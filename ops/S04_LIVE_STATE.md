# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:12 UTC

> Fast continuation file for any new Chat/Agent. Read this first, then `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md`. Do not infer server state from old attempts.

## Current server state — last verified 2026-08-27 23:57:29 UTC

- Host: `testAmir5-3`
- `S00` through `S03-DATABASE`: PASSED.
- `S04-AUTH`: **IN PROGRESS, NOT PASSED**.
- PostgreSQL schema: **2**.
- Migration row:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`
- Last failed S04 attempt rollback removed all S04 service artifacts:
  - `/opt/pvnaive/S04_AUTH.json` absent
  - `/etc/pvnaive/auth.key` absent
  - `/opt/pvnaive/bin/pvnaive` absent
  - `/opt/pvnaive/bin/pvnaive-password` absent
  - `/etc/systemd/system/pvnaive-api.service` absent
  - `/opt/pvnaive/auth/current` absent
  - `/opt/pvnaive/web/current` absent
- `pvnaive-api.service`: not loaded / inactive.
- port `8080`: not listening.
- PostgreSQL: loopback-only on `127.0.0.1:5432` and `[::1]:5432`.
- Caddy: active.
- Caddyfile SHA-256 unchanged:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- SSH/Caddy/firewall were not changed by the S04 failures.
- `/etc/pvnaive/db.env` still expects schema 1 on the live host; the new Stage fixes this atomically during recovery.

## S04 live attempt history

### Attempt 1 — missing `file`

The old S04 bundle stopped at the binary architecture preflight because Ubuntu did not have the `file` utility installed. Rollback completed before substantive S04 mutation. Package `file` was then installed and both bundled executables verified as static x86-64 ELF binaries.

### Attempt 2 — same-second encrypted backup collision

At `2026-08-27 23:46 UTC`:

- pre-S04 encrypted backup succeeded:
  `/var/backups/pvnaive/database/20260827T234607Z/pvnaive.dump.age`
- migration `0002` applied successfully and schema reached 2.
- the immediately following schema-v2 backup tried to use the same second-based directory and failed with:
  `ERROR: backup destination already exists`
- rollback correctly refused destructive v2→v1 rollback because a verified schema-v2 rollback backup had not yet been created.
- schema 2 intentionally remained; S04 artifacts were removed; Caddy/SSH/firewall stayed unchanged.

The backup implementation was fixed to use a unique final directory even for two backups in the same second. A frozen-time collision regression test was added and is part of the green PostgreSQL 18 suite.

### Attempt 3 — API systemd startup failure

Recovery preflight at `2026-08-27 23:49:59 UTC` verified schema 2, exact 0002 checksum, no S04 artifacts, no 8080 listener, and unchanged infrastructure. Stage entered:

`RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER`

It produced a valid encrypted schema-v2 backup:
`/var/backups/pvnaive/database/20260827T234959Z/pvnaive.dump.age`

Then `pvnaive-api.service` failed to remain active. Rollback completed and removed all S04 artifacts while preserving schema 2.

## Decisive service diagnosis

Read-only diagnostics at `2026-08-27 23:57:29 UTC` showed the exact systemd journal error repeatedly:

```text
PVNaive API stopped: PostgreSQL startup check: failed to connect to `user=pvnaive database=`: /var/run/postgresql/.s.PGSQL.5432 (/var/run/postgresql): server error: FATAL: role "pvnaive" does not exist (SQLSTATE 28000)
```

The live environment itself contained the intended application identity:

```text
PVNAIVE_DB_HOST=127.0.0.1
PVNAIVE_DB_PORT=5432
PVNAIVE_DB_NAME=pvnaive
PVNAIVE_DB_USER=pvnaive_app
PVNAIVE_DB_CONNECT_TIMEOUT=5
PVNAIVE_EXPECTED_SCHEMA_VERSION=1
PGPASSFILE=/etc/pvnaive/db.pgpass
```

Permissions on `db.env` and `db.pgpass` were correct.

### Root cause 1 — API DB connection wiring

Old code used `sql.Open("pgx", "")`. An empty pgx DSN does not translate `PVNAIVE_DB_*`; it fell back to PostgreSQL defaults/standard PG environment and therefore attempted OS user `pvnaive` and an empty database name.

TDD/fix evidence:

- RED test commit: `a6f6917dfec393d37787f9932c2ab32caf164962`
- RED CI: `33128145907`, exact compile failure `undefined: databaseDSN`
- production fix: `f64b62acc53d7f264946b1c5612d7f80edac7ad7`

The fixed binary explicitly constructs and validates pgx DSN from `PVNAIVE_DB_HOST`, `PVNAIVE_DB_PORT`, `PVNAIVE_DB_NAME`, `PVNAIVE_DB_USER`, and `PVNAIVE_DB_CONNECT_TIMEOUT`. Production is fail-closed to host `127.0.0.1`, DB `pvnaive`, and user `pvnaive_app`; password remains supplied through `PGPASSFILE`.

### Root cause 2 — schema expectation drift

Live DB schema is 2 after migration 0002, but the S03-created `/etc/pvnaive/db.env` still has `PVNAIVE_EXPECTED_SCHEMA_VERSION=1`. Without a transition, the DB health unit would reject the post-S04 database.

TDD/fix evidence:

- RED test: `tests/stages/S04_db_env_version_test.sh`
- RED: helper missing
- helper implementation: `916e7b37d474ddd47f7eba47f1217c0e764190a8`
- helper invocation test fix: `416fad4e03712b245ef77fbe7cd3a48fdaae5b98`
- guarded Stage runtime/env alignment: `18a11d1a5950217b770ae477ed283f6cdfaa1bb2`

The S04 Stage now atomically sets expected schema to 2, immediately requires `pvnaive-db-health.service` success, and restores expectation to 1 only if a Stage-owned migration actually rolls back to schema 1. If schema remains 2, env remains 2.

### CI masking issue also fixed

The old end-to-end rehearsal had passed `PGHOST/PGDATABASE/PGUSER`, which masked the production systemd wiring bug. The rehearsal now uses the actual `PVNAIVE_DB_*` contract. It also uses the exact production DB name `pvnaive` inside its isolated PostgreSQL container, preserving the binary's fail-closed database-name rule.

Final rehearsal adjustment commit:
`11c54dc1faae99a1491c750b30db9faa44a0c3ae`

## Final green fixed artifact — USE THIS, NOT THE OLD BUNDLE

Full fresh GitHub Actions run:
`33128780602`

All gates passed on source commit:
`11c54dc1faae99a1491c750b30db9faa44a0c3ae`

Passed jobs:

1. Go formatting + vet + tests — SUCCESS
2. Web tests + production build — SUCCESS
3. PostgreSQL 18 migration/health/backup/restore/auth + same-second backup collision + db.env schema transition — SUCCESS
4. end-to-end auth rehearsal using production-style `PVNAIVE_DB_*`, exact DB name `pvnaive`, role `pvnaive_app`, login/session `/me`/CSRF logout/revocation — SUCCESS
5. production bundle build + archive checksum + artifact upload — SUCCESS

New inner production archive:
`PVNaive-S04-11c54dc1faae.tar.gz`

Required inner archive SHA-256:
`52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

