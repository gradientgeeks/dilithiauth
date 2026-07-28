import { Link } from 'react-router-dom'
import { ShieldCheck, Zap, Wrench, Building2, Lock, ArrowRight, Terminal, Shield } from 'lucide-react'
import Navbar from '../components/Navbar'
import QuantumTokenPlayground from '../components/QuantumTokenPlayground'

const features = [
  { icon: ShieldCheck, label: 'Hybrid Signatures', desc: 'Every token is dual-signed using Ed25519 classical keys and ML-DSA-44 post-quantum keypairs.' },
  { icon: Zap, label: 'OAuth 2.0 & PKCE', desc: 'Standard authorization code flow compatible with modern web, mobile, and confidential backends.' },
  { icon: Wrench, label: 'Optimized Token Modes', desc: 'Thin JWTs (Redis sidecar reference), Jumbo JWTs (embedded signatures), and CBOR Web Tokens (CWT).' },
  { icon: Building2, label: 'Multi-Tenancy', desc: 'Partition user pools and client applications by tenant subdomain or header isolation.' },
]

export default function Landing() {
  return (
    <>
      <Navbar />
      <div className="w-full max-w-[1020px] mx-auto px-4 sm:px-6 py-8">

        {/* Hero */}
        <div className="text-center pt-8 pb-10 px-4">
          <div className="inline-flex items-center gap-1.5 bg-[var(--color-hover)] border border-[var(--color-border-mid)]
            text-[var(--color-sub)] px-3.5 py-1.5 rounded-full text-[0.77rem] font-medium mb-5">
            <Shield size={13} className="text-[var(--color-sub)]" />
            NIST FIPS 204 Standard · Post-Quantum Identity Provider
          </div>

          <h1 className="text-[clamp(2rem,6vw,3rem)] font-bold leading-[1.15] tracking-tight mb-4
            bg-gradient-to-br from-slate-100 to-slate-400 bg-clip-text text-transparent">
            Post-Quantum OAuth 2.0<br />Identity &amp; Access Management
          </h1>

          <p className="text-[var(--color-sub)] text-[clamp(0.95rem,2.5vw,1.1rem)] max-w-xl mx-auto mb-8 leading-relaxed">
            DilithiAuth is a research-oriented Identity Provider (IdP) implementing the Authorization Code Flow
            with hybrid signatures combining classical Ed25519 with NIST-standardized{' '}
            <strong className="text-[var(--color-txt)]">ML-DSA-44</strong> (FIPS 204) digital signatures,
            mitigating "Harvest Now, Forge Later" (HNFL) threats.
          </p>

          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <Link
              to="/register"
              className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-lg text-[0.95rem] font-semibold
                bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-h)]
                hover:shadow-[0_0_0_3px_rgba(59,130,246,0.2)] transition-all duration-150 no-underline hover:no-underline"
            >
              Create Developer Account <ArrowRight size={18} />
            </Link>
            <Link
              to="/login"
              className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-lg text-[0.95rem] font-semibold
                bg-transparent border border-[var(--color-border)] text-[var(--color-txt)]
                hover:bg-[var(--color-hover)] hover:border-[var(--color-border-mid)]
                transition-all duration-150 no-underline hover:no-underline"
            >
              Sign In
            </Link>
          </div>
        </div>

        {/* Quantum threat callout */}
        <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 mb-6
          hover:border-[var(--color-border-mid)] hover:shadow-lg transition-all duration-200">
          <div className="flex gap-4 items-start">
            <div className="bg-[var(--color-hover)] p-2.5 rounded-lg text-[var(--color-sub)] shrink-0">
              <Lock size={22} />
            </div>
            <div>
              <h3 className="text-[1rem] font-semibold text-[var(--color-txt)] mb-1">
                Quantum-Resilient Hybrid Authorization
              </h3>
              <p className="text-[var(--color-sub)] text-sm leading-relaxed">
                Identity claims and access tokens intercepted today are vulnerable to future decryption by
                Cryptographically Relevant Quantum Computers (CRQCs). DilithiAuth pairs{' '}
                <strong className="text-[var(--color-txt)]">ML-DSA-44</strong> with classical signatures
                to maintain backward compatibility while ensuring long-term token security.
              </p>
            </div>
          </div>
        </div>

        {/* Feature grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
          {features.map(({ icon: Icon, label, desc }) => (
            <div
              key={label}
              className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5
                hover:border-[var(--color-border-mid)] hover:shadow-lg hover:-translate-y-0.5
                transition-all duration-200"
            >
              <div className="inline-flex items-center justify-center w-10 h-10 rounded-lg
                bg-blue-500/12 text-blue-400 mb-3">
                <Icon size={20} />
              </div>
              <h3 className="text-[var(--color-txt)] font-semibold text-sm mb-1.5">{label}</h3>
              <p className="text-[var(--color-sub)] text-[0.83rem] leading-relaxed">{desc}</p>
            </div>
          ))}
        </div>

        {/* Token playground */}
        <QuantumTokenPlayground />

        {/* OpenID endpoints */}
        <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 mt-6">
          <div className="flex items-center gap-2 mb-1.5">
            <Terminal size={18} className="text-[var(--color-sub)]" />
            <h2 className="text-[var(--color-txt)] font-semibold text-base">OpenID Connect Endpoints</h2>
          </div>
          <p className="text-[var(--color-sub)] text-[0.85rem] mb-3">
            Standard discovery document and JWKS public key distribution.
          </p>
          <pre className="code-block">{`GET /.well-known/openid-configuration
GET /.well-known/jwks.json           # Includes 'mldsa44_pub' public key
POST /token/thin?sub=user_123         # Issue Thin Token (sidecar reference)
POST /token/full?sub=user_123         # Issue Jumbo Token (embedded signature)`}</pre>
        </div>

        {/* Footer */}
        <div className="flex flex-wrap justify-between items-center gap-4 mt-10 pt-5
          border-t border-[var(--color-border)] text-[0.83rem] text-[var(--color-muted)]">
          <div>© 2026 DilithiAuth · Quantum-Safe Identity Provider</div>
          <div className="flex gap-5">
            <Link to="/privacy" className="text-[var(--color-muted)] hover:text-[var(--color-sub)] no-underline">Privacy Policy</Link>
            <Link to="/terms" className="text-[var(--color-muted)] hover:text-[var(--color-sub)] no-underline">Terms of Service</Link>
            <Link to="/docs" className="text-[var(--color-muted)] hover:text-[var(--color-sub)] no-underline">Documentation</Link>
          </div>
        </div>
      </div>
    </>
  )
}
