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

The last successful fresh read-only verification confirmed `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` active and `/api/v1/health/ready` reporting `ready=true`, `db=ok`, `schema=ok`. Exact Task15 rollout evidence remains `ops/evidence/TASK15-20260901-schema20-production-pass.md`; fresh encrypted backup and rollback material from that rollout remain retained under `/var/backups/pvnaive`.

A current-run read-only Production probe could not execute because SentinelX currently has 4 connected hosts under a Free-plan one-active-host limit. `pv-primary` is connected/healthy but returned `upgrade_required`. No Production mutation, restart, reload or migration was attempted.

## Task13

Do not merge `lead/task13-kill-session-publish-20260901` wholesale. Current reconstruction is draft PR #62 on schema20 main.

Published/verified so far:

- exact full-tuple session registry: runtime credential + node + boot + session;
- sibling survival, forged-tuple rejection, repeated-kill idempotence, unregister semantics;
- local session-control protocol/client;
- initial CI formatting-only failure isolated and corrected with gofmt;
- new TDD-first local control handler: RED observed with missing `NewHandler`, then minimal implementation added;
- fresh isolated Worker verification with clean Go 1.25.0: focused handler tests PASS, race tests for `sessionkill` + `sessioncontrol` PASS, full `go test ./... -count=1` PASS, `git diff --check` PASS;
- incomplete tuple requests are rejected before live state is touched.

Exact-head CI for `2e0f485...` is running. PR #62 must stay draft until forwardproxy registration/cancel path and Unix listener ownership, API RBAC/ownership/CSRF, UI action, exact final accounting, full exact-head gates and a fresh real HTTP1/HTTP2 kill rehearsal are green.

Mandatory invariants remain: exact tuple identity, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting.

## Task16

A clean inspection workspace is available on the active Worker at exact main. Do not merge any stale candidate as-is. Caller-controlled retention/page values must not extend visible history beyond the exact 30-day retention contract or exceed bounded pagination/read limits. Add RED tests for oversized requests first, enforce server-side retention/bounds, then prove tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker capacity

Four SentinelX hosts are connected and the Free plan permits one active host. In this run `TrPaqet` (`host_311ff...`) is executable and used for Task13 verification and Task16 inspection. `pv-primary` is connected/healthy but inactive. Do not use Production as a development/test worker.

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

1. Continue Task13 surgical integration on PR #62 with TDD-first forwardproxy/listener, API authorization and UI work.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and postflight.
4. When a second worker becomes active, execute Task16 RED tests for retention >30 days and oversized pagination, then minimal server-side enforcement and PG18/RLS/rollback proof.
5. Keep draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
