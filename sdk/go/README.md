# DilithiAuth Go Client SDK

> Official Go Client SDK for DilithiAuth Post-Quantum Identity Provider

Go client library for integrating **DilithiAuth** into server-side web applications, APIs, and microservices. Supports **OAuth 2.0 PKCE Authorization Code Exchange** and **Hybrid Post-Quantum Token Verification (Ed25519 + ML-DSA-44)**.

---

## 📦 Installation

```bash
go get github.com/Uttam-Mahata/DilithiAuth/sdk/go
```

---

## 🚀 Usage Example

### 1. Initialize Client SDK

```go
import dilithiauth "github.com/Uttam-Mahata/DilithiAuth/sdk/go"

client, err := dilithiauth.NewClient(dilithiauth.Config{
    Domain:       "https://dilithiauth.gradientgeeks.tech",
    ClientID:     "YOUR_CLIENT_ID",
    ClientSecret: "YOUR_CLIENT_SECRET", // Leave empty for Public clients
    RedirectURI:  "https://myapp.com/callback",
})
```

### 2. Generate PKCE & Redirect URL

```go
verifier, challenge, err := dilithiauth.GeneratePKCEPair()
// Save verifier in session / cookie

authURL := client.GetAuthorizationURL("https://myapp.com/callback", "state123", challenge)
// Redirect user to authURL
```

### 3. Exchange Code for Post-Quantum Token

```go
tok, err := client.ExchangeCode(ctx, code, "https://myapp.com/callback", verifier)
if err != nil {
    log.Fatalf("Token exchange failed: %v", err)
}

fmt.Println("Access Token:", tok.AccessToken)
fmt.Println("PQC Signature Reference:", tok.PQCRef)
```

### 4. Verify & Parse Token Claims

```go
claims, err := client.VerifyToken(ctx, tok.AccessToken)
if err != nil {
    log.Fatalf("Invalid token: %v", err)
}

fmt.Printf("User Subject: %s (PQC Ref: %s)\n", claims.Subject, claims.PQCRef)
```

---

## 🛡️ License

MIT © DilithiAuth
