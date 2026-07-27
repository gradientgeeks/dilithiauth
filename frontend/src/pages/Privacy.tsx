import { Link } from 'react-router-dom'
import { ShieldCheck, ArrowLeft } from 'lucide-react'
import Navbar from '../components/Navbar'

export default function Privacy() {
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
              <ShieldCheck size={28} />
            </div>
            <div>
              <h1 style={{ fontSize: '1.75rem', margin: 0 }}>Privacy Policy</h1>
              <p className="text-muted" style={{ margin: 0, fontSize: '0.85rem' }}>Effective Date: July 27, 2026</p>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.75rem', color: 'var(--text-main)', lineHeight: '1.6', fontSize: '0.95rem' }}>
            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>1. Overview & Commitment</h2>
              <p>
                DilithiAuth ("we", "our", or "us") is dedicated to protecting user privacy through quantum-resistant identity security.
                This Privacy Policy explains how we collect, use, disclose, and safeguard user information when utilizing our Post-Quantum Identity Provider (IdP) platform, APIs, and client SDKs.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>2. Information We Collect</h2>
              <ul style={{ paddingLeft: '1.2rem', margin: '0.5rem 0' }}>
                <li><strong>Account Identity Information:</strong> Email addresses, user IDs (`sub`), and hashed credentials provided during registration or social OAuth federation.</li>
                <li><strong>Cryptographic & Token Telemetry:</strong> Public Ed25519 and NIST FIPS 204 ML-DSA-44 post-quantum key identifiers, signature references (`pqc_ref`), and session tokens.</li>
                <li><strong>Technical Logs:</strong> IP addresses, request user agents, and authorization code exchange timestamps used strictly for rate limiting, threat prevention, and audit compliance.</li>
              </ul>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>3. How We Use Information</h2>
              <p>We process collected data exclusively for:</p>
              <ul style={{ paddingLeft: '1.2rem', margin: '0.5rem 0' }}>
                <li>Authenticating end-users and issuing hybrid post-quantum signed JWT / CWT tokens.</li>
                <li>Providing OAuth 2.0 / OpenID Connect authorization code flow services to registered client applications.</li>
                <li>Preventing credential stuffing, brute-force attacks, and unauthorized token reuse.</li>
                <li>Fulfilling legal and security audit obligations.</li>
              </ul>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>4. Post-Quantum Cryptographic Data Protection</h2>
              <p>
                DilithiAuth utilizes <strong>NIST FIPS 204 ML-DSA-44</strong> post-quantum signatures combined with classical <strong>Ed25519</strong> signing.
                All sensitive credentials, secrets, and refresh tokens are encrypted at rest using AES-256-GCM and stored in high-security sidecar stores.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>5. Third-Party Identity Providers</h2>
              <p>
                When you choose to sign in using Google or GitHub OAuth 2.0 social connections, we only request basic profile scope (`openid`, `profile`, `email`).
                We never sell, monetize, or share your identity data with third-party advertisers.
              </p>
            </section>

            <section>
              <h2 style={{ fontSize: '1.2rem', marginBottom: '0.5rem', color: '#fff' }}>6. Contact & Data Subject Rights</h2>
              <p>
                You have the right to request deletion of your account and identity logs at any time. For privacy inquiries or data removal requests, contact our security team at:
              </p>
              <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border-color)', padding: '1rem', borderRadius: '6px', fontSize: '0.9rem' }}>
                <strong>DilithiAuth Security & Privacy Desk</strong><br />
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
