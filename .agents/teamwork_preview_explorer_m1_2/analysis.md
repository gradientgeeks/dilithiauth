# DilithiAuth Backend API & Post-Quantum Token Technical Analysis

## Executive Summary

This report provides a comprehensive analysis of the DilithiAuth Post-Quantum Identity Provider (IdP) backend API specifications, Post-Quantum token mechanics, and authentic multi-language SDK code implementations (TypeScript, Go, Python, Java, and .NET).

DilithiAuth integrates **NIST FIPS 204 (ML-DSA-44)** with **Ed25519** classical cryptography to provide hybrid post-quantum authorization code flows, OIDC discovery, and token verification while solving the ~2.4 KB signature token bloat problem through three distinct token delivery strategies.

---

## 1. Backend API Specification

DilithiAuth exposes standard OAuth 2.0 / OpenID Connect endpoints along with quantum-safe hybrid key discovery, token revocation, verification, and demo token issuance.

### Endpoints Overview Table

| Endpoint | Method | Auth Required | Content Type | Purpose / Description |
|---|---|---|---|---|
| `/.well-known/openid-configuration` | `GET` | None | `application/json` | OIDC Discovery Metadata (RFC 8414) |
| `/.well-known/jwks.json` | `GET` | None | `application/json` | Public Key Set containing Ed25519 & ML-DSA-44 public keys |
| `/authorize` | `GET` / `POST` | User Credentials | `text/html` / Redirect | PKCE Authorization Code flow login & consent entrypoint |
| `/token` | `POST` | Client Credentials | `application/json` | Exchange auth code or refresh token for access token (Rate-limited: 2 req/sec, burst 20) |
| `/userinfo` | `GET` | Bearer Token | `application/json` | OIDC protected endpoint returning user claims (`sub`, `name`, `email`) |
| `/revoke` | `POST` | Client Credentials | `application/json` / 200 OK | RFC 7009 token revocation with hybrid signature verification |
| `/verify/jwt` | `POST` | None | `application/json` | REST verification endpoint for lightweight clients |
| `/token/thin` | `POST` | None (Demo tag) | `application/json` | Issue Thin JWT + store PQC signature in SigStore |
| `/token/full` | `POST` | None (Demo tag) | `application/json` | Issue Jumbo JWT with embedded `pqc_sig` (~4KB) |
| `/token/cwt` | `POST` | None (Demo tag) | `application/cbor` | Issue binary CBOR Web Token (RFC 8392) |

---

### Detailed Endpoint Specifications

#### 1. `GET /.well-known/openid-configuration`
- **Description**: Returns OpenID Connect discovery metadata.
- **Response Example**:
```json
{
  "issuer": "https://dilithiauthapis.gradientgeeks.tech",
  "authorization_endpoint": "https://dilithiauthapis.gradientgeeks.tech/authorize",
  "token_endpoint": "https://dilithiauthapis.gradientgeeks.tech/token",
  "jwks_uri": "https://dilithiauthapis.gradientgeeks.tech/.well-known/jwks.json",
  "response_types_supported": ["code"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256", "EdDSA", "MLDSA44"],
  "pqc_hybrid_enabled": true
}
```

#### 2. `GET /.well-known/jwks.json`
- **Description**: Exposes public signing keys for dual-signature verification.
- **Key Fields**:
  - `kid`: Key identifier shared between Ed25519 and ML-DSA-44 sub-keys.
  - `kty`: `"OKP"`
  - `crv`: `"Ed25519"`
  - `x`: Base64url-encoded Ed25519 public key (32 bytes raw).
  - `mldsa44_pub`: Base64url-encoded ML-DSA-44 public key (1312 bytes raw).
  - `alg`: `"EdDSA+MLDSA44"`
- **Response Example**:
```json
{
  "keys": [
    {
      "kid": "x9K3mP1Q",
      "kty": "OKP",
      "crv": "Ed25519",
      "x": "u18d...",
      "mldsa44_pub": "A9x2...",
      "use": "sig",
      "alg": "EdDSA+MLDSA44"
    }
  ]
}
```

#### 3. `/authorize` (`GET` / `POST`)
- **Query Parameters**:
  - `client_id` (required): OAuth client ID.
  - `redirect_uri` (required): Registered client callback URL.
  - `response_type`: Must be `"code"`.
  - `state` (recommended): Anti-CSRF token.
  - `scope` (optional): Space-delimited requested scopes (e.g., `"openid profile email offline_access"`).
  - `code_challenge` (required for public clients): Base64url-encoded S256 code challenge.
  - `code_challenge_method`: Must be `"S256"`.
