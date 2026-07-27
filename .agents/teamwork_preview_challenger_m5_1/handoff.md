# Handoff & Challenge Report — Challenger 1

## 1. Observation

### Build & Backend Suite Verification
- **Frontend Build Execution**:
  - Command: `npm run build` in `/home/uttam/projects/DilithiAuth/frontend`
  - Output:
    ```
    > frontend@0.0.0 build
    > tsc -b && vite build
    vite v8.0.10 building client environment for production...
    dist/index.html                   0.86 kB │ gzip:   0.48 kB
    dist/assets/index-lubqN_EF.css   11.46 kB │ gzip:   2.95 kB
    dist/assets/index-BuPPx7sg.js   360.40 kB │ gzip: 110.78 kB
    ✓ built in 294ms
    ```
  - Artifacts in `frontend/dist`: `index.html` (860 B), `favicon.svg` (2116 B), `icons.svg` (5031 B), `logo-mark.svg` (2116 B), `assets/index-lubqN_EF.css` (11.46 kB), `assets/index-BuPPx7sg.js` (360.40 kB).

- **Backend Test Suite Execution**:
  - Command: `go test -v ./...` in `/home/uttam/projects/DilithiAuth`
  - Output:
    ```
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/oauth      5.984s
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/ratelimit  (cached)
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/signer     (cached)
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/tenant     (cached)
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/transport  0.108s
    ok  github.com/Uttam-Mahata/DilithiAuth/pkg/verifier   0.129s
    ok  github.com/Uttam-Mahata/DilithiAuth/tests          0.479s
    ```
  - Result: 0 test failures or regressions.

### `Docs.tsx` Stress Testing Observations

1. **Clipboard Timer Collision (Lines 297-311)**:
   - Code snippet:
     ```tsx
     setCopiedMap((prev) => ({ ...prev, [key]: true }))
     setTimeout(() => {
       setCopiedMap((prev) => ({ ...prev, [key]: false }))
     }, 2000)
     ```
   - Observation: When a user clicks copy multiple times rapidly, timer handles are not tracked or cleared. Empirical execution of `test_docs_state.js` showed that at `t=2050ms` (1050ms after a second click), `copiedMap['install-ts']` was prematurely cleared to `false`.

2. **Stale Copy Indicator State across Token Strategy Toggle (Lines 967, 971)**:
   - Code snippet:
     ```tsx
     copyToClipboard(`res-body-${selectedEndpoint}`, JSON.stringify(endpointRes.body, null, 2))
     ```
   - Observation: Key is `res-body-token`. If user copies response while `tokenStrategy` is `thin` and then toggles strategy to `full` (Jumbo), `copiedMap['res-body-token']` remains `true` for up to 2s despite the response body displaying completely different content (~3.8 KB Jumbo JWT vs ~500 B Thin JWT).

3. **Unhandled Promise Rejection & False Copy Success (Lines 298-308)**:
   - Code snippet:
     ```tsx
     if (navigator.clipboard && navigator.clipboard.writeText) {
       navigator.clipboard.writeText(text)
     } else { ... }
     setCopiedMap((prev) => ({ ...prev, [key]: true }))
     ```
   - Observation: `writeText` returns a `Promise`. When clipboard permission is denied or in a restricted iframe/non-secure context, the promise rejects asynchronously. Because `writeText` is un-awaited and has no `.catch()`, execution never reaches `execCommand` fallback, and `setCopiedMap` still marks `[key]: true`.

4. **Unencoded `clientId` Query Injection in `/authorize` URL (Line 374)**:
   - Code snippet:
     ```tsx
     url: `/authorize?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}...`
     ```
   - Observation: `clientId` is directly interpolated without `encodeURIComponent()`. Empirical test input `app_123 & client_secret=foo #hash` produced `/authorize?client_id=app_123 & client_secret=foo #hash&redirect_uri=...`.

5. **Double Question Mark (`??`) URL Bug in `/authorize` Redirect Location (Lines 376, 382)**:
   - Code snippet:
     ```tsx
     redirect_to: `${redirectUri}?code=da_code_8f9a2b4c1d6e3f5a7b9c0d2e4f6a8b1c&state=random_csrf_token`
     ```
   - Observation: Input `redirectUri = "https://app.com/cb?env=dev"` produced `https://app.com/cb?env=dev?code=...` with invalid double `?`.

