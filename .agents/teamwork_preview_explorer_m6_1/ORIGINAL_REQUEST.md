## 2026-07-27T18:11:34Z
<USER_REQUEST>
You are teamwork_preview_explorer_m6_1. Your working directory is /home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_1.

Objective:
Investigate `pkg/oauth/` codebase in DilithiAuth (/home/uttam/projects/DilithiAuth) for Social Authentication integration.

Tasks:
1. Examine existing HTTP handlers and server structure in `pkg/oauth/` (`authz_handlers.go`, `server.go`, `token.go`, `jwks.go`, etc.).
2. Analyze how DilithiAuth handles sessions, state validation, user store / account lookup, and token generation (hybrid Ed25519 + ML-DSA-44 tokens).
3. Design the architecture and implementation plan for Social Authentication HTTP handlers (`/auth/social/{provider}` and `/auth/social/{provider}/callback` for Google and GitHub):
   - State generation and validation (encrypted cookie or signed state) to prevent CSRF attacks.
   - Code exchange flow for Google (`https://oauth2.googleapis.com/token`) and GitHub (`https://github.com/login/oauth/access_token`).
   - Claim normalization into common structure (`sub`, `email`, `name`).
   - Account provisioning / retrieval in DilithiAuth userstore.
   - Issuing Post-Quantum Hybrid Access Token and performing HTTP 302 redirect to `redirect_uri` with authorization code.
4. Write your findings to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_1/analysis.md` and write a detailed handoff report to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_1/handoff.md`.
5. Send a completion message to the parent orchestrator.
</USER_REQUEST>
