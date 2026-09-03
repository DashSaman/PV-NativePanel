# PVNaive — Canonical Project Status

Last updated: 2026-09-03 07:40 Asia/Tehran.

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub CI, Exact Accounting and Pinned Forwardproxy are green; final real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy are green; repository-wide CI fails only in database job. Latest recorded blocker is a schema21/latest-schema assertion mismatch; do not bulk-change schema20-specific Task15 fixtures.
- No workflow runs are currently associated with the post-merge `main` commit `a5d114c9...`; do not infer post-merge CI from pre-merge runs.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary` at 2026-09-03 04:10:56Z:

- `GET http://127.0.0.1:8080/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`;
- Docker inventory was not available to the SentinelX execution user due to Docker socket permission denied, so no container-state claim is made from that probe.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Task13 — exact live-session kill

All exact-head GitHub gates are green on `3fc14825...`. The sole merge gate is a fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` is connected but inactive under the one-active-host SentinelX plan. Production must not be used as the rehearsal lane.

## Task16 — bounded IP/session history / schema21

Issue #79 and draft PR #81 are the active execution ledger. The dedicated PG18 Task16 workflow, Exact Accounting and Pinned Forwardproxy are green on `3c431033...`; generic repository CI is red in the database job. The latest CI evidence shows a latest-schema assertion mismatch; keep schema20-specific Task15 fixtures pinned to schema20 and patch only the schema21-aware generic expectation.

The required next change is a narrowly-scoped schema21-aware assertion fix plus a fresh repository-wide CI run. Keep `tests/db/ip_session_history_contract_test.sh` explicitly executed in CI.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override the canonical files.

## Worker / orchestration capacity

- `pv-primary`: executable; **Production-only**.
- `TrPaqet`: connected but inactive; assignment = Task13 final HTTP1/HTTP2 rehearsal when executable.
- `pv-worker-main`: connected but inactive; assignment = Task16 schema21 CI fix and rerun when executable.

## Immediate execution order

1. Keep #64 and #81 draft.
2. On the first executable development Worker, run the final exact-head Task13 live rehearsal.
3. On the development lane, patch only the schema21-aware generic assertion and rerun all Task16 gates.
4. If and only if Task13 live proof and Task16 full gates pass, perform fresh encrypted backup + rollback preparation and then consider promotion.
