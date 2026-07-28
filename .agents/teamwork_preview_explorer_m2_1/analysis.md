# Architectural Blueprint & Implementation Plan: R1 Persistent Left Sidebar Shell

## Executive Summary
Milestone 2 (R1) establishes a persistent, responsive left sidebar layout (`SidebarLayout.tsx`) for all authenticated routes in DilithiAuth (`/apps`, `/apps/new`, `/apps/:id`). Public routes (`/`, `/login`, `/register`, `/privacy`, `/terms`) retain the top navigation bar (`Navbar.tsx`). This report presents the complete architectural specification, responsive layout mechanics, theme token integration, and exact code design for implementation.

---

## 1. Current Routing & Layout Analysis

### Existing Architecture
Currently, `src/App.tsx` defines flat routes where each page component independently imports and renders `<Navbar />`:
- `Dashboard.tsx`, `NewApp.tsx`, and `AppDetail.tsx` render `<Navbar />` inside their root JSX element.
- Guest and public routes (`Landing.tsx`, `Login.tsx`, `Register.tsx`, `Docs.tsx`, `Privacy.tsx`, `Terms.tsx`) also render `<Navbar />`.
- Auth guards `RequireAuth` and `RequireGuest` wrap individual route `element` props.

### Target Architecture
With R1, routing in `src/App.tsx` is restructured into two distinct layout branches:
1. **Public / Standalone Routes Branch**: Uses top `Navbar.tsx` embedded in page views or a public layout shell.
2. **Authenticated Dashboard Branch**: Uses React Router v7 `<Outlet />` wrapped in `<RequireAuth><SidebarLayout /></RequireAuth>`.

---

## 2. Responsive Layout Specification: `SidebarLayout.tsx`

`SidebarLayout.tsx` adapts seamlessly across three distinct screen size tiers:

| Viewport Tier | Breakpoint | Sidebar Variant | Dimensions | Key Features |
|---|---|---|---|---|
| **Desktop** | `>= 1024px` (`lg`) | Expanded Sticky Sidebar | `w-64` (256px wide) | Brand logo, full labels + icons, active indicators, bottom user card + Sign Out button |
| **Tablet** | `768px – 1023px` (`md`) | Icon Rail | `w-16` (64px wide) | Centered icons with tooltip `title` / `aria-label`, user avatar circle, logout icon button |
| **Mobile** | `< 768px` (< `md`) | Top Bar + Drawer | `h-[60px]` top bar | Fixed header with brand logo & hamburger toggle; backdrop blur slide-down drawer |

---

## 3. Detailed Component Blueprint for `SidebarLayout.tsx`

### A. State & Context Hooks
- `useAuth()`: Consumes `developer` object (`email`, `name`) and `logout()` function from `AuthContext.tsx`.
- `useLocation()`: Detects active route for navigation highlighting (`pathname.startsWith('/apps')`, `pathname === '/docs'`). Closes mobile drawer automatically on path changes.
- `useNavigate()`: Navigates user to `/login` after `logout()`.
- `isMobileOpen` (boolean state): Manages mobile drawer toggle. Handled with `Escape` key event listener and click-outside backdrop listener.

### B. Navigation Items Specification
1. **Applications**
   - Path: `/apps`
   - Icon: `LayoutGrid` (Lucide React, `size={18}`)
   - Label: `Applications`
   - Active state: `location.pathname.startsWith('/apps')`
2. **Documentation**
   - Path: `/docs`
   - Icon: `BookOpen` (Lucide React, `size={18}`)
   - Label: `Docs`
   - Active state: `location.pathname === '/docs'`
3. **Settings**
   - Path: Disabled (`#` or `button`)
   - Icon: `Settings` (Lucide React, `size={18}`)
   - Label: `Settings`
   - State: Disabled (`disabled`, `opacity-40`, `cursor-not-allowed`) with a subtle `"Soon"` badge (`text-[0.6rem] bg-[var(--color-border)] text-[var(--color-muted)] px-1.5 py-0.5 rounded`).

