# Frontend Codebase Structure Analysis

**Target Project**: `/home/uttam/projects/DilithiAuth/frontend`  
**Date**: 2026-07-28  
**Subagent**: Explorer (`teamwork_preview_explorer_m1_1`)

---

## Executive Summary

The frontend is a single-page React 19 application configured with Vite 8, TypeScript 6, React Router DOM 7, and Tailwind CSS v4. Authentication state is managed globally via React Context (`AuthContext.tsx`). Page navigation uses component-level routing guards (`RequireAuth` and `RequireGuest`). Layout rendering is page-contained: `<Navbar />` is manually instantiated across all page components rather than wrapped in a global layout shell. Design tokens are defined directly in CSS using Tailwind v4's `@theme` directive.

---

## 1. Routing Structure & Auth Guards

### Route Configuration (`src/App.tsx` & `src/main.tsx`)
- `main.tsx` renders `<App />` within React `StrictMode`.
- `App.tsx` establishes the provider hierarchy and route definitions:
  - Top level: `<AuthProvider>` wraps `<BrowserRouter basename={import.meta.env.VITE_ROUTER_BASENAME || ''}>`.
  - Routes are defined declaratively using `<Routes>` and `<Route>`.

### Auth Guard Components (`src/App.tsx`)
Authentication logic is enforced using wrapper components consuming the `useAuth()` hook:

1. **`RequireAuth({ children })`** (Lines 13–17):
   - Accesses `{ developer, loading }` from `useAuth()`.
   - While `loading` is `true`, renders `<div style={{ padding: '2rem' }}>Loading…</div>`.
   - If `developer` is truthy, renders `children`.
   - If `developer` is null/falsy, redirects to `/login` with `replace`.

2. **`RequireGuest({ children })`** (Lines 19–23):
   - Accesses `{ developer, loading }` from `useAuth()`.
   - While `loading` is `true`, renders `<div style={{ padding: '2rem' }}>Loading…</div>`.
   - If `developer` is truthy, redirects to `/apps` with `replace`.
   - If `developer` is null/falsy, renders `children`.

### Complete Route Map

| Path | Route Element / Protection | Component | Behavior |
| --- | --- | --- | --- |
| `/` | `<RequireGuest><Landing /></RequireGuest>` | `Landing.tsx` | Guest home; logged-in users redirected to `/apps` |
| `/login` | `<RequireGuest><Login /></RequireGuest>` | `Login.tsx` | Sign-in page; logged-in users redirected to `/apps` |
| `/register` | `<RequireGuest><Register /></RequireGuest>` | `Register.tsx` | Registration page; logged-in users redirected to `/apps` |
| `/docs` | `<Docs />` | `Docs.tsx` | Public API & SDK documentation |
| `/privacy` | `<Privacy />` | `Privacy.tsx` | Public Privacy Policy |
| `/terms` | `<Terms />` | `Terms.tsx` | Public Terms of Service |
| `/apps` | `<RequireAuth><Dashboard /></RequireAuth>` | `Dashboard.tsx` | Developer dashboard (App listing) |
| `/apps/new` | `<RequireAuth><NewApp /></RequireAuth>` | `NewApp.tsx` | Application registration form |
| `/apps/:id` | `<RequireAuth><AppDetail /></RequireAuth>` | `AppDetail.tsx` | Application management & secret rotation |
| `*` | `<Navigate to="/" replace />` | — | Fallback catch-all route |

---

## 2. Navbar Component & Page Integration

### Component Architecture (`src/components/Navbar.tsx`)
- **Structure**: Sticky navigation bar (`sticky top-0 z-[200]`) with fixed height `60px` (`h-[60px]`), translucent dark background (`bg-[#090d16]/90 backdrop-blur-md`), and bottom border (`border-b border-[var(--color-border)]`).
- **Brand Link**: Brand header (`DilithiAuth PQC IdP`) links to `/apps` if `developer` exists, or `/` if guest.
- **Desktop View (`lg:flex`)**: Displays horizontal navigation links using `lucide-react` icons.
- **Mobile View (`lg:hidden`)**: Displays a hamburger toggle button (`Menu` / `X`). Toggling opens a fixed slide-down drawer (`z-[199]`) and backdrop overlay (`z-[198]`). Automatically closes drawer on location change or pressing `Escape`.
- **Dynamic Content**:
  - **Authenticated State**: Displays `developer.email` badge, link to `/apps` ("Applications"), link to `/docs` ("SDK & Docs"), and "Sign out" button.
  - **Guest State**: Displays links to `/` ("Overview"), `/docs` ("SDK & Docs"), `/login` ("Sign in"), and CTA to `/register` ("Get Started").
  - **Logout Action**: Executes `handleLogout()` -> calls `logout()` from `useAuth()`, closes drawer, and navigates to `/login`.

### Usage Across Pages
`<Navbar />` is **not** wrapped in a top-level layout route. Instead, every single page imports and renders `<Navbar />` explicitly at the root of its JSX tree:

1. `src/pages/Dashboard.tsx` — Rendered at line 39.
2. `src/pages/NewApp.tsx` — Rendered at lines 53 and 122 (both in creation success view and form view).
3. `src/pages/AppDetail.tsx` — Rendered at lines 69, 82, and top of main render (loading, error, and detail views).
4. `src/pages/Landing.tsx` — Rendered at line 16.
5. `src/pages/Login.tsx` — Rendered at line 47.
6. `src/pages/Register.tsx` — Rendered at line 105.
7. `src/pages/Docs.tsx` — Rendered at top of main layout.
8. `src/pages/DocsLoader.tsx` — Rendered at top of page container.
9. `src/pages/Privacy.tsx` — Rendered at line 8.
10. `src/pages/Terms.tsx` — Rendered at line 8.

