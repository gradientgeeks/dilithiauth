import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'

export default function Dashboard() {
  const [appList, setAppList] = useState<OAuthApp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    apps.list()
      .then(r => setAppList(r.data ?? []))
      .catch(() => setError('Failed to load apps.'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <>
      <Navbar />
      <div className="container">
        <div className="page-header">
          <div>
            <h1>My Applications</h1>
            <p>Manage your OAuth 2.0 apps and client credentials.</p>
          </div>
          <Link to="/apps/new" className="btn btn-primary">+ New Application</Link>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        {loading ? (
          <div className="card"><p className="text-muted">Loading…</p></div>
        ) : appList.length === 0 ? (
          <div className="card empty-state">
            <h3>No applications yet</h3>
            <p>Create your first OAuth app to get a client ID and secret.</p>
            <Link to="/apps/new" className="btn btn-primary">Create application</Link>
          </div>
        ) : (
          <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Client ID</th>
                  <th>Type</th>
                  <th>Created</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {appList.map(app => (
                  <tr key={app.ID}>
                    <td><strong>{app.Name}</strong></td>
                    <td><span className="mono">{app.ClientID}</span></td>
                    <td>
                      <span className={`badge ${app.IsPublic ? 'badge-public' : 'badge-confidential'}`}>
                        {app.IsPublic ? 'Public' : 'Confidential'}
                      </span>
                    </td>
                    <td className="text-muted">{new Date(app.CreatedAt).toLocaleDateString()}</td>
                    <td><Link to={`/apps/${app.ID}`} className="btn btn-outline btn-sm">View</Link></td>
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