GitHub Actions artifact:
`PVNaive-S04-11c54dc1faae99a1491c750b30db9faa44a0c3ae`

Artifact ID:
`9669443464`

GitHub artifact ZIP digest:
`182a368c111cb918dbafe4bd0f974d2c82bd77874978d31ae9d46280157e4011`

## Old bundle — NEVER RETRY

Do not rerun source commit:
`b4803e27af36bb35de33f7dcbe39750aeadc4146`

Do not execute the old extracted server directory:
`/root/pvnaive-s04-deploy-b4803e27af36/PVNaive-S04-b4803e27af36`

It is evidence only. Its old rehearsal masked the database environment bug.

## Exact next action

1. Upload the new artifact ZIP from artifact ID `9669443464` to `testAmir5-3`.
2. Extract only its `PVNaive-S04-11c54dc1faae.tar.gz` and `.sha256` file.
3. Verify inner archive SHA-256 equals exactly:
   `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`
4. Run a read-only recovery preflight verifying current schema=2, exact 0002 checksum, S04 artifacts absent, port 8080 absent, and Caddy SHA unchanged.
5. Run the **new** S04 Stage. Expected mode is `RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER`.
6. Require:
   - `PVNAIVE_EXPECTED_SCHEMA_VERSION=2`
   - DB health service success
   - `pvnaive-api.service` active
   - listener only `127.0.0.1:8080`
   - API liveness/readiness success
   - `/opt/pvnaive/S04_AUTH.json` created
   - Caddy/SSH/firewall unchanged
7. Run an independent S04 postflight. Do not mark S04 PASSED before this.
8. Bootstrap the real Owner interactively only after postflight.
9. Verify localhost login/session/logout using the real Owner.
10. Only then expose `/panel/` and `/api/` through Caddy with backup + `caddy validate` + controlled reload.
11. After external postflight, set `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
