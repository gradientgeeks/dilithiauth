# Project: DilithiAuth Post-Quantum Identity Architecture Research Paper

## Architecture
DilithiAuth is a Post-Quantum Identity Provider (IdP) implementing Ed25519 + ML-DSA-44 (FIPS 204) hybrid signatures to protect OAuth 2.0 / OIDC tokens against Harvest Now, Decrypt Later (HNDL) attacks.
Key components of the codebase:
- `pkg/signer`: Hybrid key pair generation (`HybridKeyPair`) and dual-signing primitives combining Ed25519 (64B signature) and ML-DSA-44 (2,420B signature).
- `pkg/oauth`: OIDC Discovery, JWKS (including `mldsa44_pub`), Thin JWT issuance/validation, Jumbo JWT issuance/validation, and CBOR Web Tokens (CWT).
- `pkg/transport`: Middleware enforcing `MaxHeaderBytes = 16 KiB` and sidecar reference store (`SigStore`) for Thin Token `pqc_ref` resolution.
- `cmd/server`: HTTP server entry point exposing OIDC endpoints.

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| 1 | Codebase Survey & Empirical Benchmarks | Conduct technical survey of `pkg/signer`, `pkg/oauth`, `pkg/transport`, `cmd/server` and execute Go benchmark tests to extract empirical performance data (token sizes, signing/verification latency, header overhead). | none | DONE |
| 2 | IEEEtran LaTeX Paper & Bibliography Drafting | Create `paper/main.tex` (7 sections: Abstract, Intro, Background, Architecture, Trade-offs, Security, Conclusion) and `paper/references.bib` with citations. | M1 | DONE |
| 3 | PDF Compilation, Verification & Audit | Compile `paper/main.pdf` using `pdflatex` & `bibtex`, run Reviewers, Challengers, and Forensic Auditor for verification. | M2 | DONE |

## Interface Contracts & Artifact Standards
- IEEEtran Document Class: `\documentclass[conference]{IEEEtran}` in `paper/main.tex`.
- Bibliography: `paper/references.bib` using standard BibTeX entries (FIPS 204, RFC 8392, RFC 7519, RFC 6749, NIST PQC standards, etc.).
- PDF output: `paper/main.pdf` compiled cleanly with `pdflatex -interaction=nonstopmode main.tex` and `bibtex main`.
- Technical coverage: Thin Tokens (`pqc_ref`), Jumbo JWTs (`pqc_sig`), CWT (CBOR), Hybrid `HybridKeyPair`, `MaxHeaderBytes = 16 KiB`, HNDL threat model.

## Code Layout
- `paper/main.tex`
- `paper/references.bib`
- `paper/main.pdf`
- `pkg/signer/`
- `pkg/oauth/`
- `pkg/transport/`
- `cmd/server/`
