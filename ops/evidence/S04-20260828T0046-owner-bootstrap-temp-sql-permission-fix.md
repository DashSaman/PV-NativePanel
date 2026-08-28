# S04 Owner bootstrap — temp SQL permission failure and fix

Timestamp: 2026-08-28 UTC
Host: `testAmir5-3`

## Prerequisite status

The corrected independent localhost S04 postflight passed completely at `2026-08-28T00:43:44Z` and returned:

```text
S04_POSTFLIGHT_CORE=PASSED
DATABASE_ROLE_CONTRACT=PASSED
DB_TIMER_S04_AWARE=true
API_POSTFLIGHT=PASSED
ROLLBACK_BACKUP=PASSED
INFRASTRUCTURE_POSTFLIGHT=PASSED
S04_POSTFLIGHT=PASSED
NO_CONFIGURATION_CHANGES_MADE=true
NEXT=BOOTSTRAP_REAL_OWNER
```

The live system therefore remains schema 2 with the S04-aware immutable DB health release selected, API active only on `127.0.0.1:8080`, encrypted rollback backup valid, and Caddy/SSH unchanged.

## Owner bootstrap live failure

The first real Owner bootstrap invocation prompted successfully for email, display name and password, then failed with:

```text
psql: error: /run/pvnaive-owner-bootstrap.<random>.sql: Permission denied
```

No Owner was created by this failed invocation.

## Root cause

`scripts/auth/bootstrap-owner.sh` creates the temporary SQL file as root and sets mode `0600`, but then invokes `psql --file` through `runuser -u postgres`. A root-owned `0600` file cannot be read by the `postgres` OS account.

The SQL temp file contains the Argon2id PHC hash and bootstrap metadata, not the raw password. The existing EXIT cleanup removes the temp file on failure.

## TDD evidence

1. Regression test strengthened in commit `11bf2add44200dda48c63865263ffa05624b15f9` to require an explicit ownership handoff to `postgres` while retaining mode `0600`.
2. CI run `33130812610` went RED exactly as intended in the database job with:
   `ERROR: bootstrap temp SQL is not handed to postgres before psql --file`.
3. Production fix commit: `1da5a2e3c779a3773755c3aafbc337ad0393ce79`.
4. Fix writes the complete SQL payload as root, then performs:

```text
chown postgres:postgres "${sql_file}"
chmod 0600 "${sql_file}"
```

before invoking `postgres_psql --file`.
5. The PostgreSQL18/database gate for the fix commit is GREEN; Go and Web gates are also GREEN. Remaining downstream CI jobs do not alter this shell ownership contract.

## Live continuation rule

Do not edit the already-installed immutable auth release in place. For the one-time real Owner bootstrap, fetch the fixed script pinned to commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79`, verify its Git blob SHA `bcfeee6b04e41fb8f8c98340847ebff261375b39`, execute it from a root-only temporary path, then independently verify exactly one active Owner exists. Do not expose the panel through Caddy until localhost real-owner login/session/logout passes.
