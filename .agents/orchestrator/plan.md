# Concrete Execution Plan — DilithiAuth Research Paper

## Milestone 1: Technical Survey & Empirical Benchmarking
- [ ] Step 1.1: Dispatch 3 Explorer subagents to analyze Go codebase (`pkg/signer`, `pkg/oauth`, `pkg/transport`, `cmd/server`) and examine benchmark suites/tests.
- [ ] Step 1.2: Dispatch Worker subagent to execute Go benchmarks, collect precise token payload sizes (Thin JWT, Jumbo JWT, CWT), timing data (signing/verification), and storage/header overheads.
- [ ] Step 1.3: Aggregate benchmark results and architectural findings into a structured technical survey report.

## Milestone 2: IEEEtran LaTeX Paper & BibTeX Bibliography Drafting
- [ ] Step 2.1: Dispatch Worker subagent to create `paper/` directory and draft `paper/references.bib` with comprehensive BibTeX citations.
- [ ] Step 2.2: Dispatch Worker subagent to draft `paper/main.tex` formatted for `\documentclass[conference]{IEEEtran}` covering all 7 core sections:
  1. Abstract
  2. Introduction
  3. Background & Related Work
  4. DilithiAuth System Architecture
  5. Token Strategy Trade-off Analysis (with empirical benchmark tables)
  6. Security Evaluation
  7. Conclusion & Future Directions
- [ ] Step 2.3: Verify paper completeness and formatting.

## Milestone 3: PDF Compilation, Verification & Forensic Audit
- [ ] Step 3.1: Dispatch Worker subagent to run `pdflatex -interaction=nonstopmode main.tex` and `bibtex main` inside `paper/` to build `paper/main.pdf`.
- [ ] Step 3.2: Dispatch 2 Reviewer subagents to evaluate academic rigor, technical accuracy, LaTeX syntax, and complete section coverage.
- [ ] Step 3.3: Dispatch 2 Challenger subagents to verify clean PDF compilation, check for compilation warnings/errors, and validate empirical numbers in text.
- [ ] Step 3.4: Dispatch Forensic Auditor (`teamwork_preview_auditor`) to verify zero integrity violations.
- [ ] Step 3.5: Synthesize results and report completion to Sentinel.
