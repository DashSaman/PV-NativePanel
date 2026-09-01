# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `cce50c4b198bd8ea449f385c1198fffbddfead8e`.
- Exact-main push CI run `33530087359`: **SUCCESS**.
- Open roadmap PR: draft #64, exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head status at latest observation: Exact Accounting **SUCCESS**, Pinned Forwardproxy **SUCCESS**, CI **IN PROGRESS**.
- Old draft #4 is not current roadmap work.
- Production schema/runtime remains the previously verified Task15/schema20 rollout; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

This run attempted a fresh read-only probe on `pv-primary`, but SentinelX returned `upgrade_required` because the Free plan has one active slot and the active slot is currently the development Worker. Therefore prior schema20 health/provenance remains the latest valid Production evidence and must not be described as freshly re-verified in this run.

No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 now contains:

- exact full-tuple registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- live CONNECT registration after accounting open + trusted peer recording, with teardown unregister;
- a new Unix-domain control-server primitive: bounded/strict JSON, exact kill, `0660` socket mode, stale-socket handling and owned-socket cleanup.

Fresh TDD evidence for the latest increment: the test first failed specifically because `startPVNaiveSessionControlServer` was missing; after minimal implementation, patched upstream forwardproxy normal/race tests, Task13 stage, pinned boundary, repository Go tests, focused race tests and `git diff --check` all passed. A real HTTP request over the Unix socket killed only the selected target and left its sibling untouched.

Do not count the combined local reproducible-Caddy/full-gate attempt as PASS; it exceeded the Worker execution ceiling. Exact GitHub workflows are the authority for published head `485657d...`.

PR #64 must remain draft until all of the following are proven: reload-safe Caddy-owned listener lifecycle at `/run/pvnaive/session-control.sock`; narrow Unix-socket ownership/group model allowing the intended API but not widening accounting access; API RBAC/ownership/IDOR/CSRF without credential mutation; UI kill action; full exact-head gates; real HTTP1/HTTP2 target-only kill; sibling survival; forged tuple rejection; repeated-kill idempotency; credential survival; exactly-once normal final accounting.

## Task16

Persistent Worker reconciliation found no new completion. The current Task16 run workspace is clean on an older main snapshot, so no schema21 work receives credit this run.

First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: two connected hosts, `TrPaqet` and `pv-primary`; Free plan allows one active host.

- `TrPaqet`: active/executable; assigned to Task13 lifecycle/permission/API sequence.
- `pv-primary`: connected but non-active; fresh command returns `upgrade_required`; reserved for Production when a slot is available.
- Task16: assigned to the next independently executable development Worker; currently no second slot exists.

Human action is required only for true parallelism/fresh Production access while development continues: disconnect/switch hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

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

1. Keep #64 draft and finish its exact-head CI observation.
2. Continue Task13 TDD-first on `TrPaqet`: Caddy reload-safe listener lifecycle and narrow permission boundary.
3. Add API RBAC/ownership/IDOR/CSRF and UI action.
4. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 rehearsal.
5. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and Production postflight access.
6. Execute Task16 on the next independent Worker when capacity exists.
