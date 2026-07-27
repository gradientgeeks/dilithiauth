import { useState } from 'react'
import {
  BookOpen,
  Shield,
  Code,
  Copy,
  Check,
  Terminal,
  Key,
  Cpu,
  ExternalLink,
  Layers,
  ChevronRight
} from 'lucide-react'
import Navbar from '../components/Navbar'

type Language = 'ts' | 'go' | 'python' | 'java' | 'dotnet'
type TokenStrategy = 'thin' | 'full' | 'cwt'
type EndpointKey = 'discovery' | 'jwks' | 'authorize' | 'token' | 'userinfo' | 'revoke'

interface LangInfo {
  id: Language
  name: string
  badge: string
  installCmd: string
  manifestFile: string
  manifestContent: string
  pkceCode: string
  clientInitCode: string
  tokenExchangeCode: string
  userinfoCode: string
}

const LANGUAGE_DATA: Record<Language, LangInfo> = {
  ts: {
    id: 'ts',
    name: 'TypeScript / JS',
    badge: 'npm',
    installCmd: 'npm install dilithiauth-js',
    manifestFile: 'package.json',
    manifestContent: `{\n  "dependencies": {\n    "dilithiauth-js": "^1.0.0"\n  }\n}`,
    pkceCode: `import { generateCodeVerifier, generateCodeChallenge } from 'dilithiauth-js';

// 1. Generate high-entropy Code Verifier (64 random chars)
const codeVerifier = generateCodeVerifier(64);

// 2. Compute S256 Code Challenge
const codeChallenge = await generateCodeChallenge(codeVerifier);

console.log('Verifier:', codeVerifier);
console.log('Challenge:', codeChallenge);`,
    clientInitCode: `import { DilithiAuthClient } from 'dilithiauth-js';

const client = new DilithiAuthClient({
  domain: 'https://dilithiauthapis.gradientgeeks.tech',
  clientId: 'dilithi-app-client-123',
  redirectUri: 'https://myapp.com/callback',
  tokenStrategy: 'thin', // 'thin' | 'full' | 'cwt'
});

// Trigger PKCE Login Redirect
await client.loginWithRedirect({
  scope: 'openid profile email',
  codeChallenge,
  codeChallengeMethod: 'S256'
});`,
    tokenExchangeCode: `// Callback handler after redirect
const result = await client.handleRedirectCallback();

console.log('Access Token:', result.accessToken);
console.log('PQC Reference Hash:', result.pqcRef);
console.log('PQC Signature Algorithm:', result.pqcAlg); // ML-DSA-44`,
    userinfoCode: `// Retrieve validated identity claims from /userinfo
const userInfo = await client.getUserInfo(result.accessToken);

console.log('Subject (sub):', userInfo.sub);
console.log('Name:', userInfo.name);
console.log('Email:', userInfo.email);`
  },
  go: {
    id: 'go',
    name: 'Go',
    badge: 'go get',
    installCmd: 'go get github.com/Uttam-Mahata/DilithiAuth/sdk/go',
    manifestFile: 'go.mod',
    manifestContent: `module myapp\n\ngo 1.24\n\nrequire github.com/Uttam-Mahata/DilithiAuth/sdk/go v1.0.0`,
    pkceCode: `package main

import (
	"fmt"
	"log"
	"github.com/Uttam-Mahata/DilithiAuth/sdk/go"
)

func main() {
	// Generate high-entropy PKCE verifier and S256 challenge
	verifier, challenge, err := dilithiauth.GeneratePKCEPair()
	if err != nil {
		log.Fatalf("PKCE generation failed: %v", err)
	}

	fmt.Printf("Verifier:  %s\\n", verifier)
	fmt.Printf("Challenge: %s\\n", challenge)
}`,
    clientInitCode: `package main

import (
	"fmt"
	"github.com/Uttam-Mahata/DilithiAuth/sdk/go"
)

func main() {
	client, err := dilithiauth.NewClient(dilithiauth.Config{
		Domain:       "https://dilithiauthapis.gradientgeeks.tech",
		ClientID:     "dilithi-app-client-123",
		ClientSecret: "sec_live_9f8a7b6c5d4e3f2a",
		RedirectURI:  "http://localhost:8080/callback",
	})
	if err != nil {
		panic(err)
	}

	authURL := client.GetAuthorizationURL(
		"http://localhost:8080/callback",
		"random-csrf-state",
		"CODE_CHALLENGE",
	)
	fmt.Println("Redirect user to:", authURL)
}`,
    tokenExchangeCode: `ctx := context.Background()
code := "received_authorization_code"

tokenResp, err := client.ExchangeCode(ctx, code, "http://localhost:8080/callback", verifier)
if err != nil {
	log.Fatalf("Token exchange error: %v", err)
}

fmt.Printf("Access Token: %s\\n", tokenResp.AccessToken)
fmt.Printf("ML-DSA-44 PQC Ref: %s\\n", tokenResp.PQCRef)`,
    userinfoCode: `claims, err := client.VerifyToken(ctx, tokenResp.AccessToken)
if err != nil {
	log.Fatalf("Token verification failed: %v", err)
}

fmt.Printf("Subject: %s\\n", claims.Subject)
fmt.Printf("Issuer:  %s\\n", claims.Issuer)`
  },
  python: {
    id: 'python',
    name: 'Python',
    badge: 'PyPI',
    installCmd: 'pip install dilithiauth',
    manifestFile: 'requirements.txt',
    manifestContent: `dilithiauth>=1.0.0`,
    pkceCode: `from dilithiauth.pkce import generate_pkce_pair

# Generate 32-byte secure random verifier & SHA-256 base64url challenge
code_verifier, code_challenge = generate_pkce_pair()

print(f"Code Verifier:  {code_verifier}")
print(f"Code Challenge: {code_challenge}")`,
    clientInitCode: `from dilithiauth import DilithiAuthClient

client = DilithiAuthClient(
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_id="dilithi-app-client-123",
    client_secret="sec_live_9f8a7b6c5d4e3f2a",
    redirect_uri="http://localhost:5000/callback"
)

auth_url = client.get_authorization_url(
    redirect_uri="http://localhost:5000/callback",
    state="random_state_string",
    code_challenge=code_challenge,
    scope="openid profile email"
)
print("Authorization URL:", auth_url)`,
    tokenExchangeCode: `token_resp = client.exchange_code(
    code="incoming_authorization_code",
    code_verifier=code_verifier,
    redirect_uri="http://localhost:5000/callback"
)

access_token = token_resp["access_token"]
pqc_ref = token_resp.get("pqc_ref")

print(f"Access Token: {access_token[:30]}...")
print(f"Post-Quantum Sig Hash Ref: {pqc_ref}")`,
    userinfoCode: `user_info = client.get_user_info(access_token)

print(f"User Subject: {user_info['sub']}")
print(f"User Email:   {user_info['email']}")`
  },
  java: {
    id: 'java',
    name: 'Java',
    badge: 'Maven',
    installCmd: 'mvn dependency:add -Dartifact=com.dilithiauth:dilithiauth-sdk:1.0.0',
    manifestFile: 'pom.xml',
    manifestContent: `<dependency>\n    <groupId>com.dilithiauth</groupId>\n    <artifactId>dilithiauth-sdk</artifactId>\n    <version>1.0.0</version>\n</dependency>`,
    pkceCode: `import com.dilithiauth.client.PKCE;

public class PkceDemo {
    public static void main(String[] args) {
        // Generate secure PKCE Pair
        PKCE.Pair pkce = PKCE.generatePair();
        
        System.out.println("Code Verifier:  " + pkce.getCodeVerifier());
        System.out.println("Code Challenge: " + pkce.getCodeChallenge());
    }
}`,
    clientInitCode: `import com.dilithiauth.client.DilithiAuthClient;

DilithiAuthClient client = DilithiAuthClient.builder()
        .domain("https://dilithiauthapis.gradientgeeks.tech")
        .clientId("dilithi-app-client-123")
        .clientSecret("sec_live_9f8a7b6c5d4e3f2a")
        .redirectUri("https://myapp.com/callback")
        .build();

String authUrl = client.getAuthorizationUrl(
        "https://myapp.com/callback",
        "state_123",
        codeChallenge,
        "openid profile email"
);`,
    tokenExchangeCode: `import com.dilithiauth.client.TokenResponse;

TokenResponse token = client.exchangeCode("AUTH_CODE", codeVerifier, "https://myapp.com/callback");

System.out.println("Access Token: " + token.getAccessToken());
System.out.println("PQC Ref Hash: " + token.getPqcRef());`,
    userinfoCode: `import com.dilithiauth.client.UserInfo;

UserInfo user = client.getUserInfo(token.getAccessToken());

System.out.println("Subject: " + user.getSub());
System.out.println("Email:   " + user.getEmail());`
  },
  dotnet: {
    id: 'dotnet',
    name: '.NET (C#)',
    badge: 'NuGet',
    installCmd: 'dotnet add package DilithiAuth.Client',
    manifestFile: 'App.csproj',
    manifestContent: `<ItemGroup>\n  <PackageReference Include="DilithiAuth.Client" Version="1.0.0" />\n</ItemGroup>`,
    pkceCode: `using DilithiAuth.Client;

// Generate PKCE S256 Verifier and Challenge Pair
PkcePair pkce = Pkce.GeneratePair();

Console.WriteLine($"Code Verifier:  {pkce.CodeVerifier}");
Console.WriteLine($"Code Challenge: {pkce.CodeChallenge}");`,
    clientInitCode: `using DilithiAuth.Client;

var client = new DilithiAuthClient(new DilithiAuthOptions
{
    Domain = "https://dilithiauthapis.gradientgeeks.tech",
    ClientId = "dilithi-app-client-123",
    ClientSecret = "sec_live_9f8a7b6c5d4e3f2a",
    RedirectUri = "https://localhost:5001/signin-oidc"
});

string authUrl = client.GetAuthorizationUrl(
    redirectUri: "https://localhost:5001/signin-oidc",
    state: "random_state",
    codeChallenge: pkce.CodeChallenge,
    scope: "openid profile email"
);`,
    tokenExchangeCode: `TokenResponse token = await client.ExchangeCodeAsync(
    code: "AUTHORIZATION_CODE",
    codeVerifier: pkce.CodeVerifier
);

Console.WriteLine($"Access Token: {token.AccessToken}");
Console.WriteLine($"PQC Ref Hash: {token.PqcRef}");`,
    userinfoCode: `UserInfo userInfo = await client.GetUserInfoAsync(token.AccessToken);

Console.WriteLine($"Subject (sub): {userInfo.Sub}");
Console.WriteLine($"User Email:    {userInfo.Email}");`
  }
}

