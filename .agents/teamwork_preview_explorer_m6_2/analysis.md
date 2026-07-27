# Social Authentication Route & Environment Variable Configuration Analysis

**Project:** DilithiAuth (Post-Quantum Identity Provider)  
**Agent:** teamwork_preview_explorer_m6_2  
**Date:** 2026-07-27  

---

## 1. Executive Summary

This report analyzes `cmd/server/main.go` and the `pkg/oauth` package in DilithiAuth to design the environment variable configuration and HTTP route definitions for Social Authentication (Google and GitHub OAuth 2.0 integration). 

`pkg/oauth/social.go` already defines the core primitives (`SocialProvider`, `SocialConfig`, `SocialUser`, `GetSocialAuthURL`, and `ExchangeSocialCode`). This investigation specifies how `cmd/server/main.go` and `pkg/oauth/authz_handlers.go` should be updated to read configuration from environment variables and wire up the social auth initiation and callback endpoints.

---

## 2. Examination of `cmd/server/main.go`

### 2.1 Server Architecture and Lifecyle
- **Entry Point:** `cmd/server/main.go` initializes and runs the HTTP server using `github.com/go-chi/chi/v5`.
- **Key Components Instantiated in `main()`:**
  - `transport.SigStore` (Signature sidecar store supporting `memory`, `file`, `postgres`, `redis`).
  - `signer.KeyManager` & `oauth.Provider` (Hybrid signing with NIST ML-DSA-44 and Ed25519).
  - `verifier.Verifier` (Token verification engine).
  - `oauth.AuthzStore` (In-memory or Redis-backed auth code & user store).
  - `userstore.PostgresUserStore` (Optional durable user identity store).
  - `developer.Store` & `bff.Server` (Developer portal and BFF integration).

### 2.2 Middleware Stack & Routing Hierarchy
The router is initialized at line 341 (`r := chi.NewRouter()`). The global middleware stack includes:
1. `middleware.RequestID` (line 345) — Assigns correlation IDs.
2. `newCORSMiddleware(allowedOrigins)` (line 346) — Enforces origin-restricted CORS.
3. `middleware.Logger` (line 347) — Request logging.
4. `middleware.Recoverer` (line 348) — Panic recovery.
5. `transport.JumboHeaderMiddleware` (line 349) — Configures 16 KiB header limits for quantum-safe tokens.
6. `transport.SigStoreMiddleware(sigStore)` (line 350) — SigStore context propagation.

Sub-routers mounted:
- `/api/v1/developer` -> `devHandlers.Router()` (line 379)
- `/api/v1/bff` -> `bffServer.Router()` (line 401)
- `/assets/*` and `/developer/*` -> React SPA static file handler (lines 404-432)

OAuth/OIDC Group (wrapped with tenant middleware `tenant.NewMiddleware(trustedTenantCIDRs)` at line 440):
- Discovery: `GET /.well-known/openid-configuration`, `GET /.well-known/jwks.json`
- Auth Code Flow: `GET/POST /authorize`, `GET/POST /register`, `GET/POST /forgot-password`, `GET/POST /reset-password`, `POST /consent`, `POST /token`, `POST /revoke`, `POST /verify/jwt`
- Protected routes: `v.Middleware()` guarding `/protected/whoami` and `/userinfo`

---

## 3. Environment Variable Parsing Analysis

### 3.1 Existing Pattern in `cmd/server/main.go`
Env parsing relies on a lightweight `getEnv` helper function defined at lines 694-699:
```go
func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

Current environment variables read in `main()`:
| Variable Name | Default Value | Purpose |
|---|---|---|
| `DILITHIAUTH_ISSUER` | `https://auth.dilithiauth.io` | Base URL of the IdP issuer |
| `DILITHIAUTH_ADDR` | `:8080` | Server listen address |
| `DILITHIAUTH_SIGSTORE` | `memory` | Signature sidecar backend |
| `DILITHIAUTH_SIGSTORE_PATH` | `./data/sigstore.json` | File path for file SigStore |
| `DILITHIAUTH_POSTGRES_DSN` | `postgres://localhost:5432/...` | PostgreSQL connection DSN |
| `DILITHIAUTH_REDIS_ADDR` | `localhost:6379` | Redis address |
| `DILITHIAUTH_KEYS_DIR` | `./data/keys` | Directory for persistent keys |
| `DILITHIAUTH_KEYSTORE` | `file` | Key store backend |
| `DILITHIAUTH_EPHEMERAL_KEYS` | `false` | Force in-memory ephemeral keys |
| `DILITHIAUTH_AUTHZSTORE` | `memory` | Auth code store backend |
| `DILITHIAUTH_ADMIN_TOKEN` | Generated 32-byte base64 | Admin authorization token |
| `DILITHIAUTH_SESSION_SECRET` | Generated 32-byte secret | Developer session JWT secret |
| `DILITHIAUTH_FRONTEND_URL` | `https://dilithiauth.gradientgeeks.tech` | Dev portal origin |
| `DILITHIAUTH_CORS_ALLOWED_ORIGINS` | Derived from frontend URL | CORS allow-list |
| `DILITHIAUTH_TRUSTED_TENANT_HEADER_CIDRS` | `""` | Trusted CIDR ranges for `X-Tenant-Id` |

