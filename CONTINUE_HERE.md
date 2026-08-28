# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T0046-owner-bootstrap-temp-sql-permission-fix.md` — newest live issue, root cause, RED/GREEN fix and exact continuation rule.
2. `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md` — final independent localhost S04 postflight PASSED.
3. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
4. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
5. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
6. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. Final independent localhost postflight PASSED. The first real Owner bootstrap then failed before INSERT because the root-owned temporary SQL file was mode `0600` while `psql --file` runs as OS user `postgres`. No Owner was created by that failed attempt.

The bootstrap bug is fixed in branch commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79`. Regression test commit `11bf2add44200dda48c63865263ffa05624b15f9` produced the intended RED CI run `33130812610` with `ERROR: bootstrap temp SQL is not handed to postgres before psql --file`. The fix hands the fully-written temp file to `postgres:postgres` and retains mode `0600`; the PostgreSQL18/database gate on the fix is GREEN.

## Live server state before retrying Owner bootstrap

- PostgreSQL schema `2`; migration 0002 exact checksum remains `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB health release is `0002-84bb735877d5`; timer is S04-aware and healthy.
- `pvnaive-api.service` is active only on `127.0.0.1:8080`.
- API live/ready passed.
- encrypted rollback backup validated.
- Caddy SHA remains `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; Caddy/SSH/firewall unchanged.
- final independent localhost postflight returned `S04_POSTFLIGHT=PASSED` and `NEXT=BOOTSTRAP_REAL_OWNER`.
- the failed bootstrap did not create an Owner.

## Exact next action

Do NOT modify the installed immutable auth release in place. Fetch `scripts/auth/bootstrap-owner.sh` pinned to fix commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79`, verify Git blob SHA exactly `bcfeee6b04e41fb8f8c98340847ebff261375b39`, execute that fixed copy from a root-only temporary path, and enter the real Owner email/display name/password via `/dev/tty`.

After `PVNAIVE_OWNER_BOOTSTRAP=PASSED`:

1. Independently verify exactly one active Owner exists and a password hash is present without printing the hash.
2. Test real localhost login/session `/me`/CSRF logout/revocation with the Owner.
3. Only then prepare Caddy exposure for `/panel/` + `/api/` with backup, validate, controlled reload and rollback.
4. Only after external postflight may the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

Never reuse the old `b4803e27...` bundle.
