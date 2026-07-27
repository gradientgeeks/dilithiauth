# Architectural Analysis: Social Authentication Integration in DilithiAuth

**Author:** `teamwork_preview_explorer_m6_1`  
**Date:** 2026-07-27  
**Scope:** `pkg/oauth/`, `cmd/server/`, `pkg/userstore/`, `pkg/signer/`, `pkg/transport/`  

---

## Executive Summary

DilithiAuth is a Post-Quantum Identity Provider (IdP) implementing the OAuth 2.0 / OIDC Authorization Code Flow. It mitigates the **"Harvest Now, Decrypt Later" (HNDL)** threat using hybrid signatures combining **Ed25519** (classical) and **ML-DSA-44** (NIST FIPS 204 post-quantum standard).

This analysis investigates the integration of **Social Authentication (Google & GitHub)** into DilithiAuth's existing codebase. While initial frontend UI templates and basic HTTP helper functions (`GetSocialAuthURL`, `ExchangeSocialCode`) exist in `pkg/oauth/social.go`, full HTTP handlers (`/auth/social/{provider}` and `/auth/social/{provider}/callback`) and route registrations in `cmd/server/main.go` are missing.

This document details the existing architecture, identifies structural gaps, and provides a complete architectural design and implementation plan to wire Social Authentication into DilithiAuth's Post-Quantum Hybrid Token issuance pipeline.

---

## 1. Codebase Architecture & Structural Review

### 1.1 Server Routing & Request Pipeline (`cmd/server/main.go`)
- **Router:** Built on `github.com/go-chi/chi/v5`.
- **Middleware Chain:**
  1. `middleware.RequestID` — Request correlation tracing.
  2. `newCORSMiddleware(allowedOrigins)` — Strict origin validation for credentialed CORS.
  3. `middleware.Logger` & `middleware.Recoverer` — Structured JSON logging and panic recovery.
  4. `transport.JumboHeaderMiddleware` — Configures HTTP server with `MaxHeaderBytes = 16 KiB` to support large PQC tokens.
  5. `transport.SigStoreMiddleware(sigStore)` — Attaches sidecar signature store for Thin Token PQC reference lookups.
  6. `tenant.NewMiddleware(trustedTenantCIDRs)` — Resolves tenant context (`tenantID`) from subdomain or trusted `X-Tenant-Id` header.
- **Endpoints Registered:**
  - OIDC Discovery: `GET /.well-known/openid-configuration`
  - JWKS Document: `GET /.well-known/jwks.json`
  - Authorization Code Flow: `GET /authorize`, `POST /authorize`, `GET /register`, `POST /register`, `GET /forgot-password`, `POST /forgot-password`, `GET /reset-password`, `POST /reset-password`, `POST /consent`, `POST /token`, `POST /revoke`, `POST /verify/jwt`.
- **Gap Identified:** No routes for `/auth/social/{provider}` or `/auth/social/{provider}/callback` are mounted on the Chi router in `cmd/server/main.go`.

### 1.2 AuthzHandlers Structure (`pkg/oauth/authz_handlers.go`)
`AuthzHandlers` coordinates authorization code creation, user authentication, registration, consent management, and token exchange:
- `Provider *Provider`: Mints JWT Thin (`IssueJWTThinWithScope`), JWT Full (`IssueJWTFullWithScope`), and CWT (`IssueCWTWithScope`).
- `Store AuthzStore`: Manages OAuth clients, authorization codes (`CreateCode`, `RedeemCode`), and fallback in-memory/Redis user records.
- `UserStore UserAuthenticator` & `UserRegistrar UserRegistrar`: Production Argon2id-hashed PostgreSQL user store (`*userstore.PostgresUserStore`).
- `ClientValidator ClientValidator`: Validates client ID and redirect URIs (`CompositeClientValidator`).
- `SigStore SigStorePutter`: Stores ML-DSA-44 post-quantum signatures keyed by SHA-256 reference (`pqc_ref`).
- `ConsentSecret []byte`: 32-byte secret used to sign short-lived (5 min) HS256 JWT tickets for secure multi-step redirects (e.g. consent screens).