### 3.2 Social Auth Environment Variable Plan
To enable Google and GitHub social login, `cmd/server/main.go` should read the following environment variables:

| Variable Name | Recommended Default | Description |
|---|---|---|
| `DILITHIAUTH_GOOGLE_CLIENT_ID` | `""` | Google OAuth2 Client ID |
| `DILITHIAUTH_GOOGLE_CLIENT_SECRET` | `""` | Google OAuth2 Client Secret |
| `DILITHIAUTH_GOOGLE_REDIRECT_URI` | `issuer + "/auth/social/google/callback"` | Redirect URI registered in Google Console |
| `DILITHIAUTH_GITHUB_CLIENT_ID` | `""` | GitHub OAuth2 Client ID |
| `DILITHIAUTH_GITHUB_CLIENT_SECRET` | `""` | GitHub OAuth2 Client Secret |
| `DILITHIAUTH_GITHUB_REDIRECT_URI` | `issuer + "/auth/social/github/callback"` | Redirect URI registered in GitHub Developer Settings |

### 3.3 Injection into Handlers
In `cmd/server/main.go`, construct `oauth.SocialConfig` structs and inject them into `oauth.AuthzHandlers`:

```go
// Parse Social Auth credentials
googleConfig := oauth.SocialConfig{
    ClientID:     getEnv("DILITHIAUTH_GOOGLE_CLIENT_ID", ""),
    ClientSecret: getEnv("DILITHIAUTH_GOOGLE_CLIENT_SECRET", ""),
    RedirectURI:  getEnv("DILITHIAUTH_GOOGLE_REDIRECT_URI", issuer+"/auth/social/google/callback"),
}

githubConfig := oauth.SocialConfig{
    ClientID:     getEnv("DILITHIAUTH_GITHUB_CLIENT_ID", ""),
    ClientSecret: getEnv("DILITHIAUTH_GITHUB_CLIENT_SECRET", ""),
    RedirectURI:  getEnv("DILITHIAUTH_GITHUB_REDIRECT_URI", issuer+"/auth/social/github/callback"),
}

// Log status
if googleConfig.ClientID != "" {
    logger.Info("Google social authentication enabled", "redirect_uri", googleConfig.RedirectURI)
} else {
    logger.Info("Google social authentication disabled (DILITHIAUTH_GOOGLE_CLIENT_ID not set)")
}

if githubConfig.ClientID != "" {
    logger.Info("GitHub social authentication enabled", "redirect_uri", githubConfig.RedirectURI)
} else {
    logger.Info("GitHub social authentication disabled (DILITHIAUTH_GITHUB_CLIENT_ID not set)")
}
```

Update `oauth.AuthzHandlers` struct definition in `pkg/oauth/authz_handlers.go`:
```go
type AuthzHandlers struct {
    Provider         *Provider
    Store            AuthzStore
    UserStore        UserAuthenticator
    UserRegistrar    UserRegistrar
    PasswordResetter PasswordResetter
    OTPStore         OTPSender
    SigStore         SigStorePutter
    Revocation       RevocationStore
    ClientValidator  ClientValidator
    Verifier         TokenVerifier
    RefreshTokens    RefreshTokenStore
    RefreshTTL       time.Duration
    GrantStore       GrantStore
    ConsentSecret    []byte

    // Social Authentication Configurations
    GoogleConfig SocialConfig
    GitHubConfig SocialConfig
}
```

---

## 4. HTTP Route Definitions & Flow Design

### 4.1 Route Registrations in `cmd/server/main.go`
Inside the tenant-aware router group (`r.Group(func(r chi.Router) { r.Use(tenant.NewMiddleware(trustedTenantCIDRs)) ... })`), register the 4 social authentication routes:

