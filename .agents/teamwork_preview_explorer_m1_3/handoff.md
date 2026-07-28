# Handoff Report — Technical Survey of `pkg/transport` & Benchmark Suite Inspection

## 1. Observation

### Codebase & Package Architecture Overview
Direct inspection of `pkg/transport/`, `pkg/signer/`, `pkg/oauth/`, `pkg/verifier/`, `cmd/server/`, and `tests/` revealed the following exact file locations and implementations:

#### A. HTTP Middleware & Header Limits (`pkg/transport/middleware.go`)
- **`MaxHeaderBytesForPQC`**: Defined at `pkg/transport/middleware.go:18` as `1 << 14` (16 KiB = 16,384 bytes).
- **`NewPQCServer(addr, handler)`**: Defined at `pkg/transport/middleware.go:23-29`. Returns an `*http.Server` with `MaxHeaderBytes` explicitly configured to 16 KiB to prevent 400 Bad Request / 431 Request Header Fields Too Large errors when handling Jumbo JWTs (~3.5–4 KiB).
- **`JumboHeaderMiddleware(next)`**: Defined at `pkg/transport/middleware.go:39-43`. Pass-through middleware for handler chain documentation and future metrics hooks.
- **`SigStoreMiddleware(store)`**: Defined at `pkg/transport/middleware.go:169-177`. Attaches `*SigStore` to request context (`WithSigStore`/`SigStoreFromContext`).

#### B. Sidecar Signature Store (`SigStore`) & Storage Models
- **`SigStoreBackend` Interface**: Defined at `pkg/transport/middleware.go:50-55`:
  ```go
  type SigStoreBackend interface {
      Put(jti string, sig []byte)
      Get(jti string) ([]byte, bool)
      Delete(jti string)
      Len() int
  }
  ```
- **Storage Backends Implemented**:
  1. **`MemorySigStore`** (`pkg/transport/memory_store.go:22-89`): Thread-safe map guarded by `sync.RWMutex`. Defensive byte copying on `Put` (`make([]byte, len(sig))` + `copy`).
  2. **`RedisSigStore`** (`pkg/transport/redis_store.go:14-133`): Redis backend via `github.com/redis/go-redis/v9`. Key prefix defaults to `"sigstore:"`. Native TTL via `r.client.Set(ctx, key, sig, ttl)`. ACL authentication supported via `WithAuth`.
  3. **`FileSigStore`** (`pkg/transport/file_store.go:19-145`): JSON file persistence combined with in-memory cache map. Writes/deletes use atomic flush (`.tmp` file + `os.Rename`).
  4. **`PostgresSigStore`** (`pkg/transport/postgres_store.go:14-127`): PostgreSQL persistence in `sig_store` table (`jti TEXT PRIMARY KEY, sig BYTEA NOT NULL, created_at TIMESTAMPTZ DEFAULT NOW()`). Connection pool tuned via `configureDBPool` (`DILITHIAUTH_DB_MAX_CONNS` or default 25).

#### C. `pqc_ref` Lookup Resolution
- **Issuance**: `pkg/oauth/oauth.go:190-240` (`IssueJWTThinWithScope`) and `pkg/oauth/oauth.go:337-392` (`IssueCWTWithScope`). Computes `pqc_ref = base64url(SHA-256(mldsa44_sig))`.
- **Verification Resolution**:
  - Thin JWT: `pkg/verifier/verifier.go:104-120`. Extracts `claims.PQCRef`, calls `v.SigStore.Get(claims.PQCRef)`, verifies SHA-256 match `base64url(SHA-256(sig)) == claims.PQCRef`, then verifies ML-DSA-44 signature using `kp.VerifyMLDSA(payloadBytes, hybridSig)`.
  - CWT: `pkg/verifier/verifier.go:241-260`. Extracts `claims.JWTID` (`cti`), calls `v.SigStore.Get(claims.JWTID)`, checks `base64url(SHA-256(pqcSig)) == claims.PQCRef`, then verifies ML-DSA-44 signature.

#### D. Cache TTL & Expiry Management
- **In-Memory & File**: `sigEntry.expiresAt` (`memory_store.go:8-11`, `file_store.go:14-17`). `Get()` evaluates expiration on access (`time.Now().After(entry.expiresAt)`). Background cleanup handled by `Purge()`.
- **Redis**: Expiration handled natively by Redis daemon using `WithTTL(ttl)` parameter in `r.client.Set`.
- **Postgres**: SQL row deletion via `Purge()`: `DELETE FROM sig_store WHERE created_at < NOW() - ttl`.

#### E. Replay Protection
- **Token Identification**: 16-byte base64url random UUID (`jti`/`cti`) generated in `pkg/oauth/oauth.go:394-401`.
- **Token Expiry**: Verified via `jwt.RegisteredClaims.ExpiresAt` / `CWTClaims.ExpiresAt`.
- **Revocation**: Handled via `RevocationChecker` interface in `pkg/verifier/verifier.go:39-42`.
- **OAuth Code Flow**: Single-use authorization codes redeemed in `pkg/oauth/grant_store.go` (`RedeemCode` deletes code upon use) and PKCE S256 code verifier validation (`pkg/oauth/authz_handlers.go`).

