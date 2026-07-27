# Social Authentication Code & Security Review Report

**Reviewer**: `teamwork_preview_reviewer_m6_1`  
**Date**: 2026-07-27  
**Scope**: Social Authentication Implementation (`pkg/oauth/social.go`, `pkg/oauth/authz_handlers.go`, `cmd/server/main.go`, `deployments/dilithiauth.service`, `docs/oci-vm-deployment.md`, `docs/deployment.md`, `pkg/oauth/social_test.go`)  
**Verdict**: **REJECTED** (REQUEST_CHANGES)

---

## Executive Summary

An independent code quality and security review was conducted on the Social Authentication feature in DilithiAuth. While the general structure (OAuth 2.0 state JWT signing, Google/GitHub provider integration, end-to-end flow) is implemented and well-documented across deployment files, the codebase fails automated test suite compilation (`go test ./...`) and contains critical/major security vulnerabilities regarding CSRF secret fallbacks, cross-tenant isolation, and unverified email account linking.

---

## Test & Build Verification

1. **Build Status**:
   - Command: `go build -o dilithiauth-server ./cmd/server`
   - Result: **PASS** (Executable compiled successfully).

2. **Test Status**:
   - Command: `go test ./...`
   - Result: **FAIL**
   - Fail Output:
     ```
     # github.com/Uttam-Mahata/DilithiAuth/pkg/oauth [github.com/Uttam-Mahata/DilithiAuth/pkg/oauth.test]
     pkg/oauth/social_challenge_test.go:4:2: "context" imported and not used
     FAIL	github.com/Uttam-Mahata/DilithiAuth/pkg/oauth [build failed]
     ```

---

## Detailed Findings

### [Critical] Finding 1: Test Suite Compilation Failure
- **What**: `pkg/oauth/social_challenge_test.go` fails to compile due to an unused import `"context"`.
- **Where**: `pkg/oauth/social_challenge_test.go:4:2`
- **Why**: Go compiler enforces no unused imports. Any `go test ./...` invocation fails.
- **Suggestion**: Remove `"context"` from `pkg/oauth/social_challenge_test.go`.

### [Critical / Major] Finding 2: Hardcoded Fallback Secret for Social Auth State JWTs
- **What**: Fallback to a hardcoded secret string when `ConsentSecret` is omitted or empty.
- **Where**: `pkg/oauth/authz_handlers.go:1242-1247` (`socialSecret()`)
  ```go
  func (h *AuthzHandlers) socialSecret() []byte {
      if len(h.ConsentSecret) > 0 {
          return h.ConsentSecret
      }
      return []byte("dilithiauth-default-social-state-secret")
  }
  ```
- **Why**: If `ConsentSecret` is not set or initialized without it, state JWTs are signed with a known public secret (`"dilithiauth-default-social-state-secret"`). Attackers can forge state JWTs to execute CSRF attacks and inject unauthorized parameters.
- **Suggestion**: Require `ConsentSecret` (or `SocialStateSecret`) to be non-empty, or auto-generate a random 32-byte secret during initialization if omitted. Never use a hardcoded static string.

### [Major] Finding 3: Missing Tenant ID in Social State Claims (Cross-Tenant Vulnerability)
- **What**: `socialStateClaims` struct does not record the `TenantID` under which the social login flow was initiated.
- **Where**: `pkg/oauth/authz_handlers.go:1229-1240` (`socialStateClaims`) & `SocialCallbackGET` (lines 1386-1450)
- **Why**: In `SocialCallbackGET`, `tenantID` is read from `tenant.FromContext(r.Context())`. An attacker initiating social login on Tenant A can send the callback request to Tenant B, causing user provisioning and code issuance in Tenant B context.
- **Suggestion**: Include `TenantID string json:"tenant_id"` in `socialStateClaims`. In `SocialCallbackGET`, verify that `stateClaims.TenantID == tenantID` and return HTTP 400 on mismatch.

### [Major] Finding 4: Account Takeover Risk via Unverified Email in GitHub Social Auth
- **What**: Primary GitHub email extraction ignores the `verified` status of the email address.
- **Where**: `pkg/oauth/social.go:269-285`
- **Why**: On GitHub, users can add unverified secondary/primary email addresses. The current loop accepts any primary email without checking `em["verified"].(bool)`. If an attacker adds an unverified email matching an existing DilithiAuth user, they can hijack that user's account.
- **Suggestion**: Verify `verified, _ := em["verified"].(bool)` and only accept emails where `verified == true`.

### [Major] Finding 5: Ghost User Provisioning on Registration Failure
- **What**: Registration errors during social user provisioning are silently ignored.
- **Where**: `pkg/oauth/authz_handlers.go:1425-1445`
- **Why**: If `UserRegistrar` or `Store.RegisterUser` fails (e.g. duplicate username or DB constraint), `regErr` is ignored and `userSubject` remains the generated `usr_<rand>` without persisting the record.
- **Suggestion**: Check registration errors. If provisioning fails, return an HTTP 500/400 error rather than issuing an authorization code for a non-existent subject.

### [Minor] Finding 6: Unhandled HTTP Request Creation & Raw Error Reflection
- **What**: Ignored errors from `http.NewRequestWithContext` and raw external error string reflection.
- **Where**: `pkg/oauth/social.go:89, 113, 176, 201, 258` & `pkg/oauth/authz_handlers.go:1377`
- **Why**: Ignored request creation errors could lead to nil-pointer panics, and echoing raw external response bodies in `http.Error` can leak internal error details or allow HTML reflection.
- **Suggestion**: Check `err` on `http.NewRequestWithContext` and sanitize external provider error messages before sending them in HTTP responses.

---

## Verified Claims & Documentation Check

- **Deployment Docs & Systemd Alignment**: `deployments/dilithiauth.service`, `docs/oci-vm-deployment.md`, and `docs/deployment.md` properly document all social configuration environment variables (`DILITHIAUTH_GOOGLE_*`, `DILITHIAUTH_GITHUB_*`).
- **Build Status**: Verified binary compilation via `go build -o dilithiauth-server ./cmd/server`.
- **Test Execution**: Failed due to `social_challenge_test.go` compilation error.

---

## Summary Verdict

**Verdict**: **REJECTED**
The implementation cannot be approved until test compilation is restored and security findings (Finding 2, 3, 4, 5) are addressed.
