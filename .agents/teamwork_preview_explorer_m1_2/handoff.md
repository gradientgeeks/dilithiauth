# Technical Survey Report: `pkg/oauth` and `cmd/server` in DilithiAuth

## 1. Observation

### 1.1 OIDC Discovery & JWKS Public Key Publication
- **OIDC Discovery Endpoint:** Defined in `pkg/oauth/oauth.go` (lines 23–32, 79–94, 123–131) and mounted at `GET /.well-known/openid-configuration` in `cmd/server/main.go` (line 461).
  - `DiscoveryDocument` struct:
    ```go
    type DiscoveryDocument struct {
        Issuer                           string   `json:"issuer"`
        AuthorizationEndpoint            string   `json:"authorization_endpoint"`
        TokenEndpoint                    string   `json:"token_endpoint"`
        JWKsURI                          string   `json:"jwks_uri"`
        ResponseTypesSupported           []string `json:"response_types_supported"`
        SubjectTypesSupported            []string `json:"subject_types_supported"`
        IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
        PQCHybridEnabled                 bool     `json:"pqc_hybrid_enabled"`
    }
    ```
  - Configured response values:
    - `id_token_signing_alg_values_supported`: `["RS256", "EdDSA", "MLDSA44"]`
    - `pqc_hybrid_enabled`: `true`
- **JWKS Endpoint & Custom Public Key Field:** Defined in `pkg/oauth/oauth.go` (lines 35–49, 98–119, 134–142) and mounted at `GET /.well-known/jwks.json` in `cmd/server/main.go` (line 462).
  - `JWK` struct:
    ```go
    type JWK struct {
        Kid        string `json:"kid"`
        Kty        string `json:"kty"`
        Crv        string `json:"crv,omitempty"`
        X          string `json:"x,omitempty"`           // Ed25519 public key (base64url)
        MLDSA44Pub string `json:"mldsa44_pub,omitempty"` // ML-DSA-44 public key (base64url)
        Use        string `json:"use"`
        Alg        string `json:"alg"`
    }
    ```
  - Published attributes per key entry:
    - `kid`: Key ID (8 random bytes base64url raw encoded, 11 chars).
    - `kty`: `"OKP"` (Octet Key Pair, RFC 8037).
    - `crv`: `"Ed25519"`.
    - `x`: Base64url raw unpadded string of 32-byte Ed25519 public key (43 characters).
    - `mldsa44_pub`: Base64url raw unpadded string of 1,312-byte ML-DSA-44 public key (NIST FIPS 204, 1,750 characters).
    - `use`: `"sig"`.
    - `alg`: `"EdDSA+MLDSA44"`.
  - **Key Observation:** Both classical (`x`) and post-quantum (`mldsa44_pub`) public keys share the exact same `kid` in the JWKS document.

### 1.2 Token Issuance Strategies
DilithiAuth implements three token strategies in `pkg/oauth/oauth.go` and exposes them in `cmd/server/demo_routes.go` (`//go:build demo`) and `pkg/oauth/authz_handlers.go`:

1. **Thin Tokens (`POST /token/thin` & `IssueJWTThinWithScope`):**
   - **File & Lines:** `pkg/oauth/oauth.go` lines 180–240; `pkg/oauth/authz_handlers.go` lines 1116–1128; `cmd/server/demo_routes.go` lines 29–51.
   - **Claims Struct (`HybridClaims`):** Standard claims (`iss`, `sub`, `exp`, `iat`, `jti`) + `tid` (tenant ID), `scope`, and `pqc_ref` (SHA-256 hash reference to ML-DSA-44 signature). `pqc_sig` is omitted (`omitempty`).
   - **Header & Signature:** Signed using Ed25519 (`jwt.SigningMethodEdDSA`). Header: `{"alg":"EdDSA","kid":"...","typ":"JWT"}`.
   - **Signature Reference (`pqc_ref`):** `sha256.Sum256(hybridSig.MLDSASig)` encoded as raw Base64url (43 chars).
   - **Sidecar Storage:** Full 2,420-byte ML-DSA-44 signature (`hybridSig.MLDSASig`) is stored in `SigStore` (backed by Memory, File, PostgreSQL, or Redis) keyed by `pqc_ref`.
   - **Default Strategy:** Standard OAuth 2.0 authorization code token exchange (`POST /token`) uses `IssueJWTThinWithScope`.