### 1.3 Login UI UI Hooks (`pkg/oauth/authz_handlers.go` lines 457-465)
The server-rendered HTML login template (`loginTmpl`) already renders Social Authentication buttons:
```html
<div class="divider">or continue with</div>
<div style="display: flex; gap: 0.5rem; margin-bottom: 1rem;">
  <a href="/auth/social/google?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" ...>
    Google
  </a>
  <a href="/auth/social/github?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" ...>
    GitHub
  </a>
</div>
```
When a user clicks Google or GitHub on the sign-in page, the browser issues a `GET` request to `/auth/social/google` or `/auth/social/github` with the original client app's `client_id`, `redirect_uri`, and `state`.

### 1.4 Existing Social Auth Helpers (`pkg/oauth/social.go`)
`pkg/oauth/social.go` currently defines:
- `SocialProvider` ("google", "github")
- `SocialUser`: Normalized structure containing `ID`, `Email`, `Name`, `Provider`.
- `SocialConfig`: Holds `ClientID`, `ClientSecret`, `RedirectURI` (the IdP's callback URL).
- `GetSocialAuthURL(provider, cfg, state)`: Constructs OAuth authorization redirect URL for Google (`https://accounts.google.com/o/oauth2/v2/auth`) or GitHub (`https://github.com/login/oauth/authorize`).
- `ExchangeSocialCode(ctx, provider, cfg, code)`: Exchanges social authorization code for access token and fetches userinfo/user profile.

---

## 2. Analysis of Core Mechanisms

### 2.1 State Management & CSRF Prevention
During Social Login, the client application redirects the end-user to DilithiAuth, which in turn redirects the user to Google or GitHub. To prevent CSRF attacks, state tampering, and open redirect vulnerabilities, DilithiAuth must preserve the original client OAuth parameters (`client_id`, `redirect_uri`, `state`, `scope`, `code_challenge`, `code_challenge_method`) across external redirects.

**Recommended Approach — Cryptographically Signed State Token (JWT):**
Instead of relying on stateful session cookies (which can suffer from SameSite restrictions or cookie-loss across cross-site redirects), DilithiAuth can encode all request metadata into a short-lived (10 minute) HS256-signed JWT state token using `ConsentSecret` (or `SocialStateSecret`).

```
Client App ---> GET /auth/social/google?client_id=X&redirect_uri=Y&state=S
            ---> DilithiAuth mints HS256 stateJWT(client_id, redirect_uri, state, scope, PKCE, CSRF nonce)
            ---> HTTP 302 Redirect to Google with state=stateJWT
Google     ---> User authenticates
            ---> HTTP 302 Redirect to /auth/social/google/callback?code=G_CODE&state=stateJWT
DilithiAuth ---> Verifies stateJWT signature & exp -> extracts original client_id, redirect_uri, S
```

### 2.2 Account Provisioning & User Lookup Strategy
When a user authenticates via Google or GitHub, `ExchangeSocialCode` returns a `SocialUser` containing `ID`, `Email`, `Name`, and `Provider`.

**Lookup and Provisioning Sequence:**
1. **Tenant-Aware Email Lookup:** Search for an existing account with `email == SocialUser.Email` in the tenant's userstore (`UserStore` when configured, otherwise `AuthzStore`).
2. **Existing Account Matched:** If an account exists, retrieve its stable `Subject` (`usr_<16hex>`).
3. **New Account Provisioning:** If no account exists:
   - Generate a unique username (e.g., `google_<social_id>` or email-derived handle).
   - Generate an unguessable random password hash (since password authentication is bypassed for social logins).
   - Register the user via `UserRegistrar.RegisterForSubject(ctx, tenantID, username, email, randomPass)` or `AuthzStore.RegisterUser(tenantID, user)`.
   - Retrieve the newly created user's `Subject`.

### 2.3 Post-Quantum Hybrid Token Generation
DilithiAuth issues hybrid tokens signed by both Ed25519 and ML-DSA-44:
1. Once the user's `Subject` is retrieved or provisioned, the handler delegates to `h.finishAuthorization(...)` (or `completeAuthorization` if scope consent is enforced).
2. `finishAuthorization` generates an internal OAuth authorization code via `h.Store.CreateCode(tenantID, subject, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod)`.
3. IdP redirects the browser back to the client application's `redirect_uri` with `code` and the original client `state`.
4. Client application exchanges `code` at `POST /token`.
5. `TokenExchange` handler calls `h.Provider.IssueJWTThinWithScope(...)`:
   - Computes ML-DSA-44 PQC signature over claims.
   - Hashes ML-DSA-44 signature to form `pqc_ref` (SHA-256 base64url).
   - Stores ML-DSA-44 signature in `SigStore` keyed by `pqc_ref`.
   - Signs JWT with Ed25519 private key.
   - Returns token response to client application.

---

## 3. Implementation Plan & Architecture Design

### 3.1 Proposed Data Structures (`pkg/oauth/social.go` & `pkg/oauth/authz_handlers.go`)

```go
// SocialStateClaims represents the encrypted/signed state payload carried
// through external social IdP authorization redirects.
type SocialStateClaims struct {
	jwt.RegisteredClaims
	TenantID            string         `json:"tenant_id"`
	ClientID            string         `json:"client_id"`
	RedirectURI         string         `json:"redirect_uri"`
	ClientState         string         `json:"client_state"`
	Scope               string         `json:"scope"`
	CodeChallenge       string         `json:"code_challenge"`
	CodeChallengeMethod string         `json:"code_challenge_method"`
	Provider            SocialProvider `json:"provider"`
	Nonce               string         `json:"nonce"`
}
```

Add social configuration fields to `AuthzHandlers`:
```go
type AuthzHandlers struct {
	// ... existing fields ...
	SocialConfigs map[SocialProvider]SocialConfig // Map of provider -> SocialConfig
}
```

### 3.2 Handler 1: `SocialLoginGET` (`GET /auth/social/{provider}`)
**Responsibilities:**
1. Extract `{provider}` path param (google or github). Validate against `h.SocialConfigs`.
2. Extract client params: `client_id`, `redirect_uri`, `state`, `scope`, `code_challenge`, `code_challenge_method`.
3. Validate client ID and redirect URI via `h.clientValidator().ValidateClient(tenantID, clientID, redirectURI)`.
4. Generate signed state JWT (`SocialStateClaims` signed with `h.ConsentSecret`, 10 min TTL).
5. Call `GetSocialAuthURL(provider, cfg, stateJWT)`.
6. Issue HTTP 302 redirect to external provider authorization URL.

### 3.3 Handler 2: `SocialCallbackGET` (`GET /auth/social/{provider}/callback`)
**Responsibilities:**
1. Extract `code`, `state` (signed state JWT), `error` from callback query params.
2. If `error` parameter is present, redirect to client `redirect_uri?error=access_denied&state=<ClientState>`.
3. Parse and verify signed `state` JWT with `h.ConsentSecret`. Verify `exp` and signature.
4. Call `ExchangeSocialCode(r.Context(), provider, cfg, code)` to exchange social code for normalized `SocialUser`.
5. Perform account lookup or auto-provisioning in `UserStore`/`AuthzStore`.
6. Hand off to `h.finishAuthorization(w, r, tenantID, subject, clientID, redirectURI, clientState, scope, codeChallenge, codeChallengeMethod, onError)`.

### 3.4 Refactoring Helper Functions in `pkg/oauth/social.go`
To improve testability (preventing external HTTP calls in unit tests):
- Update `ExchangeSocialCode` (or introduce `ExchangeSocialCodeWithClient`) to accept an optional `HTTPClient interface` or `*http.Client`, allowing unit tests to inject mock HTTP transports (`httptest.Server`).

### 3.5 Router Integration in `cmd/server/main.go`
Mount the social auth handlers inside the tenant-aware router group:
```go
r.Get("/auth/social/{provider}", authzHandlers.SocialLoginGET)
r.Get("/auth/social/{provider}/callback", authzHandlers.SocialCallbackGET)
```

---

## 4. Verification & Testing Strategy

1. **Unit Tests (`pkg/oauth/social_test.go`):**
   - Test state token signing & verification (`signSocialState` / `parseSocialState`).
   - Mock HTTP server (`httptest.NewServer`) for Google (`/token` & `/userinfo`) and GitHub (`/access_token` & `/user`).
   - Verify claim normalization into `SocialUser`.
2. **Integration Tests (`pkg/oauth/authz_test.go` or `cmd/server/main_test.go`):**
   - End-to-end simulation of `/auth/social/google` -> signed state generation -> callback handling -> user account auto-provisioning in `PostgresUserStore` / `MemoryAuthzStore` -> code issuance -> `/token` exchange -> PQC Hybrid Access Token validation.
