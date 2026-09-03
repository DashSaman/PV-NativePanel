# PVNaive — Canonical Handoff

Last updated: 2026-09-04

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- No workflow run or status row is currently visible for this exact docs-only main commit.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head gates and focused tests green, real HTTP1/HTTP2 proof pending.
- Task16: draft #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; dedicated PG18/Exact Accounting/Pinned Forwardproxy green, generic CI still blocked by the schema21-aware RLS expectation.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Production state

Fresh read-only evidence on `pv-primary` at 2026-09-03 23:11 UTC:

- `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`: active;
- `/api/v1/health/ready` HTTP 200: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `/api/v1/health/live` HTTP 200: `service=pvnaive-api`, `status=ok`;
- local HTTPS probe returned TLS alert `internal error` / HTTP 000; external HTTPS health is not claimed.

The historical S04 startup blocker is not present in this fresh observation. No Production mutation, restart, reload, migration, DB write, credential rotation or deployment occurred.

## Gates and blockers

- Do not merge #64 until final live HTTP1/HTTP2 exact-session rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI passes with a narrowly-scoped schema21-aware RLS expectation; preserve schema20-specific Task15 fixtures and keep Task16 integration test explicitly wired into CI.
- Do not deploy without fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` on `pv-primary` remain historical S04-era ledgers last updated 2026-08-27 and do not override the fresh observations above.

## Worker capacity / assignments

- `pv-primary`: executable; Production-only.
- `TrPaqet`: connected but inactive; Task13 rehearsal when executable.
- `pv-worker-main`: connected but inactive; Task16 CI fix and rerun when executable.

No development/testing on Production. Keep truthful accounting/session semantics under retry, race, kill and disconnect.
