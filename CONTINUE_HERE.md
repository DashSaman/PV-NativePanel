# CONTINUE HERE — PVNaive

Last updated: 2026-09-03 15:41 Asia/Tehran

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates green; real HTTP/1.1 + HTTP/2 rehearsal pending.
- Draft Task16 PR #81 exact head: `b96c65903e5fc314284ea777ceea236913a03842`; dedicated gates green; generic CI database job fails at `RLS coverage check failed: 43/42`.
- No post-merge workflow run is visible for the current `main` commit.
- Production remains on Task15/schema20; no Task13/schema21 code deployed.
- Fresh localhost Production probe: ready/live HTTP 200 and all four PVNaive services active; direct local HTTPS probe returned TLS internal error / HTTP 000.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Next execution

- `TrPaqet`: final exact-head Task13 live HTTP1/HTTP2 rehearsal when an executable development slot is available.
- `pv-worker-main`: narrowly-scoped schema21-aware RLS assertion fix and full gate rerun when an executable development slot is available.
- `pv-primary`: Production-only audit, backup, rollback and deploy lane.

Never use Production as a development database or test lane; never claim completion from stale worker reports or partial evidence.
