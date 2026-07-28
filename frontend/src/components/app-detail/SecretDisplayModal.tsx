import { useState } from 'react'
import { AlertTriangle, Copy, Check, KeyRound, ShieldCheck } from 'lucide-react'

export interface SecretDisplayModalProps {
  clientId?: string
  clientSecret: string
  onClose: () => void
}

export default function SecretDisplayModal({ clientId, clientSecret, onClose }: SecretDisplayModalProps) {
  const [copiedSecret, setCopiedSecret] = useState(false)
  const [copiedId, setCopiedId] = useState(false)

  const handleCopySecret = () => {
    navigator.clipboard.writeText(clientSecret)
    setCopiedSecret(true)
    setTimeout(() => setCopiedSecret(false), 2000)
  }

  const handleCopyId = () => {
    if (!clientId) return
    navigator.clipboard.writeText(clientId)
    setCopiedId(true)
    setTimeout(() => setCopiedId(false), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fadeIn">
      <div className="w-full max-w-lg bg-[var(--color-card)] border border-[var(--color-border)] rounded-2xl p-6 shadow-2xl space-y-5 relative">
        {/* Header */}
        <div className="flex items-center gap-3 border-b border-[var(--color-border)] pb-4">
          <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
            <KeyRound size={20} />
          </div>
          <div>
            <h2 className="text-lg font-bold text-[var(--color-txt)]">New Client Secret Generated</h2>
            <p className="text-xs text-[var(--color-sub)]">Your client secret has been rotated successfully.</p>
          </div>
        </div>

        {/* Amber Warning Banner */}
        <div className="flex items-start gap-3 p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-300 text-xs leading-relaxed">
          <AlertTriangle size={18} className="shrink-0 mt-0.5 text-amber-400" />
          <div>
            <span className="font-bold text-amber-200 uppercase tracking-wider block mb-0.5">Crucial Warning</span>
            Save this secret now! It will <strong>never be displayed again</strong>. Make sure to update your application settings with this new secret immediately.
          </div>
        </div>

        {/* Client ID (if provided) */}
        {clientId && (
          <div>
            <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5">
              Client ID
            </label>
            <div className="flex gap-2 items-center">
              <span className="mono-inline flex-1 px-3.5 py-2 text-xs bg-[#050810] border border-[var(--color-border)] rounded-lg font-mono text-[var(--color-txt)] select-all truncate">
                {clientId}
              </span>
              <button
                type="button"
                onClick={handleCopyId}
                className="inline-flex items-center gap-1 px-3 py-2 rounded-lg text-xs font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all cursor-pointer shrink-0"
              >
                {copiedId ? (
                  <>
                    <Check size={14} className="text-emerald-400" /> Copied
                  </>
                ) : (
                  <>
                    <Copy size={14} /> Copy ID
                  </>
                )}
              </button>
            </div>
          </div>
        )}

        {/* Client Secret Display */}
        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider">
              Newly Generated Client Secret
            </label>
            <button
              type="button"
              onClick={handleCopySecret}
              className="inline-flex items-center gap-1 text-xs text-[var(--color-primary)] hover:text-blue-400 font-semibold cursor-pointer"
            >
              {copiedSecret ? (
                <>
                  <Check size={13} className="text-emerald-400" /> Copied to Clipboard!
                </>
              ) : (
                <>
                  <Copy size={13} /> 1-Click Copy
                </>
              )}
            </button>
          </div>

          <pre className="secret-box w-full bg-[#050810] border border-amber-500/30 rounded-xl p-4 text-xs font-mono text-amber-200 select-all overflow-x-auto break-all whitespace-pre-wrap">
            {clientSecret}
          </pre>
        </div>

        {/* Action Button */}
        <div className="pt-2">
          <button
            type="button"
            onClick={onClose}
            className="w-full py-2.5 rounded-xl text-sm font-semibold bg-[var(--color-primary)] text-white hover:bg-blue-600 active:translate-y-px shadow-lg shadow-blue-500/20 transition-all cursor-pointer flex items-center justify-center gap-2"
          >
            <ShieldCheck size={16} /> I've Saved My Secret
          </button>
        </div>
      </div>
    </div>
  )
}
