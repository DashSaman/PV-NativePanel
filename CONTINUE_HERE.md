# CONTINUE HERE — PVNaive

Last updated: 2026-09-05 01:37 Asia/Tehran

Start here after interruption. Re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before any mutation.

## Current verified state

- `main`: `287cb5880747d83c4d82aed3994ad37628bee782`.
- No status rows are visible for this exact docs-only head; post-merge CI is not proven.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates green; live HTTP/1.1 + HTTP/2 rehearsal pending.
- Draft Task16 PR #81 actual GitHub head: `3c4310335ab4907d28bac995bba1be3545e14f6e`; body is stale. Dedicated Task16/Exact Accounting/Pinned Forwardproxy gates green; generic CI `33678134360` failed in `database`; failed jobs were re-run and result is pending.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh `pv-primary` read-only probe: all four services active; readiness/liveness HTTP 200; SNI-correct local HTTPS HTTP 200.

## Task accounting

- Task12: DONE / Production, schema17.
- Task14: DONE / Production, schema19.
- Task15: DONE / Production, schema20.
- Task35 security P0: DONE in main.
- Task13: IN PROGRESS / draft #64 / live proof pending.
- Task16: IN PROGRESS / draft #81 / generic CI pending.

## Next execution

- Development lane: final exact-head Task13 live HTTP1/HTTP2 rehearsal when a development host is executable.
- Development lane: narrow schema21-aware generic RLS/latest-schema fixture fix and full gate rerun.
- `pv-primary`: Production-only audit, backup, rollback and deploy lane.

Never use Production as a development database or test lane; never claim completion from stale worker reports or partial evidence.
