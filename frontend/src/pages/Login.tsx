import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Mail, Lock, AlertCircle, Loader2, LogIn, ShieldCheck } from 'lucide-react'
import { useAuth } from '../AuthContext'
import Navbar from '../components/Navbar'

const GOOGLE_SVG = (
  <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/>
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
  </svg>
)

const GITHUB_SVG = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
    <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
  </svg>
)

export default function Login() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await login(email, password)
      navigate('/apps')
    } catch {
      setError('Invalid email or password.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <>
      <Navbar />
      <div className="w-full max-w-[460px] mx-auto px-4 py-8">
        <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 mt-6">

          {/* Icon + title */}
          <div className="text-center mb-7">
            <div className="inline-flex bg-blue-500/15 border border-sky-400/30 p-3 rounded-2xl text-blue-300 mb-3">
              <ShieldCheck size={28} />
            </div>
            <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight">Sign in to Portal</h1>
            <p className="text-[var(--color-muted)] text-[0.9rem] mt-1">Access your quantum-safe developer dashboard</p>
          </div>

          {/* Error */}
          {error && (
            <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4
              bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
              <AlertCircle size={18} className="shrink-0 mt-0.5" />
              <div>{error}</div>
            </div>
          )}

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-[0.72rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5">
                Developer Email
              </label>
              <div className="relative">
                <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)] pointer-events-none" />
                <input
                  type="email"
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  required
                  placeholder="developer@example.com"
                  autoFocus
                  className="w-full pl-9 pr-3 py-2.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded-md
                    text-sm text-[var(--color-txt)] placeholder:text-[var(--color-muted)]
                    focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)]
                    hover:border-[var(--color-border-mid)] transition-all duration-150"
                />
              </div>
            </div>

            <div>
              <label className="block text-[0.72rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5">
                Password
              </label>
              <div className="relative">
                <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)] pointer-events-none" />
                <input
                  type="password"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  required
                  placeholder="••••••••••••"
                  className="w-full pl-9 pr-3 py-2.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded-md
                    text-sm text-[var(--color-txt)] placeholder:text-[var(--color-muted)]
                    focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)]
                    hover:border-[var(--color-border-mid)] transition-all duration-150"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-semibold
                bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
                transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed mt-2"
            >
              {loading ? <><Loader2 size={18} className="spin" /> Signing in…</> : <><LogIn size={18} /> Sign in to Dashboard</>}
            </button>
          </form>

          {/* Divider */}
          <div className="relative flex items-center my-5">
            <div className="flex-1 h-px bg-[var(--color-border)]" />
            <span className="px-3 text-[0.78rem] text-[var(--color-muted)] bg-[var(--color-card)]">or continue with</span>
            <div className="flex-1 h-px bg-[var(--color-border)]" />
          </div>

          {/* Social buttons */}
          <div className="flex gap-3">
            <a
              href="https://dilithiauthapis.gradientgeeks.tech/api/v1/bff/login?provider=google"
              className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-semibold
                bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)]
                transition-all duration-150 no-underline hover:no-underline"
            >
              {GOOGLE_SVG} Google
            </a>
            <a
              href="https://dilithiauthapis.gradientgeeks.tech/api/v1/bff/login?provider=github"
              className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-semibold
                bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)]
                transition-all duration-150 no-underline hover:no-underline"
            >
              {GITHUB_SVG} GitHub
            </a>
          </div>

          {/* Footer link */}
          <div className="text-center mt-5 pt-5 border-t border-[var(--color-border)]">
            <p className="text-[var(--color-muted)] text-[0.88rem]">
              Don't have a developer account?{' '}
              <Link to="/register" className="text-[var(--color-primary)] font-semibold hover:text-blue-400">
                Create one now →
              </Link>
            </p>
          </div>
        </div>
      </div>
    </>
  )
}
