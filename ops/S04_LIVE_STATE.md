# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:41 UTC after corrected diagnosis of the latest independent postflight

> This is the authoritative fast continuation file. Read `CONTINUE_HERE.md`, the newest file under `ops/evidence/`, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` for supporting history. Do not infer live state from old attempts.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, **not PASSED yet**.
- S04 API/auth/web Stage is installed on `testAmir5-3` and the Stage run itself returned `S04_RESULT=PASSED` / `S04_MODE=LOCALHOST_READY`.
- The periodic DB health blocker discovered by the first postflight has been repaired live.
- `/opt/pvnaive/db/current` now selects schema2 immutable release `0002-84bb735877d5` and live verification returned `DB_TIMER_S04_AWARE=true`.
- A fresh independent postflight at `2026-08-28T00:41:21Z` stopped only because the harness expected DB-role booleans as `t/f`; PostgreSQL returned `true/false`. The observed privilege values are correct.
- The latest postflight was read-only and reported `NO_CONFIGURATION_CHANGES_MADE=true`.
- **The only immediate gate is rerunning the corrected read-only postflight.**
- Do not bootstrap Owner and do not expose the panel through Caddy until that corrected postflight returns `S04_POSTFLIGHT=PASSED`.

## Latest independent postflight — false negative, not server failure

Evidence:
`ops/evidence/S04-20260828T004121Z-postflight-role-format-harness-false-negative.md`

Before stopping, the postflight independently passed:

- S04 marker contract.
- Auth/web release linkage.
- Installed API binary/password helper/systemd unit/web integrity against the pinned bundle.
- `/etc/pvnaive/auth.key` metadata exactly `root|pvnaive|640|32`.
- PostgreSQL schema `2`.
- Exact migration 0002 identity:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.

It observed:

```text
pvnaive_app|true|false|false|false|false|false|false
pvnaive_owner|false|false|false|false|false|false|false
```

This is the intended contract: application role has LOGIN only, no listed elevated privileges; owner role has no LOGIN and no listed elevated privileges. The harness incorrectly compared against `t/f` and therefore raised `ERROR=pvnaive_app privilege contract failed` even though the values themselves were correct.

## Live server facts after DB health release repair at 2026-08-28 00:37:59 UTC

- PostgreSQL 18 schema: `2`.
- Migration 0002 checksum:
  `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- `/etc/pvnaive/db.env`: `PVNAIVE_EXPECTED_SCHEMA_VERSION=2`.
- S04 marker exists: `/opt/pvnaive/S04_AUTH.json`.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- Encrypted schema-v2 rollback backup:
  `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age`.
- `/opt/pvnaive/db/current`:
  `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Old S03 immutable release preserved:
  `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- Real periodic DB health returned `Result=success`, `ExecMainStatus=0` and:
  - `PVNAIVE_DB_HEALTH=OK`
  - `PVNAIVE_SCHEMA_VERSION=2`
  - `PVNAIVE_DB_USER=pvnaive_app`
  - `PVNAIVE_SECRET_DIRECT_SELECT=DENIED`
  - `PVNAIVE_MFA_DIRECT_SELECT=DENIED`
  - `DB_TIMER_S04_AWARE=true`
- `pvnaive-api.service` active.
- API listener only `127.0.0.1:8080`.
- Liveness `{"service":"pvnaive-api","status":"ok"}`.
- Readiness `{"ready":true,"status":"ready"}`.
- Caddy active; Caddyfile SHA-256 unchanged:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- Caddy validate passed; existing formatting warning remains non-blocking.
- SSH active as `ssh.service`.
- Repair explicitly reported schema/API/Caddy/SSH/firewall unchanged.

## Repository repair — TDD complete and final CI green

Active branch: `s04-auth`
PR: `#2`
Repair head used for final verification:
`20ed774d06969a3f4c301fd6072a4db83fcffcca`

Final clean-head CI `33130012929`: **SUCCESS** across all five gates: Go, Web, PostgreSQL18 regression suite, end-to-end auth rehearsal, production bundle.

## Installed S04 artifact

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse old `b4803e27...` bundle.

## Earlier live failures already diagnosed and fixed

1. Ubuntu lacked `file`; package installed and bundled binaries verified x86-64 static ELF.
2. Two backups in one second collided; backup directories now have unique suffixes and collision regression test passes.
3. API used an empty pgx DSN and attempted OS role `pvnaive`; fixed to explicit validated `PVNAIVE_DB_*` DSN.
4. DB environment stayed at expected schema1 after migration 0002; fixed atomic schema expectation transition to 2.
5. CI rehearsal had masked the production DB-env contract; fixed to exact production-style variables.
6. Independent postflight found immutable DB health release drift; TDD fix passed and live promotion succeeded.
7. Corrected postflight diagnosis: the latest stop was only a boolean-output-format mismatch in the harness (`t/f` vs `true/false`), not a privilege drift.

## Exact next action

Run one corrected **independent S04 postflight** with no configuration mutation. The role contract must compare against:

```text
pvnaive_app|true|false|false|false|false|false|false
pvnaive_owner|false|false|false|false|false|false|false
```

The command must then continue through schema2 immutable DB release, MFA-aware periodic health, API active/loopback-only/live/ready, encrypted rollback backup validity, Caddy SHA + validation, SSH and required ports. Only if it finally prints:

```text
S04_POSTFLIGHT_CORE=PASSED
DB_TIMER_S04_AWARE=true
S04_POSTFLIGHT=PASSED
NEXT=BOOTSTRAP_REAL_OWNER
```

may real Owner bootstrap begin. Do not advance the official ledger yet.