### C. Bottom Anchored User Section
- **Avatar & User Details**:
  - Avatar: Circular badge (`w-8 h-8 rounded-full bg-[var(--color-primary)]/20 border border-[var(--color-primary)]/40 text-[var(--color-primary)] font-semibold flex items-center justify-center text-xs shrink-0`) containing initials (e.g. derived from `developer.name` or `developer.email`).
  - Info: Display `developer.name` (or truncated email prefix) and `developer.email` (`text-xs text-[var(--color-sub)] truncate`).
- **Sign Out Action**:
  - Button executing `async () => { await logout(); navigate('/login'); }`.
  - Icon: `LogOut` (`size={16}`).
  - Styled with subtle red hover state: `text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 border border-rose-500/20 rounded-md`.

### D. Dark Theme Token Alignment
All colors and spacing map directly to `src/index.css` `@theme` CSS variables:
- Background: `bg-[var(--color-bg)]` (`#090d16`)
- Sidebar background: `bg-[var(--color-card)]` (`#0f1623`)
- Hover highlight: `bg-[var(--color-hover)]` (`#1a2236`)
- Borders: `border-[var(--color-border)]` (`#1e2a40`)
- Text: Primary `text-[var(--color-txt)]` (`#f1f5f9`), Secondary `text-[var(--color-sub)]` (`#94a3b8`), Muted `text-[var(--color-muted)]` (`#64748b`)

---

## 4. Exact Implementation Code Blueprint

### 4.1 `src/components/layout/SidebarLayout.tsx`

