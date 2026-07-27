# DilithiAuth Social Authentication Deployment, Documentation & Testing Analysis

## 1. Executive Summary

This report documents the findings and proposed changes for integrating Social Authentication (Google & GitHub OAuth 2.0) into DilithiAuth's deployment configuration, system documentation, and unit test suite.

---

## 2. Deployment & Systemd Service Investigation

### 2.1 Existing State
- **Systemd Unit Location**: `/etc/systemd/system/dilithiauth.service` on the OCI VM (`80.225.222.81`).
- **Current Version Control Status**: No committed systemd service template exists under `deployments/` (currently contains only `certs/`, `nginx.conf`, and `sdk-ci-cd.yml`).
- **Live VM Environment Variables**:
  - `docs/oci-vm-deployment.md` documents the live systemd unit configuration.
  - Currently missing environment variables for social authentication (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`).

### 2.2 Systemd Service Template Additions (`deployments/dilithiauth.service`)
To ensure version control alignment across deployments, a new file `deployments/dilithiauth.service` should be created with the following configuration:

```ini
[Unit]
Description=DilithiAuth Post-Quantum Identity Provider
After=network.target postgresql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/dilithiauth
ExecStart=/opt/dilithiauth/server
Restart=always
RestartSec=3

Environment="DILITHIAUTH_ISSUER=https://dilithiauthapis.gradientgeeks.tech"
Environment="DILITHIAUTH_ADDR=:8080"
Environment="DILITHIAUTH_KEYS_DIR=/opt/dilithiauth/keys"
Environment="DILITHIAUTH_POSTGRES_DSN=postgres://dilithiauth:CHANGE_ME@127.0.0.1:5432/dilithiauth?sslmode=disable"
Environment="DILITHIAUTH_SIGSTORE=redis"
Environment="DILITHIAUTH_REDIS_ADDR=127.0.0.1:6380"
Environment="DILITHIAUTH_REDIS_USER=dilithiauth"
Environment="DILITHIAUTH_REDIS_PASS=CHANGE_ME"
Environment="DILITHIAUTH_FRONTEND_URL=https://dilithiauth.gradientgeeks.tech"

# Social Authentication Credentials (Google & GitHub OAuth 2.0)
Environment="DILITHIAUTH_GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com"
Environment="DILITHIAUTH_GOOGLE_CLIENT_SECRET=your-google-client-secret"
Environment="DILITHIAUTH_GITHUB_CLIENT_ID=your-github-client-id"
Environment="DILITHIAUTH_GITHUB_CLIENT_SECRET=your-github-client-secret"

# OCI SMTP Settings
Environment="OCI_SMTP_HOST=smtp.email.ap-mumbai-1.oci.oraclecloud.com"
Environment="OCI_SMTP_PORT=587"
Environment="OCI_SMTP_USER=CHANGE_ME"
Environment="OCI_SMTP_PASS=CHANGE_ME"
Environment="OCI_SMTP_FROM=no-reply@gradientgeeks.tech"

[Install]
WantedBy=multi-user.target
```

### 2.3 Documentation Updates (`docs/oci-vm-deployment.md` & `docs/deployment.md`)

#### A. Updates to `docs/oci-vm-deployment.md`
In Section `### systemd service` (lines 82–96):
Append the redacted social auth environment lines:
```ini
Environment="DILITHIAUTH_GOOGLE_CLIENT_ID=***"
Environment="DILITHIAUTH_GOOGLE_CLIENT_SECRET=***"
Environment="DILITHIAUTH_GITHUB_CLIENT_ID=***"
Environment="DILITHIAUTH_GITHUB_CLIENT_SECRET=***"
```

#### B. Updates to `docs/deployment.md`
Add the following entries to the **Environment Variables** reference table:

| Variable | Default | Description |
|----------|---------|-------------|
| `DILITHIAUTH_GOOGLE_CLIENT_ID` | *(empty)* | Google OAuth 2.0 Web Application Client ID for Social Login |
| `DILITHIAUTH_GOOGLE_CLIENT_SECRET` | *(empty)* | Google OAuth 2.0 Web Application Client Secret |
| `DILITHIAUTH_GITHUB_CLIENT_ID` | *(empty)* | GitHub OAuth 2.0 App Client ID for Social Login |
| `DILITHIAUTH_GITHUB_CLIENT_SECRET` | *(empty)* | GitHub OAuth 2.0 App Client Secret |

Add a new section `## Social Authentication Configuration`:
- Cross-reference `social-login-integration.md` for step-by-step setup guides for Google Cloud Console and GitHub Developer Settings.
- Detail the authorized callback URLs:
  - Google: `https://dilithiauthapis.gradientgeeks.tech/auth/social/google/callback`
  - GitHub: `https://dilithiauthapis.gradientgeeks.tech/auth/social/github/callback`
- Note fallback behavior: when credentials are left empty/unset, standard username/password and TOTP 2FA remain active, and social login buttons gracefully inform users that the provider is unconfigured.

---

## 3. Examination of Existing Unit Test Infrastructure

### 3.1 Survey of `pkg/oauth/*_test.go`
1. **`pkg/oauth/social_test.go`**:
   - Contains `TestGetSocialAuthURL` testing URL construction for Google and GitHub.
   - Missing tests for `ExchangeSocialCode` and callback HTTP handlers due to external HTTP client calls.
2. **`pkg/oauth/oauth_test.go`**:
   - Exercises discovery documents, JWKS generation, JWT issuance (thin and full jumbo modes), CWT binary size optimization, and KeyManager key rotation.
   - Uses helper functions (`newProvider`, `newKeyManagerProvider`) and `httptest.NewRecorder` / `httptest.NewRequest`.
3. **`pkg/oauth/authz_test.go`**:
   - Exercises user authentication, client validation, authorization code creation and redemption, consent handling, password reset, and token exchange.
   - Uses helper functions (`newHandlers`, `newHandlersWithConsent`, `injectTenant`).
4. **Project-wide Patterns (`sdk/go/client_test.go`, `tests/e2e_oauth_test.go`)**:
   - Standalone HTTP servers use `httptest.NewServer(handler)`.
   - Assertions rely exclusively on standard Go `t.Fatalf`, `t.Errorf`, `t.Error`, `t.Logf` (no 3rd-party assertion packages like `testify`).

---

## 4. Unit Testing Strategy for Social Auth Endpoints

### 4.1 Refactoring `pkg/oauth/social.go` for Testability
Currently `ExchangeSocialCode` creates an internal `&http.Client{}` and hardcodes endpoint URLs (`https://oauth2.googleapis.com/token`, `https://www.googleapis.com/oauth2/v2/userinfo`, `https://github.com/login/oauth/access_token`, `https://api.github.com/user`).

To enable unit testing against `httptest.Server`:
1. Enhance `SocialConfig` to accept optional custom endpoint URLs and an optional `*http.Client`:
```go
type SocialConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	TokenURL     string       // Optional override for testing (defaults to Google/GitHub token URL)
	UserInfoURL  string       // Optional override for testing (defaults to Google/GitHub userinfo URL)
	HTTPClient   *http.Client // Optional HTTP client override (defaults to &http.Client{})
}
```

### 4.2 Mock HTTP Servers Design (`httptest.Server`)

#### A. Mock Google OAuth Server
Handles two endpoints:
- `POST /token`:
  - Validates `Content-Type: application/x-www-form-urlencoded`.
  - Checks parameters: `grant_type=authorization_code`, `code`, `client_id`, `client_secret`, `redirect_uri`.
  - Returns `200 OK` with JSON:
    ```json
    {
      "access_token": "mock_google_access_token_123",
      "token_type": "Bearer",
      "expires_in": 3600
    }
    ```
  - Returns `400 Bad Request` if `code` equals `"invalid_code"`.
- `GET /userinfo`:
  - Validates header `Authorization: Bearer mock_google_access_token_123`.
  - Returns `200 OK` with JSON:
    ```json
    {
      "id": "google_user_10987654321",
      "email": "alice.google@example.com",
      "name": "Alice Google",
      "verified_email": true
    }
    ```
  - Returns `401 Unauthorized` if authorization header is invalid.

#### B. Mock GitHub OAuth Server
Handles two endpoints:
- `POST /login/oauth/access_token`:
  - Validates `Content-Type: application/x-www-form-urlencoded` and `Accept: application/json`.
  - Checks parameters: `code`, `client_id`, `client_secret`, `redirect_uri`.
  - Returns `200 OK` with JSON:
    ```json
    {
      "access_token": "mock_github_access_token_456",
      "token_type": "bearer",
      "scope": "read:user,user:email"
    }
    ```
  - Returns `400 Bad Request` if `code` equals `"invalid_code"`.
- `GET /user`:
  - Validates `Authorization: Bearer mock_github_access_token_456` and `User-Agent: DilithiAuth-Server`.
  - Returns `200 OK` with JSON:
    ```json
    {
      "id": 12345678,
      "login": "alicegithub",
      "name": "Alice GitHub",
      "email": "alice.github@example.com"
    }
    ```

### 4.3 Proposed Unit Test Cases in `pkg/oauth/social_test.go`

1. **`TestExchangeSocialCode_Google_Success`**:
   - Set up `mockGoogleServer`.
   - Pass `SocialConfig` with `TokenURL = mockServer.URL + "/token"` and `UserInfoURL = mockServer.URL + "/userinfo"`.
   - Assert `SocialUser.ID == "google_user_10987654321"`, `SocialUser.Email == "alice.google@example.com"`, `SocialUser.Provider == ProviderGoogle`.

2. **`TestExchangeSocialCode_Google_TokenError`**:
   - Pass `code = "invalid_code"`.
   - Assert error is non-nil and contains status code `400`.

3. **`TestExchangeSocialCode_Google_UserInfoError`**:
   - Configure mock server to return `401` on `/userinfo`.
   - Assert error is returned.

4. **`TestExchangeSocialCode_GitHub_Success`**:
   - Set up `mockGitHubServer`.
   - Pass `SocialConfig` with `TokenURL = mockServer.URL + "/login/oauth/access_token"` and `UserInfoURL = mockServer.URL + "/user"`.
   - Assert `SocialUser.ID == "12345678"`, `SocialUser.Email == "alice.github@example.com"`, `SocialUser.Provider == ProviderGitHub`.

5. **`TestExchangeSocialCode_GitHub_TokenError`**:
   - Pass `code = "invalid_code"`.
   - Assert error is non-nil.

6. **`TestExchangeSocialCode_UnsupportedProvider`**:
   - Pass invalid provider `SocialProvider("twitter")`.
   - Assert error returned immediately without network call.

---

## 5. Summary of Recommended Action Items

1. **Create `deployments/dilithiauth.service`**: Add the complete, production-ready systemd unit template.
2. **Update `docs/oci-vm-deployment.md`**: Add redacted social auth environment variables to systemd documentation block.
3. **Update `docs/deployment.md`**: Add social auth variables table and configuration notes.
4. **Update `pkg/oauth/social.go`**: Add `TokenURL` and `UserInfoURL` optional overrides to `SocialConfig`.
5. **Implement social auth tests in `pkg/oauth/social_test.go`**: Use `httptest.NewServer` to test Google & GitHub token and userinfo exchange flows.
