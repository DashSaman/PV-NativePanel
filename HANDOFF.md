# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `3f775195439163a20935caff1bb4c2b8c3225c84`.
- Exact-main push CI `33602254294`: **SUCCESS**.
- Open roadmap PR: draft #64, exact head `e47fb53ca2885a0aee7ffc2fc8fc7a0b7d461ac7` after PR #78 reconciled current main without force-push/history rewrite.
- Exact-head CI `33602350141`, WS1 Exact Accounting `33602350122`, WS1 Pinned Forwardproxy `33602350340`: all **SUCCESS**.
- Production remains on Task15/schema20; no Task13 code is deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / #64 draft / CI green / final real protocol proof pending**.
- Task16: **IN PROGRESS / issue #79 / schema21 RED established**.

## Production state

Fresh read-only evidence at `2026-09-02T08:18:36Z` on `pv-primary`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: **active**;
- `GET /api/v1/health/ready`: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET /api/v1/health/live`: `service=pvnaive-api`, `status=ok`;
- 45-minute journal scan: no panic/fatal/schema-mismatch/segmentation/error matches.

No Production mutation, restart, reload, migration, DB write, credential rotation or deployment occurred.

## Task13

PR #64 validated scope includes exact tuple registry/live CONNECT boundary, local Unix control path, narrow `pvnaive-session-control` permissions, ownership/RLS/CSRF-safe DELETE API, per-session Web/UI action, R1 pinned-Caddy packaging/rollback and PostgreSQL18 auth/tenant proof.

All exact-head GitHub gates are green. The sole merge gate is now a fresh real HTTP1/HTTP2 rehearsal on exact head `e47fb53c...` proving target-only termination, sibling survival, forged-tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once normal final accounting.

A focused exact-head Worker run was attempted this cycle; it reached the race/build stage but timed out and `pv-worker-main` disconnected afterward. This is not PASS evidence. Do not merge #64 until the real live proof is completed.

## Task16

Issue #79 is the active ledger. Genuine RED evidence exists on the Worker: `tests/db/ip_session_history_contract_test.sh` was written before implementation and fails because the schema21 migration pair is missing. Required implementation proof: exact 30-day retention, hard server-side pagination/read bounds, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 behavior and rollback safety.

Direct Worker push is still not trusted/available; use connector-side publication only after local validation if that remains true.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are historical S04-era ledgers last updated 2026-08-27. They contain no newer Task13/Task16 completion and must not override the canonical files.

## Worker capacity / assignments

Fresh SentinelX state currently shows two connected hosts: `pv-primary` and `TrPaqet`; `pv-worker-main` disconnected during the long Task13 exact-head run.

- `pv-primary`: executable; **Production-only**.
- `TrPaqet`: assignment = Task13 final live HTTP1/HTTP2 rehearsal if executable under current plan capacity.
- `pv-worker-main`: assignment on reconnect = Task16 schema21 TDD/PG18 lane.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.
