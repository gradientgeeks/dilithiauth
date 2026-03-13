# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test

```bash
go build ./cmd/server                              # build the server binary
go test ./...                                       # run all tests
go test ./pkg/oauth -run TestIssueJWTThin           # run a single test
```

No Makefile or linter configuration exists — use standard Go tooling.

## Environment Variables

- `DILITHIAUTH_ISSUER` — base URL of the issuer (default: `https://auth.dilithiauth.io`)
- `DILITHIAUTH_ADDR` — listen address (default: `:8080`)

## Architecture

DilithiAuth is a post-quantum OAuth 2.0 identity provider that dual-signs tokens with Ed25519 (classical) and ML-DSA-44 (quantum-safe, FIPS 204). It mitigates "Harvest Now, Decrypt Later" attacks.

- **cmd/server** — entrypoint. Generates hybrid key pair, creates `oauth.Provider`, wires chi router with middleware, and exposes OIDC discovery/JWKS plus `/token/{thin,full,cwt}` handlers.
- **pkg/signer** — hybrid key pair (Ed25519 + ML-DSA-44) under a shared `kid`, with sign/verify helpers and JWKS public key export.
- **pkg/oauth** — OIDC discovery/JWKS endpoints and three token issuance strategies: thin JWT (Ed25519 only, PQC sig in sidecar), jumbo JWT (full PQC sig embedded), and CWT (CBOR-encoded, smallest).
- **pkg/transport** — PQC-tuned HTTP server (16 KiB `MaxHeaderBytes`), jumbo-header middleware, and in-memory `SigStore` sidecar with context helpers.

## Key Conventions

- Both Ed25519 and ML-DSA-44 keys share a single `kid` in JWKS.
- Thin tokens carry `pqc_ref` (SHA-256 of the ML-DSA-44 signature) and store the full signature in `SigStore` keyed by `jti`/`cti`. Full/jumbo tokens embed the signature directly in `pqc_sig`.
- CWTs use RFC 8392 integer claim keys and a `[claims_bytes, ed25519_sig]` CBOR envelope.
- `transport.NewPQCServer` sets 16 KiB header limit to accommodate jumbo JWTs; the nginx config in `deployments/` mirrors this.
- `SigStore` is in-memory (dev); production should use a distributed store.

## Dependencies

- `cloudflare/circl` — ML-DSA-44 post-quantum crypto
- `fxamacker/cbor/v2` — CBOR encoding for CWTs
- `go-chi/chi/v5` — HTTP routing
- `golang-jwt/jwt/v5` — JWT signing/parsing
