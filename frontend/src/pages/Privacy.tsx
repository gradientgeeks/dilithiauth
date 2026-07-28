import { Link } from 'react-router-dom'
import { ShieldCheck, ArrowLeft } from 'lucide-react'
import Navbar from '../components/Navbar'

export default function Privacy() {
  return (
    <div className="min-h-screen bg-[var(--color-bg)] text-[var(--color-txt)]">
      <Navbar />
      <div className="w-full max-w-[840px] mx-auto px-4 sm:px-6 py-8 pb-16">
        <div className="mb-6">
          <Link
            to="/"
            className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md
              border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
              hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors"
          >
            <ArrowLeft size={14} /> Back to Home
          </Link>
        </div>

        <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 sm:p-10 space-y-6">
          <div className="flex items-center gap-3 pb-6 border-b border-[var(--color-border)]">
            <div className="p-3 rounded-xl bg-blue-500/15 border border-blue-500/30 text-blue-400">
              <ShieldCheck size={28} />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[var(--color-txt)]">Privacy Policy</h1>
              <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">Effective Date: July 27, 2026</p>
            </div>
          </div>

          <div className="space-y-6 text-sm text-[var(--color-sub)] leading-relaxed">
            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">1. Overview &amp; Commitment</h2>
              <p>
                DilithiAuth ("we", "our", or "us") is dedicated to protecting user privacy through quantum-resistant identity security.
                This Privacy Policy explains how we collect, use, disclose, and safeguard user information when utilizing our Post-Quantum Identity Provider (IdP) platform, APIs, and client SDKs.
              </p>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">2. Information We Collect</h2>
              <ul className="list-disc list-inside space-y-1.5 pl-2">
                <li><strong className="text-[var(--color-txt)]">Account Identity Information:</strong> Email addresses, user IDs (<code className="text-xs bg-[#050810] px-1.5 py-0.5 rounded text-sky-300">sub</code>), and hashed credentials provided during registration or social OAuth federation.</li>
                <li><strong className="text-[var(--color-txt)]">Cryptographic &amp; Token Telemetry:</strong> Public Ed25519 and NIST FIPS 204 ML-DSA-44 post-quantum key identifiers, signature references (<code className="text-xs bg-[#050810] px-1.5 py-0.5 rounded text-sky-300">pqc_ref</code>), and session tokens.</li>
                <li><strong className="text-[var(--color-txt)]">Technical Logs:</strong> IP addresses, request user agents, and authorization code exchange timestamps used strictly for rate limiting, threat prevention, and audit compliance.</li>
              </ul>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">3. How We Use Information</h2>
              <p>We process collected data exclusively for:</p>
              <ul className="list-disc list-inside space-y-1.5 pl-2">
                <li>Authenticating end-users and issuing hybrid post-quantum signed JWT / CWT tokens.</li>
                <li>Providing OAuth 2.0 / OpenID Connect authorization code flow services to registered client applications.</li>
                <li>Preventing credential stuffing, brute-force attacks, and unauthorized token reuse.</li>
                <li>Fulfilling legal and security audit obligations.</li>
              </ul>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">4. Post-Quantum Cryptographic Data Protection</h2>
              <p>
                DilithiAuth utilizes <strong className="text-[var(--color-txt)]">NIST FIPS 204 ML-DSA-44</strong> post-quantum signatures combined with classical <strong className="text-[var(--color-txt)]">Ed25519</strong> signing.
                All sensitive credentials, secrets, and refresh tokens are encrypted at rest using AES-256-GCM and stored in high-security sidecar stores.
              </p>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">5. Third-Party Identity Providers</h2>
              <p>
                When you choose to sign in using Google or GitHub OAuth 2.0 social connections, we only request basic profile scope (<code className="text-xs bg-[#050810] px-1.5 py-0.5 rounded text-sky-300">openid</code>, <code className="text-xs bg-[#050810] px-1.5 py-0.5 rounded text-sky-300">profile</code>, <code className="text-xs bg-[#050810] px-1.5 py-0.5 rounded text-sky-300">email</code>).
                We never sell, monetize, or share your identity data with third-party advertisers.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">6. Contact &amp; Data Subject Rights</h2>
              <p>
                You have the right to request deletion of your account and identity logs at any time. For privacy inquiries or data removal requests, contact our security team at:
              </p>
              <div className="p-4 rounded-lg bg-[#050810] border border-[var(--color-border)] text-xs font-mono space-y-1 text-slate-300">
                <div className="font-bold text-[var(--color-txt)]">DilithiAuth Security &amp; Privacy Desk</div>
                <div>Email: <a href="mailto:support@gradientgeeks.tech" className="text-blue-400 hover:text-blue-300">support@gradientgeeks.tech</a></div>
                <div>Domain: <code className="text-sky-300">https://dilithiauthapis.gradientgeeks.tech</code></div>
              </div>
            </section>
          </div>
        </div>
      </div>
    </div>
  )
}
