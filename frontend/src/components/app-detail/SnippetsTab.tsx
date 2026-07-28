import { useState } from 'react'
import { Check, Copy, Terminal, FileCode, Code, Info, Layers } from 'lucide-react'
import type { OAuthApp } from '../../api'

export interface SnippetsTabProps {
  app: OAuthApp
}

type LangType = 'curl' | 'javascript' | 'go' | 'python'
type EndpointType = 'thin' | 'full' | 'cwt'

export default function SnippetsTab({ app }: SnippetsTabProps) {
  const [activeLang, setActiveLang] = useState<LangType>('curl')
  const [activeEndpoint, setActiveEndpoint] = useState<EndpointType>('thin')
  const [copied, setCopied] = useState(false)

  const redirectUri = app.RedirectURIs?.[0] || 'https://myapp.example.com/callback'
  const domain = 'https://dilithiauthapis.gradientgeeks.tech'

  const getSnippets = (): Record<LangType, string> => {
    const secretPlaceholder = app.IsPublic ? '' : 'YOUR_CLIENT_SECRET'

    if (activeEndpoint === 'thin') {
      return {
        curl: `# Step 1: Redirect user for authorization:
# ${domain}/authorize?client_id=${app.ClientID}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile

# Step 2: Request Thin JWT Token (Ed25519 + sidecar pqc_ref hash):
curl -X POST ${domain}/token/thin \\
  -u "${app.ClientID}:${secretPlaceholder}" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${encodeURIComponent(redirectUri)}"`,

        javascript: `// Request Thin JWT Token (Ed25519 signature + PQC sidecar hash reference)
const response = await fetch('${domain}/token/thin', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
    'Authorization': 'Basic ' + btoa('${app.ClientID}:${secretPlaceholder}')
  },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: 'AUTHORIZATION_CODE',
    redirect_uri: '${redirectUri}'
  })
});

const data = await response.json();
console.log('Thin JWT Access Token:', data.access_token);
console.log('PQC Reference Hash:', data.pqc_ref);`,

        go: `package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

func main() {
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("code", "AUTHORIZATION_CODE")
    data.Set("redirect_uri", "${redirectUri}")

    req, _ := http.NewRequestWithContext(context.Background(), "POST", "${domain}/token/thin", strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.SetBasicAuth("${app.ClientID}", "${secretPlaceholder}")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    fmt.Println("Thin Token Response:", string(body))
}`,

        python: `import requests

url = "${domain}/token/thin"
data = {
    "grant_type": "authorization_code",
    "code": "AUTHORIZATION_CODE",
    "redirect_uri": "${redirectUri}"
}
auth = ("${app.ClientID}", "${secretPlaceholder}")

response = requests.post(url, data=data, auth=auth)
token_data = response.json()
print("Thin Token:", token_data.get("access_token"))
print("PQC Ref:", token_data.get("pqc_ref"))`
      }
    }

    if (activeEndpoint === 'full') {
      return {
        curl: `# Step 1: Redirect user for authorization:
# ${domain}/authorize?client_id=${app.ClientID}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile

# Step 2: Request Jumbo JWT Token (embeds ~2.4 KB ML-DSA-44 quantum signature):
curl -X POST ${domain}/token/full \\
  -u "${app.ClientID}:${secretPlaceholder}" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${encodeURIComponent(redirectUri)}"`,

        javascript: `// Request Jumbo JWT Token (includes inline NIST FIPS 204 ML-DSA-44 signature)
const response = await fetch('${domain}/token/full', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
    'Authorization': 'Basic ' + btoa('${app.ClientID}:${secretPlaceholder}')
  },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: 'AUTHORIZATION_CODE',
    redirect_uri: '${redirectUri}'
  })
});

const data = await response.json();
// Jumbo token size is ~4 KB due to full ML-DSA-44 signature
console.log('Jumbo Post-Quantum JWT:', data.access_token);`,

        go: `package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

func main() {
    // Request full Jumbo JWT with embedded ML-DSA-44 post-quantum signature
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("code", "AUTHORIZATION_CODE")
    data.Set("redirect_uri", "${redirectUri}")

    req, _ := http.NewRequestWithContext(context.Background(), "POST", "${domain}/token/full", strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.SetBasicAuth("${app.ClientID}", "${secretPlaceholder}")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    fmt.Println("Jumbo Token Response:", string(body))
}`,

        python: `import requests

# Fetch Jumbo JWT with full embedded NIST FIPS 204 ML-DSA-44 signature (~4KB)
url = "${domain}/token/full"
data = {
    "grant_type": "authorization_code",
    "code": "AUTHORIZATION_CODE",
    "redirect_uri": "${redirectUri}"
}

response = requests.post(url, data=data, auth=("${app.ClientID}", "${secretPlaceholder}"))
print("Jumbo JWT:", response.json().get("access_token"))`
      }
    }

    // CWT endpoint
    return {
      curl: `# Step 1: Redirect user for authorization:
# ${domain}/authorize?client_id=${app.ClientID}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile

# Step 2: Request Binary CBOR Web Token (CWT, RFC 8392 - 20-30% smaller):
curl -X POST ${domain}/token/cwt \\
  -u "${app.ClientID}:${secretPlaceholder}" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${encodeURIComponent(redirectUri)}" \\
  --output token.cwt`,

      javascript: `// Request Binary CBOR Web Token (CWT RFC 8392)
const response = await fetch('${domain}/token/cwt', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
    'Authorization': 'Basic ' + btoa('${app.ClientID}:${secretPlaceholder}')
  },
  body: new URLSearchParams({
    grant_type: 'authorization_code',
    code: 'AUTHORIZATION_CODE',
    redirect_uri: '${redirectUri}'
  })
});

const cwtBuffer = await response.arrayBuffer();
console.log('Received CWT binary byte length:', cwtBuffer.byteLength);`,

      go: `package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

func main() {
    // Request Binary CBOR Web Token (CWT)
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("code", "AUTHORIZATION_CODE")
    data.Set("redirect_uri", "${redirectUri}")

    req, _ := http.NewRequestWithContext(context.Background(), "POST", "${domain}/token/cwt", strings.NewReader(data.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.SetBasicAuth("${app.ClientID}", "${secretPlaceholder}")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
    cwtBytes, _ := io.ReadAll(resp.Body)
    fmt.Printf("Received %d bytes CBOR Web Token\\n", len(cwtBytes))
}`,

      python: `import requests

# Request binary CBOR Web Token (CWT, 20-30% smaller payload)
url = "${domain}/token/cwt"
data = {
    "grant_type": "authorization_code",
    "code": "AUTHORIZATION_CODE",
    "redirect_uri": "${redirectUri}"
}

response = requests.post(url, data=data, auth=("${app.ClientID}", "${secretPlaceholder}"))
cwt_bytes = response.content
print(f"Received CWT payload ({len(cwt_bytes)} bytes)")`
    }
  }

  const snippets = getSnippets()
  const currentSnippet = snippets[activeLang]

  const handleCopy = () => {
    navigator.clipboard.writeText(currentSnippet)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const languages: { key: LangType; label: string; icon: React.ElementType }[] = [
    { key: 'curl', label: 'cURL', icon: Terminal },
    { key: 'javascript', label: 'JS / Fetch', icon: FileCode },
    { key: 'go', label: 'Go', icon: Code },
    { key: 'python', label: 'Python', icon: FileCode },
  ]

  const endpoints: { key: EndpointType; label: string; desc: string }[] = [
    { key: 'thin', label: '/token/thin', desc: 'Thin JWT + PQC sidecar hash ref' },
    { key: 'full', label: '/token/full', desc: 'Jumbo JWT (~4 KB ML-DSA-44 signature)' },
    { key: 'cwt', label: '/token/cwt', desc: 'CBOR Web Token (Binary RFC 8392)' },
  ]

  return (
    <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 space-y-6">
      <div>
        <h2 className="text-[var(--color-txt)] font-semibold text-lg mb-1">
          OAuth 2.0 / DilithiAuth Integration Snippets
        </h2>
        <p className="text-[var(--color-sub)] text-sm">
          Select a token strategy endpoint and your language to generate pre-configured authentication code.
        </p>
      </div>

      {/* Endpoint Strategy Switcher */}
      <div>
        <label className="block text-xs font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2 flex items-center gap-1.5">
          <Layers size={14} className="text-cyan-400" /> Token Endpoint Strategy
        </label>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {endpoints.map(({ key, label, desc }) => (
            <button
              key={key}
              type="button"
              onClick={() => setActiveEndpoint(key)}
              className={`p-3 rounded-lg border text-left transition-all cursor-pointer ${
                activeEndpoint === key
                  ? 'bg-cyan-500/10 border-cyan-500/50 text-[var(--color-txt)] ring-1 ring-cyan-500/30'
                  : 'bg-[#050810] border-[var(--color-border)] text-[var(--color-sub)] hover:border-[var(--color-border-mid)] hover:text-[var(--color-txt)]'
              }`}
            >
              <div className="font-mono font-bold text-xs text-cyan-400 mb-0.5">{label}</div>
              <div className="text-xs text-[var(--color-muted)]">{desc}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Language Selector Bar & Copy Action */}
      <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
        <div className="flex flex-wrap gap-1.5">
          {languages.map(({ key, label, icon: Icon }) => (
            <button
              key={key}
              type="button"
              onClick={() => setActiveLang(key)}
              className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all cursor-pointer ${
                activeLang === key
                  ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)] shadow-md shadow-blue-500/20'
                  : 'bg-transparent border-[var(--color-border)] text-[var(--color-sub)] hover:bg-white/[0.04] hover:text-[var(--color-txt)]'
              }`}
            >
              <Icon size={14} /> {label}
            </button>
          ))}
        </div>

        <button
          type="button"
          onClick={handleCopy}
          className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-white/[0.04] transition-all cursor-pointer"
        >
          {copied ? (
            <>
              <Check size={14} className="text-emerald-400" /> Copied Code!
            </>
          ) : (
            <>
              <Copy size={14} /> Copy Code
            </>
          )}
        </button>
      </div>

      {/* Dark Theme Code Block */}
      <div className="relative rounded-xl overflow-hidden border border-[var(--color-border)] bg-[#050810]">
        <div className="px-4 py-2 bg-white/[0.03] border-b border-[var(--color-border)] flex items-center justify-between text-xs text-[var(--color-muted)] font-mono">
          <span>{activeLang.toUpperCase()} &bull; {endpoints.find(e => e.key === activeEndpoint)?.label}</span>
          <span>DilithiAuth Hybrid Quantum Protocol</span>
        </div>
        <pre className="code-block p-4 overflow-x-auto text-xs font-mono leading-relaxed text-slate-200">
          {currentSnippet}
        </pre>
      </div>

      {/* Informational Banner */}
      <div className="flex items-start gap-3 p-4 rounded-xl bg-blue-500/10 border border-blue-500/25 text-blue-300 text-xs leading-relaxed">
        <Info size={18} className="shrink-0 mt-0.5 text-blue-400" />
        <div>
          DilithiAuth tokens use hybrid signatures combining Ed25519 and NIST FIPS 204 ML-DSA-44. Verify token signatures against public keys retrieved from{' '}
          <code className="bg-white/10 px-1.5 py-0.5 rounded font-mono text-cyan-300">GET /.well-known/jwks.json</code>.
        </div>
      </div>
    </div>
  )
}
