# CONTINUE HERE — PVNaive

Last updated: 2026-09-02

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `6e7e14aabe22719099443d80d65798bb53b2769c`.
- Draft Task13 PR #64 exact head: `c79f5385b7a751c30282948423b8c34d1ba89deb`.
- Task13 compare: 44 ahead / 0 behind; merge-base = exact current main.
- Exact-head Task13 CI `33612706915`, Exact Accounting `33612706979`, and Pinned Forwardproxy `33612706951`: all **SUCCESS**.
- Production remains on Task15/schema20; no Task13/schema21 code is deployed.
- Fresh Production probe: all four services active; `/api/v1/health/ready` = `db=ok/schema=ok/ready=true/status=ready`; `/api/v1/health/live` = `service=pvnaive-api/status=ok`; 45-minute critical journal scan = zero matches.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.
- Task16 draft PR #81 / issue #79 has genuine pre-implementation RED evidence.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft #64 / all exact-head GitHub gates green / final real live proof pending**.
- Task16 bounded session/IP history: **IN PROGRESS / draft #81 / issue #79 / schema21 RED established**.

## Task13 next sequence

1. Keep #64 draft.
2. On an executable development Worker, run a fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal on `c79f5385...` proving target-only kill, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload and exactly-once final accounting.
3. `TrPaqet` owns this rehearsal but currently returns `upgrade_required` under the one-active-host SentinelX plan while `pv-primary` remains Production-only.
4. Merge only after the live rehearsal is green.
5. Deploy only after fresh encrypted Production backup + rollback state, exact artifact verification and postflight checks.

## Task16

Continue from issue #79 and draft PR #81. Important correction: current CI syntax-checks `tests/db/ip_session_history_contract_test.sh` but does not invoke it in the PostgreSQL18 database test list, so generic PR #81 green CI is not Task16 GREEN proof.

Next slice: wire a real schema21 PG18 integration test into CI; implement the minimal up/down pair preserving exact 30-day retention, maximum 500 server-side reads, tenant forced-RLS, trusted `direct_naive_accounting_session_peers` lineage, final-accounting-safe purge, coherent `SHA256SUMS`, and disposable rollback. Do not accept client-provided IP/session facts.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain S04-era (2026-08-27) and contain no current Task13/Task16 completion.

## Worker access

- `pv-primary`: executable, **Production-only**.
- `TrPaqet`: connected; Task13 final rehearsal assignment; currently plan-blocked.
- `pv-worker-main`: Task16 schema21 TDD/PG18 assignment when executable.

Never use Production as a development database or test lane.