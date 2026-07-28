import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle2, AlertTriangle, AlertCircle, ShieldCheck, Copy, Check, Loader2, Globe, Lock, ArrowLeft } from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'

interface CreatedResult { app: OAuthApp; secret?: string }

const inputCls = "w-full px-3 py-2.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded-md text-sm text-[var(--color-txt)] placeholder:text-[var(--color-muted)] focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)] hover:border-[var(--color-border-mid)] transition-all duration-150"
const labelCls = "block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5"

export default function NewApp() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [homepageURL, setHomepageURL] = useState('')
  const [redirectURIs, setRedirectURIs] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [created, setCreated] = useState<CreatedResult | null>(null)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    const uris = redirectURIs.split('\n').map(s => s.trim()).filter(Boolean)
    if (uris.length === 0) { setError('At least one redirect URI is required.'); return }
    setLoading(true)
    try {
      const r = await apps.create({ name, description, homepage_url: homepageURL, redirect_uris: uris, is_public: isPublic })
      setCreated({ app: r.data.app, secret: r.data.client_secret })
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Failed to create application.')
    } finally {
      setLoading(false)
    }
  }

  const copySecret = () => {
    if (created?.secret) {
      navigator.clipboard.writeText(created.secret)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  if (created) {
    return (
      <>
        <Navbar />
        <div className="w-full max-w-[460px] mx-auto px-4 py-8">
          <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 mt-6">
            <div className="text-center mb-6">
              <div className="inline-flex bg-emerald-500/15 border border-emerald-400/30 p-3 rounded-2xl text-emerald-400 mb-3">
                <CheckCircle2 size={32} />
              </div>
              <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight">Application Created!</h1>
              <p className="text-[var(--color-muted)] text-[0.9rem] mt-1">Your OAuth 2.0 application credentials are ready.</p>
            </div>

            <div className="mb-4">
              <label className={labelCls}>Client ID</label>
              <pre className="secret-box text-sky-400 border-sky-400/30">{created.app.ClientID}</pre>
            </div>

            {created.secret && (
              <div className="mb-4">
                <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-3 bg-amber-500/10 border border-amber-500/28 text-amber-300 text-sm">
                  <AlertTriangle size={18} className="shrink-0 mt-0.5" />
                  <div>Save this Client Secret now! It will <strong>never be shown again</strong>.</div>
                </div>
                <label className={labelCls}>Client Secret</label>
                <pre className="secret-box mb-2">{created.secret}</pre>
                <button
                  onClick={copySecret}
                  className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md
                    border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
                    hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
                >
                  {copied ? <><Check size={14} className="text-emerald-400" /> Copied to Clipboard!</> : <><Copy size={14} /> Copy Secret</>}
                </button>
              </div>
            )}

            {isPublic && (
              <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-blue-500/10 border border-blue-500/28 text-blue-300 text-sm">
                <ShieldCheck size={18} className="shrink-0 mt-0.5" />
                <div>Registered as a <strong>Public Client</strong> (PKCE mode). No client secret is required.</div>
              </div>
            )}

            <div className="h-px bg-[var(--color-border)] my-4" />
            <div className="flex gap-3">
              <button
                onClick={() => navigate(`/apps/${created.app.ID}`)}
                className="flex-1 flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] transition-all duration-150 cursor-pointer"
              >
                View App Details →
              </button>
              <button
                onClick={() => navigate('/apps')}
                className="inline-flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg text-sm font-semibold
                  bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                  hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)] transition-all duration-150 cursor-pointer"
              >
                Dashboard
              </button>
            </div>
          </div>
        </div>
      </>
    )
  }

  return (
    <>
      <Navbar />
      <div className="w-full max-w-[460px] mx-auto px-4 py-8">
        <div className="mb-3">
          <button
            onClick={() => navigate('/apps')}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md
              border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
              hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
          >
            <ArrowLeft size={14} /> Back to Applications
          </button>
        </div>

        <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8">
          <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight mb-1">Register New Application</h1>
          <p className="text-[var(--color-muted)] text-[0.88rem] mb-5">
            Get client credentials to integrate DilithiAuth Post-Quantum identity into your service.
          </p>

          {error && (
            <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
              <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className={labelCls}>Application Name *</label>
              <input
                type="text" value={name} onChange={e => setName(e.target.value)}
                required placeholder="My Secure App" autoFocus className={inputCls}
              />
            </div>

            {/* Client Type Selector */}
            <div>
              <label className={labelCls}>Client Type</label>
              <div className="grid grid-cols-2 gap-3">
                {[
                  { pub: false, icon: Lock, label: 'Confidential Client', desc: 'For server-side web apps with protected backend secrets.', color: 'text-blue-300', activeBg: 'bg-blue-500/12 border-blue-500/50', hoverBg: 'hover:border-[var(--color-border-mid)]' },
                  { pub: true, icon: Globe, label: 'Public Client (PKCE)', desc: 'For SPAs, Mobile Apps, or CLI tools without backends.', color: 'text-cyan-400', activeBg: 'bg-cyan-500/12 border-cyan-500/50', hoverBg: 'hover:border-[var(--color-border-mid)]' },
                ].map(({ pub, icon: Icon, label, desc, color, activeBg, hoverBg }) => (
                  <button
                    key={label}
                    type="button"
                    onClick={() => setIsPublic(pub)}
                    className={`text-left p-3.5 rounded-lg border cursor-pointer transition-all duration-150 bg-transparent
                      ${isPublic === pub ? activeBg : `border-[var(--color-border)] bg-[var(--color-input)] ${hoverBg}`}`}
                  >
                    <div className={`flex items-center gap-1.5 font-bold text-[0.87rem] mb-1 ${isPublic === pub ? color : 'text-[var(--color-txt)]'}`}>
                      <Icon size={15} /> {label}
                    </div>
                    <p className="text-[0.77rem] text-[var(--color-muted)] leading-relaxed">{desc}</p>
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className={labelCls}>Homepage URL</label>
              <input type="text" value={homepageURL} onChange={e => setHomepageURL(e.target.value)} placeholder="https://myapp.com" className={inputCls} />
            </div>

            <div>
              <label className={labelCls}>Description</label>
              <textarea
                value={description} onChange={e => setDescription(e.target.value)}
                placeholder="Brief summary of your app..."
                className={`${inputCls} min-h-[65px] resize-y`}
              />
            </div>

            <div>
              <label className={labelCls}>Redirect URIs *</label>
              <textarea
                value={redirectURIs} onChange={e => setRedirectURIs(e.target.value)}
                required
                placeholder={"https://myapp.com/callback\nhttps://myapp.com/auth/callback"}
                className={`${inputCls} min-h-[80px] resize-y font-mono text-[0.85rem]`}
              />
              <p className="text-[0.77rem] text-[var(--color-muted)] mt-1.5">
                One URI per line. Must use HTTPS (http://localhost permitted for dev).
              </p>
            </div>

            <div className="flex gap-3 pt-2">
              <button
                type="submit" disabled={loading}
                className="flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-semibold
                  bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                  hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
                  transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? <><Loader2 size={16} className="spin" /> Registering App…</> : 'Register Application'}
              </button>
              <button
                type="button" onClick={() => navigate('/apps')}
                className="inline-flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-semibold
                  bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                  hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)] transition-all duration-150 cursor-pointer"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </>
  )
}