---

## 3. AuthContext Analysis (`src/AuthContext.tsx`)

### Location & Exports
- **File**: `src/AuthContext.tsx` (Note: `src/AuthContext.js` exists as a compiled build artifact, but TS imports resolve to `src/AuthContext.tsx`).
- **Exports**: `AuthProvider` (React Context Provider) and `useAuth` (custom hook).

### Interface Definitions & Data Types (`src/api.ts` & `src/AuthContext.tsx`)
```typescript
// Developer user object shape (src/api.ts)
export interface Developer {
  id: string
  email: string
  name: string
  csrf_token?: string
}

// Context interface (src/AuthContext.tsx)
interface AuthCtx {
  developer: Developer | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, name: string, password: string) => Promise<void>
  logout: () => Promise<void>
}
```

### Key Clarifications on Terminology
- **User Object**: Named **`developer`** (type `Developer | null`), NOT `user`.
- **Sign Out Method**: Named **`logout`** (type `() => Promise<void>`), NOT `signOut`.
- **Login State**: Authenticated state is indicated by `developer !== null` once `loading === false`.

### State Lifecycle & API Operations
1. **Initial Mount (`useEffect`)**:
   - Executes `auth.me()` (performs GET to `/api/v1/bff/me` with `withCredentials: true`).
   - On success: sets `developer` state with `res.data` and stores `csrf_token` in memory module variable.
   - On error: sets `developer` state to `null`.
   - In `finally`: sets `loading` to `false`.
2. **`login(email, password)`**:
   - Calls `auth.login(email, password)` (POST `/api/v1/developer/login`).
   - Updates `developer` state and sets `csrfToken`.
3. **`register(email, name, password)`**:
   - Calls `auth.register(email, name, password)` (POST `/api/v1/developer/register`).
   - Updates `developer` state.
4. **`logout()`**:
   - Clears in-memory `csrfToken` variable.
   - Triggers concurrent logout requests using `Promise.allSettled`:
     - `POST /api/v1/developer/logout`
     - `POST /api/v1/bff/logout`
   - In `finally`: sets `developer` to `null`.

---

## 4. Design Tokens & Styling Architecture (`src/index.css`)

### Tailwind CSS v4 Setup
- Configured using Tailwind CSS v4 (`@tailwindcss/vite` v4.3.3, `tailwindcss` v4.3.3).
- Direct import in `src/index.css`: `@import "tailwindcss";`.
- Uses Tailwind v4's zero-config `@theme` block in CSS (no standalone `tailwind.config.js`).

### Design Tokens (`@theme` block)

```css
@theme {
  /* Backgrounds */
  --color-bg:         #090d16;
  --color-card:       #0f1623;
  --color-card2:      #141c2e;
  --color-hover:      #1a2236;
  --color-input:      #0c1120;

  /* Borders */
  --color-border:     #1e2a40;
  --color-border-mid: #263347;
  --color-focus:      #3b82f6;

  /* Accent & Status Colors */
  --color-primary:    #3b82f6;
  --color-primary-h:  #2563eb;
  --color-emerald:    #10b981;
  --color-emerald-h:  #059669;
  --color-rose:       #ef4444;
  --color-rose-h:     #dc2626;
  --color-amber:      #f59e0b;
  --color-cyan:       #06b6d4;
  --color-sky:        #38bdf8;
  --color-violet:     #a5b4fc;

  /* Text Hierarchy */
  --color-txt:        #f1f5f9;
  --color-sub:        #94a3b8;
  --color-muted:      #64748b;

  /* Typography */
  --font-sans:        'Inter', system-ui, -apple-system, sans-serif;
  --font-mono:        'JetBrains Mono', 'Fira Code', monospace;

  /* Border Radii */
  --radius-xs:        4px;
  --radius-sm:        6px;
  --radius:           10px;
  --radius-lg:        14px;

  /* Spacing */
  --spacing-nav:      60px;
}
```

### CSS Layers & Custom Utilities
- **`@layer base`**:
  - Global `box-sizing: border-box`, `margin: 0`, `padding: 0`.
  - Body background set to `var(--color-bg)` (`#090d16`) and text to `var(--color-txt)` (`#f1f5f9`).
  - Webkit scrollbar styling with `var(--color-border-mid)` thumb.
- **`@layer components`**:
  - `.code-block`: Dark code container (`#050810`), monospace, overflow handling.
  - `.secret-box`: Monospace box for API secrets (`#a3e635` text, break-all).
  - `.mono-inline`: Inline code snippets.
  - `.skeleton`: Animated shimmer loading box using `@keyframes shimmer`.
  - `.tab-active` & `.nav-active`: Navigation active state helpers.
  - `.spin`: Rotating spinner animation.

---

## Summary Table of Key Findings

| Domain | Key Finding | Source Location |
| --- | --- | --- |
| Router | React Router DOM v7; `<AuthProvider>` wraps router tree | `src/App.tsx` |
| Auth Guards | `RequireAuth` and `RequireGuest` wrappers check `developer` & `loading` | `src/App.tsx:13-23` |
| Navigation Component | `Navbar` handles desktop links & mobile drawer; dynamic guest/auth links | `src/components/Navbar.tsx` |
| Navbar Usage | Instantiated per-page across all 10 page components | `src/pages/*.tsx` |
| Auth State | Object named `developer` (`Developer \| null`); method named `logout` | `src/AuthContext.tsx` |
| CSRF Handling | Interceptor attaches `X-CSRF-Token` header from `auth.me()` / `login()` | `src/api.ts:21-29` |
| Styling Engine | Tailwind CSS v4 using `@import "tailwindcss"` and `@theme` | `src/index.css` |
