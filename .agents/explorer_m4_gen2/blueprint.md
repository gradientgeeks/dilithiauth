# Blueprint: Milestone 4 (R3 Create Application Page - NewApp.tsx Wizard Redesign)

## 1. Overview & Goal

The objective of Milestone 4 is to transform the single-page application creation form in `frontend/src/pages/NewApp.tsx` into a multi-step, Auth0-style guided wizard (`NewApp.tsx`) with modular subcomponents in `frontend/src/components/new-app/`.

### Key Design Principles:
1. **Guided Multi-Step Wizard**: Divide application registration into three clear, digestible steps:
   - **Step 1: Basic Info** (App Name, Description, Homepage URL)
   - **Step 2: Application Type** (Visual card selector for Public SPA vs Confidential Server-Side App)
   - **Step 3: URIs & Configuration** (Interactive multi-input list for Redirect URIs, review summary, and submission)
2. **Auth0 Console Dark Aesthetic**: Modern dark theme utilizing DilithiAuth CSS design tokens (`var(--color-bg)`, `var(--color-card)`, `var(--color-border)`, `var(--color-primary)`), card-based option selector with glowing focus borders and badges.
3. **Post-Creation Credentials Overlay**: Reusable `SecretSuccessModal` component that pops up after `apps.create()` succeeds, presenting Client ID & Client Secret with 1-click clipboard copy, amber warning banner for non-retrievable secret, and navigation CTAs.
4. **Layout Integration**: Fully compatible inside `<SidebarLayout />` without any top `<Navbar />` imports or layout collisions.

---

## 2. File & Directory Architecture

```
frontend/src/
├── components/
│   └── new-app/
│       ├── TypeSelector.tsx          # Visual card selector for client types (Public vs Confidential)
│       ├── SecretSuccessModal.tsx    # Modal overlay displaying Client ID & Client Secret post-creation
│       └── WizardStepper.tsx         # Visual progress indicator bar / step header
├── pages/
│   └── NewApp.tsx                    # Wizard controller page managing state, steps, and API submission
└── __tests__/
    ├── m4_verification_test.tsx      # Automated SSR test suite for M4 components & wizard
    └── run_m4_test.mjs               # Vite SSR runner script for M4 test suite
```

---

## 3. Component Specifications & Interfaces

### 3.1 `TypeSelector.tsx` (`frontend/src/components/new-app/TypeSelector.tsx`)

A visual card selector component allowing developers to choose between Confidential Server-Side Client and Public SPA / Native Client.

#### Interface:
```typescript
export interface TypeSelectorProps {
  isPublic: boolean
  onChange: (isPublic: boolean) => void
}
```

#### Visual Card Design Specs:
- **Card 1: Confidential Server-Side App**
  - `isPublic: false`
  - Icon: `Lock` or `ShieldCheck` (blue/purple tone)
  - Title: **Confidential Server-Side App**
  - Tag: `Client Secret Flow`
  - Description: *"For Node.js, Go, Python, or Java web applications running on a server. Secures credentials via client secret."*
- **Card 2: Public SPA / Native App**
  - `isPublic: true`
  - Icon: `Globe` or `Laptop` (cyan/sky tone)
  - Title: **Public SPA / Native App**
  - Tag: `PKCE Authorization Flow`
  - Description: *"For Single Page Apps (React, Vue), Mobile Apps (iOS/Android), or CLI tools. Uses PKCE without backend secrets."*
- **State & Styling**:
  - Active selected card gets active border (`border-blue-500/60`), subtle background glow (`bg-blue-500/10` or `bg-cyan-500/10`), active checkmark badge on top right, and bright title color.
  - Non-selected card uses neutral dark styling (`bg-[var(--color-input)] border-[var(--color-border)]`) with hover border transitions (`hover:border-[var(--color-border-mid)]`).

---

### 3.2 `SecretSuccessModal.tsx` (`frontend/src/components/new-app/SecretSuccessModal.tsx`)

A modal popup layer rendered over the layout immediately following successful application registration when `client_secret` is present.

#### Interface:
```typescript
import type { OAuthApp } from '../../api'

export interface SecretSuccessModalProps {
  app: OAuthApp
  secret?: string
  onClose: () => void
  onNavigateToApp: (appId: string) => void
  onNavigateToDashboard: () => void
}
```

#### Modal Layout & Features:
1. **Backdrop**: `fixed inset-0 z-[500] bg-black/75 backdrop-blur-sm flex items-center justify-center p-4`
2. **Card Container**: `bg-[var(--color-card)] border border-[var(--color-border-mid)] rounded-xl p-6 sm:p-8 max-w-lg w-full shadow-2xl animate-fadeIn`
3. **Header**: Success icon (`CheckCircle2` in emerald glow), title *"Application Created Successfully!"*, subtitle showing app name and client type.
4. **Client ID Box**: Masked/formatted string with 1-click Copy button and feedback tooltip state ("Copied!").
5. **Client Secret Box (if `secret` exists)**:
   - **Amber Warning Banner**: `AlertTriangle` icon + *"Save your Client Secret now! It will never be shown again."*
   - Pre-styled code box (`secret-box`) with full secret text.
   - 1-click **Copy Secret** button (`navigator.clipboard.writeText`) with instant feedback state.
