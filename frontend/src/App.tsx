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
