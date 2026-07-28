import React from 'react'
import { Link } from 'react-router-dom'
import { Globe, Lock, Copy, Check, ChevronRight, Calendar } from 'lucide-react'
import type { OAuthApp } from '../../api'

export interface AppCardProps {
  app: OAuthApp
  copiedId: string | null
  onCopyClientId: (id: string, e: React.MouseEvent) => void
}

export const formatClientId = (id: string): string => {
  if (!id) return ''
  if (id.length > 16) {
    return `${id.slice(0, 11)}...`
  }
  return id
}

export const AppCard: React.FC<AppCardProps> = ({ app, copiedId, onCopyClientId }) => {
  const isCopied = copiedId === app.ClientID
  const isPublic = app.IsPublic

  const formattedDate = app.CreatedAt
    ? new Date(app.CreatedAt).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      })
    : 'Recently'

  const initial = app.Name ? app.Name.charAt(0).toUpperCase() : 'A'

  return (
    <div className="bg-[var(--color-card)] border border-[var(--color-border)] hover:border-[var(--color-border-mid)] rounded-2xl p-5 transition-all duration-200 shadow-md hover:shadow-lg hover:shadow-blue-500/5 group flex flex-col justify-between gap-4">
      {/* Header: App Avatar, Name, Description, Type Badge */}
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-start gap-3.5 min-w-0">
          {/* App Icon Badge / Avatar */}
          <div
            className={`w-11 h-11 rounded-xl flex items-center justify-center font-bold text-base shrink-0 shadow-inner border transition-transform duration-200 group-hover:scale-105 ${
              isPublic
                ? 'bg-cyan-500/10 border-cyan-500/20 text-cyan-400'
                : 'bg-purple-500/10 border-purple-500/20 text-purple-400'
            }`}
          >
            {initial}
          </div>

          <div className="min-w-0 flex-1">
            <Link
              to={`/apps/${app.ID}`}
              className="text-[var(--color-txt)] font-bold text-base hover:text-blue-400 transition-colors line-clamp-1 block"
            >
              {app.Name}
            </Link>
            <p className="text-[var(--color-sub)] text-xs mt-0.5 line-clamp-2 leading-relaxed">
              {app.Description || 'No description provided.'}
            </p>
          </div>
        </div>

        {/* Type Badge */}
        <div className="shrink-0">
          {isPublic ? (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-cyan-500/10 border border-cyan-500/30 text-sky-400 shadow-sm">
              <Globe size={12} className="shrink-0" />
              <span>Public SPA</span>
            </span>
          ) : (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-purple-500/10 border border-purple-500/30 text-purple-300 shadow-sm">
              <Lock size={12} className="shrink-0" />
              <span>Confidential Server</span>
            </span>
          )}
        </div>
      </div>

      {/* Footer Info: Masked Client ID & Quick Copy, Created Date, Action Link */}
      <div className="pt-3.5 border-t border-[var(--color-border)]/60 flex flex-wrap items-center justify-between gap-3 text-xs">
        {/* Masked Client ID with 1-click Copy */}
        <div className="flex items-center gap-2">
          <span className="text-[var(--color-muted)] font-medium">Client ID:</span>
          <div className="inline-flex items-center gap-1 bg-[#050810] border border-[var(--color-border)] rounded-lg px-2 py-1">
            <code className="font-mono text-xs text-blue-400 tracking-tight" title={app.ClientID}>
              {formatClientId(app.ClientID)}
            </code>
            <button
              type="button"
              onClick={(e) => onCopyClientId(app.ClientID, e)}
              title={isCopied ? 'Copied!' : 'Copy Client ID'}
              className={`p-1 rounded transition-all duration-150 cursor-pointer ${
                isCopied
                  ? 'text-emerald-400 bg-emerald-500/10'
                  : 'text-[var(--color-muted)] hover:text-white hover:bg-white/10'
              }`}
            >
              {isCopied ? <Check size={13} className="text-emerald-400" /> : <Copy size={13} />}
            </button>
          </div>
        </div>

        <div className="flex items-center gap-4 ml-auto">
          {/* Created Date */}
          <div className="inline-flex items-center gap-1.5 text-[var(--color-muted)]">
            <Calendar size={13} />
            <span>{formattedDate}</span>
          </div>

          {/* Details Navigation Link */}
          <Link
            to={`/apps/${app.ID}`}
            className="inline-flex items-center gap-1 text-xs font-semibold text-blue-400 hover:text-blue-300 transition-colors group/link"
          >
            <span>Settings</span>
            <ChevronRight size={14} className="transition-transform duration-150 group-hover/link:translate-x-0.5" />
          </Link>
        </div>
      </div>
    </div>
  )
}

export default AppCard