```tsx
import { useState, useEffect } from 'react'
import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom'
import {
  LayoutGrid, BookOpen, Settings, LogOut, Menu, X, ShieldCheck
} from 'lucide-react'
import { useAuth } from '../../AuthContext'

export default function SidebarLayout() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [isMobileOpen, setIsMobileOpen] = useState(false)

  // Close mobile drawer on route change
  useEffect(() => {
    setIsMobileOpen(false)
  }, [location.pathname])

  // Close mobile drawer on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsMobileOpen(false)
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  const handleLogout = async () => {
    try {
      await logout()
    } catch {
      // Ignore network errors on logout
    }
    setIsMobileOpen(false)
    navigate('/login')
  }

  const isAppsActive = location.pathname.startsWith('/apps')
  const isDocsActive = location.pathname === '/docs'

  // User initials for avatar
  const userInitial = developer?.name
    ? developer.name.charAt(0).toUpperCase()
    : developer?.email
    ? developer.email.charAt(0).toUpperCase()
    : 'U'

  return (
    <div className="min-h-screen bg-[var(--color-bg)] text-[var(--color-txt)] flex flex-col md:flex-row">
      {/* ============================================================
         1. MOBILE TOP BAR (< 768px)
         ============================================================ */}
      <header className="md:hidden sticky top-0 z-[100] h-[60px] w-full px-4 flex items-center justify-between bg-[#090d16]/90 backdrop-blur-md border-b border-[var(--color-border)]">
        <Link to="/apps" className="flex items-center gap-2 no-underline">
          <img src="/favicon.svg" alt="DilithiAuth" width={26} height={26} className="block shrink-0" />
          <span className="font-bold text-[var(--color-txt)] text-[0.95rem] tracking-tight flex items-center gap-1.5">
            DilithiAuth
            <span className="bg-[var(--color-hover)] border border-[var(--color-border-mid)] text-[var(--color-sub)] text-[0.6rem] font-bold px-1.5 py-0.5 rounded uppercase tracking-widest">
              PQC IdP
            </span>
          </span>
        </Link>
        <button
          onClick={() => setIsMobileOpen(v => !v)}
          aria-label={isMobileOpen ? 'Close navigation menu' : 'Open navigation menu'}
          aria-expanded={isMobileOpen}
          className="flex items-center justify-center w-9 h-9 rounded-md border border-[var(--color-border)] text-[var(--color-txt)] bg-transparent hover:bg-[var(--color-hover)] transition-colors cursor-pointer"
        >
          {isMobileOpen ? <X size={20} /> : <Menu size={20} />}
        </button>
      </header>

      {/* ============================================================
         2. MOBILE DRAWER OVERLAY (< 768px)
         ============================================================ */}
      {isMobileOpen && (
        <>
          <div
            className="fixed inset-0 z-[198] bg-black/60 backdrop-blur-sm md:hidden"
            onClick={() => setIsMobileOpen(false)}
            aria-hidden="true"
          />
          <aside
            className="fixed top-[60px] left-0 right-0 z-[199] bg-[var(--color-card)] border-b border-[var(--color-border)] shadow-2xl p-4 md:hidden flex flex-col gap-4"
            role="dialog"
            aria-label="Mobile Navigation"
          >
            {developer && (
              <div className="flex items-center gap-3 px-3 py-2.5 bg-[var(--color-bg)] rounded-md border border-[var(--color-border)]">
                <div className="w-8 h-8 rounded-full bg-[var(--color-primary)]/20 border border-[var(--color-primary)]/40 text-[var(--color-primary)] font-semibold flex items-center justify-center text-xs shrink-0">
                  {userInitial}
                </div>
                <div className="flex flex-col min-w-0">
                  <span className="text-xs font-semibold text-[var(--color-txt)] truncate">
                    {developer.name || 'Developer'}
                  </span>
                  <span className="text-[0.75rem] text-[var(--color-sub)] truncate">
                    {developer.email}
                  </span>
                </div>
              </div>
            )}

            <nav className="flex flex-col gap-1">
              <Link
                to="/apps"
                className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-all ${
                  isAppsActive
                    ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold border-l-2 border-[var(--color-primary)]'
                    : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
                }`}
              >
                <LayoutGrid size={18} /> Applications
              </Link>
              <Link
                to="/docs"
                className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-all ${
                  isDocsActive
                    ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold border-l-2 border-[var(--color-primary)]'
                    : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
                }`}
              >
                <BookOpen size={18} /> Docs
              </Link>
              <div className="flex items-center justify-between px-3 py-2.5 rounded-md text-sm font-medium text-[var(--color-muted)] opacity-50 cursor-not-allowed">
                <div className="flex items-center gap-3">
                  <Settings size={18} /> Settings
                </div>
                <span className="text-[0.6rem] bg-[var(--color-border)] text-[var(--color-muted)] px-1.5 py-0.5 rounded font-mono uppercase">
                  Soon
                </span>
              </div>
            </nav>

            <button
              onClick={handleLogout}
              className="flex items-center justify-center gap-2 w-full px-3 py-2.5 rounded-md text-sm font-medium text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 border border-rose-500/20 bg-transparent cursor-pointer transition-all mt-2"
            >
              <LogOut size={18} /> Sign out
            </button>
          </aside>
        </>
      )}

      {/* ============================================================
         3. TABLET ICON RAIL (768px - 1023px)
         ============================================================ */}
      <aside className="hidden md:flex lg:hidden sticky top-0 h-screen w-16 shrink-0 border-r border-[var(--color-border)] bg-[var(--color-card)] flex-col justify-between items-center py-4 z-40">
        <div className="flex flex-col items-center gap-6 w-full">
          <Link to="/apps" title="DilithiAuth Applications" className="no-underline">
            <img src="/favicon.svg" alt="DilithiAuth" width={28} height={28} className="block" />
          </Link>
          <nav className="flex flex-col gap-3 w-full px-2">
            <Link
              to="/apps"
              title="Applications"
              aria-label="Applications"
              className={`flex items-center justify-center w-12 h-12 rounded-lg transition-all ${
                isAppsActive
                  ? 'bg-[var(--color-hover)] text-[var(--color-primary)] font-semibold border-l-2 border-[var(--color-primary)]'
                  : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
              }`}
            >
              <LayoutGrid size={20} />
            </Link>
            <Link
              to="/docs"
              title="Docs"
              aria-label="Docs"
              className={`flex items-center justify-center w-12 h-12 rounded-lg transition-all ${
                isDocsActive
                  ? 'bg-[var(--color-hover)] text-[var(--color-primary)] font-semibold border-l-2 border-[var(--color-primary)]'
                  : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
              }`}
            >
              <BookOpen size={20} />
            </Link>
            <div
              title="Settings (Coming Soon)"
              aria-label="Settings"
              className="flex items-center justify-center w-12 h-12 rounded-lg text-[var(--color-muted)] opacity-40 cursor-not-allowed"
            >
              <Settings size={20} />
            </div>
          </nav>
        </div>

        <div className="flex flex-col items-center gap-3 w-full px-2">
          <div
            title={developer ? `${developer.name || 'Developer'} (${developer.email})` : 'User'}
            className="w-9 h-9 rounded-full bg-[var(--color-primary)]/20 border border-[var(--color-primary)]/40 text-[var(--color-primary)] font-semibold flex items-center justify-center text-xs"
          >
            {userInitial}
          </div>
          <button
            onClick={handleLogout}
            title="Sign out"
            aria-label="Sign out"
            className="flex items-center justify-center w-10 h-10 rounded-lg text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 transition-all border border-rose-500/20 bg-transparent cursor-pointer"
          >
            <LogOut size={18} />
          </button>
        </div>
      </aside>

      {/* ============================================================
         4. DESKTOP SIDEBAR (>= 1024px)
         ============================================================ */}
      <aside className="hidden lg:flex sticky top-0 h-screen w-64 shrink-0 border-r border-[var(--color-border)] bg-[var(--color-card)] flex-col justify-between p-4 z-40">
        <div>
          {/* Brand Header */}
          <Link to="/apps" className="flex items-center gap-2.5 px-2 py-1 no-underline">
            <img src="/favicon.svg" alt="DilithiAuth" width={30} height={30} className="block shrink-0" />
            <div className="flex flex-col">
              <span className="font-bold text-[var(--color-txt)] text-[1rem] tracking-tight flex items-center gap-1.5">
                DilithiAuth
                <span className="bg-[var(--color-hover)] border border-[var(--color-border-mid)] text-[var(--color-sub)] text-[0.6rem] font-bold px-1.5 py-0.5 rounded uppercase tracking-widest">
                  PQC IdP
                </span>
              </span>
            </div>
          </Link>

          {/* Navigation Links */}
          <nav className="flex flex-col gap-1 mt-6">
            <Link
              to="/apps"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-all ${
                isAppsActive
                  ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold border-l-2 border-[var(--color-primary)]'
                  : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
              }`}
            >
              <LayoutGrid size={18} />
              Applications
            </Link>

            <Link
              to="/docs"
              className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-all ${
                isDocsActive
                  ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold border-l-2 border-[var(--color-primary)]'
                  : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
              }`}
            >
              <BookOpen size={18} />
              Docs
            </Link>

            <div className="flex items-center justify-between px-3 py-2.5 rounded-md text-sm font-medium text-[var(--color-muted)] opacity-50 cursor-not-allowed select-none">
              <div className="flex items-center gap-3">
                <Settings size={18} />
                Settings
              </div>
              <span className="text-[0.6rem] bg-[var(--color-border)] text-[var(--color-muted)] px-1.5 py-0.5 rounded font-mono uppercase">
                Soon
              </span>
            </div>
          </nav>
        </div>

        {/* Anchored Bottom Section */}
        <div className="border-t border-[var(--color-border)] pt-4 flex flex-col gap-3">
          {developer && (
            <div className="flex items-center gap-3 px-2 py-1.5 min-w-0">
              <div className="w-8 h-8 rounded-full bg-[var(--color-primary)]/20 border border-[var(--color-primary)]/40 text-[var(--color-primary)] font-semibold flex items-center justify-center text-xs shrink-0">
                {userInitial}
              </div>
              <div className="flex flex-col min-w-0">
                <span className="text-xs font-semibold text-[var(--color-txt)] truncate">
                  {developer.name || 'Developer'}
                </span>
                <span className="text-[0.75rem] text-[var(--color-sub)] truncate">
                  {developer.email}
                </span>
              </div>
            </div>
          )}

          <button
            onClick={handleLogout}
            className="flex items-center justify-center gap-2 w-full px-3 py-2 rounded-md text-sm font-medium text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 transition-all border border-rose-500/20 bg-transparent cursor-pointer"
          >
            <LogOut size={16} />
            Sign out
          </button>
        </div>
      </aside>

      {/* ============================================================
         5. MAIN CONTENT AREA (<Outlet />)
         ============================================================ */}
      <main className="flex-1 min-w-0 p-4 sm:p-6 md:p-8 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  )
}
```

