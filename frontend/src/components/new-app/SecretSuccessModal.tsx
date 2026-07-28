import { useState } from 'react'
import { CheckCircle2, AlertTriangle, ShieldCheck, Copy, Check, ArrowRight, LayoutDashboard } from 'lucide-react'
import type { OAuthApp } from '../../api'

export interface SecretSuccessModalProps {
  app: OAuthApp
  secret?: string
  onClose?: () => void
  onNavigateToApp: (appId: string) => void
  onNavigateToDashboard: () => void
}

const labelCls = "block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5"

export default function SecretSuccessModal({
  app,
  secret,
  onNavigateToApp,
  onNavigateToDashboard,
}: SecretSuccessModalProps) {
  const [copiedId, setCopiedId] = useState(false)
  const [copiedSecret, setCopiedSecret] = useState(false)

  const copyClientId = () => {
    if (app.ClientID) {
      navigator.clipboard.writeText(app.ClientID)
      setCopiedId(true)
      setTimeout(() => setCopiedId(false), 2000)
    }
  }

  const copySecret = () => {
    if (secret) {
      navigator.clipboard.writeText(secret)
      setCopiedSecret(true)
      setTimeout(() => setCopiedSecret(false), 2000)
    }
  }

  return (
    <div className="fixed inset-0 z-[500] bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 max-w-lg w-full shadow-2xl animate-fadeIn relative">
        {/* Header */}
        <div className="text-center mb-6">
          <div className="inline-flex bg-emerald-500/15 border border-emerald-400/30 p-3 rounded-2xl text-emerald-400 mb-3 shadow-[0_0_15px_rgba(16,185,129,0.2)]">
            <CheckCircle2 size={32} />
          </div>
          <h2 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight">
            Application Created Successfully!
          </h2>
          <p className="text-[var(--color-muted)] text-[0.9rem] mt-1">
            <strong className="text-[var(--color-txt)]">{app.Name}</strong> &bull;{' '}
            {app.IsPublic ? 'Public Client (PKCE)' : 'Confidential Client'}
          </p>
        </div>

        {/* Client ID Box */}
        <div className="mb-4">
          <label className={labelCls}>Client ID</label>
          <div className="flex items-center gap-2">
            <div className="flex-1 bg-[#050810] border border-[var(--color-border)] rounded-md px-3.5 py-2.5 font-mono text-xs sm:text-sm text-sky-400 select-all overflow-x-auto">
              {app.ClientID}
            </div>
            <button
              type="button"
              onClick={copyClientId}
              className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-semibold rounded-md
                border border-[var(--color-border)] text-[var(--color-sub)] bg-[var(--color-input)]
                hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer shrink-0"
            >
              {copiedId ? (
                <>
                  <Check size={14} className="text-emerald-400" />
                  <span className="text-emerald-400">Copied!</span>
                </>
              ) : (
                <>
                  <Copy size={14} />
                  <span>Copy</span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Client Secret Box (Confidential apps) */}
        {secret && (
          <div className="mb-5">
            <div className="flex items-start gap-2.5 p-3.5 rounded-lg mb-3 bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs sm:text-sm leading-relaxed">
              <AlertTriangle size={18} className="shrink-0 mt-0.5 text-amber-400" />
              <div>
                Save your Client Secret now. <strong>It will not be shown again!</strong>
              </div>
            </div>
            <label className={labelCls}>Client Secret</label>
            <div className="secret-box mb-2 text-xs sm:text-sm select-all">
              {secret}
            </div>
            <button
              type="button"
              onClick={copySecret}
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold rounded-md
                border border-amber-500/40 text-amber-300 bg-amber-500/10
                hover:bg-amber-500/20 transition-all duration-150 cursor-pointer"
            >
              {copiedSecret ? (
                <>
                  <Check size={14} className="text-emerald-400" />
                  <span className="text-emerald-400">Copied to Clipboard!</span>
                </>
              ) : (
                <>
                  <Copy size={14} />
                  <span>Copy Secret</span>
                </>
              )}
            </button>
          </div>
        )}

        {/* Public Client Banner */}
        {app.IsPublic && (
          <div className="flex items-start gap-2.5 p-3.5 rounded-lg mb-5 bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-xs sm:text-sm leading-relaxed">
            <ShieldCheck size={18} className="shrink-0 mt-0.5 text-cyan-400" />
            <div>
              Registered as a <strong>Public Client (PKCE mode)</strong>. No client secret is required.
            </div>
          </div>
        )}

        <div className="h-px bg-[var(--color-border)] my-5" />

        {/* Navigation CTAs */}
        <div className="flex flex-col sm:flex-row gap-3">
          <button
            type="button"
            onClick={() => onNavigateToApp(app.ID)}
            className="flex-1 flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
              bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
              hover:shadow-[0_0_0_3px_rgba(59,130,246,0.25)] transition-all duration-150 cursor-pointer"
          >
            <span>Continue to App Details</span>
            <ArrowRight size={16} />
          </button>
          <button
            type="button"
            onClick={onNavigateToDashboard}
            className="inline-flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
              bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
              hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)] transition-all duration-150 cursor-pointer"
          >
            <LayoutDashboard size={15} />
            <span>Dashboard</span>
          </button>
        </div>
      </div>
    </div>
  )
}
