import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle2, AlertTriangle, AlertCircle, ShieldCheck, Copy, Check, Loader2, Globe, Lock, ArrowLeft } from 'lucide-react'
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
      setError(err?.response?.data?.error ?? 'Failed to create application.')
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
          <div className="card card-glow" style={{ marginTop: '2.5rem' }}>
            <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
              <div style={{ display: 'inline-flex', background: 'rgba(16, 185, 129, 0.15)', padding: '0.75rem', borderRadius: '14px', color: '#34d399', marginBottom: '0.75rem', border: '1px solid rgba(16, 185, 129, 0.3)' }}>
                <CheckCircle2 size={32} />
              </div>
              <h1>Application Created!</h1>
              <p className="text-muted" style={{ fontSize: '0.9rem' }}>
                Your OAuth 2.0 application credentials are ready.
              </p>
            </div>

            <div className="form-group">
              <label className="label">Client ID</label>
              <div className="secret-box" style={{ color: '#38bdf8', borderColor: 'rgba(56, 189, 248, 0.3)' }}>
                {created.app.ClientID}
              </div>
            </div>

            {created.secret && (
              <div className="form-group">
                <div className="alert alert-warning" style={{ marginBottom: '0.75rem' }}>
                  <AlertTriangle size={18} />
                  <div>Save this Client Secret now! It will <strong>never be shown again</strong>.</div>
                </div>
                <label className="label">Client Secret</label>
                <div className="secret-box">{created.secret}</div>
                <button className="btn btn-outline btn-sm copy-btn mt-1" onClick={copySecret}>
                  {copied ? <><Check size={14} color="var(--emerald)" /> Copied to Clipboard!</> : <><Copy size={14} /> Copy Secret</>}
                </button>
              </div>
            )}

            {isPublic && (
              <div className="alert alert-info">
                <ShieldCheck size={18} />
                <div>Registered as a <strong>Public Client</strong> (PKCE mode). No client secret is required.</div>
              </div>
            )}

            <div className="divider" />
            <div style={{ display: 'flex', gap: '0.75rem' }}>
              <button className="btn btn-primary btn-full" onClick={() => navigate(`/apps/${created.app.ID}`)}>
                View App Details →
              </button>
              <button className="btn btn-outline" onClick={() => navigate('/apps')}>
                Dashboard
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
        <div style={{ marginBottom: '1rem' }}>
          <button className="btn btn-outline btn-sm" onClick={() => navigate('/apps')}>
            <ArrowLeft size={14} /> Back to Applications
          </button>
        </div>

        <div className="card card-glow">
          <h1>Register New Application</h1>
          <p className="text-muted" style={{ marginBottom: '1.5rem', fontSize: '0.9rem' }}>
            Get client credentials to integrate DilithiAuth Post-Quantum identity into your service.
          </p>

          {error && <div className="alert alert-error"><AlertCircle size={18} /><div>{error}</div></div>}

          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label className="label">Application Name *</label>
              <input
                type="text"
                value={name}
                onChange={e => setName(e.target.value)}
                required
                placeholder="My Secure App"
                autoFocus
              />
            </div>

            {/* Client Type Cards */}
            <div className="form-group">
              <label className="label">Client Type</label>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
                <div
                  style={{
                    border: `1px solid ${!isPublic ? 'var(--primary)' : 'var(--border-color)'}`,
                    background: !isPublic ? 'rgba(37, 99, 235, 0.12)' : 'var(--bg-input)',
                    borderRadius: 'var(--radius-sm)',
                    padding: '0.85rem',
                    cursor: 'pointer',
                    transition: 'all 0.2s ease'
                  }}
                  onClick={() => setIsPublic(false)}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontWeight: 700, color: !isPublic ? '#93c5fd' : 'var(--text-main)', fontSize: '0.88rem' }}>
                    <Lock size={15} /> Confidential Client
                  </div>
                  <p style={{ fontSize: '0.78rem', color: 'var(--text-muted)', marginTop: '0.3rem' }}>
                    For server-side web apps with protected backend secrets.
                  </p>
                </div>

                <div
                  style={{
                    border: `1px solid ${isPublic ? 'var(--cyan)' : 'var(--border-color)'}`,
                    background: isPublic ? 'rgba(6, 182, 212, 0.12)' : 'var(--bg-input)',
                    borderRadius: 'var(--radius-sm)',
                    padding: '0.85rem',
                    cursor: 'pointer',
                    transition: 'all 0.2s ease'
                  }}
                  onClick={() => setIsPublic(true)}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', fontWeight: 700, color: isPublic ? '#38bdf8' : 'var(--text-main)', fontSize: '0.88rem' }}>
                    <Globe size={15} /> Public Client (PKCE)
                  </div>
                  <p style={{ fontSize: '0.78rem', color: 'var(--text-muted)', marginTop: '0.3rem' }}>
                    For SPAs, Mobile Apps, or CLI tools without backends.
                  </p>
                </div>
              </div>
            </div>

            <div className="form-group">
              <label className="label">Homepage URL</label>
              <input
                type="text"
                value={homepageURL}
                onChange={e => setHomepageURL(e.target.value)}
                placeholder="https://myapp.com"
              />
            </div>

            <div className="form-group">
              <label className="label">Description</label>
              <textarea
                value={description}
                onChange={e => setDescription(e.target.value)}
                placeholder="Brief summary of your app..."
                style={{ minHeight: '65px' }}
              />
            </div>

            <div className="form-group">
              <label className="label">Redirect URIs *</label>
              <textarea
                value={redirectURIs}
                onChange={e => setRedirectURIs(e.target.value)}
                required
                placeholder={"https://myapp.com/callback\nhttps://myapp.com/auth/callback"}
                style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}
              />
              <p className="form-hint">One URI per line. Must use HTTPS (http://localhost permitted for dev).</p>
            </div>

            <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
              <button type="submit" className="btn btn-primary" disabled={loading} style={{ flex: 1 }}>
                {loading ? <><Loader2 size={16} className="spin" /> Registering App…</> : 'Register Application'}
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
