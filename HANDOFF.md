# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `26aa74dddfd23535e45837f21531cf67ea2fd238`, verified merge of PR #54 / Task15.
- Exact-main CI run `33471780919`: **SUCCESS**.
- Open PRs: old draft #4 only; it is not current roadmap work.
- Production schema: **20**.
- Production source: `26aa74dddfd23535e45837f21531cf67ea2fd238`.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS**, current-schema20 publication/recovery incomplete.
- Task16: **IN PROGRESS**, schema21 now unblocked by stable schema20 but not merged/deployed.

## Fresh Production state

The Task15 guarded Production rollout completed on 2026-09-01 after a fresh encrypted DB/config backup and separately checksummed rollback snapshot.

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: active.
- `/api/v1/health/live`: HTTP 200.
- `/api/v1/health/ready`: HTTP 200 with DB/schema ready.
- public SNI-correct panel: HTTP 200.
- public SNI-correct API readiness: HTTP 200.
- database schema and expected schema: 20/20.
- release source: exact main `26aa74dddfd23535e45837f21531cf67ea2fd238`.
- Caddy `v2.11.2`, pinned Forwardproxy module present, live Caddyfile validates.

Exact rollout evidence: `ops/evidence/TASK15-20260901-schema20-production-pass.md`.

The initial direct DB backup attempt inherited `pvnaive_app` and correctly failed on protected secret tables before any mutation. The installed canonical backup wrapper uses the local postgres OS/DB role and then succeeded. Do not grant application-role access to protected tables to solve backup failures.

## Task13

Remote publication remains partial. Reconstruct/recover the complete integration on top of exact schema20 main and rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence. Mandatory invariants: exact runtime-credential/node/boot/session tuple, sibling survival, no credential revoke, no Caddy reload/restart, idempotent kill, tenant/role isolation and exactly-once final accounting.

## Task16

Reconcile schema21 directly onto current schema20 main. Prove exact 30-day retention, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, bounded reads/pagination, coherent migration/checksums, PostgreSQL18 behavior and rollback safety before merge/DONE.

## Worker capacity

Three SentinelX hosts are connected but the current plan permits one active host. `pv-primary` is currently executable; worker hosts remain unavailable for command execution. Existing worker reports are historical until a worker is freshly reachable. Repository/CI work can continue independently.

To resume parallel workers, disconnect an unused connected host so a worker becomes active or change the SentinelX host limit.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.

## Exact next sequence

1. Merge the Task15 Production evidence/docs reconciliation only after its CI is green.
2. Recover/reconcile Task13 against exact schema20 main and publish only after its complete exact-session-kill gates pass.
3. Rebase/reconcile Task16 onto schema20 main and rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before schema21 publication.
4. Keep draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
5. Keep canonical docs/evidence synchronized only from verified GitHub and Production truth.
