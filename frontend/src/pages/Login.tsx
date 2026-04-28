import { useState } from "react"
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
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
        <div className="card" style={{ marginTop: '2rem' }}>
          <h1 style={{ marginBottom: '0.25rem' }}>Sign in</h1>
          <p className="text-muted" style={{ marginBottom: '1.5rem', fontSize: '0.9rem' }}>
            Access your developer dashboard
          </p>
          {error && <div className="alert alert-error">{error}</div>}
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Email</label>
              <input
                type="email" value={email} onChange={e => setEmail(e.target.value)}
                required autoFocus placeholder="you@example.com"
              />
            </div>
            <div className="form-group">
              <label className="label">Password</label>
              <input
                type="password" value={password} onChange={e => setPassword(e.target.value)}
                required placeholder="••••••••"
              />
            </div>
            <button type="submit" className="btn btn-primary btn-full" disabled={loading}>
              {loading ? 'Signing in…' : 'Sign in'}
            </button>
          </form>
          <p className="text-muted mt-2" style={{ fontSize: '0.88rem', textAlign: 'center' }}>
            No account? <Link to="/register">Create one</Link>
          </p>
        </div>
      </div>
    </>
  )
}
