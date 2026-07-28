# Context & Decisions — DilithiAuth Research Paper

## Project Goal
Produce a formal academic research paper in IEEEtran LaTeX format (`paper/main.tex`, `paper/references.bib`, `paper/main.pdf`) analyzing DilithiAuth's post-quantum identity provider architecture, hybrid Ed25519 + ML-DSA-44 signatures, token strategies (Thin JWT, Jumbo JWT, CWT), and large header middleware.

## Requirements Summary
- R1: Technical & Cryptographic Survey of Go codebase (`pkg/signer`, `pkg/oauth`, `pkg/transport`, `cmd/server`).
- R2: Formal Academic Research Paper in IEEEtran LaTeX (`paper/main.tex` with 7 sections + `paper/references.bib`).
- R3: Automated PDF Compilation (`paper/main.pdf` generated cleanly with `pdflatex` & `bibtex`).

## Architectural Decisions
- Topology: Decomposed into 3 sequential milestones.
- Benchmarks: Real empirical numbers to be measured from `pkg/signer`, `pkg/oauth`, `pkg/transport` unit/benchmark tests.
- Audit & Verification: Gated by Forensic Auditor binary veto.
