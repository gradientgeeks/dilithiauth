import { AlertTriangle, Loader2 } from 'lucide-react'

interface ConfirmModalProps {
  title: string
  message: string
  confirmLabel: string
  loading?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function ConfirmModal({ title, message, confirmLabel, loading, onConfirm, onCancel }: ConfirmModalProps) {
  return (
    <div
      className="fixed inset-0 z-[500] flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
      style={{ animation: 'fadeIn .15s ease' }}
      onClick={onCancel}
    >
      <div
        className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 w-full max-w-md shadow-2xl"
        onClick={e => e.stopPropagation()}
      >
        <h3 className="flex items-center gap-2 text-[var(--color-txt)] font-semibold text-[0.97rem] mb-2">
          <AlertTriangle size={20} className="text-rose-400 shrink-0" />
          {title}
        </h3>
        <p className="text-[var(--color-sub)] text-sm mb-5 leading-relaxed">{message}</p>
        <div className="flex gap-2 justify-end flex-wrap">
          <button
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-md text-sm font-semibold
              bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
              hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)]
              transition-all duration-150 cursor-pointer disabled:opacity-50"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </button>
          <button
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-md text-sm font-semibold
              bg-rose-500 text-white hover:bg-rose-600 hover:shadow-[0_0_0_3px_rgba(239,68,68,0.2)]
              transition-all duration-150 cursor-pointer disabled:opacity-50"
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? <><Loader2 size={16} className="spin" /> Working…</> : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
