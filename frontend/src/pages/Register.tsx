import { useState, useRef } from 'react'
import type { FormEvent, KeyboardEvent, ClipboardEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { User, Mail, Lock, AlertCircle, CheckCircle2, Loader2, MailCheck, RotateCw, ShieldCheck, ArrowRight } from 'lucide-react'
import { auth } from '../api'
import Navbar from '../components/Navbar'

export default function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  
  // 6-digit OTP state array
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

  // Handle OTP digit changes
  const handleDigitChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return
    const newDigits = [...otpDigits]
    newDigits[index] = value.slice(-1)
    setOtpDigits(newDigits)

    if (value && index < 5) {
      inputRefs.current[index + 1]?.focus()
    }
  }

  const handleDigitKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      inputRefs.current[index - 1]?.focus()
    }
  }

  const handleDigitPaste = (e: ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault()
    const pastedData = e.clipboardData.getData('text').trim()
    if (/^\d{6}$/.test(pastedData)) {
      const digits = pastedData.split('')
      setOtpDigits(digits)
      inputRefs.current[5]?.focus()
    }
  }

  const handleVerifySubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    const code = otpDigits.join('')
    if (code.length !== 6) {
      setError('Please enter a valid 6-digit verification code.')
      return
    }
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
      <div className="container-sm">
        <div className="card card-glow" style={{ marginTop: '2.5rem' }}>
          {step === 'form' ? (
            <>
              <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
                <div style={{ display: 'inline-flex', background: 'rgba(6, 182, 212, 0.15)', padding: '0.75rem', borderRadius: '14px', color: 'var(--cyan)', marginBottom: '0.75rem', border: '1px solid rgba(6, 182, 212, 0.3)' }}>
                  <ShieldCheck size={28} />
                </div>
                <h1>Create Developer Account</h1>
                <p className="text-muted" style={{ fontSize: '0.9rem' }}>
                  Start building quantum-safe identity integrations
                </p>
              </div>

              {error && (
                <div className="alert alert-error">
                  <AlertCircle size={18} />
                  <div>{error}</div>
                </div>
              )}

              <form onSubmit={handleRegisterSubmit}>
                <div className="form-group">
                  <label className="label">Full Name</label>
                  <div className="input-icon-wrap">
                    <input
                      type="text"
                      value={name}
                      onChange={e => setName(e.target.value)}
                      required
                      placeholder="Alice Smith"
                      autoFocus
                    />
                    <User size={16} />
                  </div>
                </div>

                <div className="form-group">
                  <label className="label">Email Address</label>
                  <div className="input-icon-wrap">
                    <input
                      type="email"
                      value={email}
                      onChange={e => setEmail(e.target.value)}
                      required
                      placeholder="alice@example.com"
                    />
                    <Mail size={16} />
                  </div>
                </div>

                <div className="form-group">
                  <label className="label">Password</label>
                  <div className="input-icon-wrap">
                    <input
                      type="password"
                      value={password}
                      onChange={e => setPassword(e.target.value)}
                      required
                      placeholder="At least 8 characters"
                    />
                    <Lock size={16} />
                  </div>
                </div>

                <div className="form-group">
                  <label className="label">Confirm Password</label>
                  <div className="input-icon-wrap">
                    <input
                      type="password"
                      value={confirm}
                      onChange={e => setConfirm(e.target.value)}
                      required
                      placeholder="••••••••••••"
                    />
                    <Lock size={16} />
                  </div>
                </div>

                <button type="submit" className="btn btn-primary btn-full" disabled={loading} style={{ padding: '0.75rem', marginTop: '0.5rem' }}>
                  {loading ? (
                    <><Loader2 size={18} className="spin" /> Creating Account…</>
                  ) : (
                    <>Continue to Verification <ArrowRight size={18} /></>
                  )}
                </button>
              </form>

              <div style={{ margin: '1.5rem 0', textAlign: 'center', position: 'relative' }}>
                <div style={{ height: '1px', background: 'var(--border-color)', width: '100%', position: 'absolute', top: '50%', zIndex: 0 }}></div>
                <span style={{ background: 'var(--card-bg)', padding: '0 0.75rem', color: 'var(--text-muted)', fontSize: '0.8rem', position: 'relative', zIndex: 1 }}>
                  or continue with
                </span>
              </div>

              <div className="social-btn-group">
                <a
                  href="https://dilithiauthapis.gradientgeeks.tech/api/v1/bff/login?provider=google"
                  className="btn btn-outline"
                  style={{ flex: 1, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', textDecoration: 'none', padding: '0.65rem' }}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24">
                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/>
                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
                  </svg>
                  Google
                </a>
                <a
                  href="https://dilithiauthapis.gradientgeeks.tech/api/v1/bff/login?provider=github"
                  className="btn btn-outline"
                  style={{ flex: 1, display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem', textDecoration: 'none', padding: '0.65rem' }}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                    <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/>
                  </svg>
                  GitHub
                </a>
              </div>

              <div style={{ textAlign: 'center', marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.25rem' }}>
                <p className="text-muted" style={{ fontSize: '0.88rem' }}>
                  Already have an account?{' '}
                  <Link to="/login" style={{ fontWeight: 600 }}>Sign in →</Link>
                </p>
              </div>
            </>
          ) : (
            <>
              <div style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
                <div style={{ display: 'inline-flex', background: 'rgba(16, 185, 129, 0.15)', padding: '0.75rem', borderRadius: '14px', color: '#34d399', marginBottom: '0.75rem', border: '1px solid rgba(16, 185, 129, 0.3)' }}>
                  <MailCheck size={28} />
                </div>
                <h1>Verify Your Email</h1>
                <p className="text-muted" style={{ fontSize: '0.9rem', marginTop: '0.3rem' }}>
                  We sent a 6-digit verification code to <strong style={{ color: 'var(--text-main)' }}>{email}</strong>
                </p>
              </div>

              {error && <div className="alert alert-error"><AlertCircle size={18} /><div>{error}</div></div>}
              {success && <div className="alert alert-success"><CheckCircle2 size={18} /><div>{success}</div></div>}

              <form onSubmit={handleVerifySubmit}>
                <div className="form-group">
                  <label className="label" style={{ textAlign: 'center' }}>Enter 6-Digit Security Code</label>
                  
                  {/* Split OTP Boxes */}
                  <div className="otp-container">
                    {otpDigits.map((digit, idx) => (
                      <input
                        key={idx}
                        ref={el => { inputRefs.current[idx] = el }}
                        type="text"
                        maxLength={1}
                        className="otp-box"
                        value={digit}
                        onChange={e => handleDigitChange(idx, e.target.value)}
                        onKeyDown={e => handleDigitKeyDown(idx, e)}
                        onPaste={handleDigitPaste}
                        autoFocus={idx === 0}
                      />
                    ))}
                  </div>
                </div>

                <button type="submit" className="btn btn-primary btn-full" disabled={loading} style={{ padding: '0.75rem' }}>
                  {loading ? <><Loader2 size={18} className="spin" /> Verifying Code…</> : 'Verify Email & Enter Dashboard'}
                </button>
              </form>

              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1.5rem', borderTop: '1px solid var(--border-color)', paddingTop: '1.25rem' }}>
                <button type="button" className="btn btn-outline btn-sm" onClick={handleResend}>
                  <RotateCw size={14} /> Resend Code
                </button>
                <button type="button" className="btn btn-outline btn-sm" onClick={() => setStep('form')}>
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
