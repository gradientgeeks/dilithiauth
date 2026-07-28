# Deep Technical & Cryptographic Survey of `pkg/signer` in DilithiAuth

## 1. Architectural Overview of `pkg/signer`

The `pkg/signer` package in DilithiAuth provides a post-quantum hybrid digital signature scheme that pairs a fast classical signature algorithm (**Ed25519**) with a NIST-standardized post-quantum signature algorithm (**ML-DSA-44 / FIPS 204**).

### Core Components and Files
- **`pkg/signer/signer.go`**: Defines `HybridKeyPair`, `HybridSignature`, keypair generation (`GenerateHybridKeyPair`), dual-signing (`Sign`), verification methods (`VerifyEd25519`, `VerifyMLDSA`, `VerifyBoth`), and public key serialization (`PublicKeyInfo`).
- **`pkg/signer/keymanager.go`**: Provides thread-safe key management, key rotation (`Rotate`), active key tracking, caching with periodic store refresh, and lookup by Key ID (`KeyByKid`).
- **`pkg/signer/keystore.go`**: Defines the `KeyStore` interface and `SerializedKeyPair` JSON marshalling/unmarshalling logic using CIRCL binary encoding.
- **`pkg/signer/file_keystore.go`**: Implements filesystem-based key persistence (`FileKeyStore`) with secure `0600` file permissions.
- **`pkg/signer/redis_keystore.go`**: Implements distributed Redis-based key persistence (`RedisKeyStore`) supporting per-tenant prefix isolation.
- **`pkg/signer/signer_test.go` & `keystore_test.go`**: Comprehensive unit tests validating key sizes, round-trip serialization, dual verification, message tampering detection, and key rotation persistence.

---

## 2. HybridKeyPair Implementation & Key Management

### Key Pair Structure (`pkg/signer/signer.go:20-31`)
```go
type HybridKeyPair struct {
    Kid            string
    Ed25519Public  ed25519.PublicKey
    Ed25519Private ed25519.PrivateKey
    MLDSAPublic    *mldsa44.PublicKey
    MLDSAPrivate   *mldsa44.PrivateKey
}
```

### Key Generation (`pkg/signer/signer.go:52-79`)
1. **Key ID (`Kid`) Generation**:
   - Generates 8 cryptographically secure random bytes via `crypto/rand.Read()`.
   - Encodes bytes using `base64.RawURLEncoding` resulting in an 11-character URL-safe identifier (e.g., `aB3_x91K2qA`).
2. **Ed25519 Key Generation**:
   - Calls `ed25519.GenerateKey(rand.Reader)` from standard library `crypto/ed25519`.
   - Produces 32-byte public key and 64-byte private key.
3. **ML-DSA-44 Key Generation**:
   - Calls `mldsa44.GenerateKey(rand.Reader)` from `github.com/cloudflare/circl/sign/mldsa/mldsa44`.
   - Produces 1,312-byte public key and 2,560-byte private key.

### Dual Signing Mechanism (`pkg/signer/signer.go:84-101`)
When `kp.Sign(msg)` is called:
1. **Ed25519 Signature Generation**:
   - Executes `ed25519.Sign(kp.Ed25519Private, msg)`.
   - Output: 64-byte deterministic Ed25519 signature byte slice.
2. **ML-DSA-44 Signature Generation**:
   - Allocates byte buffer `make([]byte, mldsa44.SignatureSize)` (2,420 bytes).
   - Executes `mldsa44.SignTo(kp.MLDSAPrivate, msg, nil, false, mlSig)`.
   - Output: 2,420-byte ML-DSA-44 signature.
3. Returns `HybridSignature` containing both raw signatures.

### Dual Verification Methods (`pkg/signer/signer.go:107-127`)
- **`VerifyEd25519(msg, sig)`**: Verifies classical Ed25519 component via `ed25519.Verify(kp.Ed25519Public, msg, sig.Ed25519Sig)`.
- **`VerifyMLDSA(msg, sig)`**: Verifies post-quantum component via `mldsa44.Verify(kp.MLDSAPublic, msg, nil, sig.MLDSASig)`.
- **`VerifyBoth(msg, sig)`**: Returns `VerifyEd25519(...) && VerifyMLDSA(...)`. Ensures zero compromise on either algorithm.

---

## 3. Cryptographic Primitives & NIST FIPS 204 ML-DSA-44 Specifications

### NIST FIPS 204 Standard Overview
ML-DSA (Module-Lattice Digital Signature Algorithm) is standardized in NIST FIPS 204 (August 2024), derived from CRYSTALS-Dilithium.
- **Underlying Hardness Assumption**: Module Learning With Errors (M-LWE) and Module Short Integer Solution (M-SIS) problems in algebraic lattices.
- **Security Category**: **NIST Security Category 1** (AES-128 equivalent security against both classical and quantum computers).

### Precise Byte Sizes and Parameter Comparison

| Parameter | Classical (Ed25519) | Post-Quantum (ML-DSA-44) | Hybrid Composite | Expansion Factor (ML-DSA / Ed25519) |
|---|---|---|---|---|
| **Public Key Size** | 32 bytes | 1,312 bytes | 1,344 bytes | **41.0x** |
| **Private Key Size** | 64 bytes | 2,560 bytes | 2,624 bytes | **40.0x** |
| **Signature Size** | 64 bytes | 2,420 bytes | 2,484 bytes | **37.8x** |
| **Key ID (`Kid`)** | N/A | N/A | 11 chars (8 bytes raw) | N/A |
| **PQC Reference (`PQCRef`)** | N/A | N/A | 43 chars (32 bytes SHA-256 base64url) | N/A |

---

## 4. Harvest Now, Decrypt Later (HNDL) Threat Model & Mitigation Strategy

### The HNDL Threat Model
1. **Adversary Behavior**: Adversaries (state-sponsored or persistent threat groups) actively record and store encrypted network traffic, authentication handshakes, and identity tokens today.
2. **Quantum Vulnerability**: When a Cryptographically Relevant Quantum Computer (CRQC) becomes operational, Shor's Algorithm will solve discrete logarithms (breaking RSA, ECDSA, Ed25519, ECDH) in polynomial time.
3. **Impact on Identity Systems**:
   - Harvested signed JWTs / OIDC tokens signed with RSA/Ed25519 could be forged retroactively or manipulated to break historical non-repudiation and audit logs.
   - Long-lived assertion credentials or session artifacts could be spoofed.

### How Hybrid Signatures Mitigate HNDL
1. **Quantum Non-Repudiation**: Every token issued by DilithiAuth includes or references an ML-DSA-44 signature. Even decades later, a CRQC cannot forge an ML-DSA-44 signature due to lattice hardness.
2. **Dual-Key Defense in Depth**:
   - If ML-DSA-44 were ever found to have an unforeseen algorithmic flaw, Ed25519 protects the current system.
   - When Ed25519 is broken by quantum computers, ML-DSA-44 maintains signature integrity.
3. **Mitigating "Token Bloat" via Thin Tokens**:
   - Storing a 2,420-byte signature directly in every HTTP header (Jumbo Token ~4 KB) exceeds standard proxy header limits (e.g. 8 KB limits, buffer exhaustion).
   - DilithiAuth solves this in `pkg/oauth` and `pkg/transport` by issuing **Thin Tokens** (< 500 bytes) containing the Ed25519 signature + a 32-byte SHA-256 hash reference (`pqc_ref`) to the ML-DSA-44 signature stored in a sidecar (e.g. Redis).
