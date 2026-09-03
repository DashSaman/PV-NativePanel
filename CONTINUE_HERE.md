# CONTINUE HERE — PVNaive

Last updated: 2026-09-03 07:40 Asia/Tehran.

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates green; real HTTP/1.1 + HTTP/2 rehearsal pending.
- Draft Task16 PR #81 exact head: `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated Task16/PG18, Exact Accounting and Pinned Forwardproxy green; generic CI database job fails on a latest-schema/RLS expectation mismatch.
- No post-merge workflow run is currently associated with `a5d114c9...`.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh Production read-only probe on `pv-primary`: ready/live HTTP 200. Docker inventory was unavailable to the execution user because of Docker socket permissions.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: DONE / Production, schema17.
- Task14: DONE / Production, schema19.
- Task15: DONE / Production, schema20.
- Task35 security P0: DONE in main.
- Task13: IN PROGRESS / draft #64 / live proof pending.
- Task16: IN PROGRESS / draft #81 / generic CI expectation pending.

## Next execution

- `TrPaqet`: final exact-head Task13 live HTTP1/HTTP2 rehearsal when active.
- `pv-worker-main`: narrowly-scoped schema21-aware RLS/latest-schema assertion fix and full gate rerun when active.
- `pv-primary`: Production-only audit, backup, rollback and deploy lane.

Never use Production as a development database or test lane; never claim completion from stale worker reports or partial evidence.
