# Handoff Report: Social Authentication Server Route & Environment Configuration

**Agent:** teamwork_preview_explorer_m6_2  
**Working Directory:** `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_2`  
**Date:** 2026-07-27  

---

## 1. Observation

Direct observations from the codebase investigation:

1. **`cmd/server/main.go` Entrypoint and Routing:**
   - **Router Initialization:** `r := chi.NewRouter()` at line 341.
   - **Global Middleware:** Lines 345-350 apply `middleware.RequestID`, `newCORSMiddleware(allowedOrigins)`, `middleware.Logger`, `middleware.Recoverer`, `transport.JumboHeaderMiddleware`, and `transport.SigStoreMiddleware(sigStore)`.
   - **Tenant Scoped Group:** Lines 438-440 wrap OAuth/OIDC endpoints with `r.Group(func(r chi.Router) { r.Use(tenant.NewMiddleware(trustedTenantCIDRs))` where tenant resolution is handled automatically.
   - **Handler Instantiation:** `authzHandlers` is initialized as a pointer struct `&oauth.AuthzHandlers{...}` at lines 298-323.

2. **Environment Variable Parsing Pattern:**
   - Lines 694-699 define `getEnv(key, fallback string) string`:
     ```go
     func getEnv(key, fallback string) string {
         if v := os.Getenv(key); v != "" {
             return v
         }
         return fallback
     }
     ```
   - Used throughout `main.go` for `DILITHIAUTH_ISSUER` (line 89), `DILITHIAUTH_ADDR` (line 90), `DILITHIAUTH_SIGSTORE` (line 104), `DILITHIAUTH_POSTGRES_DSN` (line 116), etc.

3. **Existing Social Authentication Primitives:**
   - `pkg/oauth/social.go` lines 13-34 define `SocialProvider` (`ProviderGoogle = "google"`, `ProviderGitHub = "github"`), `SocialUser`, and `SocialConfig`:
     ```go
     type SocialConfig struct {
         ClientID     string
         ClientSecret string
         RedirectURI  string
     }
     ```
   - `pkg/oauth/social.go` lines 37-63 implement `GetSocialAuthURL(provider SocialProvider, cfg SocialConfig, state string) (string, error)`.
   - `pkg/oauth/social.go` lines 66-166 implement `ExchangeSocialCode(ctx context.Context, provider SocialProvider, cfg SocialConfig, code string) (*SocialUser, error)`.

4. **`AuthzHandlers` Struct Definition:**
   - `pkg/oauth/authz_handlers.go` lines 97-112 define `AuthzHandlers` struct holding dependencies (`Provider`, `Store`, `UserStore`, `UserRegistrar`, `PasswordResetter`, `OTPStore`, `SigStore`, `Revocation`, `ClientValidator`, `Verifier`, `RefreshTokens`, `GrantStore`, `ConsentSecret`).

---

## 2. Logic Chain

1. **Observation 1 & 3** show that `pkg/oauth/social.go` contains the core functions for constructing social auth URLs and exchanging codes, but `cmd/server/main.go` currently lacks env variable parsing for social provider credentials and route registration for social endpoints.
2. **Observation 2** establishes that all server configuration parameters follow the `getEnv("DILITHIAUTH_<NAME>", fallback)` pattern.
3. Therefore, adding `DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GOOGLE_REDIRECT_URI`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`, and `DILITHIAUTH_GITHUB_REDIRECT_URI` using `getEnv` in `cmd/server/main.go` maintains project consistency.
4. **Observation 4** shows that HTTP handlers for authorization code flows are methods on `*oauth.AuthzHandlers`. Extending `AuthzHandlers` to include `GoogleConfig SocialConfig` and `GitHubConfig SocialConfig` allows social authentication state and credentials to be seamlessly injected at server startup.
5. **Observation 1** shows that tenant isolation is enforced by registering OAuth endpoints inside the `r.Group(func(r chi.Router) { r.Use(tenant.NewMiddleware(trustedTenantCIDRs)) ... })` block. Registering `/auth/social/google`, `/auth/social/google/callback`, `/auth/social/github`, and `/auth/social/github/callback` inside this group ensures all social auth requests receive appropriate tenant context.

---

## 3. Caveats

- **External OAuth Provider Connectivity:** Live testing of `ExchangeSocialCode` requires actual client secrets and network connectivity to `oauth2.googleapis.com` and `github.com`. In test/offline environments, HTTP mock servers or stub handlers should be used.
- **Session/State Cookie Management:** The callback state verification requires setting a short-lived HTTP-only state cookie during the initiate GET request and validating it in the callback handler to prevent CSRF attacks.

---

## 4. Conclusion

1. **Environment Variables to Add:**
   - `DILITHIAUTH_GOOGLE_CLIENT_ID`
   - `DILITHIAUTH_GOOGLE_CLIENT_SECRET`
   - `DILITHIAUTH_GOOGLE_REDIRECT_URI` (default: `issuer + "/auth/social/google/callback"`)
   - `DILITHIAUTH_GITHUB_CLIENT_ID`
   - `DILITHIAUTH_GITHUB_CLIENT_SECRET`
   - `DILITHIAUTH_GITHUB_REDIRECT_URI` (default: `issuer + "/auth/social/github/callback"`)

2. **Struct Updates:**
   - Add `GoogleConfig SocialConfig` and `GitHubConfig SocialConfig` to `oauth.AuthzHandlers` in `pkg/oauth/authz_handlers.go`.

3. **HTTP Route Definitions:**
   - `GET /auth/social/google` -> Initiates Google OAuth2 redirect flow.
   - `GET /auth/social/google/callback` -> Processes Google auth code, resolves/provisions identity, and completes authentication.
   - `GET /auth/social/github` -> Initiates GitHub OAuth2 redirect flow.
   - `GET /auth/social/github/callback` -> Processes GitHub auth code, resolves/provisions identity, and completes authentication.

---

## 5. Verification Method

1. **Build Verification:**
   ```bash
   go build ./cmd/server
   ```
   Ensures `cmd/server` compiles cleanly with no syntax errors.

2. **Package Unit Tests:**
   ```bash
   go test ./pkg/oauth/...
   ```
   Runs existing social authentication and authz handler tests.

3. **HTTP Endpoint Inspection:**
   Inspect routing definitions in `cmd/server/main.go` to confirm routes are registered under the tenant middleware group.
