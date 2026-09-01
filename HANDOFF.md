# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `b5d32352f99f362c6c4850703c9efc27544f966c`, verified merge of PR #59 (current-state reconciliation only).
- Exact-main CI run `33491705169`: **SUCCESS**.
- Open roadmap PRs: none. Old draft #4 is not current roadmap work.
- Production schema: **20**.
- Production runtime source remains the Task15 rollout commit `26aa74dddfd23535e45837f21531cf67ea2fd238`; PRs #57/#58/#59 did not change runtime artifacts.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS**, complete current-schema20 publication plus fresh real HTTP1/HTTP2 rehearsal still required.
- Task16: **IN PROGRESS**, schema21 candidate fails design acceptance until retention/pagination are server-bounded.

## Fresh Production state

Fresh read-only verification at 2026-09-01T10:13Z confirmed all four services `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent` are active. `/api/v1/health/ready` reports `ready=true`, `db=ok`, `schema=ok`. No Production mutation, restart, reload or migration was performed during this reconciliation.

Exact Task15 rollout evidence remains `ops/evidence/TASK15-20260901-schema20-production-pass.md`; fresh encrypted backup and rollback material remain retained under `/var/backups/pvnaive`.

## Task13

Remote publication remains partial. The publication branch `lead/task13-kill-session-publish-20260901` still points to `2af0e4edfb3d66047835cd886d46b94221cf77b7`; its latest commit modifies only the session-control client test file and remains stale relative to schema20 main. Do not merge it wholesale. Reconstruct/recover the complete integration on top of exact schema20 main and rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence. Mandatory invariants: exact runtime-credential/node/boot/session tuple, sibling survival, no credential revoke, no Caddy reload/restart, idempotent kill, tenant/role isolation and exactly-once final accounting.

## Task16

Do not merge the existing worker candidate as-is. Review identified a contract bypass risk: caller-controlled retention/page values must not extend visible history beyond the exact 30-day retention contract or exceed bounded pagination/read limits. Add RED tests for oversized requests first, enforce server-side retention/bounds, then prove tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker capacity

Four SentinelX hosts are connected and the Free plan currently permits one active host. `pv-primary` is executable. A fresh execution attempt on `pv-worker-main` at 2026-09-01T10:13Z returned `upgrade_required` while the worker remained connected/healthy. Do not use Production as a development or PostgreSQL test worker. Persistent worker reports remain historical until a worker becomes executable.

To resume parallel workers, disconnect unused connected hosts so a worker becomes active or change the SentinelX host limit.

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

1. Recover/reconcile Task13 against exact schema20 main and publish only after its complete exact-session-kill gates pass, including real HTTP1+HTTP2 rehearsal.
2. Repair Task16 retention/pagination bypass with RED tests first and rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before schema21 publication.
3. Keep draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
4. Keep canonical docs/evidence synchronized only from verified GitHub and Production truth.
