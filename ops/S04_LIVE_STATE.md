# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 after independent postflight + green DB-release fix CI

> This is the authoritative fast continuation file. Read `CONTINUE_HERE.md`, the newest file under `ops/evidence/`, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` for supporting history. Do not infer live state from old attempts.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, **not PASSED yet**.
- S04 API/auth/web Stage is installed on `testAmir5-3` and the Stage run itself returned `S04_RESULT=PASSED` / `S04_MODE=LOCALHOST_READY`.
- Independent S04 postflight core passed, but final postflight is blocked only by the periodic DB health timer selecting the S03-era immutable DB tooling release.
- Do not bootstrap Owner and do not expose the panel through Caddy until the DB tooling release is promoted and independent postflight returns `S04_POSTFLIGHT=PASSED`.

## Live server facts verified by independent postflight at 2026-08-28 00:20:41 UTC

- PostgreSQL 18 schema: `2`.
- Migration 0002:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- `/etc/pvnaive/db.env`: `PVNAIVE_EXPECTED_SCHEMA_VERSION=2`.
- S04 marker: `/opt/pvnaive/S04_AUTH.json` — contract passed.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- Encrypted schema-v2 rollback backup:
  `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age` — checksum + decrypt/parse passed.
- Auth key: `root:pvnaive`, `0640`, 32 bytes.
- `pvnaive-api.service`: enabled + active, runs as `pvnaive:pvnaive`, zero restarts at postflight.
- API listener: only `127.0.0.1:8080`.
- Liveness: `{"service":"pvnaive-api","status":"ok"}`.
- Readiness: `{"ready":true,"status":"ready"}`.
- PostgreSQL remains loopback-only.
- Caddy active; Caddyfile SHA-256 unchanged:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH active; Caddy/SSH/firewall unchanged by S04.

## Independent postflight blocker

Evidence:
`ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md`

Observed:

```text
S04_POSTFLIGHT_CORE=PASSED
S04_POSTFLIGHT=BLOCKED
DB_TIMER_S04_AWARE=false
BLOCKER=periodic pvnaive-db-health.service still uses the S03-era health release and does not verify S04 MFA secret tables
```

Direct S04-aware health from `/opt/pvnaive/auth/current/scripts/db/health.sh` already passed schema2, `pvnaive_app`, signing-key denial and MFA secret-table denial. The timer unit itself also exits 0, but its selected immutable script is older and does not test the S04 MFA boundary.

## Root cause

`pvnaive-db-health.service` intentionally executes:

`/opt/pvnaive/db/current/scripts/db/health.sh`

`/opt/pvnaive/db/current` still points to the immutable S03 release:

`/opt/pvnaive/db/releases/0001-7f66adefd8f0`

The correct repair is to keep the systemd unit unchanged, preserve the S03 release, create a schema2 immutable DB tooling release and atomically promote the `current` symlink.

Expected new release:

`/opt/pvnaive/db/releases/0002-84bb735877d5`

## Repository repair — TDD complete and final CI green

Active branch: `s04-auth`
PR: `#2`
Clean branch head used for final verification:
`20ed774d06969a3f4c301fd6072a4db83fcffcca`

TDD sequence:

1. Added `tests/stages/S04_db_release_promotion_test.sh`.
2. RED run `33129595441`: `scripts/db/promote-release.sh` missing.
3. Implemented `scripts/db/promote-release.sh` at `d8c4751b77e59e3c2cdcad2e55e34729c9e51403`; helper atomic/idempotent behavior passed.
4. Strengthened test to require Stage wiring.
5. RED run `33129769272`: `ERROR: S04 stage does not require the DB release promotion helper`.
6. Guarded Stage wiring patch succeeded in commit `708a4e7fd71011e5b21f136ae7305612f295a258`.
7. Temporary one-shot workflow removed in normal-user clean-head commit `20ed774d06969a3f4c301fd6072a4db83fcffcca`.
8. Final clean-head CI run `33130012929`: **SUCCESS**.
9. All five gates passed: Go, Web, PostgreSQL18 regression suite, end-to-end authentication rehearsal, production bundle.
10. Compare deployed source `11c54dc1...` → repair head `20ed774d...` shows the only change under `scripts/db` is the addition of `promote-release.sh`; migration files and existing DB scripts are unchanged.

The fixed Stage now performs the DB tooling promotion on fresh/recovery paths and existing-marker verification. Its rollback restores the old DB tooling symlink only if a Stage-owned migration actually returns schema to 1; if schema remains 2, schema2 tooling remains selected.

## Previously deployed pinned S04 artifact

Live API/auth/web are still the validated deployment from:

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse old `b4803e27...` bundle.

## Earlier live failures already diagnosed and fixed

1. Ubuntu lacked `file`; package installed and bundled binaries verified x86-64 static ELF.
2. Two backups in one second collided; backup directories now have unique suffixes and collision regression test passes.
3. API opened pgx with an empty DSN and attempted OS role `pvnaive`; fixed to explicit validated `PVNAIVE_DB_*` DSN for `pvnaive_app`/`pvnaive` and production-style rehearsal.
4. DB environment stayed at expected schema1 after migration 0002; fixed atomic schema expectation transition to 2 with rollback consistency.
5. CI rehearsal had masked the production DB-env contract; fixed to exact production DB name and `PVNAIVE_DB_*` variables.
6. Independent postflight then found the remaining immutable DB health release drift; repository fix above is green.

## Exact next action

1. On the live server, preflight schema2, exact migration checksum, S04 marker/API, current S03 DB release, Caddy SHA and timer state.
2. Build an exact schema2 DB tooling source from the already-installed pinned S04 DB files plus the immutable commit-pinned `promote-release.sh` from repair head `20ed774d...`.
3. Verify the downloaded helper against its Git blob SHA `0f83469e8f7928d8dbc58d1984fb236552a97e29` before execution.
4. Atomically create/select `/opt/pvnaive/db/releases/0002-84bb735877d5`; preserve `0001-7f66adefd8f0`.
5. Start `pvnaive-db-health.service`; require `Result=success`, exit 0, schema2 and `PVNAIVE_MFA_DIRECT_SELECT=DENIED` from the selected release.
6. If any post-switch invariant fails, rollback only the DB `current` symlink to the S03 release; do not alter schema/API/Caddy/SSH/firewall.
7. Rerun independent S04 postflight; require `DB_TIMER_S04_AWARE=true` and `S04_POSTFLIGHT=PASSED`.
8. Only then bootstrap the real Owner, verify localhost authentication, expose `/panel/` + `/api/` through Caddy with backup/validate/controlled reload, perform external postflight, then mark `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
