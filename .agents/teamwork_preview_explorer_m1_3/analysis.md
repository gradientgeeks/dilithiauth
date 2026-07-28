# Developer Portal Redesign: Technical Design & Architecture Proposal

## 1. Executive Summary & Core Objectives

This specification outlines the technical design and architectural proposal for redesigning the DilithiAuth Developer Portal (`frontend/src`). The redesign transitions the interface from a simple top-navbar layout to an **Auth0-inspired, enterprise-grade developer console**. 

The redesign preserves all post-quantum authorization mechanics (ML-DSA-44 / FIPS 204 + Ed25519 hybrid tokens, PKCE flows, thin/jumbo token strategies) while elevating usability, visual polish, layout responsiveness, and security workflows.

---

## 2. Architectural Blueprint & Component Hierarchy

```
+-----------------------------------------------------------------------------------+
| SidebarLayout (Layout Container)                                                  |
| +-------------------------+ +---------------------------------------------------+ |
| | Sidebar                 | | Header Bar (Mobile top-bar / Desktop context bar)  | |
| | - Logo (DilithiAuth)    | +---------------------------------------------------+ |
| | - PQC IdP Badge         | | Main Content Outlet (`<Outlet />` or Page)        | |
| | - Collapse Toggle       | |                                                   | |
| |                         | | [ Dashboard.tsx | NewApp.tsx | AppDetail.tsx ]     | |
| | Nav Links:              | |                                                   | |
| | - Applications (/apps)  | |                                                   | |
| | - Docs (/docs)          | |                                                   | |
| | - Settings (Disabled)   | |                                                   | |
| |                         | |                                                   | |
| | Footer:                 | |                                                   | |
| | - User Profile Pill     | |                                                   | |
| | - Sign Out Button       | |                                                   | |
| +-------------------------+ +---------------------------------------------------+ |
+-----------------------------------------------------------------------------------+
```

---

## 3. Detailed Component Specifications

### 3.1 Requirement 1: `SidebarLayout.tsx` (R1)

`SidebarLayout.tsx` serves as the primary layout wrapper for authenticated routes (`/apps`, `/apps/new`, `/apps/:id`, `/docs` when authenticated).

#### Design Specs & Layout States
- **Desktop (>= 1024px / `lg`)**:
  - Persistent left sidebar with width `w-64` (256px) in expanded mode, or `w-16` (64px) in collapsed rail mode.
  - Smooth CSS transition for width expand/collapse (`transition-all duration-300 ease-in-out`).
  - Collapse toggle button at bottom or header section to switch between full sidebar and icon-only rail.
- **Tablet (768px to 1023px / `md:max-lg`)**:
  - Default icon-only left rail (`w-16` / 64px width).
  - Tooltips display link labels on hover (`group-hover:opacity-100`).
- **Mobile (< 768px / `<md`)**:
  - Top header bar (`h-14` / 56px) featuring brand logo and hamburger menu button.
  - Off-canvas drawer sliding from left (`-translate-x-full` to `translate-x-0`) with dark backdrop blur (`bg-black/60 backdrop-blur-sm`).
  - Automatic drawer closure on route change or backdrop click / `Escape` key press.

#### Elements & Features
1. **Brand Header**:
   - Logo (`/favicon.svg`, 28x28px).
   - Text "DilithiAuth" with a subtle post-quantum badge (`PQC IdP` / `ML-DSA-44` in cyan/emerald text).
2. **Navigation Items**:
   - **Applications** (`/apps`): `LayoutGrid` icon. Active when `pathname.startsWith('/apps')`.
   - **Docs & SDKs** (`/docs`): `BookOpen` icon. Active when `pathname === '/docs'`.
   - **Settings** (`/settings`): `Settings` icon. Disabled state with tooltip *"Settings coming soon"* and `cursor-not-allowed opacity-50`.
3. **Bottom Profile & Actions**:
   - Developer initials avatar badge (e.g. `AQ` for Alice Quantum).
   - Truncated developer email (`developer.email`).
   - Sign Out button: `LogOut` icon, red hover background accent (`hover:bg-rose-500/10 hover:text-rose-400`), invokes `auth.logout()` and redirects to `/login`.

---

### 3.2 Requirement 2: `Dashboard.tsx` (R2)

`Dashboard.tsx` provides an Auth0-console aesthetic application overview and management interface.

#### Design Specs & Sections
1. **Console Header**:
   - Title: **Applications**
   - Subtitle: *"Manage your OAuth 2.0 applications, client credentials, and Post-Quantum hybrid token settings."*
   - CTA Button: **+ Create Application** (Primary blue button linking to `/apps/new`).
