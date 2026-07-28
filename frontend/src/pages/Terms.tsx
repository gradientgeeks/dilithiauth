import { Link } from 'react-router-dom'
import { Scale, ArrowLeft } from 'lucide-react'
import Navbar from '../components/Navbar'

export default function Terms() {
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
              <Scale size={28} />
            </div>
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[var(--color-txt)]">Terms of Service</h1>
              <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">Effective Date: July 27, 2026</p>
            </div>
          </div>

          <div className="space-y-6 text-sm text-[var(--color-sub)] leading-relaxed">
            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">1. Acceptance of Terms</h2>
              <p>
                By accessing or using the DilithiAuth platform, APIs, developer dashboard, or client SDKs, you agree to be bound by these Terms of Service.
                If you do not agree to these terms, you may not access or use the service.
              </p>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">2. Identity Services &amp; API Usage</h2>
              <p>
                DilithiAuth provides a quantum-safe Identity Provider (IdP) implementing OAuth 2.0 and OpenID Connect (OIDC) with hybrid classical (Ed25519) and post-quantum (ML-DSA-44) signatures.
                Developers and client applications must:
              </p>
              <ul className="list-disc list-inside space-y-1.5 pl-2">
                <li>Maintain the confidentiality of issued Client Secrets and API credentials.</li>
                <li>Implement OAuth 2.0 PKCE (S256) for public applications.</li>
                <li>Refrain from attempting to bypass rate limits, probe infrastructure vulnerabilities, or issue unauthorized tokens.</li>
              </ul>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">3. Service Level &amp; Modifications</h2>
              <p>
                DilithiAuth strives to maintain high service availability. We reserve the right to modify, update, or deprecate API endpoints with appropriate notice to registered developers.
              </p>
            </section>

            <section className="space-y-2">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">4. Limitation of Liability</h2>
              <p>
                DilithiAuth is provided "as is" without warranty of any kind. In no event shall DilithiAuth or its maintainers be liable for any indirect, incidental, or consequential damages arising out of API availability or identity service usage.
              </p>
            </section>

            <section className="space-y-3">
              <h2 className="text-base font-semibold text-[var(--color-txt)]">5. Contact Information</h2>
              <p>For inquiries regarding these Terms of Service, contact us at:</p>
              <div className="p-4 rounded-lg bg-[#050810] border border-[var(--color-border)] text-xs font-mono space-y-1 text-slate-300">
                <div className="font-bold text-[var(--color-txt)]">DilithiAuth Legal &amp; Compliance Desk</div>
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
