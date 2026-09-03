# PVNaive — Canonical Project Status

Last updated: 2026-09-03 16:43 Asia/Tehran

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Main CI run `33623286003` = SUCCESS; combined status endpoint currently returns no status rows.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub CI, Exact Accounting and Pinned Forwardproxy are green; final real HTTP/1.1 + HTTP/2 rehearsal remains the sole merge gate.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy are green; repository-wide CI fails at `RLS coverage check failed: 43/42`.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-03 16:43 Asia/Tehran:

- `GET http://127.0.0.1:8080/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`;
- systemd active: `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`;
- local HTTPS probe to `https://127.0.0.1/api/v1/health/live` failed with TLS alert internal error (`HTTP 000`), so external HTTPS end-to-end health is not claimed.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Task13 — exact live-session kill

All exact-head GitHub gates are green on `3fc14825...`. The sole merge gate is a fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` is connected but inactive under the one-active-host SentinelX plan. Production must not be used as the rehearsal lane.

## Task16 — bounded IP/session history / schema21

Issue #79 and draft PR #81 are the active execution ledger. The dedicated PG18 Task16 workflow is green, but generic repository CI remains blocked by the RLS assertion that still expects 42 policies after schema21; the actual migration state exposes 43 covered relations. Do not bulk-change schema20-specific Task15 fixtures.

The required next change is a narrowly-scoped schema21-aware health assertion plus a fresh repository-wide CI run. The `tests/db/ip_session_history_contract_test.sh` integration test must remain explicitly executed in CI.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override the canonical files.

## Worker / orchestration capacity

- `pv-primary`: executable; **Production-only**.
- `TrPaqet`: connected but inactive; assignment = Task13 final HTTP1/HTTP2 rehearsal when executable.
- `pv-worker-main`: connected but inactive; assignment = Task16 schema21 CI fix and rerun when executable.

## Immediate execution order

1. Keep #64 and #81 draft.
2. On the first executable development Worker, run the final exact-head Task13 live rehearsal.
3. On the development lane, patch only the schema21-aware RLS expectation and rerun all Task16 gates.
4. Investigate the HTTPS TLS alert on `pv-primary` without changing Caddy/credentials until evidence is captured.
5. If and only if Task13 live proof and Task16 full gates pass, perform fresh encrypted backup + rollback preparation and then consider promotion.
