## 2026-07-28T14:45:45Z
You are a Worker subagent for Milestone 2 (R1 Persistent Left Sidebar Shell).
Your metadata working directory is `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m2_1`.
Project code location: `/home/uttam/projects/DilithiAuth/frontend`.

MANDATORY INTEGRITY WARNING:
DO NOT CHEAT. All implementations must be genuine. DO NOT hardcode test results, create dummy/facade implementations, or circumvent the intended task. A Forensic Auditor will independently verify your work. Integrity violations WILL be detected and your work WILL be rejected.

Tasks for Milestone 2:
1. Read the blueprint in `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m2_1/analysis.md` and `/home/uttam/projects/DilithiAuth/PROJECT.md`.
2. Create `src/components/layout/SidebarLayout.tsx`:
   - Brand logo at top: "DilithiAuth" with post-quantum identity badge ("ML-DSA-44").
   - Nav items:
     - Applications: links to `/apps`, Lucide `LayoutGrid` icon. Highlight active link when path starts with `/apps`.
     - Docs: links to `/docs`, Lucide `BookOpen` or `FileText` icon.
     - Settings: disabled link / button with `Lock` or `Settings` icon and "Soon" badge.
   - User profile section anchored at bottom:
     - User avatar with initials (from `developer.name` or `developer.email`).
     - Display `developer.email` (or `developer.name`).
     - "Sign Out" button that invokes `logout()` from `useAuth()` (from `src/AuthContext.tsx`).
   - Viewport Responsiveness:
     - Desktop (>=1024px): Persistent 256px wide left sidebar (`w-64`), sticky left side.
     - Tablet (768px - 1023px): 64px icon-only rail (`w-16`), hiding labels, showing icons with title/tooltip.
     - Mobile (<768px): Top bar with brand logo + hamburger toggle button; sliding drawer overlay when `isMobileOpen` is true, dark backdrop overlay that closes on click or nav click.
   - Styling: Dark theme using CSS custom properties from `src/index.css` (`--color-bg`, `--color-card`, `--color-border`, `--color-txt`, `--color-primary`, etc.).
   - Include `<Outlet />` in the main content container.
3. Update `src/App.tsx`:
   - Wrap authenticated routes (`/apps`, `/apps/new`, `/apps/:id`) under a route element layout `<RequireAuth><SidebarLayout /></RequireAuth>` using an Outlet, or wrap each inside `<RequireAuth><SidebarLayout><Dashboard /></SidebarLayout></RequireAuth>`.
   - Preserve public routes (`/`, `/login`, `/register`, `/docs`, `/privacy`, `/terms`) using `Navbar.tsx`.
4. Clean up leftover `<Navbar />` calls:
   - Remove `<Navbar />` from `src/pages/Dashboard.tsx`, `src/pages/NewApp.tsx`, and `src/pages/AppDetail.tsx` so authenticated pages don't have dual top navbars.
5. Verification:
   - Run `npm run build` in `frontend/` using `run_command`. Ensure 0 TypeScript errors and 0 Vite errors.
   - Document build command and result in your handoff report.

Write your changes report to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m2_1/changes.md` and handoff report to `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_worker_m2_1/handoff.md`. Communicate your completion via `send_message` to parent (ID: 145f166d-e8fc-4e12-bee7-7f454f7d2939).
