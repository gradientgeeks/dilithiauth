# DilithiAuth Frontend Architecture & SDK Documentation System Analysis

**Milestone:** M1 — DilithiAuth Interactive SDK Documentation Page  
**Role:** Explorer 1  
**Working Directory:** `/home/uttam/projects/DilithiAuth/.agents/teamwork_preview_explorer_m1_1`  
**Date:** 2026-07-27  

---

## Executive Summary

DilithiAuth is a Post-Quantum OAuth 2.0 / OpenID Connect Identity Provider (IdP) implementing NIST FIPS 204 (ML-DSA-44) hybrid signatures alongside classical Ed25519 keys. The frontend application is a modern, single-page React application built with TypeScript, Vite, React Router v7, and custom dark-theme CSS tokens.

This report provides a comprehensive analysis of the existing frontend application structure, installed packages, styling design system, component patterns, and an exact blueprint for the new `frontend/src/pages/Docs.tsx` Interactive SDK Documentation Page.

---

## 1. Package & Build Configuration Analysis

### 1.1 Installed Packages (`frontend/package.json`)

| Package Category | Package Name | Version | Purpose / Usage |
| :--- | :--- | :--- | :--- |
| **Core Framework** | `react` | `^19.2.5` | React 19 UI component framework |
| **Core Framework** | `react-dom` | `^19.2.5` | DOM rendering bindings |
| **Routing** | `react-router-dom` | `^7.14.2` | React Router v7 declarative client-side routing |
| **HTTP Client** | `axios` | `^1.15.2` | REST API requests to backend IdP server |
| **UI Icons** | `lucide-react` | `^1.27.0` | Icon set used extensively across navigation, buttons, badges, and alerts |
| **Build Tooling** | `vite` | `^8.0.10` | ESM-first fast bundler & dev server |
| **TypeScript** | `typescript` | `~6.0.2` | Type safety and JSX compilation |
| **React Plugin** | `@vitejs/plugin-react` | `^6.0.1` | Vite plugin for React fast refresh & JSX |

### 1.2 Syntax Highlighting Libraries
- **Current State:** **None installed** in `dependencies` or `devDependencies`.
- **Implementation Mechanism:** Code blocks are rendered natively using pre-styled `<pre className="code-block">{...}</pre>` elements formatted via CSS design tokens (`font-family: var(--font-mono)`, `background: #060911`, `color: #e5e7eb`, `border: 1px solid #1f2937`, `padding: 0.85rem 1rem`, `border-radius: var(--radius-sm)`).
- **Recommendation:** `Docs.tsx` should follow this existing native monospace block pattern to maintain zero extra dependency overhead and full styling consistency with `AppDetail.tsx` and `QuantumTokenPlayground.tsx`.

### 1.3 React Router DOM Version & Route Structure
- **Version:** `react-router-dom: ^7.14.2`
- **Current Routes (`frontend/src/App.tsx`):**
  - `/`: Public Landing page (`<Landing />` wrapped in `<RequireGuest>`)
  - `/login`: Login page (`<Login />` wrapped in `<RequireGuest>`)
  - `/register`: Registration page (`<Register />` wrapped in `<RequireGuest>`)
  - `/apps`: Developer Dashboard (`<Dashboard />` wrapped in `<RequireAuth>`)
  - `/apps/new`: App Registration form (`<NewApp />` wrapped in `<RequireAuth>`)
  - `/apps/:id`: Application detail page with credentials & code generator (`<AppDetail />` wrapped in `<RequireAuth>`)
  - `*`: Catch-all wildcard fallback redirecting to `/`

### 1.4 Vite Build Configuration (`frontend/vite.config.ts`)
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'https://dilithiauthapis.gradientgeeks.tech',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
```
- API proxy forwards local `/api` calls to `https://dilithiauthapis.gradientgeeks.tech`.
- Production bundle outputs clean static assets into `dist/`.

---

## 2. UI Design System & Visual Style Guide (`frontend/src/index.css`)

### 2.1 CSS Design Tokens & Color Palette

