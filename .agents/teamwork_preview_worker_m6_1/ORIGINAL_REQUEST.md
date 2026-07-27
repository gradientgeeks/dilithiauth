## 2026-07-27T18:13:29Z
You are teamwork_preview_worker_m6_1. Your working directory is /home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m6_1.

Objective:
Implement live Social Authentication HTTP endpoints (`/auth/social/google`, `/auth/social/google/callback`, `/auth/social/github`, `/auth/social/github/callback`), wire environment variables (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`), update systemd service docs, write unit tests, and verify tests and build.

Read the Explorer reports for guidance:
- `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_1/handoff.md`
- `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_2/handoff.md`
- `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_3/handoff.md`

Tasks:
1. Update `pkg/oauth/social.go`:
   - Add `TokenURL string` and `UserInfoURL string` fields to `SocialConfig` (fall back to Google/GitHub standard URLs if empty).
   - Update `ExchangeSocialCode` to use `cfg.TokenURL` and `cfg.UserInfoURL`. Also ensure GitHub email retrieval handles private emails if main profile email is empty (fetch from `/user/emails`).

2. Update `pkg/oauth/authz_handlers.go`:
   - Add `GoogleConfig SocialConfig` and `GitHubConfig SocialConfig` fields to `AuthzHandlers`.
   - Implement `SocialLoginGET(w http.ResponseWriter, r *http.Request)` handler for `/auth/social/{provider}`:
     - Extract `{provider}` via `chi.URLParam(r, "provider")`.
     - Validate incoming parameters (`client_id`, `redirect_uri`, `scope`, `state` [client state], `code_challenge`, `code_challenge_method`).
     - Retrieve provider configuration (`h.GoogleConfig` or `h.GitHubConfig`).
     - If `ClientID` or `ClientSecret` is unconfigured, return HTTP 400 Bad Request with clear error message.
     - Generate a signed state JWT using `h.ConsentSecret` containing client_id, redirect_uri, client_state, scope, code_challenge, code_challenge_method, provider, nonce, exp (10 mins).
     - Obtain provider authorization URL via `GetSocialAuthURL(provider, cfg, signedState)` and redirect (HTTP 302) to provider.
   - Implement `SocialCallbackGET(w http.ResponseWriter, r *http.Request)` handler for `/auth/social/{provider}/callback`:
     - Extract `{provider}`.
     - Extract query parameters `state` and `code`.
     - Validate state JWT signature using `h.ConsentSecret`, check expiration, and extract claims.
     - Call `ExchangeSocialCode(ctx, provider, cfg, code)` to get `SocialUser` identity claims.
     - Look up user by email in `h.UserStore` (if PostgresUserStore) or `h.Store` (MemoryAuthzStore).
     - If user does not exist, provision user with a new `Subject` (`usr_<16hex_rand>`), email, and name.
     - Handoff to `h.finishAuthorization(w, r, client_id, redirect_uri, scope, client_state, code_challenge, code_challenge_method, userSubject)` to issue an IdP code and redirect HTTP 302 back to client `redirect_uri`.

3. Update `cmd/server/main.go`:
   - Parse `DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GOOGLE_REDIRECT_URI` using `getEnv`.
   - Parse `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_REDIRECT_URI` using `getEnv`.
   - Set default redirect URIs to `issuer + "/auth/social/google/callback"` and `issuer + "/auth/social/github/callback"` if empty.
   - Inject `GoogleConfig` and `GitHubConfig` into `oauth.AuthzHandlers`.
   - Register routes under tenant router group:
     - `r.Get("/auth/social/{provider}", authzHandlers.SocialLoginGET)`
     - `r.Get("/auth/social/{provider}/callback", authzHandlers.SocialCallbackGET)`

4. Deployment & Service Docs:
   - Create `deployments/dilithiauth.service` systemd unit template file with placeholders for `DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`.
   - Update `docs/oci-vm-deployment.md` systemd service snippet.
   - Update `docs/deployment.md` environment variables reference table.

5. Unit Testing:
   - Add unit tests in `pkg/oauth/social_test.go` or `pkg/oauth/social_auth_test.go`:
     - Test `SocialLoginGET` (redirects with signed state, handles unconfigured credentials).
     - Test `SocialCallbackGET` using `httptest.NewServer` for mock Google and GitHub OAuth token/userinfo endpoints (verifies state JWT, code exchange, user lookup/provisioning, finishAuthorization, 302 redirect with IdP authorization code).

6. Verification:
   - Execute `go test ./...` in project root and record results in handoff report.
   - Execute `go build -o dilithiauth-server ./cmd/server` and record results.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Write your changes to `changes.md` and handoff report to `handoff.md` in `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m6_1/`.
Send a completion message to the parent orchestrator when done.
