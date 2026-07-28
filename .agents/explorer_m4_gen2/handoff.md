# Handoff Report: Milestone 4 (R3 Create Application Page - NewApp.tsx Wizard Redesign)

## 1. Observation

- **Current Implementation**:
  - `frontend/src/pages/NewApp.tsx` (225 lines) is a monolithic single-page form inside `SidebarLayout`.
  - Form state: `name`, `description`, `homepageURL`, `redirectURIs` (newline-separated string), `isPublic` (boolean), `error`, `loading`, `created` (inline success card), `copied`.
  - `apps.create` API payload: `{ name, description, homepage_url, redirect_uris: string[], is_public: boolean }`.
  - Return type: `{ data: { app: OAuthApp, client_secret?: string } }`.
- **Layout & Navigation Context**:
  - In `frontend/src/App.tsx` line 40: `<Route element={<RequireAuth><SidebarLayout /></RequireAuth>}> <Route path="/apps/new" element={<NewApp />} /> </Route>`.
  - `<SidebarLayout />` provides top header (mobile) and sidebar (desktop). Main page content renders via `<Outlet />` inside `<main className="flex-1 min-w-0 p-4 sm:p-6 md:p-8 overflow-y-auto">`.
  - Zero imports or usages of `Navbar.tsx` in `NewApp.tsx` or `SidebarLayout.tsx`.
- **Design Tokens & System**:
  - `frontend/src/index.css` Tailwind v4 design tokens: `--color-bg` (#090d16), `--color-card` (#0f1623), `--color-border` (#1e2a40), `--color-border-mid` (#263347), `--color-primary` (#3b82f6), `--color-cyan` (#06b6d4), `--color-amber` (#f59e0b).
  - Pre-styled component helper classes: `.secret-box`, `.code-block`, `.mono-inline`, `.spin`.
- **Test Infrastructure**:
  - `node frontend/src/__tests__/run_m3_test.mjs` runs Vite SSR tests cleanly (26 passed, 0 failed).

---

## 2. Logic Chain

1. **Observation**: `NewApp.tsx` currently collects basic info, client type selection, homepage URL, description, and redirect URIs all in one form container.
   - **Reasoning**: Grouping these into a 3-step wizard improves user flow, reduces cognitive overload, and aligns with modern developer portals like Auth0 Console.
   - **Step 1**: Basic Info (Name, Description, Homepage URL).
   - **Step 2**: Application Type (`TypeSelector.tsx` for Confidential Client vs Public SPA PKCE Client).
   - **Step 3**: URIs & Configuration (Interactive URI list manager with Add/Remove rows).

2. **Observation**: Client type selector currently uses two small inline buttons inside `NewApp.tsx` lines 158-171.
   - **Reasoning**: Extracting client type selection into a dedicated `TypeSelector.tsx` component under `src/components/new-app/` allows rich visual cards with icons (`Lock`, `Globe`), architecture tags (`PKCE Flow`, `Client Secret Flow`), descriptions, and dark theme highlight states.

3. **Observation**: Post-creation success state currently replaces the entire form view inline on `NewApp.tsx` (lines 49-114).
   - **Reasoning**: Moving post-creation display into a modal component `SecretSuccessModal.tsx` under `src/components/new-app/` provides a focused overlay experience that cleanly separates creation form flow from credentials display, while supporting 1-click clipboard copy, amber secret warning banner, and direct navigation buttons (`/apps/${app.ID}`).

4. **Observation**: `App.tsx` routes `/apps/new` inside `SidebarLayout`.
   - **Reasoning**: `NewApp.tsx` must omit any `Navbar` references to prevent duplicate headers or breaking layout constraints.

5. **Observation**: Existing test pattern in `run_m3_test.mjs` uses Vite SSR to render React components in Node without browser overhead.
   - **Reasoning**: Worker M4 can adopt this exact test pattern by creating `frontend/src/__tests__/m4_verification_test.tsx` and `run_m4_test.mjs`.

---

## 3. Caveats

- **API Compatibility**: No backend API changes are needed; `apps.create` parameters remain `{ name, description, homepage_url, redirect_uris, is_public }`.
- **Local Storage / Session State**: Stepper state is ephemeral within the component state of `NewApp.tsx`; refreshing the page resets the wizard to Step 1.
- **Clipboard Permissions**: `navigator.clipboard.writeText` is supported in secure contexts and modern browsers; fallback inline text selection is preserved via `<pre className="secret-box">`.

---

## 4. Conclusion

The blueprint for Milestone 4 provides a complete, production-ready specification for redesigning `NewApp.tsx` into an Auth0-style 3-step wizard with modular components (`TypeSelector.tsx`, `SecretSuccessModal.tsx`, `WizardStepper.tsx`). The plan preserves full backend API compatibility, ensures zero Navbar collisions inside `SidebarLayout`, and defines an automated Vite SSR test suite for Worker M4 to verify implementation.

---

## 5. Verification Method

To verify the implementation once completed by Worker M4:

1. **Automated Verification**:
   ```bash
   node frontend/src/__tests__/run_m4_test.mjs
   ```
   *Expected output*: 100% passing tests for `TypeSelector`, `SecretSuccessModal`, `WizardStepper`, and `NewApp` wizard step transitions.

2. **Full Frontend Test Verification**:
   ```bash
   node frontend/src/__tests__/run_m3_test.mjs
   ```
   *Expected output*: 26 PASSED, 0 FAILED.

3. **File Layout Inspection**:
   - Confirm presence of `frontend/src/components/new-app/TypeSelector.tsx`
   - Confirm presence of `frontend/src/components/new-app/SecretSuccessModal.tsx`
   - Confirm presence of `frontend/src/components/new-app/WizardStepper.tsx`
   - Confirm update to `frontend/src/pages/NewApp.tsx`
   - Confirm zero imports/references to `Navbar` in `NewApp.tsx`.
