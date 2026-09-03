# PVNaive — Canonical Project Status

Last updated: 2026-09-04

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- No workflow run and no status row are currently visible for this exact `main` commit; post-merge CI is therefore not proven for this docs-only head.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates and focused tests are green; final real HTTP/1.1 + HTTP/2 rehearsal remains the merge gate.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; dedicated PG18/Exact Accounting/Pinned Forwardproxy gates are green; repository-wide CI remains blocked by the schema21-aware RLS expectation and must be rerun on the exact head after the narrow fix.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-03 23:11 UTC:

- `pvnaive-api.service`: active (running), main PID `4014753`, since `2026-09-01 05:11:16 UTC`;
- `caddy-naive.service`: active;
- `pvnaive-runtime-agent.service`: active;
- `pvnaive-telemetry-agent.service`: active;
- `GET http://127.0.0.1:8080/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`;
- direct local HTTPS probe to `https://127.0.0.1/...` fails because SNI is not the production hostname (TLS alert `internal error`, HTTP 000); this is not an external-health result;
- SNI-correct local HTTPS probe using `--resolve namir.softarg.ir:443:127.0.0.1` returns HTTP 200 for `/api/v1/health/live` and HTTP 200 for `/`.

The historical S04 startup blocker is not present in this fresh read-only observation. No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed in this run.

## Task13 — exact live-session kill

All exact-head GitHub gates are green on `3fc14825...`. The sole merge gate is a fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` is connected but inactive under the one-active-host SentinelX plan. Production must not be used as the rehearsal lane.

## Task16 — bounded IP/session history / schema21

The dedicated PG18 Task16 gates are green on `b96c659...`, but generic repository CI is not green yet. The next change must be narrowly scoped to the schema21-aware RLS expectation; preserve schema20-specific Task15 fixtures and keep `tests/db/ip_session_history_contract_test.sh` explicitly executed in CI.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` on `pv-primary` remain historical S04-era ledgers last updated 2026-08-27. They do not override the fresh observations above and contain no newer Task13/Task16 completion.

## Worker / orchestration capacity

- `pv-primary`: executable; Production-only.
- `TrPaqet`: connected but inactive; assignment = Task13 final HTTP1/HTTP2 rehearsal when executable.
- `pv-worker-main`: connected but inactive; assignment = Task16 schema21 CI fix and rerun when executable.

## Immediate execution order

1. Keep #64 and #81 draft.
2. On the first executable development Worker, run the final exact-head Task13 live rehearsal.
3. On the development lane, patch only the schema21-aware RLS expectation and rerun all Task16 gates.
4. If and only if Task13 live proof and Task16 full gates pass, perform fresh encrypted backup + rollback preparation and then consider promotion.
