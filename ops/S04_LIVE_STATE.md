# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after real Owner bootstrap PASSED

> Authoritative fast continuation file. Read `CONTINUE_HERE.md`, newest S04 evidence, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` before acting on the live host.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, not official PASSED yet because real localhost auth verification plus public Caddy exposure/external postflight have not completed.
- S04 API/auth/web localhost deployment is installed and healthy on `testAmir5-3`.
- Final corrected independent read-only localhost postflight PASSED at `2026-08-28T00:43:44Z`.
- Real one-time Owner bootstrap PASSED at `2026-08-28T00:50:33Z` using the fixed, commit-pinned bootstrap script.
- Exactly one active Owner exists and a password hash is present; the hash was not disclosed.
- The Owner email is intentionally omitted from this public repository.
- Next permitted action: real Owner localhost login/session/me/CSRF logout/revocation test.
- Do not expose `/panel/` or `/api/` through Caddy until that real localhost auth test passes.

## Real Owner bootstrap evidence

Newest evidence:
`ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md`

Live success output included:

```text
PVNAIVE_OWNER_BOOTSTRAP=PASSED
ACTIVE_OWNER_WITH_PASSWORD_COUNT=1
TOTAL_OWNER_COUNT=1
API_LISTENER=127.0.0.1:8080
OWNER_BOOTSTRAP_FINAL=PASSED
ACTIVE_OWNER_COUNT=1
PASSWORD_HASH_PRESENT=true
PASSWORD_HASH_DISCLOSED=false
NEXT=REAL_OWNER_LOCALHOST_LOGIN_TEST
```

Before execution the launcher verified schema2, Owner count zero and API active state. It downloaded the fixed bootstrap script from:

`1da5a2e3c779a3773755c3aafbc337ad0393ce79`

and independently verified Git blob SHA:

`bcfeee6b04e41fb8f8c98340847ebff261375b39`.

The previously failed bootstrap attempt created no Owner. Its root cause was a root-owned mode-0600 temporary SQL file being unreadable by OS user `postgres`. Regression test commit `11bf2add44200dda48c63865263ffa05624b15f9` produced intended RED CI run `33130812610`; production fix commit `1da5a2e3...` then passed follow-up CI run `33130860909` across all five jobs: Go, Web, PostgreSQL18/database, end-to-end auth rehearsal and production bundle.

## Live facts currently expected to remain true

- PostgreSQL 18 schema: `2`.
- Migration 0002 SHA-256: `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB role contract passed.
- `/etc/pvnaive/db.env` expects schema `2`.
- S04 marker: `/opt/pvnaive/S04_AUTH.json`.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- DB current release: `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Old S03 immutable DB release preserved: `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- Periodic DB health returns schema2, `pvnaive_app`, signing secret denial and MFA secret-table denial.
- `pvnaive-api.service` enabled/active as `pvnaive:pvnaive`, listener only `127.0.0.1:8080`.
- API live/ready passed.
- Encrypted rollback backup `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age` validated.
- PostgreSQL listeners loopback-only.
- Caddy active; Caddyfile SHA unchanged: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; validation passed.
- SSH active.

## Exact next action

Run a real Owner localhost authentication test against `http://127.0.0.1:8080`. Prompt for password using `/dev/tty` with echo disabled and do not place the password in argv, environment, command history or repo.

The test must:

1. identify the single active Owner internally without printing sensitive credential material;
2. snapshot current successful `auth.login` audit count for that Owner;
3. POST `/api/v1/auth/login` and require authenticated Owner response;
4. retain session + CSRF cookies;
5. GET `/api/v1/me` and require Owner role;
6. verify exactly one additional successful `auth.login` audit event;
7. POST `/api/v1/auth/logout` with `X-CSRF-Token` from the CSRF cookie;
8. require old session `/api/v1/me` to return HTTP 401;
9. verify the newly-created session token hash length is 32 bytes and `revoked_at` is non-null;
10. verify API remains loopback-only and Caddy/SSH/firewall invariants remain unchanged.

Only after this real localhost test passes may Caddy exposure be prepared. Only after controlled Caddy reload plus external postflight may `S04-AUTH` be marked PASSED and `S05-USERS` become NEXT.