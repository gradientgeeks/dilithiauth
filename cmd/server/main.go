// Command server is the DilithiAuth Identity Provider entrypoint.
//
// It starts a quantum-safe OAuth 2.0 / OIDC server with:
//   - Hybrid signing: ML-DSA-44 (NIST FIPS 204) + Ed25519
//   - OIDC Discovery endpoint (/.well-known/openid-configuration)
//   - JWKS endpoint (/.well-known/jwks.json)
//   - JWT issuance in both "thin" (sidecar) and "full" (jumbo) modes
//   - CBOR Web Token (CWT, RFC 8392) issuance
//   - OAuth 2.0 Authorization Code flow with multi-tenant scoping
//   - Token verification middleware for protected endpoints
//   - Key rotation with persistence
//   - MaxHeaderBytes tuned to 16 KiB for PQC-sized tokens
//   - Developer portal React SPA at /developer (REST API at /api/v1/developer)
//
// Configuration is read from environment variables:
//
//	DILITHIAUTH_ISSUER          – base URL of the issuer (default: https://auth.dilithiauth.io)
//	DILITHIAUTH_ADDR            – listen address (default: :8080)
//	DILITHIAUTH_SIGSTORE        – SigStore backend: "memory", "file", "postgres", or "redis" (default: memory)
//	DILITHIAUTH_SIGSTORE_PATH   – file path for file-backed SigStore (default: ./data/sigstore.json)
//	DILITHIAUTH_POSTGRES_DSN    – PostgreSQL DSN (default: postgres://localhost:5432/dilithiauth?sslmode=disable)
//	DILITHIAUTH_REDIS_ADDR      – Redis address (default: localhost:6379)
//	DILITHIAUTH_KEYS_DIR        – directory for persistent keys (default: empty = ephemeral)
//	DILITHIAUTH_KEYSTORE        – key store backend: "file" or "redis" (default: file)
//	DILITHIAUTH_AUTHZSTORE      – auth code store backend: "memory" or "redis" (default: memory)
//	DILITHIAUTH_SESSION_SECRET  – base64-encoded 32-byte secret for developer session JWTs (required in prod)
//	DILITHIAUTH_STATIC_DIR      – directory of built React assets (default: ./static)
package main

