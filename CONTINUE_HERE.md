# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T005544Z-real-owner-localhost-auth-pass.md` — newest live evidence; real Owner localhost auth/session/logout/revocation PASSED.
2. `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md` — real Owner bootstrap PASSED.
3. `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md` — final independent localhost S04 postflight PASSED.
4. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
5. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
6. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
7. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. Final independent localhost postflight PASSED, the real one-time Owner bootstrap PASSED, and the real Owner localhost login/session/me/CSRF logout/revocation test also PASSED. The API remains healthy and loopback-only on `127.0.0.1:8080`; Caddy/SSH/firewall remained unchanged.

The next permitted action is **read-only inspection of the exact live Caddy configuration and service contract** before designing the panel exposure. Do not mutate Caddy yet and do not advance the official ledger to S04 PASSED yet.

The real Owner email/password are intentionally not copied into this public repository.

## Real localhost auth PASS

At `2026-08-28T00:55:44Z` the real Owner test returned:

```text
LOGIN_HTTP=200
ME_HTTP=200
SUCCESS_LOGIN_AUDIT_AFTER=1
SESSION_ACTIVE_CONTRACT=32|0
LOGOUT_HTTP=200
OLD_SESSION_ME_HTTP=401
SESSION_REVOKED_CONTRACT=32|1
REAL_OWNER_LOCALHOST_AUTH=PASSED
CADDY_CHANGED=false
SSH_CHANGED=false
FIREWALL_CHANGED=false
NEXT=PREPARE_CADDY_PANEL_EXPOSURE
```

Meaning:

- real Owner credentials authenticate successfully;
- session and CSRF cookies are issued;
- `/api/v1/me` resolves the live Owner session;
- login audit increments exactly once;
- DB stores only a 32-byte session token hash;
- CSRF logout succeeds;
- old session is rejected with HTTP 401 and its DB row is revoked.

## Live S04 state that remains valid

- PostgreSQL schema `2`; migration 0002 checksum `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB health release `0002-84bb735877d5`; timer is S04-aware and healthy.
- exactly one active Owner exists with a password hash.
- `pvnaive-api.service` active as `pvnaive:pvnaive`; listener only `127.0.0.1:8080`.
- API live/ready passed after real Owner auth.
- encrypted rollback backup validated.
- PostgreSQL loopback-only.
- Caddy active and previously validated; Caddyfile SHA remains exactly `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH active; Caddy/SSH/firewall unchanged.
- Owner bootstrap permission regression is fixed in `s04-auth` commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79`; CI `33130860909` passed Go, Web, PostgreSQL18/database, end-to-end rehearsal and production bundle.

## Exact next action

Run one **read-only Caddy inspection** on `testAmir5-3`. It must capture without changing anything:

1. exact `/etc/caddy/Caddyfile` content with line numbers plus SHA/owner/mode;
2. `systemctl cat/show caddy-naive.service`, especially `ExecStart`, `ExecReload`, `User`, `Group`;
3. `caddy validate` and adapted JSON shape;
4. current TCP listeners for 22/80/443/8080 and Caddy process identity;
5. current certificate/storage/service paths if visible from the unit/config;
6. exact `/opt/pvnaive/web/current` target and top-level files;
7. confirmation API is still loopback-only and live/ready.

Only after that evidence is reviewed should the Caddy exposure finalizer be implemented/tested. It must use exact backup + candidate validation + controlled reload (never restart) + automatic exact rollback/reload on failure, while preserving NaiveProxy/forward-proxy behavior.

Only after external panel/API smoke and independent external postflight may the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

Never reuse the old `b4803e27...` bundle.
