import { useState } from "react"
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'

interface CreatedResult { app: OAuthApp; secret?: string }

export default function NewApp() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [homepageURL, setHomepageURL] = useState('')
  const [redirectURIs, setRedirectURIs] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [created, setCreated] = useState<CreatedResult | null>(null)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    const uris = redirectURIs.split('\n').map(s => s.trim()).filter(Boolean)
    if (uris.length === 0) { setError('At least one redirect URI is required.'); return }
    setLoading(true)
    try {
      const r = await apps.create({ name, description, homepage_url: homepageURL, redirect_uris: uris, is_public: isPublic })
      setCreated({ app: r.data.app, secret: r.data.client_secret })
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Failed to create app.')
    } finally {
      setLoading(false)
    }
  }

  const copySecret = () => {
    if (created?.secret) {
      navigator.clipboard.writeText(created.secret)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  if (created) {
    return (
      <>
        <Navbar />
        <div className="container-sm">
          <div className="card" style={{ marginTop: '2rem' }}>
            <h1>Application created! 🎉</h1>
            <p className="text-muted" style={{ marginBottom: '1.25rem', fontSize: '0.9rem' }}>
              Save your credentials now. The client secret <strong>will not be shown again.</strong>
            </p>

            <div className="form-group">
              <label className="label">Client ID</label>
              <div className="secret-box" style={{ color: '#93c5fd' }}>{created.app.ClientID}</div>
            </div>

            {created.secret && (
              <div className="form-group">
                <div className="alert alert-warning" style={{ marginBottom: '0.5rem' }}>
                  ⚠️ Copy this secret now — it will <strong>never</strong> be shown again.
                </div>
                <label className="label">Client Secret</label>
                <div className="secret-box">{created.secret}</div>
                <button className="btn btn-outline btn-sm mt-1" onClick={copySecret}>
                  {copied ? '✓ Copied!' : 'Copy to clipboard'}
                </button>
              </div>
            )}

            {isPublic && (
              <div className="alert alert-info">
                This is a <strong>public</strong> app — use PKCE instead of a client secret.
              </div>
            )}

            <div className="divider" />
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button className="btn btn-primary" onClick={() => navigate(`/apps/${created.app.ID}`)}>
                View app →
              </button>
              <button className="btn btn-outline" onClick={() => navigate('/apps')}>
                Back to dashboard
              </button>
            </div>
          </div>
        </div>
      </>
    )
  }

  return (
    <>
      <Navbar />
      <div className="container-sm">
        <div className="card" style={{ marginTop: '2rem' }}>
          <h1 style={{ marginBottom: '0.25rem' }}>New Application</h1>
          <p className="text-muted" style={{ marginBottom: '1.5rem', fontSize: '0.9rem' }}>
            Register an OAuth 2.0 app to receive a client ID and secret.
          </p>
          {error && <div className="alert alert-error">{error}</div>}
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Application name *</label>
              <input type="text" value={name} onChange={e => setName(e.target.value)} required placeholder="My App" autoFocus />
            </div>
            <div className="form-group">
              <label className="label">Homepage URL</label>
              <input type="text" value={homepageURL} onChange={e => setHomepageURL(e.target.value)} placeholder="https://myapp.com" />
            </div>
            <div className="form-group">
              <label className="label">Description</label>
              <textarea value={description} onChange={e => setDescription(e.target.value)} placeholder="What does your app do?" style={{ minHeight: '60px' }} />
            </div>
            <div className="form-group">
              <label className="label">Redirect URIs *</label>
              <textarea
                value={redirectURIs}
                onChange={e => setRedirectURIs(e.target.value)}
                required
                placeholder={"https://myapp.com/callback\nhttps://myapp.com/auth/callback"}
              />
              <p className="form-hint">One URI per line. Must use HTTPS (http://localhost is allowed for development).</p>
            </div>
            <div className="form-group">
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', cursor: 'pointer' }}>
                <input type="checkbox" checked={isPublic} onChange={e => setIsPublic(e.target.checked)} style={{ width: 'auto' }} />
                <span style={{ fontWeight: 500 }}>Public client (PKCE only — no client secret)</span>
              </label>
              <p className="form-hint">Use for browser SPAs and mobile apps where secrets can't be kept confidential.</p>
            </div>
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button type="submit" className="btn btn-primary" disabled={loading}>
                {loading ? 'Creating…' : 'Create application'}
              </button>
              <button type="button" className="btn btn-outline" onClick={() => navigate('/apps')}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </>
  )
}