import (
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/developer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

//go:embed static
var staticFiles embed.FS

func main() {
	issuer := getEnv("DILITHIAUTH_ISSUER", "https://auth.dilithiauth.io")
	addr := getEnv("DILITHIAUTH_ADDR", ":8080")

	log.Printf("DilithiAuth starting – issuer=%s addr=%s", issuer, addr)

	// --- SigStore setup ---
	sigStoreType := getEnv("DILITHIAUTH_SIGSTORE", "memory")
	var sigStore *transport.SigStore
	switch sigStoreType {
	case "file":
		path := getEnv("DILITHIAUTH_SIGSTORE_PATH", "./data/sigstore.json")
		fb, err := transport.NewFileSigStore(path)
		if err != nil {
			log.Fatalf("FATAL: creating file SigStore: %v", err)
		}
		sigStore = transport.NewSigStoreWithBackend(fb)
		log.Printf("SigStore: file-backed at %s", path)
	case "postgres":
		dsn := getEnv("DILITHIAUTH_POSTGRES_DSN", "postgres://localhost:5432/dilithiauth?sslmode=disable")
		pg, err := transport.NewPostgresSigStore(dsn)
		if err != nil {
			log.Fatalf("FATAL: creating PostgreSQL SigStore: %v", err)
		}
		sigStore = transport.NewSigStoreWithBackend(pg)
		log.Printf("SigStore: PostgreSQL-backed at %s", dsn)
	case "redis":
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := transport.NewRedisSigStore(redisAddr)
		if err != nil {
			log.Fatalf("FATAL: creating Redis SigStore: %v", err)
		}
		sigStore = transport.NewSigStoreWithBackend(rs)
		log.Printf("SigStore: Redis-backed at %s", redisAddr)
	default:
		sigStore = transport.NewSigStore()
		log.Print("SigStore: in-memory")
	}

	// --- Key management setup ---
	keyStoreType := getEnv("DILITHIAUTH_KEYSTORE", "file")
	keysDir := getEnv("DILITHIAUTH_KEYS_DIR", "")
	var provider *oauth.Provider
	var km *signer.KeyManager

	if keyStoreType == "redis" {
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := signer.NewRedisKeyStore(redisAddr)
		if err != nil {
			log.Fatalf("FATAL: creating Redis KeyStore: %v", err)
		}
		km, err = signer.NewKeyManagerWithStore(rs)
		if err != nil {
			log.Fatalf("FATAL: creating key manager with Redis: %v", err)
		}
		provider = oauth.NewProviderWithKeyManager(issuer, km)
		log.Printf("Key manager: Redis-backed at %s – active kid=%s", redisAddr, km.ActiveKey().Kid)
	} else if keysDir != "" {
		var err error
		km, err = signer.NewKeyManager(keysDir)
		if err != nil {
			log.Fatalf("FATAL: creating key manager: %v", err)
		}
		provider = oauth.NewProviderWithKeyManager(issuer, km)
		log.Printf("Key manager: persistent at %s – active kid=%s", keysDir, km.ActiveKey().Kid)
	} else {
		kp, err := signer.GenerateHybridKeyPair()
		if err != nil {
			log.Fatalf("FATAL: generating hybrid key pair: %v", err)
		}
		provider = oauth.NewProvider(issuer, kp)
		log.Printf("Hybrid key pair generated (ephemeral) – kid=%s", kp.Kid)
	}

	// --- Verifier setup ---
	var v *verifier.Verifier
	if km != nil {
		v = verifier.NewVerifier(km, sigStore)
	} else {
		v = verifier.NewVerifierWithKeyPair(provider.KeyPair, sigStore)
	}

	// --- Authorization code flow setup ---
	var authzStore oauth.AuthzStore
	authzStoreType := getEnv("DILITHIAUTH_AUTHZSTORE", "memory")
	switch authzStoreType {
	case "redis":
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := oauth.NewRedisAuthzStore(redisAddr)
		if err != nil {
			log.Fatalf("FATAL: creating Redis AuthzStore: %v", err)
		}
		authzStore = rs
		log.Printf("AuthzStore: Redis-backed at %s", redisAddr)
	default:
		authzStore = oauth.NewAuthzStore()
		log.Print("AuthzStore: in-memory")
	}

	// --- Revocation store setup ---
	revStore := oauth.NewMemoryRevocationStore()
	// Purge expired entries once per hour.
	go func() {
		for range time.Tick(time.Hour) {
			revStore.Purge()
		}
	}()

	// Wire revocation into the verifier.
	v.Revocation = revStore

	// --- Admin token setup ---
	adminToken := getEnv("DILITHIAUTH_ADMIN_TOKEN", "")
	if adminToken == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("FATAL: generating admin token: %v", err)
		}
		adminToken = base64.RawURLEncoding.EncodeToString(b)
		log.Printf("WARNING: DILITHIAUTH_ADMIN_TOKEN not set – ephemeral token: %s", adminToken)
	}

	// --- Developer portal setup ---
	postgresDSN := getEnv("DILITHIAUTH_POSTGRES_DSN", "postgres://localhost:5432/dilithiauth?sslmode=disable")

	sessionSecretRaw := getEnv("DILITHIAUTH_SESSION_SECRET", "")
	var sessionSecret []byte
	if sessionSecretRaw != "" {
		var err error
		sessionSecret, err = base64.StdEncoding.DecodeString(sessionSecretRaw)
		if err != nil || len(sessionSecret) < 32 {
			log.Fatalf("FATAL: DILITHIAUTH_SESSION_SECRET must be base64-encoded 32+ bytes")
		}
	} else {
		sessionSecret = make([]byte, 32)
		if _, err := rand.Read(sessionSecret); err != nil {
			log.Fatalf("FATAL: generating session secret: %v", err)
		}
		log.Printf("WARNING: DILITHIAUTH_SESSION_SECRET not set – ephemeral session secret generated")
	}

	// Developer store and postgres authz are optional — graceful degradation when
	// DILITHIAUTH_POSTGRES_DSN is not set (local testing without PostgreSQL).
	var devStore *developer.Store
	var pgAuthz *oauth.PostgresAuthzStore
	if postgresDSN != "" {
		var storeErr error
		devStore, storeErr = developer.NewStore(postgresDSN)
		if storeErr != nil {
			log.Printf("WARNING: developer store unavailable (%v) — /developer portal disabled", storeErr)
		}
		pgAuthz, storeErr = oauth.NewPostgresAuthzStore(postgresDSN)
		if storeErr != nil {
			log.Printf("WARNING: postgres authz store unavailable (%v) — client auth uses in-memory store", storeErr)
		}
	} else {
		log.Print("INFO: DILITHIAUTH_POSTGRES_DSN not set — developer portal and postgres client auth disabled")
	}

	// Seed demo users and a local-test client under the default tenant.
	authzStore.AddUser(tenant.DefaultTenantID, &oauth.User{Username: "alice", Password: "password", Subject: "alice"})
	authzStore.AddUser(tenant.DefaultTenantID, &oauth.User{Username: "bob", Password: "password", Subject: "bob"})
	localRedirectURI := getEnv("DILITHIAUTH_LOCAL_REDIRECT_URI", "http://localhost:8080/api/auth/dilithiauth/callback")
	localClientSecret := getEnv("DILITHIAUTH_LOCAL_CLIENT_SECRET", "rootaccess-local-dev-secret")
	localSecretHash, err := bcrypt.GenerateFromPassword([]byte(localClientSecret), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("FATAL: hashing local client secret: %v", err)
	}
	authzStore.AddClient(tenant.DefaultTenantID, &oauth.OAuthClient{
		ClientID:         "local-rootaccess",
		ClientSecretHash: string(localSecretHash),
		RedirectURI:      localRedirectURI,
		Name:             "RootAccess Local",
	})
	log.Printf("Local OAuth client ready: client_id=local-rootaccess secret=%s redirect=%s", localClientSecret, localRedirectURI)

	authzHandlers := &oauth.AuthzHandlers{
		Provider:   provider,
		Store:      authzStore,
		SigStore:   sigStore,
		Revocation: revStore,
	}
	if pgAuthz != nil {
		authzHandlers.ClientValidator = pgAuthz
	}

	// --- Static assets for React SPA ---
	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("FATAL: embedding static files: %v", err)
	}
	staticFS := http.FileServer(http.FS(staticSub))

	// Build the router.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(transport.JumboHeaderMiddleware)
	r.Use(transport.SigStoreMiddleware(sigStore))

	// Developer portal — only mounted when PostgreSQL is available.
	if devStore != nil {
		devHandlers := &developer.Handlers{
			Store:         devStore,
			SessionSecret: sessionSecret,
		}
		r.Mount("/api/v1/developer", devHandlers.Router())
		r.Get("/developer", http.RedirectHandler("/developer/", http.StatusMovedPermanently).ServeHTTP)
		r.Get("/developer/*", func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/developer")
			if path == "" || path == "/" {
				path = "/index.html"
			}
			r.URL.Path = path
			if _, err := staticSub.(interface {
				Open(string) (fs.File, error)
			}).Open(strings.TrimPrefix(path, "/")); err != nil {
				r.URL.Path = "/index.html"
			}
			staticFS.ServeHTTP(w, r)
		})
	}

	// All OAuth and OIDC routes — apply tenant middleware.
	r.Group(func(r chi.Router) {
		r.Use(tenant.Middleware)

		// OIDC Discovery & JWKS.
		r.Get("/.well-known/openid-configuration", provider.DiscoveryHandler())
		r.Get("/.well-known/jwks.json", provider.JWKSHandler())

		// Token issuance endpoints (demo – no auth guard for brevity).
		r.Post("/token/thin", makeThinTokenHandler(provider, sigStore))
		r.Post("/token/full", makeFullTokenHandler(provider))
		r.Post("/token/cwt", makeCWTTokenHandler(provider, sigStore))

		// OAuth 2.0 Authorization Code flow.
		r.Get("/authorize", authzHandlers.AuthorizeGET)
		r.Post("/authorize", authzHandlers.AuthorizePOST)
		r.Get("/register", authzHandlers.RegisterGET)
		r.Post("/register", authzHandlers.RegisterPOST)
		r.Post("/token", authzHandlers.TokenExchange)

		// Token revocation endpoint (RFC 7009).
		r.Post("/revoke", authzHandlers.RevokeToken)

		// Key rotation admin endpoint — protected by bearer admin token.
		if km != nil {
			r.With(requireAdminToken(adminToken)).Post("/admin/rotate", makeRotateHandler(km))
		}

		// Protected endpoints (require valid token).
		r.Group(func(r chi.Router) {
			r.Use(v.Middleware())
			r.Get("/protected/whoami", func(w http.ResponseWriter, r *http.Request) {
				claims := verifier.ClaimsFromContext(r.Context())
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(claims)
			})
			// OIDC userinfo endpoint — returns subject identity for OAuth clients.
			r.Get("/userinfo", func(w http.ResponseWriter, r *http.Request) {
				claims := verifier.ClaimsFromContext(r.Context())
				sub := claims.Subject
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{
					"sub":   sub,
					"name":  sub,
					"email": sub + "@dilithiauth.local",
				})
			})
		})
	})

	// Health check.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Create a PQC-ready server with MaxHeaderBytes = 16 KiB.
	srv := transport.NewPQCServer(addr, r)
	srv.ReadTimeout = 15 * time.Second
	srv.WriteTimeout = 15 * time.Second
	srv.IdleTimeout = 60 * time.Second

	log.Printf("Listening on %s (MaxHeaderBytes=%d)", addr, srv.MaxHeaderBytes)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: %v", err)
	}
}

