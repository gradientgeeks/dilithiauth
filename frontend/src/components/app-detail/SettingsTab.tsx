import React, { useState, useEffect } from 'react'
import { Copy, Check, Plus, Trash2, Globe, Lock, ShieldCheck, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react'
import type { OAuthApp } from '../../api'

export interface SettingsTabProps {
  app: OAuthApp
  onSave: (updated: {
    name: string
    description: string
    homepageURL: string
    redirectURIs: string[]
  }) => Promise<void>
  saving: boolean
  saveSuccess: boolean
  saveError: string | null
}

export default function SettingsTab({ app, onSave, saving, saveSuccess, saveError }: SettingsTabProps) {
  const [name, setName] = useState(app.Name || '')
  const [description, setDescription] = useState(app.Description || '')
  const [homepageURL, setHomepageURL] = useState(app.HomepageURL || '')
  const [redirectURIs, setRedirectURIs] = useState<string[]>(
    app.RedirectURIs && app.RedirectURIs.length > 0 ? [...app.RedirectURIs] : ['']
  )
  const [copiedId, setCopiedId] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  useEffect(() => {
    setName(app.Name || '')
    setDescription(app.Description || '')
    setHomepageURL(app.HomepageURL || '')
    setRedirectURIs(app.RedirectURIs && app.RedirectURIs.length > 0 ? [...app.RedirectURIs] : [''])
  }, [app])

  const handleURIChange = (index: number, value: string) => {
    const next = [...redirectURIs]
    next[index] = value
    setRedirectURIs(next)
  }

  const addURIRow = () => {
    setRedirectURIs([...redirectURIs, ''])
  }

  const removeURIRow = (index: number) => {
    if (redirectURIs.length <= 1) return
    setRedirectURIs(redirectURIs.filter((_, i) => i !== index))
  }

  const handleCopyClientId = () => {
    navigator.clipboard.writeText(app.ClientID)
    setCopiedId(true)
    setTimeout(() => setCopiedId(false), 2000)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setValidationError(null)

    if (!name.trim()) {
      setValidationError('Application Name is required.')
      return
    }

    const filteredURIs = redirectURIs.map(u => u.trim()).filter(Boolean)
    if (filteredURIs.length === 0) {
      setValidationError('At least one Redirect URI is required.')
      return
    }

    await onSave({
      name: name.trim(),
      description: description.trim(),
      homepageURL: homepageURL.trim(),
      redirectURIs: filteredURIs,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Toast / Banner Messages */}
      {saveSuccess && (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-sm">
          <CheckCircle2 size={18} className="shrink-0 text-emerald-400" />
          <div className="font-medium">Settings updated successfully!</div>
        </div>
      )}

      {(saveError || validationError) && (
        <div className="flex items-center gap-3 p-4 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm">
          <AlertCircle size={18} className="shrink-0 text-rose-400" />
          <div>{saveError || validationError}</div>
        </div>
      )}

      {/* Main Settings Card */}
      <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 space-y-5">
        <h2 className="text-[var(--color-txt)] font-semibold text-lg border-b border-[var(--color-border)] pb-3">
          General Settings
        </h2>

        {/* Application Name */}
        <div>
          <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
            Application Name <span className="text-rose-400">*</span>
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Acme Web Portal"
            required
            className="w-full bg-[#050810] border border-[var(--color-border)] rounded-lg px-3.5 py-2.5 text-sm text-[var(--color-txt)] placeholder-[var(--color-muted)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] transition-all"
          />
        </div>

        {/* Description */}
        <div>
          <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
            Description
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe the purpose of this application..."
            rows={3}
            className="w-full bg-[#050810] border border-[var(--color-border)] rounded-lg px-3.5 py-2.5 text-sm text-[var(--color-txt)] placeholder-[var(--color-muted)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] transition-all resize-none"
          />
        </div>

        {/* Homepage URL */}
        <div>
          <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
            Homepage URL
          </label>
          <input
            type="url"
            value={homepageURL}
            onChange={(e) => setHomepageURL(e.target.value)}
            placeholder="https://myapp.example.com"
            className="w-full bg-[#050810] border border-[var(--color-border)] rounded-lg px-3.5 py-2.5 text-sm text-[var(--color-txt)] placeholder-[var(--color-muted)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] transition-all"
          />
        </div>

        {/* Multi-row Redirect URIs Manager */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider">
              Authorized Redirect URIs <span className="text-rose-400">*</span>
            </label>
            <span className="text-xs text-[var(--color-muted)]">
              OAuth 2.0 / OIDC callback endpoints
            </span>
          </div>

          <div className="space-y-2.5 mb-3">
            {redirectURIs.map((uri, index) => (
              <div key={index} className="flex gap-2 items-center">
                <input
                  type="text"
                  value={uri}
                  onChange={(e) => handleURIChange(index, e.target.value)}
                  placeholder="https://myapp.example.com/callback"
                  className="flex-1 bg-[#050810] border border-[var(--color-border)] rounded-lg px-3.5 py-2.5 text-sm font-mono text-[var(--color-txt)] placeholder-[var(--color-muted)] focus:outline-none focus:border-[var(--color-primary)] focus:ring-1 focus:ring-[var(--color-primary)] transition-all"
                />
                <button
                  type="button"
                  onClick={() => removeURIRow(index)}
                  disabled={redirectURIs.length <= 1}
                  className={`p-2.5 rounded-lg border transition-all ${
                    redirectURIs.length <= 1
                      ? 'opacity-40 cursor-not-allowed border-[var(--color-border)] text-[var(--color-muted)]'
                      : 'border-rose-500/30 text-rose-400 bg-rose-500/5 hover:bg-rose-500/15 hover:border-rose-500/50 cursor-pointer'
                  }`}
                  title={redirectURIs.length <= 1 ? 'At least one URI required' : 'Remove URI'}
                >
                  <Trash2 size={16} />
                </button>
              </div>
            ))}
          </div>

          <button
            type="button"
            onClick={addURIRow}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all cursor-pointer"
          >
            <Plus size={14} /> Add Redirect URI
          </button>
        </div>
      </div>

      {/* App Credentials Info Card */}
      <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 space-y-5">
        <h2 className="text-[var(--color-txt)] font-semibold text-lg border-b border-[var(--color-border)] pb-3">
          Application Identity &amp; Type
        </h2>

        {/* Client ID Display */}
        <div>
          <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
            Client ID
          </label>
          <div className="flex gap-2 items-center">
            <span className="mono-inline flex-1 px-3.5 py-2.5 text-sm bg-[#050810] border border-[var(--color-border)] rounded-lg font-mono text-[var(--color-txt)] select-all overflow-x-auto">
              {app.ClientID}
            </span>
            <button
              type="button"
              onClick={handleCopyClientId}
              className="inline-flex items-center gap-1.5 px-3.5 py-2.5 rounded-lg text-xs font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all cursor-pointer shrink-0"
            >
              {copiedId ? (
                <>
                  <Check size={14} className="text-emerald-400" /> Copied!
                </>
              ) : (
                <>
                  <Copy size={14} /> Copy Client ID
                </>
              )}
            </button>
          </div>
        </div>

        {/* Client Type Indicator */}
        <div>
          <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">
            Client Type
          </label>
          <div className="flex items-center gap-3">
            {app.IsPublic ? (
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-sky-500/10 border border-sky-500/30 text-sky-400">
                <Globe size={14} /> Public SPA (PKCE Mode)
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
                <Lock size={14} /> Confidential Client (Client Secret)
              </span>
            )}
          </div>
          <div className="flex items-start gap-2.5 p-3.5 rounded-lg mt-3 bg-blue-500/10 border border-blue-500/25 text-blue-300 text-xs leading-relaxed">
            <ShieldCheck size={16} className="shrink-0 mt-0.5 text-blue-400" />
            <div>
              {app.IsPublic
                ? 'Public applications (SPAs, mobile apps) execute in untrusted environments and use Proof Key for Code Exchange (PKCE) with S256 code challenge verification.'
                : 'Confidential server applications store client credentials securely on backend servers and authenticate using client secrets alongside authorization codes.'}
            </div>
          </div>
        </div>
      </div>

      {/* Submit Action */}
      <div className="flex justify-end pt-2">
        <button
          type="submit"
          disabled={saving}
          className="inline-flex items-center justify-center gap-2 px-6 py-2.5 rounded-lg text-sm font-semibold bg-[var(--color-primary)] text-white hover:bg-blue-600 active:translate-y-px shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed transition-all cursor-pointer"
        >
          {saving ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Saving Changes...
            </>
          ) : (
            'Save Changes'
          )}
        </button>
      </div>
    </form>
  )
}
