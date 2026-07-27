# Handoff Report — Explorer 2 (Backend API & PQC Token Mechanics)

## 1. Observation

Direct observations from the DilithiAuth codebase:

1. **Backend Endpoints**:
   - `cmd/server/main.go:443-444`: `GET /.well-known/openid-configuration` (Discovery) and `GET /.well-known/jwks.json` (JWKS).
   - `cmd/server/main.go:457-466`: `/authorize` (`GET`/`POST`), `/register` (`GET`/`POST`), `/forgot-password`, `/reset-password`, `/consent`, `/token`.
   - `cmd/server/main.go:469`: `/revoke` (`POST` RFC 7009 token revocation).
   - `cmd/server/main.go:472`: `/verify/jwt` (`POST` REST verification endpoint).
   - `cmd/server/main.go:491`: `/userinfo` (`GET` OIDC protected identity endpoint).
   - `cmd/server/demo_routes.go:23-25`: Unauthenticated demo token endpoints `/token/thin`, `/token/full`, `/token/cwt` compiled under `-tags demo`.

2. **JWKS Key Parameters**:
   - `pkg/oauth/oauth.go:108-117`: JWK structure returns `kty: "OKP"`, `crv: "Ed25519"`, `x` (base64url Ed25519 public key), `mldsa44_pub` (base64url ML-DSA-44 public key), `alg: "EdDSA+MLDSA44"`.

3. **Post-Quantum Token Strategies**:
   - **Thin JWT**: `pkg/oauth/oauth.go:183-240` (`IssueJWTThinWithScope`) signs payload with Ed25519 and ML-DSA-44, embeds `pqc_ref = base64url(SHA256(mldsa_sig))`, returns small JWT (<500 bytes) + ML-DSA-44 signature (~2420 bytes) for SigStore sidecar.
   - **Jumbo / Full JWT**: `pkg/oauth/oauth.go:245-295` (`IssueJWTFullWithScope`) embeds full ML-DSA-44 signature directly in `pqc_sig` claim (~3.5-4KB total token).
   - **CWT (CBOR Web Token)**: `pkg/oauth/oauth.go:336-392` (`IssueCWTWithScope`) encodes claims in binary CBOR using integer keys (`1:iss`, `2:sub`, `4:exp`, `6:iat`, `7:cti`, `8:pqc_ref`, `9:tid`, `10:scope`), enveloped in `[claims_bytes, ed25519_sig]`, PQC signature in SigStore keyed by `cti`. Returns binary `application/cbor` content (~200-300 bytes).

4. **16 KiB Header Requirement**:
   - `pkg/transport/middleware.go:18`: `MaxHeaderBytesForPQC = 1 << 14` (16 KiB).
   - `cmd/server/main.go:527`: `srv := transport.NewPQCServer(addr, r)` sets `MaxHeaderBytes: 16 KiB`.

5. **5 SDK Implementations**:
   - JavaScript/TypeScript: `sdk/js/src/DilithiAuthClient.ts`, `pkce.ts`, `types.ts` (`dilithiauth-js`).
   - Go: `sdk/go/client.go` (`github.com/DilithiAuth/dilithiauth-go`).
   - Python: `sdk/python/dilithiauth/client.py`, `pkce.py` (`dilithiauth`).
   - Java: `sdk/java/src/main/java/com/dilithiauth/client/DilithiAuthClient.java`, `PKCE.java` (`com.dilithiauth:dilithiauth-sdk`).
   - .NET: `sdk/dotnet/DilithiAuth.Client/DilithiAuthClient.cs`, `Pkce.cs`, `Models.cs` (`DilithiAuth.SDK` / `DilithiAuth.Client`).

---

## 2. Logic Chain

1. **Premise**: Milestone M1 requires an exhaustive investigation of backend API specifications, Post-Quantum token mechanics, and authentic code snippets across 5 SDK languages for the interactive SDK Documentation page (`frontend/src/pages/Docs.tsx`).
2. **Step 1**: Traced server endpoints in `cmd/server/main.go` and `pkg/oauth/authz_handlers.go` to document request/response formats, query parameters, headers, rate limits, and scopes for all 8 standard + 3 demo endpoints.
3. **Step 2**: Analyzed cryptographic mechanics in `pkg/signer/signer.go`, `pkg/oauth/oauth.go`, and `pkg/transport/middleware.go` to formalize the three token issuance strategies (Thin, Jumbo, CWT), explaining why ML-DSA-44 (~2,420 bytes) creates token bloat and how SigStore vs Jumbo headers mitigate it.
4. **Step 3**: Analyzed source code across all 5 language directories (`sdk/js`, `sdk/go`, `sdk/python`, `sdk/java`, `sdk/dotnet`) to construct authentic, working code snippets for SDK initialization, PKCE authorization URL generation, authorization code exchange, and user claims retrieval.
5. **Conclusion**: Full specification produced in `analysis.md`, providing workers in M2/M3 with exact data to build `frontend/src/pages/Docs.tsx`.

---

## 3. Caveats

- Demo routes (`/token/thin`, `/token/full`, `/token/cwt`) are only active when compiled with `-tags demo`; standard production builds use standard OAuth `/token` endpoint with `grant_type=authorization_code`.
- Thin tokens require a running `SigStore` backend (in-memory, file, PostgreSQL, or Redis) on the server to preserve PQC signature references.

---

## 4. Conclusion

All backend API endpoints, token mechanics (Thin, Jumbo, CWT), and SDK code snippets across 5 languages (TypeScript, Go, Python, Java, .NET) have been fully investigated and documented in `.agents/teamwork_preview_explorer_m1_2/analysis.md`.

---

## 5. Verification Method

Independent verification can be executed via:

1. **Go Tests**:
   ```bash
   go test ./...
   ```
2. **File Inspection**:
   - `analysis.md` in `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_2/analysis.md`
   - `pkg/oauth/oauth.go` for token issuance strategies
   - `sdk/` subdirectories for multi-language SDK code
