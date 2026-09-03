# PVNaive — Canonical Project Status

Last updated: 2026-09-03 08:43 Asia/Tehran

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Main post-merge workflow runs for this docs-only commit: **none observed**.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub gates are green, but a fresh real HTTP/1.1 + HTTP/2 rehearsal remains the sole merge gate.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; Task16 PG18, Exact Accounting and Pinned Forwardproxy are green; repository-wide CI remains blocked in the database job at `RLS coverage check failed: 43/42`.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-03 08:43 Asia/Tehran:

- `GET http://127.0.0.1:8080/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`;
- `systemctl is-active pvnaive-api`: active;
- `systemctl is-active caddy`: inactive;
- Docker inventory could not be read because the execution context lacks permission for `/var/run/docker.sock`.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` on `/opt/pvnaive-backup-hotfix` remain historical S04-era ledgers (last updated 2026-08-27 / 2026-08-29 in the observed worktrees) and contain no newer validated Task13/Task16 completion. They do not override this file.

## Worker / orchestration capacity

Three SentinelX hosts are connected: `TrPaqet`, `pv-worker-main`, and `pv-primary`. All are currently connected; development work remains subject to the available execution slot and must not use Production as a test lane.

- `TrPaqet`: Task13 final HTTP/1.1 + HTTP/2 rehearsal.
- `pv-worker-main`: schema21-aware RLS assertion fix and full Task16 gate rerun.
- `pv-primary`: Production-only audit, backup, rollback and deploy lane.

## Immediate execution order

1. Keep #64 and #81 draft.
2. On the first executable development worker, run the final exact-head Task13 live rehearsal.
3. On the development lane, patch only the schema21-aware RLS expectation and rerun all Task16 gates.
4. If and only if Task13 live proof and Task16 full gates pass, perform fresh encrypted backup + rollback preparation and then consider promotion.
