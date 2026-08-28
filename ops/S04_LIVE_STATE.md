# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after real Owner localhost authentication/session lifecycle PASSED

> Authoritative fast continuation file. Read `CONTINUE_HERE.md`, newest S04 evidence, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` before acting on the live host.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, not official PASSED yet because public Caddy exposure and independent external postflight have not completed.
- S04 API/auth/web localhost deployment is installed and healthy on `testAmir5-3`.
- Final corrected independent read-only localhost postflight PASSED at `2026-08-28T00:43:44Z`.
- Real one-time Owner bootstrap PASSED at `2026-08-28T00:50:33Z`.
- Real Owner localhost login/session/me/CSRF logout/revocation test PASSED at `2026-08-28T00:55:44Z`.
- Exactly one active Owner exists with a password hash; real Owner email/password are intentionally omitted from this public repository.
- Next permitted action: **read-only inspection of the live Caddy configuration/service contract**.
- Do not mutate Caddy until the exact current NaiveProxy/forward-proxy route structure has been captured and reviewed.

## Newest live evidence

`ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md`

Real auth output included:

```text
LOGIN_HTTP=200
LOGIN_AUTHENTICATED=PASSED
ME_HTTP=200
OWNER_ME=PASSED
SUCCESS_LOGIN_AUDIT_AFTER=1
SESSION_ACTIVE_CONTRACT=32|0
LOGIN_AUDIT=PASSED
SESSION_PERSISTENCE=PASSED
LOGOUT_HTTP=200
CSRF_LOGOUT=PASSED
OLD_SESSION_ME_HTTP=401
SESSION_REVOKED_CONTRACT=32|1
SESSION_REVOCATION=PASSED
REAL_OWNER_LOCALHOST_AUTH=PASSED
NEXT=PREPARE_CADDY_PANEL_EXPOSURE
```

## Auth security properties independently observed

- Before the live test, the Owner had zero successful login audits and zero auth sessions.
- Real login returned HTTP `200`.
- Session cookie and CSRF cookie were both issued.
- `/api/v1/me` returned HTTP `200` and Owner role.
- Successful `auth.login` audit count increased by exactly one.
- A single new auth session was persisted with a 32-byte token hash and was initially not revoked (`32|0`).
- Logout with `X-CSRF-Token` returned HTTP `200`.
- Reusing the old session against `/api/v1/me` returned HTTP `401`.
- The same session row then had non-null `revoked_at` while retaining the 32-byte hash (`32|1`).
- No raw session token or password was printed into repository evidence.

## Live facts currently verified

- PostgreSQL 18 schema: `2`.
- Migration 0002 SHA-256: `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB role contract passed.
- `/etc/pvnaive/db.env` expects schema `2`.
- S04 marker: `/opt/pvnaive/S04_AUTH.json`.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- DB current release: `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Old S03 immutable DB release preserved: `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- Periodic DB health returns schema2, `pvnaive_app`, signing-secret denial and MFA-secret-table denial.
- `pvnaive-api.service` enabled/active as `pvnaive:pvnaive`, listener only `127.0.0.1:8080`.
- API liveness after real auth: `{"service":"pvnaive-api","status":"ok"}`.
- API readiness after real auth: `{"ready":true,"status":"ready"}`.
- Encrypted rollback backup `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age` previously validated.
- PostgreSQL listeners loopback-only.
- Caddy active; Caddyfile SHA still exactly `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH active as `ssh.service`.
- Real Owner auth test reported `CADDY_CHANGED=false`, `SSH_CHANGED=false`, `FIREWALL_CHANGED=false`.

## Repository/bootstrap fix provenance

- Owner bootstrap permission fix commit: `1da5a2e3c779a3773755c3aafbc337ad0393ce79`.
- Fixed bootstrap Git blob SHA: `bcfeee6b04e41fb8f8c98340847ebff261375b39`.
- Intended RED CI: `33130812610`.
- Follow-up fix CI: `33130860909` — Go, Web, PostgreSQL18/database, end-to-end auth rehearsal and production bundle all SUCCESS.
- Installed API/auth/web artifact remains from source commit `11c54dc1faae99a1491c750b30db9faa44a0c3ae`.
- Never reuse old `b4803e27...` bundle.

## Exact next action

Perform **one read-only Caddy inspection** before any public exposure change. Capture:

1. `/etc/caddy/Caddyfile` verbatim with line numbers, SHA-256, owner/group/mode;
2. `systemctl cat caddy-naive.service` and `systemctl show` fields `ExecStart`, `ExecReload`, `User`, `Group`, `FragmentPath`, `DropInPaths`;
3. `caddy validate --config /etc/caddy/Caddyfile` output and adapted JSON structure;
4. listeners/process identity for ports 22, 80, 443 and 8080;
5. current TLS/storage paths visible from the config/unit;
6. `/opt/pvnaive/web/current` symlink target and top-level static files;
7. API loopback-only + live/ready still healthy.

Do not infer the Caddy route insertion point before reading the exact live config. Once reviewed, implement the S04 Caddy finalizer with exact pre-change backup/SHA, candidate generation, validation before install, controlled `systemctl reload caddy-naive.service` only, and exact rollback+reload on any failure. NaiveProxy must remain functional.

Only after controlled exposure plus independent external postflight may `S04-AUTH` be marked PASSED and `S05-USERS` become NEXT.
