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
  const [codeLang, setCodeLang] = useState<'curl' | 'javascript' | 'go' | 'python' | 'java' | 'dotnet'>('curl')

  useEffect(() => {
    if (!id) return
    apps.get(id)
      .then(r => setApp(r.data))
      .catch(() => setError('App not found.'))
      .finally(() => setLoading(false))
  }, [id])

  const handleRotate = async () => {
    if (!id || !app) return
    setConfirmRotate(false)
    setRotating(true)
    try {
      const r = await apps.rotateSecret(id)
      setRotatedSecret(r.data.client_secret)
      setApp(r.data.app)
    } catch {
      setError('Failed to rotate secret.')
    } finally {
      setRotating(false)
    }
  }

  const handleDelete = async () => {
    if (!id) return
    setDeleting(true)
    try {
      await apps.delete(id)
      navigate('/apps')
    } catch {
      setError('Failed to delete app.')
      setDeleting(false)
      setConfirmDelete(false)
    }
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
        <div className="container">
          <div className="card">
            <div className="skeleton skeleton-row" style={{ width: '40%', height: '1.6rem' }} />
            <div className="skeleton skeleton-row" style={{ width: '80%' }} />
            <div className="skeleton skeleton-row" style={{ width: '60%' }} />
          </div>
        </div>
      </>
    )
  }

  if (error && !app) {
    return (
      <>
        <Navbar />
        <div className="container">
          <div className="alert alert-error"><AlertCircle size={18} /><div>{error}</div></div>
          <Link to="/apps" className="btn btn-outline btn-sm"><ArrowLeft size={14} /> Back to Applications</Link>
        </div>
      </>
    )
  }

  if (!app) return null

  const redirectUri = app.RedirectURIs[0] || 'https://myapp.com/callback'

  const snippets = {
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

// 1. Trigger PKCE login redirect
await authClient.loginWithRedirect();

// 2. Handle callback on redirect route (/callback)
const { user, accessToken, pqcRef } = await authClient.handleRedirectCallback();
console.log('User claims:', user);`,

    go: `// Install SDK: go get github.com/Uttam-Mahata/DilithiAuth/sdk/go
package main

import (
    "context"
    "fmt"
    dilithiauth "github.com/Uttam-Mahata/DilithiAuth/sdk/go"
)

func main() {
    client, _ := dilithiauth.NewClient(dilithiauth.Config{
        Domain:       "https://dilithiauthapis.gradientgeeks.tech",
        ClientID:     "${app.ClientID}",
        ${app.IsPublic ? '' : `ClientSecret: "YOUR_CLIENT_SECRET",`}
        RedirectURI:  "${redirectUri}",
    })

    // Exchange code for Post-Quantum Token
    tok, _ := client.ExchangeCode(context.Background(), "AUTH_CODE", "${redirectUri}", "VERIFIER")
    fmt.Println("Access Token:", tok.AccessToken)
}`,


    python: `# Install SDK: pip install dilithiauth
from dilithiauth import DilithiAuthClient, generate_pkce_pair

client = DilithiAuthClient(
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_id="${app.ClientID}",
    ${app.IsPublic ? '' : 'client_secret="YOUR_CLIENT_SECRET",'}
    redirect_uri="${redirectUri}"
)

# Exchange authorization code for token
token_resp = client.exchange_code(
    code="AUTHORIZATION_CODE",
    code_verifier="PKCE_VERIFIER",
    redirect_uri="${redirectUri}"
)
user_info = client.get_user_info(token_resp['access_token'])
print(f"Logged in user: {user_info['sub']}")`,

    java: `// Install SDK: com.dilithiauth:dilithiauth-sdk:1.0.0
import com.dilithiauth.client.DilithiAuthClient;
import com.dilithiauth.client.TokenResponse;

DilithiAuthClient client = DilithiAuthClient.builder()
    .domain("https://dilithiauthapis.gradientgeeks.tech")
    .clientId("${app.ClientID}")
    ${app.IsPublic ? '' : '.clientSecret("YOUR_CLIENT_SECRET")'}
    .redirectUri("${redirectUri}")
    .build();

TokenResponse token = client.exchangeCode("AUTH_CODE", "VERIFIER", "${redirectUri}");
System.out.println("Access Token: " + token.getAccessToken());`,

    dotnet: `// Install SDK: dotnet add package DilithiAuth.Client
using DilithiAuth.Client;

var client = new DilithiAuthClient(new DilithiAuthOptions
{
    Domain = "https://dilithiauthapis.gradientgeeks.tech",
    ClientId = "${app.ClientID}",
    ${app.IsPublic ? '' : 'ClientSecret = "YOUR_CLIENT_SECRET",'}
    RedirectUri = "${redirectUri}"
});

TokenResponse token = await client.ExchangeCodeAsync("AUTH_CODE", "VERIFIER", "${redirectUri}");
UserInfo user = await client.GetUserInfoAsync(token.AccessToken);
Console.WriteLine($"User: {user.Sub}");`
  }

  return (
    <>
      <Navbar />
      <div className="container">
        {/* Navigation Breadcrumb */}
        <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '0.6rem' }}>
          <Link to="/apps">My Apps</Link> / <span style={{ color: 'var(--text-main)' }}>{app.Name}</span>
        </div>

        <div className="page-header">
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '0.3rem' }}>
              <h1>{app.Name}</h1>
              <span className={`badge ${app.IsPublic ? 'badge-public' : 'badge-confidential'}`}>
                {app.IsPublic ? <Globe size={12} /> : <Lock size={12} />}
                {app.IsPublic ? 'Public Client (PKCE)' : 'Confidential Client'}
              </span>
            </div>
            {app.Description ? <p>{app.Description}</p> : <p className="text-muted">No description provided.</p>}
          </div>
        </div>

        {error && <div className="alert alert-error"><AlertCircle size={18} /><div>{error}</div></div>}

        {/* Tab Headers */}
        <div className="tab-header">
          <button
            className={`tab-btn ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            <KeyRound size={16} /> Credentials & Overview
          </button>
          <button
            className={`tab-btn ${activeTab === 'snippets' ? 'active' : ''}`}
            onClick={() => setActiveTab('snippets')}
          >
            <Code size={16} /> SDK & Code Snippets
          </button>
          <button
            className={`tab-btn ${activeTab === 'danger' ? 'active' : ''}`}
            onClick={() => setActiveTab('danger')}
          >
            <Trash2 size={16} /> Danger Zone
          </button>
        </div>

        {/* TAB 1: OVERVIEW & CREDENTIALS */}
        {activeTab === 'overview' && (
          <>
            {/* Credentials Card */}
            <div className="card">
              <h2><KeyRound size={18} style={{ verticalAlign: '-3px', marginRight: '0.4rem', color: 'var(--cyan)' }} />Client Credentials</h2>
              
              <div className="form-group">
                <label className="label">Client ID</label>
                <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                  <span className="mono" style={{ flex: 1, padding: '0.65rem 0.9rem', fontSize: '0.9rem' }}>
                    {app.ClientID}
                  </span>
                  <button className="btn btn-outline btn-sm copy-btn" onClick={() => copy(app.ClientID, setCopiedId)}>
                    {copiedId ? <><Check size={14} color="var(--emerald)" /> Copied!</> : <><Copy size={14} /> Copy</>}
                  </button>
                </div>
              </div>

              {!app.IsPublic && (
                <div>
                  {rotatedSecret ? (
                    <div>
                      <div className="alert alert-warning">
                        <AlertTriangle size={18} />
                        <div>New secret generated! Copy it now as it will <strong>never be displayed again</strong>.</div>
                      </div>
                      <label className="label">New Client Secret</label>
                      <div className="secret-box">{rotatedSecret}</div>
                      <div className="flex gap-2 mt-1">
                        <button className="btn btn-outline btn-sm copy-btn" onClick={() => copy(rotatedSecret, setCopiedSecret)}>
                          {copiedSecret ? <><Check size={14} color="var(--emerald)" /> Copied Secret!</> : <><Copy size={14} /> Copy</>}
                        </button>
                        <button className="btn btn-outline btn-sm" onClick={() => setRotatedSecret(null)}>
                          I've Saved My Secret
                        </button>
                      </div>
                    </div>
                  ) : (
                    <div>
                      <label className="label">Client Secret</label>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
                        <span className="mono" style={{ letterSpacing: '0.2em', color: 'var(--text-muted)' }}>
                          ••••••••••••••••••••••••••••••••
                        </span>
                        <button className="btn btn-outline btn-sm" onClick={() => setConfirmRotate(true)} disabled={rotating}>
                          <RotateCw size={14} className={rotating ? 'spin' : ''} /> {rotating ? 'Rotating…' : 'Rotate Secret'}
                        </button>
                      </div>
                      {app.SecretRotatedAt && (
                        <p className="form-hint mt-1">
                          Last rotated: {new Date(app.SecretRotatedAt).toLocaleString()}
                        </p>
                      )}
                    </div>
                  )}
                </div>
              )}

              {app.IsPublic && (
                <div className="alert alert-info mt-1">
                  <ShieldCheck size={18} />
                  <div>Public application client — configured for PKCE flow (<code>code_challenge</code> & <code>code_verifier</code>). No client secret required.</div>
                </div>
              )}
            </div>

            {/* App Details Card */}
            <div className="card">
              <h2>Application Configuration</h2>
              <div className="detail-grid">
                <div>
                  <div className="detail-label">Homepage URL</div>
                  <div className="detail-value">
                    {app.HomepageURL ? <a href={app.HomepageURL} target="_blank" rel="noreferrer">{app.HomepageURL}</a> : <span className="text-muted">—</span>}
                  </div>
                </div>
                <div>
                  <div className="detail-label">Creation Date</div>
                  <div className="detail-value">{new Date(app.CreatedAt).toLocaleString()}</div>
                </div>
              </div>

              <div className="divider" />
              <div className="detail-label" style={{ marginBottom: '0.5rem' }}>Authorized Redirect URIs</div>
              {app.RedirectURIs.map(uri => (
                <div key={uri} className="mono" style={{ marginBottom: '0.4rem', display: 'block', padding: '0.45rem 0.75rem' }}>
                  {uri}
                </div>
              ))}
            </div>
          </>
        )}

        {/* TAB 2: CODE SNIPPET GENERATOR */}
        {activeTab === 'snippets' && (
          <div className="card">
            <h2>OAuth 2.0 Integration Code Generator</h2>
            <p className="text-muted" style={{ marginBottom: '1.25rem', fontSize: '0.9rem' }}>
              Copy pre-configured code snippets to initiate the Post-Quantum authorization code exchange.
            </p>

            <div style={{ display: 'flex', gap: '0.4rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
              <button className={`btn btn-sm ${codeLang === 'curl' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('curl')}>
                <Terminal size={14} /> cURL
              </button>
              <button className={`btn btn-sm ${codeLang === 'javascript' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('javascript')}>
                <FileCode size={14} /> JS / Node
              </button>
              <button className={`btn btn-sm ${codeLang === 'go' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('go')}>
                <Code size={14} /> Go
              </button>
              <button className={`btn btn-sm ${codeLang === 'python' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('python')}>
                <FileCode size={14} /> Python
              </button>
              <button className={`btn btn-sm ${codeLang === 'java' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('java')}>
                <Code size={14} /> Java
              </button>
              <button className={`btn btn-sm ${codeLang === 'dotnet' ? 'btn-primary' : 'btn-outline'}`} onClick={() => setCodeLang('dotnet')}>
                <Code size={14} /> .NET
              </button>
            </div>

            <div style={{ position: 'relative' }}>
              <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '0.4rem' }}>
                <button className="btn btn-outline btn-sm" onClick={() => copy(snippets[codeLang], setCopiedSnippet)}>
                  {copiedSnippet ? <><Check size={13} color="var(--emerald)" /> Copied!</> : <><Copy size={13} /> Copy Code</>}
                </button>
              </div>
              <pre className="code-block">
                {snippets[codeLang]}
              </pre>
            </div>

            <div className="alert alert-info" style={{ marginTop: '1.25rem' }}>
              <Info size={18} />
              <div>
                DilithiAuth issues hybrid signatures containing Ed25519 & FIPS 204 ML-DSA-44. Verify tokens using the OIDC discovery key set at <code>GET /.well-known/jwks.json</code>.
              </div>
            </div>
          </div>
        )}

        {/* TAB 3: DANGER ZONE */}
        {activeTab === 'danger' && (
          <div className="card" style={{ borderColor: 'rgba(244, 63, 94, 0.4)', background: 'rgba(244, 63, 94, 0.05)' }}>
            <h2 style={{ color: 'var(--rose)' }}>Danger Zone</h2>
            <p className="text-muted" style={{ marginBottom: '1.25rem', fontSize: '0.9rem' }}>
              Deleting this application will revoke all client credentials and immediately invalidate issued tokens.
            </p>
            <button
              className="btn btn-danger"
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
