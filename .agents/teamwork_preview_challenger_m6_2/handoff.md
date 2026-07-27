# Handoff Report — Social Authentication Stress Testing

## 1. Observation

### Build & Core Suite Execution
- Command executed: `go test ./... && go build -o dilithiauth-server ./cmd/server`
- Output:
  ```
  ok  	github.com/Uttam-Mahata/DilithiAuth/cmd/server	(cached)
  ok  	github.com/Uttam-Mahata/DilithiAuth/pkg/bff	(cached)
  ok  	github.com/Uttam-Mahata/DilithiAuth/pkg/developer	(cached)
  ok  	github.com/Uttam-Mahata/DilithiAuth/pkg/oauth	3.971s
  ok  	github.com/Uttam-Mahata/DilithiAuth/pkg/signer	(cached)
  ok  	github.com/Uttam-Mahata/DilithiAuth/pkg/transport	(cached)
  ```
  Binary produced: `/home/uttam/projects/DilithiAuth/dilithiauth-server` (Size: 24,082,509 bytes).

### Social Auth Source Code Observations
1. **GitHub Private Email Resolution**:
   - Source: `pkg/oauth/social.go`, lines 248–285
   - Quotes:
     ```go
     if emailVal == "" {
         emailsURL := userinfoURL + "/emails"
         ...
         for _, em := range emailsList {
             eStr := fmt.Sprintf("%v", em["email"])
             if eStr != "" && eStr != "<nil>" {
                 primary, _ := em["primary"].(bool)
                 if primary {
                     emailVal = eStr
                     break
                 }
                 if emailVal == "" {
                     emailVal = eStr
                 }
             }
         }
     }
     ```
2. **GitHub Login Handle Fallback**:
   - Source: `pkg/oauth/social.go`, lines 239–246
   - Quotes:
     ```go
     if nameVal == "" {
         if login, ok := ghUser["login"]; ok && login != nil {
             nameVal = fmt.Sprintf("%v", login)
             if nameVal == "<nil>" {
                 nameVal = ""
             }
         }
     }
     ```
3. **User Provisioning & Store Error Suppression**:
   - Source: `pkg/oauth/authz_handlers.go`, lines 1438–1444
   - Quotes:
     ```go
     _ = h.Store.RegisterUser(tenantID, &User{
         Username: username,
         Email:    socialUser.Email,
         Password: randomPassword,
         Subject:  userSubject,
     })
     ```
4. **Provider Error Callback Rejection**:
   - Source: `pkg/oauth/authz_handlers.go`, lines 1345–1351
   - Quotes:
     ```go
     stateStr := r.URL.Query().Get("state")
     code := r.URL.Query().Get("code")

     if stateStr == "" || code == "" {
         http.Error(w, "missing state or code", http.StatusBadRequest)
         return
     }
     ```

### Empirical Test Execution
- Command executed: `go test -v ./pkg/oauth -run TestEmpirical_`
- Output:
  ```
  === RUN   TestEmpirical_1_GitHubPrivateEmailFallback
  --- PASS: TestEmpirical_1_GitHubPrivateEmailFallback (0.00s)
  === RUN   TestEmpirical_2_GitHubLoginHandleFallback
  --- PASS: TestEmpirical_2_GitHubLoginHandleFallback (0.00s)
  === RUN   TestEmpirical_3_UserProvisioningAndUsernameCollision
      social_challenge_test.go:652: EDGE CASE UNCOVERED: User with subject usr_3525d36c0764bdb2 was NOT stored in AuthzStore because RegisterUser failed due to username collision (username 'CollidingUsername' already taken by another user). Lookup error: user not found
  --- PASS: TestEmpirical_3_UserProvisioningAndUsernameCollision (0.14s)
  === RUN   TestEmpirical_4_PKCEPropagationAndVerification
  --- PASS: TestEmpirical_4_PKCEPropagationAndVerification (0.05s)
  === RUN   TestEmpirical_5_ProviderErrorCallbackPropagation
      social_challenge_test.go:773: Empirical result for provider error callback: status=400, body=missing state or code
      social_challenge_test.go:775: EDGE CASE UNCOVERED: Provider error callback ?error=access_denied results in 400 Bad Request ('missing state or code') instead of propagating 302 redirect with ?error=access_denied to client redirect_uri
  --- PASS: TestEmpirical_5_ProviderErrorCallbackPropagation (0.00s)
  ```

---

## 2. Logic Chain

1. **Observation 1 & 2** demonstrate that GitHub private email resolution (`GET /user/emails`) and login handle fallback are implemented as intended and function correctly under normal and edge conditions.
2. **Observation 3** shows that when `RegisterUser` returns an error (e.g., due to username collision with an existing user), `SocialCallbackGET` ignores the error (`_ = ...`). The code issuance flow continues anyway, minting an auth code for a `userSubject` that was never saved to the store.
3. **Observation 4** shows that `SocialCallbackGET` mandates both `state` and `code` query parameters. When a social provider returns an error callback (such as `?error=access_denied`), `code` is omitted by the provider. The handler immediately short-circuits with a `400 Bad Request` ("missing state or code"), preventing the error status from reaching the OAuth client's `redirect_uri`.
4. Therefore, while core happy-path flows and PKCE propagation work properly, edge cases involving provider denial and username collisions fail to handle errors correctly.

---

## 3. Caveats

- Testing of live Google and GitHub endpoints relied on local mock `httptest.Server` fixtures rather than live external network calls (governed by `CODE_ONLY` network restriction).
- Real production PostgreSQL database integration (`PostgresUserStore`) was evaluated structurally; empirical unit tests ran against `MemoryAuthzStore`.

---

## 4. Conclusion

- **Verdict**: **FAIL**
- Summary: Core Social Auth functionality (GitHub email resolution, login fallback, state signing, PKCE propagation) builds and passes tests. However, two significant edge-case bugs were empirically identified:
  1. Provider error callbacks (`?error=access_denied`) are blocked with HTTP 400 Bad Request instead of being redirected to client applications.
  2. Username collisions during new user registration silently discard user store creation while still issuing authorization codes.

---

## 5. Verification Method

To verify these results independently:

1. Run standard build and unit test commands:
   ```bash
   cd /home/uttam/projects/DilithiAuth
   go test ./...
   go build -o dilithiauth-server ./cmd/server
   ```
2. Run empirical social authentication challenge tests:
   ```bash
   go test -v ./pkg/oauth -run TestEmpirical_
   ```
3. Inspect `pkg/oauth/social_challenge_test.go` and `pkg/oauth/authz_handlers.go` lines 1345–1351 & 1438–1444.
