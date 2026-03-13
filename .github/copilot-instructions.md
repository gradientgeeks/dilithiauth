# Copilot Instructions for DilithiAuth

## Build and test
- Build the server: `go build ./cmd/server`
- Run all tests: `go test ./...`
- Run a single test (example): `go test ./pkg/oauth -run TestIssueJWTThin`

## Architecture
- `cmd/server` is the entrypoint. It reads `DILITHIAUTH_ISSUER` and `DILITHIAUTH_ADDR`, generates the hybrid key pair, builds an `oauth.Provider`, wires chi middleware, and exposes discovery/JWKS plus `/token/*` handlers.
- `pkg/signer` owns the hybrid key pair (Ed25519 + ML-DSA-44) under a shared `kid`, with sign/verify helpers and JWKS public key material.
- `pkg/oauth` serves OIDC discovery/JWKS and issues tokens: thin JWTs, full jumbo JWTs, and CWTs. It computes `pqc_ref` from ML-DSA-44 signatures via SHA-256.
- `pkg/transport` provides PQC server defaults (16 KiB header limit), jumbo-header middleware, and the in-memory `SigStore` sidecar with context helpers.

## Key conventions
- Hybrid tokens use a shared `kid` for Ed25519 and ML-DSA-44; JWKS publishes both under the same key.
- Thin tokens include `pqc_ref` and store the full ML-DSA-44 signature in `SigStore` keyed by `jti`/`cti`; full tokens embed `pqc_sig`.
- CWTs are encoded as a CBOR envelope `[claims_bytes, ed25519_sig]` with `pqc_ref` included in the claims.
- Full/jumbo JWTs assume the server is configured via `transport.NewPQCServer` to accept 16 KiB headers.
