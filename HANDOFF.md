# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `b5f9466c1464afa9bc3183418aaf8e124b890563`.
- Exact-main push CI `33556193807`: **SUCCESS**.
- Open roadmap PR: draft #64, exact published head `5bc42d8dedd682eaf560a99777b21b9e82062c79`.
- Exact-head workflows on `5bc42d8...`: WS1 Exact Accounting `33557107038` is **SUCCESS**; CI `33557107036` and WS1 Pinned Forwardproxy `33557107045` are still running on the exact new head.
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
- dedicated `pvnaive-session-control` system group, separated from telemetry, with `0660` Caddy-owned socket;
- per-session Web/UI kill action that sends only path identifiers/CSRF, leaves credentials/subscriptions unchanged and refreshes the trusted list;
- **release proof:** R1 packages the exact reproducible pinned Task13 Caddy candidate, its provenance/build metadata and Caddy service drop-in; predeploy verifies checksums/provenance/config/module, mandatory encrypted backup precedes binary/drop-in mutation, activation performs exactly one controlled release-time Caddy binary-swap restart, and rollback restores/restarts the prior Caddy binary/drop-in state.

TDD/evidence on `TrPaqet`: release contract first failed RED because R1 lacked the Task13 Caddy candidate; after the implementation, `TASK13_R1_RELEASE_CONTRACT=PASSED`, `TASK13_SESSION_CONTROL_PERMISSIONS=PASSED`, focused API/session/ops race PASS, full Go PASS, reproducible pinned Caddy build PASS, and `TASK13_FORWARDPROXY_SESSION_CONTROL=PASSED`. Pinned candidate SHA: `0e44d42a63b5e1001b6c2410f6fa7108256aabb89dfd86cbb50334030bdddb0e`.

A new TDD-first DB/auth rehearsal is published on head `5bc42d8...`: CI-contract RED was observed before workflow wiring; local contract/Go/race/Caddy checks are green. Exact-head PostgreSQL18 CI now owns proof that missing-CSRF and cross-tenant IDOR attempts never reach the Unix control side effect and an owned kill emits exactly one trusted full tuple without credential mutation.

The prior exact head was fully green. On new head `5bc42d8...`, Exact Accounting is green while CI and Pinned Forwardproxy are still running; the new PostgreSQL18 DB/auth rehearsal is not credited until CI finishes successfully.

PR #64 must remain draft until the newly published DB-integrated ownership/IDOR/CSRF rehearsal passes exact-head PostgreSQL18 CI and a fresh HTTP1/HTTP2 rehearsal prove target-only termination, sibling survival, forged tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once normal final accounting.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: three connected hosts, `TrPaqet`, `pv-worker-main`, and `pv-primary`; Free plan allows one active host.

- `TrPaqet`: active development Worker; assigned Task13 DB-integrated authorization + final live-rehearsal lane.
- `pv-primary`: Production-only; current fresh execution is blocked by `upgrade_required`.
- `pv-worker-main`: connected/healthy but inactive under the same one-active-host limit; Task16 remains assigned there for the first independently executable slot.

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

1. Keep #64 draft while the new exact head is re-gated; require PostgreSQL18 ownership/IDOR/CSRF proof on exact head `5bc42d8...`.
2. Run fresh real HTTP1+HTTP2 exact-kill rehearsal with target-only termination, sibling survival and exactly-once accounting, and verify kill requests cause no Caddy lifecycle action.
3. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and Production postflight access.
4. Execute Task16 on `pv-worker-main` when an independent active slot exists.
