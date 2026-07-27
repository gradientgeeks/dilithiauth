# dilithiauth-js

> Official Post-Quantum Identity Client SDK for DilithiAuth (JavaScript & TypeScript)

`dilithiauth-js` is an Auth0-style client SDK for integrating **DilithiAuth** into single-page applications (React, Vue, Vanilla JS) and Node.js environments. It implements **OAuth 2.0 PKCE Authorization Code Flow** and supports Post-Quantum hybrid token validation (**Ed25519 + NIST FIPS 204 ML-DSA-44**).

---

## 📦 Installation

```bash
npm install dilithiauth-js
```

---

## 🚀 Quick Start (React / SPA)

### 1. Initialize the SDK Client

```typescript
import { DilithiAuthClient } from 'dilithiauth-js';

const authClient = new DilithiAuthClient({
  domain: 'https://dilithiauth.gradientgeeks.tech',
  clientId: 'YOUR_CLIENT_ID',
  redirectUri: window.location.origin + '/callback',
  tokenStrategy: 'thin' // 'thin' | 'jumbo' | 'cwt'
});
```

### 2. Initiate Login

```typescript
// Redirect user to DilithiAuth authorization server with PKCE (S256)
async function handleLogin() {
  await authClient.loginWithRedirect();
}
```

### 3. Handle OAuth Callback

```typescript
// On your callback route (/callback)
async function handleCallback() {
  try {
    const result = await authClient.handleRedirectCallback();
    console.log('Logged in user:', result.user);
    console.log('Access Token:', result.accessToken);
    console.log('Post-Quantum Signature Ref:', result.pqcRef);
  } catch (error) {
    console.error('Authentication error:', error);
  }
}
```

### 4. Access User Session & Tokens

```typescript
// Get active user claims
const user = authClient.getUser();
console.log('User Subject:', user?.sub);

// Get token silently (throws if expired)
const token = await authClient.getTokenSilently();
```

### 5. Sign Out

```typescript
authClient.logout({ returnTo: window.location.origin });
```

---

## 🔑 Features

- **Automated PKCE**: Generates cryptographic `code_verifier` and S256 `code_challenge` automatically.
- **CSRF State Verification**: Validates state parameters against timing attacks.
- **Token Decoding & Claims**: Inspects Ed25519 and ML-DSA-44 post-quantum reference claims (`pqc_ref`).
- **OIDC Discovery & JWKS**: Fetches public keys (`mldsa44_pub`) directly from `/.well-known/jwks.json`.
- **Zero Dependencies**: Pure TypeScript implementation using Web Crypto API.

---

## 🛡️ License

MIT © DilithiAuth
