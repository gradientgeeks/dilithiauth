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
  const [mobileOpen, setMobileOpen] = useState(false)

  // Close menu on route change
  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  // Close menu on ESC key
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMobileOpen(false)
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [])

  const handleLogout = async () => {
    try { await logout() } catch { /* ignore */ }
    setMobileOpen(false)
    navigate('/login')
  }

  const isActive = (path: string) =>
    location.pathname === path ? 'active' : ''

  const guestLinks = (
    <>
      <Link to="/" className={`nav-link ${isActive('/')}`}>
        <ShieldCheck size={16} />
        <span>Overview</span>
      </Link>
      <Link to="/docs" className={`nav-link ${isActive('/docs')}`}>
        <BookOpen size={16} />
        <span>SDK & Docs</span>
      </Link>
      <Link to="/login" className={`nav-link ${isActive('/login')}`}>
        <LogIn size={16} />
        <span>Sign in</span>
      </Link>
      <Link to="/register" className="nav-link nav-cta">
        <UserPlus size={16} />
        <span>Get Started</span>
      </Link>
    </>
  )

  const authLinks = (
    <>
      <span className="nav-email-pill">
        <Mail size={13} />
        <span>{developer?.email}</span>
      </span>
      <Link to="/apps" className={`nav-link ${isActive('/apps')}`}>
        <LayoutGrid size={16} />
        <span>Applications</span>
      </Link>
      <Link to="/docs" className={`nav-link ${isActive('/docs')}`}>
        <BookOpen size={16} />
        <span>SDK & Docs</span>
      </Link>
      <button className="nav-link nav-signout" onClick={handleLogout}>
        <LogOut size={16} />
        <span>Sign out</span>
      </button>
    </>
  )

  return (
    <>
      <nav className="navbar">
        <Link to={developer ? '/apps' : '/'} className="nav-brand">
          <img src="/favicon.svg" alt="DilithiAuth" width={28} height={28} />
          <span className="nav-brand-name">
            DilithiAuth
            <span className="nav-brand-badge">PQC IdP</span>
          </span>
        </Link>

        {/* Desktop links */}
        <div className="nav-desktop-links">
          {developer ? authLinks : guestLinks}
        </div>

        {/* Hamburger — shown on tablet/mobile */}
        <button
          className="nav-hamburger"
          onClick={() => setMobileOpen(v => !v)}
          aria-label={mobileOpen ? 'Close menu' : 'Open menu'}
          aria-expanded={mobileOpen}
        >
          {mobileOpen ? <X size={22} /> : <Menu size={22} />}
        </button>
      </nav>

      {/* Slide-down mobile drawer */}
      {mobileOpen && (
        <>
          <div
            className="nav-backdrop"
            onClick={() => setMobileOpen(false)}
            aria-hidden="true"
          />
          <div className="nav-mobile-drawer" role="dialog" aria-label="Navigation menu">
            <div className="nav-drawer-inner">
              {developer && (
                <div className="nav-drawer-user">
                  <Mail size={14} />
                  <span>{developer.email}</span>
                </div>
              )}
              <div className="nav-drawer-links">
                {developer ? (
                  <>
                    <Link to="/apps" className={`drawer-link ${isActive('/apps')}`}>
                      <LayoutGrid size={18} /> Applications
                    </Link>
                    <Link to="/docs" className={`drawer-link ${isActive('/docs')}`}>
                      <BookOpen size={18} /> SDK & Docs
                    </Link>
                    <button className="drawer-link drawer-signout" onClick={handleLogout}>
                      <LogOut size={18} /> Sign out
                    </button>
                  </>
                ) : (
                  <>
                    <Link to="/" className={`drawer-link ${isActive('/')}`}>
                      <ShieldCheck size={18} /> Overview
                    </Link>
                    <Link to="/docs" className={`drawer-link ${isActive('/docs')}`}>
                      <BookOpen size={18} /> SDK & Docs
                    </Link>
                    <Link to="/login" className={`drawer-link ${isActive('/login')}`}>
                      <LogIn size={18} /> Sign in
                    </Link>
                    <Link to="/register" className="drawer-link drawer-cta">
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
