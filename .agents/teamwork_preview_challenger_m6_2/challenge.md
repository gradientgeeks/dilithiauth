# Adversarial Challenge Report — Social Authentication Stress Testing

## Challenge Summary

**Overall Risk Assessment**: **HIGH**
**Verdict**: **FAIL** (Requires remediations for Provider Error Propagation and Provisioning Error Handling)

---

## Task Verification & Empirical Results

### 1. Build & Core Test Suite Verification
- `go test ./...`: **PASS** (All packages pass cleanly).
- `go build -o dilithiauth-server ./cmd/server`: **PASS** (Executable binary generated successfully, size ~24MB).

---

## Detailed Edge-Case Stress Testing Findings

### Challenge 1: Provider Error Callback Propagation (`?error=access_denied`) [HIGH RISK]
- **Location**: `pkg/oauth/authz_handlers.go`, lines 1341–1351 (`SocialCallbackGET`)
- **Observed Behavior**:
  When a user denies access or cancels authorization on the social provider page (GitHub or Google), the provider redirects back to DilithiAuth callback with:
  `GET /auth/social/{provider}/callback?error=access_denied&error_description=User+denied&state=<signedState>`
  In `SocialCallbackGET`:
  ```go
  stateStr := r.URL.Query().Get("state")
  code := r.URL.Query().Get("code")

  if stateStr == "" || code == "" {
      http.Error(w, "missing state or code", http.StatusBadRequest)
      return
  }
  ```
  Since `code` is empty (`code == ""`), the handler immediately returns `400 Bad Request` with text `"missing state or code"`.
- **Attack / Failure Scenario**:
  1. End user clicks "Cancel" / "Deny" on Google or GitHub consent prompt.
  2. Google/GitHub redirects to DilithiAuth callback with `?error=access_denied&state=<signedState>`.
  3. DilithiAuth responds to the user with a plain text `400 Bad Request` error page.
  4. The client application's `redirect_uri` never receives the `error=access_denied` redirect. The client app remains in a pending state, and the end user is stranded on a 400 error page.
- **Recommended Defense**:
  Modify `SocialCallbackGET` to check for `r.URL.Query().Get("error")`. If present and `state` is valid, decode `stateClaims` to retrieve `ClientID`, `RedirectURI`, and `ClientState`, then issue a 302 redirect to `claims.RedirectURI?error=<error>&state=<clientState>`.

---

### Challenge 2: Username Collision During New User Provisioning [MEDIUM RISK]
- **Location**: `pkg/oauth/authz_handlers.go`, lines 1437–1444
- **Observed Behavior**:
  When a new social user logs in via GitHub or Google, their email (`socialUser.Email`) is checked against `GetUserByEmail`. If no account exists, a new `userSubject` (`usr_<16hex>`) is generated, and `username` is assigned from `socialUser.Name`.
  Then line 1438 executes:
  ```go
  _ = h.Store.RegisterUser(tenantID, &User{
      Username: username,
      Email:    socialUser.Email,
      Password: randomPassword,
      Subject:  userSubject,
  })
  ```
  If a local user already exists with `Username == username` but a different email address, `RegisterUser` returns `error("username already taken")`.
  However, `SocialCallbackGET` **ignores** the error (`_ = ...`) and proceeds to call `finishAuthorization`, issuing an authorization code for `userSubject`.
- **Failure Scenario**:
  1. Local user "John Doe" registers with email `john.local@example.com`.
  2. New user "John Doe" signs in via GitHub with email `john.github@example.com`.
  3. `GetUserByEmail` returns not found. `userSubject` `usr_a1b2c3d4` is generated.
  4. `RegisterUser` fails because username `"John Doe"` is already taken. The error is silently discarded.
  5. Authorization code and JWT are issued for subject `usr_a1b2c3d4`.
  6. Subsequent lookups in `UserStore` or `MemoryAuthzStore` for subject `usr_a1b2c3d4` fail with `user not found`.
- **Recommended Defense**:
  If `username` is already taken, append a unique suffix (e.g. `username + "_" + hexSuffix`) or fall back to `socialUser.Email` or `userSubject` before calling `RegisterUser`. Handle any non-nil registration error properly.

---

### Challenge 3: GitHub Private Email Resolution Fallback (`GET /user/emails`) [LOW RISK / VERIFIED PASS]
- **Location**: `pkg/oauth/social.go`, lines 248–285
- **Verified Behavior**:
  - When `/user` returns `email: null` or `""`, `ExchangeSocialCode` queries `/user/emails`.
  - Accurately selects `primary: true` email if present.
  - If no email is marked `primary: true`, safely selects the first available email in the list.
  - If `/user/emails` fails or returns 400/403/500, `SocialUser.Email` remains `""`, and `SocialCallbackGET` returns 400 Bad Request ("failed to obtain user email from social provider").
- **Minor Caveat**: If `userinfoURL` contains URL query parameters (e.g. `http://custom/user?param=1`), appending `/emails` naively creates a malformed URL.

---

### Challenge 4: GitHub Login Handle Fallback When Full Name Is Empty [VERIFIED PASS]
- **Location**: `pkg/oauth/social.go`, lines 239–246
- **Verified Behavior**:
  - When `name` is null or empty in GitHub profile, `ExchangeSocialCode` extracts `login` handle.
  - If both `name` and `login` are empty/null, `socialUser.Name` remains `""`, and `SocialCallbackGET` falls back to `socialUser.Email` as the username.

---

### Challenge 5: PKCE Challenge Parameter Propagation [VERIFIED PASS]
- **Location**: `pkg/oauth/authz_handlers.go`, lines 1317–1325, 1451–1462; `pkg/oauth/authz.go`, lines 341–352
- **Verified Behavior**:
  - `code_challenge` and `code_challenge_method` (defaulting to `S256`) are saved in `socialStateClaims` and signed into the state JWT.
  - `SocialCallbackGET` propagates `CodeChallenge` and `CodeChallengeMethod` into `CreateCode`.
  - Token exchange (`POST /token`) strictly validates the `code_verifier` against the code's challenge.
  - Token exchange with incorrect `code_verifier` returns `400 Bad Request` (`pkce: code_verifier mismatch`).
  - Public clients attempting to exchange codes without PKCE are rejected at `/token` with `400 Bad Request` (`PKCE is required for public clients`).

---

## Empirical Test Suite
Empirical verification was performed via `pkg/oauth/social_challenge_test.go`, executing 5 test routines (`TestEmpirical_1_` through `TestEmpirical_5_`).
- Command: `go test -v ./pkg/oauth -run TestEmpirical_`
- Result: 5/5 tests executed, confirming verified passes and reproducing both uncovered edge-case flaws.