2. **Metrics Ribbon (Stat Cards)**:
   - Total Registered Apps
   - Public PKCE Clients (Cyan badge icon)
   - Confidential Clients (Emerald badge icon)
   - Post-Quantum Status (*FIPS 204 ML-DSA-44 Active*)
3. **Search & Filter Controls**:
   - Real-time search bar filtering by `Name` and `ClientID`.
   - Filter pills: `All`, `Confidential`, `Public (PKCE)`.
   - View mode toggle: `Table View` (default) vs `Card Grid View`.
   - Active count indicator: *"Showing X of Y applications"*.
4. **App List Table & Card Grid**:
   - **Masked Client ID**: Displayed as `client_...3f8a` or `••••••••••••3f8a` with a toggle eye button to unmask, plus a 1-click **Quick Copy** button with toast/checkmark feedback.
   - **Client Type Badges**:
     - *Confidential*: Emerald border/bg badge + `Lock` icon.
     - *Public (PKCE)*: Sky/Cyan border/bg badge + `Globe` icon.
   - **Actions**: "Manage" link button leading to `/apps/:id`.
5. **Empty State Illustration**:
   - High-contrast empty state container for zero-search or zero-app states.
   - Vector/Icon graphic with message: *"No OAuth applications found"*.
   - CTA button to register the first app.

---

### 3.3 Requirement 3: `NewApp.tsx` (R3)

`NewApp.tsx` adopts a multi-step wizard form pattern with visual card selection and a post-creation credential disclosure modal.

#### Design Specs & Wizard Steps
1. **Stepper Progress Header**:
   - `1. Basic Info` -> `2. App Type` -> `3. OAuth URIs`
2. **Step 1: Basic Information**:
   - Application Name (Required text input).
   - Description (Optional textarea).
   - Homepage URL (Optional text input).
3. **Step 2: Visual Card Selector (Client Type)**:
   - Two large interactive cards:
     - **Confidential Client (Server-Side Web App)**:
       - Icon: `Lock` / `Server`.
       - Description: *"For traditional web apps with server backends capable of securely storing client secrets."*
       - Badges: `Client Secret` | `Auth Code Flow`
     - **Public Client (SPA / Mobile / Native)**:
       - Icon: `Globe` / `Smartphone`.
       - Description: *"For Single Page Apps (React/Vue), Mobile Apps, or CLI tools executing on untrusted client environments."*
       - Badges: `PKCE Flow` | `No Secret Required`
   - Active state styling: Cyan/Blue glow border (`border-sky-500 bg-sky-500/10`), radio check mark indicator.
4. **Step 3: OAuth URIs & Review**:
   - Line-separated Redirect URIs text field with URL validation.
   - Pre-submit summary review card.
5. **Post-Creation Success Screen / Modal**:
   - Triggered upon successful response from `apps.create(...)`.
   - **Client ID Box**: Monospace box with 1-click copy button.
   - **Client Secret Disclosure Box (Confidential Apps Only)**:
     - Prominent Amber Alert Banner: **"Save your Client Secret now! This secret will NEVER be displayed again."**
     - Secret display with show/hide toggle and 1-click copy button.
     - "Download Credentials (.json)" action.
   - **Public Client Notice (Public Apps Only)**:
     - Blue Info Banner explaining PKCE requirement.
   - Action CTA buttons: **"Go to App Settings"** and **"Back to Dashboard"**.

---

### 3.4 Requirement 4: `AppDetail.tsx` (R4)

`AppDetail.tsx` provides a tabbed settings console with sub-navigation for comprehensive application management.

#### Design Specs & Sub-Navigation Tabs
1. **Header & Navigation Breadcrumbs**:
   - Breadcrumbs: `Applications / {app.Name}`
   - Header with App Name, Client ID quick copy, and type badge.
2. **Sub-Nav Tab Bar**:
   - **Settings & Credentials** (`KeyRound` icon)
   - **OAuth URIs & Config** (`Globe` icon)
   - **SDK & Code Snippets** (`Code` icon)
   - **Danger Zone** (`AlertTriangle` icon)
3. **Tab 1: Credentials & General Settings**:
   - **Client Credentials Panel**: Client ID copy, Client Secret status.
   - **Secret Rotation (Confidential Apps)**:
     - "Rotate Client Secret" button.
     - Confirmation Modal (`ConfirmModal`) warning of API disruption.
     - Temporary display box for newly rotated secret with 1-click copy button and one-time notice.
   - **General Form Groups**: Name, Description, Homepage URL editable fields with "Save Changes" action.