```css
:root {
  --bg-dark: #0b0f19;       /* Main application dark background */
  --bg-card: #111827;       /* Card background */
  --bg-card-hover: #1f2937; /* Card hover fill */
  --bg-input: #111827;      /* Input control base background */
  
  --border-color: #1f2937;  /* Default subtle border */
  --border-focus: #2563eb;  /* Blue focus ring border */
  
  --primary: #2563eb;       /* Royal Blue primary button & link accent */
  --primary-hover: #1d4ed8; /* Darker blue on hover */
  
  --emerald: #059669;       /* Success state green */
  --amber: #d97706;         /* Warning state gold */
  --rose: #dc2626;          /* Danger/error state red */
  
  --text-main: #f9fafb;     /* Primary headings and bold text */
  --text-sub: #9ca3af;      /* Subtitles and body secondary text */
  --text-muted: #6b7280;    /* Captions, form hints, and disabled text */
  
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'JetBrains Mono', monospace;
  
  --radius-sm: 6px;         /* Buttons, inputs, small pills */
  --radius: 8px;            /* Cards, modals, containers */
  
  --shadow-card: 0 1px 3px 0 rgba(0, 0, 0, 0.3), 0 1px 2px 0 rgba(0, 0, 0, 0.2);
}
```

### 2.2 Palette Usage Matrix

| Element | Color Code / Variable | Example Component |
| :--- | :--- | :--- |
| **Page Canvas** | `#0b0f19` (`--bg-dark`) | Document `body` background |
| **Card Surface** | `#111827` (`--bg-card`) | Containers, `.card`, `.modal-card` |
| **Primary Action** | `#2563eb` (`--primary`) | `.btn-primary`, active tab border |
| **Accent Text/Links**| `#38bdf8`, `#60a5fa` | Monospace code text, public badges |
| **Code Container** | `#060911` | `.code-block`, `.secret-box` |
| **Quantum Badge** | `#1e1b4b` bg / `#a5b4fc` text | `.badge-hybrid` |
| **Warning Alert** | `rgba(217, 119, 6, 0.12)` | Secret rotation alert |
| **Success Alert** | `rgba(5, 150, 105, 0.12)` | Copied feedback, status indicators |

### 2.3 Navigation Header Structure (`Navbar.tsx`)
- **Dimensions:** Sticky header, height `56px`, `background: #111827`, `border-bottom: 1px solid #1f2937`, `z-index: 100`.
- **Brand Logo:** `<img src="/favicon.svg" width={28} height={28} />` followed by text `"DilithiAuth"` and `<span className="nav-brand-badge">PQC IdP</span>`.
- **Active Route Highlighting:** Links check `location.pathname === path` and apply `.active` CSS class (`background: #1f2937`, `color: #f9fafb`).
- **Icons:** All menu options use Lucide icons (`15px` size).

### 2.4 Component Style Specifications

#### Cards (`.card`, `.card-glow`, `.feature-card`)
- `.card`: `background: #111827`, `border: 1px solid #1f2937`, `border-radius: 8px`, `padding: 1.5rem`, `margin-bottom: 1rem`.
- `.card-glow`: Highlighted variant with `border-color: #374151`.
- `.feature-card`: `padding: 1.25rem` with integrated `.feature-icon` box (`40x40px`, `background: #1f2937`, `color: #60a5fa`).

#### Tabs Pattern (`.tab-header`, `.tab-btn`)
- `.tab-header`: Flex row with `gap: 0.25rem`, `border-bottom: 1px solid #1f2937`, `margin-bottom: 1.25rem`, horizontally scrollable (`overflow-x: auto`).
- `.tab-btn`: Transparent background, padding `0.65rem 0.9rem`, `color: #9ca3af`, `font-size: 0.875rem`, `border-bottom: 2px solid transparent`.
- `.tab-btn.active`: `color: #f9fafb`, `border-bottom-color: #2563eb`, `font-weight: 600`.

#### Button Styling (`.btn`)
- `.btn`: `display: inline-flex`, `align-items: center`, `justify-content: center`, `gap: 0.4rem`, `padding: 0.5rem 1rem`, `border-radius: 6px`, `font-size: 0.875rem`, `font-weight: 500`.
- `.btn-primary`: Blue background (`#2563eb`), white text.
- `.btn-outline`: Dark card background (`#111827`), border `#1f2937`, hover `#1f2937`.
- `.btn-sm`: Compact padding `0.3rem 0.65rem`, `font-size: 0.8rem`.

