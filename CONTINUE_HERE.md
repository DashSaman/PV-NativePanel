# CONTINUE HERE — PVNaive

Last updated: 2026-09-02

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Draft Task13 PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`.
- Task13 exact-head CI `33623363327`, Exact Accounting `33623363299`, Pinned Forwardproxy `33623363389`: all **SUCCESS**.
- Fresh exact-head worker checks also PASS: diff-check, focused race tests for sessionkill/sessioncontrol/httpapi, forwardproxy session-control and permission contracts.
- Task13 is **not DONE**: fresh real HTTP/1.1 + HTTP/2 live proof remains mandatory.
- Production remains on Task15/schema20; no Task13/schema21 code is deployed.
- No fresh Production command-level health claim this cycle because the one executable SentinelX slot is currently on development worker `TrPaqet`; `pv-primary` is connected but inactive.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.
- Task16 draft PR #81 current head: `b96c65903e5fc314284ea777ceea236913a03842`.
- Task16 has real pre-implementation RED and prior dedicated PostgreSQL18 GREEN evidence. Generic CI on earlier head `a366c359...` failed only the database job after exposing stale generic newest-schema=20 fixtures; four such fixtures have now been advanced to 21 without changing intentionally schema20-pinned Task15 fixtures.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft #64 / exact-head GitHub + focused gates green / final real live proof pending**.
- Task16 bounded session/IP history: **IN PROGRESS / draft #81 / issue #79 / schema21 implementation and final exact-head gates pending**.

## Task13 next sequence

1. Keep #64 draft.
2. On `TrPaqet`, run a fresh real HTTP/1.1 + HTTP/2 rehearsal on exact `3fc14825...` proving target-only kill, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload and exactly-once final accounting.
3. Historical protocol rehearsal evidence does not substitute for this exact-head proof.
4. Merge only after the live rehearsal is green.
5. Before deployment, activate `pv-primary`, perform a fresh read-only audit, create encrypted Production backup + rollback snapshot, verify exact R1 artifact, then deploy and postflight.

## Task16 next sequence

Continue from issue #79 and draft PR #81. Current head `b96c659...` contains schema21 implementation plus generic fixture corrections for auth refresh, DB health, backup/restore and periodic reset executor. Do not bulk-change intentionally frozen schema20 fixtures.

Final promotion requires exact-head normal CI, Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy all green. Keep exact 30-day retention, hard server-side maximum 500 reads, FORCE RLS, trusted `direct_naive_accounting_session_peers` lineage, final/accounting-complete materialization, explicit confirmation-gated maintenance purge, coherent checksums and rollback safety.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain S04-era (2026-08-27) and contain no current Task13/Task16 completion.

## Worker access

- `TrPaqet`: currently executable; Task13 final rehearsal owner.
- `pv-worker-main`: connected but inactive under one-active-host limit; Task16 CI/PG18 owner when executable.
- `pv-primary`: connected but inactive; **Production-only**, activate for fresh audit/deploy lane.

Never use Production as a development database or rehearsal host.