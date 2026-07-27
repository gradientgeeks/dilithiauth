import { useState } from 'react'
import { ShieldCheck, Cpu, Database, HardDrive, Copy, Check, FileCode, Layers, Info } from 'lucide-react'

type TokenType = 'thin' | 'full' | 'cwt'

export default function QuantumTokenPlayground() {
  const [tokenType, setTokenType] = useState<TokenType>('thin')
  const [copied, setCopied] = useState(false)
  const [subject, setSubject] = useState('alice@example.com')

  const sampleThinJWT = {
    header: {
      alg: "Ed25519+ML-DSA-44",
      typ: "JWT",
      kid: "dilithi-key-2026-01"
    },
    payload: {
      iss: "https://auth.example.com",
      sub: subject,
      aud: "dilithi-client-app",
      exp: 1774640000,
      iat: 1774636400,
      pqc_ref: "sha256:8f9a2b4c1e7d3f9e... (stored in sidecar)",
      pqc_alg: "ML-DSA-44"
    },
    signature_info: {
      ed25519_bytes: 64,
      mldsa44_bytes: 2420,
      mldsa44_location: "Sidecar Store (Redis/InMemory)",
      total_header_size: "~450 Bytes"
    }
  }

  const sampleJumboJWT = {
    header: {
      alg: "Ed25519+ML-DSA-44",
      typ: "JWT",
      kid: "dilithi-key-2026-01"
    },
    payload: {
      iss: "https://auth.example.com",
      sub: subject,
      aud: "dilithi-client-app",
      exp: 1774640000,
      iat: 1774636400,
      pqc_alg: "ML-DSA-44",
      pqc_sig_embedded: "0x416c696365... [2,420 bytes raw signature embedded directly]"
    },
    signature_info: {
      ed25519_bytes: 64,
      mldsa44_bytes: 2420,
      mldsa44_location: "Embedded in JWT Header/Payload",
      total_header_size: "~3.8 KB"
    }
  }

  const sampleCWT = {
    header: {
      1: -7, // EdDSA / Hybrid
      4: "dilithi-key-2026-01"
    },
    payload: {
      1: "https://auth.example.com",
      2: subject,
      3: "dilithi-client-app",
      4: 1774640000,
      5: 1774636400
    },
    signature_info: {
      format: "CBOR / RFC 8392",
      encoding: "Binary (Compact COSE_Sign1)",
      compression: "25-30% footprint reduction vs JSON",
      total_size: "~310 Bytes (Thin) / ~2.7 KB (Jumbo)"
    }
  }

  const activeData = tokenType === 'thin' ? sampleThinJWT : tokenType === 'full' ? sampleJumboJWT : sampleCWT

  const copySampleText = () => {
    navigator.clipboard.writeText(JSON.stringify(activeData, null, 2))
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="card card-glow" style={{ marginTop: '2rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem', marginBottom: '1.25rem' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Cpu size={18} color="#9ca3af" />
            <h2>Token Architecture Inspector</h2>
          </div>
          <p className="text-muted" style={{ fontSize: '0.88rem' }}>
            Explore how DilithiAuth constructs hybrid quantum-safe identity tokens (FIPS 204 ML-DSA-44).
          </p>
        </div>
        <span className="badge badge-hybrid">
          <ShieldCheck size={13} /> FIPS 204 Compliant
        </span>
      </div>

      {/* Sub controls */}
      <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap', marginBottom: '1.25rem' }}>
        <div style={{ flex: 1, minWidth: '220px' }}>
          <label className="label">Subject Claim (sub)</label>
          <input
            type="text"
            value={subject}
            onChange={e => setSubject(e.target.value || 'alice@example.com')}
            placeholder="User Subject"
            style={{ fontSize: '0.88rem', padding: '0.45rem 0.75rem' }}
          />
        </div>

        <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', marginTop: '1.25rem' }}>
          <button
            className={`btn btn-sm ${tokenType === 'thin' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setTokenType('thin')}
          >
            <Database size={14} /> Thin Token (Hash Ref)
          </button>
          <button
            className={`btn btn-sm ${tokenType === 'full' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setTokenType('full')}
          >
            <HardDrive size={14} /> Jumbo Token (~4 KB)
          </button>
          <button
            className={`btn btn-sm ${tokenType === 'cwt' ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => setTokenType('cwt')}
          >
            <FileCode size={14} /> CWT (CBOR RFC 8392)
          </button>
        </div>
      </div>

      {/* Visual Token Architecture Explanation */}
      <div className="alert alert-info" style={{ marginBottom: '1.25rem' }}>
        <Info size={18} />
        <div>
          {tokenType === 'thin' && (
            <span>
              <strong>Thin Token Strategy:</strong> Emits an Ed25519 signature + <code>pqc_ref</code> hash in the JWT. The full 2.4 KB ML-DSA-44 signature is cached in a sidecar store (e.g. Redis) to avoid header overflow in legacy reverse proxies.
            </span>
          )}
          {tokenType === 'full' && (
            <span>
              <strong>Jumbo Token Strategy:</strong> Embeds the complete ~2.4 KB ML-DSA-44 post-quantum signature directly inside the JWT payload. Requires IdP proxy <code>MaxHeaderBytes = 16 KiB</code> configuration.
            </span>
          )}
          {tokenType === 'cwt' && (
            <span>
              <strong>CBOR Web Token (CWT):</strong> Formatted as binary CBOR/COSE structures (RFC 8392). Reduces quantum token payload size by 25–30% compared to JSON string representation.
            </span>
          )}
        </div>
      </div>

      {/* Signature size comparison visualizer */}
      <div style={{ background: '#040711', padding: '1rem', borderRadius: '8px', border: '1px solid var(--border-color)', marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.8rem', color: 'var(--text-sub)', marginBottom: '0.4rem', fontWeight: 600 }}>
          <span>Signature Overhead Visualizer</span>
          <span>Ed25519 (64B) vs ML-DSA-44 (2,420B)</span>
        </div>
        <div style={{ height: '10px', background: 'rgba(255,255,255,0.06)', borderRadius: '5px', overflow: 'hidden', display: 'flex' }}>
          <div style={{ width: tokenType === 'full' ? '3%' : '20%', background: '#38bdf8', height: '100%' }} title="Ed25519 Signature (64 bytes)" />
          <div style={{ width: tokenType === 'full' ? '97%' : '80%', background: '#2563eb', height: '100%' }} title="ML-DSA-44 Quantum Signature (2,420 bytes)" />
        </div>
        <div style={{ display: 'flex', gap: '1.5rem', marginTop: '0.6rem', fontSize: '0.78rem', color: 'var(--text-sub)' }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#38bdf8' }} /> Classical Ed25519: 64 B
          </span>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.35rem' }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#2563eb' }} /> Post-Quantum ML-DSA-44: 2,420 B
          </span>
        </div>
      </div>

      {/* Code Inspector */}
      <div style={{ position: 'relative' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.4rem' }}>
          <span className="label" style={{ marginBottom: 0 }}>
            <Layers size={13} style={{ verticalAlign: '-2px', marginRight: '0.3rem' }} /> Token Structure Inspection
          </span>
          <button className="btn btn-outline btn-sm" onClick={copySampleText}>
            {copied ? <><Check size={13} /> Copied!</> : <><Copy size={13} /> Copy JSON</>}
          </button>
        </div>
        <pre className="code-block">
          {JSON.stringify(activeData, null, 2)}
        </pre>
      </div>
    </div>
  )
}
