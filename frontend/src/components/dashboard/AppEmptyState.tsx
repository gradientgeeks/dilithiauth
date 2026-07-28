import React from 'react'
import { Link } from 'react-router-dom'
import { Plus, Boxes, SearchX, RotateCcw } from 'lucide-react'
import type { ClientTypeFilter } from './AppSearchFilter'

export interface AppEmptyStateProps {
  search?: string
  typeFilter?: ClientTypeFilter
  onClearFilters?: () => void
}

export const AppEmptyState: React.FC<AppEmptyStateProps> = ({
  search,
  typeFilter,
  onClearFilters,
}) => {
  const isFiltered = Boolean(search || (typeFilter && typeFilter !== 'all'))

  return (
    <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-2xl p-12 text-center shadow-md max-w-2xl mx-auto my-6">
      {/* Icon Illustration Container */}
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-blue-500/10 border border-blue-500/20 text-blue-400 mb-5 shadow-inner">
        {isFiltered ? <SearchX size={32} /> : <Boxes size={32} />}
      </div>

      {/* Heading & Subtitle */}
      <h3 className="text-lg font-bold text-[var(--color-txt)] mb-2 tracking-tight">
        {isFiltered ? 'No matching applications found' : 'No applications registered yet'}
      </h3>
      <p className="text-sm text-[var(--color-sub)] max-w-md mx-auto mb-6 leading-relaxed">
        {isFiltered
          ? 'No applications match your current search query or filter selection. Try adjusting your search term or clearing filters.'
          : 'Register your first post-quantum OAuth 2.0 / OIDC application to obtain client credentials and ML-DSA-44 token protection.'}
      </p>

      {/* CTA Buttons */}
      <div className="flex flex-wrap items-center justify-center gap-3">
        {isFiltered && onClearFilters && (
          <button
            type="button"
            onClick={onClearFilters}
            className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold
              bg-[#1a2236] border border-[var(--color-border-mid)] text-[var(--color-txt)]
              hover:bg-[#242f4a] hover:border-blue-500/40 transition-all duration-150 cursor-pointer shadow-sm"
          >
            <RotateCcw size={16} /> Clear Filters
          </button>
        )}

        <Link
          to="/apps/new"
          className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold
            bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20
            hover:shadow-blue-500/30 active:translate-y-px transition-all duration-150 no-underline hover:no-underline"
        >
          <Plus size={18} /> Create Application
        </Link>
      </div>
    </div>
  )
}

export default AppEmptyState