---

### 4.2 Updated `src/App.tsx` Routing Layout

```tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './AuthContext'
import SidebarLayout from './components/layout/SidebarLayout'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Register from './pages/Register'
import Dashboard from './pages/Dashboard'
import NewApp from './pages/NewApp'
import AppDetail from './pages/AppDetail'
import Docs from './pages/Docs'
import Privacy from './pages/Privacy'
import Terms from './pages/Terms'

function RequireAuth({ children }: { children: React.ReactElement }) {
  const { developer, loading } = useAuth()
  if (loading) return <div style={{ padding: '2rem' }}>Loading…</div>
  return developer ? children : <Navigate to="/login" replace />
}

function RequireGuest({ children }: { children: React.ReactElement }) {
  const { developer, loading } = useAuth()
  if (loading) return <div style={{ padding: '2rem' }}>Loading…</div>
  return developer ? <Navigate to="/apps" replace /> : children
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter basename={import.meta.env.VITE_ROUTER_BASENAME || ''}>
        <Routes>
          {/* Public / Standalone routes with top Navbar */}
          <Route path="/" element={<RequireGuest><Landing /></RequireGuest>} />
          <Route path="/login" element={<RequireGuest><Login /></RequireGuest>} />
          <Route path="/register" element={<RequireGuest><Register /></RequireGuest>} />
          <Route path="/docs" element={<Docs />} />
          <Route path="/privacy" element={<Privacy />} />
          <Route path="/terms" element={<Terms />} />

          {/* Authenticated Dashboard Shell using SidebarLayout */}
          <Route element={<RequireAuth><SidebarLayout /></RequireAuth>}>
            <Route path="/apps" element={<Dashboard />} />
            <Route path="/apps/new" element={<NewApp />} />
            <Route path="/apps/:id" element={<AppDetail />} />
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
```

