# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `3f775195439163a20935caff1bb4c2b8c3225c84` (PR #77).
- Exact-main push CI run `33602254294`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact head `e47fb53ca2885a0aee7ffc2fc8fc7a0b7d461ac7` after mechanical PR #78 reconciled current main without force-push/history rewrite.
- Exact-head Task13 workflows on `e47fb53c...`: CI `33602350141` **SUCCESS**; WS1 Exact Accounting `33602350122` **SUCCESS**; WS1 Pinned Forwardproxy `33602350340` **SUCCESS**.
- Task13 exact live-session kill: **IN PROGRESS / draft / final real HTTP1+HTTP2 live proof pending**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task16 bounded IP/session history: **IN PROGRESS / issue #79 / schema21 TDD RED established**.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 code has been deployed.

Fresh read-only observation at `2026-09-02T08:18:36Z` on `pv-primary`:

- `pvnaive-api`: **active**;
- `caddy-naive`: **active**;
- `pvnaive-runtime-agent`: **active**;
- `pvnaive-telemetry-agent`: **active**;
- `GET http://127.0.0.1:8080/api/v1/health/ready`: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: `service=pvnaive-api`, `status=ok`;
- 45-minute journal scan found no `panic`, `fatal`, `schema mismatch`, segmentation or error matches for the four services.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Task13 — exact live-session kill

PR #64 contains the validated exact-tuple registry/control/API/UI/release implementation. All three exact-head GitHub gates are green on `e47fb53c...`.

The remaining merge gate is a fresh real HTTP/1.1 + HTTP/2 rehearsal on the exact published tree proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy lifecycle action, and exactly-once final accounting. A focused exact-head Worker run was attempted this orchestration cycle, but the long race/build execution timed out and the Worker subsequently disconnected; it is **not** credited as PASS.

Only after that live proof may Task13 be merged. Production deployment additionally requires fresh encrypted backup + rollback snapshot and complete postflight.

## Task16 — bounded IP/session history / schema21

Issue #79 is the current execution ledger. On `pv-worker-main`, `tests/db/ip_session_history_contract_test.sh` was authored before implementation and produced the expected RED result `schema21 migration pair missing`. Required acceptance remains exact 30-day retention, hard server-side pagination/read limits, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 proof and rollback safety.

The Worker currently lacks GitHub write credentials, but connector-side GitHub writes remain available for validated publication.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override this file.

## Worker / orchestration capacity

Fresh SentinelX state currently shows two connected hosts: `pv-primary` and `TrPaqet`; `pv-worker-main` disconnected while an exact-head Task13 background run was executing.

- `pv-primary`: executable, **Production-only**.
- `TrPaqet`: connected; assignment = Task13 final HTTP1/HTTP2 rehearsal if it becomes executable under the plan limit.
- `pv-worker-main`: assignment on reconnect = Task16 schema21 TDD/PG18 lane.

Do not move development/testing onto Production.

## Immediate execution order

1. Keep #64 draft; exact-head CI is already fully green.
2. On the first executable development Worker, run the fresh real HTTP1+HTTP2 exact-kill rehearsal with exactly-once final accounting.
3. If and only if that proof passes, merge Task13, build the final R1 artifact, take fresh encrypted Production backup + rollback snapshot, deploy and postflight.
4. In parallel, continue Task16 from the established RED test through schema21 migration, bounded query/purge semantics, PG18 migration/rollback tests and publication through GitHub connector if direct Worker push remains unavailable.
