import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Plus, Boxes, AlertCircle, ShieldCheck, Globe, Lock } from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import AppSearchFilter, { type ClientTypeFilter } from '../components/dashboard/AppSearchFilter'
import AppCard from '../components/dashboard/AppCard'
import AppEmptyState from '../components/dashboard/AppEmptyState'

export default function Dashboard() {
  const [appList, setAppList] = useState<OAuthApp[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState<ClientTypeFilter>('all')
  const [copiedId, setCopiedId] = useState<string | null>(null)

  useEffect(() => {
    apps
      .list()
      .then((r) => setAppList(r.data ?? []))
      .catch(() => setError('Failed to load applications.'))
      .finally(() => setLoading(false))
  }, [])

  const copyClientId = (id: string, e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(id)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const publicCount = appList.filter((a) => a.IsPublic).length
  const confidentialCount = appList.filter((a) => !a.IsPublic).length

  const filteredApps = appList.filter((app) => {
    const matchesSearch =
      app.Name.toLowerCase().includes(search.toLowerCase()) ||
      app.ClientID.toLowerCase().includes(search.toLowerCase())

    const matchesType =
      typeFilter === 'all'
        ? true
        : typeFilter === 'public'
        ? app.IsPublic
        : !app.IsPublic

    return matchesSearch && matchesType
  })

  const clearFilters = () => {
    setSearch('')
    setTypeFilter('all')
  }

  return (
    <div className="w-full max-w-6xl mx-auto px-4 sm:px-6 py-8">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-[var(--color-txt)] tracking-tight">
            Applications
          </h1>
          <p className="text-[var(--color-sub)] text-sm mt-1">
            Manage your post-quantum OAuth 2.0 / OIDC applications and client credentials.
          </p>
        </div>
        <Link
          to="/apps/new"
          className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold
            bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20
            hover:shadow-blue-500/30 active:translate-y-px transition-all duration-150 no-underline hover:no-underline shrink-0"
        >
          <Plus size={18} /> Create Application
        </Link>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="flex items-start gap-3 p-4 rounded-xl mb-6 bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm shadow-sm">
          <AlertCircle size={18} className="shrink-0 mt-0.5" />
          <div className="font-medium">{error}</div>
        </div>
      )}

      {/* Stats Summary Cards (4 Metric Grid) */}
      {!loading && (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          {[
            {
              label: 'Total Applications',
              val: appList.length,
              icon: Boxes,
              iconCls: 'bg-indigo-500/10 border-indigo-500/20 text-indigo-400',
            },
            {
              label: 'Public SPAs (PKCE)',
              val: publicCount,
              icon: Globe,
              iconCls: 'bg-cyan-500/10 border-cyan-500/20 text-cyan-400',
            },
            {
              label: 'Confidential Servers',
              val: confidentialCount,
              icon: Lock,
              iconCls: 'bg-purple-500/10 border-purple-500/20 text-purple-400',
            },
            {
              label: 'PQC Scheme Active',
              val: 'ML-DSA-44',
              icon: ShieldCheck,
              iconCls: 'bg-blue-500/10 border-blue-500/20 text-blue-400',
              valCls: 'text-base font-bold text-blue-400 tracking-wide',
            },
          ].map(({ label, val, icon: Icon, iconCls, valCls }) => (
            <div
              key={label}
              className="bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-border-mid)] rounded-2xl p-4.5 flex items-center gap-3.5 transition-all duration-150 shadow-sm"
            >
              <div
                className={`w-11 h-11 rounded-xl border flex items-center justify-center shrink-0 shadow-inner ${iconCls}`}
              >
                <Icon size={20} />
              </div>
              <div className="min-w-0">
                <div
                  className={
                    valCls ?? 'text-2xl font-bold text-[var(--color-txt)] leading-none tracking-tight'
                  }
                >
                  {val}
                </div>
                <div className="text-xs text-[var(--color-sub)] font-medium mt-1 truncate">
                  {label}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Search & Type Filter Bar */}
      <AppSearchFilter
        search={search}
        onSearchChange={setSearch}
        typeFilter={typeFilter}
        onTypeFilterChange={setTypeFilter}
        showingCount={filteredApps.length}
        totalCount={appList.length}
        publicCount={publicCount}
        confidentialCount={confidentialCount}
      />

      {/* Loading Skeleton */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div
              key={i}
              className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-2xl p-5 space-y-4 shadow-sm"
            >
              <div className="flex items-center gap-3">
                <div className="skeleton w-11 h-11 rounded-xl" />
                <div className="space-y-2 flex-1">
                  <div className="skeleton h-4 w-1/3 rounded" />
                  <div className="skeleton h-3 w-2/3 rounded" />
                </div>
              </div>
              <div className="pt-3 border-t border-[var(--color-border)]/50 flex justify-between items-center">
                <div className="skeleton h-4 w-28 rounded" />
                <div className="skeleton h-4 w-20 rounded" />
              </div>
            </div>
          ))}
        </div>
      ) : filteredApps.length === 0 ? (
        /* Empty State */
        <AppEmptyState
          search={search}
          typeFilter={typeFilter}
          onClearFilters={clearFilters}
        />
      ) : (
        /* Grid Rendering of App Cards */
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {filteredApps.map((app) => (
            <AppCard
              key={app.ID}
              app={app}
              copiedId={copiedId}
              onCopyClientId={copyClientId}
            />
          ))}
        </div>
      )}
    </div>
  )
}
