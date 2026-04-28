import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'

export default function Navbar() {
  const { developer, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <nav>
      <Link to={developer ? '/apps' : '/'} className="nav-brand" style={{ textDecoration: 'none' }}>
        DilithiAuth <span>Developer</span>
      </Link>
      <div className="nav-actions">
        {developer ? (
          <>
            <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: '0.85rem' }}>{developer.email}</span>
            <Link to="/apps">My Apps</Link>
            <button onClick={handleLogout}>Sign out</button>
          </>
        ) : (
          <>
            <Link to="/login">Sign in</Link>
            <Link to="/register" style={{ color: 'white', background: 'rgba(79,142,247,0.3)', padding: '0.3rem 0.75rem', borderRadius: '6px' }}>
              Register
            </Link>
          </>
        )}
      </div>
    </nav>
  )
}
