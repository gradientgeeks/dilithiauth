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

          <div style={{ margin: '1.5rem 0', textAlign: 'center', position: 'relative' }}>
            <div style={{ height: '1px', background: 'var(--border-color)', width: '100%', position: 'absolute', top: '50%', zIndex: 0 }}></div>
            <span style={{ background: 'var(--card-bg)', padding: '0 0.75rem', color: 'var(--text-muted)', fontSize: '0.8rem', position: 'relative', zIndex: 1 }}>
              or continue with
            </span>
          </div>

          <div style={{ display: 'flex', gap: '0.75rem' }}>
            <a
              href="https://dilithiauthapis.gradientgeeks.tech/auth/social/google"
              className="btn btn-outline"
              style={{ flex: 1, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', textDecoration: 'none', padding: '0.65rem' }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24">
                <path fill="#EA4335" d="M12 5c1.6 0 3 .6 4.1 1.6l3.1-3.1C17.3 1.8 14.8 1 12 1 7.5 1 3.7 3.6 1.9 7.3l3.7 2.9C6.5 7.4 9 5 12 5z"/>
                <path fill="#4285F4" d="M23.5 12.3c0-.8-.1-1.6-.2-2.3H12v4.5h6.5c-.3 1.5-1.1 2.8-2.4 3.7l3.7 2.9c2.2-2 3.7-5 3.7-8.8z"/>
                <path fill="#FBBC05" d="M5.6 14.8c-.2-.7-.4-1.5-.4-2.3s.2-1.6.4-2.3L1.9 7.3C.7 9.7 0 12.4 0 15.3s.7 5.6 1.9 8l3.7-2.9z"/>
                <path fill="#34A853" d="M12 23.5c3.2 0 6-1.1 8-3l-3.7-2.9c-1.1.7-2.5 1.2-4.3 1.2-3 0-5.5-2.4-6.4-5.2L1.9 16.5C3.7 20.2 7.5 23.5 12 23.5z"/>
              </svg>
              Google
            </a>
            <a
              href="https://dilithiauthapis.gradientgeeks.tech/auth/social/github"
              className="btn btn-outline"
              style={{ flex: 1, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', textDecoration: 'none', padding: '0.65rem' }}
            >
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
              </svg>
              GitHub
            </a>
          </div>

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
