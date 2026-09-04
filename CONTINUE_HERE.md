# CONTINUE HERE — PVNaive

Last updated: 2026-09-05 02:38 Asia/Tehran

Start here after interruption. Re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before any mutation.

## Current verified state

- `main`: `eb212dfb40836a6bb91546e2bcc9f0cfc1afde7b` before this docs checkpoint; current canonical docs correction advances main again.
- No status rows or workflow runs are visible for the current docs-only `main` head; post-merge CI is not proven.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates green; live HTTP/1.1 + HTTP/2 rehearsal pending.
- Draft Task16 PR #81 actual GitHub head: `3c4310335ab4907d28bac995bba1be3545e14f6e`; body is stale. Dedicated Task16/Exact Accounting/Pinned Forwardproxy gates green; generic CI `33678134360` failed in `database`; failed jobs were re-run and result is pending.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh `pv-primary` read-only probe: all four services active; readiness/liveness bodies report ready/ok; correct SNI-preserving local HTTPS probe using `--resolve namir.softarg.ir:443:127.0.0.1` returned HTTP 200 and TLS verify result 0. The earlier HTTP 000 was caused by probing the IP with only a Host header.

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
- `pv-primary`: Production-only audit, encrypted backup, rollback and deploy lane; do not use it as a development database/test lane.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
