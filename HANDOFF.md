# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `a29b5ef434a72004af80cf489f47fffe0b0a03a8`, verified merge of PR #61 (documentation/state reconciliation only).
- Open roadmap PR: draft #62, `lead/task13-reconstruct-a29b5ef`, exact published head `2e0f485d61f2dd70647b6f626b1f8a18178336d7`.
- Old draft #4 is not current roadmap work.
- Production schema: **20**.
- Production runtime source remains Task15 commit `26aa74dddfd23535e45837f21531cf67ea2fd238`; later documentation merges did not change runtime artifacts.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #62 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

Fresh read-only verification in the current run succeeded after the SentinelX active slot moved to `pv-primary`: all four core services are active; `/api/v1/health/ready` reports `ready=true`, `db=ok`, `schema=ok`; the API process expects schema20; `/opt/pvnaive/deploy-src` is clean at exact Task15 source `26aa74dddfd23535e45837f21531cf67ea2fd238`; no `panic`, `fatal` or `schema mismatch` match appeared in the checked last 30 minutes of the four core-service journals. Exact Task15 rollout evidence remains `ops/evidence/TASK15-20260901-schema20-production-pass.md` and retained backup/rollback material remains under `/var/backups/pvnaive`.

No Production mutation, restart, reload or migration was performed in this run.

## Task13

Do not merge `lead/task13-kill-session-publish-20260901` wholesale. Current reconstruction is draft PR #62 on schema20 main.

Published/verified so far:

- exact full-tuple session registry: runtime credential + node + boot + session;
- sibling survival, forged-tuple rejection, repeated-kill idempotence, unregister semantics;
- local session-control protocol/client;
- initial CI formatting-only failure isolated and corrected with gofmt;
- new TDD-first local control handler: RED observed with missing `NewHandler`, then minimal implementation added;
- fresh isolated Worker verification with clean Go 1.25.0: focused handler tests PASS, race tests for `sessionkill` + `sessioncontrol` PASS, full `go test ./... -count=1` PASS, `git diff --check` PASS;
- incomplete tuple requests are rejected before live state is touched;
- exact head `2e0f485...`: Exact Accounting and Pinned Forwardproxy SUCCESS; CI database/web/go/rehearsal SUCCESS, bundle still in progress at the latest check.

PR #62 must stay draft until forwardproxy registration/cancel path and Unix listener ownership, API RBAC/ownership/CSRF, UI action, exact final accounting, full exact-head gates and a fresh real HTTP1/HTTP2 kill rehearsal are green.

Mandatory invariants remain: exact tuple identity, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting.

## Task16

A clean exact-main inspection was completed before the development Worker slot moved. Do not merge any stale candidate as-is. Caller-controlled retention/page values must not extend visible history beyond the exact 30-day retention contract or exceed bounded pagination/read limits. Add RED tests for oversized requests first, enforce server-side retention/bounds, then prove tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker capacity

`TrPaqet` was executable earlier in this run and used for Task13 verification, then disconnected while the next overlay RED test was being staged. Latest host listing returned three connected hosts (`pv-worker-main`, `pv-primary`, `ubuntu-4gb-hel1-1`), although inactive-worker plan responses still reported four connected. The executable slot is currently `pv-primary`; tested development workers return `upgrade_required`. Production must not be used as a development/test worker.

True parallelism still requires human action: disconnect unused connected hosts or change the SentinelX active-host limit.

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

1. Resume Task13 surgical integration on PR #62 when a development worker is executable: TDD-first forwardproxy/listener, API authorization and UI work.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and postflight.
4. When a second development worker becomes active, execute Task16 RED tests for retention >30 days and oversized pagination, then minimal server-side enforcement and PG18/RLS/rollback proof.
5. Keep draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
