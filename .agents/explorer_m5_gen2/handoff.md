# Handoff Report: Explorer M5 — App Detail Page Redesign Blueprint

## 1. Observation

### 1.1 Existing Layout & Implementation Analysis
- **Target File**: `frontend/src/pages/AppDetail.tsx` (Total lines: 396). Currently a monolithic component handling loading, app fetching (`apps.get(id)`), credentials display, basic code snippets, and modal confirmations.
- **Routing Shell**: `frontend/src/App.tsx` routes `/apps/:id` inside `<RequireAuth><SidebarLayout /></RequireAuth>`. There is no `<Navbar />` reference in `AppDetail.tsx`.
- **API Integration**: `frontend/src/api.ts` provides `apps.get`, `apps.create`, `apps.rotateSecret`, `apps.delete`. `apps.update` should be added to `api.ts` for form saving.
- **Design Tokens**: `frontend/src/index.css` provides CSS variables (`--color-bg`, `--color-card`, `--color-border`, `--color-txt`, `--color-sub`, `--color-primary`, `--color-input`) and utility classes (`.code-block`, `.secret-box`, `.mono-inline`, `.skeleton`).

### 1.2 Identified Redesign Targets
1. **Header & Sub-Nav Layout**: Header displaying App Name, Client ID badge, App Type badge (`Public SPA (PKCE)` vs `Confidential Client`), and back link (`/apps`). Sub-nav bar switching between **Settings**, **Code Snippets**, and **Danger Zone**.
2. **Settings Tab**: Form groups for App Name, Description, Homepage URL, interactive multi-row Redirect URIs list manager (`+ Add URI`, row trash button), Client ID display with copy action, Client Type indicator, and Save Changes submit button.
3. **Code Snippets Tab**: OAuth integration code snippets viewer with language switcher (cURL, JS / Fetch, Go, Python) highlighting DilithiAuth endpoints (`/token/thin`, `/token/full`, `/token/cwt`).
4. **Danger Zone Tab**: Redesigned danger box with "Rotate Client Secret" button (opens secret rotation modal displaying newly generated secret with copy button & warning) and "Delete Application" button (opens confirmation modal with type-app-name validation).

---

## 2. Logic Chain

1. **Routing Alignment**: `AppDetail.tsx` is an `<Outlet />` inside `SidebarLayout.tsx`. It must maintain standard page outer margins (`max-w-6xl mx-auto px-4 sm:px-6 py-8`) and omit top Navbar elements.
2. **Component Modularization**: Splitting `AppDetail.tsx` into modular subcomponents inside `src/components/app-detail/` improves code maintainability and testability:
   - `src/components/app-detail/SettingsTab.tsx`
   - `src/components/app-detail/SnippetsTab.tsx`
   - `src/components/app-detail/DangerZoneTab.tsx`
   - `src/components/app-detail/SecretDisplayModal.tsx`
   - `src/components/app-detail/DeleteAppModal.tsx`
3. **Safety & UX Enhancements**: Adding type-app-name validation to the deletion modal prevents destructive accidental deletions. Displaying specific token exchange endpoints (`/token/thin`, `/token/full`, `/token/cwt`) in snippets directly educates developers on DilithiAuth's post-quantum hybrid architecture.

---

## 3. Caveats
- **Backend API for App Editing**: The Go backend (`pkg/developer/handlers.go`) currently provides `GET /apps`, `POST /apps`, `GET /apps/{id}`, `POST /apps/{id}/rotate-secret`, and `DELETE /apps/{id}`. If full backend edit (`PUT /apps/{id}`) is not implemented on backend, `SettingsTab` updates local state and provides user feedback gracefully.
- **No Direct Source Changes**: As Explorer M5, this investigation is read-only. All blueprint specifications are delivered to Worker M5 for implementation.

---

## 4. Conclusion
The comprehensive blueprint (`blueprint.md`) and component breakdown provide exact props interfaces, component contracts, snippet definitions, and styling specifications for Worker M5 to execute the Milestone 5 redesign cleanly.

---

## 5. Verification Method

### 5.1 Inspection Targets
- Inspection of `frontend/src/pages/AppDetail.tsx` to verify clean orchestration of header, tabs, and subcomponents.
- Inspection of subcomponents in `frontend/src/components/app-detail/`:
  - `SettingsTab.tsx`
  - `SnippetsTab.tsx`
  - `DangerZoneTab.tsx`
  - `SecretDisplayModal.tsx`
  - `DeleteAppModal.tsx`

### 5.2 Build Command
From `frontend/`:
```bash
npm run build
```
Verify 0 TypeScript compiler errors and 0 Vite build errors.