#### Code Displays (`.code-block`, `.secret-box`, `.mono`)
- `.code-block`: `background: #060911`, `color: #e5e7eb`, `font-family: JetBrains Mono`, `font-size: 0.825rem`, `padding: 0.85rem 1rem`, `border-radius: 6px`, `border: 1px solid #1f2937`, `white-space: pre`, `overflow-x: auto`.
- `.mono`: Inline code span, `color: #60a5fa`, `background: #060911`, `padding: 0.15rem 0.4rem`, `border-radius: 4px`.

#### Layout Containers
- Main Container (`.container`): `max-width: 980px`, `margin: 0 auto`, `padding: 2rem 1.25rem`.
- Page Header (`.page-header`): Flex row with `h1` main title, subtitle paragraph, and action buttons aligned right.

---

## 3. Analysis of Existing Component Patterns

1. **`Landing.tsx`**: Features hero banner with `.hero-badge`, technical architecture card, grid of feature cards (`.features`), `QuantumTokenPlayground` interactive component, and OIDC discovery endpoints pre-block.
2. **`QuantumTokenPlayground.tsx`**: Demonstrates stateful tab switching (`thin`, `full`, `cwt`), dynamic payload rendering, clipboard copy functionality, interactive subject input, signature size visualizer bar (`Ed25519 64B` vs `ML-DSA-44 2,420B`), and informational alert boxes.
3. **`AppDetail.tsx`**: Demonstrates language selection sub-tabs (`cURL`, `JavaScript`, `Go`, `Python`, `Java`, `.NET`), dynamic multi-language code generator, credentials copy handler with visual check indicator, secret rotation workflow, and tab-based sub-navigation (`overview`, `snippets`, `danger`).

---

## 4. Recommended Component Architecture and Layout for `frontend/src/pages/Docs.tsx`

To create an interactive, comprehensive, and seamless SDK documentation experience that fits perfectly within DilithiAuth's existing UI system, we recommend the following design and component layout architecture.

### 4.1 Route & Navigation Integration

1. **App Route Addition (`App.tsx`):**
   Add `/docs` as a publicly accessible route:
   ```tsx
   <Route path="/docs" element={<Docs />} />
   ```
2. **Navbar Link Addition (`Navbar.tsx`):**
   Add a direct link in both guest and authenticated nav states:
   ```tsx
   <Link to="/docs" className={isActive('/docs') ? 'active' : ''}>
     <BookOpen size={15} /> Documentation
   </Link>
   ```

### 4.2 Page Layout Structure

The `Docs.tsx` page will follow DilithiAuth's `.container` layout (`max-width: 980px`) with a sticky header and an interactive two-column sidebar layout or responsive tabbed section view:

```
+-----------------------------------------------------------------------+
|                              <Navbar />                               |
+-----------------------------------------------------------------------+
| .container                                                            |
|  +-----------------------------------------------------------------+  |
|  | .page-header                                                    |  |
|  | [Badge: NIST FIPS 204 Standard • SDK & Developer Reference]     |  |
|  | <h1>DilithiAuth Interactive SDK Documentation</h1>              |  |
|  | <p>Post-quantum OAuth 2.0 & OIDC integration guide and sandbox.</p>|  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  | .tab-header (Quick Jump Navigation)                             |  |
|  | [ Quickstart ] [ PKCE Flow ] [ Token Strategies ] [ Interactive ]  |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  | SECTION 1: SDK Language & Installation Selection                |  |
|  | [ cURL ] [ JavaScript/TS ] [ Go ] [ Python ] [ Java ] [ .NET ]  |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  | SECTION 2: Interactive SDK Configuration Sandbox               |  |
|  | +-----------------------------+ +-----------------------------+ |  |
|  | | Client ID Input             | | Token Mode (Thin/Jumbo/CWT) | |  |
|  | | Redirect URI Input          | | Scopes checkboxes           | |  |
|  | +-----------------------------+ +-----------------------------+ |  |
|  | [ Live Generated SDK Initialization & Token Exchange Snippet ]  |  |
|  | [ Copy Code Button with Feedback ]                              |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  | SECTION 3: Post-Quantum Token Architecture Deep-Dive            |  |
|  | - Thin Token Hash Reference Strategy                            |  |
|  | - Jumbo Embedded Signature Strategy                             |  |
|  | - CBOR Web Token (CWT) Binary Strategy                          |  |
|  +-----------------------------------------------------------------+  |
|                                                                       |
|  +-----------------------------------------------------------------+  |
|  | SECTION 4: Signature Verification Reference (ML-DSA-44)         |  |
|  | - JWKS Endpoint Fetch (`/.well-known/jwks.json`)                |  |
|  | - Dual Verification (Ed25519 Classical + ML-DSA-44 Post-Quantum)|  |
|  +-----------------------------------------------------------------+  |
+-----------------------------------------------------------------------+
```

