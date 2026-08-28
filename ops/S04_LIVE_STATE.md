# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:07 UTC

> This is the fast continuation file for a new Chat/Agent. Read `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` as authoritative history, then use this file for the latest S04 state.

## Current server state

- Host: `testAmir5-3`
- `S03-DATABASE=PASSED`
- S04 is **NOT PASSED** yet.
- PostgreSQL schema is currently **2**.
- Migration row is:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`
- The last S04 failure rollback completed and removed all S04 service artifacts:
  - `/opt/pvnaive/S04_AUTH.json` absent
  - `/etc/pvnaive/auth.key` absent
  - `/opt/pvnaive/bin/pvnaive` absent
  - `/opt/pvnaive/bin/pvnaive-password` absent
  - `/etc/systemd/system/pvnaive-api.service` absent
  - `/opt/pvnaive/auth/current` absent
  - `/opt/pvnaive/web/current` absent
- `pvnaive-api.service` is not loaded/active.
- port 8080 is not listening.
- PostgreSQL listens only on loopback port 5432.
- Caddy is active and Caddyfile SHA remains:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- SSH/Caddy/firewall have not been changed by S04.

## Decisive service failure diagnosis

Read-only diagnostics at `2026-08-27 23:57:29 UTC` showed the real systemd journal failure repeatedly:

```text
PVNaive API stopped: PostgreSQL startup check: failed to connect to `user=pvnaive database=`: /var/run/postgresql/.s.PGSQL.5432 (/var/run/postgresql): server error: FATAL: role "pvnaive" does not exist (SQLSTATE 28000)
```

The production DB environment itself was correct for the application role:

```text
PVNAIVE_DB_HOST=127.0.0.1
PVNAIVE_DB_PORT=5432
PVNAIVE_DB_NAME=pvnaive
PVNAIVE_DB_USER=pvnaive_app
PVNAIVE_DB_CONNECT_TIMEOUT=5
PVNAIVE_EXPECTED_SCHEMA_VERSION=1
PGPASSFILE=/etc/pvnaive/db.pgpass
```

Permissions on `/etc/pvnaive/db.env` and `/etc/pvnaive/db.pgpass` were correct.

### Root cause 1: API ignored PVNAIVE_DB_* connection identity

The old binary used `sql.Open("pgx", "")`. With an empty DSN, pgx fell back to standard PostgreSQL environment/defaults, so under systemd it attempted OS user `pvnaive` and an empty database name instead of `pvnaive_app` / `pvnaive`.

TDD evidence:

- regression test commit: `a6f6917dfec393d37787f9932c2ab32caf164962`
- RED GitHub Actions run: `33128145907`
- exact RED: `undefined: databaseDSN`
- production fix commit: `f64b62acc53d7f264946b1c5612d7f80edac7ad7`

The fixed API builds an explicit validated pgx DSN from `PVNAIVE_DB_HOST`, `PVNAIVE_DB_PORT`, `PVNAIVE_DB_NAME`, `PVNAIVE_DB_USER`, and `PVNAIVE_DB_CONNECT_TIMEOUT`. It fail-closes unless host is `127.0.0.1`, DB is `pvnaive`, and user is `pvnaive_app`. `PGPASSFILE` remains the password source.

The S04 end-to-end rehearsal was also changed so it no longer masks this class of bug with `PGHOST/PGDATABASE/PGUSER`; commit: `bf545bfaf31d1f1418697e78c6ab40684c3e5f35`.

### Root cause 2: expected schema version drift

After S04 migration the DB is schema 2, but `/etc/pvnaive/db.env` still said `PVNAIVE_EXPECTED_SCHEMA_VERSION=1`. That would make the existing DB health service fail after an otherwise successful S04.

TDD evidence:

- test: `tests/stages/S04_db_env_version_test.sh`
- RED CI: missing `scripts/db/set-expected-schema-version.sh`
- helper implementation commit: `916e7b37d474ddd47f7eba47f1217c0e764190a8`
- helper test invocation fix: `416fad4e03712b245ef77fbe7cd3a48fdaae5b98`
- Stage runtime/env alignment commit produced by guarded one-shot workflow:
  `18a11d1a5950217b770ae477ed283f6cdfaa1bb2`

The Stage now updates expected schema to 2 atomically, requires DB health success, and restores expected schema to 1 only if a Stage-owned migration actually rolls back to v1. In recoverable schema-v2 state it keeps expected schema 2.

### Root cause 3 from previous attempt: same-second backup collision

The previous Stage could create two backups in the same second. `scripts/db/backup.sh` now gives each completed backup a unique final directory using the mktemp suffix + BASHPID. A regression test reproduces two backups with an identical frozen timestamp.

- backup fix is present on `s04-auth`
- collision harness repair commit: `3cb954882601a4d66a5df567fb10b5e92a5bb769`

## Active branch / CI

- Development branch: `s04-auth`
- PR: `#2`
- Full verification trigger commit: `d5a7c640bb17b72ab97aeddab8f6f06ba26533be`
- Full CI run: `33128625046`

Do **not** deploy until this run (or a later head run) has all of these green:

1. Go formatting/vet/tests
2. Web tests/build
3. PostgreSQL 18 gates, including backup collision and db.env transition
4. end-to-end S04 rehearsal using production-style `PVNAIVE_DB_*` variables
5. production bundle build + checksum

## Old bundle — DO NOT RETRY

Do not rerun the old bundle from source commit `b4803e27af36bb35de33f7dcbe39750aeadc4146` even though it passed its old CI. Its rehearsal masked the production environment wiring bug.

Old server copy remains only as evidence:
`/root/pvnaive-s04-deploy-b4803e27af36/PVNaive-S04-b4803e27af36`

## Exact next action

1. Finish full CI on the latest `s04-auth` head.
2. Build/download the **new** green S04 artifact and record its exact source commit, filename and SHA-256 here.
3. User uploads that artifact to `testAmir5-3`.
4. Run a read-only recovery preflight: schema=2, migration checksum match, no S04 artifacts, no 8080 listener, Caddy SHA unchanged.
5. Run the new S04 bundle in `RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER`.
6. Require API active only on `127.0.0.1:8080`, DB health success with expected schema 2, marker creation, and unchanged Caddy/SSH/firewall.
7. Run an independent S04 postflight.
8. Only then bootstrap the real Owner interactively.
9. Only after localhost authentication is proven, expose `/panel/` and `/api/` through Caddy with backup + validate + controlled reload.
10. After final external postflight, mark S04 PASSED and S05 NEXT.