4. **Tab 2: OAuth URIs Configuration**:
   - Interactive Redirect URIs manager (add URI row, delete URI row, URL validation).
5. **Tab 3: SDK & Code Snippets Generator**:
   - Language selector buttons: `cURL`, `JavaScript / Node`, `Go`, `Python`, `Java`, `.NET`.
   - Pre-filled authorization code exchange code blocks incorporating `app.ClientID` and `app.RedirectURIs[0]`.
   - Post-Quantum Token Architecture callout explaining hybrid signature verification via `/.well-known/jwks.json`.
6. **Tab 4: Danger Zone**:
   - Red-bordered danger container (`border-rose-500/30 bg-rose-500/5`).
   - "Delete Application" action triggering `ConfirmModal` with double-confirmation dialog (requires user confirmation before executing `apps.delete(id)`).

---

## 4. Requirement 5: Component File Structure Specification

The exact modular file structure to create and update in `frontend/src`:

```
frontend/src/
├── components/
│   ├── layout/
│   │   ├── SidebarLayout.tsx        # Responsive persistent sidebar layout container
│   │   ├── SidebarNav.tsx           # Nav links, brand header, rail toggle & profile
│   │   └── MobileHeader.tsx         # Mobile top bar & off-canvas drawer trigger
│   ├── common/
│   │   ├── Badge.tsx                # Reusable type/status badges (Confidential, Public, PQC)
│   │   ├── CopyButton.tsx           # Monospace copy button with feedback state
│   │   ├── ConfirmModal.tsx         # Shared double-confirmation modal dialog
│   │   ├── EmptyState.tsx           # Graphic empty state placeholder
│   │   └── StatCard.tsx             # Metric statistic card component
│   ├── dashboard/
│   │   ├── AppTable.tsx             # Tabular layout for application list
│   │   ├── AppCardGrid.tsx          # Card grid view layout for application list
│   │   └── DashboardHeader.tsx      # Search input, type filter pills, view toggles
│   ├── new-app/
│   │   ├── Stepper.tsx              # Wizard step progress indicator
│   │   ├── AppTypeSelector.tsx      # Interactive card selector for Public vs Confidential
│   │   └── SuccessModal.tsx         # Post-creation credentials disclosure modal
│   └── app-detail/
│       ├── CredentialsCard.tsx      # Client ID & Secret rotation section
│       ├── GeneralSettingsForm.tsx  # General settings & Redirect URIs form
│       ├── CodeSnippetsTab.tsx      # Multi-language OAuth 2.0 code generator
│       └── DangerZoneTab.tsx        # Application deletion card with confirm modal
├── pages/
│   ├── Dashboard.tsx                # Main developer applications dashboard (R2)
│   ├── NewApp.tsx                   # Application creation wizard (R3)
│   ├── AppDetail.tsx                # Application settings & credentials console (R4)
│   ├── Docs.tsx                     # Developer documentation page (wrapped in SidebarLayout)
│   ├── Landing.tsx                  # Public marketing landing page
│   ├── Login.tsx                    # Developer login page
│   └── Register.tsx                 # Developer registration page
├── api.ts                           # Axios API client & type definitions
├── AuthContext.tsx                  # React Context for authentication state
├── App.tsx                          # Router configuration & protected routes
└── index.css                        # Design system tokens & Tailwind v4 styles
```

---

## 5. State Management & API Contracts

### Data Types (`api.ts`)
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

export interface CreateAppPayload {
  name: string
  description: string
  homepage_url: string
  redirect_uris: string[]
  is_public: boolean
}

export interface CreateAppResponse {
  app: OAuthApp
  client_secret?: string
}

export interface RotateSecretResponse {
  app: OAuthApp
  client_secret: string
}
```

---

## 6. Verification & Test Plan

1. **Layout & Responsiveness**:
   - Verify desktop view (>=1024px) renders expanded sidebar / toggleable rail.
   - Verify tablet view (768px-1023px) defaults to icon rail.
   - Verify mobile view (<768px) renders mobile header and sliding drawer.
2. **Auth & Route Guarding**:
   - Ensure `/apps`, `/apps/new`, `/apps/:id` remain protected behind `RequireAuth`.
   - Ensure `/docs` is accessible to both guest and authenticated users with appropriate layout context.
3. **Application Lifecycle Operations**:
   - Test application registration via wizard (Public and Confidential types).
   - Test one-time secret display upon creation and rotation.
   - Test Client ID masking and quick copy feedback.
   - Test application deletion with modal confirmation.
