import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Boxes, AlertCircle, Eye, Search, ShieldCheck, Key, Globe, Lock, Copy, Check } from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'

export default function Dashboard() {
  const [appList, setAppList] = useState<OAuthApp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [copiedId, setCopiedId] = useState<string | null>(null)

  useEffect(() => {
    apps.list()
      .then(r => setAppList(r.data ?? []))
      .catch(() => setError('Failed to load applications.'))
      .finally(() => setLoading(false))
  }, [])

  const copyClientId = (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(id)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const filteredApps = appList.filter(app => 
    app.Name.toLowerCase().includes(search.toLowerCase()) ||
    app.ClientID.toLowerCase().includes(search.toLowerCase())
  )

  const publicCount = appList.filter(a => a.IsPublic).length
  const confidentialCount = appList.filter(a => !a.IsPublic).length

  return (
    <>
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div>
            <h1>Developer Dashboard</h1>
            <p>Manage your OAuth 2.0 applications and client credentials.</p>
          </div>
          <Link to="/apps/new" className="btn btn-primary">
            <Plus size={18} /> Register New Application
          </Link>
        </div>

        {error && <div className="alert alert-error"><AlertCircle size={18} /><div>{error}</div></div>}

        {/* Overview Stats Cards */}
        {!loading && (
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-icon"><Boxes size={22} /></div>
              <div>
                <div className="stat-val">{appList.length}</div>
                <div className="stat-lbl">Registered Apps</div>
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon" style={{ background: 'rgba(6, 182, 212, 0.15)', color: '#38bdf8' }}>
                <Globe size={22} />
              </div>
              <div>
                <div className="stat-val">{publicCount}</div>
                <div className="stat-lbl">Public PKCE Clients</div>
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon" style={{ background: 'rgba(16, 185, 129, 0.15)', color: '#34d399' }}>
                <Lock size={22} />
              </div>
              <div>
                <div className="stat-val">{confidentialCount}</div>
                <div className="stat-lbl">Confidential Clients</div>
              </div>
            </div>

            <div className="stat-card">
              <div className="stat-icon" style={{ background: 'rgba(2, 132, 199, 0.15)', color: '#38bdf8' }}>
                <ShieldCheck size={22} />
              </div>
              <div>
                <div className="stat-val" style={{ fontSize: '1.1rem', color: '#38bdf8' }}>ML-DSA-44</div>
                <div className="stat-lbl">FIPS 204 Active</div>
              </div>
            </div>
          </div>
        )}

        {/* Filter and Search header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.25rem', flexWrap: 'wrap', gap: '1rem' }}>
          <div className="search-box">
            <Search size={16} />
            <input
              type="text"
              placeholder="Search by app name or Client ID…"
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <div style={{ fontSize: '0.85rem', color: 'var(--text-sub)' }}>
            Showing {filteredApps.length} of {appList.length} applications
          </div>
        </div>

        {loading ? (
          <div className="card">
            <div className="skeleton skeleton-row" style={{ width: '40%', height: '1.5rem' }} />
            <div className="skeleton skeleton-row" style={{ width: '85%' }} />
            <div className="skeleton skeleton-row" style={{ width: '60%' }} />
            <div className="skeleton skeleton-row" style={{ width: '75%' }} />
          </div>
        ) : filteredApps.length === 0 ? (
          <div className="card empty-state">
            <div className="empty-icon"><Boxes size={32} /></div>
            <h3>{search ? 'No matching applications found' : 'No applications registered yet'}</h3>
            <p>{search ? 'Try clearing your search query filter.' : 'Register your first OAuth 2.0 app to obtain client credentials.'}</p>
            {!search && (
              <Link to="/apps/new" className="btn btn-primary">
                <Plus size={16} /> Create Application
              </Link>
            )}
          </div>
        ) : (
          <div className="card" style={{ padding: 0, overflowX: 'auto', maxWidth: '100%' }}>
            <table>
              <thead>
                <tr>
                  <th>Application Name</th>
                  <th>Client ID</th>
                  <th>Client Type</th>
                  <th>Created Date</th>
                  <th style={{ textAlign: 'right' }}>Action</th>
                </tr>
              </thead>
              <tbody>
                {filteredApps.map(app => (
                  <tr key={app.ID}>
                    <td>
                      <Link to={`/apps/${app.ID}`} style={{ color: 'var(--text-main)', fontWeight: 700 }}>
                        {app.Name}
                      </Link>
                      {app.Description && (
                        <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginTop: '0.15rem' }}>
                          {app.Description}
                        </div>
                      )}
                    </td>
                    <td>
                      <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
                        <span className="mono">{app.ClientID}</span>
                        <button
                          className="btn btn-outline btn-sm"
                          style={{ padding: '0.15rem 0.4rem' }}
                          onClick={e => copyClientId(app.ClientID, e)}
                          title="Copy Client ID"
                        >
                          {copiedId === app.ClientID ? <Check size={12} color="var(--emerald)" /> : <Copy size={12} />}
                        </button>
                      </div>
                    </td>
                    <td>
                      <span className={`badge ${app.IsPublic ? 'badge-public' : 'badge-confidential'}`}>
                        {app.IsPublic ? <Globe size={11} /> : <Key size={11} />}
                        {app.IsPublic ? 'Public (PKCE)' : 'Confidential'}
                      </span>
                    </td>
                    <td style={{ color: 'var(--text-sub)' }}>
                      {new Date(app.CreatedAt).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <Link to={`/apps/${app.ID}`} className="btn btn-outline btn-sm">
                        <Eye size={14} /> Details
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  )
}
