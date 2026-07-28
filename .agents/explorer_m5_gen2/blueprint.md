# Milestone 5 Architecture Blueprint: App Detail Page Redesign (`AppDetail.tsx`)

## 1. Overview & Objective
The goal of Milestone 5 (R4) is to redesign `frontend/src/pages/AppDetail.tsx` into an **Auth0/Clerk-style tabbed settings panel** hosted cleanly within the persistent `<SidebarLayout />` shell.

The redesigned page features:
1. **Header Area**: Breadcrumb back-link to `/apps`, App Name, Client ID badge, App Type badge (`Public SPA (PKCE)` vs `Confidential Client`), and quick copy action.
2. **Sub-Nav / Tab Switcher**: Tabs for **Settings**, **Code Snippets**, and **Danger Zone**.
3. **Modular Subcomponents** under `src/components/app-detail/`:
   - `SettingsTab.tsx`: Interactive settings form including App Name, Description, Homepage URL, multi-row Redirect URIs manager, Client ID, Client Type display, and "Save Changes" action.
   - `SnippetsTab.tsx`: Code snippets viewer with language switcher (cURL, JS / Fetch, Go, Python) demonstrating DilithiAuth token endpoints (`/token/thin`, `/token/full`, `/token/cwt`).
   - `DangerZoneTab.tsx`: Dedicated danger zone box with secret rotation modal and application deletion modal with type-app-name validation.

---

## 2. Current State Analysis

### 2.1 File: `frontend/src/pages/AppDetail.tsx`
- Currently line-count ~396 lines, single monolithic component.
- Uses `useParams<{ id: string }>()` and fetches app using `apps.get(id)`.
- Has existing basic tabs (`overview`, `snippets`, `danger`), but without settings form editing or modular subcomponents.
- Uses `ConfirmModal.tsx` for deletion confirmation without type-to-confirm validation.

### 2.2 API Layer: `frontend/src/api.ts`
- Currently provides:
  - `apps.get(id)`
  - `apps.create(...)`
  - `apps.rotateSecret(id)`
  - `apps.delete(id)`
- Worker M5 should add `apps.update(id, data)` to `api.ts` (or update parent component state on form save):
```ts
update: (id: string, data: {
  name?: string
  description?: string
  homepage_url?: string
  redirect_uris?: string[]
}) => api.put<OAuthApp>(`/apps/${id}`, data),
```

---

## 3. Page Architecture & Sub-Nav Layout (`AppDetail.tsx`)

### 3.1 Header Area Structure
- **Back Navigation**: `<Link to="/apps" className="inline-flex items-center gap-1.5 text-xs text-[var(--color-sub)] hover:text-[var(--color-txt)] transition-colors mb-3">` `<ArrowLeft size={14} /> Back to Applications</Link>`
- **App Title Row**:
  - `h1`: App Name (`text-2xl font-bold text-[var(--color-txt)] tracking-tight`)
  - **Client ID Badge**: `mono-inline text-xs flex items-center gap-1.5 px-2.5 py-1 rounded bg-[#050810] border border-[var(--color-border)]` with inline Copy icon button.
  - **App Type Badge**:
    - Public: `<span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-sky-500/10 border border-sky-500/30 text-sky-400">` `<Globe size={12} /> Public SPA (PKCE)</span>`
    - Confidential: `<span className="inline-flex items-center gap-1 px-2.5 py-1 rounded text-xs font-semibold bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">` `<Lock size={12} /> Confidential Client</span>`
- **Sub-Nav Bar**:
  - Border-bottom navigation tab bar:
    ```tsx
    const TABS = [
      { id: 'settings', label: 'Settings', icon: Settings },
      { id: 'snippets', label: 'Code Snippets', icon: Code },
      { id: 'danger', label: 'Danger Zone', icon: AlertTriangle },
    ]
    ```
  - Styling:
    - Active: `text-[var(--color-txt)] border-b-2 border-[var(--color-primary)] font-semibold bg-transparent`
    - Inactive: `text-[var(--color-sub)] border-b-2 border-transparent hover:text-[var(--color-txt)] hover:bg-white/[0.03]`

---

## 4. Modular Subcomponents Blueprint (`src/components/app-detail/`)

### 4.1 Component: `SettingsTab.tsx`
- **File**: `frontend/src/components/app-detail/SettingsTab.tsx`
- **Purpose**: Render and manage application settings form.
- **Props Interface**:
```typescript
import type { OAuthApp } from '../../api'

export interface SettingsTabProps {
  app: OAuthApp
  onSave: (updated: {
    name: string
    description: string
    homepageURL: string
    redirectURIs: string[]
  }) => Promise<void>
  saving: boolean
  saveSuccess: boolean
  saveError: string | null
}
```
- **Fields & Controls**:
  1. **Application Name**: Input text field (required).
  2. **Description**: Textarea field (optional).
  3. **Homepage URL**: Input text field (optional, URL format).
  4. **Authorized Redirect URIs**: Interactive multi-row list manager:
     - Multi-row inputs with index-based state (`redirectURIs: string[]`).
     - Button `+ Add Redirect URI` (`addURIRow`).
     - Trash button per row (`removeURIRow`), preserving at least 1 item.
  5. **Client ID**: Read-only input box with inline `Copy` button.
  6. **Client Type**: Read-only indicator showing Public/Confidential type badge & Security Note.
  7. **Save Changes Button**: Primary submit button (`Save Changes`) with loading spinner state and success toast banner (`Settings updated successfully`).

