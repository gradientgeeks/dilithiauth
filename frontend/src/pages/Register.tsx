import { useState } from "react"
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../AuthContext'
import Navbar from '../components/Navbar'

export default function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirm) { setError('Passwords do not match.'); return }
    if (password.length < 8) { setError('Password must be at least 8 characters.'); return }
    setLoading(true)
    try {
      await register(email, name, password)
      navigate('/apps')
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Registration failed.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Navbar />
      <div className="container-sm">
        <div className="card" style={{ marginTop: '2rem' }}>
          <h1 style={{ marginBottom: '0.25rem' }}>Create account</h1>
          <p className="text-muted" style={{ marginBottom: '1.5rem', fontSize: '0.9rem' }}>
            Start building with DilithiAuth
          </p>
          {error && <div className="alert alert-error">{error}</div>}
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Full name</label>
              <input type="text" value={name} onChange={e => setName(e.target.value)} required placeholder="Jane Smith" autoFocus />
            </div>
            <div className="form-group">
              <label className="label">Email</label>
              <input type="email" value={email} onChange={e => setEmail(e.target.value)} required placeholder="you@example.com" />
            </div>
            <div className="form-group">
              <label className="label">Password</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)} required placeholder="At least 8 characters" />
            </div>
            <div className="form-group">
              <label className="label">Confirm password</label>
              <input type="password" value={confirm} onChange={e => setConfirm(e.target.value)} required placeholder="••••••••" />
            </div>
            <button type="submit" className="btn btn-primary btn-full" disabled={loading}>
              {loading ? 'Creating account…' : 'Create account'}
            </button>
          </form>
          <p className="text-muted mt-2" style={{ fontSize: '0.88rem', textAlign: 'center' }}>
            Already have an account? <Link to="/login">Sign in</Link>
          </p>
        </div>
      </div>
    </>
  )
}
