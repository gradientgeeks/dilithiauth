# Production Deployment: OCI VM + Firebase Hosting

This documents how DilithiAuth is actually configured and running in production, as
verified directly against the live infrastructure. It's a live-state record, not a
generic how-to — see [`deployment.md`](./deployment.md) for the general
environment-variable reference and local dev quick start.

## Topology

Two independent deployment targets serve two different domains — they are **not**
the same host, and mixing them up (e.g. redeploying the frontend to the VM) will not
work:

```mermaid
graph LR
    U["Browser"]
    CF["Cloudflare<br/>(proxied)"]
    FB["Firebase Hosting<br/>project: dilithiauth"]
    N["nginx (OCI VM)<br/>TLS termination"]
    GO["DilithiAuth server<br/>127.0.0.1:8080"]
    PG[("PostgreSQL<br/>local, 127.0.0.1:5432")]
    RD[("Redis<br/>local, 127.0.0.1:6380<br/>ACL auth")]

    U -->|dilithiauth.gradientgeeks.tech<br/>CNAME to dilithiauth.web.app| FB
    U -->|dilithiauthapis.gradientgeeks.tech| CF
    CF -->|origin cert| N
    N -->|proxy_pass| GO
    GO --> PG
    GO --> RD
```

| Domain | Serves | Backing infra |
|---|---|---|
| `dilithiauth.gradientgeeks.tech` | Developer portal SPA (React/Vite) | **Firebase Hosting**, project ID `dilithiauth` (CNAME → `dilithiauth.web.app`) |
| `dilithiauthapis.gradientgeeks.tech` | OAuth/OIDC + developer REST API | **OCI VM** (`80.225.222.81`), fronted by Cloudflare (orange-cloud proxy) |

There is an nginx `server` block on the VM for `dilithiauth.gradientgeeks.tech` too
(see below), but it is **not receiving traffic** — DNS for that hostname resolves to
Firebase, not the VM. It's effectively dead config left over from an earlier
approach where the SPA was served directly from the Go binary's static-file handler;
harmless to leave, but don't rely on it.

## OCI VM

- Host: `80.225.222.81` (SSH as `ubuntu`)
- OS: Ubuntu 24.04.4 LTS ("noble"), ARM64 (Ampere)
- **Important:** this VM is shared with other, unrelated services (there was a
  separate "openwa" app stack running here before it was decommissioned in July
  2026 — always check `systemctl list-units`, `podman ps -a` / `docker ps -a`, and
  `ss -tlnp` before assuming a port is free or a process is safe to touch).

### Directory layout

```
/opt/dilithiauth/
├── server              # the deployed binary (linux/arm64, stripped)
└── keys/                # persistent hybrid signing keys (Ed25519 + ML-DSA-44)
    ├── active_kid.txt    # current active key id
    └── <kid>.json        # key material, mode 600
```

Owned by `ubuntu:ubuntu`. `keys/` must survive redeploys — deleting or regenerating
it invalidates every previously issued token, since JWKS/`kid` changes.

### systemd service

Unit: `/etc/systemd/system/dilithiauth.service`

```ini
[Unit]
Description=DilithiAuth Post-Quantum Identity Provider
After=network.target postgresql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/dilithiauth
ExecStart=/opt/dilithiauth/server
Restart=always
RestartSec=3

Environment="DILITHIAUTH_ISSUER=https://dilithiauthapis.gradientgeeks.tech"
Environment="DILITHIAUTH_ADDR=:8080"
Environment="DILITHIAUTH_KEYS_DIR=/opt/dilithiauth/keys"
Environment="DILITHIAUTH_POSTGRES_DSN=postgres://dilithiauth:***@127.0.0.1:5432/dilithiauth?sslmode=disable"
Environment="DILITHIAUTH_SIGSTORE=redis"
Environment="DILITHIAUTH_REDIS_ADDR=127.0.0.1:6380"
Environment="DILITHIAUTH_REDIS_USER=dilithiauth"
Environment="DILITHIAUTH_REDIS_PASS=***"
Environment="DILITHIAUTH_FRONTEND_URL=https://dilithiauth.gradientgeeks.tech"
Environment="DILITHIAUTH_GOOGLE_CLIENT_ID=your-google-client-id"
Environment="DILITHIAUTH_GOOGLE_CLIENT_SECRET=***"
Environment="DILITHIAUTH_GOOGLE_REDIRECT_URI=https://dilithiauthapis.gradientgeeks.tech/auth/social/google/callback"
Environment="DILITHIAUTH_GITHUB_CLIENT_ID=your-github-client-id"
Environment="DILITHIAUTH_GITHUB_CLIENT_SECRET=***"
Environment="DILITHIAUTH_GITHUB_REDIRECT_URI=https://dilithiauthapis.gradientgeeks.tech/auth/social/github/callback"
Environment="OCI_SMTP_HOST=smtp.email.ap-mumbai-1.oci.oraclecloud.com"
Environment="OCI_SMTP_PORT=587"
Environment="OCI_SMTP_USER=***"
Environment="OCI_SMTP_PASS=***"
Environment="OCI_SMTP_FROM=no-reply@gradientgeeks.tech"

[Install]
WantedBy=multi-user.target
```

