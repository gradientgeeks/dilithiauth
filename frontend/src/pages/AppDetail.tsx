import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'

export default function AppDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [app, setApp] = useState<OAuthApp | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [rotatedSecret, setRotatedSecret] = useState<string | null>(null)
  const [rotating, setRotating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [copiedId, setCopiedId] = useState(false)
  const [copiedSecret, setCopiedSecret] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  useEffect(() => {
    if (!id) return
    apps.get(id)
      .then(r => setApp(r.data))
      .catch(() => setError('App not found.'))
      .finally(() => setLoading(false))
  }, [id])

  const handleRotate = async () => {
    if (!id || !app) return
    if (!window.confirm('Rotate the client secret? The old secret will stop working immediately.')) return
    setRotating(true)
    try {
      const r = await apps.rotateSecret(id)
      setRotatedSecret(r.data.client_secret)
      setApp(r.data.app)
    } catch {
      setError('Failed to rotate secret.')
    } finally {
      setRotating(false)
    }
  }

  const handleDelete = async () => {
    if (!id) return
    setDeleting(true)
    try {
      await apps.delete(id)
      navigate('/apps')
    } catch {
      setError('Failed to delete app.')
      setDeleting(false)
      setConfirmDelete(false)
    }
  }

  const copy = (text: string, setCopied: (v: boolean) => void) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (loading) return <><Navbar /><div className="container"><p>Loading…</p></div></>
  if (error && !app) return <><Navbar /><div className="container"><div className="alert alert-error">{error}</div><Link to="/apps">← Back</Link></div></>
  if (!app) return null

  return (
    <>
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div>
            <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '0.3rem' }}>
              <Link to="/apps">My Apps</Link> / {app.Name}
            </div>
            <h1>{app.Name}</h1>
            {app.Description && <p>{app.Description}</p>}
          </div>
          <span className={`badge ${app.IsPublic ? 'badge-public' : 'badge-confidential'}`} style={{ fontSize: '0.8rem', padding: '0.3rem 0.7rem' }}>
            {app.IsPublic ? 'Public' : 'Confidential'}
          </span>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        {/* Credentials */}
        <div className="card">
          <h2>Credentials</h2>
          <div className="form-group">
            <label className="label">Client ID</label>
            <div className="flex items-center gap-2">
              <span className="mono" style={{ flex: 1, display: 'block', padding: '0.55rem 0.8rem', background: '#f1f5f9', borderRadius: '6px' }}>{app.ClientID}</span>
              <button className="btn btn-outline btn-sm" onClick={() => copy(app.ClientID, setCopiedId)}>
                {copiedId ? '✓' : 'Copy'}
              </button>
            </div>
          </div>

          {!app.IsPublic && (
            <div>
              {rotatedSecret ? (
                <div>
                  <div className="alert alert-warning">
                    ⚠️ New secret shown <strong>once</strong>. Copy it now.
                  </div>
                  <label className="label">New Client Secret</label>
                  <div className="secret-box">{rotatedSecret}</div>
                  <div className="flex gap-2 mt-1">
                    <button className="btn btn-outline btn-sm" onClick={() => copy(rotatedSecret, setCopiedSecret)}>
                      {copiedSecret ? '✓ Copied!' : 'Copy'}
                    </button>
                    <button className="btn btn-outline btn-sm" onClick={() => setRotatedSecret(null)}>
                      I've saved it
                    </button>
                  </div>
                </div>
              ) : (
                <div>
                  <label className="label">Client Secret</label>
                  <div className="flex items-center gap-2">
                    <span style={{ color: 'var(--text-muted)', fontFamily: 'monospace', letterSpacing: '0.1em' }}>••••••••••••••••••••</span>
                    <button className="btn btn-outline btn-sm" onClick={handleRotate} disabled={rotating}>
                      {rotating ? 'Rotating…' : 'Rotate secret'}
                    </button>
                  </div>
                  {app.SecretRotatedAt && (
                    <p className="form-hint mt-1">Last rotated: {new Date(app.SecretRotatedAt).toLocaleString()}</p>
                  )}
                </div>
              )}
            </div>
          )}

          {app.IsPublic && (
            <div className="alert alert-info mt-1">
              Public app — use PKCE (<code>code_challenge</code> + <code>code_verifier</code>). No client secret required.
            </div>
          )}
        </div>

        {/* App details */}
        <div className="card">
          <h2>Details</h2>
          <div className="detail-grid">
            <div>
              <div className="detail-label">Homepage URL</div>
              <div className="detail-value">{app.HomepageURL || <span className="text-muted">—</span>}</div>
            </div>
            <div>
              <div className="detail-label">Created</div>
              <div className="detail-value">{new Date(app.CreatedAt).toLocaleString()}</div>
            </div>
          </div>
          <div className="divider" />
          <div className="detail-label" style={{ marginBottom: '0.5rem' }}>Redirect URIs</div>
          {app.RedirectURIs.map(uri => (
            <div key={uri} className="mono" style={{ marginBottom: '0.35rem', display: 'block', padding: '0.35rem 0.6rem', background: '#f1f5f9', borderRadius: '4px' }}>
              {uri}
            </div>
          ))}
        </div>

        {/* Integration guide */}
        <div className="card">
          <h2>Quick Integration</h2>
          <p className="text-muted" style={{ marginBottom: '1rem', fontSize: '0.88rem' }}>
            Redirect your users to the authorization endpoint:
          </p>
          <div className="secret-box" style={{ color: '#94a3b8', fontSize: '0.8rem' }}>
            {`https://dilithiauth.gradientgeeks.tech/authorize\n  ?client_id=${app.ClientID}\n  &redirect_uri=YOUR_REDIRECT_URI\n  &response_type=code\n  &state=RANDOM_STATE${app.IsPublic ? '\n  &code_challenge=PKCE_CHALLENGE\n  &code_challenge_method=S256' : ''}`}
          </div>
          <p className="text-muted mt-1" style={{ fontSize: '0.85rem' }}>
            Then exchange the code at <code style={{ background: '#f1f5f9', padding: '0.1rem 0.4rem', borderRadius: 4 }}>POST /token</code>
            {!app.IsPublic ? ' using HTTP Basic Auth with your client_id and client_secret.' : ' with your code_verifier.'}
          </p>
        </div>

        {/* Danger zone */}
        <div className="card" style={{ borderColor: '#fca5a5' }}>
          <h2 style={{ color: 'var(--danger)' }}>Danger Zone</h2>
          <p className="text-muted" style={{ marginBottom: '1rem', fontSize: '0.9rem' }}>
            Deleting this app will immediately revoke all issued tokens and disable the integration for your users.
          </p>
          {confirmDelete ? (
            <div className="flex gap-2">
              <button className="btn btn-danger" onClick={handleDelete} disabled={deleting}>
                {deleting ? 'Deleting…' : 'Yes, delete permanently'}
              </button>
              <button className="btn btn-outline" onClick={() => setConfirmDelete(false)}>Cancel</button>
            </div>
          ) : (
            <button className="btn btn-danger btn-outline" style={{ borderColor: 'var(--danger)', color: 'var(--danger)', background: 'transparent' }}
              onClick={() => setConfirmDelete(true)}>
              Delete application
            </button>
          )}
        </div>
      </div>
    </>
  )
}