#### F. Test & Benchmark Suite Inspection & Code Coverage
- **Benchmark Functions**: Global `grep_search` for `func Benchmark` and `testing.B` returned **0 matches**. No standard Go benchmark functions exist in the repository.
- **Unit & Integration Test Coverage** (measured via `go test -cover ./...`):
  | Package | Test Coverage | Key Test Files |
  |---|---|---|
  | `pkg/signer` | **60.3%** | `pkg/signer/signer_test.go`, `pkg/signer/keystore_test.go` |
  | `pkg/oauth` | **66.7%** | `pkg/oauth/oauth_test.go`, `pkg/oauth/authz_test.go`, `grant_store_test.go`, `social_test.go` |
  | `pkg/transport` | **50.5%** | `middleware_test.go`, `memory_store_test.go`, `file_store_test.go`, `redis_store_test.go`, `postgres_store_test.go` |
  | `pkg/verifier` | **65.8%** | `pkg/verifier/verifier_test.go`, `pkg/verifier/middleware_test.go` |
  | `cmd/server` | **6.0%** | `cmd/server/main_test.go` |
  | `pkg/tenant` | **83.7%** | `pkg/tenant/tenant_test.go` |
  | `pkg/ratelimit` | **77.8%** | `pkg/ratelimit/ratelimit_test.go` |
  | `pkg/mfa` | **87.1%** | `pkg/mfa/totp_test.go` |
  | `tests` | **Passing** | `tests/e2e_oauth_test.go` |

---

## 2. Logic Chain

1. **Observation**: `pkg/transport/middleware.go` defines `MaxHeaderBytesForPQC = 1 << 14` (16 KiB) and `NewPQCServer`.
   **Reasoning**: Quantum-safe ML-DSA-44 signatures are ~2.4 KB binary (~3.3 KB base64url encoded). Embedded in a "Jumbo" JWT (`pqc_sig`), total HTTP Authorization header size reaches ~3.8–4.0 KB. Standard HTTP servers / proxies (e.g., Node.js, Nginx defaults) reject headers > 4–8 KB with HTTP 431. Configuring `MaxHeaderBytes = 16 KiB` at the Go `http.Server` layer ensures Jumbo tokens are accepted without header truncation.

2. **Observation**: Thin JWTs omit `pqc_sig` and embed `pqc_ref`, delegating full signature storage to `SigStore`.
   **Reasoning**: Thin JWT tokens remain under 500 bytes (comparable to classical Ed25519 JWTs). When high-security verifiers require post-quantum auditability, `pkg/verifier` retrieves the ML-DSA-44 signature from `SigStore` using `pqc_ref` or `cti`/`jti` and asserts cryptographic hash equality `SHA-256(sig) == pqc_ref` before executing `kp.VerifyMLDSA`.

3. **Observation**: No `Benchmark*` functions currently exist in `pkg/signer`, `pkg/oauth`, `pkg/transport`, or `cmd/server`.
   **Reasoning**: Benchmark functions need to be constructed or executed via dedicated benchmark tests in order to collect empirical metrics for signing latency, verification latency, token sizes, and HTTP header overhead for the research paper.

---

## 3. Caveats

- **External Store Test Skipping**: `redis_store_test.go` and `postgres_store_test.go` skip automatically during `go test` unless `REDIS_ADDR` and `POSTGRES_DSN` environment variables are set.
- **cmd/server Coverage**: Low statement coverage (6.0%) in `cmd/server` is expected because `main.go` acts primarily as an initialization binary while core logic lives in `pkg/`.

---

## 4. Conclusion

- `pkg/transport` provides a clean, decoupled abstraction (`SigStoreBackend` interface with In-Memory, Redis, File, and PostgreSQL implementations) for offloading PQC signatures and enforcing 16 KiB header limits for PQC tokens.
- All unit and integration test suites pass successfully across all core packages (`pkg/signer`, `pkg/oauth`, `pkg/transport`, `pkg/verifier`, `tests`).
- `Benchmark*` functions are currently absent and need to be implemented or benchmarked in downstream tasks to collect empirical latency and token size data for the research paper.

---

## 5. Verification Method

To independently verify these findings, run the following commands from the repository root:

```bash
# 1. Run all unit tests and verify coverage
go test -cover ./pkg/signer ./pkg/oauth ./pkg/transport ./pkg/verifier ./cmd/server

# 2. Run E2E integration test suite
go test -v ./tests/e2e_oauth_test.go

# 3. Confirm absence of Benchmark functions
grep -rn "func Benchmark" pkg/ cmd/ tests/
```
