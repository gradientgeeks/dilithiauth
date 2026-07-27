# Project: DilithiAuth Interactive SDK Documentation Page

## Overview
DilithiAuth is a Post-Quantum Identity Provider (IdP) featuring hybrid Ed25519 + ML-DSA-44 signatures and custom token formats (Thin, Jumbo, CWT). This project adds an interactive, enterprise SaaS-styled SDK Documentation Page (`/docs`) to the React SPA (`frontend/src/pages/Docs.tsx`), complete with 5 language tabs (TypeScript, Go, Python, Java, .NET), quickstart guides, PKCE S256 code samples, copy-to-clipboard functionality, interactive API reference (Discovery, JWKS, /authorize, /token, /userinfo, Post-Quantum token formats), and navbar/router integration.

## Code Layout
- Frontend Root: `frontend/`
- Page components: `frontend/src/pages/`
  - `Docs.tsx` (New interactive SDK documentation page)
- UI components: `frontend/src/components/`
  - `Navbar.tsx` (Navigation bar with active state for `/docs`)
  - `QuantumTokenPlayground.tsx` (Existing playground for reference/reuse patterns)
- Main Routing: `frontend/src/App.tsx`
- Styles: `frontend/src/index.css` & `App.css` (Tailwind CSS / modern CSS classes)
- API Specs/Constants: `frontend/src/api.ts`

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | Base Exploration & Specs | Investigate frontend setup, dependencies, existing component styling, and draft exact structure for Docs.tsx | None | DONE |
| M2 | SDK Docs Component (`Docs.tsx`) & Multi-Lang Quickstart | Implement `Docs.tsx` with language switcher tabs (TS, Go, Python, Java, .NET), installation commands, PKCE S256 snippets, token exchange, userinfo retrieval, and copy buttons | M1 | DONE |
| M3 | Interactive API & Identity Reference | Add API reference sections to `Docs.tsx` covering OpenID Discovery, JWKS, `/authorize`, `/token`, `/userinfo`, and PQ token formats (Thin, Jumbo, CWT) | M2 | DONE |
| M4 | Navigation & Routing | Wire up `/docs` in `frontend/src/App.tsx` router and update `Navbar.tsx` to include "SDK & Docs" navigation link | M3 | DONE |
| M5 | Quality, Build & Forensic Audit | Run `npm run build` in `frontend/`, ensure 0 TypeScript / Vite warnings/errors, and complete forensic audit | M4 | DONE |

## Interface Contracts
- `/docs` route renders `Docs.tsx`
- Navbar includes a link to `/docs` with active state highlighting
- Language tabs: TypeScript (`ts`), Go (`go`), Python (`python`), Java (`java`), .NET (`dotnet`)
- Code snippets copyable with toast/copied feedback notification
- Clean TypeScript types, standard React state management
