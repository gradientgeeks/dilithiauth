# Implementation Plan: DilithiAuth Client SDK Libraries (Python, Java, .NET)

Extend DilithiAuth SDK ecosystem by building official client SDK libraries for **Python**, **Java**, and **.NET (C#)**, targeting live DilithiAuth IdP endpoints (`/authorize`, `/token`, `/userinfo`, `/.well-known/jwks.json`), PKCE S256 challenge generation, and Post-Quantum hybrid token support.

## User Review Required

> [!NOTE]
> All SDKs will target live production endpoints at `https://dilithiauthapis.gradientgeeks.tech` (or custom self-hosted domains), supporting:
> - PKCE S256 Code Challenge Generation
> - Recognized Scopes (`openid`, `profile`, `email`, `offline_access`)
> - Token Exchange (`/token`) returning `{ access_token, token_type, pqc_ref, refresh_token }`
> - User Info fetching (`/userinfo`) returning `{ sub, name, email }`
> - Localhost (`http://localhost`) redirect URI support

## Proposed Changes

### 1. Python SDK (`sdk/python`)

#### [NEW] [pyproject.toml](file:///home/uttam/projects/DilithiAuth/sdk/python/pyproject.toml) & [setup.py](file:///home/uttam/projects/DilithiAuth/sdk/python/setup.py)
- Package configuration for `dilithiauth` Python package.

#### [NEW] [dilithiauth/pkce.py](file:///home/uttam/projects/DilithiAuth/sdk/python/dilithiauth/pkce.py)
- Cryptographic PKCE S256 challenge & code verifier generator using `hashlib` and `secrets`.

#### [NEW] [dilithiauth/client.py](file:///home/uttam/projects/DilithiAuth/sdk/python/dilithiauth/client.py)
- `DilithiAuthClient` class:
  - `__init__(domain, client_id, client_secret=None, redirect_uri=None)`
  - `get_authorization_url(redirect_uri, state, code_challenge, scope)`
  - `exchange_code(code, code_verifier, redirect_uri=None)`
  - `get_user_info(access_token)`
  - `verify_token(token)`

#### [NEW] [README.md](file:///home/uttam/projects/DilithiAuth/sdk/python/README.md)
- Complete Python SDK documentation with Flask / FastAPI / Django examples.

---

### 2. Java SDK (`sdk/java`)

#### [NEW] [pom.xml](file:///home/uttam/projects/DilithiAuth/sdk/java/pom.xml)
- Maven project definition for `com.dilithiauth:dilithiauth-sdk`.

#### [NEW] [PKCE.java](file:///home/uttam/projects/DilithiAuth/sdk/java/src/main/java/com/dilithiauth/client/PKCE.java)
- Java PKCE helper generating secure `code_verifier` and S256 `code_challenge`.

#### [NEW] [DilithiAuthClient.java](file:///home/uttam/projects/DilithiAuth/sdk/java/src/main/java/com/dilithiauth/client/DilithiAuthClient.java)
- Builder-pattern Java client with `getAuthorizationUrl()`, `exchangeCode()`, `getUserInfo()`, and `verifyToken()`.

#### [NEW] [README.md](file:///home/uttam/projects/DilithiAuth/sdk/java/README.md)
- Java Spring Boot / Jakarta EE SDK integration guide.

---

### 3. .NET Client SDK (`sdk/dotnet`)

#### [NEW] [DilithiAuth.Client.csproj](file:///home/uttam/projects/DilithiAuth/sdk/dotnet/DilithiAuth.Client/DilithiAuth.Client.csproj)
- C# .NET 8.0 SDK project file.

#### [NEW] [Pkce.cs](file:///home/uttam/projects/DilithiAuth/sdk/dotnet/DilithiAuth.Client/Pkce.cs)
- C# `Pkce` generator using `System.Security.Cryptography`.

#### [NEW] [DilithiAuthClient.cs](file:///home/uttam/projects/DilithiAuth/sdk/dotnet/DilithiAuth.Client/DilithiAuthClient.cs)
- C# client with async methods: `GetAuthorizationUrl()`, `ExchangeCodeAsync()`, `GetUserInfoAsync()`, `VerifyToken()`.

#### [NEW] [README.md](file:///home/uttam/projects/DilithiAuth/sdk/dotnet/DilithiAuth.Client/README.md)
- ASP.NET Core integration guide.

---

### 4. Developer Portal Integration

#### [MODIFY] [AppDetail.tsx](file:///home/uttam/projects/DilithiAuth/frontend/src/pages/AppDetail.tsx)
- Add Java and .NET SDK tabs to the Developer Portal Code Generator.

---

## Verification Plan

### Automated Tests
- Test Python SDK:
  ```bash
  cd sdk/python && python3 -m unittest discover tests
  ```
- Test Java SDK compilation:
  ```bash
  cd sdk/java && mvn test-compile || javac -d target/classes src/main/java/com/dilithiauth/client/*.java
  ```
- Test .NET SDK build:
  ```bash
  dotnet build sdk/dotnet/DilithiAuth.Client/DilithiAuth.Client.csproj
  ```

### Manual Verification
- Test code exchange and userinfo responses using each language SDK.
