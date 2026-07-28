## 2026-07-28T21:02:31Z
You are Worker M5 for Milestone 5 (R4 App Detail Page redesign in AppDetail.tsx).
Your working directory is `/home/uttam/projects/DilithiAuth/.agents/worker_m5_gen2`.
Your parent orchestrator workspace is `/home/uttam/projects/DilithiAuth/.agents/orchestrator`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Your Task:
Implement the redesign of `AppDetail.tsx` and subcomponents according to the Explorer blueprint in `/home/uttam/projects/DilithiAuth/.agents/explorer_m5_gen2/blueprint.md` and `handoff.md`.

Specifically:
1. Create subcomponents under `frontend/src/components/app-detail/`:
   - `SettingsTab.tsx`: Clean editable form groups for App Name, Description, Homepage URL, multi-row Redirect URIs manager (with Add URI and Remove URI buttons), Client ID with copy button, Client Type indicator (Public SPA vs Confidential Server App), and Save Changes button.
   - `SnippetsTab.tsx`: Code snippets viewer with language tabs (cURL, JS / Fetch, Go, Python) featuring token endpoints (`/token/thin`, `/token/full`, `/token/cwt`), copy code action, and dark theme code block styling.
   - `DangerZoneTab.tsx`: Redesigned Danger Zone container. "Rotate Client Secret" button (opens `SecretDisplayModal.tsx`) and "Delete Application" button (opens `DeleteAppModal.tsx`).
   - `SecretDisplayModal.tsx`: Secret rotation success modal displaying Client ID and newly generated Client Secret with 1-click copy buttons and amber warning banner.
   - `DeleteAppModal.tsx`: Confirmation modal requiring user to type the application name before confirming deletion via `apps.delete(id)`.
2. Refactor `frontend/src/pages/AppDetail.tsx`:
   - Page header with App Name, Client ID badge, App Type badge (Public SPA PKCE vs Confidential Client), and back button to `/apps`.
   - Sub-navigation tabs bar: **Settings**, **Code Snippets**, **Danger Zone**.
   - Integrates `SettingsTab`, `SnippetsTab`, `DangerZoneTab`, `SecretDisplayModal`, and `DeleteAppModal`.
   - Ensures ZERO imports or references to `Navbar.tsx`.
   - Works seamlessly inside `<SidebarLayout />`.
3. Create test harness `frontend/src/__tests__/m5_verification_test.tsx` and `run_m5_test.mjs` (Vite SSR test pattern), and run tests to verify tab switching, form editing, secret rotation, and deletion modals.
4. Verify build (`npm run build` with 0 TS/Vite errors).

Deliver your implementation report to `/home/uttam/projects/DilithiAuth/.agents/worker_m5_gen2/handoff.md` and send a message back when completed.
