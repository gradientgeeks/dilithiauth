# Project Plan: DilithiAuth Developer Portal Redesign

## Objective
Redesign the DilithiAuth developer portal authenticated screens (`/apps`, `/apps/new`, `/apps/:id`) into an Auth0 / Clerk-style dashboard with a persistent left sidebar navigation.

## Milestones & Strategy

### Milestone 1: Exploration & Architectural Contract
- **Goal**: Read and inspect existing frontend codebase in `frontend/src/`. Map router layout, AuthContext usage, design tokens in `index.css`, components (`Navbar.tsx`, `Dashboard.tsx`, `NewApp.tsx`, `AppDetail.tsx`), and API bindings (`apps` service).
- **Deliverable**: `PROJECT.md` at root defining architecture, interfaces, code layout, and milestone list.

### Milestone 2: R1 - Persistent Left Sidebar Shell (`SidebarLayout.tsx`)
- **Goal**: Create a sticky sidebar shell component for authenticated routes.
- **Features**: Logo at top, nav items (Applications `/apps`, Docs `/docs`, Settings disabled), user avatar/email + Sign Out at bottom.
- **Responsiveness**: Sticky rail on desktop (≥1024px), icon rail on tablet (768–1023px), drawer/overlay on mobile (<768px). Styled with dark theme design tokens.

### Milestone 3: R2 - Applications List Page (`Dashboard.tsx`)
- **Goal**: Redesign `Dashboard.tsx` within sidebar shell.
- **Features**: Auth0 console aesthetic, page header with title/subtitle and "Create Application" button, search/filter bar, app cards/table rows with masked Client ID, quick copy button, type badge, empty state illustration/CTA.

### Milestone 4: R3 - Create Application Wizard Page (`NewApp.tsx`)
- **Goal**: Redesign `NewApp.tsx` within sidebar shell.
- **Features**: Multi-step/grouped wizard (Basic Info → Application Type → URIs), visual card selector for App Type (Public SPA vs Confidential Server-Side), post-creation success modal/screen with Client ID & Client Secret copy buttons and warning.

### Milestone 5: R4 - App Detail Page (`AppDetail.tsx`)
- **Goal**: Redesign `AppDetail.tsx` within sidebar shell.
- **Features**: Sub-navigation/tabs for Settings, Code Snippets (OAuth code samples / configs), Danger Zone (Rotate Secret, Delete App). Clean form groups for settings.

### Milestone 6: E2E Integration, Verification & Final Audit
- **Goal**: Verify full build (`npm run build` with 0 TS and 0 Vite errors), responsiveness across viewports, no leftover Navbar references in authenticated pages, all API calls and auth actions functioning. Forensic Audit check.
