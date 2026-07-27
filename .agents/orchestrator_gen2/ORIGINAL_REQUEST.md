# Original User Request

## Follow-up — 2026-07-27T18:10:30Z

<USER_REQUEST>
Implement live Social Authentication HTTP endpoints (`/auth/social/google`, `/auth/social/google/callback`, `/auth/social/github`, `/auth/social/github/callback`) in the DilithiAuth server (`pkg/oauth/` and `cmd/server/main.go`), wire environment variables (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`), and update production systemd service docs (`docs/oci-vm-deployment.md`).

Working directory: `/home/uttam/projects/DilithiAuth`
Integrity mode: `development`

## Requirements

### R1. Social Login HTTP Handlers & Callback Flow
Implement HTTP handlers in `pkg/oauth/authz_handlers.go` and `cmd/server/main.go` for `/auth/social/{provider}` and `/auth/social/{provider}/callback`:
- State validation via encrypted state cookies or session parameters to prevent CSRF attacks.
- Exchange authorization code with Google (`https://oauth2.googleapis.com/token`) and GitHub (`https://github.com/login/oauth/access_token`).
- Normalize profile claims (`sub`, `email`, `name`).
- Provision or retrieve user account in the DilithiAuth userstore.
- Issue Post-Quantum Hybrid Access Token (Ed25519 + ML-DSA-44) and redirect to client `redirect_uri` with authorization code.

### R2. Environment Variable Wiring
Read Google and GitHub client credentials from environment variables (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`) in `cmd/server/main.go` and pass to social handlers.

### R3. Deployment Environment & Service Config Update
Update `docs/oci-vm-deployment.md` and `deployments/dilithiauth.service` systemd configuration template with Google and GitHub OAuth environment variable placeholders.

## Acceptance Criteria

### Verification & Functionality
- [ ] Server compiles cleanly (`go build -o dilithiauth-server ./cmd/server`).
- [ ] All Go unit tests pass (`go test ./...` in root).
- [ ] Social auth endpoints (`/auth/social/google`, `/auth/social/github`) respond with proper 302 redirects when credentials are configured.
</USER_REQUEST>
