## 2026-07-27T19:25:07Z
<USER_REQUEST>
Identity: teamwork_preview_worker
Working Directory: /home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m2

Objective: Create the `paper/` directory, draft `paper/references.bib`, and draft the complete IEEEtran LaTeX research paper `paper/main.tex`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Input Data to Incorporate:
- Empirical survey & benchmark results from `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m1/handoff.md`, `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_1/handoff.md`, `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_2/handoff.md`, and `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_3/handoff.md`.
- Key Empirical Figures to inclusion:
  - Cryptographic Primitives: Ed25519 (32B pub, 64B sig, 28.72 µs sign, 36.73 µs verify); ML-DSA-44 (1,312B pub, 2,420B sig, 121.58 µs sign, 27.94 µs verify); Hybrid (2,484B sig, 156.57 µs sign, 176.54 µs verify).
  - Token Strategies: Thin JWT (419B token, 441B HTTP header, 276.05 µs issue, 65.31 µs verify, 43.37 µs fast path); Jumbo JWT (4,665B token, 4,687B HTTP header, 194.89 µs issue, 108.33 µs verify); CWT (211B binary / 282B base64url, 304B HTTP header, 257.35 µs issue, 58.20 µs verify).
  - Header Limits & Sidecar Store: `MaxHeaderBytes = 16 KiB`, `SigStore` sidecar entry overhead (2,463B for Thin JWT, 2,442B for CWT).

Tasks:
1. Create `paper/` directory.
2. Write `paper/references.bib` with comprehensive BibTeX citations including:
   - FIPS 204 (NIST ML-DSA)
   - RFC 8032 (Ed25519)
   - RFC 8392 (CWT)
   - RFC 7519 (JWT)
   - RFC 6749 & RFC 6750 (OAuth 2.0)
   - OpenID Connect Core 1.0
   - NIST SP 800-208 / NIST PQC Standards
   - HNDL threat literature and related post-quantum identity papers.
3. Write `paper/main.tex` using `\documentclass[conference]{IEEEtran}` with packages `cite`, `amsmath`, `amssymb`, `graphicx`, `booktabs`, `array`, `url`, `microtype`, `hyperref`.
4. Include all 7 mandatory sections:
   I. Abstract
   II. Introduction (HNDL threat model, token bloat in HTTP headers)
   III. Background & Related Work (NIST PQC FIPS 204 ML-DSA-44, signature size comparison, OAuth 2.0 / OIDC)
   IV. DilithiAuth System Architecture (`pkg/signer` hybrid primitives, `pkg/oauth` pipeline, `pkg/transport` sidecar store)
   V. Token Strategy Trade-off Analysis (Mathematical modeling, empirical tables comparing Thin JWT, Jumbo JWT, CWT, size ratios, latency, memory allocations, header limits)
   VI. Security Evaluation (Post-quantum lattice security, sidecar integrity verification via `pqc_ref`, replay protection)
   VII. Conclusion & Future Directions
5. Verify LaTeX code syntax, section cross-references, cite keys, and table formatting.
6. Record completion details in `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m2/handoff.md`.
7. Send a message to parent when finished.
</USER_REQUEST>
