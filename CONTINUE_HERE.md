# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md` — newest live evidence; corrected final independent localhost postflight PASSED.
2. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
3. `ops/evidence/S04-20260828T003759Z-db-health-release-promotion-pass.md` — live DB health release promotion evidence.
4. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
5. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
6. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. The S04 API/auth/web localhost deployment is installed and independently verified healthy on `testAmir5-3`. The final corrected read-only postflight at `2026-08-28T00:43:44Z` returned `S04_POSTFLIGHT=PASSED`, `DB_TIMER_S04_AWARE=true`, and `NEXT=BOOTSTRAP_REAL_OWNER`.

The next permitted action is the one-time real Owner bootstrap. Do **not** expose the panel through Caddy yet and do **not** advance the official ledger to S04 PASSED yet.

## Final localhost postflight facts

- PostgreSQL schema `2`.
- Migration 0002 exact checksum `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- Restricted DB roles passed with deterministic 0/1 encoding.
- `/opt/pvnaive/db/current` = `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Real periodic DB health verifies schema2, `pvnaive_app`, signing secret denial and MFA secret-table denial.
- `pvnaive-api.service` active as `pvnaive:pvnaive`, `NRestarts=0`.
- API listener only `127.0.0.1:8080`.
- API live/ready healthy.
- Encrypted rollback backup checksum + decrypt/archive parse passed.
- PostgreSQL loopback-only.
- Caddy active and validated; Caddyfile SHA unchanged at `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH active.
- Postflight was read-only and reported `NO_CONFIGURATION_CHANGES_MADE=true`.

## Exact next action

Run the installed one-time Owner bootstrap interactively from a root TTY:

`/opt/pvnaive/auth/current/scripts/auth/bootstrap-owner.sh`

The script must require schema version 2, refuse if any Owner already exists, prompt for Owner email/display name/password/confirmation via `/dev/tty`, hash the password with `/opt/pvnaive/bin/pvnaive-password`, and finally print `PVNAIVE_OWNER_BOOTSTRAP=PASSED` plus the normalized Owner email.

After successful Owner creation:

1. Independently verify exactly one active Owner exists and password hash is present; do not print the password hash.
2. Exercise real localhost login/session `/me`/CSRF logout/revocation with that Owner.
3. Only after localhost auth passes, prepare Caddy exposure for `/panel/` + `/api/` with exact backup, `caddy validate`, controlled reload (never restart), rollback, and external smoke.
4. Only after external postflight should the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.

Never reuse the old `b4803e27...` bundle.