- **Flow**:
  1. `GET /authorize` renders login HTML form.
  2. `POST /authorize` authenticates user credentials (using bcrypt / Argon2id).
  3. If first-time consent is needed, issues a short-lived (5 min) HS256-signed consent ticket and renders consent page.
  4. Upon approval, creates authorization code and redirects to `redirect_uri?code=...&state=...`.

#### 4. `POST /token`
- **Content-Type**: `application/x-www-form-urlencoded`
- **Request Parameters** (`grant_type=authorization_code`):
  - `grant_type`: `"authorization_code"`
  - `client_id`: OAuth client ID.
  - `code`: Authorization code issued by `/authorize`.
  - `redirect_uri`: Matching redirect URI.
  - `code_verifier`: PKCE code verifier (SHA256(code_verifier) == code_challenge).
  - `client_secret` (optional for public PKCE clients, required for confidential clients via Basic Auth or body).
- **Request Parameters** (`grant_type=refresh_token`):
  - `grant_type`: `"refresh_token"`
  - `refresh_token`: Rotated refresh token string.
- **Response Example**:
```json
{
  "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6Ing5SzNtUDFRIiwidHlwIjoiSldUIn0...",
  "token_type": "Bearer",
  "pqc_ref": "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
  "refresh_token": "da_rt_9f8a..."
}
```

#### 5. `GET /userinfo`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: Returns identity claims based on granted scopes:
```json
{
  "sub": "alice",
  "name": "alice",
  "email": "alice@example.com"
}
```

---

## 2. Post-Quantum Token Mechanics & Delivery Strategies

Quantum-safe signatures like NIST ML-DSA-44 require ~2,420 bytes per signature (compared to 64 bytes for Ed25519). If included in standard HTTP Authorization headers, tokens can exceed typical server header limits (4 KB to 8 KB). DilithiAuth solves this using **3 token strategies**:

```
+-----------------------------------------------------------------------------------------+
|                                    DilithiAuth Strategies                               |
+------------------------------+------------------------------+---------------------------+
| 1. Thin Token (Default)      | 2. Full / Jumbo Token        | 3. CBOR Web Token (CWT)   |
| - Format: JWT (JSON)         | - Format: JWT (JSON)         | - Format: CBOR (Binary)   |
| - Size: ~500 bytes           | - Size: ~3.5 - 4.0 KB        | - Size: ~200 - 300 bytes  |
| - Ed25519 Sig: In Token      | - Ed25519 Sig: In Token      | - Ed25519 Sig: In Env     |
| - ML-DSA-44: SigStore        | - ML-DSA-44: Embedded claim  | - ML-DSA-44: SigStore     |
|   (referenced via `pqc_ref`) |   (`pqc_sig` claim)          |   (referenced via `cti`)  |
| - Header Limit: Standard     | - Requires 16 KiB Header Limit| - Binary / IoT optimized  |
+------------------------------+------------------------------+---------------------------+
```

### Strategy Breakdown

1. **Thin Tokens (`pqc_ref` Sidecar Strategy - Recommended Default)**:
   - **Header / Signature**: Signed classically via Ed25519 (`alg: "EdDSA"`).
   - **Payload**: Standard JWT claims + `pqc_ref` claim containing `base64url(SHA-256(mldsa44_sig))`.
   - **Sidecar Storage (`SigStore`)**: The full ~2,420 byte ML-DSA-44 signature is stored out-of-band in SigStore (In-Memory, File, PostgreSQL, or Redis) with a TTL matching token expiration (1 hour).
   - **Verification**: Lightweight microservices verify Ed25519 instantly. High-security verifiers fetch the ML-DSA-44 signature from SigStore using `pqc_ref` or `jti` to execute full post-quantum verification.

2. **Jumbo Tokens (Embedded ML-DSA-44 Signature)**:
   - **Payload**: Contains `pqc_sig` claim holding the raw base64url-encoded ML-DSA-44 signature (~2,420 bytes).
   - **Token Size**: ~3,500 – 4,000 bytes.
   - **Self-Contained**: 100% self-contained token; no database or cache lookup needed for verification.
   - **Infrastructure Impact**: Server and reverse proxies (Nginx, Envoy, Cloudflare) must be configured with `MaxHeaderBytes = 16 KiB` (`transport.MaxHeaderBytesForPQC`).

3. **CBOR Web Tokens (CWT - RFC 8392)**:
   - **Encoding**: Binary Concise Binary Object Representation (CBOR).
   - **Integer Claim Mapping**:
     - `1`: `iss` (Issuer)
     - `2`: `sub` (Subject)
     - `4`: `exp` (Expiration)
     - `6`: `iat` (Issued At)
     - `7`: `cti` (Token ID)
     - `8`: `pqc_ref` (PQC Signature Hash Reference)
     - `9`: `tid` (Tenant ID)
     - `10`: `scope` (Granted Scopes)
   - **Envelope Structure**: CBOR array `[claims_bytes, ed25519_sig]`.
   - **Efficiency**: 20% to 30% smaller than JSON equivalent; ideal for low-power IoT and mobile network constraints.

