import { Link } from 'react-router-dom'
import Navbar from '../components/Navbar'

export default function Landing() {
  return (
    <>
      <Navbar />
      <div className="hero-section">
        <h1>Post-Quantum OAuth 2.0<br />Identity Provider</h1>
        <p>
          DilithiAuth signs tokens with both Ed25519 and ML-DSA-44 (NIST FIPS 204),
          protecting your users against "harvest now, decrypt later" attacks.
        </p>
        <div className="hero-actions">
          <Link to="/register" className="btn btn-primary" style={{ fontSize: '1rem', padding: '0.7rem 1.6rem' }}>
            Get started — it's free
          </Link>
          <Link to="/login" className="btn btn-outline" style={{ fontSize: '1rem', padding: '0.7rem 1.6rem' }}>
            Sign in
          </Link>
        </div>
      </div>

      <div className="features">
        <div className="feature-card">
          <div className="feature-icon">🔐</div>
          <h3>Hybrid Signatures</h3>
          <p>Every token is signed with Ed25519 <em>and</em> ML-DSA-44. Quantum-safe from day one.</p>
        </div>
        <div className="feature-card">
          <div className="feature-icon">⚡</div>
          <h3>Standard OAuth 2.0</h3>
          <p>Drop-in replacement for any OAuth 2.0 provider. Authorization Code + PKCE flow.</p>
        </div>
        <div className="feature-card">
          <div className="feature-icon">🛠️</div>
          <h3>Easy Integration</h3>
          <p>Register your app, get a client ID and secret, and redirect users to our authorization endpoint.</p>
        </div>
        <div className="feature-card">
          <div className="feature-icon">🏢</div>
          <h3>Multi-Tenant</h3>
          <p>Scope users by subdomain or header. One DilithiAuth instance, unlimited tenants.</p>
        </div>
      </div>
    </>
  )
}
