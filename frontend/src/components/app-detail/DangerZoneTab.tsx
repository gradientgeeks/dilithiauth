import { RotateCw, Trash2, ShieldAlert, KeyRound } from 'lucide-react'
import type { OAuthApp } from '../../api'

export interface DangerZoneTabProps {
  app: OAuthApp
  onOpenRotateModal: () => void
  onOpenDeleteModal: () => void
  rotating?: boolean
  deleting?: boolean
}

export default function DangerZoneTab({
  app,
  onOpenRotateModal,
  onOpenDeleteModal,
  rotating = false,
  deleting = false,
}: DangerZoneTabProps) {
  return (
    <div className="bg-rose-500/5 border border-rose-500/25 rounded-xl p-6 space-y-6">
      <div className="flex items-center gap-2 border-b border-rose-500/20 pb-4">
        <ShieldAlert size={20} className="text-rose-400" />
        <h2 className="text-rose-400 font-semibold text-lg">Danger Zone</h2>
      </div>

      <div className="space-y-6">
        {/* Row 1: Rotate Client Secret */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-lg bg-rose-500/5 border border-rose-500/15">
          <div className="space-y-1">
            <div className="font-semibold text-sm text-[var(--color-txt)] flex items-center gap-2">
              <KeyRound size={16} className="text-amber-400" /> Rotate Client Secret
            </div>
            <p className="text-xs text-[var(--color-sub)]">
              Generate a new client secret. The current secret will be immediately revoked and cannot be recovered.
            </p>
          </div>

          <div>
            {app.IsPublic ? (
              <span className="text-xs text-[var(--color-muted)] italic px-3 py-2 block">
                Public applications do not use client secrets.
              </span>
            ) : (
              <button
                type="button"
                onClick={onOpenRotateModal}
                disabled={rotating}
                className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 hover:bg-amber-500/20 active:translate-y-px transition-all cursor-pointer disabled:opacity-50 shrink-0"
              >
                <RotateCw size={15} className={rotating ? 'animate-spin' : ''} />
                {rotating ? 'Rotating...' : 'Rotate Client Secret'}
              </button>
            )}
          </div>
        </div>

        {/* Row 2: Delete Application */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-lg bg-rose-500/10 border border-rose-500/25">
          <div className="space-y-1">
            <div className="font-semibold text-sm text-rose-300 flex items-center gap-2">
              <Trash2 size={16} className="text-rose-400" /> Delete Application
            </div>
            <p className="text-xs text-[var(--color-sub)]">
              Permanently remove this application, revoking all credentials and invalidating all issued tokens.
            </p>
          </div>

          <button
            type="button"
            onClick={onOpenDeleteModal}
            disabled={deleting}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-semibold bg-rose-500 text-white hover:bg-rose-600 shadow-lg shadow-rose-500/20 active:translate-y-px transition-all cursor-pointer disabled:opacity-50 shrink-0"
          >
            <Trash2 size={15} />
            {deleting ? 'Deleting...' : 'Delete Application'}
          </button>
        </div>
      </div>
    </div>
  )
}
