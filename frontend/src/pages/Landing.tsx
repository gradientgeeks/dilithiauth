import { Link } from 'react-router-dom'
import { ShieldCheck, Zap, Wrench, Building2, Lock, ArrowRight, Terminal, Shield } from 'lucide-react'
import Navbar from '../components/Navbar'
import QuantumTokenPlayground from '../components/QuantumTokenPlayground'

export default function Landing() {
  return (
    <>
      <Navbar />
      <div className="container">
        {/* Hero Section */}
        <div className="hero-section">
          <div className="hero-badge">
            <Shield size={13} color="#9ca3af" /> NIST FIPS 204 Standard • Post-Quantum Identity Provider
          </div>
          <h1>
            Post-Quantum OAuth 2.0 <br />
            Identity & Access Management
          </h1>
          <p>
            DilithiAuth issues hybrid tokens combining classical Ed25519 with NIST-standardized 
            <strong> ML-DSA-44</strong> signatures, protecting authentication infrastructure against 
            "Harvest Now, Decrypt Later" (HNDL) quantum threats.
          </p>
          <div className="hero-actions">
            <Link to="/register" className="btn btn-primary" style={{ padding: '0.65rem 1.4rem' }}>
              Create Developer Account <ArrowRight size={16} />
            </Link>
            <Link to="/login" className="btn btn-outline" style={{ padding: '0.65rem 1.4rem' }}>
              Sign In
            </Link>
          </div>
        </div>

        {/* Technical Architecture Overview */}
        <div className="card" style={{ marginBottom: '2rem' }}>
          <div style={{ display: 'flex', gap: '1rem', alignItems: 'flex-start' }}>
            <div style={{ background: '#1f2937', padding: '0.65rem', borderRadius: '6px', color: '#9ca3af', flexShrink: 0 }}>
              <Lock size={22} />
            </div>
            <div>
              <h3 style={{ fontSize: '1rem', marginBottom: '0.25rem' }}>
                Quantum-Resilient Hybrid Authorization
              </h3>
              <p style={{ color: 'var(--text-sub)', fontSize: '0.875rem', lineHeight: '1.6' }}>
                Identity claims and access tokens intercepted today are vulnerable to future decryption by Cryptographically Relevant Quantum Computers (CRQCs). DilithiAuth pairs <strong>ML-DSA-44</strong> with classical signatures to maintain backward compatibility while ensuring long-term token security.
              </p>
            </div>
          </div>
        </div>

        {/* Feature Cards Grid */}
        <div className="features">
          <div className="feature-card">
            <div className="feature-icon"><ShieldCheck size={20} /></div>
            <h3>Hybrid Signatures</h3>
            <p>Every token is dual-signed using Ed25519 classical keys and ML-DSA-44 post-quantum keypairs.</p>
          </div>

          <div className="feature-card">
            <div className="feature-icon"><Zap size={20} /></div>
            <h3>OAuth 2.0 & PKCE</h3>
            <p>Standard authorization code flow compatible with modern web, mobile, and confidential backends.</p>
          </div>

          <div className="feature-card">
            <div className="feature-icon"><Wrench size={20} /></div>
            <h3>Optimized Token Modes</h3>
            <p>Thin JWTs (Redis sidecar reference), Jumbo JWTs (embedded signatures), and CBOR Web Tokens (CWT).</p>
          </div>

          <div className="feature-card">
            <div className="feature-icon"><Building2 size={20} /></div>
            <h3>Multi-Tenancy</h3>
            <p>Partition user pools and client applications by tenant subdomain or header isolation.</p>
          </div>
        </div>

        {/* Interactive Quantum Token Playground */}
        <QuantumTokenPlayground />

        {/* OpenID Discovery Endpoints */}
        <div className="card" style={{ marginTop: '2rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.5rem' }}>
            <Terminal size={18} color="#9ca3af" />
            <h2>OpenID Connect Endpoints</h2>
          </div>
          <p className="text-muted" style={{ fontSize: '0.85rem', marginBottom: '0.85rem' }}>
            Standard discovery document and JWKS public key distribution.
          </p>
          <div className="code-block">
{`GET /.well-known/openid-configuration
GET /.well-known/jwks.json           # Includes 'mldsa44_pub' public key
POST /token/thin?sub=user_123         # Issue Thin Token (sidecar reference)
POST /token/full?sub=user_123         # Issue Jumbo Token (embedded signature)`}
          </div>
        </div>
      </div>
    </>
  )
}
