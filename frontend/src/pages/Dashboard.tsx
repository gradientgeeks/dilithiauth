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
      <div className="w-full max-w-[1020px] mx-auto px-4 sm:px-6 py-8">

        {/* Page header */}
        <div className="flex flex-wrap items-start justify-between gap-4 mb-6">
          <div>
            <h1 className="text-[1.6rem] font-bold text-[var(--color-txt)] tracking-tight">Developer Dashboard</h1>
            <p className="text-[var(--color-sub)] text-sm mt-1">Manage your OAuth 2.0 applications and client credentials.</p>
          </div>
          <Link
            to="/apps/new"
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold
              bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
              hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
              transition-all duration-150 no-underline hover:no-underline"
          >
            <Plus size={18} /> Register New Application
          </Link>
        </div>

        {error && (
          <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
            <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
          </div>
        )}

        {/* Stats */}
        {!loading && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5">
            {[
              { icon: Boxes, val: appList.length, lbl: 'Registered Apps', iconCls: 'bg-[var(--color-hover)] text-[var(--color-sub)]' },
              { icon: Globe, val: publicCount, lbl: 'Public PKCE Clients', iconCls: 'bg-cyan-500/15 text-sky-400' },
              { icon: Lock, val: confidentialCount, lbl: 'Confidential Clients', iconCls: 'bg-emerald-500/15 text-emerald-400' },
              { icon: ShieldCheck, val: 'ML-DSA-44', lbl: 'FIPS 204 Active', iconCls: 'bg-blue-500/15 text-sky-400', valCls: 'text-[1.05rem] text-sky-400' },
            ].map(({ icon: Icon, val, lbl, iconCls, valCls }) => (
              <div
                key={lbl}
                className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-4 flex items-center gap-3
                  hover:border-[var(--color-border-mid)] transition-colors duration-150"
              >
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 ${iconCls}`}>
                  <Icon size={20} />
                </div>
                <div>
                  <div className={`font-bold leading-tight ${valCls ?? 'text-[1.35rem] text-[var(--color-txt)]'}`}>{val}</div>
                  <div className="text-[0.73rem] text-[var(--color-muted)] mt-0.5">{lbl}</div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Search + count */}
        <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
          <div className="relative flex-1 max-w-xs">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)] pointer-events-none" />
            <input
              type="text"
              placeholder="Search by name or Client ID…"
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="w-full pl-9 pr-3 py-2 text-sm bg-[var(--color-input)] border border-[var(--color-border)] rounded-md
                text-[var(--color-txt)] placeholder:text-[var(--color-muted)]
                focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)]
                hover:border-[var(--color-border-mid)] transition-all duration-150"
            />
          </div>
          <span className="text-[0.85rem] text-[var(--color-sub)]">
            Showing {filteredApps.length} of {appList.length} applications
          </span>
        </div>

        {/* Table / Empty / Skeleton */}
        {loading ? (
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 space-y-3">
            {[40, 85, 60, 75].map((w, i) => (
              <div key={i} className="skeleton h-4 rounded" style={{ width: `${w}%` }} />
            ))}
          </div>
        ) : filteredApps.length === 0 ? (
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-12 text-center">
            <div className="inline-flex items-center justify-center w-14 h-14 rounded-xl bg-[var(--color-hover)] text-[var(--color-muted)] mb-4">
              <Boxes size={32} />
            </div>
            <h3 className="text-[var(--color-txt)] font-semibold mb-2">
              {search ? 'No matching applications found' : 'No applications registered yet'}
            </h3>
            <p className="text-[var(--color-sub)] text-sm mb-4">
              {search ? 'Try clearing your search query.' : 'Register your first OAuth 2.0 app to obtain client credentials.'}
            </p>
            {!search && (
              <Link
                to="/apps/new"
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  transition-all duration-150 no-underline hover:no-underline"
              >
                <Plus size={16} /> Create Application
              </Link>
            )}
          </div>
        ) : (
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    {['Application Name', 'Client ID', 'Client Type', 'Created Date', ''].map(h => (
                      <th
                        key={h}
                        className="text-left text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider
                          px-4 py-3 border-b border-[var(--color-border)] bg-[#0a0e1a] whitespace-nowrap last:text-right"
                      >
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {filteredApps.map(app => (
                    <tr key={app.ID} className="hover:bg-white/[0.02] transition-colors">
                      <td className="px-4 py-3.5 border-b border-[var(--color-border)]">
                        <Link
                          to={`/apps/${app.ID}`}
                          className="text-[var(--color-txt)] font-bold hover:text-[var(--color-primary)] transition-colors"
                        >
                          {app.Name}
                        </Link>
                        {app.Description && (
                          <div className="text-[0.78rem] text-[var(--color-muted)] mt-0.5">{app.Description}</div>
                        )}
                      </td>
                      <td className="px-4 py-3.5 border-b border-[var(--color-border)]">
                        <div className="inline-flex items-center gap-1.5">
                          <span className="mono-inline text-[0.8rem]">{app.ClientID}</span>
                          <button
                            onClick={e => copyClientId(app.ClientID, e)}
                            title="Copy Client ID"
                            className="p-1 rounded border border-[var(--color-border)] text-[var(--color-muted)]
                              hover:bg-[var(--color-hover)] hover:text-[var(--color-sub)] transition-all duration-150 cursor-pointer"
                          >
                            {copiedId === app.ClientID ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
                          </button>
                        </div>
                      </td>
                      <td className="px-4 py-3.5 border-b border-[var(--color-border)]">
                        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[0.68rem] font-bold uppercase tracking-wide
                          ${app.IsPublic ? 'bg-[#1e293b] border border-[#334155] text-sky-400' : 'bg-[#052e16] border border-[#166534] text-emerald-400'}`}>
                          {app.IsPublic ? <Globe size={11} /> : <Key size={11} />}
                          {app.IsPublic ? 'Public (PKCE)' : 'Confidential'}
                        </span>
                      </td>
                      <td className="px-4 py-3.5 border-b border-[var(--color-border)] text-[var(--color-sub)] text-sm whitespace-nowrap">
                        {new Date(app.CreatedAt).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
                      </td>
                      <td className="px-4 py-3.5 border-b border-[var(--color-border)] text-right">
                        <Link
                          to={`/apps/${app.ID}`}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md
                            border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
                            hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] hover:border-[var(--color-border-mid)]
                            transition-all duration-150 no-underline hover:no-underline"
                        >
                          <Eye size={14} /> Details
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </>
  )
}
