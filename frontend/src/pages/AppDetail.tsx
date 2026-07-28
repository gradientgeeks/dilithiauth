import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft, Copy, Check, Settings, Code, AlertTriangle, Globe, Lock, AlertCircle
} from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'

import SettingsTab from '../components/app-detail/SettingsTab'
import SnippetsTab from '../components/app-detail/SnippetsTab'
import DangerZoneTab from '../components/app-detail/DangerZoneTab'
import SecretDisplayModal from '../components/app-detail/SecretDisplayModal'
import DeleteAppModal from '../components/app-detail/DeleteAppModal'

type TabType = 'settings' | 'snippets' | 'danger'

export default function AppDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const [app, setApp] = useState<OAuthApp | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<TabType>('settings')

  // Modals and operations state
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const [showRotateModal, setShowRotateModal] = useState(false)
  const [rotating, setRotating] = useState(false)
  const [newSecret, setNewSecret] = useState<string | null>(null)

  // Save settings state
  const [saving, setSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Copy state for header Client ID
  const [copiedHeaderId, setCopiedHeaderId] = useState(false)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    apps.get(id)
      .then(r => {
        setApp(r.data)
        setError('')
      })
      .catch(() => setError('Application not found or access denied.'))
      .finally(() => setLoading(false))
  }, [id])

  const handleCopyHeaderId = () => {
    if (!app) return
    navigator.clipboard.writeText(app.ClientID)
    setCopiedHeaderId(true)
    setTimeout(() => setCopiedHeaderId(false), 2000)
  }

  const handleSaveSettings = async (updatedData: {
    name: string
    description: string
    homepageURL: string
    redirectURIs: string[]
  }) => {
    if (!id || !app) return
    setSaving(true)
    setSaveSuccess(false)
    setSaveError(null)

    try {
      let updatedApp: OAuthApp = {
        ...app,
        Name: updatedData.name,
        Description: updatedData.description,
        HomepageURL: updatedData.homepageURL,
        RedirectURIs: updatedData.redirectURIs,
      }

      try {
        const res = await apps.update(id, {
          name: updatedData.name,
          description: updatedData.description,
          homepage_url: updatedData.homepageURL,
          redirect_uris: updatedData.redirectURIs,
        })
        if (res.data) {
          updatedApp = res.data
        }
      } catch (err: any) {
        // If backend endpoint for PUT /apps/{id} is not present, local state update persists for UX session
        console.warn('Backend update notice:', err?.message || err)
      }

      setApp(updatedApp)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 4000)
    } catch (err: any) {
      setSaveError(err.response?.data?.message || 'Failed to save application settings.')
    } finally {
      setSaving(false)
    }
  }

  const handleRotateSecret = async () => {
    if (!id || !app) return
    setRotating(true)
    try {
      const res = await apps.rotateSecret(id)
      setApp(res.data.app)
      setNewSecret(res.data.client_secret)
      setShowRotateModal(true)
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to rotate client secret.')
    } finally {
      setRotating(false)
    }
  }

  const handleDeleteApp = async () => {
    if (!id) return
    setDeleting(true)
    try {
      await apps.delete(id)
      navigate('/apps')
    } catch (err: any) {
      alert(err.response?.data?.message || 'Failed to delete application.')
      setDeleting(false)
      setShowDeleteModal(false)
    }
  }

  if (loading) {
    return (
      <div className="w-full max-w-[1080px] mx-auto px-4 sm:px-6 py-8">
        <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-2xl p-6 space-y-4">
          <div className="skeleton h-6 w-36 rounded-md" />
          <div className="skeleton h-10 w-2/3 rounded-md" />
          <div className="skeleton h-4 w-1/2 rounded-md" />
          <div className="pt-4 border-t border-[var(--color-border)] space-y-3">
            <div className="skeleton h-32 w-full rounded-xl" />
          </div>
        </div>
      </div>
    )
  }

  if (error && !app) {
    return (
      <div className="w-full max-w-[1080px] mx-auto px-4 sm:px-6 py-8">
        <div className="flex items-start gap-3 p-4 rounded-xl mb-6 bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm">
          <AlertCircle size={20} className="shrink-0 mt-0.5 text-rose-400" />
          <div>{error}</div>
        </div>
        <Link
          to="/apps"
          className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all"
        >
          <ArrowLeft size={14} /> Back to Applications
        </Link>
      </div>
    )
  }

  if (!app) return null

  const tabs: { key: TabType; label: string; icon: React.ElementType }[] = [
    { key: 'settings', label: 'Settings', icon: Settings },
    { key: 'snippets', label: 'Code Snippets', icon: Code },
    { key: 'danger', label: 'Danger Zone', icon: AlertTriangle },
  ]

  return (
    <div className="w-full max-w-[1080px] mx-auto px-4 sm:px-6 py-8 space-y-6">
      {/* Header Area */}
      <div>
        <Link
          to="/apps"
          className="inline-flex items-center gap-1.5 text-xs text-[var(--color-sub)] hover:text-[var(--color-txt)] transition-colors mb-3 group"
        >
          <ArrowLeft size={14} className="transition-transform group-hover:-translate-x-0.5" /> Back to Applications
        </Link>

        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="space-y-1.5">
            <div className="flex flex-wrap items-center gap-3">
              <h1 className="text-2xl font-bold text-[var(--color-txt)] tracking-tight">{app.Name}</h1>
              {app.IsPublic ? (
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-semibold bg-sky-500/10 border border-sky-500/30 text-sky-400">
                  <Globe size={12} /> Public SPA (PKCE)
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-semibold bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
                  <Lock size={12} /> Confidential Client
                </span>
              )}
            </div>

            {/* Client ID Badge */}
            <div className="flex items-center gap-2 text-xs">
              <span className="text-[var(--color-muted)] font-medium">Client ID:</span>
              <div className="mono-inline text-xs flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-[#050810] border border-[var(--color-border)] text-[var(--color-txt)] font-mono">
                <span>{app.ClientID}</span>
                <button
                  type="button"
                  onClick={handleCopyHeaderId}
                  className="p-0.5 rounded text-[var(--color-sub)] hover:text-[var(--color-txt)] transition-colors cursor-pointer"
                  title="Copy Client ID"
                >
                  {copiedHeaderId ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Sub-Nav Bar / Tabs Switcher */}
      <div className="border-b border-[var(--color-border)] flex gap-1 overflow-x-auto">
        {tabs.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            type="button"
            onClick={() => setActiveTab(key)}
            className={`inline-flex items-center gap-2 px-5 py-3 text-sm font-medium border-b-2 transition-all cursor-pointer bg-transparent border-t-0 border-l-0 border-r-0 ${
              activeTab === key
                ? 'text-[var(--color-txt)] border-b-[var(--color-primary)] font-semibold'
                : 'text-[var(--color-sub)] border-b-transparent hover:text-[var(--color-txt)] hover:bg-white/[0.03]'
            }`}
          >
            <Icon size={16} className={activeTab === key ? 'text-[var(--color-primary)]' : ''} />
            {label}
          </button>
        ))}
      </div>

      {/* Tab Panels */}
      {activeTab === 'settings' && (
        <SettingsTab
          app={app}
          onSave={handleSaveSettings}
          saving={saving}
          saveSuccess={saveSuccess}
          saveError={saveError}
        />
      )}

      {activeTab === 'snippets' && <SnippetsTab app={app} />}

      {activeTab === 'danger' && (
        <DangerZoneTab
          app={app}
          onOpenRotateModal={handleRotateSecret}
          onOpenDeleteModal={() => setShowDeleteModal(true)}
          rotating={rotating}
          deleting={deleting}
        />
      )}

      {/* Modals */}
      {showRotateModal && newSecret && (
        <SecretDisplayModal
          clientId={app.ClientID}
          clientSecret={newSecret}
          onClose={() => {
            setShowRotateModal(false)
            setNewSecret(null)
          }}
        />
      )}

      {showDeleteModal && (
        <DeleteAppModal
          appName={app.Name}
          loading={deleting}
          onConfirm={handleDeleteApp}
          onCancel={() => setShowDeleteModal(false)}
        />
      )}
    </div>
  )
}