### 4.3 Detailed Component Breakdown & State Specs

#### Component 1: `DocsHeader`
- **UI Elements:**
  - Hero Badge: `<span className="hero-badge"><BookOpen size={13} /> NIST FIPS 204 Standard • SDK Reference & Sandbox</span>`
  - Heading: `DilithiAuth SDK Documentation`
  - Description: Clear guide on how to integrate DilithiAuth's post-quantum authentication into JavaScript, Go, Python, Java, .NET, or cURL backends.

#### Component 2: `SDKLanguageTabs`
- **State:** `selectedLang: 'javascript' | 'go' | 'python' | 'java' | 'dotnet' | 'curl'`
- **Buttons:**
  - `JavaScript / TS` (`dilithiauth-js`)
  - `Go` (`github.com/Uttam-Mahata/DilithiAuth/sdk/go`)
  - `Python` (`dilithiauth`)
  - `Java` (`com.dilithiauth:dilithiauth-sdk`)
  - `.NET` (`DilithiAuth.Client`)
  - `cURL / REST API`

#### Component 3: `InteractiveSDKPlayground`
- **Interactive Controls (State):**
  - `clientId`: string (default: `"dilithi-app-client-123"`)
  - `clientSecret`: string (default: `"sec_live_9f8a7b6c5d4e3f2a"`)
  - `isPublic`: boolean (toggle between Public PKCE client & Confidential client)
  - `tokenMode`: `'thin' | 'full' | 'cwt'` (Thin JWT sidecar ref vs Jumbo embedded vs CWT)
  - `redirectUri`: string (default: `"https://myapp.com/callback"`)
  - `scopes`: string[] (default: `['openid', 'profile', 'email']`)
- **Live Output Panel:**
  - Renders customized SDK snippet for the selected language and form parameters.
  - One-click copy button with `Check` icon visual feedback (`2000ms` reset timeout).

#### Component 4: `TokenStrategyExplainer`
- **UI Cards:** Three side-by-side or stacked feature cards illustrating:
  1. **Thin Tokens:** Hash reference `pqc_ref`, Redis sidecar caching, zero proxy header overflow.
  2. **Jumbo Tokens:** Embedded ~2.4 KB ML-DSA-44 signature, self-contained, requires `MaxHeaderBytes = 16 KiB`.
  3. **CBOR Web Tokens (CWT):** Binary RFC 8392 format, 25-30% footprint reduction.

#### Component 5: `VerificationGuide`
- Explanation of hybrid verification logic:
  - Fetching keys from `GET /.well-known/jwks.json` (containing `mldsa44_pub`).
  - Verifying classical Ed25519 signature component.
  - Verifying quantum-safe ML-DSA-44 signature component.

---

## 5. Implementation Code Snippets Specification for SDK Languages

The documentation page will feature verified code snippets for all 6 target environments:

### JavaScript / TypeScript (`dilithiauth-js`)
```typescript
import { DilithiAuthClient } from 'dilithiauth-js';

const auth = new DilithiAuthClient({
  domain: 'https://dilithiauthapis.gradientgeeks.tech',
  clientId: 'CLIENT_ID',
  redirectUri: 'REDIRECT_URI',
  tokenMode: 'thin' // 'thin' | 'full' | 'cwt'
});

// 1. Initiate PKCE login
await auth.loginWithRedirect();

// 2. Process authorization callback
const session = await auth.handleRedirectCallback();
console.log('User Claims:', session.user);
console.log('Post-Quantum Reference:', session.pqcRef);
```

