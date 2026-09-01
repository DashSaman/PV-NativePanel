# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `5f7fd64f34d951ec7f9c16907123ddd659484515`.
- Exact-main push CI `33548065775`: **SUCCESS**.
- Open roadmap PR: draft #64, exact head `8bb5804545cd977ec1e01f6331d40d1aa9148279`.
- Exact-head workflows were restarted after the socket-permission publication. WS1 Exact Accounting is **SUCCESS**; CI and WS1 Pinned Forwardproxy were still running at this checkpoint, so prior-head greens are not reused.
- Production remains on Task15/schema20; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

A fresh read-only Production probe was attempted, but `pv-primary` is inactive under the SentinelX Free-plan one-active-host limit while development Worker `TrPaqet` owns the slot. The probe returned `upgrade_required`; therefore no new Production-health PASS is credited.

No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 now contains:

- exact full-tuple registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- live CONNECT registration after accounting open + trusted peer recording, with teardown unregister;
- Unix-domain control server with bounded/strict JSON, exact tuple control, stale-socket handling and reload-safe shared lifecycle;
- ownership-checked API route selecting the exact tuple from trusted customer/RLS active-session state; client tuple fields are not accepted and credential mutation stays false;
- dedicated `pvnaive-session-control` system group, separated from telemetry;
- API service gets only the dedicated group; Caddy retains telemetry plus the dedicated group;
- fresh foundation and same-schema deploy paths provision the group idempotently;
- session-control socket resolves/chowns to the dedicated GID, then applies `0660`, failing closed on permission setup errors.

Permission TDD/evidence on `TrPaqet`: initial contract RED, forwardproxy group-assignment RED, permission contract PASS, patched pinned forwardproxy `go test -race ./...` PASS, focused API/session race tests PASS, full `go test ./...` PASS, and `git diff --check` PASS.

PR #64 must remain draft until stronger handler-level ownership/IDOR/CSRF failure tests, UI exact-session action, release packaging/install/rollback proof, full final-tree gates, and a fresh HTTP1/HTTP2 rehearsal prove target-only termination, sibling survival, forged tuple rejection, repeated-request idempotency, credential survival and exactly-once normal final accounting.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: three connected hosts, `TrPaqet`, `pv-worker-main`, and `pv-primary`; Free plan allows one active host.

- `TrPaqet`: active development Worker; assigned Task13 handler/UI/release-rehearsal lane.
- `pv-primary`: Production-only; current fresh execution is blocked by `upgrade_required`.
- `pv-worker-main`: connected/healthy, but fresh execution also returns `upgrade_required`.
- Task16: assigned to the next independently executable development Worker.

Human action is required only for true parallelism or simultaneous fresh Production probing: switch/disconnect hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.

## Exact next sequence

1. Keep #64 draft while exact-head checks complete.
2. Continue Task13 on `TrPaqet`: handler-level authorization/ownership/CSRF/IDOR failure tests, UI action, then release packaging/rollback proof.
3. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 rehearsal.
4. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and Production postflight access.
5. Execute Task16 on the next independent Worker when capacity exists.
