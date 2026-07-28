import React, { useState } from 'react'
import { AlertTriangle, Trash2, X, Loader2 } from 'lucide-react'

export interface DeleteAppModalProps {
  appName: string
  loading: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function DeleteAppModal({ appName, loading, onConfirm, onCancel }: DeleteAppModalProps) {
  const [typedName, setTypedName] = useState('')

  const isConfirmed = typedName.trim() === appName.trim()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isConfirmed && !loading) {
      onConfirm()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fadeIn">
      <div className="w-full max-w-md bg-[var(--color-card)] border border-[var(--color-border)] rounded-2xl p-6 shadow-2xl space-y-5 relative">
        {/* Close button */}
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className="absolute top-4 right-4 p-1 rounded-lg text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.05] transition-all cursor-pointer"
        >
          <X size={18} />
        </button>

        {/* Modal Header */}
        <div className="flex items-center gap-3 border-b border-[var(--color-border)] pb-4">
          <div className="w-10 h-10 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center justify-center text-rose-400">
            <AlertTriangle size={20} />
          </div>
          <div>
            <h2 className="text-lg font-bold text-rose-400">Delete Application</h2>
            <p className="text-xs text-[var(--color-sub)]">This action cannot be undone.</p>
          </div>
        </div>

        {/* Warning text */}
        <p className="text-xs text-[var(--color-sub)] leading-relaxed">
          You are about to permanently delete <strong className="text-[var(--color-txt)]">{appName}</strong>. All associated API keys, client credentials, and issued tokens will be permanently revoked.
        </p>

        {/* Name Verification Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
              To confirm deletion, please type <span className="text-[var(--color-txt)] font-mono select-all">"{appName}"</span> below:
            </label>
            <input
              type="text"
              value={typedName}
              onChange={(e) => setTypedName(e.target.value)}
              placeholder={appName}
              disabled={loading}
              className="w-full bg-[#050810] border border-[var(--color-border)] rounded-lg px-3.5 py-2.5 text-sm text-[var(--color-txt)] placeholder-[var(--color-muted)] focus:outline-none focus:border-rose-500 focus:ring-1 focus:ring-rose-500 transition-all font-mono"
            />
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onCancel}
              disabled={loading}
              className="px-4 py-2 rounded-lg text-sm font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all cursor-pointer"
            >
              Cancel
            </button>

            <button
              type="submit"
              disabled={!isConfirmed || loading}
              className={`inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-semibold transition-all cursor-pointer ${
                isConfirmed && !loading
                  ? 'bg-rose-500 text-white hover:bg-rose-600 shadow-lg shadow-rose-500/20 active:translate-y-px'
                  : 'bg-rose-500/30 text-rose-300/50 border border-rose-500/20 cursor-not-allowed'
              }`}
            >
              {loading ? (
                <>
                  <Loader2 size={16} className="animate-spin" /> Deleting...
                </>
              ) : (
                <>
                  <Trash2 size={15} /> Permanently Delete Application
                </>
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
