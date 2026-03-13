// Command server is the DilithiAuth Identity Provider entrypoint.
//
// It starts a quantum-safe OAuth 2.0 / OIDC server with:
//   - Hybrid signing: ML-DSA-44 (NIST FIPS 204) + Ed25519
//   - OIDC Discovery endpoint (/.well-known/openid-configuration)
//   - JWKS endpoint (/.well-known/jwks.json)
//   - JWT issuance in both "thin" (sidecar) and "full" (jumbo) modes
//   - CBOR Web Token (CWT, RFC 8392) issuance
//   - OAuth 2.0 Authorization Code flow
//   - Token verification middleware for protected endpoints
//   - Key rotation with persistence
//   - MaxHeaderBytes tuned to 16 KiB for PQC-sized tokens
//
// Configuration is read from environment variables:
//
//	DILITHIAUTH_ISSUER         – base URL of the issuer (default: https://auth.dilithiauth.io)
//	DILITHIAUTH_ADDR           – listen address (default: :8080)
//	DILITHIAUTH_SIGSTORE       – SigStore backend: "memory", "file", "postgres", or "redis" (default: memory)
//	DILITHIAUTH_SIGSTORE_PATH  – file path for file-backed SigStore (default: ./data/sigstore.json)
//	DILITHIAUTH_POSTGRES_DSN   – PostgreSQL DSN for postgres-backed SigStore (default: postgres://localhost:5432/dilithiauth?sslmode=disable)
//	DILITHIAUTH_REDIS_ADDR     – Redis address for redis-backed SigStore (default: localhost:6379)
//	DILITHIAUTH_KEYS_DIR       – directory for persistent keys (default: empty = ephemeral)
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

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
		addr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := transport.NewRedisSigStore(addr)
		if err != nil {
			log.Fatalf("FATAL: creating Redis SigStore: %v", err)
		}
		sigStore = transport.NewSigStoreWithBackend(rs)
		log.Printf("SigStore: Redis-backed at %s", addr)
	default:
		sigStore = transport.NewSigStore()
		log.Print("SigStore: in-memory")
	}

	// --- Key management setup ---
	keysDir := getEnv("DILITHIAUTH_KEYS_DIR", "")
	var provider *oauth.Provider
	var km *signer.KeyManager

	if keysDir != "" {
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
	authzStore := oauth.NewAuthzStore()
	authzStore.AddUser(&oauth.DemoUser{Username: "alice", Password: "password", Subject: "alice"})
	authzStore.AddUser(&oauth.DemoUser{Username: "bob", Password: "password", Subject: "bob"})
	authzStore.AddClient(&oauth.DemoClient{
		ClientID:    "demo-client",
		RedirectURI: "http://localhost:8080/callback",
		Name:        "Demo Client",
	})
	authzHandlers := &oauth.AuthzHandlers{
		Provider: provider,
		Store:    authzStore,
		SigStore: sigStore,
	}

	// Build the router.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(transport.JumboHeaderMiddleware)
	r.Use(transport.SigStoreMiddleware(sigStore))

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
	r.Post("/token", authzHandlers.TokenExchange)

	// Key rotation admin endpoint.
	if km != nil {
		r.Post("/admin/rotate", makeRotateHandler(km))
	}

	// Protected endpoints (require valid token).
	r.Group(func(r chi.Router) {
		r.Use(v.Middleware())
		r.Get("/protected/whoami", func(w http.ResponseWriter, r *http.Request) {
			claims := verifier.ClaimsFromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(claims)
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

		// Store the ML-DSA-44 signature in the sidecar for later verification.
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

		// Store PQC signature in sidecar keyed by JWTID.
		store.Put(result.JWTID, result.PQCSig)

		// Return the raw CBOR bytes with the appropriate content type.
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
