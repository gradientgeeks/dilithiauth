# DilithiAuth.Client (.NET SDK)

> Official .NET Client SDK for DilithiAuth Post-Quantum Identity Provider

C# .NET client library for integrating **DilithiAuth** into ASP.NET Core, Blazor, and .NET 8 backend services. Supports **OAuth 2.0 PKCE S256 Authorization Code Exchange** and **Post-Quantum Hybrid Token Verification**.

---

## 📦 Installation

```bash
dotnet add package DilithiAuth.Client
```

---

## 🚀 Quick Start (ASP.NET Core)

### 1. Initialize Client SDK

```csharp
using DilithiAuth.Client;

var client = new DilithiAuthClient(new DilithiAuthOptions
{
    Domain = "https://dilithiauthapis.gradientgeeks.tech",
    ClientId = "YOUR_CLIENT_ID",
    ClientSecret = "YOUR_CLIENT_SECRET", // Null for Public Clients
    RedirectUri = "http://localhost:5000/callback"
});
```

### 2. Generate PKCE & Authorization URL

```csharp
var pkce = Pkce.GeneratePair();
string state = Pkce.GenerateState();

// Save pkce.CodeVerifier & state in HTTP Session / Encrypted Cookie

string authUrl = client.GetAuthorizationUrl(
    redirectUri: "http://localhost:5000/callback",
    state: state,
    codeChallenge: pkce.CodeChallenge,
    scope: "openid profile email"
);
// Redirect user to authUrl
```

### 3. Exchange Code for Tokens & Fetch User Info

```csharp
// In your callback controller / endpoint
TokenResponse token = await client.ExchangeCodeAsync(code, savedVerifier, "http://localhost:5000/callback");
Console.WriteLine($"Access Token: {token.AccessToken}");
Console.WriteLine($"ML-DSA-44 Signature Ref: {token.PqcRef}");

// Fetch user info
UserInfo userInfo = await client.GetUserInfoAsync(token.AccessToken);
Console.WriteLine($"Logged in user: {userInfo.Sub}");
```

---

## 🛡️ License

MIT © DilithiAuth