export default function Docs() {
  const [selectedLang, setSelectedLang] = useState<Language>('ts')
  const [activeCodeStage, setActiveCodeStage] = useState<'pkce' | 'init' | 'exchange' | 'userinfo'>('pkce')
  const [copiedMap, setCopiedMap] = useState<Record<string, boolean>>({})

  // Endpoint reference state
  const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointKey>('discovery')
  const [endpointTab, setEndpointTab] = useState<'body' | 'headers' | 'metrics'>('body')
  const [clientId, setClientId] = useState('dilithi-demo-client-123')
  const [redirectUri, setRedirectUri] = useState('https://app.example.com/callback')
  const [subject, setSubject] = useState('alice@example.com')
  const [tokenStrategy, setTokenStrategy] = useState<TokenStrategy>('thin')

  const copyToClipboard = async (key: string, text: string) => {
    let copied = false
    if (navigator.clipboard && navigator.clipboard.writeText) {
      try {
        await navigator.clipboard.writeText(text)
        copied = true
      } catch {
        copied = false
      }
    }
    if (!copied) {
      try {
        const textArea = document.createElement('textarea')
        textArea.value = text
        document.body.appendChild(textArea)
        textArea.select()
        document.execCommand('copy')
        document.body.removeChild(textArea)
        copied = true
      } catch {
        // Fallback copy failed
      }
    }
    if (copied) {
      setCopiedMap((prev) => ({ ...prev, [key]: true }))
      setTimeout(() => {
        setCopiedMap((prev) => ({ ...prev, [key]: false }))
      }, 2000)
    }
  }

  const langInfo = LANGUAGE_DATA[selectedLang]

  const toBase64Url = (obj: object) => {
    try {
      const json = JSON.stringify(obj)
      const encoded = encodeURIComponent(json).replace(/%([0-9A-F]{2})/g, (_, p1) =>
        String.fromCharCode(parseInt(p1, 16))
      )
      return btoa(encoded)
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '')
    } catch {
      return 'eyJ...'
    }
  }

  const getEndpointResponse = () => {
    switch (selectedEndpoint) {
      case 'discovery':
        return {
          status: 200,
          statusText: 'OK',
          method: 'GET',
          url: '/.well-known/openid-configuration',
          body: {
            issuer: 'https://dilithiauthapis.gradientgeeks.tech',
            authorization_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/authorize',
            token_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/token',
            userinfo_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/userinfo',
            jwks_uri: 'https://dilithiauthapis.gradientgeeks.tech/.well-known/jwks.json',
            revocation_endpoint: 'https://dilithiauthapis.gradientgeeks.tech/revoke',
            response_types_supported: ['code'],
            subject_types_supported: ['public'],
            id_token_signing_alg_values_supported: ['EdDSA', 'MLDSA44'],
            code_challenge_methods_supported: ['S256'],
            pqc_hybrid_enabled: true,
            pqc_algorithms: ['ML-DSA-44', 'Ed25519']
          },
          headers: {
            'Content-Type': 'application/json; charset=utf-8',
            'Cache-Control': 'public, max-age=3600',
            'X-DilithiAuth-PQC': 'ML-DSA-44 (FIPS 204)'
          }
        }
      case 'jwks':
        return {
          status: 200,
          statusText: 'OK',
          method: 'GET',
          url: '/.well-known/jwks.json',
          body: {
            keys: [
              {
                kid: 'x9K3mP1Q',
                kty: 'OKP',
                crv: 'Ed25519',
                x: 'u18d9fA2mL5pQ8sT4vW7xZ1aB3cD5eF7gH9iJ1kL3mN',
                mldsa44_pub: 'A9x2kL7mP4qS9tV2wY5zB8cD1eF4gH7iJ0kL3mN6pQ9rS2tV5wY8zA1bC4dE7fG0hJ3kL6mN9pQ2rS5tV8wY1zB4cD7eF0gH3iJ6kL9mN2pQ5rS8tV1wY4zB7cD0eF3gH6iJ9kL2mN5pQ8rS1tV4wY7zB0cD3eF6gH9iJ2kL5mN8pQ1rS4tV7wY0zB3cD6eF9gH2iJ5kL8mN1pQ4rS7tV0wY3zB6cD9eF2gH5iJ8kL1mN4pQ7rS0tV3wY6zB9cD2eF5gH8iJ1kL4mN7pQ0',
                use: 'sig',
                alg: 'EdDSA+MLDSA44'
              }
            ]
          },
          headers: {
            'Content-Type': 'application/json',
            'Cache-Control': 'public, max-age=86400',
            'Access-Control-Allow-Origin': '*'
          }
        }
      case 'authorize': {
        const redirectSep = redirectUri.includes('?') ? '&' : '?'
        return {
          status: 302,
          statusText: 'Found (Redirect)',
          method: 'GET',
          url: `/authorize?client_id=${encodeURIComponent(clientId)}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=openid+profile+email&code_challenge=E9Mel-pGsD-Zc4jKWBfZ6KBG_gJS5524j_ic71_715k&code_challenge_method=S256`,
          body: {
            redirect_to: `${redirectUri}${redirectSep}code=da_code_8f9a2b4c1d6e3f5a7b9c0d2e4f6a8b1c&state=random_csrf_token`,
            grant_type: 'authorization_code',
            code_challenge_method: 'S256',
            status: 'authorization_granted'
          },
          headers: {
            'Location': `${redirectUri}${redirectSep}code=da_code_8f9a2b4c1d6e3f5a7b9c0d2e4f6a8b1c`,
            'Set-Cookie': 'dilithiauth_session=sess_987654321; HttpOnly; Secure; SameSite=Lax'
          }
        }
      }
      case 'token': {
        const effectiveSub = subject.trim() || 'alice@example.com'
        const effectiveClient = clientId.trim() || 'dilithi-demo-client-123'
        if (tokenStrategy === 'thin') {
          const thinHeader = { alg: 'EdDSA', kid: 'x9K3mP1Q', typ: 'JWT' }
          const thinPayload = {
            iss: 'https://dilithiauthapis.gradientgeeks.tech',
            sub: effectiveSub,
            aud: effectiveClient,
            exp: 1771716000,
            iat: 1771712400,
            pqc_ref: '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU'
          }
          return {
            status: 200,
            statusText: 'OK',
            method: 'POST',
            url: '/token',
            body: {
              access_token: `${toBase64Url(thinHeader)}.${toBase64Url(thinPayload)}.ed25519_sig_64bytes_sample`,
              token_type: 'Bearer',
              expires_in: 3600,
              pqc_ref: '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU',
              pqc_alg: 'ML-DSA-44',
              refresh_token: 'da_rt_7f8a9b0c1d2e3f4a5b6c7d8e9f0a',
              note: 'Custom inputs (clientId, subject) reflect live verification'
            },
            headers: {
              'Content-Type': 'application/json',
              'X-DilithiAuth-Token-Strategy': 'thin',
              'X-DilithiAuth-PQC-Ref': '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU'
            }
          }
        } else if (tokenStrategy === 'full') {
          const fullHeader = { alg: 'EdDSA+', kid: 'x9K3mP1Q', typ: 'JWT' }
          const fullPayload = {
            iss: 'https://dilithiauthapis.gradientgeeks.tech',
            sub: effectiveSub,
            aud: effectiveClient,
            exp: 1771716000,
            iat: 1771712400,
            pqc_sig: 'A19x2k7mP4qS2...[ML-DSA-44 ~2.4KB Signature Embedded]'
          }
          return {
            status: 200,
            statusText: 'OK',
            method: 'POST',
            url: '/token',
            body: {
              access_token: `${toBase64Url(fullHeader)}.${toBase64Url(fullPayload)}.ed25519_sig_64bytes`,
              token_type: 'Bearer',
              expires_in: 3600,
              token_size_bytes: 3840,
              pqc_alg: 'ML-DSA-44',
              pqc_sig_bytes: 2420,
              refresh_token: 'da_rt_7f8a9b0c1d2e3f4a5b6c7d8e9f0a',
              note: 'Custom inputs (clientId, subject) reflect live verification'
            },
            headers: {
              'Content-Type': 'application/json',
              'X-DilithiAuth-Token-Strategy': 'full (Jumbo)',
              'X-DilithiAuth-MaxHeaderBytes': '16384 (16 KiB)'
            }
          }
        } else {
          return {
            status: 200,
            statusText: 'OK',
            method: 'POST',
            url: '/token',
            body: {
              cwt_hex: 'd83d8448a10168697373756572a501686973737565720265616c696365041a6627000008582047444551506a38484253612d5f54496d572d354a4365755165526b6d354e4d704a575a473368537546555840...',
              token_type: 'Bearer',
              format: 'CBOR Web Token (RFC 8392)',
              token_size_bytes: 284,
              pqc_ref: '47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU',
              note: 'Custom inputs (clientId, subject) reflect live verification'
            },
            headers: {
              'Content-Type': 'application/cbor',
              'X-DilithiAuth-Token-Strategy': 'cwt'
            }
          }
        }
      }
      case 'userinfo': {
        const namePrefix = subject.trim() ? subject.split('@')[0].toUpperCase() : 'ANONYMOUS'
        const effectiveSub = subject.trim() || 'ANONYMOUS'
        return {
          status: 200,
          statusText: 'OK',
          method: 'GET',
          url: '/userinfo',
          body: {
            sub: effectiveSub,
            name: `${namePrefix} Identity`,
            email: effectiveSub,
            email_verified: true,
            pqc_verified: true,
            hybrid_signature: 'Ed25519 + ML-DSA-44'
          },
          headers: {
            'Content-Type': 'application/json',
            'X-DilithiAuth-Sub': effectiveSub
          }
        }
      }
      case 'revoke':
        return {
          status: 200,
          statusText: 'OK',
          method: 'POST',
          url: '/revoke',
          body: {
            status: 'revoked',
            token_type_hint: 'access_token',
            pqc_sigstore_purged: true
          },
          headers: {
            'Content-Type': 'application/json'
          }
        }
    }
  }

  const endpointRes = getEndpointResponse()

  return (
    <div>
      <Navbar />

      <div className="container">
        {/* Page Header */}
        <div className="page-header" style={{ marginBottom: '2rem' }}>
          <div>
            <span className="badge badge-hybrid" style={{ marginBottom: '0.6rem' }}>
              <Shield size={12} /> NIST FIPS 204 Standard • ML-DSA-44 & Ed25519
            </span>
            <h1>DilithiAuth SDK & Interactive Documentation</h1>
            <p style={{ maxWidth: '680px' }}>
              Post-quantum OAuth 2.0 & OIDC identity integration guides, authentic PKCE S256 flow snippets,
              and live Post-Quantum Token strategy reference.
            </p>
          </div>
          <div className="flex gap-2">
            <a
              href="https://csrc.nist.gov/pubs/fips/204/final"
              target="_blank"
              rel="noopener noreferrer"
              className="btn btn-outline btn-sm"
            >
              <ExternalLink size={14} /> FIPS 204 Spec
            </a>
          </div>
        </div>

        {/* Navigation Quick Jump Tabs */}
        <div className="tab-header" style={{ marginBottom: '2rem' }}>
          <a href="#quickstart" className="tab-btn active">
            <Terminal size={15} /> SDK Quickstart
          </a>
          <a href="#code-snippets" className="tab-btn">
            <Code size={15} /> Code Snippets
          </a>
          <a href="#api-reference" className="tab-btn">
            <Key size={15} /> API Reference
          </a>
          <a href="#token-strategies" className="tab-btn">
            <Cpu size={15} /> PQC Token Strategies
          </a>
        </div>

        {/* SECTION 1: SDK Quickstart & Language Switcher */}
        <section id="quickstart" style={{ marginBottom: '2.5rem' }}>
          <div className="card">
            <div className="flex items-center justify-between flex-wrap gap-2" style={{ marginBottom: '1.25rem' }}>
              <div>
                <h2>1. SDK Installation & Quickstart</h2>
                <p className="text-muted" style={{ fontSize: '0.85rem' }}>
                  Select your primary language stack to get copy-ready package installation commands and manifests.
                </p>
              </div>
              <span className="badge badge-public">5 Languages Supported</span>
            </div>

            {/* Language Switcher Tab Bar */}
            <div
              style={{
                display: 'flex',
                gap: '0.5rem',
                marginBottom: '1.5rem',
                overflowX: 'auto',
                paddingBottom: '0.4rem',
                borderBottom: '1px solid var(--border-color)'
              }}
            >
              {(Object.keys(LANGUAGE_DATA) as Language[]).map((langKey) => {
                const lang = LANGUAGE_DATA[langKey]
                const isSelected = selectedLang === langKey
                return (
                  <button
                    key={langKey}
                    onClick={() => setSelectedLang(langKey)}
                    className={`btn ${isSelected ? 'btn-primary' : 'btn-outline'}`}
                    style={{ padding: '0.5rem 0.9rem', fontSize: '0.85rem' }}
                  >
                    <span
                      className="badge"
                      style={{
                        fontSize: '0.68rem',
                        padding: '1px 5px',
                        background: isSelected ? 'rgba(255,255,255,0.2)' : '#1f2937',
                        color: isSelected ? 'white' : '#9ca3af'
                      }}
                    >
                      {lang.badge}
                    </span>
                    {lang.name}
                  </button>
                )
              })}
            </div>

            {/* Installation Box */}
            <div className="detail-grid" style={{ marginBottom: '1.25rem' }}>
              <div>
                <div className="detail-label">CLI Package Installation</div>
                <div
                  style={{
                    background: '#060911',
                    border: '1px solid #1f2937',
                    borderRadius: '6px',
                    padding: '0.75rem 1rem',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: '0.5rem',
                    marginTop: '0.3rem'
                  }}
                >
                  <code style={{ fontFamily: 'var(--font-mono)', color: '#38bdf8', fontSize: '0.85rem' }}>
                    {langInfo.installCmd}
                  </code>
                  <button
                    className="btn btn-outline btn-sm"
                    onClick={() => copyToClipboard(`install-${langInfo.id}`, langInfo.installCmd)}
                    style={{ flexShrink: 0 }}
                  >
                    {copiedMap[`install-${langInfo.id}`] ? (
                      <>
                        <Check size={14} color="#10b981" />
                        <span style={{ color: '#10b981', fontWeight: 600 }}>Copied!</span>
                      </>
                    ) : (
                      <>
                        <Copy size={14} />
                        <span>Copy</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              <div>
                <div className="detail-label">Manifest File ({langInfo.manifestFile})</div>
                <div
                  style={{
                    background: '#060911',
                    border: '1px solid #1f2937',
                    borderRadius: '6px',
                    padding: '0.75rem 1rem',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    gap: '0.5rem',
                    marginTop: '0.3rem'
                  }}
                >
                  <code style={{ fontFamily: 'var(--font-mono)', color: '#9ca3af', fontSize: '0.8rem', whiteSpace: 'pre' }}>
                    {langInfo.manifestContent}
                  </code>
                  <button
                    className="btn btn-outline btn-sm"
                    onClick={() => copyToClipboard(`manifest-${langInfo.id}`, langInfo.manifestContent)}
                    style={{ flexShrink: 0 }}
                  >
                    {copiedMap[`manifest-${langInfo.id}`] ? (
                      <>
                        <Check size={14} color="#10b981" />
                        <span style={{ color: '#10b981', fontWeight: 600 }}>Copied!</span>
                      </>
                    ) : (
                      <>
                        <Copy size={14} />
                        <span>Copy</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* SECTION 2: Authentic PKCE & Auth Code Snippets */}
        <section id="code-snippets" style={{ marginBottom: '2.5rem' }}>
          <div className="card">
            <div className="flex items-center justify-between flex-wrap gap-2" style={{ marginBottom: '1rem' }}>
              <div>
                <h2>2. Code Challenge & Integration Snippets</h2>
                <p className="text-muted" style={{ fontSize: '0.85rem' }}>
                  Authentic PKCE S256 code challenge generation, OAuth2 authorization, token exchange, and userinfo claims for {langInfo.name}.
                </p>
              </div>
              <span className="badge badge-hybrid">
                <Shield size={11} /> PKCE S256 Mandated
              </span>
            </div>

            {/* Snippet Stage Sub-tabs */}
            <div className="tab-header" style={{ marginBottom: '1rem' }}>
              <button
                className={`tab-btn ${activeCodeStage === 'pkce' ? 'active' : ''}`}
                onClick={() => setActiveCodeStage('pkce')}
              >
                <Key size={14} /> 1. PKCE S256 Challenge
              </button>
              <button
                className={`tab-btn ${activeCodeStage === 'init' ? 'active' : ''}`}
                onClick={() => setActiveCodeStage('init')}
              >
                <Code size={14} /> 2. Client Init & Auth URL
              </button>
              <button
                className={`tab-btn ${activeCodeStage === 'exchange' ? 'active' : ''}`}
                onClick={() => setActiveCodeStage('exchange')}
              >
                <Cpu size={14} /> 3. Code Exchange & PQC Token
              </button>
              <button
                className={`tab-btn ${activeCodeStage === 'userinfo' ? 'active' : ''}`}
                onClick={() => setActiveCodeStage('userinfo')}
              >
                <BookOpen size={14} /> 4. UserInfo Retrieval
              </button>
            </div>

            {/* Active Code Block */}
            <div style={{ position: 'relative' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: '#0b0f19',
                  padding: '0.5rem 1rem',
                  borderTopLeftRadius: '6px',
                  borderTopRightRadius: '6px',
                  border: '1px solid #1f2937',
                  borderBottom: 'none'
                }}
              >
                <span className="mono" style={{ fontSize: '0.75rem' }}>
                  {selectedLang.toUpperCase()} — Stage:{' '}
                  {activeCodeStage === 'pkce'
                    ? 'PKCE S256 Generator'
                    : activeCodeStage === 'init'
                    ? 'OAuth2 Initialization'
                    : activeCodeStage === 'exchange'
                    ? 'Token Exchange'
                    : 'UserInfo Claims'}
                </span>
                <button
                  className="btn btn-outline btn-sm"
                  onClick={() =>
                    copyToClipboard(
                      `code-${selectedLang}-${activeCodeStage}`,
                      activeCodeStage === 'pkce'
                        ? langInfo.pkceCode
                        : activeCodeStage === 'init'
                        ? langInfo.clientInitCode
                        : activeCodeStage === 'exchange'
                        ? langInfo.tokenExchangeCode
                        : langInfo.userinfoCode
                    )
                  }
                >
                  {copiedMap[`code-${selectedLang}-${activeCodeStage}`] ? (
                    <>
                      <Check size={14} color="#10b981" />
                      <span style={{ color: '#10b981', fontWeight: 600 }}>Copied Code!</span>
                    </>
                  ) : (
                    <>
                      <Copy size={14} />
                      <span>Copy Code</span>
                    </>
                  )}
                </button>
              </div>

              <pre
                className="code-block"
                style={{
                  borderTopLeftRadius: 0,
                  borderTopRightRadius: 0,
                  maxHeight: '380px'
                }}
              >
                {activeCodeStage === 'pkce' && langInfo.pkceCode}
                {activeCodeStage === 'init' && langInfo.clientInitCode}
                {activeCodeStage === 'exchange' && langInfo.tokenExchangeCode}
                {activeCodeStage === 'userinfo' && langInfo.userinfoCode}
              </pre>
            </div>
          </div>
        </section>

        {/* SECTION 3: Interactive API & Endpoint Reference */}
        <section id="api-reference" style={{ marginBottom: '2.5rem' }}>
          <div className="card card-glow">
            <div className="flex items-center justify-between flex-wrap gap-2" style={{ marginBottom: '1.25rem' }}>
              <div>
                <h2>3. Interactive API & Post-Quantum Identity Reference</h2>
                <p className="text-muted" style={{ fontSize: '0.85rem' }}>
                  Test endpoints, configure live parameters (Client ID, Redirect URI, Token Strategy), and inspect exact response formats.
                </p>
              </div>
              <span className="badge badge-confidential">Live Interactive Playground</span>
            </div>

            <div className="detail-grid" style={{ gridTemplateColumns: '260px 1fr', gap: '1.25rem' }}>
              {/* Endpoint List */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
                <div className="detail-label" style={{ marginBottom: '0.4rem' }}>DilithiAuth Endpoints</div>

                <button
                  onClick={() => setSelectedEndpoint('discovery')}
                  className={`btn ${selectedEndpoint === 'discovery' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-public" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>GET</span>
                  <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>/.well-known/openid-configuration</span>
                </button>

                <button
                  onClick={() => setSelectedEndpoint('jwks')}
                  className={`btn ${selectedEndpoint === 'jwks' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-public" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>GET</span>
                  <span style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>/.well-known/jwks.json</span>
                </button>

                <button
                  onClick={() => setSelectedEndpoint('authorize')}
                  className={`btn ${selectedEndpoint === 'authorize' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-public" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>GET</span>
                  <span>/authorize (PKCE)</span>
                </button>

                <button
                  onClick={() => setSelectedEndpoint('token')}
                  className={`btn ${selectedEndpoint === 'token' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-confidential" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>POST</span>
                  <span>/token</span>
                </button>

                <button
                  onClick={() => setSelectedEndpoint('userinfo')}
                  className={`btn ${selectedEndpoint === 'userinfo' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-public" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>GET</span>
                  <span>/userinfo</span>
                </button>

                <button
                  onClick={() => setSelectedEndpoint('revoke')}
                  className={`btn ${selectedEndpoint === 'revoke' ? 'btn-primary' : 'btn-outline'}`}
                  style={{ justifyContent: 'flex-start', fontSize: '0.8rem', textAlign: 'left' }}
                >
                  <span className="badge badge-confidential" style={{ padding: '1px 4px', fontSize: '0.65rem' }}>POST</span>
                  <span>/revoke</span>
                </button>
              </div>

              {/* Endpoint Tester & Inspector */}
              <div>
                {/* Interactive Controls Header */}
                <div
                  style={{
                    background: '#0b0f19',
                    border: '1px solid #1f2937',
                    borderRadius: '6px',
                    padding: '0.85rem',
                    marginBottom: '1rem'
                  }}
                >
                  <div className="detail-label" style={{ marginBottom: '0.6rem' }}>Interactive Parameter Inputs</div>
                  <div className="detail-grid" style={{ gridTemplateColumns: '1fr 1fr', gap: '0.75rem' }}>
                    <div>
                      <label className="label">Client ID</label>
                      <input
                        type="text"
                        value={clientId}
                        onChange={(e) => setClientId(e.target.value)}
                        placeholder="Client ID"
                      />
                    </div>
                    <div>
                      <label className="label">Redirect URI</label>
                      <input
                        type="text"
                        value={redirectUri}
                        onChange={(e) => setRedirectUri(e.target.value)}
                        placeholder="Redirect URI"
                      />
                    </div>
                    <div>
                      <label className="label">User Subject (sub)</label>
                      <input
                        type="text"
                        value={subject}
                        onChange={(e) => setSubject(e.target.value)}
                        placeholder="Subject email/ID"
                      />
                    </div>
                    <div>
                      <label className="label">Token Strategy</label>
                      <select
                        value={tokenStrategy}
                        onChange={(e) => setTokenStrategy(e.target.value as TokenStrategy)}
                      >
                        <option value="thin">Thin Token (Ed25519 + pqc_ref)</option>
                        <option value="full">Jumbo Token (Embedded ML-DSA-44)</option>
                        <option value="cwt">CBOR Web Token (CWT RFC 8392)</option>
                      </select>
                    </div>
                  </div>
                </div>

                {/* Response Preview Header */}
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    borderBottom: '1px solid #1f2937',
                    paddingBottom: '0.5rem',
                    marginBottom: '0.75rem'
                  }}
                >
                  <div className="flex items-center gap-2">
                    <span
                      className={`badge ${
                        endpointRes.method === 'GET' ? 'badge-public' : 'badge-confidential'
                      }`}
                    >
                      {endpointRes.method}
                    </span>
                    <span className="mono" style={{ fontSize: '0.85rem', color: '#f9fafb' }}>
                      {endpointRes.url}
                    </span>
                    <span
                      className="badge"
                      style={{
                        background: 'rgba(5, 150, 105, 0.15)',
                        color: '#6ee7b7',
                        border: '1px solid rgba(5, 150, 105, 0.3)'
                      }}
                    >
                      {endpointRes.status} {endpointRes.statusText}
                    </span>
                  </div>

                  <div className="flex gap-2">
                    <button
                      className={`btn btn-sm ${endpointTab === 'body' ? 'btn-primary' : 'btn-outline'}`}
                      onClick={() => setEndpointTab('body')}
                    >
                      Body
                    </button>
                    <button
                      className={`btn btn-sm ${endpointTab === 'headers' ? 'btn-primary' : 'btn-outline'}`}
                      onClick={() => setEndpointTab('headers')}
                    >
                      Headers
                    </button>
                    <button
                      className={`btn btn-sm ${endpointTab === 'metrics' ? 'btn-primary' : 'btn-outline'}`}
                      onClick={() => setEndpointTab('metrics')}
                    >
                      PQC Metrics
                    </button>
                  </div>
                </div>

                {/* Inspector Output Panel */}
                {endpointTab === 'body' && (
                  <div style={{ position: 'relative' }}>
                    <button
                      className="btn btn-outline btn-sm"
                      onClick={() =>
                        copyToClipboard(`res-body-${selectedEndpoint}-${tokenStrategy}`, JSON.stringify(endpointRes.body, null, 2))
                      }
                      style={{ position: 'absolute', top: '0.5rem', right: '0.5rem', zIndex: 10 }}
                    >
                      {copiedMap[`res-body-${selectedEndpoint}-${tokenStrategy}`] ? (
                        <>
                          <Check size={13} color="#10b981" />
                          <span style={{ color: '#10b981' }}>Copied</span>
                        </>
                      ) : (
                        <>
                          <Copy size={13} />
                          <span>Copy Response</span>
                        </>
                      )}
                    </button>
                    <pre className="code-block" style={{ maxHeight: '340px' }}>
                      {JSON.stringify(endpointRes.body, null, 2)}
                    </pre>
                  </div>
                )}

                {endpointTab === 'headers' && (
                  <pre className="code-block" style={{ maxHeight: '340px' }}>
                    {Object.entries(endpointRes.headers)
                      .map(([k, v]) => `${k}: ${v}`)
                      .join('\n')}
                  </pre>
                )}

                {endpointTab === 'metrics' && (
                  <div
                    style={{
                      background: '#060911',
                      border: '1px solid #1f2937',
                      borderRadius: '6px',
                      padding: '1rem'
                    }}
                  >
                    <h3 style={{ fontSize: '0.9rem', marginBottom: '0.5rem', color: '#60a5fa' }}>
                      Post-Quantum Signature Specification Breakdown
                    </h3>
                    <ul style={{ paddingLeft: '1.2rem', color: 'var(--text-sub)', fontSize: '0.85rem', lineHeight: '1.7' }}>
                      <li>
                        <strong>Ed25519 Classical Signature Size:</strong> <span className="mono">64 Bytes</span>
                      </li>
                      <li>
                        <strong>NIST ML-DSA-44 Signature Size:</strong> <span className="mono">2,420 Bytes</span>
                      </li>
                      <li>
                        <strong>Selected Token Mode ({tokenStrategy.toUpperCase()}):</strong>{' '}
                        {tokenStrategy === 'thin' && (
                          <span>Standard JWT (~500 Bytes) + SigStore <span className="mono">pqc_ref</span> Hash</span>
                        )}
                        {tokenStrategy === 'full' && (
                          <span>Jumbo Embedded JWT (~3,840 Bytes) with full <span className="mono">pqc_sig</span> claim</span>
                        )}
                        {tokenStrategy === 'cwt' && (
                          <span>Binary CBOR Web Token (~284 Bytes RFC 8392 format)</span>
                        )}
                      </li>
                      <li>
                        <strong>DilithiAuth Server Proxy Configuration:</strong>{' '}
                        <span className="mono">MaxHeaderBytes = 16 KiB (16,384 Bytes)</span>
                      </li>
                    </ul>
                  </div>
                )}
              </div>
            </div>
          </div>
        </section>

        {/* SECTION 4: Post-Quantum Token Architecture Comparison */}
        <section id="token-strategies" style={{ marginBottom: '2.5rem' }}>
          <div className="card">
            <div className="flex items-center justify-between flex-wrap gap-2" style={{ marginBottom: '1.25rem' }}>
              <div>
                <h2>4. Post-Quantum Token Strategy Architecture</h2>
                <p className="text-muted" style={{ fontSize: '0.85rem' }}>
                  Solving the NIST ML-DSA-44 2.4 KB "token bloat" problem through modular token delivery mechanisms.
                </p>
              </div>
              <span className="badge badge-hybrid">
                <Cpu size={12} /> 16 KiB Header Engine
              </span>
            </div>

            <div className="features" style={{ padding: 0 }}>
              {/* Thin Token Card */}
              <div className="feature-card">
                <div className="feature-icon">
                  <Layers size={20} />
                </div>
                <h3>1. Thin Tokens (Recommended Default)</h3>
                <p style={{ marginBottom: '0.75rem' }}>
                  JWT contains classical Ed25519 signature and a cryptographic SHA-256 hash reference (<span className="mono">pqc_ref</span>)
                  pointing to the ML-DSA-44 signature stored out-of-band in SigStore (Redis/Memory).
                </p>
                <div className="detail-label">Size & Compatibility</div>
                <div className="mono mt-1" style={{ fontSize: '0.75rem', color: '#6ee7b7' }}>
                  ~500 Bytes • Zero Proxy Header Bloat
                </div>
              </div>

              {/* Jumbo Token Card */}
              <div className="feature-card">
                <div className="feature-icon">
                  <Key size={20} />
                </div>
                <h3>2. Jumbo Tokens (Full Embedded)</h3>
                <p style={{ marginBottom: '0.75rem' }}>
                  JWT directly embeds the entire ~2.4 KB ML-DSA-44 signature inside the <span className="mono">pqc_sig</span> claim.
                  Fully self-contained for offline verification without cache lookups.
                </p>
                <div className="detail-label">Size & Header Config</div>
                <div className="mono mt-1" style={{ fontSize: '0.75rem', color: '#fcd34d' }}>
                  ~4.0 KB • MaxHeaderBytes = 16 KiB
                </div>
              </div>

              {/* CBOR Web Token Card */}
              <div className="feature-card">
                <div className="feature-icon">
                  <ChevronRight size={20} />
                </div>
                <h3>3. CBOR Web Tokens (CWT RFC 8392)</h3>
                <p style={{ marginBottom: '0.75rem' }}>
                  Binary encoded token using CBOR integer claim mapping (1:iss, 2:sub, 4:exp, 8:pqc_ref).
                  Delivers 25-30% footprint reduction for low-bandwidth mobile and IoT networks.
                </p>
                <div className="detail-label">Size & Encoding</div>
                <div className="mono mt-1" style={{ fontSize: '0.75rem', color: '#38bdf8' }}>
                  ~280 Bytes • Binary Compact Format
                </div>
              </div>
            </div>

            {/* Architecture Footer Callout */}
            <div className="alert alert-info" style={{ marginTop: '1.25rem', marginBottom: 0 }}>
              <Shield size={16} />
              <div>
                <strong>Reverse Proxy Configuration Note:</strong> When operating Jumbo Tokens in production, reverse proxies (Nginx, Envoy, Cloudflare) must be configured with <span className="mono">large_client_header_buffers 4 16k;</span> to accommodate ~4 KB Authorization header tokens.
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}
