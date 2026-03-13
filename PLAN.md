 DilithiAuth: Production Hardening, Performance Optimization & FFI/Client SDKs                                                                           
                                                                                                                                                         
 Context

 DilithiAuth is a functional post-quantum OAuth 2.0 identity provider, but a thorough audit revealed critical security gaps (plaintext passwords, no
 rate limiting, unbounded memory growth), performance bottlenecks (double-parse/marshal, no connection pooling, no benchmarks), and no client SDK
 strategy for non-Go consumers. This plan addresses all three areas in priority order so the project can be deployed as a real OAuth provider and
 consumed from Python, Node.js, Java, and Rust.

 ---
 Phase 1: Security-Critical Fixes

 1.1 Password Hashing

 - File: pkg/oauth/authz.go — line 72 does u.Password != password (plaintext)
 - Add golang.org/x/crypto/bcrypt to go.mod
 - Change DemoUser.Password → DemoUser.PasswordHash
 - Add NewDemoUser(username, plaintext, subject) that bcrypt-hashes
 - AuthenticateUser() → bcrypt.CompareHashAndPassword()
 - Update cmd/server/main.go:112-113 and pkg/oauth/authz_test.go

 1.2 Admin Endpoint Authentication

 - File: cmd/server/main.go:148 — /admin/rotate has zero auth
 - Add env var DILITHIAUTH_ADMIN_TOKEN
 - Create AdminAuthMiddleware(token) in pkg/transport/middleware.go
 - Wrap /admin/rotate route; refuse to register if token is empty

 1.3 Memory Leak Fixes (MemorySigStore + AuthzStore)

 - File: pkg/transport/memory_store.go — map grows unbounded
 - Add ttl field, sigEntry{sig, createdAt} wrapper, cleanup goroutine (5-min sweep, 1-hour default TTL), Stop() method
 - File: pkg/oauth/authz.go:41 — codes map never cleaned
 - Add startCodeCleanup() goroutine to remove expired/used codes
 - Add Close() method

 1.4 SigStoreBackend Error Propagation

 - File: pkg/transport/middleware.go:50-55 — interface returns no errors
 - Change Put(jti, sig) → Put(jti, sig) error, Delete(jti) → Delete(jti) error
 - Update all 4 implementations + SigStore wrapper + callers in main.go, authz_handlers.go
 - File: pkg/transport/postgres_store.go:44 — _, _ = p.db.Exec(...) stops swallowing errors

 1.5 PKCE Support (RFC 7636)

 - File: pkg/oauth/authz.go — AuthCode struct
 - Add CodeChallenge, CodeChallengeMethod fields
 - CreateCode() accepts challenge params
 - RedeemCode() accepts code_verifier, validates S256
 - Update authz_handlers.go to pass through PKCE params

 Verification

 go test ./...
 # Manual: curl /admin/rotate without token → 401
 # Manual: curl /admin/rotate with correct token → 200

 ---
 Phase 2: Reliability & Observability

 2.1 Graceful Shutdown

 - File: cmd/server/main.go:174 — no signal handling
 - Run ListenAndServe() in goroutine
 - Use signal.NotifyContext() for SIGINT/SIGTERM
 - Call srv.Shutdown(ctx) with 30s timeout
 - Close SigStore backends (Postgres .Close(), Redis .Close(), Memory .Stop())

 2.2 Deep Health Check

 - File: cmd/server/main.go:162-165 — always returns 200
 - Add HealthChecker interface with Ping(ctx) error
 - Implement for Postgres (db.PingContext) and Redis (client.Ping)
 - /healthz returns 503 if any backend unhealthy

 2.3 Rate Limiting

 - Add golang.org/x/time/rate dependency
 - Create pkg/transport/ratelimit.go — per-IP rate limiter middleware using sync.Map
 - Apply to /authorize, /token, /admin/rotate
 - Configure via DILITHIAUTH_RATE_LIMIT (req/s) and DILITHIAUTH_RATE_BURST

 2.4 Structured Logging

 - Replace log.Printf with log/slog (stdlib, Go 1.21+)
 - Add DILITHIAUTH_LOG_FORMAT env var (json/text)
 - Update all log calls in cmd/server/main.go

 2.5 Prometheus Metrics

 - Add github.com/prometheus/client_golang dependency
 - Create pkg/metrics/metrics.go with counters: tokens_issued_total{type}, token_verifications_total{result}, sigstore_operations_total{op,backend}
 - Add /metrics endpoint
 - Instrument token handlers and verifier middleware

 Verification

 go test ./...
 # Start server, send SIGTERM → observe clean shutdown
 # curl /healthz with Redis down → 503
 # Rapid curl /authorize → 429 after burst
 # curl /metrics → prometheus output

 ---
 Phase 3: Performance Optimizations

 3.1 Eliminate Verifier Double-Parse

 - File: pkg/verifier/verifier.go:50-67 — ParseUnverified then ParseWithClaims
 - Do ParseWithClaims once in VerifyJWT(), then branch on claims.PQCSig
 - Extract internal helpers verifyMLDSAFromStore() and verifyMLDSAEmbedded()
 - Remove ParseUnverified call entirely
 - Expected gain: ~30% faster JWT verification

 3.2 Deduplicate Struct Construction in IssueJWTFull

 - File: pkg/oauth/oauth.go:237-263 — RegisteredClaims constructed twice (lines 237 and 255)
 - Build HybridClaims once, marshal RegisteredClaims for MLDSA signing, then set PQCSig
 - Note: double-marshal is unavoidable (signed content ≠ final content) — document this

 3.3 KeyManager: Map-Based Kid Lookup

 - File: pkg/signer/keymanager.go:67-84 — linear search O(n)
 - Add keysByKid map[string]*HybridKeyPair field
 - Populate on load and Rotate()
 - ActiveKey() and KeyByKid() become O(1)

 3.4 PostgresSigStore: Connection Pool + Prepared Statements

 - File: pkg/transport/postgres_store.go:20-28
 - After sql.Open: db.SetMaxOpenConns(25), db.SetMaxIdleConns(10), db.SetConnMaxLifetime(5*min)
 - db.Prepare() for PUT/GET/DELETE queries in constructor, store as fields
 - Add configurable env vars for pool sizes

 3.5 FileSigStore: Debounced Writes

 - File: pkg/transport/file_store.go:46-52 — full JSON flush under mutex on every Put
 - Add dirty flag + flush goroutine (every 5s)
 - Put/Delete set dirty flag and return immediately
 - Close() does final flush

 3.6 Add Benchmarks

 - pkg/oauth/oauth_test.go: BenchmarkIssueJWTThin, BenchmarkIssueJWTFull, BenchmarkIssueCWT
 - pkg/verifier/verifier_test.go: BenchmarkVerifyJWTThin, BenchmarkVerifyJWTFull, BenchmarkVerifyCWT
 - pkg/signer/*_test.go: BenchmarkSign, BenchmarkVerifyMLDSA, BenchmarkKeyByKid

 Verification

 go test -bench=. -benchmem ./...
 go test ./...

 ---
 Phase 4: Token Revocation

 4.1 Blocklist Infrastructure

 - Create pkg/transport/blocklist.go — TokenBlocklist interface: Revoke(jti, expiresAt), IsRevoked(jti) bool
 - MemoryBlocklist — map with TTL cleanup (same pattern as fixed MemorySigStore)
 - Optional: RedisBlocklist using Redis SET with TTL

 4.2 Integration

 - Add POST /admin/revoke endpoint (guarded by admin auth from 1.2)
 - In verifier.Middleware() — check blocklist before accepting token
 - Add Blocklist field to Verifier struct

 Verification

 # Issue token → verify works → revoke → verify returns 401
 go test ./pkg/verifier -run TestRevocation

 ---
 Phase 5: FFI Shared Library & Client SDKs

 5.1 Core FFI Layer

 New files:
 - pkg/ffi/ffi.go — Go functions with //export CGO directives
 - cmd/libdilithiauth/main.go — build target (package main + import "C")

 Exported C functions:

 ┌─────────────────────────────┬─────────────────────────┬──────────────────────────────────────────────────────────────┐
 │          Function           │         Purpose         │                          Signature                           │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_VerifyJWTThin   │ Verify thin JWT         │ (token, jwksJSON, pqcSigB64 *C.char, outJSON **C.char) C.int │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_VerifyJWTFull   │ Verify full JWT         │ (token, jwksJSON *C.char, outJSON **C.char) C.int            │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_VerifyCWT       │ Verify CWT              │ (cwtBytes, cwtLen, jwksJSON, pqcSig, pqcLen, outJSON) C.int  │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_GenerateKeyPair │ Generate hybrid keypair │ (outJSON **C.char) C.int                                     │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_Sign            │ Dual sign message       │ (keyJSON, msg, msgLen, outJSON) C.int                        │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_VerifyMLDSA     │ Verify ML-DSA-44 only   │ (pubKeyB64, msg, msgLen, sigB64) C.int                       │
 ├─────────────────────────────┼─────────────────────────┼──────────────────────────────────────────────────────────────┤
 │ DilithiAuth_Free            │ Free C-allocated memory │ (ptr *C.char)                                                │
 └─────────────────────────────┴─────────────────────────┴──────────────────────────────────────────────────────────────┘

 Memory convention: All output via C.CString (malloc'd); callers must call DilithiAuth_Free. Return 0 = success, negative = error.

 Build:
 go build -buildmode=c-shared -o libdilithiauth.so ./cmd/libdilithiauth
 # Produces: libdilithiauth.so + libdilithiauth.h

 5.2 REST Verification Endpoint (Simpler Alternative)

 For clients that don't want FFI complexity:
 - Add POST /verify/jwt — accepts {token, pqc_sig?}, returns {claims} or {error}
 - Add POST /verify/cwt — accepts raw CBOR body + X-PQC-Sig header
 - Rate-limited, requires no native library

 5.3 Python SDK (sdk/python/dilithiauth/)

 - ctypes wrapper loading libdilithiauth.so
 - Class: DilithiAuth(so_path) with methods verify_jwt_thin(), verify_jwt_full(), verify_cwt()
 - Auto DilithiAuth_Free in __del__ or context manager
 - Package as PyPI wheel with bundled .so

 5.4 Node.js SDK (sdk/node/)

 - Use koffi (modern FFI) to bind libdilithiauth.so
 - Async wrappers returning Promise<Claims>
 - npm package with prebuilt binaries

 5.5 Java SDK (sdk/java/)

 - JNA (Java Native Access) binding
 - Maven artifact with platform-specific .so/.dylib/.dll

 5.6 Rust SDK (sdk/rust/)

 - bindgen from libdilithiauth.h to generate FFI types
 - Safe Rust wrapper crate
 - Publish to crates.io

 5.7 Go Client SDK (sdk/go/)

 - Pure Go package that wraps pkg/verifier + JWKS HTTP fetching
 - No FFI needed — direct import of DilithiAuth packages
 - dilithiauth.NewClient(issuerURL) → auto-fetches JWKS, creates Verifier
 - Cacheable JWKS with background refresh
 - Publish as Go module github.com/Uttam-Mahata/dilithiauth-go

 5.8 .NET SDK (sdk/dotnet/)

 - P/Invoke wrapper loading libdilithiauth.so / libdilithiauth.dll
 - DilithiAuth.Verifier class with VerifyJwtThin(), VerifyJwtFull(), VerifyCwt()
 - NuGet package with platform-specific native libraries (runtime RIDs)
 - Targets .NET 8+

 Verification

 # Build shared library
 go build -buildmode=c-shared -o libdilithiauth.so ./cmd/libdilithiauth

 # Python
 cd sdk/python && python -m pytest tests/

 # Node.js
 cd sdk/node && npm test

 # REST endpoint
 curl -X POST http://localhost:8080/verify/jwt -d '{"token":"eyJ..."}'

 ---
 Phase 6: CI/CD

 - .github/workflows/ci.yml: go test ./..., go build ./cmd/server, shared lib build, benchmark comparison on PRs
 - Cross-platform builds: Linux (x86, ARM), macOS (Intel, Apple Silicon)
 - SDK package publishing (PyPI, npm, Maven Central, crates.io)

 ---
 Dependency Ordering

 Phase 1 (Security) → Phase 2 (Observability) → Phase 3 (Performance) → Phase 4 (Revocation) → Phase 5 (FFI) → Phase 6 (CI)

 Phase 1.4 (error interface change) is a prerequisite for Phase 4 (blocklist uses same patterns).
 Phase 3.1 (verifier fix) should precede Phase 5 (FFI wraps verifier).

 ---
 New Dependencies

 ┌─────────────────────────────────────┬───────┬──────────────────┐
 │               Package               │ Phase │     Purpose      │
 ├─────────────────────────────────────┼───────┼──────────────────┤
 │ golang.org/x/crypto/bcrypt          │ 1.1   │ Password hashing │
 ├─────────────────────────────────────┼───────┼──────────────────┤
 │ golang.org/x/time/rate              │ 2.3   │ Rate limiting    │
 ├─────────────────────────────────────┼───────┼──────────────────┤
 │ github.com/prometheus/client_golang │ 2.5   │ Metrics          │
 └─────────────────────────────────────┴───────┴──────────────────┘

 ---
 Files Modified (by phase)

 Phase 1: pkg/oauth/authz.go, pkg/oauth/authz_test.go, pkg/oauth/authz_handlers.go, pkg/transport/middleware.go, pkg/transport/memory_store.go,
 pkg/transport/postgres_store.go, pkg/transport/redis_store.go, pkg/transport/file_store.go, cmd/server/main.go, go.mod

 Phase 2: cmd/server/main.go, pkg/transport/ratelimit.go (new), pkg/metrics/metrics.go (new)

 Phase 3: pkg/verifier/verifier.go, pkg/oauth/oauth.go, pkg/signer/keymanager.go, pkg/transport/postgres_store.go, pkg/transport/file_store.go,
 *_test.go files (benchmarks)

 Phase 4: pkg/transport/blocklist.go (new), pkg/verifier/verifier.go, pkg/verifier/middleware.go, cmd/server/main.go

 Phase 5: pkg/ffi/ffi.go (new), cmd/libdilithiauth/main.go (new), sdk/python/ (new), sdk/node/ (new), sdk/java/ (new), sdk/rust/ (new), sdk/go/ (new),
 sdk/dotnet/ (new)