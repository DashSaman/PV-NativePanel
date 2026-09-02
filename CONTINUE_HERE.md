# CONTINUE HERE — PVNaive

Last updated: 2026-09-02

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `c876b20343c6ae3aae27d096e9034955e88195c9`.
- Exact-main push CI `33578213894`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact published head `64acfd2593a19cf2048e45f8a63d9a1173ad8240`.
- Task13 is reconciled onto exact current main without force-push/history rewrite; fresh comparison reports **40 ahead / 0 behind** with merge base = exact current main.
- New CI / WS1 Exact Accounting / WS1 Pinned Forwardproxy runs are executing for exact head `64acfd2...`; do not reuse old-head greens.
- Task13 validated scope includes exact tuple/live CONNECT registration, Unix control lifecycle, dedicated `pvnaive-session-control` socket permissions, trusted-tuple DELETE API, per-session Web/UI kill with no credential mutation, R1 patched-Caddy packaging/rollback, and PostgreSQL18 auth/tenant proof.
- Old draft #4 remains outside the current roadmap.
- No Task13 runtime/schema change has been deployed; Production remains on Task15/schema20.
- Fresh Production read-only probe at `2026-09-02T02:11:55Z`: all four PVNaive/Caddy services active; readiness `db=ok`, `schema=ok`, `ready=true`; liveness `status=ok`; inspected previous 45-minute logs contain no panic/fatal/schema-mismatch match.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64 / current-main reconciled / final live proof pending**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / design gate pending**.

## Task13 next sequence

1. Keep #64 draft until all exact-head GitHub gates on `64acfd2...` are green.
2. On the first executable development Worker, run the fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal proving target-only kill, sibling survival, forged tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.
3. Merge only when both exact-head gates and the live rehearsal are green.
4. Deploy only with fresh Production access, encrypted backup, rollback state and postflight verification.

## Task16

No fresh current-main Task16 delta is credited. First step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are still S04-era (2026-08-27) and contain no current Task13/Task16 completion. Treat them as historical evidence only.

## Worker access

Three SentinelX hosts are connected: `TrPaqet`, `pv-worker-main`, `pv-primary`; Free plan permits one active host.

- `pv-primary`: active/executable, **Production-only**.
- `TrPaqet`: connected/healthy, fresh execution `upgrade_required`; assignment = Task13 final HTTP1/HTTP2 rehearsal.
- `pv-worker-main`: connected/healthy, fresh execution `upgrade_required`; assignment = Task16 RED retention/pagination lane.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.
