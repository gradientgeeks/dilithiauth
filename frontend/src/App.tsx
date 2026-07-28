import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './AuthContext'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Register from './pages/Register'
import Dashboard from './pages/Dashboard'
import NewApp from './pages/NewApp'
import AppDetail from './pages/AppDetail'
import DocsLoader from './pages/DocsLoader'  // ← Micro frontend loader (lazy remote)
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
          <Route path="/" element={<RequireGuest><Landing /></RequireGuest>} />
          <Route path="/login" element={<RequireGuest><Login /></RequireGuest>} />
          <Route path="/register" element={<RequireGuest><Register /></RequireGuest>} />

          {/*
           * /docs is served by the Docs MICRO FRONTEND (docs-frontend/).
           * DocsLoader wraps the lazy remote import with a Suspense skeleton
           * and an Error boundary — no bundle-time dependency on Docs.tsx.
           */}
          <Route path="/docs" element={<DocsLoader />} />

          <Route path="/privacy" element={<Privacy />} />
          <Route path="/terms" element={<Terms />} />
          <Route path="/apps" element={<RequireAuth><Dashboard /></RequireAuth>} />
          <Route path="/apps/new" element={<RequireAuth><NewApp /></RequireAuth>} />
          <Route path="/apps/:id" element={<RequireAuth><AppDetail /></RequireAuth>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
