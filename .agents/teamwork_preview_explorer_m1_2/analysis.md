# Milestone 1: Component Logic & API Bindings Analysis

**Date:** 2026-07-28  
**Scope:** Frontend Application Logic & API Bindings (`frontend/src`)  
**Target Files Analyzed:**
- `src/api.ts` (API Client & Type Definitions)
- `src/pages/Dashboard.tsx` (App Listing, Search, & Stats)
- `src/pages/NewApp.tsx` (App Registration & One-time Secret Display)
- `src/pages/AppDetail.tsx` (App Details, Credentials, SDK Snippets, Secret Rotation, Deletion)

---

## 1. API Module & Data Models (`src/api.ts`)

### 1.1 HTTP Client Configuration
- **Base URL:** Defaults to `import.meta.env.VITE_API_BASE_URL` or fallback `'https://dilithiauthapis.gradientgeeks.tech/api/v1/developer'`.
- **Credentials:** Created using `axios.create({ baseURL: apiBaseURL, withCredentials: true })`.
- **CSRF Defense:**
  - In-memory CSRF token store (`let csrfToken: string | undefined`).
  - Captured on `auth.login`, `auth.verifyEmail`, and `auth.me` responses.
  - Request interceptor automatically attaches `X-CSRF-Token` header for state-changing HTTP methods (`POST`, `PUT`, `DELETE`, `PATCH`).

### 1.2 Type Definitions
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

