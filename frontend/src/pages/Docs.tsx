import { useState, useEffect, useRef } from 'react'
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
  ChevronRight,
  Menu,
  X
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
    installCmd: 'go get github.com/Uttam-Mahata/dilithiauth-go',
    manifestFile: 'go.mod',
    manifestContent: `module myapp\n\ngo 1.24\n\nrequire github.com/Uttam-Mahata/dilithiauth-go v1.0.0`,
    pkceCode: `package main

import (
	"fmt"
	"log"
	"github.com/Uttam-Mahata/dilithiauth-go"
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
	"github.com/Uttam-Mahata/dilithiauth-go"
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
    manifestContent: `<dependency>\n    <groupId>com.github.Uttam-Mahata</groupId>\n    <artifactId>dilithiauth-java</artifactId>\n    <version>1.0.1</version>\n</dependency>`,
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

  // Mobile navigation drawer state & accessibility refs
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const drawerRef = useRef<HTMLDivElement>(null)
  const triggerBtnRef = useRef<HTMLButtonElement>(null)

  // Lock Body Scroll when Mobile Drawer is open
  useEffect(() => {
    if (mobileNavOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [mobileNavOpen])

  // Escape key listener for Mobile Drawer accessibility
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && mobileNavOpen) {
        setMobileNavOpen(false)
        triggerBtnRef.current?.focus()
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [mobileNavOpen])

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
    <div className="min-h-screen bg-[var(--color-bg)] text-[var(--color-txt)]">
      <Navbar />

      {/* Mobile Sub-Header & Sticky Navigation Trigger (< md) */}
      <div className="md:hidden sticky top-[60px] z-30 flex items-center justify-between px-4 py-2.5 bg-[#090d16]/95 backdrop-blur-md border-b border-[var(--color-border)]">
        <button
          ref={triggerBtnRef}
          className="inline-flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-semibold border border-[var(--color-border)] text-[var(--color-txt)] bg-[var(--color-card)] hover:bg-[var(--color-hover)] transition-colors cursor-pointer"
          onClick={() => setMobileNavOpen(!mobileNavOpen)}
          aria-expanded={mobileNavOpen}
          aria-controls="docs-sidebar-drawer"
          aria-label="Toggle documentation navigation menu"
        >
          <Menu size={16} />
          <span>Docs Menu</span>
        </button>
        <span className="text-[var(--color-sub)] text-xs font-mono">
          Endpoint: <span className="text-sky-400">/{selectedEndpoint}</span>
        </span>
      </div>

      {/* Backdrop Overlay */}
      {mobileNavOpen && (
        <div
          className="fixed inset-0 z-[190] bg-black/65 backdrop-blur-sm"
          onClick={() => setMobileNavOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Slide-over Mobile Sidebar Drawer */}
      <aside
        id="docs-sidebar-drawer"
        ref={drawerRef}
        className={`fixed top-[60px] left-0 bottom-0 z-[191] w-[280px] bg-[var(--color-card)] border-r border-[var(--color-border)] transform transition-transform duration-200 ease-out flex flex-col ${
          mobileNavOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
        role="dialog"
        aria-modal="true"
        aria-label="Documentation Navigation Drawer"
      >
        <div className="flex items-center justify-between p-4 border-b border-[var(--color-border)]">
          <div className="flex items-center gap-2 font-semibold text-sm text-[var(--color-txt)]">
            <BookOpen size={18} className="text-blue-400" /> Documentation Menu
          </div>
          <button
            className="p-1 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors"
            onClick={() => setMobileNavOpen(false)}
            aria-label="Close navigation drawer"
          >
            <X size={18} />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-5 text-sm">
          <div>
            <div className="text-[0.7rem] font-bold uppercase tracking-wider text-[var(--color-sub)] mb-2">Quick Jump</div>
            <div className="space-y-1">
              <a href="#quickstart" onClick={() => setMobileNavOpen(false)} className="flex items-center gap-2 px-3 py-2 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Terminal size={15} /> 1. SDK Quickstart
              </a>
              <a href="#code-snippets" onClick={() => setMobileNavOpen(false)} className="flex items-center gap-2 px-3 py-2 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Code size={15} /> 2. Code Snippets
              </a>
              <a href="#api-reference" onClick={() => setMobileNavOpen(false)} className="flex items-center gap-2 px-3 py-2 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Key size={15} /> 3. API Reference
              </a>
              <a href="#token-strategies" onClick={() => setMobileNavOpen(false)} className="flex items-center gap-2 px-3 py-2 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Cpu size={15} /> 4. PQC Token Strategies
              </a>
            </div>
          </div>

          <div>
            <div className="text-[0.7rem] font-bold uppercase tracking-wider text-[var(--color-sub)] mb-2">API Endpoints</div>
            <div className="space-y-1 font-mono text-xs">
              {(['discovery', 'jwks', 'authorize', 'token', 'userinfo', 'revoke'] as EndpointKey[]).map((ep) => (
                <button
                  key={ep}
                  onClick={() => { setSelectedEndpoint(ep); setMobileNavOpen(false); }}
                  className={`w-full text-left px-3 py-2 rounded-md transition-colors ${
                    selectedEndpoint === ep
                      ? 'bg-[var(--color-hover)] text-blue-400 font-semibold border-l-2 border-blue-500'
                      : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
                  }`}
                >
                  /{ep === 'discovery' ? '.well-known/openid' : ep === 'jwks' ? '.well-known/jwks.json' : ep}
                </button>
              ))}
            </div>
          </div>
        </div>
      </aside>

      {/* Enterprise Documentation 3-Column Layout */}
      <div className="w-full max-w-[1280px] mx-auto px-4 sm:px-6 py-8 flex gap-8 items-start">
        
        {/* Left Sticky Sidebar Navigation (Desktop/Tablet) */}
        <aside className="hidden md:block w-56 shrink-0 sticky top-[76px] max-h-[calc(100vh-90px)] overflow-y-auto pr-2 space-y-6 text-sm">
          <div>
            <div className="text-[0.7rem] font-bold uppercase tracking-wider text-[var(--color-sub)] mb-2.5">Documentation</div>
            <div className="space-y-1">
              <a href="#quickstart" className="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Terminal size={15} /> SDK Quickstart
              </a>
              <a href="#code-snippets" className="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Code size={15} /> Code Snippets
              </a>
              <a href="#api-reference" className="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Key size={15} /> API Reference
              </a>
              <a href="#token-strategies" className="flex items-center gap-2 px-2.5 py-1.5 rounded-md text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
                <Cpu size={15} /> PQC Strategies
              </a>
            </div>
          </div>

          <div>
            <div className="text-[0.7rem] font-bold uppercase tracking-wider text-[var(--color-sub)] mb-2.5">API Endpoints</div>
            <div className="space-y-1 font-mono text-xs">
              {(['discovery', 'jwks', 'authorize', 'token', 'userinfo', 'revoke'] as EndpointKey[]).map((ep) => (
                <button
                  key={ep}
                  onClick={() => setSelectedEndpoint(ep)}
                  className={`w-full text-left px-2.5 py-1.5 rounded-md transition-colors ${
                    selectedEndpoint === ep
                      ? 'bg-[var(--color-hover)] text-blue-400 font-semibold border-l-2 border-blue-500'
                      : 'text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)]'
                  }`}
                >
                  /{ep === 'discovery' ? '.well-known/openid' : ep === 'jwks' ? '.well-known/jwks.json' : ep}
                </button>
              ))}
            </div>
          </div>
        </aside>

        {/* Main Content Column */}
        <main className="flex-1 min-w-0 space-y-8">
          
          {/* Page Header */}
          <div className="space-y-3 pb-4 border-b border-[var(--color-border)]">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-violet-500/15 border border-violet-500/30 text-violet-300 text-xs font-semibold">
                <Shield size={13} /> NIST FIPS 204 Standard • ML-DSA-44 &amp; Ed25519
              </span>
              <a
                href="https://csrc.nist.gov/pubs/fips/204/final"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-md border border-[var(--color-border)] text-[var(--color-sub)] hover:text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors"
              >
                <ExternalLink size={14} /> FIPS 204 Spec
              </a>
            </div>

            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-[var(--color-txt)]">
              DilithiAuth SDK &amp; Interactive Documentation
            </h1>

            <p className="text-[var(--color-sub)] text-sm sm:text-base leading-relaxed max-w-3xl">
              DilithiAuth is a research-oriented Identity Provider (IdP) designed to mitigate the "Harvest Now, Forge Later" (HNFL) threat. It implements the Authorization Code Flow using hybrid signatures combining classical Ed25519 with NIST-standardized ML-DSA-44 (FIPS 204) to prevent future quantum-enabled token forgery and identity impersonation.
            </p>
          </div>

          {/* Quick Jump Bar */}
          <div className="flex flex-wrap gap-2 text-xs">
            <a href="#quickstart" className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg bg-[var(--color-card)] border border-[var(--color-border)] text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
              <Terminal size={14} className="text-sky-400" /> 1. SDK Quickstart
            </a>
            <a href="#code-snippets" className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg bg-[var(--color-card)] border border-[var(--color-border)] text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
              <Code size={14} className="text-emerald-400" /> 2. Code Snippets
            </a>
            <a href="#api-reference" className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg bg-[var(--color-card)] border border-[var(--color-border)] text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
              <Key size={14} className="text-amber-400" /> 3. API Reference
            </a>
            <a href="#token-strategies" className="inline-flex items-center gap-1.5 px-3 py-2 rounded-lg bg-[var(--color-card)] border border-[var(--color-border)] text-[var(--color-txt)] hover:bg-[var(--color-hover)] transition-colors">
              <Cpu size={14} className="text-violet-400" /> 4. PQC Token Strategies
            </a>
          </div>

          {/* SECTION 1: SDK Quickstart & Language Switcher */}
          <section id="quickstart" className="scroll-mt-24 space-y-4">
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 sm:p-6 space-y-5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h2 className="text-lg font-semibold text-[var(--color-txt)]">1. SDK Installation &amp; Quickstart</h2>
                  <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">
                    Select your primary language stack to get copy-ready package installation commands and manifests.
                  </p>
                </div>
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-emerald-500/15 border border-emerald-500/30 text-emerald-300">
                  5 Languages Supported
                </span>
              </div>

              {/* Language Switcher Tab Bar */}
              <div className="flex gap-2 overflow-x-auto pb-2 border-b border-[var(--color-border)]">
                {(Object.keys(LANGUAGE_DATA) as Language[]).map((langKey) => {
                  const lang = LANGUAGE_DATA[langKey]
                  const isSelected = selectedLang === langKey
                  return (
                    <button
                      key={langKey}
                      onClick={() => setSelectedLang(langKey)}
                      className={`inline-flex items-center gap-2 px-3 py-2 rounded-md text-xs font-semibold transition-colors shrink-0 cursor-pointer ${
                        isSelected
                          ? 'bg-[var(--color-primary)] text-white'
                          : 'bg-transparent text-[var(--color-sub)] border border-[var(--color-border)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)]'
                      }`}
                    >
                      <span className={`px-1.5 py-0.5 rounded text-[0.65rem] uppercase ${isSelected ? 'bg-white/20 text-white' : 'bg-[#1f2937] text-gray-400'}`}>
                        {lang.badge}
                      </span>
                      {lang.name}
                    </button>
                  )
                })}
              </div>

              {/* Installation & Manifest Command Boxes */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <div className="text-xs font-bold uppercase tracking-wider text-[var(--color-sub)]">CLI Package Installation</div>
                  <div className="flex items-center justify-between gap-3 p-3 bg-[#050810] border border-[var(--color-border)] rounded-lg font-mono text-xs text-sky-300 overflow-x-auto">
                    <code className="whitespace-nowrap">{langInfo.installCmd}</code>
                    <button
                      className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors shrink-0 cursor-pointer"
                      onClick={() => copyToClipboard(`install-${langInfo.id}`, langInfo.installCmd)}
                    >
                      {copiedMap[`install-${langInfo.id}`] ? (
                        <>
                          <Check size={13} className="text-emerald-400" />
                          <span className="text-emerald-400 font-semibold">Copied!</span>
                        </>
                      ) : (
                        <>
                          <Copy size={13} />
                          <span>Copy</span>
                        </>
                      )}
                    </button>
                  </div>
                </div>

                <div className="space-y-1.5">
                  <div className="text-xs font-bold uppercase tracking-wider text-[var(--color-sub)]">Manifest File ({langInfo.manifestFile})</div>
                  <div className="flex items-start justify-between gap-3 p-3 bg-[#050810] border border-[var(--color-border)] rounded-lg font-mono text-xs text-gray-400 overflow-x-auto">
                    <code className="whitespace-pre">{langInfo.manifestContent}</code>
                    <button
                      className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors shrink-0 cursor-pointer"
                      onClick={() => copyToClipboard(`manifest-${langInfo.id}`, langInfo.manifestContent)}
                    >
                      {copiedMap[`manifest-${langInfo.id}`] ? (
                        <>
                          <Check size={13} className="text-emerald-400" />
                          <span className="text-emerald-400 font-semibold">Copied!</span>
                        </>
                      ) : (
                        <>
                          <Copy size={13} />
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
          <section id="code-snippets" className="scroll-mt-24 space-y-4">
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 sm:p-6 space-y-5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h2 className="text-lg font-semibold text-[var(--color-txt)]">2. Code Challenge &amp; Integration Snippets</h2>
                  <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">
                    Authentic PKCE S256 code challenge generation, OAuth2 authorization, token exchange, and userinfo claims for {langInfo.name}.
                  </p>
                </div>
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-violet-500/15 border border-violet-500/30 text-violet-300">
                  <Shield size={12} /> PKCE S256 Mandated
                </span>
              </div>

              {/* Snippet Stage Sub-tabs */}
              <div className="flex gap-2 overflow-x-auto pb-1 border-b border-[var(--color-border)]">
                {[
                  { id: 'pkce', label: '1. PKCE S256 Challenge', icon: Key },
                  { id: 'init', label: '2. Client Init & Auth URL', icon: Code },
                  { id: 'exchange', label: '3. Code Exchange & PQC Token', icon: Cpu },
                  { id: 'userinfo', label: '4. UserInfo Retrieval', icon: BookOpen }
                ].map(({ id, label, icon: Icon }) => (
                  <button
                    key={id}
                    className={`inline-flex items-center gap-1.5 px-3 py-2 rounded-md text-xs font-semibold transition-colors shrink-0 cursor-pointer ${
                      activeCodeStage === id
                        ? 'bg-[var(--color-primary)] text-white'
                        : 'bg-transparent text-[var(--color-sub)] border border-[var(--color-border)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)]'
                    }`}
                    onClick={() => setActiveCodeStage(id as any)}
                  >
                    <Icon size={14} /> {label}
                  </button>
                ))}
              </div>

              {/* Active Code Block Card */}
              <div className="border border-[var(--color-border)] rounded-lg overflow-hidden bg-[#050810]">
                <div className="flex items-center justify-between px-4 py-2.5 bg-[#0b0f19] border-b border-[var(--color-border)] text-xs font-mono text-gray-400">
                  <span>
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
                    className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold rounded border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors shrink-0 cursor-pointer"
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
                        <Check size={13} className="text-emerald-400" />
                        <span className="text-emerald-400 font-semibold">Copied Code!</span>
                      </>
                    ) : (
                      <>
                        <Copy size={13} />
                        <span>Copy Code</span>
                      </>
                    )}
                  </button>
                </div>

                <pre className="p-4 font-mono text-xs text-slate-200 overflow-x-auto max-h-[380px] leading-relaxed">
                  {activeCodeStage === 'pkce' && langInfo.pkceCode}
                  {activeCodeStage === 'init' && langInfo.clientInitCode}
                  {activeCodeStage === 'exchange' && langInfo.tokenExchangeCode}
                  {activeCodeStage === 'userinfo' && langInfo.userinfoCode}
                </pre>
              </div>
            </div>
          </section>

          {/* SECTION 3: Interactive API & Endpoint Reference */}
          <section id="api-reference" className="scroll-mt-24 space-y-4">
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 sm:p-6 space-y-5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h2 className="text-lg font-semibold text-[var(--color-txt)]">3. Interactive API &amp; Post-Quantum Identity Reference</h2>
                  <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">
                    Test endpoints, configure live parameters (Client ID, Redirect URI, Token Strategy), and inspect exact response formats.
                  </p>
                </div>
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-cyan-500/15 border border-cyan-500/30 text-cyan-300">
                  Live Interactive Playground
                </span>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Endpoint List Sidebar */}
                <div className="lg:col-span-4 space-y-2">
                  <div className="text-xs font-bold uppercase tracking-wider text-[var(--color-sub)] mb-2">DilithiAuth Endpoints</div>

                  {[
                    { key: 'discovery', method: 'GET', path: '/.well-known/openid', badge: 'bg-emerald-500/15 text-emerald-400' },
                    { key: 'jwks', method: 'GET', path: '/.well-known/jwks.json', badge: 'bg-emerald-500/15 text-emerald-400' },
                    { key: 'authorize', method: 'GET', path: '/authorize (PKCE)', badge: 'bg-emerald-500/15 text-emerald-400' },
                    { key: 'token', method: 'POST', path: '/token', badge: 'bg-cyan-500/15 text-cyan-400' },
                    { key: 'userinfo', method: 'GET', path: '/userinfo', badge: 'bg-emerald-500/15 text-emerald-400' },
                    { key: 'revoke', method: 'POST', path: '/revoke', badge: 'bg-cyan-500/15 text-cyan-400' },
                  ].map(({ key, method, path, badge }) => (
                    <button
                      key={key}
                      onClick={() => setSelectedEndpoint(key as EndpointKey)}
                      className={`w-full flex items-center gap-2 px-3 py-2 rounded-lg text-xs font-mono text-left transition-colors cursor-pointer border ${
                        selectedEndpoint === key
                          ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
                          : 'bg-transparent text-[var(--color-sub)] border-[var(--color-border)] hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)]'
                      }`}
                    >
                      <span className={`px-1.5 py-0.5 rounded text-[0.65rem] font-bold ${badge}`}>
                        {method}
                      </span>
                      <span className="truncate">{path}</span>
                    </button>
                  ))}
                </div>

                {/* Endpoint Tester & Inspector */}
                <div className="lg:col-span-8 space-y-4 min-w-0">
                  {/* Interactive Controls Header */}
                  <div className="p-4 bg-[#080d19] border border-[var(--color-border)] rounded-lg space-y-3">
                    <div className="text-xs font-bold uppercase tracking-wider text-[var(--color-sub)]">Interactive Parameter Inputs</div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                      <div>
                        <label className="block text-[var(--color-sub)] mb-1 font-medium">Client ID</label>
                        <input
                          type="text"
                          value={clientId}
                          onChange={(e) => setClientId(e.target.value)}
                          placeholder="Client ID"
                          className="w-full px-2.5 py-1.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded text-[var(--color-txt)] focus:outline-none focus:border-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-[var(--color-sub)] mb-1 font-medium">Redirect URI</label>
                        <input
                          type="text"
                          value={redirectUri}
                          onChange={(e) => setRedirectUri(e.target.value)}
                          placeholder="Redirect URI"
                          className="w-full px-2.5 py-1.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded text-[var(--color-txt)] focus:outline-none focus:border-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-[var(--color-sub)] mb-1 font-medium">User Subject (sub)</label>
                        <input
                          type="text"
                          value={subject}
                          onChange={(e) => setSubject(e.target.value)}
                          placeholder="Subject email/ID"
                          className="w-full px-2.5 py-1.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded text-[var(--color-txt)] focus:outline-none focus:border-blue-500"
                        />
                      </div>
                      <div>
                        <label className="block text-[var(--color-sub)] mb-1 font-medium">Token Strategy</label>
                        <select
                          value={tokenStrategy}
                          onChange={(e) => setTokenStrategy(e.target.value as TokenStrategy)}
                          className="w-full px-2.5 py-1.5 bg-[var(--color-input)] border border-[var(--color-border)] rounded text-[var(--color-txt)] focus:outline-none focus:border-blue-500"
                        >
                          <option value="thin">Thin Token (Ed25519 + pqc_ref)</option>
                          <option value="full">Jumbo Token (Embedded ML-DSA-44)</option>
                          <option value="cwt">CBOR Web Token (CWT RFC 8392)</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  {/* Response Preview Header */}
                  <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-[var(--color-border)]">
                    <div className="flex items-center gap-2 flex-wrap min-w-0">
                      <span className={`px-2 py-0.5 rounded text-xs font-bold font-mono ${endpointRes.method === 'GET' ? 'bg-emerald-500/15 text-emerald-400' : 'bg-cyan-500/15 text-cyan-400'}`}>
                        {endpointRes.method}
                      </span>
                      <span className="font-mono text-xs text-slate-100 break-all">
                        {endpointRes.url}
                      </span>
                      <span className="px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 shrink-0">
                        {endpointRes.status} {endpointRes.statusText}
                      </span>
                    </div>

                    <div className="flex gap-1 shrink-0">
                      {(['body', 'headers', 'metrics'] as const).map((t) => (
                        <button
                          key={t}
                          className={`px-2.5 py-1 text-xs font-semibold rounded transition-colors cursor-pointer ${
                            endpointTab === t
                              ? 'bg-[var(--color-primary)] text-white'
                              : 'bg-transparent text-[var(--color-sub)] border border-[var(--color-border)] hover:bg-[var(--color-hover)]'
                          }`}
                          onClick={() => setEndpointTab(t)}
                        >
                          {t === 'body' ? 'Body' : t === 'headers' ? 'Headers' : 'PQC Metrics'}
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Inspector Output Panels */}
                  {endpointTab === 'body' && (
                    <div className="border border-[var(--color-border)] rounded-lg overflow-hidden bg-[#050810]">
                      <div className="flex items-center justify-between px-4 py-2 bg-[#0b0f19] border-b border-[var(--color-border)] text-xs font-mono text-gray-400">
                        <span>Response Body (JSON)</span>
                        <button
                          className="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-semibold rounded border border-[var(--color-border)] text-[var(--color-sub)] bg-transparent hover:bg-[var(--color-hover)] hover:text-[var(--color-txt)] transition-colors cursor-pointer"
                          onClick={() =>
                            copyToClipboard(`res-body-${selectedEndpoint}-${tokenStrategy}`, JSON.stringify(endpointRes.body, null, 2))
                          }
                        >
                          {copiedMap[`res-body-${selectedEndpoint}-${tokenStrategy}`] ? (
                            <>
                              <Check size={13} className="text-emerald-400" />
                              <span className="text-emerald-400">Copied</span>
                            </>
                          ) : (
                            <>
                              <Copy size={13} />
                              <span>Copy Response</span>
                            </>
                          )}
                        </button>
                      </div>
                      <pre className="p-4 font-mono text-xs text-slate-200 overflow-x-auto max-h-[340px] leading-relaxed">
                        {JSON.stringify(endpointRes.body, null, 2)}
                      </pre>
                    </div>
                  )}

                  {endpointTab === 'headers' && (
                    <div className="border border-[var(--color-border)] rounded-lg overflow-hidden bg-[#050810]">
                      <div className="px-4 py-2 bg-[#0b0f19] border-b border-[var(--color-border)] text-xs font-mono text-gray-400">
                        Response Headers
                      </div>
                      <pre className="p-4 font-mono text-xs text-slate-200 overflow-x-auto max-h-[340px] leading-relaxed">
                        {Object.entries(endpointRes.headers)
                          .map(([k, v]) => `${k}: ${v}`)
                          .join('\n')}
                      </pre>
                    </div>
                  )}

                  {endpointTab === 'metrics' && (
                    <div className="p-4 bg-[#060911] border border-[var(--color-border)] rounded-lg space-y-3">
                      <h3 className="text-xs font-bold text-sky-400 uppercase tracking-wider">
                        Post-Quantum Signature Specification Breakdown
                      </h3>
                      <ul className="space-y-2 text-xs text-[var(--color-sub)] leading-relaxed list-disc list-inside">
                        <li>
                          <strong>Ed25519 Classical Signature Size:</strong> <span className="font-mono text-slate-200">64 Bytes</span>
                        </li>
                        <li>
                          <strong>NIST ML-DSA-44 Signature Size:</strong> <span className="font-mono text-slate-200">2,420 Bytes</span>
                        </li>
                        <li>
                          <strong>Selected Token Mode ({tokenStrategy.toUpperCase()}):</strong>{' '}
                          {tokenStrategy === 'thin' && (
                            <span>Standard JWT (~500 Bytes) + SigStore <span className="font-mono text-emerald-400">pqc_ref</span> Hash</span>
                          )}
                          {tokenStrategy === 'full' && (
                            <span>Jumbo Embedded JWT (~3,840 Bytes) with full <span className="font-mono text-amber-400">pqc_sig</span> claim</span>
                          )}
                          {tokenStrategy === 'cwt' && (
                            <span>Binary CBOR Web Token (~284 Bytes RFC 8392 format)</span>
                          )}
                        </li>
                        <li>
                          <strong>DilithiAuth Server Proxy Configuration:</strong>{' '}
                          <span className="font-mono text-slate-200">MaxHeaderBytes = 16 KiB (16,384 Bytes)</span>
                        </li>
                      </ul>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </section>

          {/* SECTION 4: Post-Quantum Token Architecture Comparison */}
          <section id="token-strategies" className="scroll-mt-24 space-y-4">
            <div className="bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-5 sm:p-6 space-y-5">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <h2 className="text-lg font-semibold text-[var(--color-txt)]">4. Post-Quantum Token Strategy Architecture</h2>
                  <p className="text-[var(--color-sub)] text-xs sm:text-sm mt-0.5">
                    Solving the NIST ML-DSA-44 2.4 KB "token bloat" problem through modular token delivery mechanisms.
                  </p>
                </div>
                <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-violet-500/15 border border-violet-500/30 text-violet-300">
                  <Cpu size={12} /> 16 KiB Header Engine
                </span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {/* Thin Token Card */}
                <div className="p-4 rounded-xl bg-[var(--color-card)] border border-[var(--color-border)] space-y-3 hover:border-blue-500/50 transition-all">
                  <div className="w-10 h-10 rounded-lg bg-emerald-500/15 border border-emerald-500/30 text-emerald-400 flex items-center justify-center">
                    <Layers size={20} />
                  </div>
                  <h3 className="font-semibold text-sm text-[var(--color-txt)]">1. Thin Tokens (Recommended Default)</h3>
                  <p className="text-xs text-[var(--color-sub)] leading-relaxed">
                    JWT contains classical Ed25519 signature and a cryptographic SHA-256 hash reference (<span className="font-mono text-emerald-400">pqc_ref</span>)
                    pointing to the ML-DSA-44 signature stored out-of-band in SigStore (Redis/Memory).
                  </p>
                  <div className="pt-2 border-t border-[var(--color-border)] text-xs font-mono text-emerald-400">
                    ~500 Bytes • Zero Proxy Header Bloat
                  </div>
                </div>

                {/* Jumbo Token Card */}
                <div className="p-4 rounded-xl bg-[var(--color-card)] border border-[var(--color-border)] space-y-3 hover:border-blue-500/50 transition-all">
                  <div className="w-10 h-10 rounded-lg bg-amber-500/15 border border-amber-500/30 text-amber-400 flex items-center justify-center">
                    <Key size={20} />
                  </div>
                  <h3 className="font-semibold text-sm text-[var(--color-txt)]">2. Jumbo Tokens (Full Embedded)</h3>
                  <p className="text-xs text-[var(--color-sub)] leading-relaxed">
                    JWT directly embeds the entire ~2.4 KB ML-DSA-44 signature inside the <span className="font-mono text-amber-400">pqc_sig</span> claim.
                    Fully self-contained for offline verification without cache lookups.
                  </p>
                  <div className="pt-2 border-t border-[var(--color-border)] text-xs font-mono text-amber-400">
                    ~4.0 KB • MaxHeaderBytes = 16 KiB
                  </div>
                </div>

                {/* CBOR Web Token Card */}
                <div className="p-4 rounded-xl bg-[var(--color-card)] border border-[var(--color-border)] space-y-3 hover:border-blue-500/50 transition-all">
                  <div className="w-10 h-10 rounded-lg bg-sky-500/15 border border-sky-500/30 text-sky-400 flex items-center justify-center">
                    <ChevronRight size={20} />
                  </div>
                  <h3 className="font-semibold text-sm text-[var(--color-txt)]">3. CBOR Web Tokens (CWT RFC 8392)</h3>
                  <p className="text-xs text-[var(--color-sub)] leading-relaxed">
                    Binary encoded token using CBOR integer claim mapping (1:iss, 2:sub, 4:exp, 8:pqc_ref).
                    Delivers 25-30% footprint reduction for low-bandwidth mobile and IoT networks.
                  </p>
                  <div className="pt-2 border-t border-[var(--color-border)] text-xs font-mono text-sky-400">
                    ~280 Bytes • Binary Compact Format
                  </div>
                </div>
              </div>

              {/* Architecture Footer Callout */}
              <div className="p-4 rounded-lg bg-blue-500/10 border border-blue-500/30 text-blue-300 text-xs sm:text-sm flex items-start gap-3 leading-relaxed">
                <Shield size={18} className="shrink-0 mt-0.5" />
                <div>
                  <strong>Reverse Proxy Configuration Note:</strong> When operating Jumbo Tokens in production, reverse proxies (Nginx, Envoy, Cloudflare) must be configured with <span className="font-mono bg-blue-950 px-1.5 py-0.5 rounded text-blue-200">large_client_header_buffers 4 16k;</span> to accommodate ~4 KB Authorization header tokens.
                </div>
              </div>
            </div>
          </section>
        </main>

        {/* Right Sticky Table of Contents (Desktop ≥1024px) */}
        <aside className="hidden lg:block w-48 shrink-0 sticky top-[76px] space-y-2 text-xs border-l border-[var(--color-border)] pl-4">
          <div className="font-bold text-[var(--color-sub)] uppercase tracking-wider mb-2">On This Page</div>
          <a href="#quickstart" className="block text-[var(--color-sub)] hover:text-blue-400 transition-colors py-1">
            1. SDK Quickstart
          </a>
          <a href="#code-snippets" className="block text-[var(--color-sub)] hover:text-blue-400 transition-colors py-1">
            2. Code Snippets
          </a>
          <a href="#api-reference" className="block text-[var(--color-sub)] hover:text-blue-400 transition-colors py-1">
            3. API Reference
          </a>
          <a href="#token-strategies" className="block text-[var(--color-sub)] hover:text-blue-400 transition-colors py-1">
            4. PQC Token Strategies
          </a>
        </aside>

      </div>
    </div>
  )
}
