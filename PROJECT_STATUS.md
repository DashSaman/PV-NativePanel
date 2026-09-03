# PVNaive — Canonical Project Status

Last updated: 2026-09-03 15:41 Asia/Tehran

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `a5d114c9cc74bcd0ef1ae9d27badbda3d493053b`.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub CI, Exact Accounting and Pinned Forwardproxy are green; final real HTTP/1.1 + HTTP/2 rehearsal remains the merge gate.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; dedicated gates are green; repository-wide CI remains blocked by `RLS coverage check failed: 43/42`.
- No workflow run is currently visible for `main` after the docs merge; do not infer post-merge CI.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation on `pv-primary`:

- `GET http://127.0.0.1:8080/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`;
- systemd active: `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`;
- direct local HTTPS probe returned TLS internal error / HTTP 000, so external HTTPS health is not claimed.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override the canonical files.

## Worker / orchestration capacity

All three hosts are connected, but the current plan exposes only one executable slot. In this run `pv-primary` was the executable slot and was used only for read-only Production verification.

- `pv-primary`: Production-only.
- `TrPaqet`: Task13 final HTTP1/HTTP2 rehearsal when an executable development slot is available.
- `pv-worker-main`: Task16 schema21 CI fix and full gate rerun when an executable development slot is available.

## Immediate execution order

1. Keep #64 and #81 draft.
2. On the first executable development Worker, run the final exact-head Task13 live rehearsal.
3. On the development lane, patch only the schema21-aware RLS expectation and rerun all Task16 gates.
4. If and only if Task13 live proof and Task16 full gates pass, perform fresh encrypted backup + rollback preparation and then consider promotion.