// makeThinTokenHandler issues a "thin" JWT with a pqc_ref sidecar entry.
func makeThinTokenHandler(p *oauth.Provider, store *transport.SigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		result, err := p.IssueJWTThin(subject, time.Hour)
		if err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}

		store.Put(result.PQCRef, result.PQCSig)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": result.Token,
			"token_type":   "Bearer",
			"pqc_ref":      result.PQCRef,
		})
	}
}

// makeFullTokenHandler issues a "full" / jumbo JWT with embedded ML-DSA-44 sig.
func makeFullTokenHandler(p *oauth.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		tokenStr, err := p.IssueJWTFull(subject, time.Hour)
		if err != nil {
			http.Error(w, "token issuance failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": tokenStr,
			"token_type":   "Bearer",
		})
	}
}

// makeCWTTokenHandler issues a CBOR Web Token with sidecar PQC signature.
func makeCWTTokenHandler(p *oauth.Provider, store *transport.SigStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subject := r.URL.Query().Get("sub")
		if subject == "" {
			subject = "anonymous"
		}

		result, err := p.IssueCWT(subject, time.Hour)
		if err != nil {
			http.Error(w, "CWT issuance failed", http.StatusInternalServerError)
			return
		}

		store.Put(result.JWTID, result.PQCSig)

		w.Header().Set("Content-Type", "application/cbor")
		w.Header().Set("X-Token-Id", result.JWTID)
		w.Header().Set("X-PQC-Ref", result.PQCRef)
		_, _ = w.Write(result.Token)
	}
}

// makeRotateHandler creates a handler that generates a new key pair.
func makeRotateHandler(km *signer.KeyManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newKP, err := km.Rotate()
		if err != nil {
			http.Error(w, "key rotation failed", http.StatusInternalServerError)
			return
		}
		log.Printf("Key rotated – new active kid=%s", newKP.Kid)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"kid":    newKP.Kid,
			"status": "rotated",
		})
	}
}

// requireAdminToken is middleware that enforces an Authorization: Bearer <token>
// check using constant-time comparison to prevent timing attacks.
func requireAdminToken(token string) func(http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("Authorization")
			if !strings.HasPrefix(got, "Bearer ") ||
				subtle.ConstantTimeCompare([]byte(got), want) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
