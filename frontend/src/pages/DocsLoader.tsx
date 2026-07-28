import React, { Component, Suspense } from 'react'
import { Shield, RefreshCw, AlertTriangle } from 'lucide-react'
import Navbar from '../components/Navbar'

// ---------------------------------------------------------------------------
// Lazy remote import — loaded at runtime from the docs micro frontend app.
// The actual URL is configured via the vite federation plugin.
// ---------------------------------------------------------------------------
const RemoteDocsApp = React.lazy(() => import('docs/DocsApp'))

// ---------------------------------------------------------------------------
// Branded skeleton shown while the remote JS chunk is downloading
// ---------------------------------------------------------------------------
function DocsSkeleton() {
  return (
    <div className="min-h-screen bg-[var(--color-bg)] text-[var(--color-txt)]">
      <Navbar />
      <div className="w-full max-w-[1280px] mx-auto px-4 sm:px-6 py-8 flex gap-8">
        {/* Sidebar skeleton */}
        <div className="hidden md:flex flex-col gap-3 w-56 shrink-0 pt-2">
          {[80, 65, 75, 60].map((w, i) => (
            <div
              key={i}
              className="h-5 rounded bg-[var(--color-card)] border border-[var(--color-border)] relative overflow-hidden"
              style={{ width: `${w}%` }}
            >
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
          ))}
        </div>

        {/* Main content skeleton */}
        <div className="flex-1 min-w-0 space-y-6">
          {/* Header */}
          <div className="space-y-3 pb-4 border-b border-[var(--color-border)]">
            <div className="h-6 w-48 rounded-full bg-[var(--color-card)] border border-[var(--color-border)] relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
            <div className="h-8 w-3/4 rounded bg-[var(--color-card)] border border-[var(--color-border)] relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
            <div className="h-4 w-full max-w-xl rounded bg-[var(--color-card)] border border-[var(--color-border)] relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
          </div>

          {/* Quick-jump bar */}
          <div className="flex gap-2">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="h-9 w-36 rounded-lg bg-[var(--color-card)] border border-[var(--color-border)] relative overflow-hidden">
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
              </div>
            ))}
          </div>

          {/* Card skeleton */}
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 space-y-4">
            <div className="h-6 w-64 rounded bg-[var(--color-hover)] relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
            <div className="grid grid-cols-5 gap-2">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="h-8 rounded-md bg-[var(--color-hover)] relative overflow-hidden">
                  <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
                </div>
              ))}
            </div>
            <div className="h-32 rounded-lg bg-[var(--color-hover)] relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/[0.04] to-transparent animate-[shimmer_1.4s_infinite]" />
            </div>
          </div>

          {/* Loading badge */}
          <div className="flex items-center justify-center gap-3 py-4 text-sm text-[var(--color-sub)]">
            <Shield size={18} className="text-blue-400 animate-pulse" />
            <span>Loading DilithiAuth Docs micro frontend…</span>
            <RefreshCw size={14} className="animate-spin text-[var(--color-muted)]" />
          </div>
        </div>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Error boundary — shown when the remote JS fails to load
// (remote app down, network error, version mismatch, etc.)
// ---------------------------------------------------------------------------
interface EBState { hasError: boolean; error: Error | null }

class DocsErrorBoundary extends Component<{ children: React.ReactNode }, EBState> {
  constructor(props: { children: React.ReactNode }) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): EBState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[DocsLoader] Remote module failed to load:', error, info)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-[var(--color-bg)] text-[var(--color-txt)]">
          <Navbar />
          <div className="flex flex-col items-center justify-center min-h-[60vh] px-4 text-center gap-5">
            <div className="p-4 rounded-xl bg-amber-500/15 border border-amber-500/30 text-amber-400">
              <AlertTriangle size={32} />
            </div>
            <div className="space-y-2 max-w-md">
              <h1 className="text-xl font-bold text-[var(--color-txt)]">Docs unavailable</h1>
              <p className="text-sm text-[var(--color-sub)] leading-relaxed">
                The documentation micro frontend could not be loaded. This usually means the docs
                remote app is not running or unreachable.
              </p>
              {this.state.error && (
                <p className="text-xs font-mono text-red-400 bg-red-500/10 border border-red-500/20 rounded p-2 mt-2">
                  {this.state.error.message}
                </p>
              )}
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => this.setState({ hasError: false, error: null })}
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-[var(--color-primary)] text-white text-sm font-semibold hover:bg-[var(--color-primary-h)] transition-colors cursor-pointer"
              >
                <RefreshCw size={14} /> Retry
              </button>
              <a
                href="https://github.com/Uttam-Mahata/DilithiAuth"
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-transparent border border-[var(--color-border)] text-[var(--color-sub)] text-sm font-semibold hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors"
              >
                GitHub Repo
              </a>
            </div>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}

// ---------------------------------------------------------------------------
// Public export — drop this in App.tsx instead of <Docs />
// ---------------------------------------------------------------------------
export default function DocsLoader() {
  return (
    <DocsErrorBoundary>
      <Suspense fallback={<DocsSkeleton />}>
        {/* standalone=false — the host Navbar is already rendered by Navbar import above
            inside the skeleton. The remote DocsApp suppresses its own nav when standalone=false. */}
        <RemoteDocsApp standalone={false} />
      </Suspense>
    </DocsErrorBoundary>
  )
}
