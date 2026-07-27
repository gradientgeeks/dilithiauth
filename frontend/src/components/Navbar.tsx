import { Link, useLocation, useNavigate } from 'react-router-dom'
import { LogOut, LayoutGrid, Mail, ShieldCheck, LogIn, UserPlus, BookOpen } from 'lucide-react'
import { useAuth } from '../AuthContext'

export default function Navbar() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const handleLogout = async () => {
    await logout()
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

      <div className="nav-actions">
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
    </nav>
  )
}