export interface Developer {
  id: string
  email: string
  name: string
  csrf_token?: string
}
```

### 1.3 `apps` Object Method Signatures
| Method | HTTP Verb & Path | Request Parameters / Payload | Response Type | Notes |
| :--- | :--- | :--- | :--- | :--- |
| `apps.list()` | `GET /apps` | None | `Promise<AxiosResponse<OAuthApp[]>>` | Retrieves all OAuth 2.0 apps owned by developer |
| `apps.get(id)` | `GET /apps/${id}` | `id: string` (App ID) | `Promise<AxiosResponse<OAuthApp>>` | Retrieves single application details |
| `apps.create(data)` | `POST /apps` | `{ name: string; description: string; homepage_url: string; redirect_uris: string[]; is_public: boolean }` | `Promise<AxiosResponse<{ app: OAuthApp; client_secret?: string }>>` | Returns `client_secret` only for confidential apps |
| `apps.rotateSecret(id)` | `POST /apps/${id}/rotate-secret` | `id: string` (App ID) | `Promise<AxiosResponse<{ app: OAuthApp; client_secret: string }>>` | Rotates secret and returns new `client_secret` |
| `apps.delete(id)` | `DELETE /apps/${id}` | `id: string` (App ID) | `Promise<AxiosResponse<any>>` | Revokes app credentials and deletes app |

*Note: In `src/api.ts`, types are named `OAuthApp` (not `Application`), input payload for create is defined inline, and output payload for secret rotation is defined inline.*

---

## 2. Dashboard Component Logic (`src/pages/Dashboard.tsx`)

### 2.1 State Structure
- `appList`: `OAuthApp[]` (State holding list of applications)
- `loading`: `boolean` (Loading state during initial API call)
- `error`: `string` (Error message banner string)
- `search`: `string` (Search query string for filtering)
- `copiedId`: `string | null` (Tracks ClientID currently copied to clipboard for UI checkmark feedback)

### 2.2 Execution & Data Fetching
- `useEffect` runs once on component mount (lines 15-20):
  ```typescript
  apps.list()
    .then(r => setAppList(r.data ?? []))
    .catch(() => setError('Failed to load applications.'))
    .finally(() => setLoading(false))
  ```

### 2.3 Search & Filter Logic
- Real-time case-insensitive filter applied to `appList` (lines 29-32):
  ```typescript
  const filteredApps = appList.filter(app =>
    app.Name.toLowerCase().includes(search.toLowerCase()) ||
    app.ClientID.toLowerCase().includes(search.toLowerCase())
  )
  ```
- **Stats Counter Logic:**
  - Total Apps: `appList.length`
  - Public PKCE Clients: `appList.filter(a => a.IsPublic).length`
  - Confidential Clients: `appList.filter(a => !a.IsPublic).length`
  - Active Cryptography Protocol: Static indicator for `ML-DSA-44` (FIPS 204).

### 2.4 UI Features & Copy Handling
- `copyClientId(id: string, e: React.MouseEvent)`: Calls `navigator.clipboard.writeText(id)`, sets `copiedId = id`, resets after 2000ms. Stops click propagation to prevent table row navigation.
- **Rendering States:**
  - Loading: Animated skeleton cards (`.skeleton h-4`).
  - Empty (No apps or no search matches): Empty state graphic with conditional call-to-action button.
  - Table: Displays Name, Description, Client ID (with copy button), Type badge (Public vs Confidential), Creation Date (formatted via `toLocaleDateString`), and Details action link.

---

## 3. New Application Component Logic (`src/pages/NewApp.tsx`)

### 3.1 Form State Management
- `name`: `string` (Required)
- `description`: `string` (Optional)
- `homepageURL`: `string` (Optional)
- `redirectURIs`: `string` (Required, multiline string parsed into array)
- `isPublic`: `boolean` (Default: `false` for Confidential Client; `true` for Public Client/PKCE)
- `loading`: `boolean` (Form submission state)
- `error`: `string` (Submission error text)
- `created`: `CreatedResult | null` where `interface CreatedResult { app: OAuthApp; secret?: string }`
- `copied`: `boolean` (Copy feedback for Client Secret)

### 3.2 Form Submission & Validation Logic
- Handler `handleSubmit` (lines 26-40):
  1. Prevents default form submit.
  2. Splits `redirectURIs` input by newline, trims each entry, and filters out empty lines.
  3. Validates that at least one URI exists; sets error if zero URIs.
  4. Calls `apps.create({ name, description, homepage_url: homepageURL, redirect_uris: uris, is_public: isPublic })`.
  5. On success: sets `created` state with returned `app` object and `client_secret` (if confidential).
  6. On failure: extracts error message from `err?.response?.data?.error` or defaults to `'Failed to create application.'`.

### 3.3 Post-Creation / Secret Display Screen
- When `created` state is present, screen transitions to success state (lines 50-118):
  - Displays `ClientID`.
  - **Secret Handling:** If `created.secret` exists, renders an amber alert banner ("Save this Client Secret now! It will **never be shown again**") and a styled `<pre className="secret-box">`.
  - Copy action: `copySecret()` writes `created.secret` to clipboard and provides a 2-second checkmark notification.
  - If `isPublic` is true, displays info banner explaining PKCE mode requirement without client secrets.
  - Actions: "View App Details →" (`navigate(/apps/${created.app.ID})`) or "Dashboard" (`navigate('/apps')`).

---

## 4. Application Detail Component Logic (`src/pages/AppDetail.tsx`)

### 4.1 Data Fetching & State Management
- `id`: Route parameter from `useParams<{ id: string }>()`.
- `app`: `OAuthApp | null`.
- `loading`: `boolean` (Initial fetch).
- `error`: `string` (Error alert banner).
- `activeTab`: `'overview' | 'snippets' | 'danger'` (Tab state, default `'overview'`).
- `codeLang`: `'curl' | 'javascript' | 'go' | 'python' | 'java' | 'dotnet'` (SDK code snippet language selector).

### 4.2 Lifecycle & API Integration
- **Fetch App (`apps.get`):** Runs `useEffect` on `id` change (lines 36-39). Calls `apps.get(id)`, updates `app` or sets `'App not found.'` on error.
- **Rotate Secret (`apps.rotateSecret`):**
  - Handler `handleRotate` (lines 41-51): Confirmed via `ConfirmModal`. Sets `rotating` state to true. Calls `apps.rotateSecret(id)`.
  - Updates `app` state with new app object and stores `client_secret` in `rotatedSecret` state.
  - UI displays amber warning banner and one-time secret view until user clicks "I've Saved My Secret".
- **Delete Application (`apps.delete`):**
  - Handler `handleDelete` (lines 53-58): Confirmed via `ConfirmModal`. Sets `deleting` state to true. Calls `apps.delete(id)`.
  - On success, redirects to `/apps` via `navigate('/apps')`.
- **Edit Settings Handling:**
  - **Analysis Finding:** `AppDetail.tsx` currently **does NOT implement application editing**.
  - Application Configuration (Homepage URL, Creation Date, Redirect URIs) is displayed read-only in Tab 1.
  - No `apps.update` method exists in `src/api.ts` nor is there an edit modal/form in `AppDetail.tsx`.

### 4.3 Tab Features & Code Generator
- **Tab 1: Overview:** Client ID copy button, Client Secret rotation / PKCE info, and read-only Application Configuration.
- **Tab 2: Code Snippets:** Dynamically builds OAuth 2.0 / Post-Quantum authorization code exchange snippets for 6 programming languages (`cURL`, `JS/Node`, `Go`, `Python`, `Java`, `.NET`) populated with the current app's `ClientID` and `RedirectURI`.
- **Tab 3: Danger Zone:** Permanent deletion trigger with modal confirmation.

---

## 5. Summary of Findings & Next Steps (Milestone 1 Context)

1. **Type Naming Consistency:** API module uses `OAuthApp` type.
2. **Missing Edit Feature:** Application updating functionality (editing name, description, homepage URL, or redirect URIs) is currently unhandled in both `api.ts` (`apps` object) and `AppDetail.tsx`.
3. **Secret Security Model:** Both `NewApp.tsx` and `AppDetail.tsx` follow strict ephemeral display patterns for client secrets, storing them in component state only after creation or rotation and warning the user that secrets are shown only once.
