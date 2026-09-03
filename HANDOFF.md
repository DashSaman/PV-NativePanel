# PVNaive — Canonical Handoff

Last updated: 2026-09-03 07:40 Asia/Tehran.

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`, GitHub gates green, real HTTP1/HTTP2 proof pending.
- Task16: draft #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated PG18, Exact Accounting and Pinned Forwardproxy green; generic CI database job fails on a latest-schema/RLS expectation mismatch.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.
- Post-merge CI for `a5d114c9...` is not currently observed; do not infer it from pre-merge runs.

## Production state

Fresh read-only evidence on `pv-primary` at 2026-09-03 04:10:56Z:

- `/api/v1/health/ready` HTTP 200: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `/api/v1/health/live` HTTP 200: `service=pvnaive-api`, `status=ok`;
- Docker inventory unavailable to SentinelX execution user because access to `/var/run/docker.sock` was denied.

No Production mutation, restart, reload, migration, DB write, credential rotation or deployment occurred.

## Gates and blockers

- Do not merge #64 until final live HTTP1/HTTP2 exact-session rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI passes with a narrowly-scoped schema21-aware latest-schema/RLS expectation; preserve schema20-specific Task15 fixtures and keep Task16 integration test explicitly wired into CI.
- Do not deploy without fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` on connected hosts remain historical S04-era ledgers and do not override these canonical facts.

## Worker capacity / assignments

- `pv-primary`: executable; Production-only.
- `TrPaqet`: connected but inactive; Task13 rehearsal when executable.
- `pv-worker-main`: connected but inactive; Task16 CI fix and rerun when executable.

No development/testing on Production. Keep truthful accounting/session semantics under retry, race, kill and disconnect.
