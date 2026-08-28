# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:20 UTC

> Fast continuation file for any new Chat/Agent. Read this first, then `CONTINUE_HERE.md`, `AGENT_HANDOFF.md`, `ops/DEPLOYMENT_PROGRESS.md`, and the newest S04 evidence file. Do not infer live state from earlier failed attempts.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`; the fixed S04 Stage itself is installed and running successfully on localhost, and the independent postflight core passed. **The only remaining blocker before Owner bootstrap is that the periodic DB health timer still selects the immutable S03-era DB tooling release.**

Do NOT mark S04 PASSED yet.

## Current live server state — independently verified 2026-08-28 00:20:41 UTC

Host: `testAmir5-3`

- PostgreSQL 18 schema: `2`.
- Migration 0002:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- `/etc/pvnaive/db.env` expects schema 2.
- S04 marker exists and passed its independent contract check:
  `/opt/pvnaive/S04_AUTH.json`.
- Auth release:
  `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release:
  `/opt/pvnaive/web/releases/20260828T001418Z`.
- Encrypted schema-v2 rollback backup:
  `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age`.
- `/etc/pvnaive/auth.key`: `root:pvnaive`, mode `0640`, exactly 32 bytes.
- `pvnaive-api.service`: enabled + active, running as `pvnaive:pvnaive`, zero restarts at postflight.
- API listener: only `127.0.0.1:8080`.
- API liveness: `{"service":"pvnaive-api","status":"ok"}`.
- API readiness: `{"ready":true,"status":"ready"}`.
- PostgreSQL remains loopback-only.
- Caddy remains active and Caddyfile SHA-256 is unchanged:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH remains active.
- S04 has not changed Caddy config, SSH config, or firewall.

## Independent postflight result

Evidence file:
`ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md`

Result:

```text
S04_POSTFLIGHT_CORE=PASSED
S04_POSTFLIGHT=BLOCKED
DB_TIMER_S04_AWARE=false
BLOCKER=periodic pvnaive-db-health.service still uses the S03-era health release and does not verify S04 MFA secret tables
```

Everything else in the independent postflight passed: marker/artifact identity, DB schema/roles, direct S04-aware DB health including MFA-secret boundary, API systemd/listener, encrypted rollback backup, Caddy and SSH invariants.

## Why the periodic health check is the only blocker

The health unit is intentionally stable and executes:

`/opt/pvnaive/db/current/scripts/db/health.sh`

That symlink still points to the immutable S03 tooling release (`0001-7f66adefd8f0`). The S04-aware health script under `/opt/pvnaive/auth/current/scripts/db/health.sh` already passes schema-2 and MFA-table checks, but the periodic timer does not select it.

Correct design: keep the unit unchanged, preserve the old S03 immutable DB release, create a schema-2 immutable DB tooling release, then atomically promote `/opt/pvnaive/db/current`.

Expected new DB tooling release ID:
`0002-84bb735877d5`.

## Repository fix — TDD status

Development branch: `s04-auth`
PR: `#2`

TDD sequence:

1. Added `tests/stages/S04_db_release_promotion_test.sh`.
2. RED run `33129595441`: `scripts/db/promote-release.sh` did not exist.
3. Added `scripts/db/promote-release.sh` at `d8c4751b77e59e3c2cdcad2e55e34729c9e51403`.
4. Helper itself passed atomic/idempotent immutable release promotion test.
5. Strengthened regression test to require Stage wiring.
6. RED run `33129769272`: `S04 stage does not require the DB release promotion helper`.
7. Guarded Stage patch succeeded; production Stage patch commit:
   `708a4e7fd71011e5b21f136ae7305612f295a258`.
8. Temporary one-shot patch workflow removed in user-authored commit:
   `20ed774d06969a3f4c301fd6072a4db83fcffcca`.
9. Clean-head full CI run: `33130012929`; verify final result before any live mutation.

The fixed S04 Stage now promotes DB tooling for both fresh/recovery installs and existing-marker verification. If a Stage-owned migration genuinely rolls back to schema 1, its rollback path also restores the prior DB tooling release symlink; if schema remains 2, the v2 tooling release remains selected.

## Previously deployed pinned S04 artifact

The live API/auth/web installation remains from the verified artifact:

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- archive SHA-256:
  `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse the older `b4803e27...` bundle.

## Earlier failures already fixed

1. Missing `file` utility on Ubuntu — installed and architecture checks passed.
2. Same-second encrypted-backup directory collision — backup release path now has a unique suffix; regression tested.
3. API used empty pgx DSN and tried OS role `pvnaive` — fixed to explicit fail-closed `PVNAIVE_DB_*` DSN; production-style rehearsal passed.
4. `PVNAIVE_EXPECTED_SCHEMA_VERSION` stayed at 1 after migration 0002 — fixed atomically to 2 with rollback alignment.
5. CI rehearsal previously masked production DB environment behavior — fixed to use exact production-style `PVNAIVE_DB_*` and DB name `pvnaive`.

## Exact next action

1. Require final clean-head CI for the DB release promotion fix to pass Go, Web, PostgreSQL18 regression gates, end-to-end rehearsal, and bundle.
2. Do one atomic live DB tooling release promotion from immutable S03 release to `0002-84bb735877d5`; do not change DB schema, API binary, Caddy, SSH or firewall.
3. Require periodic `pvnaive-db-health.service` to return `Result=success`, `ExecMainStatus=0`, and its selected health script to contain the S04 MFA checks.
4. Rerun independent S04 postflight and require `DB_TIMER_S04_AWARE=true` plus `S04_POSTFLIGHT=PASSED`.
5. Only then bootstrap the real Owner interactively.
6. Verify localhost Owner login/session/logout.
7. Only after localhost auth is proven, expose `/panel/` and `/api/` through Caddy using backup + validate + controlled reload.
8. Run external postflight; only then mark `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
