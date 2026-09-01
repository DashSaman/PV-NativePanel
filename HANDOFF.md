# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `8b8549b8d1431dfa8858c207e55b9a52eaa2c4e8`, verified merge of PR #60 (documentation/state reconciliation only).
- Exact-main CI run `33500952494`: **SUCCESS**.
- Open roadmap PRs: none. Old draft #4 is not current roadmap work.
- Production schema: **20**.
- Production runtime source remains the Task15 rollout commit `26aa74dddfd23535e45837f21531cf67ea2fd238`; later documentation merges did not change runtime artifacts.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS**, active-worker reconstruction on current schema20 main; full integration/publication and fresh real HTTP1/HTTP2 rehearsal still required.
- Task16: **IN PROGRESS**, schema21 candidate fails design acceptance until retention/pagination are server-bounded.

## Production state

The last successful fresh read-only verification confirmed all four services `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent` active and `/api/v1/health/ready` reporting `ready=true`, `db=ok`, `schema=ok`. Exact Task15 rollout evidence remains `ops/evidence/TASK15-20260901-schema20-production-pass.md`; fresh encrypted backup and rollback material remain retained under `/var/backups/pvnaive`.

A new read-only Production probe in the latest run could not execute because the one SentinelX active slot was assigned to another connected host. `pv-primary` remained connected/healthy but returned `upgrade_required`. No Production mutation, restart, reload or migration was attempted.

## Task13

Remote publication remains partial and stale. Do not merge `lead/task13-kill-session-publish-20260901` wholesale.

The currently executable worker is `ubuntu-4gb-hel1-1`. A clean clone was created there at exact main `8b8549b8...`. Historical Task13 primitives for the session-control wire/client and exact-tuple registry were reapplied locally on top of schema20. Focused `go test -race ./internal/sessionkill ./internal/sessioncontrol -count=1` passed, but this partial historical candidate still lacks a committed registry test and lacks the complete forwardproxy/API/UI integration. It is recovery evidence only, not mergeable work.

Continue with TDD-first surgical integration. Mandatory invariants: exact runtime-credential/node/boot/session tuple, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once final accounting. Then rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy gates before publication.

## Task16

Do not merge the existing worker candidate as-is. Caller-controlled retention/page values must not extend visible history beyond the exact 30-day retention contract or exceed bounded pagination/read limits. Add RED tests for oversized requests first, enforce server-side retention/bounds, then prove tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker capacity

Four SentinelX hosts are connected and the Free plan permits one active host. In the latest run `pv-primary`, `pv-worker-main`, and `host_311ff...` were inactive with `upgrade_required`; `ubuntu-4gb-hel1-1` was executable and used about 32% RAM. Use that worker for development while it remains active. Do not use Production as a development/test worker.

True parallelism still requires human action: disconnect unused connected hosts or change the SentinelX host limit.

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

1. Finish Task13 surgical integration on the active worker with RED tests first for registry/data-plane/API/UI behavior.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Publish only the exact verified Task13 tree, merge only after exact-head CI is green, and deploy only after fresh encrypted backup + rollback snapshot and postflight.
4. Repair Task16 retention/pagination bypass with RED tests first and rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before schema21 publication.
5. Keep draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
