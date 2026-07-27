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
//	DILITHIAUTH_REDIS_USER      – Redis ACL username (optional; empty uses legacy requirepass auth)
//	DILITHIAUTH_REDIS_PASS      – Redis password (optional; required if the Redis instance enforces auth)
//	DILITHIAUTH_KEYS_DIR        – directory for persistent keys (default: ./data/keys)
//	DILITHIAUTH_KEYSTORE        – key store backend: "file" or "redis" (default: file)
//	DILITHIAUTH_EPHEMERAL_KEYS  – "true" to generate in-memory keys instead of persisting to DILITHIAUTH_KEYS_DIR
//	                              (default: false; every restart invalidates all previously issued tokens when true)
//	DILITHIAUTH_AUTHZSTORE      – auth code store backend: "memory" or "redis" (default: memory)
//	DILITHIAUTH_SESSION_SECRET  – base64-encoded 32-byte secret for developer session JWTs (required in prod)
//	DILITHIAUTH_STATIC_DIR      – directory of built React assets (default: ./static)
//	DILITHIAUTH_FRONTEND_URL    – origin of the developer portal frontend (default: https://dilithiauth.gradientgeeks.tech)
//	DILITHIAUTH_CORS_ALLOWED_ORIGINS – comma-separated list of origins allowed for credentialed CORS requests
//	                              (default: DILITHIAUTH_FRONTEND_URL plus https://dilithiauth.gradientgeeks.tech)
//	DILITHIAUTH_TRUSTED_TENANT_HEADER_CIDRS – comma-separated CIDR ranges allowed to set X-Tenant-Id
//	                              (default: empty = the header is never honored; all callers resolve via subdomain/default)
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"

	"github.com/Uttam-Mahata/DilithiAuth/pkg/bff"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/developer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/oauth"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/ratelimit"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/signer"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/tenant"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/transport"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/userstore"
	"github.com/Uttam-Mahata/DilithiAuth/pkg/verifier"
)

//go:embed static
var staticFiles embed.FS

// logger is the process-wide structured logger. JSON output is used
// unconditionally (rather than switching on an environment/dev flag) to keep
// this change bounded: JSON is safe for both local `| jq` inspection and
// production log aggregation (e.g. journald/systemd forwarding to a log
// pipeline), and every existing log line here becomes a structured event
// with fields instead of an interpolated string.
var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// fatal logs msg as an error with fields, then exits the process with a
// non-zero status — a structured drop-in replacement for log.Fatalf.
func fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}

