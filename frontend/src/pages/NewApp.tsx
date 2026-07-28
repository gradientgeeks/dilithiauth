import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  ArrowRight,
  Plus,
  Trash2,
  AlertCircle,
  Loader2,
} from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import WizardStepper from '../components/new-app/WizardStepper'
import type { WizardStep } from '../components/new-app/WizardStepper'
import TypeSelector from '../components/new-app/TypeSelector'
import SecretSuccessModal from '../components/new-app/SecretSuccessModal'

const STEPS: WizardStep[] = [
  { id: 1, title: 'Basic Details', subtitle: 'Name & Description' },
  { id: 2, title: 'App Type', subtitle: 'Confidential or Public' },
  { id: 3, title: 'URIs & Config', subtitle: 'Redirect URIs & Finalize' },
]

const inputCls =
  "w-full px-3.5 py-2.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded-lg text-sm text-[var(--color-txt)] placeholder:text-[var(--color-muted)] focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)] hover:border-[var(--color-border-mid)] transition-all duration-150"
const labelCls = "block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5"

export default function NewApp() {
  const navigate = useNavigate()
  const [step, setStep] = useState<1 | 2 | 3>(1)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [homepageURL, setHomepageURL] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [redirectURIs, setRedirectURIs] = useState<string[]>(['http://localhost:3000/callback'])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [createdResult, setCreatedResult] = useState<{ app: OAuthApp; secret?: string } | null>(null)

  // Step 1 Validation & Progression
  const handleStep1Next = (e?: FormEvent) => {
    e?.preventDefault()
    setError('')
    if (!name.trim()) {
      setError('Application Name is required.')
      return
    }
    setStep(2)
  }

  // Step 2 Progression
  const handleStep2Next = () => {
    setError('')
    setStep(3)
  }

  // URI Manager Row Actions
  const handleURIChange = (index: number, val: string) => {
    const next = [...redirectURIs]
    next[index] = val
    setRedirectURIs(next)
  }

  const addURIRow = () => {
    setRedirectURIs([...redirectURIs, ''])
  }

  const removeURIRow = (index: number) => {
    if (redirectURIs.length <= 1) return
    setRedirectURIs(redirectURIs.filter((_, i) => i !== index))
  }

  // Final API Submission
  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    const validURIs = redirectURIs.map(u => u.trim()).filter(Boolean)
    if (validURIs.length === 0) {
      setError('At least one redirect URI is required.')
      return
    }
    setLoading(true)
    try {
      const res = await apps.create({
        name: name.trim(),
        description: description.trim(),
        homepage_url: homepageURL.trim(),
        redirect_uris: validURIs,
        is_public: isPublic,
      })
      setCreatedResult({ app: res.data.app, secret: res.data.client_secret })
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Failed to create application.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="w-full max-w-[760px] mx-auto px-4 py-6 sm:py-8">
      {/* Back Button Header */}
      <div className="mb-4">
        <button
          type="button"
          onClick={() => navigate('/apps')}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md
            border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
            hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
        >
          <ArrowLeft size={14} /> Back to Applications
        </button>
      </div>

      {/* Card Container */}
      <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 shadow-xl">
        <div className="mb-6">
          <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight mb-1">
            Register New Application
          </h1>
          <p className="text-[var(--color-muted)] text-[0.88rem]">
            Get client credentials to integrate DilithiAuth Post-Quantum identity into your service.
          </p>
        </div>

        {/* Visual Progress Stepper */}
        <WizardStepper
          currentStep={step}
          steps={STEPS}
          onStepClick={(targetStep) => {
            if (targetStep === 1) setStep(1)
            else if (targetStep === 2 && name.trim()) setStep(2)
            else if (targetStep === 3 && name.trim()) setStep(3)
            else if (!name.trim()) setError('Application Name is required.')
          }}
        />

        {/* Error Alert Banner */}
        {error && (
          <div className="flex items-start gap-2.5 p-3.5 rounded-lg mb-6 bg-rose-500/10 border border-rose-500/30 text-rose-300 text-sm">
            <AlertCircle size={18} className="shrink-0 mt-0.5 text-rose-400" />
            <div>{error}</div>
          </div>
        )}

        {/* STEP 1: Basic Info */}
        {step === 1 && (
          <form onSubmit={handleStep1Next} className="space-y-5">
            <div>
              <label className={labelCls}>
                Application Name <span className="text-rose-400">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g., Portal Service, Dashboard SPA"
                autoFocus
                className={inputCls}
              />
              <p className="text-[0.77rem] text-[var(--color-muted)] mt-1">
                A descriptive display name for your application.
              </p>
            </div>

            <div>
              <label className={labelCls}>Homepage URL</label>
              <input
                type="text"
                value={homepageURL}
                onChange={(e) => setHomepageURL(e.target.value)}
                placeholder="https://myapp.example.com"
                className={inputCls}
              />
              <p className="text-[0.77rem] text-[var(--color-muted)] mt-1">
                Your application's main website or landing page.
              </p>
            </div>

            <div>
              <label className={labelCls}>Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Brief summary of your application's purpose..."
                className={`${inputCls} min-h-[90px] resize-y`}
              />
            </div>

            <div className="flex justify-end pt-3 border-t border-[var(--color-border)]">
              <button
                type="submit"
                className="inline-flex items-center justify-center gap-2 py-2.5 px-5 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] transition-all duration-150 cursor-pointer"
              >
                <span>Next: Application Type</span>
                <ArrowRight size={16} />
              </button>
            </div>
          </form>
        )}

        {/* STEP 2: Application Type */}
        {step === 2 && (
          <div className="space-y-6">
            <div>
              <label className={labelCls}>Select Client Architecture</label>
              <p className="text-xs text-[var(--color-muted)] mb-4">
                Choose the integration type based on how and where your code executes.
              </p>
              <TypeSelector isPublic={isPublic} onChange={setIsPublic} />
            </div>

            <div className="flex items-center justify-between pt-4 border-t border-[var(--color-border)]">
              <button
                type="button"
                onClick={() => setStep(1)}
                className="inline-flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
                  bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                  hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)] transition-all duration-150 cursor-pointer"
              >
                <ArrowLeft size={16} />
                <span>Back to Basic Info</span>
              </button>
              <button
                type="button"
                onClick={handleStep2Next}
                className="inline-flex items-center justify-center gap-2 py-2.5 px-5 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] transition-all duration-150 cursor-pointer"
              >
                <span>Next: URIs & Config</span>
                <ArrowRight size={16} />
              </button>
            </div>
          </div>
        )}

        {/* STEP 3: URIs & Configuration */}
        {step === 3 && (
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className={labelCls}>
                  Allowed Redirect URIs <span className="text-rose-400">*</span>
                </label>
                <button
                  type="button"
                  onClick={addURIRow}
                  className="inline-flex items-center gap-1 text-xs font-semibold text-[var(--color-primary)] hover:underline cursor-pointer"
                >
                  <Plus size={14} /> Add URI
                </button>
              </div>

              <div className="space-y-2 mb-2">
                {redirectURIs.map((uri, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    <input
                      type="text"
                      value={uri}
                      onChange={(e) => handleURIChange(idx, e.target.value)}
                      placeholder="http://localhost:3000/callback"
                      className={`${inputCls} font-mono text-xs sm:text-sm`}
                    />
                    {redirectURIs.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removeURIRow(idx)}
                        title="Remove URI"
                        className="p-2.5 text-[var(--color-sub)] hover:text-rose-400 hover:bg-rose-500/10 border border-[var(--color-border)] rounded-lg transition-all cursor-pointer"
                      >
                        <Trash2 size={16} />
                      </button>
                    )}
                  </div>
                ))}
              </div>
              <p className="text-[0.77rem] text-[var(--color-muted)]">
                Redirect callback URIs where authorization codes are returned. Must use HTTPS (except localhost).
              </p>
            </div>

            {/* Summary Review Box */}
            <div className="bg-[var(--color-input)] border border-[var(--color-border)] rounded-xl p-4 space-y-2 text-xs text-[var(--color-sub)]">
              <div className="font-bold uppercase tracking-wider text-[var(--color-muted)] text-[0.68rem] mb-2">
                Application Summary
              </div>
              <div className="flex justify-between">
                <span>Name:</span>
                <strong className="text-[var(--color-txt)]">{name}</strong>
              </div>
              <div className="flex justify-between">
                <span>Type:</span>
                <strong className={isPublic ? 'text-cyan-400' : 'text-blue-400'}>
                  {isPublic ? 'Public SPA (PKCE)' : 'Confidential Server App'}
                </strong>
              </div>
              {homepageURL && (
                <div className="flex justify-between">
                  <span>Homepage:</span>
                  <span className="text-[var(--color-txt)] font-mono">{homepageURL}</span>
                </div>
              )}
            </div>

            <div className="flex items-center justify-between pt-4 border-t border-[var(--color-border)]">
              <button
                type="button"
                onClick={() => setStep(2)}
                disabled={loading}
                className="inline-flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
                  bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                  hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)] transition-all duration-150 cursor-pointer disabled:opacity-50"
              >
                <ArrowLeft size={16} />
                <span>Back to App Type</span>
              </button>
              <button
                type="submit"
                disabled={loading}
                className="inline-flex items-center justify-center gap-2 py-3 px-6 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
                  transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? (
                  <>
                    <Loader2 size={16} className="spin" /> Creating Application…
                  </>
                ) : (
                  'Create Application'
                )}
              </button>
            </div>
          </form>
        )}
      </div>

      {/* Post-Creation Success Modal Overlay */}
      {createdResult && (
        <SecretSuccessModal
          app={createdResult.app}
          secret={createdResult.secret}
          onNavigateToApp={(appId) => navigate(`/apps/${appId}`)}
          onNavigateToDashboard={() => navigate('/apps')}
        />
      )}
    </div>
  )
}
