# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:38 UTC after live DB health release promotion

> This is the authoritative fast continuation file. Read `CONTINUE_HERE.md`, the newest file under `ops/evidence/`, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` for supporting history. Do not infer live state from old attempts.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, **not PASSED yet**.
- S04 API/auth/web Stage is installed on `testAmir5-3` and the Stage run itself returned `S04_RESULT=PASSED` / `S04_MODE=LOCALHOST_READY`.
- The first independent postflight core passed and found one blocker: the periodic DB health timer still selected the immutable S03 DB tooling release.
- That blocker has now been repaired live and the repair returned `DB_TIMER_S04_AWARE=true`.
- **A fresh independent postflight is now the only gate before real Owner bootstrap.**
- Do not bootstrap Owner and do not expose the panel through Caddy until the fresh independent postflight returns `S04_POSTFLIGHT=PASSED`.

## Live server facts after repair at 2026-08-28 00:37:59 UTC

- PostgreSQL 18 schema: `2`.
- Migration 0002 checksum:
  `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- `/etc/pvnaive/db.env`: `PVNAIVE_EXPECTED_SCHEMA_VERSION=2`.
- S04 marker exists: `/opt/pvnaive/S04_AUTH.json`.
- Auth release remains `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release remains `/opt/pvnaive/web/releases/20260828T001418Z`.
- Encrypted schema-v2 rollback backup from S04 Stage remains:
  `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age`.
- `/opt/pvnaive/db/current` now selects:
  `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Old S03 immutable release is preserved:
  `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- Real periodic DB health returned `Result=success`, `ExecMainStatus=0` and:
  - `PVNAIVE_DB_HEALTH=OK`
  - `PVNAIVE_SCHEMA_VERSION=2`
  - `PVNAIVE_DB_USER=pvnaive_app`
  - `PVNAIVE_SECRET_DIRECT_SELECT=DENIED`
  - `PVNAIVE_MFA_DIRECT_SELECT=DENIED`
  - `DB_TIMER_S04_AWARE=true`
- `pvnaive-api.service` remains active.
- API listener remains only `127.0.0.1:8080`.
- Liveness remains `{"service":"pvnaive-api","status":"ok"}`.
- Readiness remains `{"ready":true,"status":"ready"}`.
- Caddy remains active; Caddyfile SHA-256 remains exactly:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- `caddy validate` passed. Its existing formatting warning is non-blocking and no Caddyfile mutation occurred.
- SSH remains active as `ssh.service`.
- Repair explicitly reported:
  - `SCHEMA_CHANGED=false`
  - `API_CHANGED=false`
  - `CADDY_CHANGED=false`
  - `SSH_CHANGED=false`
  - `FIREWALL_CHANGED=false`

Newest evidence:
`ops/evidence/S04-20260828T003759Z-db-health-release-promotion-pass.md`

## Prior independent postflight

Evidence:
`ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md`

It returned:

```text
S04_POSTFLIGHT_CORE=PASSED
S04_POSTFLIGHT=BLOCKED
DB_TIMER_S04_AWARE=false
```

All its other core checks passed. The exact blocker it found is now closed by the live immutable DB tooling promotion above.

## Repository repair — TDD complete and final CI green

Active branch: `s04-auth`
PR: `#2`
Repair head used for final verification:
`20ed774d06969a3f4c301fd6072a4db83fcffcca`

TDD sequence:

1. Added `tests/stages/S04_db_release_promotion_test.sh`.
2. RED run `33129595441`: `scripts/db/promote-release.sh` missing.
3. Implemented `scripts/db/promote-release.sh` at `d8c4751b77e59e3c2cdcad2e55e34729c9e51403`; helper atomic/idempotent behavior passed.
4. Strengthened test to require Stage wiring.
5. RED run `33129769272`: S04 Stage did not require the DB release promotion helper.
6. Guarded Stage wiring patch succeeded in `708a4e7fd71011e5b21f136ae7305612f295a258`.
7. Temporary one-shot workflow removed in clean-head commit `20ed774d06969a3f4c301fd6072a4db83fcffcca`.
8. Final clean-head CI `33130012929`: **SUCCESS**.
9. All five gates passed: Go, Web, PostgreSQL18 regression suite, end-to-end authentication rehearsal, production bundle.
10. Live helper download was pinned to `20ed774d...` and independently matched Git blob SHA `0f83469e8f7928d8dbc58d1984fb236552a97e29` before execution.

## Previously deployed pinned S04 artifact

Live API/auth/web are the validated deployment from:

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse old `b4803e27...` bundle.

## Earlier live failures already diagnosed and fixed

1. Ubuntu lacked `file`; package installed and bundled binaries verified x86-64 static ELF.
2. Two backups in one second collided; backup directories now have unique suffixes and collision regression test passes.
3. API opened pgx with an empty DSN and attempted OS role `pvnaive`; fixed to explicit validated `PVNAIVE_DB_*` DSN.
4. DB environment stayed at expected schema1 after migration 0002; fixed atomic schema expectation transition to 2.
5. CI rehearsal had masked the production DB-env contract; fixed to exact production-style variables.
6. Independent postflight found immutable DB health release drift; TDD fix passed and live promotion is now successful.

## Exact next action

Run one fresh **independent S04 postflight** with no configuration mutation. It must independently verify:

1. S04 marker contract, schema2, exact migration identity and restricted DB roles.
2. `/opt/pvnaive/db/current` resolves exactly to `0002-84bb735877d5` and selected `health.sh` is MFA-aware.
3. DB health service last result/exit are successful and timer is active.
4. Direct selected health execution as `pvnaive` returns schema2, app role, signing-secret denial and MFA-secret denial.
5. API service active, loopback-only `127.0.0.1:8080`, live/ready healthy.
6. S04 encrypted rollback backup checksum + decrypt/archive parse remain valid.
7. Caddy active, expected Caddy SHA unchanged, config validates; SSH active; required ports remain present.
8. `DB_TIMER_S04_AWARE=true`.

Only if that independent command returns `S04_POSTFLIGHT=PASSED` may real Owner bootstrap begin. Do not mark the official ledger PASSED yet.
