import { AlertTriangle } from 'lucide-react'

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
    <div className="modal-backdrop" onClick={onCancel}>
      <div className="modal-card" onClick={e => e.stopPropagation()}>
        <h3><AlertTriangle size={20} color="var(--danger)" /> {title}</h3>
        <p>{message}</p>
        <div className="modal-actions">
          <button className="btn btn-outline" onClick={onCancel} disabled={loading}>Cancel</button>
          <button className="btn btn-danger" onClick={onConfirm} disabled={loading}>
            {loading ? 'Working…' : confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
