import { Link } from 'react-router-dom'
import { Scale, ArrowLeft } from 'lucide-react'
import Navbar from '../components/Navbar'

export default function Terms() {
  return (
    <>
      <Navbar />
      <div className="container" style={{ maxWidth: '840px', paddingTop: '2rem', paddingBottom: '4rem' }}>
        <Link to="/" className="btn btn-outline btn-sm" style={{ marginBottom: '1.5rem', display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
          <ArrowLeft size={14} /> Back to Home
        </Link>

        <div className="card" style={{ padding: '2.5rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1.5rem' }}>
            <div style={{ background: 'rgba(37, 99, 235, 0.1)', padding: '0.6rem', borderRadius: '8px', color: 'var(--primary)' }}>
              <Scale size={28} />
            </div>
            <div>
              <h1 style={{ fontSize: '1.75rem', margin: 0 }}>Terms of Service</h1>
              <p className="text-muted" style={{ margin: 0, fontSize: '0.85rem' }}>Effective Date: July 27, 2026</p>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.75rem', color: 'var(--text-main)', lineHeight: '1.6', fontSize: '0.95rem' }}>
            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>1. Acceptance of Terms</h2>
              <p>
                By accessing or using the DilithiAuth platform, APIs, developer dashboard, or client SDKs, you agree to be bound by these Terms of Service.
                If you do not agree to these terms, you may not access or use the service.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>2. Identity Services & API Usage</h2>
              <p>
                DilithiAuth provides a quantum-safe Identity Provider (IdP) implementing OAuth 2.0 and OpenID Connect (OIDC) with hybrid classical (Ed25519) and post-quantum (ML-DSA-44) signatures.
                Developers and client applications must:
              </p>
              <ul style={{ paddingLeft: '1.2rem', margin: '0.5rem 0' }}>
                <li>Maintain the confidentiality of issued Client Secrets and API credentials.</li>
                <li>Implement OAuth 2.0 PKCE (S256) for public applications.</li>
                <li>Refrain from attempting to bypass rate limits, probe infrastructure vulnerabilities, or issue unauthorized tokens.</li>
              </ul>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>3. Service Level & Modifications</h2>
              <p>
                DilithiAuth strives to maintain high service availability. We reserve the right to modify, update, or deprecate API endpoints with appropriate notice to registered developers.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>4. Limitation of Liability</h2>
              <p>
                DilithiAuth is provided "as is" without warranty of any kind. In no event shall DilithiAuth or its maintainers be liable for any indirect, incidental, or consequential damages arising out of API availability or identity service usage.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>5. Contact Information</h2>
              <p>For inquiries regarding these Terms of Service, contact us at:</p>
              <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border-color)', padding: '1rem', borderRadius: '6px', fontSize: '0.9rem' }}>
                <strong>DilithiAuth Legal & Compliance Desk</strong><br />
                Email: <a href="mailto:support@gradientgeeks.tech" style={{ color: 'var(--primary)' }}>support@gradientgeeks.tech</a><br />
                Domain: <code>https://dilithiauthapis.gradientgeeks.tech</code>
              </div>
            </section>
          </div>
        </div>
      </div>
    </>
  )
}
