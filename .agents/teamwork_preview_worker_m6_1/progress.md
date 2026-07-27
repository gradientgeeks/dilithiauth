# Progress Log

Last visited: 2026-07-27T18:17:25Z

- Initialized BRIEFING.md and ORIGINAL_REQUEST.md.
- Read Explorer handoff reports (m6_1, m6_2, m6_3).
- Updated `pkg/oauth/social.go`: added `TokenURL`/`UserInfoURL` to `SocialConfig`, updated `ExchangeSocialCode` with GitHub private email fallback (`/user/emails`).
- Updated `pkg/oauth/authz.go`: preserved `Email` and `Subject` in `MemoryAuthzStore`.
- Updated `pkg/oauth/authz_handlers.go`: added `GoogleConfig` & `GitHubConfig` to `AuthzHandlers`, implemented `SocialLoginGET` and `SocialCallbackGET` with state JWT generation/validation, user lookup & provisioning.
- Updated `cmd/server/main.go`: parsed `DILITHIAUTH_GOOGLE_*` and `DILITHIAUTH_GITHUB_*` environment variables, injected configs into `AuthzHandlers`, registered routes under tenant router group.
- Created `deployments/dilithiauth.service` systemd unit template.
- Updated `docs/oci-vm-deployment.md` and `docs/deployment.md` documentation.
- Created unit tests in `pkg/oauth/social_test.go`.
- Ran `go test -v -count=1 ./pkg/oauth` (PASS - 64/64 tests passed).
- Ran `go build -o dilithiauth-server ./cmd/server` (PASS - binary created).
- Written `changes.md` and `handoff.md`.
- Completed task.
