# Changes Summary

## Remediation Overview
Resolved the Forensic Auditor finding in `pkg/oauth/social_m6_3_empirical_test.go` by verifying and ensuring that HTTP requests constructed for tenant-middleware-protected routes set `req.RemoteAddr = "127.0.0.1:12345"`.

## Files Modified / Inspected
1. **`pkg/oauth/social_m6_3_empirical_test.go`**:
   - Inspected test request creation logic inside `makeReq`.
   - Confirmed `req.RemoteAddr = "127.0.0.1:12345"` is set on test requests so `tenant.NewMiddleware([]string{"127.0.0.1/32"})` recognizes the incoming request as originating from a trusted loopback CIDR.
   - Ensures `X-Tenant-Id: tenant-emp-1` header is honored by tenant middleware instead of falling back to the `"default"` tenant.

2. **`pkg/oauth/*_test.go`**:
   - Inspected all test files in `pkg/oauth/` (`authz_test.go`, `grant_store_test.go`, `oauth_test.go`, `refresh_store_test.go`, `social_challenge_test.go`, `social_test.go`).
   - Verified no other tests bypass tenant trusted network checks or fail tenant resolution.

## Verification
- `go test -count=1 ./...`: 100% PASS across all packages.
- `go test -v -count=1 ./pkg/oauth -run TestM63Empirical_AllScenarios`: All 5 subtests (A, B, C, D, E) PASS.
- `go build -o dilithiauth-server ./cmd/server`: Clean compilation (binary size ~24MB).
