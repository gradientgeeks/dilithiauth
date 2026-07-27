# Forensic Audit Report: Social Authentication (Milestone 6 Remediated)

**Work Product**: Social Authentication Implementation (`pkg/oauth/social.go`, `pkg/oauth/authz_handlers.go`, `cmd/server/main.go`, `deployments/dilithiauth.service`, `docs/oci-vm-deployment.md`, `docs/deployment.md`, `pkg/oauth/social_test.go`, `pkg/oauth/social_challenge_test.go`, `pkg/oauth/social_m6_3_empirical_test.go`)
**Profile**: General Project
**Verdict**: INTEGRITY VIOLATION

---

## Phase 1: Forensic Code Analysis

### 1. Hardcoded Output & Result Detection
- **`pkg/oauth/social.go`**: `GetSocialAuthURL` builds authorization redirect URLs dynamically using query parameters. `ExchangeSocialCode` performs dynamic HTTP POST/GET requests to exchange tokens and fetch user profiles. No hardcoded or static profile results exist.
- **`pkg/oauth/authz_handlers.go`**: `SocialLoginGET` generates signed HS256 JWT state tokens with 16-byte cryptographically secure random nonces (`rand.Read`). `SocialCallbackGET` dynamically parses state claims, verifies tenant ID and provider parity, provisions users with random `usr_<16hex_rand>` subjects, generates random 16-byte hex passwords, handles username collision fallbacks (`baseUsername_<4hex_rand>`), and issues genuine OAuth authorization codes.
- **`pkg/oauth/social_test.go` & `social_challenge_test.go`**: Tests construct real `httptest.Server` mock OAuth identity providers and make real assertions on response codes, state JWT claims, PKCE challenges, user provisioning, and error redirection.

### 2. Facade & Dummy Implementation Detection
- `ExchangeSocialCode` implements complete HTTP exchanges for Google and GitHub, including header handling, status code checking, JSON unmarshaling, and GitHub secondary email fetching (`/user/emails` with `verified` check).
- `SocialLoginGET` and `SocialCallbackGET` perform full client validation, tenant isolation check, user store registration, and code creation.
- No facade or dummy implementations were detected.

### 3. Pre-populated Artifact Inspection
- Workspace search confirmed no static log files, pre-generated state tokens, or mock response files exist in the repository.

---

## Phase 2: Behavioral & Build Verification

### 1. Binary Compilation
- **Command**: `go build -o dilithiauth-server ./cmd/server`
- **Result**: `PASS` — Compiled successfully (`dilithiauth-server`, size ~24MB).

### 2. Test Suite Execution (`go test ./...`)
- **Command**: `go test ./...`
- **Result**: `FAIL` — The package `pkg/oauth` failed due to 2 failing subtests in `TestM63Empirical_AllScenarios`:

```
--- FAIL: TestM63Empirical_AllScenarios (0.24s)
    --- FAIL: TestM63Empirical_AllScenarios/Subtest_A:_Social_Login_Redirects_&_Tenant_ID_in_State_JWT (0.00s)
        social_m6_3_empirical_test.go:117: Google Login: expected HTTP 302, got 400 (body: invalid client: unknown client
            )
    --- FAIL: TestM63Empirical_AllScenarios/Subtest_B:_Unconfigured_Credentials_return_400_Bad_Request (0.00s)
        social_m6_3_empirical_test.go:185: expected body to mention 'unconfigured', got: invalid client: unknown client
FAIL	github.com/Uttam-Mahata/DilithiAuth/pkg/oauth	4.619s
```

#### Cause of Test Failure:
In `pkg/oauth/social_m6_3_empirical_test.go`, the test configures tenant middleware via `r.Use(tenant.NewMiddleware([]string{"0.0.0.0/0", "::/0"}))`. However, the test constructs HTTP requests using `http.NewRequest(...)` without setting `RemoteAddr`. As a result, `tenant.fromTrustedNetwork` evaluates `RemoteAddr == ""` to `false`, header `X-Tenant-Id: tenant-emp-1` is ignored, and the tenant ID resolves to `"default"`. When `SocialLoginGET` attempts to validate the client (`emp_client_id`) which was registered under `"tenant-emp-1"`, `ValidateClient` fails with `invalid client: unknown client` (HTTP 400).

---

## Forensic Audit Conclusion

While the core implementation code (`pkg/oauth/social.go`, `pkg/oauth/authz_handlers.go`) and primary unit/challenge test files (`pkg/oauth/social_test.go`, `pkg/oauth/social_challenge_test.go`) are free of hardcoding, facades, or shortcuts, the test suite `go test ./...` fails during behavioral execution.

Per Integrity Forensics guidelines, a work product whose test suite fails execution cannot be certified as clean.

**Verdict: INTEGRITY VIOLATION**