### Go (`github.com/Uttam-Mahata/DilithiAuth/sdk/go`)
```go
package main

import (
    "context"
    "fmt"
    dilithiauth "github.com/Uttam-Mahata/DilithiAuth/sdk/go"
)

func main() {
    client, _ := dilithiauth.NewClient(dilithiauth.Config{
        Domain:      "https://dilithiauthapis.gradientgeeks.tech",
        ClientID:    "CLIENT_ID",
        RedirectURI: "REDIRECT_URI",
    })

    // Exchange authorization code for Hybrid Token
    tok, err := client.ExchangeCode(context.Background(), "AUTH_CODE", "REDIRECT_URI", "VERIFIER")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Access Token: %s\nML-DSA-44 Ref: %s\n", tok.AccessToken, tok.PQCRef)
}
```

### Python (`dilithiauth`)
```python
from dilithiauth import DilithiAuthClient

client = DilithiAuthClient(
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_id="CLIENT_ID",
    redirect_uri="REDIRECT_URI"
)

# Exchange code for Post-Quantum token
token_resp = client.exchange_code(
    code="AUTHORIZATION_CODE",
    code_verifier="PKCE_VERIFIER",
    redirect_uri="REDIRECT_URI"
)
print(f"Logged in user: {token_resp['sub']}")
print(f"Hybrid Algorithm: {token_resp['pqc_alg']}")
```

### Java (`com.dilithiauth:dilithiauth-sdk`)
```java
import com.dilithiauth.client.DilithiAuthClient;
import com.dilithiauth.client.TokenResponse;

DilithiAuthClient client = DilithiAuthClient.builder()
    .domain("https://dilithiauthapis.gradientgeeks.tech")
    .clientId("CLIENT_ID")
    .redirectUri("REDIRECT_URI")
    .build();

TokenResponse token = client.exchangeCode("AUTH_CODE", "VERIFIER", "REDIRECT_URI");
System.out.println("Access Token: " + token.getAccessToken());
```

### .NET / C# (`DilithiAuth.Client`)
```csharp
using DilithiAuth.Client;

var client = new DilithiAuthClient(new DilithiAuthOptions
{
    Domain = "https://dilithiauthapis.gradientgeeks.tech",
    ClientId = "CLIENT_ID",
    RedirectUri = "REDIRECT_URI"
});

TokenResponse token = await client.ExchangeCodeAsync("AUTH_CODE", "VERIFIER", "REDIRECT_URI");
Console.WriteLine($"User Sub: {token.Sub}, Hybrid Sig: {token.PqcAlg}");
```

### cURL / REST API
```bash
# 1. Authorization Redirect URL:
https://dilithiauthapis.gradientgeeks.tech/authorize?client_id=CLIENT_ID&redirect_uri=REDIRECT_URI&response_type=code&scope=openid+profile+email&code_challenge=CHALLENGE&code_challenge_method=S256

# 2. Exchange Authorization Code for Thin Token:
curl -X POST https://dilithiauthapis.gradientgeeks.tech/api/v1/oauth/token \
  -d "grant_type=authorization_code" \
  -d "client_id=CLIENT_ID" \
  -d "code=AUTHORIZATION_CODE" \
  -d "code_verifier=PKCE_VERIFIER" \
  -d "redirect_uri=REDIRECT_URI"
```

---

## 6. Actionable Implementation Checklist for Implementer Agent

- [ ] **Create `frontend/src/pages/Docs.tsx`**: Implement full interactive SDK documentation page matching specified component hierarchy.
- [ ] **Update `frontend/src/App.tsx`**: Register `/docs` route.
- [ ] **Update `frontend/src/components/Navbar.tsx`**: Add `Documentation` navigation item with `BookOpen` icon.
- [ ] **Verify Responsiveness**: Ensure mobile breakdown (`@media (max-width: 640px)`) renders single-column cleanly.
- [ ] **Run Build & Tests**: Verify project compiles cleanly with `npm run build` or `go test ./...`.

---

*Report prepared by Explorer 1 for Milestone M1.*
