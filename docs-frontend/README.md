# docs-frontend Deployment

This directory contains the DilithiAuth **Docs Micro Frontend** — an independently built
and deployable Vite + React application that exposes the `DocsApp` component via
**Module Federation** (`@module-federation/vite`).

## Architecture

```
┌──────────────────────────────────────────────────┐
│         HOST APP  (frontend/)                    │
│  Route /docs → <DocsLoader /> ─────────────────→ │
│  loads remoteEntry.js from docs-frontend at      │
│  runtime (NOT bundled into the host)             │
└──────────────────────────────────────────────────┘
                       ↓  runtime fetch
┌──────────────────────────────────────────────────┐
│         DOCS REMOTE  (docs-frontend/)            │
│  remoteEntry.js  exposes ./DocsApp               │
│  Standalone dev: http://localhost:5174/          │
└──────────────────────────────────────────────────┘
```

## Development

```bash
# Terminal 1 — start docs remote
cd docs-frontend
npm install
npm run dev  # listens on http://localhost:5174

# Terminal 2 — start host app (loads remote at runtime)
cd frontend
npm run dev  # listens on http://localhost:5173

# Visit http://localhost:5173/docs  → docs from micro frontend
# Visit http://localhost:5174/      → docs standalone preview
```

## Building

```bash
# Build remote first (generates remoteEntry.js)
cd docs-frontend && npm run build

# Then build host
cd frontend && npm run build
```

## Production Remote URL

Set `VITE_DOCS_REMOTE_URL` in `frontend/.env.production` to point to where
`docs-frontend/dist/` is served:

```
VITE_DOCS_REMOTE_URL=https://docs.dilithiauth.gradientgeeks.tech/remoteEntry.js
```

Or deploy `docs-frontend/dist/` under the same Firebase Hosting site at `/docs-remote/`
and set:

```
VITE_DOCS_REMOTE_URL=https://dilithiauth.gradientgeeks.tech/docs-remote/remoteEntry.js
```
