import { Lock, Globe, CheckCircle2 } from 'lucide-react'

export interface TypeSelectorProps {
  isPublic: boolean
  onChange: (isPublic: boolean) => void
}

export default function TypeSelector({ isPublic, onChange }: TypeSelectorProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {/* Confidential Server-Side Client */}
      <button
        type="button"
        onClick={() => onChange(false)}
        className={`relative text-left p-5 rounded-xl border cursor-pointer transition-all duration-200 bg-[var(--color-card)] ${
          !isPublic
            ? 'border-[var(--color-primary)] bg-blue-500/10 shadow-[0_0_20px_rgba(59,130,246,0.15)] ring-1 ring-[var(--color-primary)]'
            : 'border-[var(--color-border)] hover:border-[var(--color-border-mid)] hover:bg-[var(--color-hover)]'
        }`}
      >
        {!isPublic && (
          <div className="absolute top-3 right-3 text-[var(--color-primary)]">
            <CheckCircle2 size={20} className="fill-[var(--color-primary)]/20" />
          </div>
        )}
        <div className="flex items-center gap-2.5 mb-2">
          <div
            className={`p-2 rounded-lg ${
              !isPublic
                ? 'bg-blue-500/20 text-blue-400'
                : 'bg-[var(--color-input)] text-[var(--color-sub)] border border-[var(--color-border)]'
            }`}
          >
            <Lock size={20} />
          </div>
          <div>
            <h3
              className={`font-bold text-sm sm:text-base ${
                !isPublic ? 'text-white' : 'text-[var(--color-txt)]'
              }`}
            >
              Confidential Server-Side App
            </h3>
            <span className="inline-block mt-0.5 px-2 py-0.5 text-[0.68rem] font-semibold uppercase tracking-wider rounded-md bg-blue-500/15 border border-blue-500/30 text-blue-300">
              Client Secret Flow
            </span>
          </div>
        </div>
        <p className="text-xs sm:text-[0.82rem] text-[var(--color-muted)] leading-relaxed mt-3">
          For Node.js, Go, Python, or Java web applications running on a server. Secures credentials via client secret.
        </p>
      </button>

      {/* Public SPA / Native App */}
      <button
        type="button"
        onClick={() => onChange(true)}
        className={`relative text-left p-5 rounded-xl border cursor-pointer transition-all duration-200 bg-[var(--color-card)] ${
          isPublic
            ? 'border-[var(--color-cyan)] bg-cyan-500/10 shadow-[0_0_20px_rgba(6,182,212,0.15)] ring-1 ring-[var(--color-cyan)]'
            : 'border-[var(--color-border)] hover:border-[var(--color-border-mid)] hover:bg-[var(--color-hover)]'
        }`}
      >
        {isPublic && (
          <div className="absolute top-3 right-3 text-[var(--color-cyan)]">
            <CheckCircle2 size={20} className="fill-[var(--color-cyan)]/20" />
          </div>
        )}
        <div className="flex items-center gap-2.5 mb-2">
          <div
            className={`p-2 rounded-lg ${
              isPublic
                ? 'bg-cyan-500/20 text-cyan-400'
                : 'bg-[var(--color-input)] text-[var(--color-sub)] border border-[var(--color-border)]'
            }`}
          >
            <Globe size={20} />
          </div>
          <div>
            <h3
              className={`font-bold text-sm sm:text-base ${
                isPublic ? 'text-white' : 'text-[var(--color-txt)]'
              }`}
            >
              Public SPA / Native App
            </h3>
            <span className="inline-block mt-0.5 px-2 py-0.5 text-[0.68rem] font-semibold uppercase tracking-wider rounded-md bg-cyan-500/15 border border-cyan-500/30 text-cyan-300">
              PKCE Authorization Flow
            </span>
          </div>
        </div>
        <p className="text-xs sm:text-[0.82rem] text-[var(--color-muted)] leading-relaxed mt-3">
          For Single Page Apps (React, Vue), Mobile Apps (iOS/Android), or CLI tools. Uses PKCE without backend secrets.
        </p>
      </button>
    </div>
  )
}
