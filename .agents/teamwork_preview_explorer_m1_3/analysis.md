# DilithiAuth Interactive SDK Documentation Page: Technical Exploration & UI Mechanics Report (M1 - Explorer 3)

**Author:** Explorer 3 (M1)  
**Target Component:** `frontend/src/pages/Docs.tsx`  
**Date:** 2026-07-27  
**Working Directory:** `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_3`  

---

## 1. Executive Summary

This report delivers the technical specifications, cryptographic algorithms, packaging matrices, and React component state strategies required to build the interactive **DilithiAuth SDK Documentation Page** (`frontend/src/pages/Docs.tsx`).

DilithiAuth is a Post-Quantum Identity Provider (IdP) implementing NIST-standardized **ML-DSA-44** (FIPS 204) hybrid dual-signatures alongside classical **Ed25519**. Providing developers with clear, copy-pasteable, multi-language code snippets and an interactive testing environment is essential to lower the barrier to adopting quantum-safe identity primitives.

### Key Investigation Outputs:
1. **PKCE S256 Algorithms:** Complete 4-step generation workflows and copy-ready snippets for **TypeScript**, **Go**, **Python**, **Java**, and **.NET (C#)** verified against the existing `sdk/` source directory.
2. **Quickstart Installation Commands:** Package registry names and single-command install scripts for `npm`, `pip`, `go get`, `mvn`/`pom.xml`, and `dotnet add package`.
3. **React UI Architecture:** Complete design and state management strategies for interactive language switching tabs, copy-to-clipboard visual feedback, and a live API Endpoint Inspector with query/response parameter controls and quantum signature metrics visualization.

---

## 2. PKCE S256 Code Challenge & Verifier Generation Algorithms

Proof Key for Code Exchange (PKCE, RFC 7636) prevents authorization code injection attacks on public clients and confidential applications. DilithiAuth enforces PKCE with the `S256` challenge method across all client integrations.

### Cryptographic Specification (RFC 7636)
1. **Code Verifier (`code_verifier`):**
   - High-entropy cryptographic random byte array (typically 32 bytes to 64 bytes).
   - Base64URL-encoded (RFC 4648 §5) without padding (`=`), yielding a string length between 43 and 128 unreserved characters (`[A-Z]`, `[a-z]`, `[0-9]`, `-`, `.`, `_`, `~`).
2. **Code Challenge (`code_challenge`):**
   - `code_challenge = BASE64URL-ENCODE( SHA-256( ASCII(code_verifier) ) )`
   - All trailing `=` padding characters **MUST** be trimmed.
3. **Code Challenge Method (`code_challenge_method`):**
   - **MUST** be `"S256"`.

---

### Step-by-Step Per-Language Implementations

#### 2.1 TypeScript / JavaScript

**Algorithm Steps:**
1. Generate 32 bytes of cryptographically secure random values using `crypto.getRandomValues(new Uint8Array(32))`.
2. Base64URL-encode the random buffer by replacing `+` with `-`, `/` with `_`, and trimming `=`.
3. Digest the UTF-8 encoded verifier string using Web Crypto API `crypto.subtle.digest('SHA-256', data)` (or Node.js `crypto.createHash('sha256')`).
4. Base64URL-encode the SHA-256 hash output without padding.

**SDK Package:** `dilithiauth-js` (Location: `sdk/js/src/pkce.ts`)

**Code Snippet for `Docs.tsx`:**
```typescript
import { generateCodeVerifier, generateCodeChallenge } from 'dilithiauth-js';

// 1. Generate high-entropy Code Verifier (43-128 chars)
const codeVerifier = generateCodeVerifier(64);

// 2. Compute S256 Code Challenge asynchronously
const codeChallenge = await generateCodeChallenge(codeVerifier);

console.log('Verifier:', codeVerifier);
console.log('Challenge:', codeChallenge);

// Standalone Web Crypto API Implementation (No Dependencies)
async function createPkcePair() {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  const verifier = btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const hash = await crypto.subtle.digest('SHA-256', data);
  const challenge = btoa(String.fromCharCode(...new Uint8Array(hash)))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');

  return { verifier, challenge };
}
```

---

#### 2.2 Go

**Algorithm Steps:**
1. Allocate a 32-byte slice and populate with `crypto/rand.Read(bytes)`.
2. Encode to Base64URL using `base64.RawURLEncoding.EncodeToString(bytes)`.
3. Compute SHA-256 hash using `crypto/sha256.Sum256([]byte(verifier))`.
4. Encode the 32-byte hash slice using `base64.RawURLEncoding.EncodeToString(hash[:])`.

**SDK Module:** `github.com/Uttam-Mahata/DilithiAuth/sdk/go` (Location: `sdk/go/client.go`)

**Code Snippet for `Docs.tsx`:**
```go
package main

import (
	"fmt"
	"log"
	"github.com/Uttam-Mahata/DilithiAuth/sdk/go"
)

func main() {
	// 1. Generate PKCE Verifier and Challenge using DilithiAuth SDK
	verifier, challenge, err := dilithiauth.GeneratePKCEPair()
	if err != nil {
		log.Fatalf("Failed to generate PKCE pair: %v", err)
	}

	fmt.Printf("Code Verifier:  %s\n", verifier)
	fmt.Printf("Code Challenge: %s\n", challenge)
}

/* Standalone Go Standard Library Implementation:
import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GeneratePKCE() (verifier, challenge string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}
*/
```

---

#### 2.3 Python

**Algorithm Steps:**
1. Generate secure random bytes using `secrets.token_bytes(32)`.
2. Encode bytes using `base64.urlsafe_b64encode(raw).decode('utf-8').rstrip('=')`.
3. Compute SHA-256 digest of verifier string using `hashlib.sha256(verifier.encode('utf-8')).digest()`.
4. Base64URL-encode digest and strip trailing `=`.

**SDK Package:** `dilithiauth` (Location: `sdk/python/dilithiauth/pkce.py`)

**Code Snippet for `Docs.tsx`:**
```python
from dilithiauth.pkce import generate_pkce_pair, generate_code_verifier, generate_code_challenge

# 1. Generate complete PKCE Pair (verifier, challenge)
code_verifier, code_challenge = generate_pkce_pair()

print(f"Code Verifier:  {code_verifier}")
print(f"Code Challenge: {code_challenge}")

"""
Standalone Standard Library Implementation (Python 3.8+):
import secrets
import base64
import hashlib

def create_pkce_pair():
    raw_bytes = secrets.token_bytes(32)
    verifier = base64.urlsafe_b64encode(raw_bytes).decode('utf-8').rstrip('=')
    
    digest = hashlib.sha256(verifier.encode('utf-8')).digest()
    challenge = base64.urlsafe_b64encode(digest).decode('utf-8').rstrip('=')
    
    return verifier, challenge
"""
```

---

#### 2.4 Java

**Algorithm Steps:**
1. Populate 32-byte array using `java.security.SecureRandom`.
2. Encode using `Base64.getUrlEncoder().withoutPadding().encodeToString(bytes)`.
3. Compute hash using `MessageDigest.getInstance("SHA-256").digest(verifier.getBytes(StandardCharsets.UTF_8))`.
4. Base64URL encode hash without padding.

**SDK Package:** `com.dilithiauth:dilithiauth-sdk` (Location: `sdk/java/src/main/java/com/dilithiauth/client/PKCE.java`)

**Code Snippet for `Docs.tsx`:**
```java
import com.dilithiauth.client.PKCE;

public class App {
    public static void main(String[] args) {
        // 1. Generate PKCE Pair
        PKCE.Pair pkcePair = PKCE.generatePair();
        
        System.out.println("Code Verifier:  " + pkcePair.getCodeVerifier());
        System.out.println("Code Challenge: " + pkcePair.getCodeChallenge());
    }
}

/* Standalone Java 11+ Implementation:
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.SecureRandom;
import java.util.Base64;

public class PkceUtil {
    public static String[] generatePkce() throws Exception {
        SecureRandom random = new SecureRandom();
        byte[] bytes = new byte[32];
        random.nextBytes(bytes);
        String verifier = Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);

        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        byte[] hash = digest.digest(verifier.getBytes(StandardCharsets.UTF_8));
        String challenge = Base64.getUrlEncoder().withoutPadding().encodeToString(hash);

        return new String[]{verifier, challenge};
    }
}
*/
```

---

#### 2.5 .NET (C#)

**Algorithm Steps:**
1. Generate random bytes using `RandomNumberGenerator.Create().GetBytes(bytes)`.
2. Base64URL encode by replacing `+` with `-`, `/` with `_`, and trimming `=`.
3. Hash using `SHA256.Create().ComputeHash(Encoding.UTF8.GetBytes(verifier))`.
4. Base64URL encode the byte hash array.

**SDK Package:** `DilithiAuth.Client` (Location: `sdk/dotnet/DilithiAuth.Client/Pkce.cs`)

**Code Snippet for `Docs.tsx`:**
```csharp
using DilithiAuth.Client;

// 1. Generate PKCE Verifier & Challenge Pair
var pkcePair = Pkce.GeneratePair();

Console.WriteLine($"Code Verifier:  {pkcePair.CodeVerifier}");
Console.WriteLine($"Code Challenge: {pkcePair.CodeChallenge}");

/* Standalone C# .NET 8.0 Implementation:
using System;
using System.Security.Cryptography;
using System.Text;

public static class PkceHelper {
    public static (string Verifier, string Challenge) GeneratePkce() {
        byte[] bytes = new byte[32];
        using (var rng = RandomNumberGenerator.Create()) {
            rng.GetBytes(bytes);
        }
        string verifier = Base64Url(bytes);

        using var sha256 = SHA256.Create();
        byte[] hash = sha256.ComputeHash(Encoding.UTF8.GetBytes(verifier));
        string challenge = Base64Url(hash);

        return (verifier, challenge);
    }

    private static string Base64Url(byte[] buffer) =>
        Convert.ToBase64String(buffer).Replace("+", "-").Replace("/", "_").TrimEnd('=');
}
*/
```

---

## 3. Quickstart Installation Commands & Packaging Matrix

Below is the verified installation matrix for all 5 enterprise SDK targets.

| Language / Framework | Package Name | Registry / Repository | CLI Installation Command | Manifest Configuration File |
| :--- | :--- | :--- | :--- | :--- |
| **TypeScript / Node.js** | `dilithiauth-js` | npm Registry | `npm install dilithiauth-js` | `package.json` |
| **Python** | `dilithiauth` | PyPI | `pip install dilithiauth` | `pyproject.toml` / `requirements.txt` |
| **Go** | `sdk/go` | Go Modules | `go get github.com/Uttam-Mahata/DilithiAuth/sdk/go` | `go.mod` |
| **Java (Maven)** | `dilithiauth-sdk` | Maven Central | `mvn dependency:add -Dartifact=com.dilithiauth:dilithiauth-sdk:1.0.0` | `pom.xml` |
| **.NET (C#)** | `DilithiAuth.Client` | NuGet | `dotnet add package DilithiAuth.Client` | `.csproj` |

---

### Detailed Package Manifest Snippets

#### 3.1 TypeScript / JavaScript (`package.json`)
```json
{
  "dependencies": {
    "dilithiauth-js": "^1.0.0"
  }
}
```

#### 3.2 Python (`pyproject.toml` / `requirements.txt`)
```toml
# pyproject.toml
[project]
dependencies = [
    "dilithiauth>=1.0.0"
]
```
```text
# requirements.txt
dilithiauth>=1.0.0
```

#### 3.3 Go (`go.mod`)
```go
module myapp

go 1.24

require github.com/Uttam-Mahata/DilithiAuth/sdk/go v1.0.0
```

#### 3.4 Java (`pom.xml` / `build.gradle`)
```xml
<!-- Maven pom.xml -->
<dependency>
    <groupId>com.dilithiauth</groupId>
    <artifactId>dilithiauth-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```
```groovy
// Gradle build.gradle
implementation 'com.dilithiauth:dilithiauth-sdk:1.0.0'
```

#### 3.5 .NET (`.csproj`)
```xml
<ItemGroup>
  <PackageReference Include="DilithiAuth.Client" Version="1.0.0" />
</ItemGroup>
```

---

## 4. React UI Implementation Strategies for `Docs.tsx`

To deliver an interactive, enterprise-grade documentation page matching DilithiAuth's existing dark-mode UI aesthetic (`var(--bg-dark)`, `var(--bg-card)`, Lucide icons), `frontend/src/pages/Docs.tsx` should be structured around three primary interactive UI mechanics.

---

### 4.1 Interactive Language Switching Tabs Strategy

#### Data Model & Type Definitions:
```typescript
export type SupportedLanguage = 'ts' | 'go' | 'python' | 'java' | 'dotnet';

export interface LanguageConfig {
  id: SupportedLanguage;
  name: string;
  badge: string;
  iconName: string;
  installCmd: string;
  altInstallCmd?: string;
  manifestSnippet: string;
  pkceCode: string;
  authFlowCode: string;
  tokenExchangeCode: string;
}
```

#### State Management:
```typescript
const [activeLang, setActiveLang] = useState<SupportedLanguage>('ts');
```

#### Component Layout Strategy:
- **Top Tab Navigation Bar:** Render a horizontal flex container with pill buttons styled via `.btn` and `.btn-outline` / `.btn-primary`.
- **Active State Highlighting:** Apply active border and glow (`border-color: var(--primary)`).
- **Tab Header Bar with Copy Button:** Display the file extension or installation type header above each code block.

```tsx
<div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1.5rem', overflowX: 'auto', paddingBottom: '0.5rem' }}>
  {languages.map((lang) => (
    <button
      key={lang.id}
      onClick={() => setActiveLang(lang.id)}
      className={`btn ${activeLang === lang.id ? 'btn-primary' : 'btn-outline'}`}
      style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', padding: '0.5rem 1rem' }}
    >
      <span className="badge" style={{ fontSize: '0.7rem', padding: '2px 6px' }}>{lang.badge}</span>
      {lang.name}
    </button>
  ))}
</div>
```

---

### 4.2 Copy-to-Clipboard State with Visual Feedback Strategy

#### Utility Hook / State Model:
Maintain a map of copied block IDs to prevent cross-button interference when copying different code blocks on the page.

```typescript
const [copiedState, setCopiedState] = useState<Record<string, boolean>>({});

const copyToClipboard = (blockId: string, text: string) => {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text);
  } else {
    // Fallback for non-HTTPS or legacy contexts
    const textArea = document.createElement('textarea');
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.select();
    document.execCommand('copy');
    document.body.removeChild(textArea);
  }

  setCopiedState(prev => ({ ...prev, [blockId]: true }));
  setTimeout(() => {
    setCopiedState(prev => ({ ...prev, [blockId]: false }));
  }, 2000);
};
```

#### Visual Feedback Rendering:
Display `lucide-react` icons `<Copy size={14} />` vs `<Check size={14} color="#10b981" />`.

```tsx
<button
  className="btn btn-outline btn-sm"
  onClick={() => copyToClipboard('install-cmd', currentLang.installCmd)}
  style={{ gap: '0.4rem', fontSize: '0.8rem' }}
>
  {copiedState['install-cmd'] ? (
    <>
      <Check size={14} color="#10b981" />
      <span style={{ color: '#10b981', fontWeight: 600 }}>Copied!</span>
    </>
  ) : (
    <>
      <Copy size={14} />
      <span>Copy Command</span>
    </>
  )}
</button>
```

---

### 4.3 Endpoint Query / Response Toggle UI (Interactive API Reference)

The Interactive API Reference section lets developers test and inspect live or simulated DilithiAuth IdP endpoints directly inside `/docs`.

#### Endpoints Covered:
1. `GET /.well-known/openid-configuration` — OIDC Discovery Document
2. `GET /.well-known/jwks.json` — Public Keys (includes `mldsa44_pub`)
3. `GET /authorize` — OAuth 2.0 PKCE Authorization Endpoint
4. `POST /token` — Token Endpoint (Supports Thin, Jumbo, and CWT token modes)
5. `GET /userinfo` — OIDC UserInfo Endpoint

#### Interactive Control & Simulation State:
```typescript
type EndpointKey = 'discovery' | 'jwks' | 'authorize' | 'token' | 'userinfo';
type ResponseTab = 'body' | 'headers' | 'metrics';

const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointKey>('discovery');
const [responseTab, setResponseTab] = useState<ResponseTab>('body');
const [subjectParam, setSubjectParam] = useState<string>('alice@example.com');
const [tokenMode, setTokenMode] = useState<'thin' | 'full' | 'cwt'>('thin');
const [isExecuting, setIsExecuting] = useState<boolean>(false);
```

#### Endpoint Inspector UI Layout:
- **Left Panel (Endpoint Selector List):** Vertical menu listing HTTP method badges (`GET` blue, `POST` green) and endpoint paths.
- **Top Right Panel (Query Controls & Parameter Modifiers):** Input fields for `sub`, `client_id`, `redirect_uri`, and token strategy selector (`Thin Token`, `Jumbo Token`, `CBOR Web Token`).
- **Bottom Right Panel (Response Inspection Tabs):**
  - **Body Tab:** Formatted JSON output of response data (e.g., JWT access token, `pqc_ref`, `pqc_alg`).
  - **Headers Tab:** HTTP headers (e.g. `Content-Type: application/json`, `X-DilithiAuth-PQC-Alg: ML-DSA-44`).
  - **Quantum Signature Metrics Tab:** Breakdown of byte sizes (Ed25519 64B vs ML-DSA-44 2420B) and proxy header requirements (`MaxHeaderBytes = 16 KiB`).

---

## 5. Verification & Downstream Handoff Plan

### 5.1 Verification Checklist for Downstream Implementer (M2/M3)
1. **TypeScript Build Validation:**
   - Execute `npm run build` inside `frontend/` to confirm zero TS or Vite compilation errors.
2. **Interactive UI Verification:**
   - Verify clicking each of the 5 language tabs updates install commands, PKCE code snippets, and token exchange examples immediately.
   - Verify clicking the copy button updates state to "Copied!" with check icon for 2000ms.
   - Verify API Inspector endpoint switching correctly updates request/response preview.

### 5.2 File Path Locations for M2/M3 Implementation:
- **New Page Component:** `frontend/src/pages/Docs.tsx`
- **App Routing Update:** `frontend/src/App.tsx` (`<Route path="/docs" element={<Docs />} />`)
- **Navbar Update:** `frontend/src/components/Navbar.tsx` (Add "SDK & Docs" nav item)

---
*End of M1 Explorer 3 Report.*
