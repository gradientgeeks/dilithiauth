## 2026-07-28T20:51:44Z
<USER_REQUEST>
You are Worker M4 for Milestone 4 (R3 Create Application Page - NewApp.tsx Wizard Redesign).
Your working directory is `/home/uttam/projects/DilithiAuth/.agents/worker_m4_gen2`.
Your parent orchestrator workspace is `/home/uttam/projects/DilithiAuth/.agents/orchestrator`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Your Task:
Implement the redesign of `NewApp.tsx` and subcomponents according to the Explorer blueprint in `/home/uttam/projects/DilithiAuth/.agents/explorer_m4_gen2/blueprint.md` and `handoff.md`.

Specifically:
1. Create `frontend/src/components/new-app/WizardStepper.tsx`:
   - Visual step indicator at top: Step 1 Basic Info -> Step 2 Application Type -> Step 3 URIs.
   - Shows completed step checkmarks, active step highlight, and step numbers.
2. Create `frontend/src/components/new-app/TypeSelector.tsx`:
   - Visual card selector for Public SPA (PKCE) vs Confidential Server-Side App.
   - Distinct icons (`Lock` for Confidential, `Globe` for Public SPA), architecture badges (`Client Secret Flow` vs `PKCE Flow`), descriptions, active border/ring selection state styled with Tailwind v4 design tokens (`--color-card`, `--color-border`, `--color-primary`, `--color-cyan`).
3. Create `frontend/src/components/new-app/SecretSuccessModal.tsx`:
   - Post-creation modal/card overlay triggered when `apps.create()` returns.
   - Displays Client ID and Client Secret in styled input fields with 1-click Copy buttons.
   - Amber warning banner ("Save your Client Secret now. It will not be shown again!").
   - Navigation button "Continue to App Details" directing to `/apps/${app.ID}`.
4. Refactor `frontend/src/pages/NewApp.tsx`:
   - Step 1: Basic Info (App Name, Description, Homepage URL). Validation for required Name.
   - Step 2: Application Type (`TypeSelector.tsx`).
   - Step 3: URIs & Configuration (Redirect URIs manager with Add/Remove input rows, default `http://localhost:3000/callback`).
   - Step navigation buttons (Back, Next / Submit "Create Application").
   - Triggers `SecretSuccessModal.tsx` upon successful API response.
   - Preserves all backend API payload fields: `{ name, description, homepage_url, redirect_uris, is_public }`.
   - Ensures ZERO imports or references to `Navbar.tsx`.
5. Create test harness `frontend/src/__tests__/m4_verification_test.tsx` and `run_m4_test.mjs` (referencing Vite SSR test pattern from `run_m3_test.mjs`), and run test scripts to verify component rendering and wizard transitions.
6. Verify build (`npm run build` or Vite compile checks).

Deliver your implementation report to `/home/uttam/projects/DilithiAuth/.agents/worker_m4_gen2/handoff.md` and send a message back when completed.
</USER_REQUEST>
