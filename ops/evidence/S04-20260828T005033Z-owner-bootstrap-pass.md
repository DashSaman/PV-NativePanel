# S04 real Owner bootstrap — PASSED

Timestamp: `2026-08-28T00:50:33Z`
Host: `testAmir5-3`

## Result

The real one-time Owner bootstrap succeeded using the fixed, commit-pinned bootstrap script.

Final live output included:

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

The real Owner email is intentionally not copied into this public repository. It is stored only in the live database and was normalized by the bootstrap script.

## Preflight / provenance

Before bootstrap:

- schema version was `2`;
- Owner count was `0`;
- `pvnaive-api.service` was active;
- API listener was only `127.0.0.1:8080`;
- the fixed script was fetched from production-fix commit:
  `1da5a2e3c779a3773755c3aafbc337ad0393ce79`;
- downloaded script Git blob SHA matched exactly:
  `bcfeee6b04e41fb8f8c98340847ebff261375b39`;
- the fixed script retained `0600` protection for the temporary SQL file and handed it to OS user `postgres` before `psql --file`.

The failed first bootstrap attempt had created no Owner. Its root cause was a root-owned `0600` temporary SQL file being unreadable by OS user `postgres`.

## TDD evidence for the bootstrap permission fix

- test-only RED commit: `11bf2add44200dda48c63865263ffa05624b15f9`;
- intended RED CI run: `33130812610`;
- exact RED message: `ERROR: bootstrap temp SQL is not handed to postgres before psql --file`;
- production fix commit: `1da5a2e3c779a3773755c3aafbc337ad0393ce79`;
- follow-up CI run `33130860909`: Go, Web, PostgreSQL18/database, end-to-end auth rehearsal and production bundle all **SUCCESS**.

## Security properties observed

- Raw Owner password was entered only through `/dev/tty` and was not echoed.
- Password hash exists in the Owner row.
- Password hash itself was not printed by verification.
- Temporary bootstrap SQL cleanup completed; the launcher found no leftover bootstrap SQL in `/run` after success.
- Installed immutable auth release was not edited in place; the fixed script was executed from a temporary, commit-pinned and blob-verified copy.

## Exact next action

Run a real localhost authentication test with the live Owner against `http://127.0.0.1:8080`:

1. prompt for the Owner password on `/dev/tty` without echo;
2. POST `/api/v1/auth/login` and require authenticated Owner response;
3. retain the returned session + CSRF cookies;
4. GET `/api/v1/me` and require Owner identity/role;
5. verify exactly one new successful `auth.login` audit event attributable to the real Owner for this test;
6. POST `/api/v1/auth/logout` with `X-CSRF-Token` from the CSRF cookie;
7. require subsequent `/api/v1/me` with the old cookie to return HTTP `401`;
8. verify the tested auth session is stored with a 32-byte token hash and is revoked;
9. preserve API/Caddy/SSH/firewall invariants.

Only after the real localhost login/session/logout/revocation test passes may Caddy exposure for `/panel/` and `/api/` be prepared. Official `S04-AUTH` remains **IN PROGRESS** until controlled Caddy exposure and external postflight pass.