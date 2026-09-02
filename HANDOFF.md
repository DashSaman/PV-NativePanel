# PVNaive — Canonical Handoff

Last updated: 2026-09-02

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `898536d8e44585fc696895bc74831b52416fba93`.
- Exact-main push CI `33560195019`: **SUCCESS**.
- Open roadmap PR: draft #64, exact published head `5bc42d8dedd682eaf560a99777b21b9e82062c79`.
- Exact-head workflows on `5bc42d8...`: CI `33557107036` **SUCCESS**, WS1 Exact Accounting `33557107038` **SUCCESS**, WS1 Pinned Forwardproxy `33557107045` **SUCCESS**.
- PR #64 is now diverged from current main: 38 commits ahead / 25 behind, merge base `62573fee8b88e4f951224da10e6a26d5b5838a54`, and GitHub currently reports `mergeable=false`. Its green old-base head is evidence, not a merge authorization.
- Production remains on Task15/schema20; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft / current-main reconciliation required**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

Fresh read-only Production evidence on `pv-primary`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` are all **active**;
- `GET /api/v1/health/ready` returns `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET /api/v1/health/live` returns `service=pvnaive-api`, `status=ok`;
- inspected recent journals produced no panic/fatal/schema-mismatch matches.

The SentinelX account cannot read `/proc/<api-pid>/environ`, so fresh process-env schema provenance was not established. No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 validated scope includes:

- exact full-tuple registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- live CONNECT registration after accounting open + trusted peer recording, with teardown unregister;
- Unix-domain control server with bounded/strict JSON, exact tuple control, stale-socket handling and reload-safe shared lifecycle;
- ownership-checked API route selecting the exact tuple from trusted customer/RLS active-session state; caller tuple fields are not accepted and credential mutation stays false;
- dedicated `pvnaive-session-control` system group, separated from telemetry, with `0660` Caddy-owned socket;
- per-session Web/UI kill action that sends only path identifiers/CSRF, leaves credentials/subscriptions unchanged and refreshes the trusted list;
- R1 packaging/install/rollback support for the exact reproducible pinned Task13 Caddy candidate and service drop-in;
- PostgreSQL18 DB/auth rehearsal proving CSRF rejection, cross-tenant IDOR rejection with zero socket side effects, trusted full-tuple owned kill, and credential survival.

The published exact head is green, but it is not based on current main and is now `mergeable=false`. Do not merge it as-is and do not force-push an unreviewed rewritten history.

Required Task13 sequence:

1. On an executable development Worker, reconstruct/cherry-pick the validated Task13 delta onto exact current `main`, resolving any changed-main semantics explicitly.
2. Rerun focused race/full Go, Web, release contracts, pinned Caddy proof, PostgreSQL18 rehearsal, then exact-head CI + Exact Accounting + Pinned Forwardproxy.
3. Run fresh real HTTP1/HTTP2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeated-request idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once normal final accounting.
4. Merge only that current-main verified tree; then deploy only after fresh encrypted backup + rollback snapshot and complete Production postflight.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: three connected hosts, `TrPaqet`, `pv-worker-main`, and `pv-primary`; Free plan allows one active host.

- `pv-primary`: currently active/executable; **Production-only**.
- `TrPaqet`: connected/healthy but fresh execution returns `upgrade_required`; first assignment is Task13 current-main reconciliation + final live rehearsal.
- `pv-worker-main`: connected/healthy but fresh execution returns `upgrade_required`; first independent assignment is Task16 RED retention/pagination lane.

Human action is required to resume development: switch/disconnect hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

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

1. Keep #64 draft while diverged/mergeable=false.
2. Make `TrPaqet` or another development host executable; republish/reconcile Task13 from exact current main and rerun all gates.
3. Complete the fresh HTTP1+HTTP2 exact-kill rehearsal with exactly-once final accounting and no kill-triggered Caddy lifecycle action.
4. Merge/deploy only after current-main verification and fresh Production backup/rollback/postflight proof.
5. Execute Task16 on `pv-worker-main` when an independent active slot exists.