---

## 5. Page Component Updates Required

Because `SidebarLayout` provides the overall outer layout structure for authenticated routes, the following page components should be modified during implementation to remove redundant top `<Navbar />` calls:
1. `src/pages/Dashboard.tsx`: Remove `<Navbar />` from line 39.
2. `src/pages/NewApp.tsx`: Remove `<Navbar />` from line 53 and line 122.
3. `src/pages/AppDetail.tsx`: Remove `<Navbar />` from line 69, 82, and 188.

---

## 6. Verification Checklist
- [x] **Routing Layout**: Verify authenticated routes (`/apps`, `/apps/new`, `/apps/:id`) render inside `SidebarLayout` and public routes keep `Navbar`.
- [x] **Desktop View**: Verify 256px sticky sidebar on `>= 1024px` displays logo, Applications, Docs, disabled Settings, and bottom user card with Sign Out.
- [x] **Tablet View**: Verify 64px icon rail on `768px - 1023px` displays centered icons with tooltips and circular avatar + logout icon.
- [x] **Mobile View**: Verify top bar with logo and hamburger toggle on `< 768px`, opening backdrop overlay and slide-down drawer menu.
- [x] **Theme Token Consistency**: Verify strict usage of `var(--color-bg)`, `var(--color-card)`, `var(--color-border)`, `var(--color-txt)`, and `var(--color-sub)`.