func main() {
	issuer := getEnv("DILITHIAUTH_ISSUER", "https://auth.dilithiauth.io")
	addr := getEnv("DILITHIAUTH_ADDR", ":8080")

	logger.Info("DilithiAuth starting", "issuer", issuer, "addr", addr)

	// tokenTTL must track the TTL passed to Provider.IssueJWTThin/IssueJWTFull/
	// IssueCWT (currently hardcoded to time.Hour at every call site) so that
	// SigStore entries expire neither before nor long after the tokens
	// referencing them are no longer valid.
	const tokenTTL = time.Hour

	// --- SigStore setup ---
	// Every backend expires entries after tokenTTL so the store doesn't grow
	// unboundedly for the life of the process/database — a signature never
	// needs to outlive the access token that references it.
	sigStoreType := getEnv("DILITHIAUTH_SIGSTORE", "memory")
	var sigStore *transport.SigStore
	switch sigStoreType {
	case "file":
		path := getEnv("DILITHIAUTH_SIGSTORE_PATH", "./data/sigstore.json")
		fb, err := transport.NewFileSigStoreWithTTL(path, tokenTTL)
		if err != nil {
			fatal("creating file SigStore", "error", err)
		}
		sigStore = transport.NewSigStoreWithBackend(fb)
		logger.Info("SigStore backend selected", "backend", "file", "path", path)
	case "postgres":
		dsn := getEnv("DILITHIAUTH_POSTGRES_DSN", "postgres://localhost:5432/dilithiauth?sslmode=disable")
		pg, err := transport.NewPostgresSigStoreWithTTL(dsn, tokenTTL)
		if err != nil {
			fatal("creating PostgreSQL SigStore", "error", err)
		}
		sigStore = transport.NewSigStoreWithBackend(pg)
		logger.Info("SigStore backend selected", "backend", "postgres")
	case "redis":
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		redisUser := getEnv("DILITHIAUTH_REDIS_USER", "")
		redisPass := getEnv("DILITHIAUTH_REDIS_PASS", "")
		rs, err := transport.NewRedisSigStore(redisAddr, transport.WithTTL(tokenTTL), transport.WithAuth(redisUser, redisPass))
		if err != nil {
			fatal("creating Redis SigStore", "error", err)
		}
		sigStore = transport.NewSigStoreWithBackend(rs)
		logger.Info("SigStore backend selected", "backend", "redis", "addr", redisAddr, "user", redisUser)
	default:
		sigStore = transport.NewSigStoreWithBackend(transport.NewMemorySigStoreWithTTL(tokenTTL))
		logger.Info("SigStore backend selected", "backend", "memory")
	}

	// --- Key management setup ---
	// Keys are persisted to disk by default (DILITHIAUTH_KEYS_DIR) so that
	// restarts and multi-instance deployments don't each mint incompatible
	// signing keys, silently invalidating every previously issued token.
	// Ephemeral in-memory keys are only used when explicitly requested via
	// DILITHIAUTH_EPHEMERAL_KEYS=true (e.g. for tests/local one-off runs).
	keyStoreType := getEnv("DILITHIAUTH_KEYSTORE", "file")
	keysDir := getEnv("DILITHIAUTH_KEYS_DIR", "./data/keys")
	ephemeralKeys := getEnv("DILITHIAUTH_EPHEMERAL_KEYS", "false") == "true"
	var provider *oauth.Provider
	var km *signer.KeyManager

	if keyStoreType == "redis" {
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := signer.NewRedisKeyStore(redisAddr)
		if err != nil {
			fatal("creating Redis KeyStore", "error", err)
		}
		km, err = signer.NewKeyManagerWithStore(rs)
		if err != nil {
			fatal("creating key manager with Redis", "error", err)
		}
		provider = oauth.NewProviderWithKeyManager(issuer, km)
		logger.Info("Key manager backend selected", "backend", "redis", "addr", redisAddr, "kid", km.ActiveKey().Kid)
	} else if ephemeralKeys {
		kp, err := signer.GenerateHybridKeyPair()
		if err != nil {
			fatal("generating hybrid key pair", "error", err)
		}
		provider = oauth.NewProvider(issuer, kp)
		logger.Warn("hybrid key pair generated ephemerally (DILITHIAUTH_EPHEMERAL_KEYS=true) — every previously issued token will be invalidated on the next restart", "kid", kp.Kid)
	} else {
		var err error
		km, err = signer.NewKeyManager(keysDir)
		if err != nil {
			fatal("creating key manager", "error", err)
		}
		provider = oauth.NewProviderWithKeyManager(issuer, km)
		logger.Info("Key manager backend selected", "backend", "file", "dir", keysDir, "kid", km.ActiveKey().Kid)
	}

	// --- Verifier setup ---
	var v *verifier.Verifier
	if km != nil {
		v = verifier.NewVerifier(km, sigStore)
	} else {
		v = verifier.NewVerifierWithKeyPair(provider.KeyPair, sigStore)
	}
	v.Issuer = issuer

	// --- Authorization code flow setup ---
	var authzStore oauth.AuthzStore
	authzStoreType := getEnv("DILITHIAUTH_AUTHZSTORE", "memory")
	switch authzStoreType {
	case "redis":
		redisAddr := getEnv("DILITHIAUTH_REDIS_ADDR", "localhost:6379")
		rs, err := oauth.NewRedisAuthzStore(redisAddr)
		if err != nil {
			fatal("creating Redis AuthzStore", "error", err)
		}
		authzStore = rs
		logger.Info("AuthzStore backend selected", "backend", "redis", "addr", redisAddr)
	default:
		authzStore = oauth.NewAuthzStore()
		logger.Info("AuthzStore backend selected", "backend", "memory")
	}

	// --- Revocation store setup ---
	revStore := oauth.NewMemoryRevocationStore()
	sessionRevStore := developer.NewMemorySessionRevocationStore()
	refreshTokenStore := oauth.NewMemoryRefreshTokenStore()
	// Purge expired entries once per hour.
	go func() {
		for range time.Tick(time.Hour) {
			revStore.Purge()
			sessionRevStore.Purge()
			sigStore.Purge()
			refreshTokenStore.Purge()
		}
	}()

	// Wire revocation into the verifier.
	v.Revocation = revStore

	// --- Admin token setup ---
	adminToken := getEnv("DILITHIAUTH_ADMIN_TOKEN", "")
	if adminToken == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			fatal("generating admin token", "error", err)
		}
		adminToken = base64.RawURLEncoding.EncodeToString(b)
		logger.Warn("DILITHIAUTH_ADMIN_TOKEN not set — generated an ephemeral admin token", "admin_token", adminToken)
	}

	// --- Developer portal setup ---
	postgresDSN := getEnv("DILITHIAUTH_POSTGRES_DSN", "postgres://localhost:5432/dilithiauth?sslmode=disable")

	sessionSecretRaw := getEnv("DILITHIAUTH_SESSION_SECRET", "")
	var sessionSecret []byte
	if sessionSecretRaw != "" {
		var err error
		sessionSecret, err = base64.StdEncoding.DecodeString(sessionSecretRaw)
		if err != nil || len(sessionSecret) < 32 {
			fatal("DILITHIAUTH_SESSION_SECRET must be base64-encoded 32+ bytes")
		}
	} else {
		sessionSecret = make([]byte, 32)
		if _, err := rand.Read(sessionSecret); err != nil {
			fatal("generating session secret", "error", err)
		}
		logger.Warn("DILITHIAUTH_SESSION_SECRET not set — generated an ephemeral session secret")
	}

	// Developer store, postgres authz, and the end-user store are optional —
	// graceful degradation when DILITHIAUTH_POSTGRES_DSN is not set (local
	// testing without PostgreSQL).
	var devStore *developer.Store
	var pgAuthz *oauth.PostgresAuthzStore
	var pgUsers *userstore.PostgresUserStore
	if postgresDSN != "" {
		var storeErr error
		devStore, storeErr = developer.NewStore(postgresDSN)
		if storeErr != nil {
			logger.Warn("developer store unavailable — /developer portal disabled", "error", storeErr)
		}
		pgAuthz, storeErr = oauth.NewPostgresAuthzStore(postgresDSN)
		if storeErr != nil {
			logger.Warn("postgres authz store unavailable — client auth uses in-memory store", "error", storeErr)
		}
		pgUsers, storeErr = userstore.NewPostgresUserStore(postgresDSN)
		if storeErr != nil {
			logger.Warn("postgres user store unavailable — end-user accounts use in-memory store (lost on restart)", "error", storeErr)
		}
	} else {
		logger.Info("DILITHIAUTH_POSTGRES_DSN not set — developer portal, postgres client auth, and durable end-user accounts disabled")
	}

	// Seed demo users and a local-test client under the default tenant.
	authzStore.AddUser(tenant.DefaultTenantID, &oauth.User{Username: "alice", Password: "password", Subject: "alice"})
	authzStore.AddUser(tenant.DefaultTenantID, &oauth.User{Username: "bob", Password: "password", Subject: "bob"})
	localRedirectURI := getEnv("DILITHIAUTH_LOCAL_REDIRECT_URI", "http://localhost:8080/api/auth/dilithiauth/callback")
	localClientSecret := getEnv("DILITHIAUTH_LOCAL_CLIENT_SECRET", "rootaccess-local-dev-secret")
	localSecretHash, err := bcrypt.GenerateFromPassword([]byte(localClientSecret), bcrypt.DefaultCost)
	if err != nil {
		fatal("hashing local client secret", "error", err)
	}
	authzStore.AddClient(tenant.DefaultTenantID, &oauth.OAuthClient{
		ClientID:         "local-rootaccess",
		ClientSecretHash: string(localSecretHash),
		RedirectURI:      localRedirectURI,
		Name:             "RootAccess Local",
	})
	logger.Info("local OAuth client ready", "client_id", "local-rootaccess", "redirect_uri", localRedirectURI)

	consentSecret := make([]byte, 32)
	if _, err := rand.Read(consentSecret); err != nil {
		fatal("generating consent secret", "error", err)
	}

	googleClientID := getEnv("DILITHIAUTH_GOOGLE_CLIENT_ID", "")
	googleClientSecret := getEnv("DILITHIAUTH_GOOGLE_CLIENT_SECRET", "")
	googleRedirectURI := getEnv("DILITHIAUTH_GOOGLE_REDIRECT_URI", issuer+"/auth/social/google/callback")

	githubClientID := getEnv("DILITHIAUTH_GITHUB_CLIENT_ID", "")
	githubClientSecret := getEnv("DILITHIAUTH_GITHUB_CLIENT_SECRET", "")
	githubRedirectURI := getEnv("DILITHIAUTH_GITHUB_REDIRECT_URI", issuer+"/auth/social/github/callback")

	authzHandlers := &oauth.AuthzHandlers{
		Provider:      provider,
		Store:         authzStore,
		SigStore:      sigStore,
		Revocation:    revStore,
		Verifier:      &verifier.OAuthAdapter{Verifier: v},
		RefreshTokens: refreshTokenStore,
		GrantStore:    oauth.NewMemoryGrantStore(),
		ConsentSecret: consentSecret,
		GoogleConfig: oauth.SocialConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURI:  googleRedirectURI,
		},
		GitHubConfig: oauth.SocialConfig{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURI:  githubRedirectURI,
		},
	}
	if pgAuthz != nil {
		authzHandlers.ClientValidator = &oauth.CompositeClientValidator{
			Primary:   pgAuthz,
			Secondary: authzStore,
		}
	}
	if pgUsers != nil {
		// UserStore and UserRegistrar must be wired together: registration and
		// authentication have to target the same backend, or accounts created
		// via one would be invisible to the other.
		authzHandlers.UserStore = pgUsers
		authzHandlers.UserRegistrar = pgUsers
		logger.Info("end-user account backend selected", "backend", "postgres")
	} else {
		logger.Info("end-user account backend selected", "backend", "memory")
	}

	// --- Static assets for React SPA ---
	staticSub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		fatal("embedding static files", "error", err)
	}
	staticFS := http.FileServer(http.FS(staticSub))

	// --- CORS allow-list setup ---
	frontendURL := getEnv("DILITHIAUTH_FRONTEND_URL", "https://dilithiauth.gradientgeeks.tech")
	corsOriginsRaw := getEnv("DILITHIAUTH_CORS_ALLOWED_ORIGINS", frontendURL+",https://dilithiauth.gradientgeeks.tech")
	allowedOrigins := parseOrigins(corsOriginsRaw)
	if len(allowedOrigins) == 0 {
		logger.Warn("no CORS allowed origins configured — credentialed cross-origin requests will be rejected")
	}

	// Build the router.
	r := chi.NewRouter()
	// RequestID must run before Logger so every access log line — and any
	// handler that pulls the ID via middleware.GetReqID(ctx) — carries a
	// correlation ID traceable across a single request's log lines.
	r.Use(middleware.RequestID)
	r.Use(newCORSMiddleware(allowedOrigins))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(transport.JumboHeaderMiddleware)
	r.Use(transport.SigStoreMiddleware(sigStore))

	// Developer portal API — mounted with Postgres or Memory fallback.
	mailer := developer.NewMailerFromEnv()
	otpStore := developer.NewOTPStore(mailer)

	// End-user password-reset OTP store — a separate instance from the
	// developer-portal one above (shares the same Mailer, but must not share
	// the same code map: a developer-portal email-verification code and an
	// end-user password-reset code for the same address would otherwise
	// overwrite each other, since OTPVerificationStore holds one active code
	// per email regardless of purpose).
	endUserOTPStore := developer.NewOTPStore(mailer)
	authzHandlers.OTPStore = endUserOTPStore
	if pgUsers != nil {
		authzHandlers.PasswordResetter = pgUsers
	}
	var effectiveDevStore developer.DevStore
	if devStore != nil {
		effectiveDevStore = devStore
	} else {
		effectiveDevStore = developer.NewMemoryDevStore()
	}
	devHandlers := &developer.Handlers{
		Store:             effectiveDevStore,
		OTPStore:          otpStore,
		SessionSecret:     sessionSecret,
		SessionRevocation: sessionRevStore,
	}
	r.Mount("/api/v1/developer", devHandlers.Router())

	// Registered OAuth Client for DilithiAuth's own Developer Portal Console (Dogfooding)
	bffCallbackURL := issuer + "/api/v1/bff/callback"
	portalClient := &oauth.OAuthClient{
		ClientID:    "da_developer_portal",
		RedirectURI: bffCallbackURL,
		Name:        "DilithiAuth Developer Portal Console",
	}
	authzStore.AddClient(tenant.DefaultTenantID, portalClient)
	authzStore.AddClient("dilithiauthapis", portalClient)

	bffServer, bffErr := bff.NewServer(bff.Config{
		IssuerURL:     issuer,
		ClientID:      "da_developer_portal",
		RedirectURI:   bffCallbackURL,
		FrontendURL:   frontendURL,
		SessionSecret: sessionSecret,
	}, v)
	if bffErr == nil {
		r.Mount("/api/v1/bff", bffServer.Router())
		logger.Info("BFF (Backend for Frontend) ready", "frontend_url", frontendURL)
	}

	// Always serve static Developer Portal React SPA.
	r.Handle("/assets/*", staticFS)
	r.Get("/developer", http.RedirectHandler("/developer/", http.StatusMovedPermanently).ServeHTTP)
	r.Get("/developer/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/developer")
		if path == "" || path == "/" {
			path = "/index.html"
		}

		relPath := strings.TrimPrefix(path, "/")
		file, err := staticSub.(interface {
			Open(string) (fs.File, error)
		}).Open(relPath)

		if err != nil || relPath == "index.html" {
			// Serve embedded index.html directly for SPA client-side routing
			indexBytes, readErr := fs.ReadFile(staticSub, "index.html")
			if readErr != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexBytes)
			return
		}
		defer file.Close()

		r.URL.Path = path
		staticFS.ServeHTTP(w, r)
	})

	// All OAuth and OIDC routes — apply tenant middleware. X-Tenant-Id is
	// only honored from a configured trusted-network range; any other
	// caller sending it is silently ignored (falls back to subdomain or
	// default resolution), preventing tenant-spoofing via that header.
	trustedTenantCIDRs := strings.Split(getEnv("DILITHIAUTH_TRUSTED_TENANT_HEADER_CIDRS", ""), ",")
	r.Group(func(r chi.Router) {
		r.Use(tenant.NewMiddleware(trustedTenantCIDRs))

		// OIDC Discovery & JWKS.
		r.Get("/.well-known/openid-configuration", provider.DiscoveryHandler())
		r.Get("/.well-known/jwks.json", provider.JWKSHandler())

		// Unauthenticated token-minting endpoints for local demos/testing only.
		// Compiled in exclusively via the "demo" build tag (see demo_routes.go);
		// a standard `go build ./cmd/server` never includes them, since they let
		// any caller mint a valid access token for an arbitrary subject.
		registerDemoRoutes(r, provider, sigStore)

		// OAuth 2.0 Authorization Code flow. POST endpoints are rate-limited
		// per client IP to blunt credential-stuffing/brute-force attempts.
		rl := func(l *ratelimit.Limiter) func(http.Handler) http.Handler {
			return ratelimit.Middleware(l, ratelimit.ClientIPKey())
		}
		r.Get("/authorize", authzHandlers.AuthorizeGET)
		r.With(rl(ratelimit.NewLimiter(1.0/2, 10))).Post("/authorize", authzHandlers.AuthorizePOST)
		r.Get("/register", authzHandlers.RegisterGET)
		r.With(rl(ratelimit.NewLimiter(1.0/5, 5))).Post("/register", authzHandlers.RegisterPOST)
		r.Get("/forgot-password", authzHandlers.ForgotPasswordGET)
		r.With(rl(ratelimit.NewLimiter(1.0/20, 3))).Post("/forgot-password", authzHandlers.ForgotPasswordPOST)
		r.Get("/reset-password", authzHandlers.ResetPasswordGET)
		r.With(rl(ratelimit.NewLimiter(1.0/5, 5))).Post("/reset-password", authzHandlers.ResetPasswordPOST)
		r.With(rl(ratelimit.NewLimiter(1.0/2, 10))).Post("/consent", authzHandlers.ConsentPOST)
		r.With(rl(ratelimit.NewLimiter(2, 20))).Post("/token", authzHandlers.TokenExchange)

		// Social Authentication endpoints.
		r.Get("/auth/social/{provider}", authzHandlers.SocialLoginGET)
		r.Get("/auth/social/{provider}/callback", authzHandlers.SocialCallbackGET)

		// Token revocation endpoint (RFC 7009).
		r.With(rl(ratelimit.NewLimiter(2, 20))).Post("/revoke", authzHandlers.RevokeToken)

		// Verification REST endpoint for lightweight clients.
		r.Post("/verify/jwt", authzHandlers.VerifyJWTPOST)

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
			// OIDC userinfo endpoint — returns the real registered identity for
			// OAuth clients, gated by the scopes actually granted to this token
			// (RFC 6749/OIDC convention: don't return claims the user never
			// consented to share).
			r.Get("/userinfo", func(w http.ResponseWriter, r *http.Request) {
				claims := verifier.ClaimsFromContext(r.Context())
				resp := map[string]string{"sub": claims.Subject}

				if oauth.ScopeContains(claims.Scope, "profile") {
					resp["name"] = claims.Subject
				}
				if oauth.ScopeContains(claims.Scope, "email") {
					if email := lookupUserEmail(r.Context(), pgUsers, authzStore, claims.TenantID, claims.Subject); email != "" {
						resp["email"] = email
					}
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			})
		})
	})

	// Liveness check — always 200 once the process is up and serving. Load
	// balancers/orchestrators should NOT use this for routing decisions
	// (see /readyz below), only for "is the process alive" restarts.
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Readiness check — verifies every dependency actually configured for
	// this process (Postgres stores that connected successfully, Redis
	// backends when selected) is currently reachable, and 503s otherwise.
	// This is what a load balancer/orchestrator should point at for routing
	// decisions: a process can be alive (/healthz) yet unable to serve
	// traffic (e.g. Postgres/Redis is down or still starting up).
	r.Get("/readyz", readyzHandler(sigStore, authzStore, pgUsers, pgAuthz, devStore))

	// Create a PQC-ready server with MaxHeaderBytes = 16 KiB.
	srv := transport.NewPQCServer(addr, r)
	srv.ReadTimeout = 15 * time.Second
	srv.WriteTimeout = 15 * time.Second
	srv.IdleTimeout = 60 * time.Second

	// --- Serve with graceful shutdown ---
	// ListenAndServe runs in its own goroutine so the main goroutine can
	// block on a shutdown signal instead of forever inside Serve — without
	// this, a SIGTERM (e.g. `systemctl restart`, container orchestrator
	// stop) kills the process and every in-flight request immediately.
	serveErrCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", addr, "max_header_bytes", srv.MaxHeaderBytes)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErrCh <- err
			return
		}
		serveErrCh <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErrCh:
		if err != nil {
			fatal("server failed", "error", err)
		}
	case <-ctx.Done():
		stop() // restore default signal behavior (e.g. a second Ctrl-C force-quits)
		logger.Info("shutdown signal received, draining in-flight requests")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown did not complete cleanly", "error", err)
		} else {
			logger.Info("shutdown complete")
		}

		// Best-effort close of database/Redis connections opened above. Each
		// is nil when its backend wasn't configured/selected, so every call
		// is guarded — Close on a nil *Store would otherwise panic.
		if devStore != nil {
			if err := devStore.Close(); err != nil {
				logger.Warn("error closing developer store during shutdown", "error", err)
			}
		}
		if pgAuthz != nil {
			if err := pgAuthz.Close(); err != nil {
				logger.Warn("error closing postgres authz store during shutdown", "error", err)
			}
		}
		if pgUsers != nil {
			if err := pgUsers.Close(); err != nil {
				logger.Warn("error closing postgres user store during shutdown", "error", err)
			}
		}
		if err := sigStore.Close(); err != nil {
			logger.Warn("error closing sig store during shutdown", "error", err)
		}
	}
}

