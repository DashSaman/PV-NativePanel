# PVNaive — Canonical Project Status

Last updated: 2026-09-02 (automation verification)

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Main CI: `33623286003` **SUCCESS**.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates green, live HTTP/1.1 + HTTP/2 rehearsal pending.
- Task16: draft PR #81 / issue #79; not promotable until repository-wide PostgreSQL18 CI and explicit Task16 contract execution are green.

## Production truth

Production remains on Task15/schema20. `pv-primary` was connected but inactive this run under the SentinelX Free-plan one-active-host limit, so no fresh command-level Production health claim is made here. No deploy, restart, reload, migration, DB write, credential rotation or other Production mutation occurred.

## Task13 blocker

`TrPaqet` was reconciled to exact Task13 head and the rehearsal binaries built with Go 1.26.3. `tests/stages/Task13_api_session_kill_rehearsal.sh` then stopped before execution because `jq` is missing on the worker. This is not counted as a pass. Install/provide `jq` on the development slot or use another executable development worker; never substitute Production.

## Task16 next slice

Keep trusted schema17 peer/accounting lineage, tenant forced RLS, exact 30-day retention, hard server-side maximum 500, explicit maintenance-only purge confirmation, checksum integrity and disposable rollback. Wire and run the real schema21 PostgreSQL18 integration test in CI before promotion.

## Persistent reports / assignments

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers and do not override exact GitHub/Production evidence.

- `TrPaqet`: Task13 final rehearsal; blocked on missing `jq`.
- `pv-worker-main`: Task16 schema21/PG18 lane when executable.
- `pv-primary`: Production-only audit/backup/rollback/deploy lane.

No task is DONE from local, stale or partial evidence. Preserve truthful accounting/session semantics and rollback safety.