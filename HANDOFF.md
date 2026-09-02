# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `c876b20343c6ae3aae27d096e9034955e88195c9`.
- Exact-main push CI `33578213894`: **SUCCESS**.
- Open roadmap PR: draft #64, exact published head `64acfd2593a19cf2048e45f8a63d9a1173ad8240`.
- Task13 was reconciled again onto current main using a non-force two-parent merge. Fresh compare reports #64 **40 ahead / 0 behind**, merge base = exact current main; validated Task13 history was not rewritten.
- New exact-head CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy runs have started for `64acfd2...`; old-head green workflows remain historical evidence only.
- Production remains on Task15/schema20; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft / current-main reconciled / final live proof pending**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

Fresh read-only Production evidence on `pv-primary` at `2026-09-02T02:11:55Z`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` are all **active**;
- `GET /api/v1/health/ready` returns `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET /api/v1/health/live` returns `service=pvnaive-api`, `status=ok`;
- the previous 45-minute journal window produced no panic/fatal/schema-mismatch matches.

No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 validated scope includes the exact tuple registry/live CONNECT boundary, local Unix control path, narrow `pvnaive-session-control` permissions, ownership/RLS/CSRF-safe DELETE API, per-session Web/UI action, R1 pinned-Caddy packaging/rollback and PostgreSQL18 auth/tenant proof.

Remaining merge gates:

1. all three new exact-head GitHub workflows on `64acfd2...` must pass;
2. fresh real HTTP1/HTTP2 rehearsal must prove target-only termination, sibling survival, forged-tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once normal final accounting.

Only then may #64 leave draft and merge. Production deployment additionally requires fresh encrypted backup + rollback snapshot and complete postflight.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are historical S04-era ledgers last updated 2026-08-27. They contain no newer Task13/Task16 completion and must not override the canonical files above.

## Worker capacity / assignments

Fresh host listing: three connected hosts, `TrPaqet`, `pv-worker-main`, and `pv-primary`; Free plan allows one active host.

- `pv-primary`: active/executable; **Production-only**.
- `TrPaqet`: connected/healthy but fresh execution returns `upgrade_required`; assignment = Task13 final live HTTP1/HTTP2 rehearsal.
- `pv-worker-main`: connected/healthy but fresh execution returns `upgrade_required`; assignment = Task16 RED retention/pagination lane.

Human action is required to resume development execution: switch/disconnect hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.
