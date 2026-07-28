# Changes Summary — Milestone 2 (R1 Persistent Left Sidebar Shell)

## Overview
Implemented the persistent, responsive left sidebar layout (`SidebarLayout.tsx`) for all authenticated routes in DilithiAuth (`/apps`, `/apps/new`, `/apps/:id`). Updated application routing and cleaned up redundant `<Navbar />` components from authenticated views.

---

## Files Modified & Created

### 1. `frontend/src/components/layout/SidebarLayout.tsx` (CREATED)
- **Purpose**: Persistent responsive shell for authenticated routes.
- **Features Implemented**:
  - **Brand Header**: Top logo displaying "DilithiAuth" and a post-quantum identity badge ("ML-DSA-44").
  - **Navigation Section**:
    - **Applications**: Links to `/apps` with `LayoutGrid` icon. Highlights when `location.pathname.startsWith('/apps')`.
    - **Docs**: Links to `/docs` with `BookOpen` icon. Highlights when `location.pathname === '/docs'`.
    - **Settings**: Disabled state with `Settings` icon and `"Soon"` badge (`opacity-50`, `cursor-not-allowed`).
  - **Bottom Anchored Profile**:
    - User avatar displaying capitalized initials from `developer.name` or `developer.email`.
    - Developer name and email display with truncation handling.
    - **Sign Out** button calling `logout()` from `useAuth()` and navigating to `/login`.
  - **Viewport Responsiveness**:
    - **Desktop (>= 1024px / `lg`)**: 256px wide persistent left sidebar (`w-64`), sticky left side (`sticky top-0 h-screen`).
    - **Tablet (768px - 1023px / `md`)**: 64px icon-only rail (`w-16`), showing icons with tooltips (`title` / `aria-label`).
    - **Mobile (< 768px)**: Sticky 60px top header with brand logo & hamburger toggle button (`isMobileOpen`), dark backdrop blur overlay closing on click or link navigation, slide-down drawer container.
  - **Content Container**: Embeds React Router v7 `<Outlet />` inside `<main className="flex-1 min-w-0 p-4 sm:p-6 md:p-8 overflow-y-auto">`.
  - **Theme Alignment**: Fully integrated with Tailwind CSS v4 design tokens from `src/index.css` (`var(--color-bg)`, `var(--color-card)`, `var(--color-border)`, `var(--color-txt)`, `var(--color-sub)`, `var(--color-muted)`).

### 2. `frontend/src/App.tsx` (MODIFIED)
- **Changes**:
  - Imported `SidebarLayout` from `./components/layout/SidebarLayout`.
  - Grouped authenticated routes (`/apps`, `/apps/new`, `/apps/:id`) under `<Route element={<RequireAuth><SidebarLayout /></RequireAuth>}>`.
  - Retained public routes (`/`, `/login`, `/register`, `/docs`, `/privacy`, `/terms`) using `Navbar.tsx`.

### 3. `frontend/src/pages/Dashboard.tsx` (MODIFIED)
- **Changes**:
  - Removed unused `Navbar` import from `../components/Navbar`.
  - Removed top `<Navbar />` component call to eliminate dual navbar rendering on authenticated pages.

### 4. `frontend/src/pages/NewApp.tsx` (MODIFIED)
- **Changes**:
  - Removed unused `Navbar` import from `../components/Navbar`.
  - Removed `<Navbar />` calls from both the creation success view (`created`) and the form view.

### 5. `frontend/src/pages/AppDetail.tsx` (MODIFIED)
- **Changes**:
  - Removed unused `Navbar` import.
  - Removed `<Navbar />` calls from loading, error, and detail layout views.

---

## Verification & Build Results
- **Command**: `npm run build` executed in `frontend/`.
- **Result**: `✓ built in 331ms` with 0 TypeScript compilation errors and 0 Vite bundler errors.
- **Go Tests**: `go test ./...` passed all packages in workspace.