### 4.2 Component: `SnippetsTab.tsx`
- **File**: `frontend/src/components/app-detail/SnippetsTab.tsx`
- **Purpose**: OAuth 2.0 / DilithiAuth integration code snippets viewer.
- **Props Interface**:
```typescript
import type { OAuthApp } from '../../api'

export interface SnippetsTabProps {
  app: OAuthApp
}
```
- **Language Tabs**:
  - `cURL`, `JS / Fetch`, `Go`, `Python`.
- **Code Block Content**:
  - Clear, syntax-highlighted code blocks for each language showing:
    - Authorization redirect step.
    - Token requests targeting DilithiAuth endpoints:
      - `/token/thin` (Thin JWT with PQC sidecar reference)
      - `/token/full` (Jumbo JWT embedding NIST FIPS 204 ML-DSA-44)
      - `/token/cwt` (CBOR Web Token, binary format)
- **Controls**:
  - Copy Code button with `Copied!` feedback state.
  - Informational banner describing hybrid signature verification (`/.well-known/jwks.json`).

### 4.3 Component: `DangerZoneTab.tsx`
- **File**: `frontend/src/components/app-detail/DangerZoneTab.tsx`
- **Purpose**: Render danger operations (Secret Rotation and Application Deletion).
- **Props Interface**:
```typescript
import type { OAuthApp } from '../../api'

export interface DangerZoneTabProps {
  app: OAuthApp
  onOpenRotateModal: () => void
  onOpenDeleteModal: () => void
  rotating: boolean
  deleting: boolean
}
```
- **Layout & Structure**:
  - Container box styled with subtle dark rose background & border (`bg-rose-500/5 border border-rose-500/25 rounded-xl p-6`).
  - **Row 1: Rotate Client Secret** (only active for Confidential Apps):
    - Subtitle: "Generate a new client secret. The current secret will be immediately revoked."
    - Button: `<button className="px-3.5 py-2 rounded-lg text-sm font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/30 hover:bg-amber-500/20">Rotate Client Secret</button>`
    - If Public App: Shows informational note ("Public applications do not use client secrets.").
  - **Row 2: Delete Application**:
    - Subtitle: "Permanently remove this application, revoking all credentials and issued tokens."
    - Button: `<button className="px-3.5 py-2 rounded-lg text-sm font-semibold bg-rose-500 text-white hover:bg-rose-600 shadow-lg shadow-rose-500/20">Delete Application</button>`

### 4.4 Component: `SecretDisplayModal.tsx`
- **File**: `frontend/src/components/app-detail/SecretDisplayModal.tsx`
- **Purpose**: Display newly rotated client secret once with copy button and warning.
- **Props Interface**:
```typescript
export interface SecretDisplayModalProps {
  clientSecret: string
  onClose: () => void
}
```
- **Features**:
  - Prominent warning callout: "Save this secret now! It will never be displayed again."
  - Secret displayed in `.secret-box` font-mono container.
  - `Copy Secret` button with temporary checked state.
  - `I've Saved My Secret` button to dismiss.

### 4.5 Component: `DeleteAppModal.tsx`
- **File**: `frontend/src/components/app-detail/DeleteAppModal.tsx`
- **Purpose**: Deletion modal with type-app-name safety verification.
- **Props Interface**:
```typescript
export interface DeleteAppModalProps {
  appName: string
  loading: boolean
  onConfirm: () => void
  onCancel: () => void
}
```
- **Features**:
  - Subtitle: `To confirm deletion, please type "${appName}" below:`
  - Verification input: `typedName` state.
  - Validation check: `typedName.trim() === appName`.
  - Confirm button (`Permanently Delete Application`) disabled unless `typedName.trim() === appName`.

---

## 5. Layout & Theme Compatibility
- **Sidebar Integration**: The page renders within `<SidebarLayout />` `<Outlet />`.
- **Navbar Dependency**: NO imported `<Navbar />` or header bar.
- **Design Tokens**: Uses index.css theme variables (`--color-bg`, `--color-card`, `--color-border`, `--color-txt`, `--color-sub`, `--color-primary`, `--color-input`, etc.).

---

## 6. Implementation Strategy for Worker M5

1. **Create Subcomponent Folder**: Create `frontend/src/components/app-detail/`.
2. **Implement Subcomponents**:
   - `SettingsTab.tsx`
   - `SnippetsTab.tsx`
   - `DangerZoneTab.tsx`
   - `SecretDisplayModal.tsx`
   - `DeleteAppModal.tsx`
3. **Refactor `AppDetail.tsx`**:
   - Import subcomponents.
   - Wire top header, tab state, modal handlers, and notification states.
4. **Build & Typecheck Verification**:
   - Run `npm run build` from `frontend/` to confirm 0 TypeScript or Vite build errors.
