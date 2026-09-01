# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `62573fee8b88e4f951224da10e6a26d5b5838a54`, verified merge of PR #63 (documentation/state reconciliation only).
- Exact-main CI run `33520634435`: **SUCCESS**.
- Open roadmap PR: draft #64, `lead/task13-reconstruct-62573fee`, exact head `263dc3e34982c739363822c7ac2cc643af7408c2`.
- PR #62 is closed/superseded because its branch diverged from current main; only its validated seven-file Task13 delta was cleanly republished in #64.
- Old draft #4 is not current roadmap work.
- Production schema remains **20** and runtime remains the guarded Task15 rollout; no later repository merge changed Production runtime artifacts.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

Fresh read-only verification in the current run succeeded on `pv-primary`: all four core services are active and `/api/v1/health/ready` reports `ready=true`, `db=ok`, `schema=ok`. No Production mutation, restart, reload, migration, credential rotation or deployment was performed.

Previous verified Task15 schema20 provenance/backup evidence remains authoritative for details not freshly re-read in this run. Never upgrade an old observation into new evidence.

## Task13

Do not merge stale `lead/task13-kill-session-publish-20260901` or closed PR #62. Active reconstruction is draft PR #64, created directly from exact current main.

Published scope in #64:

- exact full-tuple session registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- local session-control HTTP handler that rejects incomplete tuples before touching live state.

The branch is 0 commits behind exact main at publication. Exact-head workflows for `263dc3e...` started immediately; at latest observation CI and Exact Accounting were queued and Pinned Forwardproxy was in progress.

PR #64 must stay draft until forwardproxy registration/cancel path + Unix listener ownership, API RBAC/ownership/CSRF, UI action, exact final accounting, full exact-head gates and a fresh real HTTP1/HTTP2 kill rehearsal are green.

Mandatory invariants: preserve Task14/15 response/finalization and trusted `RemoteAddr`/`ClientIP` unique-IP semantics; exact tuple identity; sibling survival; no credential revoke; no Caddy reload/restart; idempotent repeated kill; tenant/role isolation; exactly-once normal final accounting.

## Task16

Do not merge any stale schema21 candidate. Caller-controlled retention/page values must not extend visible history beyond exact 30 days or exceed bounded pagination/read limits. Add RED tests first, then minimal server-side enforcement, then prove tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker capacity

Fresh host listing shows three connected SentinelX hosts: `TrPaqet`, `pv-worker-main`, and `pv-primary`. The Free plan allows only one active host. Fresh execution attempts on both development workers returned `upgrade_required`; `pv-primary` is the active slot and is Production-only.

True parallelism therefore requires human action: disconnect unused connected hosts or increase the SentinelX active-host limit. Until then, do not move development/testing onto Production.

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

1. Review exact-head CI/Accounting/Pinned-Forwardproxy results on PR #64; keep it draft while incomplete.
2. As soon as a development worker becomes executable, continue Task13 TDD-first: forwardproxy/listener, then API authorization and UI.
3. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
4. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and postflight.
5. When another development worker is active, execute Task16 RED tests for retention >30 days and oversized pagination, then minimal enforcement and PG18/RLS/rollback proof.
6. Keep draft PR #4 outside the roadmap unless a real Karing client smoke explicitly revives it.
