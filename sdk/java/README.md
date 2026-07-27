# DilithiAuth Java Client SDK

> Official Java Client SDK for DilithiAuth Post-Quantum Identity Provider

Java client library for integrating **DilithiAuth** into Spring Boot, Jakarta EE, and Java backend applications. Supports **OAuth 2.0 PKCE S256 Authorization Code Flow** and **Post-Quantum Hybrid Tokens**.

---

## 🚀 Quick Start (Java / Spring Boot)

### 1. Initialize Client SDK

```java
import com.dilithiauth.client.DilithiAuthClient;
import com.dilithiauth.client.PKCE;
import com.dilithiauth.client.TokenResponse;
import com.dilithiauth.client.UserInfo;

DilithiAuthClient client = DilithiAuthClient.builder()
    .domain("https://dilithiauthapis.gradientgeeks.tech")
    .clientId("YOUR_CLIENT_ID")
    .clientSecret("YOUR_CLIENT_SECRET") // Leave empty for Public clients
    .redirectUri("http://localhost:8080/callback")
    .build();
```

### 2. Generate PKCE Pair & Authorization Redirect URL

```java
PKCE.Pair pkce = PKCE.generatePair();
String state = PKCE.generateState();

// Save pkce.getCodeVerifier() & state in HTTPSession

String authUrl = client.getAuthorizationUrl(
    "http://localhost:8080/callback",
    state,
    pkce.getCodeChallenge()
);
// Redirect user to authUrl
```

### 3. Exchange Authorization Code for Token

```java
TokenResponse token = client.exchangeCode(code, savedVerifier, "http://localhost:8080/callback");
System.out.println("Access Token: " + token.getAccessToken());
System.out.println("ML-DSA-44 PQC Ref: " + token.getPqcRef());

// Fetch user profile
UserInfo userInfo = client.getUserInfo(token.getAccessToken());
System.out.println("User Subject: " + userInfo.getSub());
```

---

## 🛡️ License

MIT © DilithiAuth
