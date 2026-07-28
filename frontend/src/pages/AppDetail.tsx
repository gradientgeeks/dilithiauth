import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import {
  ArrowLeft, Copy, Check, RotateCw, Trash2, AlertTriangle, AlertCircle,
  KeyRound, ShieldCheck, Info, Code, Globe, Lock, Terminal, FileCode
} from 'lucide-react'
import { apps } from '../api'
import type { OAuthApp } from '../api'
import Navbar from '../components/Navbar'
import ConfirmModal from '../components/ConfirmModal'

type TabType = 'overview' | 'snippets' | 'danger'
type LangType = 'curl' | 'javascript' | 'go' | 'python' | 'java' | 'dotnet'

const btnBase = "inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold border transition-all duration-150 cursor-pointer"
const btnOutline = `${btnBase} bg-transparent border-[var(--color-border)] text-[var(--color-sub)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] hover:border-[var(--color-border-mid)]`


export default function AppDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [app, setApp] = useState<OAuthApp | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [rotatedSecret, setRotatedSecret] = useState<string | null>(null)
  const [rotating, setRotating] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [copiedId, setCopiedId] = useState(false)
  const [copiedSecret, setCopiedSecret] = useState(false)
  const [copiedSnippet, setCopiedSnippet] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [confirmRotate, setConfirmRotate] = useState(false)
  const [activeTab, setActiveTab] = useState<TabType>('overview')
  const [codeLang, setCodeLang] = useState<LangType>('curl')

  useEffect(() => {
    if (!id) return
    apps.get(id).then(r => setApp(r.data)).catch(() => setError('App not found.')).finally(() => setLoading(false))
  }, [id])

  const handleRotate = async () => {
    if (!id || !app) return
    setConfirmRotate(false)
    setRotating(true)
    try {
      const r = await apps.rotateSecret(id)
      setRotatedSecret(r.data.client_secret)
      setApp(r.data.app)
    } catch { setError('Failed to rotate secret.') }
    finally { setRotating(false) }
  }

  const handleDelete = async () => {
    if (!id) return
    setDeleting(true)
    try { await apps.delete(id); navigate('/apps') }
    catch { setError('Failed to delete app.'); setDeleting(false); setConfirmDelete(false) }
  }

  const copy = (text: string, setCopied: (v: boolean) => void) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (loading) {
    return (
      <>
        <Navbar />
        <div className="w-full max-w-[1020px] mx-auto px-4 sm:px-6 py-8">
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 space-y-3">
            {[40, 80, 60].map((w, i) => <div key={i} className="skeleton h-4 rounded" style={{ width: `${w}%` }} />)}
          </div>
        </div>
      </>
    )
  }

  if (error && !app) {
    return (
      <>
        <Navbar />
        <div className="w-full max-w-[1020px] mx-auto px-4 sm:px-6 py-8">
          <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
            <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
          </div>
          <Link to="/apps" className={btnOutline}><ArrowLeft size={14} /> Back to Applications</Link>
        </div>
      </>
    )
  }

  if (!app) return null

  const redirectUri = app.RedirectURIs[0] || 'https://myapp.com/callback'

  const snippets: Record<LangType, string> = {
    curl: `# 1. Redirect user to Authorize endpoint:
https://dilithiauthapis.gradientgeeks.tech/authorize?client_id=${app.ClientID}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&state=xyz123

# 2. Exchange authorization code for Post-Quantum JWT Token:
curl -X POST https://dilithiauthapis.gradientgeeks.tech/api/v1/oauth/token \\
  -u "${app.ClientID}:${app.IsPublic ? '' : 'YOUR_CLIENT_SECRET'}" \\
  -d "grant_type=authorization_code" \\
  -d "code=AUTHORIZATION_CODE" \\
  -d "redirect_uri=${encodeURIComponent(redirectUri)}"`,
    javascript: `// Install SDK: npm install dilithiauth-js
import { DilithiAuthClient } from 'dilithiauth-js';

const authClient = new DilithiAuthClient({
  domain: 'https://dilithiauthapis.gradientgeeks.tech',
  clientId: '${app.ClientID}',
  ${app.IsPublic ? '' : "clientSecret: 'YOUR_CLIENT_SECRET',"}
  redirectUri: '${redirectUri}'
});

await authClient.loginWithRedirect();
const { user, accessToken } = await authClient.handleRedirectCallback();
console.log('User claims:', user);`,
    go: `// go get github.com/Uttam-Mahata/dilithiauth-go
package main

import (
    "context"
    "fmt"
    dilithiauth "github.com/Uttam-Mahata/dilithiauth-go"
)

func main() {
    client, _ := dilithiauth.NewClient(dilithiauth.Config{
        Domain:      "https://dilithiauthapis.gradientgeeks.tech",
        ClientID:    "${app.ClientID}",
        ${app.IsPublic ? '' : `ClientSecret: "YOUR_CLIENT_SECRET",`}
        RedirectURI: "${redirectUri}",
    })
    tok, _ := client.ExchangeCode(context.Background(), "AUTH_CODE", "${redirectUri}", "VERIFIER")
    fmt.Println("Access Token:", tok.AccessToken)
}`,
    python: `# pip install dilithiauth
from dilithiauth import DilithiAuthClient

client = DilithiAuthClient(
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_id="${app.ClientID}",
    ${app.IsPublic ? '' : 'client_secret="YOUR_CLIENT_SECRET",'}
    redirect_uri="${redirectUri}"
)
token_resp = client.exchange_code(code="AUTHORIZATION_CODE", code_verifier="PKCE_VERIFIER", redirect_uri="${redirectUri}")
user_info = client.get_user_info(token_resp['access_token'])
print(f"Logged in user: {user_info['sub']}")`,
    java: `// com.dilithiauth:dilithiauth-sdk:1.0.0
DilithiAuthClient client = DilithiAuthClient.builder()
    .domain("https://dilithiauthapis.gradientgeeks.tech")
    .clientId("${app.ClientID}")
    ${app.IsPublic ? '' : '.clientSecret("YOUR_CLIENT_SECRET")'}
    .redirectUri("${redirectUri}")
    .build();
TokenResponse token = client.exchangeCode("AUTH_CODE", "VERIFIER", "${redirectUri}");
System.out.println("Access Token: " + token.getAccessToken());`,
    dotnet: `// dotnet add package DilithiAuth.Client
var client = new DilithiAuthClient(new DilithiAuthOptions {
    Domain = "https://dilithiauthapis.gradientgeeks.tech",
    ClientId = "${app.ClientID}",
    ${app.IsPublic ? '' : 'ClientSecret = "YOUR_CLIENT_SECRET",'}
    RedirectUri = "${redirectUri}"
});
TokenResponse token = await client.ExchangeCodeAsync("AUTH_CODE", "VERIFIER", "${redirectUri}");
Console.WriteLine($"User: {(await client.GetUserInfoAsync(token.AccessToken)).Sub}");`
  }

  const tabs: { key: TabType; icon: React.ElementType; label: string }[] = [
    { key: 'overview', icon: KeyRound, label: 'Credentials & Overview' },
    { key: 'snippets', icon: Code, label: 'SDK & Code Snippets' },
    { key: 'danger', icon: Trash2, label: 'Danger Zone' },
  ]

  const langs: { key: LangType; icon: React.ElementType; label: string }[] = [
    { key: 'curl', icon: Terminal, label: 'cURL' },
    { key: 'javascript', icon: FileCode, label: 'JS / Node' },
    { key: 'go', icon: Code, label: 'Go' },
    { key: 'python', icon: FileCode, label: 'Python' },
    { key: 'java', icon: Code, label: 'Java' },
    { key: 'dotnet', icon: Code, label: '.NET' },
  ]

  return (
    <>
      <Navbar />
      <div className="w-full max-w-[1020px] mx-auto px-4 sm:px-6 py-8">

        {/* Breadcrumb */}
        <div className="text-[0.83rem] text-[var(--color-muted)] mb-2">
          <Link to="/apps" className="hover:text-[var(--color-sub)] no-underline">My Apps</Link>
          {' / '}
          <span className="text-[var(--color-txt)]">{app.Name}</span>
        </div>

        {/* App title */}
        <div className="flex flex-wrap items-start justify-between gap-4 mb-5">
          <div>
            <div className="flex flex-wrap items-center gap-3 mb-1">
              <h1 className="text-[1.6rem] font-bold text-[var(--color-txt)] tracking-tight">{app.Name}</h1>
              <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[0.68rem] font-bold uppercase tracking-wide
                ${app.IsPublic ? 'bg-[#1e293b] border border-[#334155] text-sky-400' : 'bg-[#052e16] border border-[#166534] text-emerald-400'}`}>
                {app.IsPublic ? <Globe size={12} /> : <Lock size={12} />}
                {app.IsPublic ? 'Public Client (PKCE)' : 'Confidential Client'}
              </span>
            </div>
            {app.Description
              ? <p className="text-[var(--color-sub)] text-sm">{app.Description}</p>
              : <p className="text-[var(--color-muted)] text-sm">No description provided.</p>}
          </div>
        </div>

        {error && (
          <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-4 bg-rose-500/10 border border-rose-500/28 text-rose-300 text-sm">
            <AlertCircle size={18} className="shrink-0 mt-0.5" /><div>{error}</div>
          </div>
        )}

        {/* Tabs */}
        <div className="flex gap-0 border-b border-[var(--color-border)] mb-5 overflow-x-auto">
          {tabs.map(({ key, icon: Icon, label }) => (
            <button
              key={key}
              onClick={() => setActiveTab(key)}
              className={`inline-flex items-center gap-1.5 px-4 py-3 text-sm font-medium whitespace-nowrap
                border-b-2 transition-all duration-150 cursor-pointer bg-transparent border-l-0 border-r-0 border-t-0
                ${activeTab === key
                  ? 'text-[var(--color-txt)] border-b-[var(--color-primary)] font-semibold'
                  : 'text-[var(--color-sub)] border-b-transparent hover:text-[var(--color-txt)] hover:bg-white/[0.03]'
                }`}
            >
              <Icon size={16} /> {label}
            </button>
          ))}
        </div>

        {/* TAB: OVERVIEW */}
        {activeTab === 'overview' && (
          <div className="space-y-4">
            {/* Credentials card */}
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5">
              <h2 className="flex items-center gap-2 text-[var(--color-txt)] font-semibold text-base mb-4">
                <KeyRound size={18} className="text-cyan-400" /> Client Credentials
              </h2>

              <div className="mb-4">
                <label className="block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">Client ID</label>
                <div className="flex gap-2 items-center">
                  <span className="mono-inline flex-1 px-3 py-2.5 text-[0.9rem]">{app.ClientID}</span>
                  <button className={`${btnOutline} gap-1.5`} onClick={() => copy(app.ClientID, setCopiedId)}>
                    {copiedId ? <><Check size={14} className="text-emerald-400" /> Copied!</> : <><Copy size={14} /> Copy</>}
                  </button>
                </div>
              </div>

              {!app.IsPublic && (
                rotatedSecret ? (
                  <div>
                    <div className="flex items-start gap-2.5 p-3.5 rounded-md mb-3 bg-amber-500/10 border border-amber-500/28 text-amber-300 text-sm">
                      <AlertTriangle size={18} className="shrink-0 mt-0.5" />
                      <div>New secret generated! Copy it now as it will <strong>never be displayed again</strong>.</div>
                    </div>
                    <label className="block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">New Client Secret</label>
                    <pre className="secret-box mb-2">{rotatedSecret}</pre>
                    <div className="flex gap-2">
                      <button className={btnOutline} onClick={() => copy(rotatedSecret, setCopiedSecret)}>
                        {copiedSecret ? <><Check size={14} className="text-emerald-400" /> Copied Secret!</> : <><Copy size={14} /> Copy</>}
                      </button>
                      <button className={btnOutline} onClick={() => setRotatedSecret(null)}>I've Saved My Secret</button>
                    </div>
                  </div>
                ) : (
                  <div>
                    <label className="block text-[0.7rem] font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">Client Secret</label>
                    <div className="flex flex-wrap items-center gap-3">
                      <span className="mono-inline tracking-[0.25em] text-[var(--color-muted)]">••••••••••••••••••••••••••••••••</span>
                      <button className={btnOutline} onClick={() => setConfirmRotate(true)} disabled={rotating}>
                        <RotateCw size={14} className={rotating ? 'spin' : ''} /> {rotating ? 'Rotating…' : 'Rotate Secret'}
                      </button>
                    </div>
                    {app.SecretRotatedAt && (
                      <p className="text-[0.77rem] text-[var(--color-muted)] mt-2">
                        Last rotated: {new Date(app.SecretRotatedAt).toLocaleString()}
                      </p>
                    )}
                  </div>
                )
              )}

              {app.IsPublic && (
                <div className="flex items-start gap-2.5 p-3.5 rounded-md mt-3 bg-blue-500/10 border border-blue-500/28 text-blue-300 text-sm">
                  <ShieldCheck size={18} className="shrink-0 mt-0.5" />
                  <div>Public application client — configured for PKCE flow (<code className="text-xs bg-white/10 px-1 py-0.5 rounded">code_challenge</code> &amp; <code className="text-xs bg-white/10 px-1 py-0.5 rounded">code_verifier</code>). No client secret required.</div>
                </div>
              )}
            </div>

            {/* App config card */}
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5">
              <h2 className="text-[var(--color-txt)] font-semibold text-base mb-4">Application Configuration</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
                <div>
                  <div className="text-[0.7rem] font-bold text-[var(--color-muted)] uppercase tracking-wider mb-1">Homepage URL</div>
                  <div className="text-[0.9rem] text-[var(--color-txt)]">
                    {app.HomepageURL
                      ? <a href={app.HomepageURL} target="_blank" rel="noreferrer" className="text-[var(--color-primary)] hover:text-blue-400">{app.HomepageURL}</a>
                      : <span className="text-[var(--color-muted)]">—</span>}
                  </div>
                </div>
                <div>
                  <div className="text-[0.7rem] font-bold text-[var(--color-muted)] uppercase tracking-wider mb-1">Creation Date</div>
                  <div className="text-[0.9rem] text-[var(--color-txt)]">{new Date(app.CreatedAt).toLocaleString()}</div>
                </div>
              </div>
              <div className="h-px bg-[var(--color-border)] mb-4" />
              <div className="text-[0.7rem] font-bold text-[var(--color-muted)] uppercase tracking-wider mb-2">Authorized Redirect URIs</div>
              {app.RedirectURIs.map(uri => (
                <div key={uri} className="mono-inline block px-3 py-2 mb-1.5">{uri}</div>
              ))}
            </div>
          </div>
        )}

        {/* TAB: SNIPPETS */}
        {activeTab === 'snippets' && (
          <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5">
            <h2 className="text-[var(--color-txt)] font-semibold text-base mb-1">OAuth 2.0 Integration Code Generator</h2>
            <p className="text-[var(--color-sub)] text-[0.88rem] mb-4">
              Copy pre-configured code snippets to initiate the Post-Quantum authorization code exchange.
            </p>

            <div className="flex flex-wrap gap-1.5 mb-4">
              {langs.map(({ key, icon: Icon, label }) => (
                <button
                  key={key}
                  onClick={() => setCodeLang(key)}
                  className={`${btnBase} ${codeLang === key
                    ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
                    : 'bg-transparent border-[var(--color-border)] text-[var(--color-sub)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)]'
                  }`}
                >
                  <Icon size={14} /> {label}
                </button>
              ))}
            </div>

            <div className="flex justify-end mb-2">
              <button className={btnOutline} onClick={() => copy(snippets[codeLang], setCopiedSnippet)}>
                {copiedSnippet ? <><Check size={13} className="text-emerald-400" /> Copied!</> : <><Copy size={13} /> Copy Code</>}
              </button>
            </div>
            <pre className="code-block">{snippets[codeLang]}</pre>

            <div className="flex items-start gap-2.5 p-3.5 rounded-md mt-4 bg-blue-500/10 border border-blue-500/28 text-blue-300 text-sm">
              <Info size={18} className="shrink-0 mt-0.5" />
              <div>DilithiAuth issues hybrid signatures containing Ed25519 &amp; FIPS 204 ML-DSA-44. Verify tokens using <code className="text-xs bg-white/10 px-1 py-0.5 rounded">GET /.well-known/jwks.json</code>.</div>
            </div>
          </div>
        )}

        {/* TAB: DANGER */}
        {activeTab === 'danger' && (
          <div className="bg-rose-500/5 border border-rose-500/35 rounded-xl p-5">
            <h2 className="text-rose-400 font-semibold text-base mb-2">Danger Zone</h2>
            <p className="text-[var(--color-sub)] text-[0.88rem] mb-4">
              Deleting this application will revoke all client credentials and immediately invalidate issued tokens.
            </p>
            <button
              className={`${btnBase} bg-rose-500 text-white border-rose-500 px-4 py-2.5 text-sm
                hover:bg-rose-600 hover:shadow-[0_0_0_3px_rgba(239,68,68,0.2)] active:translate-y-px`}
              onClick={() => setConfirmDelete(true)}
            >
              <Trash2 size={16} /> Permanently Delete Application
            </button>
          </div>
        )}
      </div>

      {confirmDelete && (
        <ConfirmModal
          title="Delete this application?"
          message="This action immediately revokes all client credentials and issued tokens. This action cannot be undone."
          confirmLabel="Yes, delete permanently"
          loading={deleting}
          onConfirm={handleDelete}
          onCancel={() => setConfirmDelete(false)}
        />
      )}
      {confirmRotate && (
        <ConfirmModal
          title="Rotate Client Secret?"
          message="The existing secret will stop working immediately. Make sure to update your application settings with the new secret."
          confirmLabel="Rotate Secret Now"
          loading={rotating}
          onConfirm={handleRotate}
          onCancel={() => setConfirmRotate(false)}
        />
      )}
    </>
  )
}
