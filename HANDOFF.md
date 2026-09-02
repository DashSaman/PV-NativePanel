# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `6e7e14aabe22719099443d80d65798bb53b2769c`.
- Open Task13 PR: draft #64, exact head `c79f5385b7a751c30282948423b8c34d1ba89deb`.
- Fresh compare: 44 commits ahead / 0 behind; merge-base exact current main.
- Exact-head Task13 CI `33612706915`, WS1 Exact Accounting `33612706979`, WS1 Pinned Forwardproxy `33612706951`: all **SUCCESS**.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / #64 draft / exact-head GitHub gates green / final real protocol proof pending**.
- Task16: **IN PROGRESS / draft #81 / issue #79 / schema21 RED established**.

## Production state

Fresh read-only evidence this cycle on `pv-primary`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: **active**;
- `GET /api/v1/health/ready`: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET /api/v1/health/live`: `service=pvnaive-api`, `status=ok`;
- 45-minute critical journal scan: zero panic/fatal/schema-mismatch/segmentation-fault matches.

No Production mutation, restart, reload, migration, DB write, credential rotation or deployment occurred.

## Task13

All exact-head GitHub gates are green. The sole merge gate is a fresh real HTTP1/HTTP2 rehearsal on exact head `c79f5385...` proving target-only termination, sibling survival, forged-tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` owns this proof but currently returns `upgrade_required` while Production Primary owns the single SentinelX active-host slot. Do not substitute Production as the development/rehearsal host. Do not merge #64 until this live proof is completed.

## Task16

Issue #79 and draft PR #81 are the active ledger. Genuine RED evidence exists: `tests/db/ip_session_history_contract_test.sh` was written before implementation and failed because schema21 was absent.

Fresh CI review found that `.github/workflows/ci.yml` syntax-checks all DB shell tests but does not explicitly execute this Task16 test in the PostgreSQL18 database job. Consequently current generic PR #81 green CI is not Task16 GREEN proof.

Required implementation proof: exact 30-day retention, hard server-side read cap of 500, forced tenant RLS, trusted schema17 peer/accounting lineage only, final-accounting-safe maintenance purge, coherent schema21 checksums, explicit PostgreSQL18 test execution and disposable rollback safety.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are historical S04-era ledgers last updated 2026-08-27. They contain no newer Task13/Task16 completion and must not override the canonical files.

## Worker capacity / assignments

- `pv-primary`: executable; **Production-only**.
- `TrPaqet`: connected; assignment = Task13 final live HTTP1/HTTP2 rehearsal; currently plan-blocked.
- `pv-worker-main`: assignment when executable = Task16 schema21 TDD/PG18 lane.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.