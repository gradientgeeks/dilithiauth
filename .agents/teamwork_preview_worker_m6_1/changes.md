# Summary of Code Changes

## 1. `pkg/oauth/social.go`
- Added `TokenURL` and `UserInfoURL` optional fields to `SocialConfig` to support custom endpoints during unit testing or proxying.
- Updated `ExchangeSocialCode`:
  - Uses `cfg.TokenURL` and `cfg.UserInfoURL` with fallbacks to standard Google/GitHub endpoints if unconfigured.
  - Added private email fallback for GitHub (`/user/emails`) when primary profile email is null or empty.
  - Enhanced name resolution for GitHub profiles to fall back to `login` handle if full name is empty.

## 2. `pkg/oauth/authz_handlers.go`
- Added `GoogleConfig SocialConfig` and `GitHubConfig SocialConfig` fields to `AuthzHandlers`.
- Implemented `socialStateClaims` struct and `signSocialState` / `parseSocialState` helpers using HS256 and `ConsentSecret` (with fallback).
- Implemented `SocialLoginGET(w, r)`:
  - Validates `client_id` and `redirect_uri` via `ClientValidator`.
  - Checks provider credentials (`GoogleConfig` or `GitHubConfig`); returns HTTP 400 Bad Request if unconfigured.
  - Generates 10-minute signed state JWT carrying `client_id`, `redirect_uri`, `client_state`, `scope`, PKCE parameters, `provider`, and nonce.
  - Redirects (HTTP 302) to provider's authorization URL.
- Implemented `SocialCallbackGET(w, r)`:
  - Validates signed state JWT signature and expiration.
  - Exchanges provider code via `ExchangeSocialCode` for `SocialUser` identity.
  - Queries `UserStore` (Postgres) or `Store` (Memory) for existing user by email.
  - If user is not found, provisions new user identity with subject `usr_<16hex_rand>`, email, and name.
  - Calls `finishAuthorization` to issue IdP authorization code and redirect HTTP 302 to client's `redirect_uri`.

## 3. `cmd/server/main.go`
- Added environment variable parsing for social provider credentials:
  - `DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GOOGLE_REDIRECT_URI`
  - `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_REDIRECT_URI`
- Injected `GoogleConfig` and `GitHubConfig` into `oauth.AuthzHandlers`.
- Registered social auth routes under tenant router group:
  - `GET /auth/social/{provider}`
  - `GET /auth/social/{provider}/callback`

## 4. Deployments & Documentation
- Created `deployments/dilithiauth.service` systemd unit template with social auth environment variable placeholders.
- Updated `docs/oci-vm-deployment.md` systemd service snippet.
- Updated `docs/deployment.md` environment variables reference table and API endpoints table.

## 5. Unit Tests
- Updated `pkg/oauth/social_test.go` with tests:
  - `TestGetSocialAuthURL`
  - `TestExchangeSocialCode_Google_Success`
  - `TestExchangeSocialCode_GitHub_PrivateEmail`
  - `TestSocialLoginGET`
  - `TestSocialCallbackGET`
