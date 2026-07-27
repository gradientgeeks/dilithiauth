# DilithiAuth

DilithiAuth is a research-oriented Identity Provider (IdP) designed to mitigate the **"Harvest Now, Forge Later" (HNFL)** threat. It implements the OAuth 2.0 Authorization Code Flow using hybrid signatures combining **Ed25519** (classical) and **ML-DSA-44** (Module-Lattice Digital Signature Algorithm, the NIST FIPS 204 standardized algorithm for post-quantum digital signatures) to prevent future quantum-enabled token forgery and identity impersonation.

## The Problem: Post-Quantum Token Bloat

A standard Ed25519 signature is 64 bytes. An ML-DSA-44 signature is **~2.4 Kilobytes**. 
When signing standard JWTs with post-quantum algorithms, the resulting tokens become so large ("Jumbo JWTs") that they break standard web infrastructure (most reverse proxies and API gateways drop HTTP headers larger than 4-8 KB).

## The Solution: DilithiAuth Architecture

DilithiAuth offers multiple strategies to handle the massive signature size of post-quantum cryptography, allowing you to choose the right tradeoff between statelessness, bandwidth, and infrastructure compatibility.

### 1. The Edge Proxy Offloading Pattern (Recommended for Production)
To keep your internal microservices entirely stateless and agnostic to post-quantum cryptography, DilithiAuth includes an **Edge Proxy Gateway**.

1. The `cmd/server` IdP issues a **Jumbo JWT** (contains the full 2.4KB signature).
2. The `cmd/edgeproxy` (configured to accept 16KB headers) receives the API request from the client.
3. The Edge Proxy mathematically verifies both the Ed25519 and ML-DSA-44 signatures.
4. The Edge Proxy **strips** the massive PQC signature and re-signs the token into a tiny, standard HS256 JWT.
5. The request is forwarded to your internal microservices, which validate the token normally without any PQC overhead.

### 2. The "Thin" JWT Pattern (Stateful Alternative)
If you cannot run an Edge Proxy, the IdP can issue a "Thin JWT".
* The token contains only the classical Ed25519 signature and a `pqc_ref` hash.
* The massive ML-DSA-44 signature is stored in a backend `SigStore` (e.g., Redis).
* Your microservices must query Redis to fetch the signature before validating the token.

### 3. CBOR Web Tokens (CWT)
For bandwidth-constrained environments, DilithiAuth can issue binary tokens using RFC 8392. CWTs are roughly 20-30% smaller than base64url-encoded JWTs.

---

## High Availability & Scalability

DilithiAuth is designed to be fully stateless and horizontally scalable when deployed with a Redis cluster. It abstracts three core components into distributed interfaces:

1. **`KeyStore`**: Manages the rotation and persistence of the hybrid Ed25519 + ML-DSA-44 key pairs across multiple IdP nodes.
2. **`AuthzStore`**: Manages OAuth 2.0 Authorization Codes (with exact-once redemption guarantees via Redis transactions).
3. **`SigStore`**: Manages the off-band storage of PQC signatures (only used if you opt for the "Thin JWT" strategy).

## Getting Started

### Prerequisites
* Go 1.21+
* (Optional) Redis server for distributed testing

### Running the Demo Architecture (Edge Proxy Pattern)

You can run the complete stateless architecture locally using three terminals:

**1. Start the Identity Provider (IdP)**
```bash
go run ./cmd/server
```
*(Runs on `:8080`. Generates ephemeral hybrid keys on startup).*

**2. Start the Dummy Microservice**
```bash
go run ./cmd/microservice
```
*(Runs on `:8082`. A standard API that expects a lightweight HS256 JWT. It knows nothing about post-quantum cryptography).*

**3. Start the Edge Proxy**
```bash
go run ./cmd/edgeproxy
```
*(Runs on `:8081`. Accepts Jumbo JWTs, verifies the ML-DSA-44 signature, strips it, and proxies to the microservice).*

### Running in Distributed Production Mode

To run multiple instances of DilithiAuth behind a load balancer, configure the distributed backends via environment variables:

```bash
# Point to your Redis cluster
export DILITHIAUTH_REDIS_ADDR="localhost:6379"

# Enable distributed components
export DILITHIAUTH_KEYSTORE=redis
export DILITHIAUTH_AUTHZSTORE=redis

# (Optional) Only needed if using Thin JWTs instead of Jumbo JWTs
export DILITHIAUTH_SIGSTORE=redis 

go run ./cmd/server
```

## Infrastructure Tuning Warning
If you are passing Jumbo JWTs to the Edge Proxy, ensure that any external Load Balancers (AWS ALB, Cloudflare) or Reverse Proxies (Nginx) in front of the Edge Proxy are configured to allow HTTP headers up to **16 KiB**. The standard default is often 4 KiB or 8 KiB.

## Testing
Run the comprehensive test suite:
```bash
go test ./...
```