6. **Static JWT Claims vs Dynamic Input Mismatch in `/token` (Lines 394, 414)**:
   - Code snippet:
     ```tsx
     access_token: 'eyJhbGciOiJFZERTQS...eyJpc3Mi...InN1YiI6ImFsaWNlQGV4YW1wbGUuY29tIiwiYXVkIjoiZGlsaXRoaS1kZW1vLWNsaWVudC0xMjM...
     ```
   - Observation: `access_token` base64 payload decodes to `sub: "alice@example.com"` and `aud: "dilithi-demo-client-123"`. Changing `subject` or `clientId` in interactive input controls does not update the `/token` response payload, unlike `/userinfo` and `/authorize`.

7. **Empty Subject Formatting Defect in `/userinfo` (Line 455)**:
   - Code snippet:
     ```tsx
     name: subject.split('@')[0].toUpperCase() + ' Identity'
     ```
   - Observation: Empty subject `""` produces `name: " Identity"` (leading space).

---

## 2. Logic Chain

1. **Observation 1 & 2** show that `copyToClipboard` in `Docs.tsx` uses key strings that do not capture dynamic sub-state changes (such as `tokenStrategy`), and fails to manage timer handles across rapid user interactions.
2. **Observation 3** shows that asynchronous clipboard write errors are swallowed, causing false positive success UI feedback to the user.
3. **Observations 4, 5, 6 & 7** demonstrate that dynamic endpoint response calculations in `getEndpointResponse()` omit URL encoding on `clientId`, fail to check existing query parameters on `redirectUri`, retain static hardcoded JWT claims for `/token`, and suffer from edge-case string splitting on empty subjects.
4. **Observations on Build & Tests** confirm that the core Go backend code and TypeScript compilation pipeline are fully functional and regression-free, proving that the findings are isolated to `Docs.tsx` state logic and UI calculation handling.

---

## 3. Caveats

- Testing was performed using empirical Node test harnesses (`test_docs_state.js`) simulating the exact React state functions in `Docs.tsx`.
- Browser-specific clipboard permission dialog behavior depends on host browser security policies.
- No implementation code was modified in `frontend/src` or `pkg/` in accordance with the review-only constraint.

---

## 4. Conclusion

- **Build Reproducibility**: **VERIFIED (PASS)**. `npm run build` compiles cleanly into `frontend/dist`.
- **Backend Test Suite**: **VERIFIED (PASS)**. `go test ./...` passes all tests across all packages with 0 regressions.
- **`Docs.tsx` State & Snippet Robustness**: **NEEDS REFINEMENT**. 3 copy-to-clipboard state logic issues and 4 dynamic API response state calculation defects were identified and verified empirically.

---

## 5. Verification Method

To independently verify these findings:

1. **Build Reproducibility**:
   ```bash
   cd /home/uttam/projects/DilithiAuth/frontend && npm run build
   ```
2. **Backend Tests**:
   ```bash
   cd /home/uttam/projects/DilithiAuth && go test -v ./...
   ```
3. **Docs.tsx State Harness**:
   ```bash
   node /home/uttam/projects/DilithiAuth/.agents/teamwork_preview_challenger_m5_1/test_docs_state.js
   ```

---

## Challenge Report

### Challenge Summary
**Overall risk assessment**: MEDIUM

### Challenges

#### [Medium] Challenge 1: Unencoded `clientId` & Double `?` in `/authorize` URL Calculations
- **Assumption challenged**: User input `clientId` and `redirectUri` are simple alphanumeric strings without special characters or existing query parameters.
- **Attack scenario**: User inputs a `redirectUri` with existing query parameters (e.g. `https://app.com/callback?env=dev`) or a `clientId` with spaces or ampersands.
- **Blast radius**: Generated documentation URLs and copyable snippets become syntactically invalid URLs.
- **Mitigation**: Use `encodeURIComponent(clientId)` and parse `redirectUri` with `new URL()` to append `code` and `state` parameters cleanly.

#### [Low] Challenge 2: Copy Indicator State Stale / Timer Collision in `Docs.tsx`
- **Assumption challenged**: Copy button key `res-body-${selectedEndpoint}` remains representative of displayed content, and users only click copy once every 2 seconds.
- **Attack scenario**: User switches token strategy from `thin` to `full` on `/token` endpoint within 2 seconds of copying, or clicks copy multiple times rapidly.
- **Blast radius**: UI shows false "Copied" feedback for response content that was never copied, or prematurely clears copy feedback.
- **Mitigation**: Include `tokenStrategy` in copy key (e.g., `res-body-${selectedEndpoint}-${tokenStrategy}`), track timer refs, and handle `writeText()` promise rejections.

## Unchallenged Areas
- OIDC Discovery Document and JWKS endpoints — tested and structurally sound.
