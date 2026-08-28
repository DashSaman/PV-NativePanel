# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T005033Z-owner-bootstrap-pass.md` — newest live evidence; real Owner bootstrap PASSED.
2. `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md` — final independent localhost S04 postflight PASSED.
3. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
4. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
5. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
6. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. Final independent localhost postflight PASSED and the real one-time Owner bootstrap has now also PASSED. Exactly one active Owner exists with a password hash; the hash was not disclosed. The API remains active only on `127.0.0.1:8080`.

The next permitted action is a **real Owner localhost login/session/logout/revocation test**. Do not expose the panel through Caddy yet and do not advance the official ledger to S04 PASSED yet.

The real Owner email is intentionally not copied into this public repository.

## Owner bootstrap evidence

At `2026-08-28T00:50:33Z` the fixed bootstrap launcher verified schema2, Owner count zero and API health, then fetched the fixed bootstrap script from commit:

`1da5a2e3c779a3773755c3aafbc337ad0393ce79`

Git blob SHA matched exactly:

`bcfeee6b04e41fb8f8c98340847ebff261375b39`

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

The failed earlier bootstrap attempt created no Owner. Its permission bug is regression-covered. RED run `33130812610` caught the missing postgres-readable handoff; production fix commit `1da5a2e3...` then passed follow-up CI run `33130860909` across Go, Web, PostgreSQL18/database, end-to-end auth rehearsal and production bundle.

## Live S04 state that remains valid

- PostgreSQL schema `2`; migration 0002 checksum `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB health release `0002-84bb735877d5`; timer is S04-aware and healthy.
- `pvnaive-api.service` active as `pvnaive:pvnaive`; listener only `127.0.0.1:8080`.
- API live/ready passed.
- encrypted rollback backup validated.
- PostgreSQL loopback-only.
- Caddy active and validated; Caddyfile SHA remains `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH active; Caddy/SSH/firewall unchanged.

## Exact next action

Run one real localhost Owner authentication test against `http://127.0.0.1:8080` using the actual Owner password entered interactively on `/dev/tty`:

1. snapshot the current successful `auth.login` audit count for the Owner;
2. POST `/api/v1/auth/login`; require `status=authenticated` and `role=owner`;
3. use returned cookies to GET `/api/v1/me`; require `role=owner`;
4. verify exactly one new successful `auth.login` audit event;
5. obtain `__Host-pvnaive_csrf` from the cookie jar;
6. POST `/api/v1/auth/logout` with `X-CSRF-Token`;
7. require old cookie `/api/v1/me` to return HTTP `401`;
8. verify the newly-created session has a 32-byte token hash and a non-null `revoked_at`;
9. verify API listener/Caddy/SSH invariants remain unchanged.

Only after this real localhost auth test passes may Caddy exposure for `/panel/` + `/api/` be prepared with exact backup, `caddy validate`, controlled reload (never restart) and rollback. Only after external postflight may the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

Never reuse the old `b4803e27...` bundle.