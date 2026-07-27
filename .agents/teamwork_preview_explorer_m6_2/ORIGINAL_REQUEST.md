## 2026-07-27T18:11:34Z
<USER_REQUEST>
You are teamwork_preview_explorer_m6_2. Your working directory is /home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_2.

Objective:
Investigate `cmd/server/main.go` and route/env configuration in DilithiAuth (/home/uttam/projects/DilithiAuth) for Social Authentication integration.

Tasks:
1. Examine `cmd/server/main.go` to understand how HTTP routes, middleware, and handlers are wired up.
2. Analyze how environment variables (e.g. `DILITHIAUTH_ISSUER`, `DILITHIAUTH_ADDR`) are parsed and passed into handlers or servers.
3. Plan how `DILITHIAUTH_GOOGLE_CLIENT_ID`, `DILITHIAUTH_GOOGLE_CLIENT_SECRET`, `DILITHIAUTH_GITHUB_CLIENT_ID`, `DILITHIAUTH_GITHUB_CLIENT_SECRET` should be read in `cmd/server/main.go` and injected into social auth handlers.
4. Design the exact HTTP route definitions for `/auth/social/google`, `/auth/social/google/callback`, `/auth/social/github`, `/auth/social/github/callback`.
5. Write your findings to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_2/analysis.md` and write a detailed handoff report to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m6_2/handoff.md`.
6. Send a completion message to the parent orchestrator.
</USER_REQUEST>
