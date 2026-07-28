import { useState } from 'react'
import { ShieldCheck, Cpu, Database, HardDrive, Copy, Check, FileCode, Layers, Info } from 'lucide-react'

type TokenType = 'thin' | 'full' | 'cwt'

export default function QuantumTokenPlayground() {
  const [tokenType, setTokenType] = useState<TokenType>('thin')
  const [copied, setCopied] = useState(false)
  const [subject, setSubject] = useState('alice@example.com')

  const sampleThinJWT = {
    header: { alg: "Ed25519+ML-DSA-44", typ: "JWT", kid: "dilithi-key-2026-01" },
    payload: {
      iss: "https://auth.example.com", sub: subject, aud: "dilithi-client-app",
      exp: 1774640000, iat: 1774636400,
      pqc_ref: "sha256:8f9a2b4c1e7d3f9e... (stored in sidecar)",
      pqc_alg: "ML-DSA-44"
    },
    signature_info: {
      ed25519_bytes: 64, mldsa44_bytes: 2420,
      mldsa44_location: "Sidecar Store (Redis/InMemory)",
      total_header_size: "~450 Bytes"
    }
  }

  const sampleJumboJWT = {
    header: { alg: "Ed25519+ML-DSA-44", typ: "JWT", kid: "dilithi-key-2026-01" },
    payload: {
      iss: "https://auth.example.com", sub: subject, aud: "dilithi-client-app",
      exp: 1774640000, iat: 1774636400, pqc_alg: "ML-DSA-44",
      pqc_sig_embedded: "0x416c696365... [2,420 bytes raw signature embedded directly]"
    },
    signature_info: {
      ed25519_bytes: 64, mldsa44_bytes: 2420,
      mldsa44_location: "Embedded in JWT Header/Payload",
      total_header_size: "~3.8 KB"
    }
  }

  const sampleCWT = {
    header: { 1: -7, 4: "dilithi-key-2026-01" },
    payload: { 1: "https://auth.example.com", 2: subject, 3: "dilithi-client-app", 4: 1774640000, 5: 1774636400 },
    signature_info: {
      format: "CBOR / RFC 8392", encoding: "Binary (Compact COSE_Sign1)",
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

  const TabBtn = ({ type, icon: Icon, label }: { type: TokenType; icon: React.ElementType; label: string }) => (
    <button
      onClick={() => setTokenType(type)}
      className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold
        transition-all duration-150 cursor-pointer border
        ${tokenType === type
          ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
          : 'bg-transparent text-[var(--color-sub)] border-[var(--color-border)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)]'
        }`}
    >
      <Icon size={14} /> {label}
    </button>
  )

  return (
    <div className="bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-5 mt-6">
      {/* Header */}
      <div className="flex flex-wrap items-start justify-between gap-3 mb-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Cpu size={18} className="text-[var(--color-sub)]" />
            <h2 className="text-[var(--color-txt)] font-semibold text-base">Token Architecture Inspector</h2>
          </div>
          <p className="text-[var(--color-sub)] text-[0.87rem]">
            Explore how DilithiAuth constructs hybrid quantum-safe identity tokens (FIPS 204 ML-DSA-44).
          </p>
        </div>
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-[#1e1b4b] border border-[#3730a3]
          text-violet-300 text-[0.68rem] font-bold uppercase tracking-wide">
          <ShieldCheck size={13} /> FIPS 204 Compliant
        </span>
      </div>

      {/* Subject + type controls */}
      <div className="flex flex-wrap gap-4 items-end mb-4">
        <div className="flex-1 min-w-[220px]">
          <label className="block text-[0.72rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-1.5">
            Subject Claim (sub)
          </label>
          <input
            type="text"
            value={subject}
            onChange={e => setSubject(e.target.value || 'alice@example.com')}
            placeholder="User Subject"
            className="w-full px-3 py-2 text-sm bg-[var(--color-input)] border border-[var(--color-border)]
              rounded-md text-[var(--color-txt)] placeholder:text-[var(--color-muted)]
              focus:outline-none focus:border-[var(--color-focus)] focus:shadow-[0_0_0_3px_rgba(59,130,246,0.15)]
              hover:border-[var(--color-border-mid)] transition-all duration-150"
          />
        </div>
        <div className="flex flex-wrap gap-1.5">
          <TabBtn type="thin" icon={Database} label="Thin Token (Hash Ref)" />
          <TabBtn type="full" icon={HardDrive} label="Jumbo Token (~4 KB)" />
          <TabBtn type="cwt" icon={FileCode} label="CWT (CBOR RFC 8392)" />
        </div>
      </div>

      {/* Info alert */}
      <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4
        bg-blue-500/10 border border-blue-500/25 text-blue-300 text-sm leading-relaxed">
        <Info size={18} className="shrink-0 mt-0.5" />
        <div>
          {tokenType === 'thin' && (
            <span><strong>Thin Token Strategy:</strong> Emits an Ed25519 signature + <code className="text-xs bg-white/10 px-1 py-0.5 rounded">pqc_ref</code> hash in the JWT. The full 2.4 KB ML-DSA-44 signature is cached in a sidecar store (e.g. Redis) to avoid header overflow in legacy reverse proxies.</span>
          )}
          {tokenType === 'full' && (
            <span><strong>Jumbo Token Strategy:</strong> Embeds the complete ~2.4 KB ML-DSA-44 post-quantum signature directly inside the JWT payload. Requires IdP proxy <code className="text-xs bg-white/10 px-1 py-0.5 rounded">MaxHeaderBytes = 16 KiB</code> configuration.</span>
          )}
          {tokenType === 'cwt' && (
            <span><strong>CBOR Web Token (CWT):</strong> Formatted as binary CBOR/COSE structures (RFC 8392). Reduces quantum token payload size by 25–30% compared to JSON string representation.</span>
          )}
        </div>
      </div>

      {/* Signature visualizer */}
      <div className="bg-[#040711] p-4 rounded-lg border border-[var(--color-border)] mb-4">
        <div className="flex justify-between text-xs text-[var(--color-sub)] font-semibold mb-2">
          <span>Signature Overhead Visualizer</span>
          <span>Ed25519 (64B) vs ML-DSA-44 (2,420B)</span>
        </div>
        <div className="h-2.5 bg-white/5 rounded-full overflow-hidden flex">
          <div
            className="bg-sky-400 h-full transition-all duration-500"
            style={{ width: tokenType === 'full' ? '3%' : '20%' }}
            title="Ed25519 Signature (64 bytes)"
          />
          <div
            className="bg-blue-600 h-full transition-all duration-500"
            style={{ width: tokenType === 'full' ? '97%' : '80%' }}
            title="ML-DSA-44 Quantum Signature (2,420 bytes)"
          />
        </div>
        <div className="flex gap-5 mt-2 text-xs text-[var(--color-sub)]">
          <span className="inline-flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-sky-400 inline-block" /> Classical Ed25519: 64 B
          </span>
          <span className="inline-flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-blue-600 inline-block" /> Post-Quantum ML-DSA-44: 2,420 B
          </span>
        </div>
      </div>

      {/* Code inspector */}
      <div>
        <div className="flex justify-between items-center mb-2">
          <span className="text-[0.72rem] font-bold text-[var(--color-sub)] uppercase tracking-wider flex items-center gap-1.5">
            <Layers size={13} /> Token Structure Inspection
          </span>
          <button
            onClick={copySampleText}
            className="inline-flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-semibold rounded-md
              border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent
              hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-all duration-150 cursor-pointer"
          >
            {copied ? <><Check size={13} /> Copied!</> : <><Copy size={13} /> Copy JSON</>}
          </button>
        </div>
        <pre className="code-block">{JSON.stringify(activeData, null, 2)}</pre>
      </div>
    </div>
  )
}
