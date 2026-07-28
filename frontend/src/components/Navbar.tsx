import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { LogOut, LayoutGrid, Mail, ShieldCheck, LogIn, UserPlus, BookOpen, Menu, X } from 'lucide-react'
import { useAuth } from '../AuthContext'

export default function Navbar() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileOpen, setMobileOpen] = useState(false)

  const handleLogout = async () => {
    await logout()
    setMobileOpen(false)
    navigate('/login')
  }

  const isActive = (path: string) => location.pathname === path

  return (
    <nav>
      <Link to={developer ? '/apps' : '/'} className="nav-brand">
        <img src="/favicon.svg" alt="DilithiAuth Logo" width={28} height={28} />
        <div>
          DilithiAuth <span className="nav-brand-badge">PQC IdP</span>
        </div>
      </Link>

      {/* Desktop Navigation Links */}
      <div className="nav-actions nav-desktop">
        {developer ? (
          <>
            <span className="nav-email-pill">
              <Mail size={13} color="#9ca3af" /> {developer.email}
            </span>
            <Link to="/apps" className={isActive('/apps') ? 'active' : ''}>
              <LayoutGrid size={15} /> Applications
            </Link>
            <Link to="/docs" className={isActive('/docs') ? 'active' : ''}>
              <BookOpen size={15} /> SDK & Docs
            </Link>
            <button onClick={handleLogout}>
              <LogOut size={15} /> Sign out
            </button>
          </>
        ) : (
          <>
            <Link to="/" className={isActive('/') ? 'active' : ''}>
              <ShieldCheck size={15} /> Overview
            </Link>
            <Link to="/docs" className={isActive('/docs') ? 'active' : ''}>
              <BookOpen size={15} /> SDK & Docs
            </Link>
            <Link to="/login" className={isActive('/login') ? 'active' : ''}>
              <LogIn size={15} /> Sign in
            </Link>
            <Link to="/register" className="nav-register">
              <UserPlus size={15} /> Get Started
            </Link>
          </>
        )}
      </div>

      {/* Mobile Toggle Button (<768px) */}
      <button
        className="nav-mobile-toggle"
        onClick={() => setMobileOpen(!mobileOpen)}
        aria-label="Toggle navigation menu"
      >
        {mobileOpen ? <X size={20} /> : <Menu size={20} />}
      </button>

      {/* Mobile Dropdown Menu */}
      {mobileOpen && (
        <div className="nav-mobile-menu">
          {developer ? (
            <>
              <div className="nav-email-pill" style={{ marginBottom: '0.5rem', width: '100%', justifyContent: 'center' }}>
                <Mail size={13} color="#9ca3af" /> {developer.email}
              </div>
              <Link to="/apps" className={isActive('/apps') ? 'active' : ''} onClick={() => setMobileOpen(false)}>
                <LayoutGrid size={16} /> Applications
              </Link>
              <Link to="/docs" className={isActive('/docs') ? 'active' : ''} onClick={() => setMobileOpen(false)}>
                <BookOpen size={16} /> SDK & Docs
              </Link>
              <button onClick={handleLogout} style={{ width: '100%', justifyContent: 'center' }}>
                <LogOut size={16} /> Sign out
              </button>
            </>
          ) : (
            <>
              <Link to="/" className={isActive('/') ? 'active' : ''} onClick={() => setMobileOpen(false)}>
                <ShieldCheck size={16} /> Overview
              </Link>
              <Link to="/docs" className={isActive('/docs') ? 'active' : ''} onClick={() => setMobileOpen(false)}>
                <BookOpen size={16} /> SDK & Docs
              </Link>
              <Link to="/login" className={isActive('/login') ? 'active' : ''} onClick={() => setMobileOpen(false)}>
                <LogIn size={16} /> Sign in
              </Link>
              <Link to="/register" className="nav-register" onClick={() => setMobileOpen(false)}>
                <UserPlus size={16} /> Get Started
              </Link>
            </>
          )}
        </div>
      )}
    </nav>
  )
}
