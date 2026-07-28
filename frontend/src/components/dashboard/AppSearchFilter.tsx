import React from 'react'
import { Search, Globe, Lock, Layers } from 'lucide-react'

export type ClientTypeFilter = 'all' | 'public' | 'confidential'

export interface AppSearchFilterProps {
  search: string
  onSearchChange: (search: string) => void
  typeFilter: ClientTypeFilter
  onTypeFilterChange: (type: ClientTypeFilter) => void
  showingCount: number
  totalCount: number
  publicCount?: number
  confidentialCount?: number
}

export const AppSearchFilter: React.FC<AppSearchFilterProps> = ({
  search,
  onSearchChange,
  typeFilter,
  onTypeFilterChange,
  showingCount,
  totalCount,
  publicCount,
  confidentialCount,
}) => {
  const filterTabs: Array<{ id: ClientTypeFilter; label: string; count?: number; icon: React.ElementType }> = [
    { id: 'all', label: 'All', count: totalCount, icon: Layers },
    { id: 'public', label: 'Public SPA', count: publicCount, icon: Globe },
    { id: 'confidential', label: 'Confidential Server', count: confidentialCount, icon: Lock },
  ]

  return (
    <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 mb-6">
      {/* Search Input & Filter Tabs Container */}
      <div className="flex flex-col sm:flex-row sm:items-center gap-3 flex-1">
        {/* Search Bar */}
        <div className="relative flex-1 min-w-[240px] max-w-md">
          <Search
            size={18}
            className="absolute left-3.5 top-1/2 -translate-y-1/2 text-[var(--color-muted)] pointer-events-none"
          />
          <input
            type="text"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search applications by name or Client ID..."
            className="w-full pl-10 pr-4 py-2 text-sm bg-[var(--color-input)] border border-[var(--color-border)] rounded-xl
              text-[var(--color-txt)] placeholder:text-[var(--color-muted)]
              focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20
              hover:border-[var(--color-border-mid)] transition-all duration-150 shadow-inner"
          />
        </div>

        {/* Type Filter Buttons / Segmented Tabs */}
        <div className="inline-flex items-center p-1 bg-[#0c1120] border border-[var(--color-border)] rounded-xl shrink-0">
          {filterTabs.map((tab) => {
            const Icon = tab.icon
            const isActive = typeFilter === tab.id
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => onTypeFilterChange(tab.id)}
                className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all duration-150 cursor-pointer ${
                  isActive
                    ? 'bg-blue-600/20 text-blue-400 border border-blue-500/30 shadow-sm'
                    : 'text-[var(--color-sub)] hover:text-white hover:bg-white/5 border border-transparent'
                }`}
              >
                <Icon size={14} className={isActive ? 'text-blue-400' : 'text-[var(--color-muted)]'} />
                <span>{tab.label}</span>
                {tab.count !== undefined && (
                  <span
                    className={`ml-0.5 px-1.5 py-0.25 rounded-md text-[0.7rem] font-medium ${
                      isActive ? 'bg-blue-500/20 text-blue-300' : 'bg-white/5 text-[var(--color-muted)]'
                    }`}
                  >
                    {tab.count}
                  </span>
                )}
              </button>
            )
          })}
        </div>
      </div>

      {/* Matching Applications Count Badge */}
      <div className="self-start lg:self-center shrink-0">
        <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-[var(--color-card)] border border-[var(--color-border)] text-[var(--color-sub)] shadow-sm">
          Showing <span className="font-semibold text-white">{showingCount}</span> of{' '}
          <span className="font-semibold text-white">{totalCount}</span> applications
        </span>
      </div>
    </div>
  )
}

export default AppSearchFilter
