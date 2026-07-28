import { useState, useRef } from 'react'
import type { FormEvent, KeyboardEvent, ClipboardEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { User, Mail, Lock, AlertCircle, CheckCircle2, Loader2, MailCheck, RotateCw, ShieldCheck, ArrowRight } from 'lucide-react'
import { auth } from '../api'
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

const inputCls = "w-full py-2.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded-md text-sm text-[var(--color-txt)] placeholder:text-[var(--color-muted)] focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)] hover:border-[var(--color-border-mid)] transition-all duration-150"
const labelCls = "block text-[0.72rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5"

export default function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [otpDigits, setOtpDigits] = useState<string[]>(['', '', '', '', '', ''])
  const inputRefs = useRef<(HTMLInputElement | null)[]>([])
  const [step, setStep] = useState<'form' | 'verify'>('form')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [loading, setLoading] = useState(false)

  const handleRegisterSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirm) { setError('Passwords do not match.'); return }
    if (password.length < 8) { setError('Password must be at least 8 characters.'); return }
    setLoading(true)
    try {
      await auth.register(email, name, password)
      setStep('verify')
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Registration failed.')
    } finally {
      setLoading(false)
    }
  }

  const handleDigitChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return
    const newDigits = [...otpDigits]
    newDigits[index] = value.slice(-1)
    setOtpDigits(newDigits)
    if (value && index < 5) inputRefs.current[index + 1]?.focus()
  }

  const handleDigitKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) inputRefs.current[index - 1]?.focus()
  }

  const handleDigitPaste = (e: ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const pastedData = e.clipboardData.getData('text').trim()
    if (/^\d{6}$/.test(pastedData)) {
      setOtpDigits(pastedData.split(''))
      inputRefs.current[5]?.focus()
    }
  }

  const handleVerifySubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    const code = otpDigits.join('')
    if (code.length !== 6) { setError('Please enter a valid 6-digit verification code.'); return }
    setLoading(true)
    try {
      await auth.verifyEmail(email, code)
      setSuccess('Email verified successfully! Redirecting…')
      setTimeout(() => navigate('/apps'), 1200)
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Invalid verification code.')
    } finally {
      setLoading(false)
    }
  }

  const handleResend = async () => {
    setError('')
    try {
      await auth.resendCode(email)
      setSuccess('A new 6-digit verification code has been sent!')
      setTimeout(() => setSuccess(''), 3000)
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Failed to resend code.')
    }
  }

  return (
    <>
      <Navbar />
      <div className="w-full max-w-[460px] mx-auto px-4 py-8">
        <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 mt-6">

          {step === 'form' ? (
            <>
              <div className="text-center mb-6">
                <div className="inline-flex bg-cyan-500/15 border border-cyan-400/30 p-3 rounded-2xl text-cyan-400 mb-3">
                  <ShieldCheck size={28} />
                </div>
                <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight">Create Developer Account</h1>
                <p className="text-[var(--color-muted)] text-[0.9rem] mt-1">Start building quantum-safe identity integrations</p>
              </div>

              {error && (
                <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
                  <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
                </div>
              )}

              <form onSubmit={handleRegisterSubmit} className="space-y-4">
                {[
                  { label: 'Full Name', type: 'text', val: name, set: setName, placeholder: 'Alice Smith', icon: User, autoFocus: true },
                  { label: 'Email Address', type: 'email', val: email, set: setEmail, placeholder: 'alice@example.com', icon: Mail },
                  { label: 'Password', type: 'password', val: password, set: setPassword, placeholder: 'At least 8 characters', icon: Lock },
                  { label: 'Confirm Password', type: 'password', val: confirm, set: setConfirm, placeholder: '••••••••••••', icon: Lock },
                ].map(({ label, type, val, set, placeholder, icon: Icon, autoFocus }) => (
                  <div key={label}>
                    <label className={labelCls}>{label}</label>
                    <div className="relative">
                      <Icon size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted)] pointer-events-none" />
                      <input
                        type={type}
                        value={val}
                        onChange={e => set(e.target.value)}
                        required
                        placeholder={placeholder}
                        autoFocus={autoFocus}
                        className={`${inputCls} pl-9 pr-3`}
                      />
                    </div>
                  </div>
                ))}

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-semibold
                    bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                    hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
                    transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed mt-2"
                >
                  {loading ? <><Loader2 size={18} className="spin" /> Creating Account…</> : <>Continue to Verification <ArrowRight size={18} /></>}
                </button>
              </form>

              <div className="relative flex items-center my-5">
                <div className="flex-1 h-px bg-[var(--color-border)]" />
                <span className="px-3 text-[0.78rem] text-[var(--color-muted)] bg-[var(--color-card)]">or continue with</span>
                <div className="flex-1 h-px bg-[var(--color-border)]" />
              </div>

              <div className="flex gap-3">
                {[
                  { href: 'google', svg: GOOGLE_SVG, label: 'Google' },
                  { href: 'github', svg: GITHUB_SVG, label: 'GitHub' },
                ].map(({ href, svg, label }) => (
                  <a
                    key={href}
                    href={`https://dilithiauthapis.gradientgeeks.tech/api/v1/bff/login?provider=${href}`}
                    className="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 rounded-lg text-sm font-semibold
                      bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                      hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)]
                      transition-all duration-150 no-underline hover:no-underline"
                  >
                    {svg} {label}
                  </a>
                ))}
              </div>

              <div className="text-center mt-5 pt-5 border-t border-[var(--color-border)]">
                <p className="text-[var(--color-muted)] text-[0.88rem]">
                  Already have an account?{' '}
                  <Link to="/login" className="text-[var(--color-primary)] font-semibold hover:text-blue-400">Sign in →</Link>
                </p>
              </div>
            </>
          ) : (
            <>
              <div className="text-center mb-6">
                <div className="inline-flex bg-emerald-500/15 border border-emerald-400/30 p-3 rounded-2xl text-emerald-400 mb-3">
                  <MailCheck size={28} />
                </div>
                <h1 className="text-[1.45rem] font-bold text-[var(--color-txt)] tracking-tight">Verify Your Email</h1>
                <p className="text-[var(--color-sub)] text-[0.9rem] mt-1">
                  We sent a 6-digit verification code to{' '}
                  <strong className="text-[var(--color-txt)]">{email}</strong>
                </p>
              </div>

              {error && (
                <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
                  <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
                </div>
              )}
              {success && (
                <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-emerald-500/10 border border-emerald-500/28 text-emerald-300 text-sm">
                  <CheckCircle2 size={18} className="shrink-0 mt-0.5" /><div>{success}</div>
                </div>
              )}

              <form onSubmit={handleVerifySubmit}>
                <label className={`${labelCls} text-center`}>Enter 6-Digit Security Code</label>
                <div className="flex justify-center gap-2 my-4">
                  {otpDigits.map((digit, idx) => (
                    <input
                      key={idx}
                      ref={el => { inputRefs.current[idx] = el }}
                      type="text"
                      maxLength={1}
                      value={digit}
                      onChange={e => handleDigitChange(idx, e.target.value)}
                      onKeyDown={e => handleDigitKeyDown(idx, e)}
                      onPaste={handleDigitPaste}
                      autoFocus={idx === 0}
                      className="w-11 h-13 text-center text-xl font-bold font-mono
                        bg-[var(--color-input)] border border-[var(--color-border)] rounded-md
                        text-[var(--color-txt)] focus:outline-none focus:border-[var(--color-focus)]
                        focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)] transition-all duration-150"
                    />
                  ))}
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full flex items-center justify-center gap-2 py-3 px-4 rounded-lg text-sm font-semibold
                    bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                    hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] active:translate-y-px
                    transition-all duration-150 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? <><Loader2 size={18} className="spin" /> Verifying Code…</> : 'Verify Email & Enter Dashboard'}
                </button>
              </form>

              <div className="flex justify-between mt-5 pt-4 border-t border-[var(--color-border)]">
                <button
                  type="button"
                  onClick={handleResend}
                  className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-semibold rounded-md
                    border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
                    hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
                >
                  <RotateCw size={14} /> Resend Code
                </button>
                <button
                  type="button"
                  onClick={() => setStep('form')}
                  className="inline-flex items-center gap-1.5 px-3 py-2 text-xs font-semibold rounded-md
                    border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
                    hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
                >
                  Change Email
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  )
}
