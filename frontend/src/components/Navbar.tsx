import { useState, useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import {
  LogOut, LayoutGrid, Mail, ShieldCheck, LogIn,
  UserPlus, BookOpen, Menu, X
} from 'lucide-react'
import { useAuth } from '../AuthContext'

export default function Navbar() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [open, setOpen] = useState(false)

  useEffect(() => { setOpen(false) }, [location.pathname])
  useEffect(() => {
    const fn = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    document.addEventListener('keydown', fn)
    return () => document.removeEventListener('keydown', fn)
  }, [])

  const handleLogout = async () => {
    try { await logout() } catch { /* ignore */ }
    setOpen(false)
    navigate('/login')
  }

  const isActive = (path: string) => location.pathname === path

  const NavLink = ({ to, icon: Icon, children }: { to: string; icon: React.ElementType; children: React.ReactNode }) => (
    <Link
      to={to}
      className={`inline-flex items-center gap-1.5 px-3 py-2 rounded-md text-sm font-medium transition-all duration-150
        ${isActive(to)
          ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold'
          : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
        }`}
    >
      <Icon size={16} />
      {children}
    </Link>
  )

  const DrawerLink = ({ to, icon: Icon, children, className = '' }: { to: string; icon: React.ElementType; children: React.ReactNode; className?: string }) => (
    <Link
      to={to}
      className={`flex items-center gap-3 w-full px-3 py-2.5 rounded-md text-sm font-medium transition-all duration-150
        ${isActive(to)
          ? 'bg-[var(--color-hover)] text-[var(--color-txt)] font-semibold'
          : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
        } ${className}`}
    >
      <Icon size={18} />
      {children}
    </Link>
  )

  return (
    <>
      {/* Main nav bar */}
      <nav
        className="sticky top-0 z-[200] w-full h-[60px] flex items-center justify-between px-4 md:px-6
          bg-[#090d16]/90 backdrop-blur-md border-b border-[var(--color-border)]"
      >
        {/* Brand */}
        <Link
          to={developer ? '/apps' : '/'}
          className="flex items-center gap-2 no-underline hover:no-underline"
        >
          <img src="/favicon.svg" alt="DilithiAuth" width={28} height={28} className="block shrink-0" />
          <span className="flex items-center gap-2 font-bold text-[var(--color-txt)] text-[0.97rem] tracking-tight">
            DilithiAuth
            <span className="bg-[var(--color-hover)] border border-[var(--color-border-mid)] text-[var(--color-sub)]
              text-[0.62rem] font-bold px-1.5 py-0.5 rounded uppercase tracking-widest">
              PQC IdP
            </span>
          </span>
        </Link>

        {/* Desktop links — hidden on tablet/mobile */}
        <div className="hidden lg:flex items-center gap-1">
          {developer ? (
            <>
              <span className="inline-flex items-center gap-1.5 text-[var(--color-sub)] text-[0.8rem]
                px-2.5 py-1.5 border border-[var(--color-border)] rounded-md max-w-[200px] truncate">
                <Mail size={13} /><span className="truncate">{developer.email}</span>
              </span>
              <NavLink to="/apps" icon={LayoutGrid}>Applications</NavLink>
              <NavLink to="/docs" icon={BookOpen}>SDK & Docs</NavLink>
              <button
                onClick={handleLogout}
                className="inline-flex items-center gap-1.5 px-3 py-2 rounded-md text-sm font-medium
                  text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 transition-all duration-150 cursor-pointer border-none bg-transparent"
              >
                <LogOut size={16} /> Sign out
              </button>
            </>
          ) : (
            <>
              <NavLink to="/" icon={ShieldCheck}>Overview</NavLink>
              <NavLink to="/docs" icon={BookOpen}>SDK & Docs</NavLink>
              <NavLink to="/login" icon={LogIn}>Sign in</NavLink>
              <Link
                to="/register"
                className="inline-flex items-center gap-1.5 px-3 py-2 rounded-md text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)] hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)]
                  transition-all duration-150 ml-1 no-underline hover:no-underline"
              >
                <UserPlus size={16} /> Get Started
              </Link>
            </>
          )}
        </div>

        {/* Hamburger — shown on tablet & mobile */}
        <button
          onClick={() => setOpen(v => !v)}
          aria-label={open ? 'Close menu' : 'Open menu'}
          aria-expanded={open}
          className="lg:hidden flex items-center justify-center w-9 h-9 rounded-md border border-[var(--color-border)]
            text-[var(--color-txt)] bg-transparent hover:bg-[var(--color-hover)] transition-colors cursor-pointer"
        >
          {open ? <X size={22} /> : <Menu size={22} />}
        </button>
      </nav>

      {/* Mobile drawer */}
      {open && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-[198] bg-black/55 backdrop-blur-sm"
            style={{ animation: 'fadeIn .18s ease-out' }}
            onClick={() => setOpen(false)}
            aria-hidden="true"
          />
          {/* Slide-down drawer */}
          <div
            className="fixed top-[60px] left-0 right-0 z-[199] bg-[var(--color-card)]
              border-b border-[var(--color-border)] shadow-2xl lg:hidden"
            style={{ animation: 'slideDown .22s cubic-bezier(0.16,1,0.3,1)' }}
            role="dialog"
            aria-label="Navigation"
          >
            <div className="max-w-xl mx-auto px-4 py-3 pb-4">
              {developer && (
                <div className="flex items-center gap-2 text-[var(--color-sub)] text-[0.82rem]
                  px-3 py-2.5 border-b border-[var(--color-border)] mb-2 truncate">
                  <Mail size={14} /><span className="truncate">{developer.email}</span>
                </div>
              )}
              <div className="flex flex-col gap-1">
                {developer ? (
                  <>
                    <DrawerLink to="/apps" icon={LayoutGrid}>Applications</DrawerLink>
                    <DrawerLink to="/docs" icon={BookOpen}>SDK & Docs</DrawerLink>
                    <button
                      onClick={handleLogout}
                      className="flex items-center gap-3 w-full px-3 py-2.5 rounded-md text-sm font-medium
                        text-rose-300 hover:bg-rose-500/10 hover:text-rose-400 transition-all duration-150 cursor-pointer border-none bg-transparent mt-1"
                    >
                      <LogOut size={18} /> Sign out
                    </button>
                  </>
                ) : (
                  <>
                    <DrawerLink to="/" icon={ShieldCheck}>Overview</DrawerLink>
                    <DrawerLink to="/docs" icon={BookOpen}>SDK & Docs</DrawerLink>
                    <DrawerLink to="/login" icon={LogIn}>Sign in</DrawerLink>
                    <Link
                      to="/register"
                      className="flex items-center justify-center gap-2 w-full px-3 py-2.5 mt-1.5 rounded-md
                        text-sm font-semibold bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                        transition-all duration-150 no-underline hover:no-underline"
                    >
                      <UserPlus size={18} /> Get Started
                    </Link>
                  </>
                )}
              </div>
            </div>
          </div>
        </>
      )}
    </>
  )
}