2. **Jumbo Tokens (`POST /token/full` & `IssueJWTFullWithScope`):**
   - **File & Lines:** `pkg/oauth/oauth.go` lines 242–295; `cmd/server/demo_routes.go` lines 53–73.
   - **Claims Struct (`HybridClaims`):** Standard claims + `tid`, `scope`, and `pqc_sig` (embedded raw ML-DSA-44 signature). `pqc_ref` is omitted (`omitempty`).
   - **Embedded Signature (`pqc_sig`):** Base64url raw unpadded string of 2,420-byte ML-DSA-44 signature (~3,227 chars).
   - **Stateless Nature:** Self-contained token; no `SigStore` sidecar lookup needed during verification.
   - **Server Configuration:** Standard web servers fail on headers > 4–8 KB. DilithiAuth configures HTTP server `MaxHeaderBytes = 16 KiB` (`transport.NewPQCServer`, `cmd/server/main.go` line 549) and applies `transport.JumboHeaderMiddleware` (`cmd/server/main.go` line 368).

3. **CBOR Web Tokens / CWT (`POST /token/cwt` & `IssueCWTWithScope`):**
   - **File & Lines:** `pkg/oauth/oauth.go` lines 298–393; `cmd/server/demo_routes.go` lines 75–96.
   - **Claims Struct (`CWTClaims` with CBOR integer keys per RFC 8392):**
     - `1`: `iss` (string)
     - `2`: `sub` (string)
     - `3`: `aud` (string, omitempty)
     - `4`: `exp` (int64 UNIX timestamp)
     - `6`: `iat` (int64 UNIX timestamp)
     - `7`: `cti` (string, token ID / jti)
     - `8`: `pqc_ref` (string, SHA-256 reference hash, omitempty)
     - `9`: `tid` (string, tenant ID, omitempty)
     - `10`: `scope` (string, granted scopes, omitempty)
   - **Envelope Format:** Binary CBOR two-element array: `[tokenBytes (CBOR-encoded claims), ed25519_sig (64 raw bytes)]`.
   - **Sidecar Storage:** ML-DSA-44 signature stored in `SigStore` keyed by `cti` (token ID / JWT ID).

### 1.3 Empirical Size Measurements
Running unit test `TestCWTSmallerThanJWT` in `pkg/oauth/oauth_test.go`:
- **Thin JWT size:** `401 bytes` (Base64url text format)
- **CWT size:** `197 bytes` (Binary CBOR format)
- **Empirical Reduction:** CWT is **50.87% smaller** than Thin JWT for identical claim payload data.
- **Jumbo JWT size:** ~`3,850 bytes` to `4,100 bytes` (due to ~3,227 char `pqc_sig` Base64url payload string).

---

## 2. Logic Chain

1. **Dual-Algorithm Identity Design:**
   - Observations of `JWK` struct and `DiscoveryDocument()` show that DilithiAuth couples classical Ed25519 (fast, legacy compatibility) with ML-DSA-44 (NIST FIPS 204 quantum resistance).
   - Publishing both `x` and `mldsa44_pub` under a single `kid` ensures verifiers can validate either classical Ed25519 signatures (low-overhead edge verification) or full ML-DSA-44 signatures (quantum audit) using a unified key lookup.

