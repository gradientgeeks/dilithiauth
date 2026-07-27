# Handoff Report: Social Authentication Integration Architecture & Plan

**Author:** `teamwork_preview_explorer_m6_1`  
**Date:** 2026-07-27  
**Working Directory:** `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_1`  

---

## 1. Observation

Direct observations from the DilithiAuth codebase investigation:

1. **Existing Social Auth Helper Functions (`pkg/oauth/social.go`)**:
   - `SocialProvider` constants: `ProviderGoogle = "google"`, `ProviderGitHub = "github"` (`pkg/oauth/social.go:17-18`).
   - `SocialUser` struct (`pkg/oauth/social.go:22-27`):
     ```go
     type SocialUser struct {
         ID       string         `json:"id"`
         Email    string         `json:"email"`
         Name     string         `json:"name"`
         Provider SocialProvider `json:"provider"`
     }
     ```
   - `GetSocialAuthURL` constructs authorization URLs for Google (`https://accounts.google.com/o/oauth2/v2/auth`) and GitHub (`https://github.com/login/oauth/authorize`) (`pkg/oauth/social.go:37-63`).
   - `ExchangeSocialCode` exchanges authorization codes for access tokens and queries profile endpoints (`https://www.googleapis.com/oauth2/v2/userinfo` and `https://api.github.com/user`) (`pkg/oauth/social.go:66-166`).

2. **Frontend UI Buttons in Login Template (`pkg/oauth/authz_handlers.go:457-465`)**:
   - The login page HTML template (`loginTmpl`) embeds social login buttons targeting `/auth/social/google` and `/auth/social/github`:
     ```html
     <a href="/auth/social/google?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" ...>Google</a>
     <a href="/auth/social/github?client_id={{.ClientID}}&redirect_uri={{.RedirectURI}}&state={{.State}}" ...>GitHub</a>
     ```

3. **HTTP Server Router Configuration (`cmd/server/main.go:440-476`)**:
   - Chi router registers `/authorize`, `/register`, `/forgot-password`, `/reset-password`, `/consent`, `/token`, `/revoke`, `/verify/jwt`.
   - **Missing Routes:** `/auth/social/{provider}` and `/auth/social/{provider}/callback` are **not** currently registered on the Chi router in `cmd/server/main.go`.

4. **User Store and Account Management (`pkg/userstore/userstore.go` & `pkg/oauth/authz.go`)**:
   - Production account store: `PostgresUserStore` (`pkg/userstore/userstore.go:59-257`) with `GetByEmail`, `RegisterForSubject`, `AuthenticateForSubject`.
   - Fallback account store: `MemoryAuthzStore` (`pkg/oauth/authz.go:94-220`) with `GetUserByEmail`, `RegisterUser`.
   - `AuthzHandlers` (`pkg/oauth/authz_handlers.go:97-113`) contains `UserStore`, `UserRegistrar`, `Store`, `Provider`, and `ConsentSecret`.

5. **Token Generation (`pkg/oauth/oauth.go:184-240`)**:
   - Mints PQC Thin JWTs via `IssueJWTThinWithScope(tenantID, subject, scope, ttl)` combining Ed25519 signature + SHA-256 hash reference (`pqc_ref`) pointing to full ML-DSA-44 post-quantum signature in `SigStore`.

6. **Test Verification (`go test ./pkg/oauth/...`)**:
   - Executed `go test ./pkg/oauth/...` -> PASS (`ok github.com/Uttam-Mahata/DilithiAuth/pkg/oauth`).

---

## 2. Logic Chain

1. **From Observation 2 & 3:** The user interface in `loginTmpl` directs users to `/auth/social/{provider}`, but the server router in `cmd/server/main.go` does not mount handlers for `/auth/social/{provider}` or `/auth/social/{provider}/callback`. Clicking social sign-in buttons currently results in an HTTP 404 Not Found error.
2. **From Observation 1 & 5:** The underlying functions `GetSocialAuthURL` and `ExchangeSocialCode` in `social.go` perform provider URL formatting and OAuth code exchange, but they are disconnected from DilithiAuth's token minting pipeline (`IssueJWTThinWithScope`) and authorization code generator (`Store.CreateCode`).
3. **From Observation 1, 4 & 5:** To complete social authentication safely:
   - DilithiAuth must validate client credentials, issue a signed state JWT to prevent CSRF, and redirect the user to Google/GitHub.
   - Upon receiving the callback at `/auth/social/{provider}/callback`, DilithiAuth must verify the state JWT, exchange the social code for profile claims (`SocialUser`), look up or provision the user in `UserStore`/`AuthzStore`, and invoke `finishAuthorization` to issue an IdP authorization code.
   - The client application then exchanges this code at `POST /token` to receive DilithiAuth's Post-Quantum Hybrid Access Token (Ed25519 + ML-DSA-44 signature).

---

## 3. Caveats

1. **Network Restrictions:** Investigation was executed in `CODE_ONLY` mode. Live HTTP calls to `accounts.google.com` or `github.com` were not made. Unit tests for `ExchangeSocialCode` require mock HTTP test servers (`httptest.NewServer`).
2. **Provider Credentials:** Actual social authentication requires setting provider `client_id` and `client_secret` in environment variables (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, etc.).
3. **Private Email Scope on GitHub:** GitHub users may set their email addresses to private. `ExchangeSocialCode` queries `GET https://api.github.com/user`. If `email` is null, a fallback request to `GET https://api.github.com/user/emails` is recommended in implementation.

---

## 4. Conclusion

Social Authentication (Google & GitHub) can be seamlessly integrated into DilithiAuth by adding `SocialLoginGET` and `SocialCallbackGET` HTTP handlers to `AuthzHandlers` in `pkg/oauth/social.go` (or `pkg/oauth/authz_handlers.go`), wiring them to `cmd/server/main.go`, and bridging the social identity claims into `finishAuthorization`. This maintains 100% compatibility with DilithiAuth's multi-tenant scoping and Post-Quantum Hybrid Token (ML-DSA-44 + Ed25519) architecture.

---

## 5. Verification Method

To verify the implementation once coded:

1. **Unit Tests**:
   - Run `go test -v ./pkg/oauth -run TestSocialAuth`
   - Verify state JWT generation, expiration, and signature validation.
   - Use `httptest.NewServer` to mock Google and GitHub OAuth token & userinfo endpoints, verifying `ExchangeSocialCode` parses responses correctly.

2. **Integration Verification**:
   - Run server: `go run ./cmd/server`
   - Send `GET /auth/social/google?client_id=local-rootaccess&redirect_uri=http://localhost:8080/api/auth/dilithiauth/callback&state=xyz`
   - Confirm HTTP 302 response targeting `https://accounts.google.com/o/oauth2/v2/auth` with signed `state` parameter.
   - Simulate callback `GET /auth/social/google/callback?code=mock_code&state=<signed_state>`
   - Confirm HTTP 302 redirect to `http://localhost:8080/api/auth/dilithiauth/callback?code=<idp_code>&state=xyz`.
   - Exchange `<idp_code>` at `POST /token` and verify returned JWT contains valid `pqc_ref` and Ed25519 signature.
