import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Mail, Lock, AlertCircle, Loader2, LogIn, ShieldCheck } from 'lucide-react'
import { useAuth } from '../AuthContext'
import Navbar from '../components/Navbar'

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      navigate('/apps')
    } catch {
      setError('Invalid email or password.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Navbar />
      <div className="container-sm">
        <div className="card card-glow" style={{ marginTop: '2.5rem' }}>
          <div style={{ textAlign: 'center', marginBottom: '1.75rem' }}>
            <div style={{ display: 'inline-flex', background: 'rgba(37, 99, 235, 0.15)', padding: '0.75rem', borderRadius: '14px', color: '#93c5fd', marginBottom: '0.75rem', border: '1px solid rgba(56, 189, 248, 0.3)' }}>
              <ShieldCheck size={28} />
            </div>
            <h1>Sign in to Portal</h1>
            <p className="text-muted" style={{ fontSize: '0.9rem' }}>
              Access your quantum-safe developer dashboard
            </p>
          </div>

          {error && (
            <div className="alert alert-error">
              <AlertCircle size={18} />
              <div>{error}</div>
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Developer Email</label>
              <div className="input-icon-wrap">
                <input
                  type="email"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  required
                  placeholder="developer@example.com"
                  autoFocus
                />
                <Mail size={16} />
              </div>
            </div>

            <div className="form-group">
              <label className="label">Password</label>
              <div className="input-icon-wrap">
                <input
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  required
                  placeholder="••••••••••••"
                />
                <Lock size={16} />
              </div>
            </div>

            <button type="submit" className="btn btn-primary btn-full" disabled={loading} style={{ marginTop: '0.5rem', padding: '0.75rem' }}>
              {loading ? (
                <><Loader2 size={18} className="spin" /> Signing in…</>
              ) : (
                <><LogIn size={18} /> Sign in to Dashboard</>
              )}
            </button>
          </form>

          <div style={{ textAlign: 'center', marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.25rem' }}>
            <p className="text-muted" style={{ fontSize: '0.88rem' }}>
              Don't have a developer account?{' '}
              <Link to="/register" style={{ fontWeight: 600 }}>Create one now →</Link>
            </p>
          </div>
        </div>
      </div>
    </>
  )
}
