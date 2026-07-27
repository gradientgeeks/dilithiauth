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
