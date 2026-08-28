# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after final localhost postflight PASS and Owner bootstrap permission fix

> Authoritative fast continuation file. Read `CONTINUE_HERE.md`, newest S04 evidence, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` before acting on the live host.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, not official PASSED yet because public Caddy exposure/external postflight have not run.
- S04 API/auth/web localhost deployment is installed and healthy on `testAmir5-3`.
- Final corrected independent read-only postflight at `2026-08-28T00:43:44Z` returned complete PASS and `NEXT=BOOTSTRAP_REAL_OWNER`.
- First real Owner bootstrap attempt failed before INSERT with `psql: error: /run/pvnaive-owner-bootstrap.<random>.sql: Permission denied`.
- Root cause: temp SQL was root-owned mode `0600`, while `psql --file` deliberately runs as OS user `postgres`.
- No Owner was created by the failed attempt.
- Repo fix is in `s04-auth` commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79`: after fully writing the SQL payload, bootstrap changes ownership to `postgres:postgres` and retains mode `0600` before invoking psql.
- Regression test commit `11bf2add44200dda48c63865263ffa05624b15f9` produced intended RED CI run `33130812610` with `ERROR: bootstrap temp SQL is not handed to postgres before psql --file`.
- On fix commit, PostgreSQL18/database, Go, Web and end-to-end rehearsal gates are GREEN; bundle build may still be downstream/nonessential to the one-time script repair.
- Next permitted action: execute the fixed bootstrap script from a commit-pinned, Git-blob-verified temporary copy. Do not edit the installed immutable auth release in place.

## Final independently verified live facts before bootstrap retry

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
- `pvnaive-api.service` enabled/active as `pvnaive:pvnaive`, `NRestarts=0`.
- API listener only `127.0.0.1:8080`; live/ready passed.
- Encrypted rollback backup `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age` validated.
- PostgreSQL listeners only loopback.
- Caddy active; Caddyfile SHA unchanged: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; validation passed.
- SSH active.

## Owner bootstrap fix provenance

Newest evidence: `ops/evidence/S04-20260828T0046-owner-bootstrap-temp-sql-permission-fix.md`.

- test-only RED commit: `11bf2add44200dda48c63865263ffa05624b15f9`
- RED CI: `33130812610`
- production fix commit: `1da5a2e3c779a3773755c3aafbc337ad0393ce79`
- fixed script Git blob SHA: `bcfeee6b04e41fb8f8c98340847ebff261375b39`

The temp SQL file contains the Argon2id PHC hash and bootstrap metadata, not the raw password. Cleanup removes the temp file on failure. The fixed script still keeps the file at mode `0600`; only the intended `postgres` OS account owns/reads it when psql executes.

## Exact next action

1. Fetch `scripts/auth/bootstrap-owner.sh` pinned to commit `1da5a2e3c779a3773755c3aafbc337ad0393ce79` into a root-only temporary path.
2. Verify Git blob SHA exactly `bcfeee6b04e41fb8f8c98340847ebff261375b39`.
3. Before execution verify Owner count is zero.
4. Execute the fixed copy interactively from root TTY; enter real email/display/password/confirmation. Do not send the password in chat.
5. Require `PVNAIVE_OWNER_BOOTSTRAP=PASSED`.
6. Independently verify exactly one active Owner and password-hash presence without printing the hash.
7. Test real localhost login/session/me/CSRF logout/revocation.
8. Only then attempt Caddy exposure; only after external postflight may the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
