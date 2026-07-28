# Original User Request

## Initial Request — 2026-07-28T14:41:18Z

Redesign the DilithiAuth developer portal authenticated screens (`/apps`, `/apps/new`, `/apps/:id`) into an **Auth0 / Clerk-style dashboard** with a persistent left sidebar navigation, replacing the current top-navbar-only layout with a proper developer console experience.

Working directory: `/home/uttam/projects/DilithiAuth/frontend`
Integrity mode: `development`

## Context

- React + Vite 8 + Tailwind v4 SPA.
- Existing design tokens live in `src/index.css` (`@theme` block).
- Existing pages: `src/pages/Dashboard.tsx`, `src/pages/NewApp.tsx`, `src/pages/AppDetail.tsx`, `src/components/Navbar.tsx`.
- The top `Navbar.tsx` is used by public pages (Landing, Login, etc.) — keep it unchanged.
- The authenticated dashboard layout should get its own sidebar shell, **not** re-use `Navbar.tsx`.

## Requirements

### R1. Persistent Left Sidebar Shell for Authenticated Pages
Create a sidebar layout shell (component or layout wrapper) that appears on all authenticated pages (`/apps`, `/apps/new`, `/apps/:id`). The sidebar should include:
- DilithiAuth logo / brand at the top.
- Navigation items: **Applications** (links to `/apps`), **Docs** (links to `/docs`), **Settings** (placeholder, disabled or greyed).
- Developer avatar / email and **Sign Out** button anchored at the bottom.
- Collapses to an icon-only rail on tablet, becomes a drawer/overlay on mobile.
- Styled dark theme consistent with existing `--color-*` CSS variables.

### R2. Applications List Page (`/apps`) — Auth0 Console Aesthetic
Redesign `Dashboard.tsx` to use the sidebar shell. The main content area should feel like Auth0's Applications list:
- Page header with title, subtitle, and a prominent **"Create Application"** button.
- App cards or table rows showing app name, client ID (masked), type badge (Public / Confidential), and quick-copy client ID button.
- Empty state illustration/message with a CTA when no apps exist.
- Search/filter bar above the list.

### R3. Create Application Page (`/apps/new`) — Wizard-Style Form
Redesign `NewApp.tsx` inside the sidebar shell. Make the creation flow feel like a proper Auth0-style wizard:
- Clear step or section grouping: **Basic Info** → **Application Type** → **URIs**.
- Application Type selector with visual cards (Public SPA vs Confidential Server-Side), not just a checkbox.
- Post-creation success screen that prominently shows Client ID and Client Secret with copy buttons and a clear warning about the secret.

### R4. App Detail Page (`/apps/:id`) — Settings Panel
Redesign `AppDetail.tsx` inside the sidebar shell. Use a **left sub-nav** (or top tab bar) inside the detail page for: **Settings**, **Code Snippets**, **Danger Zone**. The settings tab should show all app fields in clean editable form groups.

## Acceptance Criteria

### Build & Type Safety
- [ ] `npm run build` in `frontend/` completes with 0 TypeScript errors and 0 Vite errors.
- [ ] No `console.error` TypeScript errors in the browser (types must be correct).

### Layout & Responsiveness
- [ ] Sidebar is visible and sticky on desktop (≥1024px).
- [ ] Sidebar collapses to icon rail on tablet (768px–1023px).
- [ ] Sidebar becomes a slide-in drawer on mobile (<768px) triggered by a hamburger button.
- [ ] No horizontal scroll on any viewport ≥320px.

### Functionality Preserved
- [ ] All existing API calls (`apps.list()`, `apps.create()`, `apps.get()`, `apps.rotateSecret()`, `apps.delete()`) continue to work correctly — no logic changes, only UI.
- [ ] Sign Out button in the sidebar works (calls the existing `signOut` from `AuthContext`).
- [ ] Copy-to-clipboard, form submission, secret rotation, and app deletion all work as before.

### Visual Quality
- [ ] The authenticated portal visually matches the quality bar of Auth0 or Clerk developer consoles — dark theme, clear hierarchy, premium spacing.
- [ ] No leftover references to `Navbar` inside authenticated dashboard pages.

## Resume Request — 2026-07-28T20:50:21Z

Resume work as Project Orchestrator (Generation 2) at workspace `/home/uttam/projects/DilithiAuth/.agents/orchestrator`.
Execute Milestone 4 (R3 Create Application Page - `NewApp.tsx` Wizard Redesign), Milestone 5 (R4 App Detail Page redesign in `AppDetail.tsx`), and Milestone 6 (Final Build Verification).