2. **Mitigating "Token Bloat" via Token Triad:**
   - ML-DSA-44 signatures are 2,420 bytes long. Base64url encoding adds 33% overhead (~3,227 chars). Embedding this in standard JWTs creates Jumbo tokens (~4 KB), exceeding default proxy and web server HTTP header buffers (typically 4–8 KB).
   - DilithiAuth solves this through two architectural paths:
     - **Path A (Sidecar Optimization):** Thin JWTs replace the 3.2 KB `pqc_sig` with a 43-character `pqc_ref` SHA-256 hash. The signature is offloaded to `SigStore`. This reduces HTTP header size from ~4 KB down to ~400 bytes.
     - **Path B (Binary Compression):** CWT replaces JSON string keys (e.g. `"iss"`, `"sub"`, `"exp"`) with integer keys (`1`, `2`, `4`), replaces Base64url text encoding with raw binary CBOR bytes, and replaces Base64url signature strings with 64 raw bytes in a CBOR array. This drops token size to 197 bytes (~50.8% reduction vs Thin JWT).
     - **Path C (Stateless Zero-Trust):** Jumbo Tokens embed the full signature for zero-trust environments where sidecar stores cannot be reached, supported by setting server `MaxHeaderBytes` to 16 KiB.

---

## 3. Caveats

- **Demo Route Compilation:** The endpoints `POST /token/thin`, `POST /token/full`, and `POST /token/cwt` in `cmd/server/demo_routes.go` are guarded by `//go:build demo`. Production builds use `demo_routes_stub.go` (no-op) and issue Thin JWTs via standard OAuth 2.0 `POST /token` authorization code exchange.
- **CWT Transport:** In the demo endpoint `POST /token/cwt`, CWT tokens are returned directly in the HTTP body as raw binary (`Content-Type: application/cbor`). When passed in HTTP `Authorization` headers in HTTP/1.1 environments, binary CWT tokens require binary-safe transport or custom headers (`X-Token-Id`).
- **SigStore TTL:** Signatures in `SigStore` must match the token expiration (`tokenTTL` = 1 hour). If `SigStore` entries expire prematurely, thin token quantum verification fails.

---

## 4. Conclusion

DilithiAuth effectively solves the Post-Quantum Cryptography token bloat challenge through a hybrid signing architecture and three distinct token issuance strategies:
1. **Thin Tokens:** The optimal default for web & microservices, keeping HTTP headers under 500 bytes while retaining a PQC audit trail in a sidecar (`SigStore`).
2. **Jumbo Tokens:** Self-contained, stateless 4 KB tokens for offline or zero-trust validation, enabled by 16 KiB HTTP server header limits.
3. **CBOR Web Tokens (CWT):** The most compact format (197 bytes, >50% smaller than Thin JWT), ideal for bandwidth-constrained networks and IoT.

---

## 5. Verification Method

### 5.1 Command Line Verification
Run the project Go tests to verify discovery document structure, JWKS key publication, token issuance, and empirical payload sizes:

```bash
# Run all oauth, signer, and server tests
go test -v ./pkg/oauth ./pkg/signer ./cmd/server
```

### 5.2 Specific Size Measurement Test
Run the CWT vs JWT size comparison unit test:

```bash
go test -v ./pkg/oauth -run TestCWTSmallerThanJWT
```

**Expected Test Output:**
```
=== RUN   TestCWTSmallerThanJWT
    oauth_test.go:220: CWT size: 197 bytes, JWT (thin) size: 401 bytes
--- PASS: TestCWTSmallerThanJWT (0.00s)
```

### 5.3 File Inspection
Inspect code locations to verify implementations:
- Discovery & JWKS: `pkg/oauth/oauth.go` (lines 23–142)
- Thin, Jumbo, CWT Issuance: `pkg/oauth/oauth.go` (lines 180–393)
- Demo Endpoints: `cmd/server/demo_routes.go` (lines 17–96)
- Standard Token Exchange: `pkg/oauth/authz_handlers.go` (lines 1003–1156)
- 16 KiB Header Config: `cmd/server/main.go` (lines 368, 549)
