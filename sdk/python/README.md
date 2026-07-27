# dilithiauth

> Official Python Client SDK for DilithiAuth Post-Quantum Identity Provider

Python client SDK for integrating **DilithiAuth** into Flask, FastAPI, Django, and backend services. Supports **OAuth 2.0 PKCE S256 Authorization Code Flow** and **Post-Quantum Hybrid Signature Verification (Ed25519 + ML-DSA-44)**.

---

## 📦 Installation

```bash
pip install dilithiauth
```

---

## 🚀 Quick Start (Flask / FastAPI)

### 1. Initialize Client SDK

```python
from dilithiauth import DilithiAuthClient, generate_pkce_pair, generate_random_state

client = DilithiAuthClient(
    domain="https://dilithiauthapis.gradientgeeks.tech",
    client_id="YOUR_CLIENT_ID",
    client_secret="YOUR_CLIENT_SECRET", # Leave None for Public Clients
    redirect_uri="http://localhost:3000/callback"
)
```

### 2. Generate PKCE & Authorization URL

```python
verifier, challenge = generate_pkce_pair()
state = generate_random_state()

# Store verifier & state in user session
session['code_verifier'] = verifier
session['state'] = state

auth_url = client.get_authorization_url(
    redirect_uri="http://localhost:3000/callback",
    state=state,
    code_challenge=challenge,
    scope="openid profile email"
)
# Redirect user to auth_url
```

### 3. Exchange Code for Tokens & Fetch User Info

```python
# In your callback route (/callback)
token_resp = client.exchange_code(
    code=request.args.get('code'),
    code_verifier=session['code_verifier'],
    redirect_uri="http://localhost:3000/callback"
)

access_token = token_resp['access_token']
pqc_ref = token_resp.get('pqc_ref') # ML-DSA-44 signature reference

# Fetch User Profile
user_info = client.get_user_info(access_token)
print(f"Logged in user: {user_info['sub']} ({user_info.get('email')})")
```

---

## 🛡️ License

MIT © DilithiAuth
