# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Open Task13 PR: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`.
- Exact-head Task13 CI `33623363327`, WS1 Exact Accounting `33623363299`, WS1 Pinned Forwardproxy `33623363389`: all **SUCCESS**.
- Fresh exact-head worker checks also PASS: diff-check; focused race tests for sessionkill/sessioncontrol/httpapi; forwardproxy session-control and permission contracts.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / #64 draft / exact-head GitHub + focused gates green / final real protocol proof pending**.
- Task16: **IN PROGRESS / draft #81 / issue #79 / current head `b96c65903e5fc314284ea777ceea236913a03842`**.

## Production state

Production Primary is connected but not executable in this cycle because SentinelX Free currently permits one active host and the active slot is on development worker `TrPaqet`. Therefore no new command-level Production health evidence is claimed here.

No Production mutation, restart, reload, migration, DB write, credential rotation or deployment occurred.

Before any future deploy, reactivate `pv-primary`, perform a fresh read-only health/journal/schema audit, verify exact artifact provenance, create a fresh encrypted backup and rollback snapshot, then mutate only after all release gates are green.

## Task13

The exact published tree `3fc14825...` has all three GitHub gates green and fresh focused worker validation green. The sole merge gate is still a fresh real HTTP1/HTTP2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` is currently executable and owns this proof. Do not use historical real-protocol evidence as a substitute. Do not merge #64 until this live proof completes.

## Task16

Issue #79 and draft PR #81 are the active ledger. Genuine RED preceded implementation. A dedicated PostgreSQL18 run on an earlier implementation head is green, but generic CI on head `a366c359...` failed the database job because multiple generic tests still assumed repository-latest schema20.

Four stale generic fixtures have now been corrected on the branch: auth refresh, DB health, backup/restore and periodic reset executor. Intentionally schema20-specific Task15 fixtures remain pinned. Current head `b96c659...` must obtain fresh exact-head normal CI, Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy success before promotion.

Task16 invariants: exact 30-day boundary; hard server-side limit 1..500; ENABLE + FORCE tenant RLS; trusted Caddy RemoteAddr/accounting lineage only; materialization from final+accounting-complete sessions; no client-provided IP/session truth; maintenance-only explicit confirmation-gated purge; rollback must not silently destroy retained evidence.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are historical S04-era ledgers last updated 2026-08-27. They contain no newer Task13/Task16 completion and must not override these canonical files.

## Worker capacity / assignments

- `TrPaqet`: currently executable; assignment = Task13 final live HTTP1/HTTP2 rehearsal; after completion, independent Task16 verification if the slot remains available.
- `pv-worker-main`: connected but inactive under one-active-host limit; assignment when executable = Task16 schema21 exact-head CI/PG18 lane.
- `pv-primary`: connected but inactive; **Production-only**; assignment when executable = fresh read-only audit, backup/rollback and gated deployment/postflight only.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.