# PVNaive — Canonical Project Status

Last updated: 2026-09-03

This file is current repository + Production truth. Historical reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage, online, IP or session history. Production changes require a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification. No credential rotation from read-only flows.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub exact-head gates are green, but fresh real HTTP/1.1 + HTTP/2 live proof is still required.
- Task16: draft PR #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`.
- Task16 exact-head runs: Schema21 TDD **SUCCESS**, Exact Accounting **SUCCESS**, Pinned Forwardproxy **SUCCESS**, repository-wide CI **FAILURE** in database job.
- CI failure is a real latest-schema fixture mismatch: `scripts/db/health.sh` reports `RLS coverage check failed: 43/42` after schema21 migration; migration, backup/restore and earlier checks pass.
- No PR is eligible for merge or promotion until its full exact-head gates are green.

## Production truth

Fresh read-only probe on `pv-primary` at 2026-09-03 02:42 local cycle:

- `/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`.
- `/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`.
- Docker inventory was not readable from the SentinelX execution context due to Docker socket permission denial; service health claims are limited to the successful HTTP probes above.
- No Production deploy, restart, reload, migration, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent reports and workers

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers and contain no newer completion that overrides exact GitHub evidence.

- `TrPaqet`: Task13 live rehearsal lane.
- `pv-worker-main`: Task16 generic fixture correction + full rerun lane when executable.
- `pv-primary`: Production-only audit/backup/rollback/deploy lane.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE until their full gates and live proof are complete.
2. Correct only the latest-schema CI expectations; preserve Task15 schema20-specific fixtures.
3. Run Task13 real HTTP/1.1 + HTTP/2 rehearsal on development infrastructure, never on Production.
4. If gates become green, take fresh encrypted backup + rollback state before any deploy, then postflight and retain rollback evidence.
