# CONTINUE HERE — PVNaive

Last updated: 2026-09-04

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- No workflow run or status row is currently visible for this exact docs-only `main` commit.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates and focused tests green; real HTTP/1.1 + HTTP/2 rehearsal pending.
- Draft Task16 PR #81 exact head: `b96c65903e5fc314284ea777ceea236913a03842`; dedicated PG18/Exact Accounting/Pinned Forwardproxy green; generic repository CI still needs the narrow schema21-aware RLS expectation fix and rerun.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh `pv-primary` read-only probe: all four PVNaive services active; localhost readiness/liveness HTTP 200; local HTTPS probe fails with TLS alert `internal error` (HTTP 000), so external HTTPS health is not claimed.
- The historical S04 startup blocker is not present in the fresh probe.

## Task accounting

- Task12: DONE / Production, schema17.
- Task14: DONE / Production, schema19.
- Task15: DONE / Production, schema20.
- Task35 security P0: DONE in main.
- Task13: IN PROGRESS / draft #64 / live proof pending.
- Task16: IN PROGRESS / draft #81 / generic CI RLS expectation pending.

## Next execution

- `TrPaqet`: final exact-head Task13 live HTTP1/HTTP2 rehearsal when active.
- `pv-worker-main`: narrowly-scoped schema21-aware RLS assertion fix and full gate rerun when active.
- `pv-primary`: Production-only audit, backup, rollback and deploy lane.

Never use Production as a development database or test lane; never claim completion from stale worker reports or partial evidence.