```go
// Social Authentication Routes
r.Get("/auth/social/google", authzHandlers.SocialGoogleInitiate)
r.Get("/auth/social/google/callback", authzHandlers.SocialGoogleCallback)
r.Get("/auth/social/github", authzHandlers.SocialGitHubInitiate)
r.Get("/auth/social/github/callback", authzHandlers.SocialGitHubCallback)
```

---

### 4.2 Endpoint Specifications

#### 1. `GET /auth/social/google`
- **Handler:** `authzHandlers.SocialGoogleInitiate(w http.ResponseWriter, r *http.Request)`
- **Behavior:**
  1. Check if `h.GoogleConfig.ClientID == ""`. If empty, return `http.StatusNotImplemented` (`501`) with JSON error `{"error": "social auth provider 'google' not configured"}`.
  2. Generate a secure 32-byte random state token (URL safe base64).
  3. Store state in an `HttpOnly`, `SameSite=Lax`, 10-minute cookie (name: `dilithiauth_social_state_google`).
  4. Generate Google authorization URL via `oauth.GetSocialAuthURL(oauth.ProviderGoogle, h.GoogleConfig, state)`.
  5. Redirect user to Google via `http.Redirect(w, r, authURL, http.StatusFound)`.

#### 2. `GET /auth/social/google/callback`
- **Handler:** `authzHandlers.SocialGoogleCallback(w http.ResponseWriter, r *http.Request)`
- **Behavior:**
  1. Retrieve `state` query param from request URL and compare with `dilithiauth_social_state_google` cookie using `crypto/subtle.ConstantTimeCompare`. Return `http.StatusBadRequest` (`400`) if invalid or missing.
  2. Clear state cookie.
  3. Check for `error` parameter returned by Google (e.g. `access_denied`).
  4. Extract `code` parameter from URL query.
  5. Call `oauth.ExchangeSocialCode(r.Context(), oauth.ProviderGoogle, h.GoogleConfig, code)` to exchange code for `SocialUser` (`ID`, `Email`, `Name`, `Provider`).
  6. **User Identity Resolution:**
     - Query `UserStore` or `AuthzStore` by subject `google:<SocialUser.ID>` or email `SocialUser.Email`.
     - If user does not exist, provision new user using `UserRegistrar` or `AuthzStore.RegisterUser` with subject `google:<SocialUser.ID>`.
  7. **Flow Completion:**
     - If initiated as part of an OAuth 2.0 authorization code flow (or direct login), issue an authorization code or session and redirect the user back to the OAuth client's redirect URI.

#### 3. `GET /auth/social/github`
- **Handler:** `authzHandlers.SocialGitHubInitiate(w http.ResponseWriter, r *http.Request)`
- **Behavior:**
  1. Check if `h.GitHubConfig.ClientID == ""`. Return `501 Not Implemented` if unconfigured.
  2. Generate random 32-byte state token.
  3. Set `dilithiauth_social_state_github` cookie (`HttpOnly`, `SameSite=Lax`, 10-min expiration).
  4. Generate GitHub authorization URL via `oauth.GetSocialAuthURL(oauth.ProviderGitHub, h.GitHubConfig, state)`.
  5. Redirect user to GitHub via `http.Redirect(w, r, authURL, http.StatusFound)`.

#### 4. `GET /auth/social/github/callback`
- **Handler:** `authzHandlers.SocialGitHubCallback(w http.ResponseWriter, r *http.Request)`
- **Behavior:**
  1. Validate state parameter against `dilithiauth_social_state_github` cookie.
  2. Clear state cookie.
  3. Extract authorization `code`.
  4. Call `oauth.ExchangeSocialCode(r.Context(), oauth.ProviderGitHub, h.GitHubConfig, code)` to fetch `SocialUser`.
  5. Resolve/provision user identity in `UserStore`/`AuthzStore` with subject `github:<SocialUser.ID>`.
  6. Issue code/token and complete authentication flow.

---

## 5. Verification Plan

1. **Static Analysis & Build Verification:**
   - Execute `go build ./cmd/server` to ensure no syntax errors, missing fields, or unused imports exist.
2. **Unit Test Verification:**
   - Run `go test ./pkg/oauth/...` to verify `GetSocialAuthURL` and `ExchangeSocialCode` tests pass.
3. **HTTP Integration Verification:**
   - Test server startup without social auth env vars -> verify log shows "disabled".
   - Test server startup with `DILITHIAUTH_GOOGLE_CLIENT_ID` set -> verify log shows "enabled".
   - Perform HTTP `GET /auth/social/google` -> verify redirect (302) to `https://accounts.google.com/o/oauth2/v2/auth` with valid parameters and state cookie header set.
