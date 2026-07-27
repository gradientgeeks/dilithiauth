# Technical Context — DilithiAuth

## Application Context
- **IdP Core**: Post-Quantum Identity Provider (IdP) implementing NIST ML-DSA-44 (FIPS 204) + Ed25519 hybrid signatures.
- **Frontend Stack**: React 18 / 19, TypeScript, Vite, Tailwind CSS, Lucide icons, React Router DOM.
- **Key Frontend Files**:
  - `frontend/src/App.tsx`: React router definitions (`/`, `/login`, `/register`, `/dashboard`, `/apps/:id`, `/new-app`, `/docs`).
  - `frontend/src/components/Navbar.tsx`: Header navigation bar.
  - `frontend/src/pages/Docs.tsx`: Target interactive documentation component to be built.
- **Token Types**:
  - Thin Tokens: Classical Ed25519 JWT + `pqc_ref` hash reference pointing to ML-DSA-44 signature stored in sidecar.
  - Jumbo Tokens: Full ~4KB JWT embedding ML-DSA-44 post-quantum signature directly.
  - CWT Tokens: CBOR Web Tokens (RFC 8392) optimized binary representations (20-30% smaller).
- **Supported SDK Languages**:
  1. TypeScript (`npm install @dilithiauth/sdk`)
  2. Go (`go get github.com/DilithiAuth/dilithiauth-go`)
  3. Python (`pip install dilithiauth`)
  4. Java (`mvn dependency: add dilithiauth-sdk`)
  5. .NET (`dotnet add package DilithiAuth.SDK`)
