# Project: DilithiAuth Developer Portal Redesign

## Architecture
The DilithiAuth developer portal frontend is a React + Vite + Tailwind CSS v4 Single Page Application (`frontend/src/`).
Authenticated routes (`/apps`, `/apps/new`, `/apps/:id`) are wrapped in a persistent left sidebar shell (`SidebarLayout.tsx`) that embeds:
- Top logo & branding (DilithiAuth Post-Quantum IdP)
- Nav links (**Applications** `/apps`, **Docs** `/docs`, **Settings** placeholder)
- Bottom developer profile (user avatar/email) and **Sign Out** button anchored at bottom
- Desktop sticky layout (≥1024px, 256px wide), Tablet icon rail (768px–1023px, 64px wide), Mobile drawer/overlay (<768px) with hamburger trigger
- Dark theme styling using design tokens from `src/index.css`

Public routes (`/`, `/login`, `/register`, `/privacy`, `/terms`) keep `Navbar.tsx` unchanged.

## Code Layout
`frontend/src/`
├── components/
│   ├── layout/
│   │   └── SidebarLayout.tsx     # R1 persistent sidebar wrapper
│   ├── dashboard/
│   │   ├── AppCard.tsx           # R2 App list item / card component
│   │   └── AppSearchFilter.tsx   # R2 Search and type filter bar
│   ├── new-app/
│   │   ├── TypeSelector.tsx      # R3 Visual Public SPA vs Confidential Server card selector
│   │   └── SecretSuccessModal.tsx# R3 Post-creation client secret display modal/card
│   └── app-detail/
│       ├── SettingsTab.tsx       # R4 App overview and details settings form
│       ├── SnippetsTab.tsx       # R4 OAuth code snippets (cURL, JS, Go, Python)
│       └── DangerZoneTab.tsx     # R4 Secret rotation & app deletion modal
├── pages/
│   ├── Dashboard.tsx             # R2 Applications list page
│   ├── NewApp.tsx                # R3 Application creation wizard page
│   └── AppDetail.tsx             # R4 Application detail page
├── context/
│   └── AuthContext.tsx           # Provides `user`, `signOut()`, auth state
└── api.ts                        # API bindings (`apps.list`, `get`, `create`, `rotateSecret`, `delete`)

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| 1 | Exploration & Architecture | Codebase investigation & `PROJECT.md` | None | DONE |
| 2 | R1 Persistent Sidebar Shell | `SidebarLayout.tsx` & route wrapping | M1 | DONE |
| 3 | R2 Applications List Page | Redesign `Dashboard.tsx` with Auth0 console look | M2 | DONE |
| 4 | R3 Create Application Wizard | Redesign `NewApp.tsx` wizard & secret modal | M2 | DONE |
| 5 | R4 App Detail Settings Panel | Redesign `AppDetail.tsx` tabs (Settings/Snippets/Danger) | M2 | IN_PROGRESS |
| 6 | E2E Integration & Build Verification | Build check (`npm run build`), zero TS/Vite errors, test pass | M2, M3, M4, M5 | PLANNED |

## Interface Contracts

### `OAuthApp` Model
```typescript
export interface OAuthApp {
  ID: string
  DeveloperID: string
  Name: string
  Description: string
  HomepageURL: string
  ClientID: string
  RedirectURIs: string[]
  IsPublic: boolean
  CreatedAt: string
  SecretRotatedAt: string | null
}
```

### `apps` API Methods
- `apps.list()` -> `Promise<{ data: OAuthApp[] }>`
- `apps.get(id: string)` -> `Promise<{ data: OAuthApp }>`
- `apps.create(data: { name: string; description: string; homepage_url: string; redirect_uris: string[]; is_public: boolean })` -> `Promise<{ data: { app: OAuthApp; client_secret?: string } }>`
- `apps.rotateSecret(id: string)` -> `Promise<{ data: { app: OAuthApp; client_secret: string } }>`
- `apps.delete(id: string)` -> `Promise<void>`

### `AuthContext`
- `user`: `{ email: string; name?: string } | null`
- `signOut`: `() => void`