---

## 3. SDK Code Snippets Across 5 Languages

Below are authentic, copy-paste ready code examples for initializing the client, generating PKCE pairs, conducting code exchange, and retrieving user info for each official DilithiAuth SDK.

---

### A. TypeScript / JavaScript (`dilithiauth-js` / `@dilithiauth/sdk`)

#### Package Installation
```bash
npm install dilithiauth-js
```

#### Initialization & PKCE Redirect
```typescript
import { DilithiAuthClient, generateCodeVerifier, generateCodeChallenge } from 'dilithiauth-js';

const client = new DilithiAuthClient({
  domain: 'https://dilithiauthapis.gradientgeeks.tech',
  clientId: 'my-web-app-client-id',
  redirectUri: 'https://myapp.com/callback',
  tokenStrategy: 'thin',
  storageType: 'sessionStorage'
});

// Trigger PKCE Login Flow
async function login() {
  await client.loginWithRedirect({
    scope: 'openid profile email offline_access'
  });
}
```

#### Callback Handling & Token Exchange
```typescript
async function handleCallback() {
  try {
    const result = await client.handleRedirectCallback();
    console.log('Access Token:', result.accessToken);
    console.log('PQC Reference Hash:', result.pqcRef);
    console.log('Authenticated User:', result.user);
  } catch (error) {
    console.error('Login callback error:', error);
  }
}
```

---

### B. Go (`github.com/DilithiAuth/dilithiauth-go` / `dilithiauth`)

#### Package Installation
```bash
go get github.com/DilithiAuth/dilithiauth-go
```

#### Client Initialization, PKCE & Code Exchange
```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/DilithiAuth/dilithiauth-go"
)

func main() {
	client, err := dilithiauth.NewClient(dilithiauth.Config{
		Domain:       "https://dilithiauthapis.gradientgeeks.tech",
		ClientID:     "my-go-backend-client",
		ClientSecret: "my-client-secret",
		RedirectURI:  "http://localhost:8080/callback",
	})
	if err != nil {
		log.Fatalf("Failed to initialize DilithiAuth client: %v", err)
	}

	// Step 1: Generate PKCE Pair
	verifier, challenge, err := dilithiauth.GeneratePKCEPair()
	if err != nil {
		log.Fatalf("Failed to generate PKCE pair: %v", err)
	}

	// Step 2: Build Authorization URL
	state := "random-csrf-state"
	authURL := client.GetAuthorizationURL("http://localhost:8080/callback", state, challenge)
	fmt.Println("Redirect user to:", authURL)

	// Step 3: Exchange Authorization Code for Token
	code := "received-authorization-code"
	tokenResp, err := client.ExchangeCode(context.Background(), code, "http://localhost:8080/callback", verifier)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}

	fmt.Printf("Access Token: %s\n", tokenResp.AccessToken)
	fmt.Printf("PQC Ref Hash: %s\n", tokenResp.PQCRef)

	// Step 4: Verify & Parse Token Claims
	claims, err := client.VerifyToken(context.Background(), tokenResp.AccessToken)
	if err != nil {
		log.Fatalf("Token verification failed: %v", err)
	}
	fmt.Printf("Authenticated Subject: %s\n", claims.Subject)
}
```

---

### C. Python (`dilithiauth`)

#### Package Installation
```bash
pip install dilithiauth
```

#### Complete Authentication Flow
```python
from dilithiauth import DilithiAuthClient, generate_pkce_pair, generate_random_state

# Step 1: Initialize Client
client = DilithiAuthClient(
    client_id="my-python-app",
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_secret="my-secret-key",
    redirect_uri="http://localhost:5000/callback"
)

# Step 2: Generate PKCE & Authorization URL
code_verifier, code_challenge = generate_pkce_pair()
state = generate_random_state()

auth_url = client.get_authorization_url(
    redirect_uri="http://localhost:5000/callback",
    state=state,
    code_challenge=code_challenge,
    scope="openid profile email"
)
print("Navigate to:", auth_url)

# Step 3: Exchange Code for Access Token
auth_code = "incoming_authorization_code"
token_response = client.exchange_code(
    code=auth_code,
    code_verifier=code_verifier,
    redirect_uri="http://localhost:5000/callback"
)

access_token = token_response["access_token"]
pqc_ref = token_response.get("pqc_ref")

print(f"Access Token: {access_token[:20]}...")
print(f"PQC Reference: {pqc_ref}")

# Step 4: Fetch UserInfo & Verify Claims
user_info = client.get_user_info(access_token)
print(f"User Info: {user_info}")

claims = client.verify_token(access_token)
print(f"Token Subject: {claims['sub']}")
```

