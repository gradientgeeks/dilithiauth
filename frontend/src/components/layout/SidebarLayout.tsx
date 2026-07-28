import { useState, useEffect } from 'react'
import { Link, useLocation, useNavigate, Outlet } from 'react-router-dom'
import {
  LayoutGrid, BookOpen, Settings, LogOut, Menu, X
} from 'lucide-react'
import { useAuth } from '../../AuthContext'

export default function SidebarLayout() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [isMobileOpen, setIsMobileOpen] = useState(false)

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
  const userInitial = developer?.name && developer.name.trim().length > 0
    ? developer.name.trim().charAt(0).toUpperCase()
    : developer?.email && developer.email.trim().length > 0
    ? developer.email.trim().charAt(0).toUpperCase()
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
              ML-DSA-44
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
                onClick={() => setIsMobileOpen(false)}
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
                onClick={() => setIsMobileOpen(false)}
                className={`flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-all ${
                  isDocsActive
                    ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold border-l-2 border-[var(--color-primary)]'
                    : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
                }`}
              >
                <BookOpen size={18} /> Docs
              </Link>
              <div className="flex items-center justify-between px-3 py-2.5 rounded-md text-sm font-medium text-[var(--color-muted)] opacity-50 cursor-not-allowed select-none">
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
              className="flex items-center justify-center w-12 h-12 rounded-lg text-[var(--color-muted)] opacity-40 cursor-not-allowed select-none"
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
                  ML-DSA-44
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