// pinger is satisfied by any configured dependency readyzHandler can check
// for reachability — *transport.SigStore, *oauth.PostgresAuthzStore,
// *userstore.PostgresUserStore, *developer.Store, and oauth.AuthzStore
// implementations that wrap a Redis client (e.g. *oauth.RedisAuthzStore) all
// satisfy this via their Ping(ctx) method.
type pinger interface {
	Ping(ctx context.Context) error
}

// readyzHandler returns a handler that pings every dependency actually
// configured for this process (skipping any that are nil, i.e. not
// configured, or that don't expose a Ping method, i.e. in-memory/file
// backends with nothing external to check) and responds 503 if any of them
// is unreachable, 200 otherwise. Unlike /healthz (always 200 once the
// process is serving), this is what a load balancer/orchestrator should use
// for routing decisions, since a live process can still be unable to serve
// traffic when its database or cache is down.
func readyzHandler(sigStore *transport.SigStore, authzStore oauth.AuthzStore, pgUsers *userstore.PostgresUserStore, pgAuthz *oauth.PostgresAuthzStore, devStore *developer.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		checks := map[string]error{}

		// sigStore.Ping is always safe to call — it no-ops for in-memory/file
		// backends and only actually pings when Postgres/Redis-backed.
		if err := sigStore.Ping(ctx); err != nil {
			checks["sigstore"] = err
		}
		if p, ok := authzStore.(pinger); ok {
			if err := p.Ping(ctx); err != nil {
				checks["authzstore"] = err
			}
		}
		if pgUsers != nil {
			if err := pgUsers.Ping(ctx); err != nil {
				checks["postgres_userstore"] = err
			}
		}
		if pgAuthz != nil {
			if err := pgAuthz.Ping(ctx); err != nil {
				checks["postgres_authzstore"] = err
			}
		}
		if devStore != nil {
			if err := devStore.Ping(ctx); err != nil {
				checks["developer_store"] = err
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if len(checks) > 0 {
			failures := make(map[string]string, len(checks))
			for name, err := range checks {
				failures[name] = err.Error()
				logger.Warn("readyz check failed", "dependency", name, "error", err)
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "not ready",
				"errors": failures,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
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
		logger.Info("key rotated", "kid", newKP.Kid)
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

// lookupUserEmail returns the real registered email for subject, using
// whichever backend owns end-user accounts (PostgreSQL when configured,
// else the in-memory/Redis AuthzStore) — the same "UserStore if set, else
// Store" precedence used throughout AuthzHandlers. Returns "" on any
// lookup failure rather than erroring, since a missing email just means
// /userinfo omits the claim.
func lookupUserEmail(ctx context.Context, pgUsers *userstore.PostgresUserStore, authzStore oauth.AuthzStore, tenantID, subject string) string {
	if pgUsers != nil {
		if u, err := pgUsers.GetBySubject(ctx, tenantID, subject); err == nil {
			return u.Email
		}
		return ""
	}
	if u, err := authzStore.GetUserBySubject(tenantID, subject); err == nil {
		return u.Email
	}
	return ""
}

// parseOrigins splits a comma-separated list of origins into a lookup set,
// trimming whitespace and dropping empty entries.
func parseOrigins(raw string) map[string]bool {
	origins := make(map[string]bool)
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = true
		}
	}
	return origins
}

// newCORSMiddleware allows cross-origin requests only from an explicit set of
// allowed origins. Reflecting an arbitrary Origin header while also setting
// Access-Control-Allow-Credentials: true would let any site make credentialed
// requests against this API, so Access-Control-Allow-Origin is only ever set
// to the request's Origin when it exactly matches an allow-listed entry.
func newCORSMiddleware(allowedOrigins map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowedOrigins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
