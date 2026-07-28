# Original User Request

## 2026-07-27T19:20:29Z

Research DilithiAuth's post-quantum identity architecture (Ed25519 + ML-DSA-44 hybrid signatures, Token bloat mitigation, Thin/Jumbo JWTs, and CWT), write a formal academic research paper in IEEEtran LaTeX format inside the `paper/` directory, and compile it to PDF using local `pdflatex`.

Working directory: `/home/uttam/projects/DilithiAuth`
Integrity mode: `development`

## Requirements

### R1. Technical & Cryptographic Codebase Survey
Survey DilithiAuth's Go codebase (`pkg/signer`, `pkg/oauth`, `pkg/transport`, `cmd/server`) and analyze:
- **Harvest Now, Decrypt Later (HNDL) Threat Model**: Mitigating long-term cryptographic compromise of classical OIDC token signatures using NIST FIPS 204 ML-DSA-44 (Dilithium).
- **Hybrid Signatures**: Dual signing strategy (`HybridKeyPair` combining classical Ed25519 64B signatures with post-quantum ML-DSA-44 2,420B signatures).
- **Token Strategies & Optimization**:
  1. *Thin Tokens*: Lightweight JWT containing Ed25519 signature + `pqc_ref` cryptographic hash referencing full ML-DSA signature in sidecar storage (Redis / memory).
  2. *Jumbo Tokens*: Self-contained ~4KB JWT embedding full ~2.4KB ML-DSA-44 signature directly in `pqc_sig` claim.
  3. *CBOR Web Tokens (CWT)*: Binary RFC 8392 format reducing token overhead by 20–30%.
- **Large Header Middleware**: Transport middleware managing `MaxHeaderBytes = 16 KiB` and sidecar resolution.

### R2. Formal Academic Research Paper in IEEEtran LaTeX
Create the research paper in directory `paper/` with the following structure:
- `paper/main.tex`: Complete IEEEtran document (`\documentclass[conference]{IEEEtran}`) with sections:
  1. *Abstract*: Problem statement, hybrid PQC approach, key empirical results.
  2. *Introduction*: Quantum threats to identity providers, token bloat challenges in HTTP headers.
  3. *Background & Related Work*: NIST PQC standardization (FIPS 204 ML-DSA-44), classical vs. PQC signature sizes, current identity protocols (OAuth 2.0 / OIDC).
  4. *DilithiAuth System Architecture*: High-level design, `pkg/signer` hybrid primitives, `pkg/oauth` token issuance & validation pipeline, `pkg/transport` sidecar architecture.
  5. *Token Strategy Trade-off Analysis*: Mathematical model & benchmark comparison (payload sizes, HTTP header overhead, storage latency vs. verification self-containment).
  6. *Security Evaluation*: Post-quantum security analysis, sidecar integrity verification via `pqc_ref`, replay protection.
  7. *Conclusion & Future Directions*: Summary of findings and recommendations for production deployment.
- `paper/references.bib`: Complete BibTeX bibliography with citations for FIPS 204, RFC 8392, OAuth 2.0, NIST PQC standards, and related literature.

### R3. Automated PDF Compilation
- Ensure `paper/main.tex` compiles cleanly with `pdflatex -interaction=nonstopmode main.tex` and `bibtex main`.
- Verify the generated `paper/main.pdf` is output in `paper/` without fatal LaTeX errors.

## Acceptance Criteria

### Verification & Deliverables
- [ ] `paper/main.tex` and `paper/references.bib` are created with comprehensive technical content and valid LaTeX formatting.
- [ ] Running `pdflatex -interaction=nonstopmode main.tex` inside `paper/` produces `paper/main.pdf`.
- [ ] The paper covers all three token strategies (Thin, Jumbo, CWT), hybrid cryptographic primitives, and empirical benchmark data.