---

### D. Java (`com.dilithiauth:dilithiauth-sdk`)

#### Dependency Setup (Maven `pom.xml`)
```xml
<dependency>
    <groupId>com.dilithiauth</groupId>
    <artifactId>dilithiauth-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

#### Initialization & Token Exchange Flow
```java
import com.dilithiauth.client.DilithiAuthClient;
import com.dilithiauth.client.PKCE;
import com.dilithiauth.client.TokenResponse;
import com.dilithiauth.client.UserInfo;

public class App {
    public static void main(String[] args) throws Exception {
        // Step 1: Initialize Client via Builder
        DilithiAuthClient client = DilithiAuthClient.builder()
                .domain("https://dilithiauthapis.gradientgeeks.tech")
                .clientId("my-java-enterprise-app")
                .clientSecret("secret-value")
                .redirectUri("https://enterprise.example.com/oauth/callback")
                .build();

        // Step 2: PKCE Generation & Auth URL
        PKCE.Pair pkce = PKCE.generatePair();
        String state = PKCE.generateState();

        String authUrl = client.getAuthorizationUrl(
                "https://enterprise.example.com/oauth/callback",
                state,
                pkce.getCodeChallenge(),
                "openid profile email"
        );
        System.out.println("Authorization URL: " + authUrl);

        // Step 3: Token Exchange
        String code = "authorization_code_from_callback";
        TokenResponse tokenResp = client.exchangeCode(code, pkce.getCodeVerifier(), "https://enterprise.example.com/oauth/callback");

        System.out.println("Access Token: " + tokenResp.getAccessToken());
        System.out.println("PQC Ref Hash: " + tokenResp.getPqcRef());

        // Step 4: UserInfo Retrieval
        UserInfo userInfo = client.getUserInfo(tokenResp.getAccessToken());
        System.out.println("User Subject: " + userInfo.getSub());
        System.out.println("User Email: " + userInfo.getEmail());
    }
}
```

---

### E. .NET (`DilithiAuth.SDK` / `DilithiAuth.Client`)

#### Package Installation
```bash
dotnet add package DilithiAuth.Client
```

#### Client Usage Example
```csharp
using System;
using System.Threading.Tasks;
using DilithiAuth.Client;

class Program
{
    static async Task Main(string[] args)
    {
        // Step 1: Initialize Client
        var options = new DilithiAuthOptions
        {
            Domain = "https://dilithiauthapis.gradientgeeks.tech",
            ClientId = "dotnet-api-client",
            ClientSecret = "client-secret-key",
            RedirectUri = "https://localhost:5001/signin-oidc"
        };

        var client = new DilithiAuthClient(options);

        // Step 2: PKCE Generation & Auth URL
        PkcePair pkce = Pkce.GeneratePair();
        string state = Pkce.GenerateState();

        string authUrl = client.GetAuthorizationUrl(
            redirectUri: "https://localhost:5001/signin-oidc",
            state: state,
            codeChallenge: pkce.CodeChallenge,
            scope: "openid profile email"
        );
        Console.WriteLine($"Redirect user to: {authUrl}");

        // Step 3: Exchange Code for Access Token
        string code = "received_auth_code";
        TokenResponse tokenResp = await client.ExchangeCodeAsync(code, pkce.CodeVerifier);

        Console.WriteLine($"Access Token: {tokenResp.AccessToken}");
        Console.WriteLine($"PQC Reference: {tokenResp.PqcRef}");

        // Step 4: UserInfo & Token Verification
        UserInfo userInfo = await client.GetUserInfoAsync(tokenResp.AccessToken);
        Console.WriteLine($"Authenticated User Sub: {userInfo.Sub}");

        Claims claims = client.VerifyToken(tokenResp.AccessToken);
        Console.WriteLine($"Valid Subject Claim: {claims.Subject}");
    }
}
```

---

## 4. Verification & Methodologies

1. **Backend Endpoints**: Verified via inspection of `cmd/server/main.go`, `cmd/server/demo_routes.go`, `pkg/oauth/oauth.go`, `pkg/oauth/authz_handlers.go`.
2. **Post-Quantum Token Strategies**: Verified via implementation signatures in `pkg/oauth/oauth.go` (`IssueJWTThin`, `IssueJWTFull`, `IssueCWT`) and `pkg/signer/signer.go` (`ML-DSA-44` signature size ~2420 bytes, Ed25519 signature size 64 bytes).
3. **SDK Implementations**: Verified via actual codebase inspection in `sdk/js`, `sdk/go`, `sdk/python`, `sdk/java`, and `sdk/dotnet`. All code examples match exact constructor parameters, method names, and payload structures.