**Real secret values live only in this file on the VM — never in git.** To view or
edit them: `sudo cat /etc/systemd/system/dilithiauth.service` on the VM (root only).

Notes on the config:
- `DILITHIAUTH_SIGSTORE=redis` — signature-sidecar storage is Redis-backed for
  shared/HA-ready lookups. `AuthzStore` and `KeyStore` are left on their defaults
  (in-memory / file respectively) — see [Redis](#redis) below for why only SigStore
  was moved.
- `DILITHIAUTH_POSTGRES_DSN` — used for the developer portal (developer accounts,
  OAuth apps) and end-user accounts (`users` table); durable across restarts.
- No `DILITHIAUTH_ADMIN_TOKEN` or `DILITHIAUTH_SESSION_SECRET` is set, so the
  process generates ephemeral ones on every restart (logged at `WARN` level on
  startup). This means the admin token changes on every deploy/restart, and any
  developer-portal sessions signed with the old session secret are invalidated.
  Setting these explicitly is a known gap — do this before treating the deployment
  as fully hardened.

Manage with the usual `sudo systemctl {status,restart,stop} dilithiauth.service`;
logs via `sudo journalctl -u dilithiauth.service -f`.

### Deployment procedure (backend)

Cross-compile locally, ship the binary, swap it in:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/server ./cmd/server
rsync -avz /tmp/server ubuntu@80.225.222.81:/opt/dilithiauth/server.new

ssh ubuntu@80.225.222.81 '
  sudo chmod +x /opt/dilithiauth/server.new
  sudo cp /opt/dilithiauth/server /opt/dilithiauth/server.bak
  sudo mv /opt/dilithiauth/server.new /opt/dilithiauth/server
  sudo systemctl restart dilithiauth.service
'
```

Verify via `curl https://dilithiauthapis.gradientgeeks.tech/healthz` (liveness) and
`/readyz` (checks every configured dependency — Postgres, Redis — is actually
reachable, not just that the process is up). Once confirmed stable, remove
`/opt/dilithiauth/server.bak`.

There is no CI/CD for the backend — deploys are manual, run from a developer
machine with SSH access to the VM.

### Deployment procedure (frontend)

Automatic: any push to `main` that changes files under `frontend/` triggers
`.github/workflows/firebase-hosting-merge.yml`, which builds
(`cd frontend && npm install && npm run build`) and deploys via
`FirebaseExtended/action-hosting-deploy`, `projectId: dilithiauth`, `channelId: live`.
PRs get their own preview channel via `firebase-hosting-pull-request.yml`. The
Firebase service account is stored as the `FIREBASE_SERVICE_ACCOUNT_DILITHIAUTH`
GitHub Actions secret.

To deploy manually instead: `cd frontend && npm run build && npx firebase deploy --only hosting --project dilithiauth`.

Cloudflare sits in front of the Firebase-served domain too (as CDN/cache), so a
fresh deploy can take a little while to become visible everywhere — check
`Cache-Control`/`ETag` on `index.html` if a change doesn't appear immediately.

## nginx (VM)

Config: `/etc/nginx/sites-available/dilithiauth` (symlinked into `sites-enabled/`).

- Handles both `dilithiauthapis.gradientgeeks.tech` (proxies to the Go backend —
  the block that actually matters) and `dilithiauth.gradientgeeks.tech` (dead,
  see [Topology](#topology)).
- HTTP → HTTPS redirect for both hostnames.
- `client_header_buffer_size 8k; large_client_header_buffers 4 16k;` and
  `proxy_buffer_size 16k; proxy_buffers 4 16k;` — required because full/jumbo PQC
  JWTs can reach 3–4 KB, exceeding nginx's 4 KB defaults. Mirrors the Go server's
  own `MaxHeaderBytes = 16 KiB` (`transport.NewPQCServer`).
- TLS cert/key: `/etc/ssl/dilithiauth/origin.crt` + `origin.key` — a Cloudflare
  Origin CA certificate (valid because Cloudflare is the only client expected to
  reach nginx directly; browsers only ever see Cloudflare's edge cert).
- `keepalive 32` to the upstream (`127.0.0.1:8080`).

## PostgreSQL

Local instance on the VM, `127.0.0.1:5432`, database `dilithiauth`, role
`dilithiauth` (password only in the systemd unit, not here). Tables in use:

- `developer_accounts` — developer-portal login accounts
- `oauth_apps` — registered third-party OAuth clients (`client_id`, hashed secret,
  `redirect_uris` as a Postgres `TEXT[]` — must be read/written via `pq.Array()`,
  see the note in `pkg/developer/store.go`)
- `users` — end-user accounts (the people who log in *through* an OAuth app, not
  developers)

## Redis

Runs as the apt-installed `redis-server` (v7.0.15) via systemd, **not** a container
— installed directly on the VM in July 2026 specifically to back `SigStore`.

- Bound to `127.0.0.1` only (`bind 127.0.0.1 -::1`), `protected-mode yes`.
- **Port 6380**, not the default 6379 — the VM already had an unrelated app's
  containerized Redis instance bound to 6379 (part of the "openwa" stack, since
  removed); DilithiAuth's Redis was deliberately placed on a different port rather
  than fighting over the default one.
- **ACL auth, not `requirepass`**: the anonymous `default` user is disabled
  (`user default off`); a dedicated `dilithiauth` user exists, restricted to keys
  under `~sigstore:*` and to `@read @write @connection @keyspace` command
  categories (`-@dangerous` — no `CONFIG`, `FLUSHALL`, `SHUTDOWN`, `ACL`, etc.).
  Config lives at the end of `/etc/redis/redis.conf` on the VM.
- Only `SigStore` uses Redis in production (`DILITHIAUTH_SIGSTORE=redis`).
  `AuthzStore` and `KeyStore` both have Redis-backed implementations available in
  the codebase (`pkg/oauth/redis_authz_store.go`, `pkg/signer/redis_keystore.go`)
  but are intentionally left on their in-memory/file defaults for now — moving them
  is a deliberate follow-up, not an oversight.

Verify connectivity: `redis-cli -p 6380 --user dilithiauth --pass <password> PING`
(anonymous `redis-cli -p 6380 PING` should fail with `NOAUTH`).

## Health & readiness

- `GET /healthz` — liveness only; always `200` once the process is up. Not for
  routing decisions.
- `GET /readyz` — actually pings every configured backend (Redis SigStore,
  Postgres stores) and returns `503` if any is unreachable. Use this for load
  balancer / orchestrator readiness checks.

## Background maintenance

An hourly goroutine in `cmd/server/main.go` purges expired entries from
`revStore`, `sessionRevStore`, `sigStore`, and `refreshTokenStore` to bound memory
growth — relevant mainly for the in-memory backends still in use (`AuthzStore`,
revocation stores).

## Known gaps / follow-ups

- `DILITHIAUTH_ADMIN_TOKEN` and `DILITHIAUTH_SESSION_SECRET` are unset, so both are
  regenerated ephemerally on every restart — set them explicitly for a stable admin
  token and to stop invalidating developer sessions on every deploy.
- Backend deploys are fully manual (no CI/CD), unlike the frontend.
- `AuthzStore`/`KeyStore` are single-node in-memory/file only — fine for the
  current single-VM deployment, but would need the existing Redis-backed
  implementations wired in before running multiple backend instances.