6. **Public App Info Banner (if `isPublic` is true)**:
   - `ShieldCheck` icon + *"Registered as a Public Client (PKCE mode). No client secret is required."*
7. **Action Buttons**:
   - Primary CTA: *"Continue to App Details →"* (`onNavigateToApp(app.ID)`)
   - Secondary CTA: *"Back to Dashboard"* (`onNavigateToDashboard()`)

---

### 3.3 `WizardStepper.tsx` (`frontend/src/components/new-app/WizardStepper.tsx`)

Visual progress bar displayed at the top of the wizard.

#### Interface:
```typescript
export interface WizardStep {
  id: number
  title: string
  subtitle: string
}

export interface WizardStepperProps {
  currentStep: number
  steps: WizardStep[]
  onStepClick?: (stepId: number) => void
}
```

#### Steps Definition:
- Step 1: **Basic Details** (*Name, Description & Homepage*)
- Step 2: **Application Type** (*Confidential vs Public PKCE*)
- Step 3: **URIs & Review** (*Redirect URIs & Finalize*)

#### Stepper Visual States:
- **Completed Step**: Step number replaced by green checkmark icon (`Check`), line separator highlighted blue/emerald.
- **Active Step**: Number circled in `var(--color-primary)` with glowing outline, bold white text.
- **Upcoming Step**: Muted gray circle and text (`text-[var(--color-muted)]`).

---

### 3.4 `NewApp.tsx` Wizard Controller Page

The main page component orchestrating wizard state, form inputs, validation, step navigation, and API call.

#### State Structure:
```typescript
const [step, setStep] = useState<1 | 2 | 3>(1)
const [name, setName] = useState('')
const [description, setDescription] = useState('')
const [homepageURL, setHomepageURL] = useState('')
const [isPublic, setIsPublic] = useState(false) // Default to Confidential
const [redirectURIs, setRedirectURIs] = useState<string[]>(['http://localhost:3000/callback'])
const [loading, setLoading] = useState(false)
const [error, setError] = useState('')
const [createdResult, setCreatedResult] = useState<{ app: OAuthApp; secret?: string } | null>(null)
```

#### Step Progression & Validation Logic:
- **Step 1 -> Step 2**:
  - Validates `name.trim() !== ''`. If empty, shows error `"Application Name is required."`.
  - Next button: *"Next: Application Type →"*
- **Step 2 -> Step 3**:
  - Validates selection.
  - Back button: *"← Back to Basic Info"*
  - Next button: *"Next: URIs & Config →"*
- **Step 3 -> Submit**:
  - Dynamic URI list management (Add URI input field, remove URI button).
  - Validates at least 1 non-empty URI.
  - Back button: *"← Back to App Type"*
  - Submit button: *"Create Application"* (displays `Loader2` spin during API call).

#### API Call Execution:
```typescript
const handleSubmit = async () => {
  setError('')
  const validURIs = redirectURIs.map(u => u.trim()).filter(Boolean)
  if (validURIs.length === 0) {
    setError('At least one redirect URI is required.')
    return
  }
  setLoading(true)
  try {
    const res = await apps.create({
      name: name.trim(),
      description: description.trim(),
      homepage_url: homepageURL.trim(),
      redirect_uris: validURIs,
      is_public: isPublic,
    })
    setCreatedResult({ app: res.data.app, secret: res.data.client_secret })
  } catch (err: any) {
    setError(err?.response?.data?.error ?? 'Failed to create application.')
  } finally {
    setLoading(false)
  }
}
```

---

## 4. Layout Compatibility Verification

- `NewApp.tsx` is mounted within `<SidebarLayout />` at route `/apps/new`.
- **Navbar Absence**: Does NOT import or render `Navbar.tsx`.
- **Responsive Width**: Centered workspace layout (`max-w-[760px] mx-auto px-4 py-6 sm:py-8`).
- **Dark Theme Tokens**: Uses `--color-bg`, `--color-card`, `--color-border`, `--color-txt`, `--color-sub`, `--color-primary` consistent with `SidebarLayout` and `Dashboard.tsx`.

---

## 5. Automated Verification Test Suite Design

Worker M4 will create `frontend/src/__tests__/m4_verification_test.tsx` and `frontend/src/__tests__/run_m4_test.mjs`.

### Test Checks to Cover:
1. **Module Exports**: Verify `TypeSelector`, `SecretSuccessModal`, `WizardStepper`, and `NewApp` default/named exports.
2. **TypeSelector Rendering**: Verify rendering of Confidential vs Public cards, icons, descriptions, and click callback execution.
3. **SecretSuccessModal Rendering**:
   - Verify modal renders Client ID and Client Secret.
   - Verify Amber alert rendered for confidential app with secret.
   - Verify Blue PKCE info alert rendered for public app without secret.
   - Verify Copy buttons and navigation callbacks.
4. **WizardStepper Rendering**:
   - Verify 3 step titles rendered.
   - Verify active step highlights step number/badge.
5. **NewApp Page Integration & Step Switching**:
   - Verify initial Step 1 render.
   - Verify validation blocking Step 1 progression when name is empty.
   - Verify URI manager list add/remove interactions.
