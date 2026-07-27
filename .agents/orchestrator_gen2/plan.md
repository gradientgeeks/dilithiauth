# Plan: Social Authentication Integration (Milestone M6)

## Objectives
1. Implement live Social Authentication HTTP endpoints (`/auth/social/google`, `/auth/social/google/callback`, `/auth/social/github`, `/auth/social/github/callback`) in `pkg/oauth/` and `cmd/server/main.go`.
2. Wire environment variables (`DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET`) in `cmd/server/main.go`.
3. Update `docs/oci-vm-deployment.md` and `deployments/dilithiauth.service` with environment variable placeholders.
4. Comprehensive unit testing and verification (`go test ./...` and `go build -o dilithiauth-server ./cmd/server`).

## Milestones & Execution Steps

### Phase 1: Multi-Agent Exploration
- **Explorer 1**: Analyze `pkg/oauth/` (handlers, sessions, state validation, token issuance, userstore).
- **Explorer 2**: Analyze `cmd/server/main.go` (routing, server initialization, environment configuration).
- **Explorer 3**: Analyze Google & GitHub OAuth token/profile schemas, systemd service, and deployment docs.
- **Aggregation**: Reconcile findings into an implementation blueprint.

### Phase 2: Implementation & Unit Testing
- **Worker**: Implement social handlers, state cookie/session validation, token exchange with Google/GitHub, claim normalization (`sub`, `email`, `name`), user provisioning/lookup, token generation (hybrid Ed25519 + ML-DSA-44), environment variable wiring, systemd/docs updates, and comprehensive Go unit tests.
- **Verification**: Worker executes `go test ./...` and `go build -o dilithiauth-server ./cmd/server`.

### Phase 3: Review, Challenge & Integrity Audit
- **Reviewers (2)**: Perform independent code reviews focusing on OAuth security, CSRF prevention, error handling, code structure, and test coverage.
- **Challengers (2)**: Perform empirical testing of endpoints, state validation, redirect behavior (302), credential checks, and mock provider interactions.
- **Forensic Auditor (1)**: Conduct forensic integrity audit to verify no hardcoded responses, facade logic, or test cheats.

### Phase 4: Gate Evaluation & Handoff
- Evaluate all pass criteria (clean audit, passing tests/builds, positive review/challenge verdicts).
- Generate `handoff.md` and notify parent.
