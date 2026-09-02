# CONTINUE HERE — PVNaive

Last updated: 2026-09-02

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `3f775195439163a20935caff1bb4c2b8c3225c84`.
- Exact-main push CI `33602254294`: **SUCCESS**.
- Draft Task13 PR #64 exact head: `e47fb53ca2885a0aee7ffc2fc8fc7a0b7d461ac7` after mechanical PR #78 reconciled current main without force-push/history rewrite.
- Exact-head Task13 CI `33602350141`, Exact Accounting `33602350122`, and Pinned Forwardproxy `33602350340`: all **SUCCESS**.
- Production remains on Task15/schema20; no Task13 runtime/schema change is deployed.
- Fresh Production probe at `2026-09-02T08:18:36Z`: all four services active, readiness `db=ok/schema=ok/ready=true`, liveness `status=ok`; 45-minute journal scan found no panic/fatal/schema-mismatch/error matches.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.
- Task16 issue #79 has genuine TDD RED evidence: `tests/db/ip_session_history_contract_test.sh` fails because schema21 migration pair is absent, before implementation.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft #64 / exact-head CI green / final real live proof pending**.
- Task16 bounded session/IP history: **IN PROGRESS / issue #79 / schema21 RED established**.

## Task13 next sequence

1. Keep #64 draft despite all exact-head CI being green.
2. On an executable development Worker, run a fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal on `e47fb53c...` proving target-only kill, sibling survival, forged tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload and exactly-once final accounting.
3. The focused exact-head Worker run attempted this cycle timed out during race/build and the Worker then disconnected; do not count it as PASS.
4. Merge only after the live rehearsal is green.
5. Deploy only after fresh encrypted Production backup + rollback state, exact artifact verification and postflight checks.

## Task16

Continue from issue #79 and the existing RED test. Implement the minimal schema21 pair only after preserving exact 30-day retention, hard server-side pagination bounds, tenant RLS, trusted peer/accounting lineage, finalization synchronization and maintenance-only purge. Require PostgreSQL18 migration/rollback proof before promotion. If direct Worker push remains unavailable, publish only validated files through the GitHub connector.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain S04-era (2026-08-27) and contain no current Task13/Task16 completion.

## Worker access

Fresh SentinelX state currently shows `pv-primary` and `TrPaqet` connected; `pv-worker-main` disconnected during the exact-head Task13 long run.

- `pv-primary`: executable, **Production-only**.
- `TrPaqet`: assignment = Task13 final HTTP1/HTTP2 rehearsal if executable.
- `pv-worker-main`: on reconnect, assignment = Task16 schema21 TDD/PG18 lane.

Never use Production as a development database or test lane.
