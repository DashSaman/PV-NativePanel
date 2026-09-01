# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `c6228290937a18c2dbe4ee06f966dc4636521d57` after PR #55.
- PR #55 was documentation-only and its exact head passed CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy before merge.
- Current roadmap PR: **#54 Task15 schema20**. Old draft #4 is not current roadmap work.
- Production schema: **19**.
- Production source: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS**, incomplete publication/recovery.
- Task15: **IN PROGRESS**, PR #54, not merged/deployed.
- Task16: **IN PROGRESS**, ordered behind schema20.

## Fresh Production state

`pv-primary` execution is available again. Fresh read-only verification shows `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, and `pvnaive-telemetry-agent.service` active. API listens on `127.0.0.1:8080`; Caddy listens on public 80/443.

Use the deployed canonical health endpoints:

- `GET /api/v1/health/live` → HTTP 200, status ok.
- `GET /api/v1/health/ready` → HTTP 200, DB/schema ready.

Legacy `/healthz` and `/readyz` return 404 on this deployed binary; that is route mismatch, not outage. DB health with `/etc/pvnaive/db.env` returns `PVNAIVE_DB_HEALTH=OK`, schema19, and direct secret/MFA SELECT denied.

No Production mutation occurred in the latest audit. Before any future deployment: fresh encrypted DB backup + rollback snapshot, verify both, apply intended exact migration/release only, then verify readiness, Runtime, Telemetry, Caddy/customer/accounting invariants and exact deployed provenance; rollback on any failed invariant.

## Task15 — current exact blocker / next action

PR #54 original head `31fd2caf...` passed Go/Web and WS1 Pinned Forwardproxy but failed PostgreSQL18 CI because `tests/stages/S04_db_env_version_test.sh` still declared schema20 unsupported. The branch is now at `ece028cb9122131f0b362474609ddd9f69701ced`, where schema20 is accepted and schema21 rejected.

Do not merge yet. Let exact-head CI expose the next real PG18 failure. Prior isolated PG18 debugging found malformed UUID fixtures and an invalid expectation that an `on_creation` term must synthesize `first_connected_at`; those are test defects only if reproduced on the exact current head. Preserve existing accounting semantics.

Required Task15 gates: exact-head **CI + WS1 Exact Accounting + WS1 Pinned Forwardproxy** all green. Required invariants: trusted Caddy `RemoteAddr` only, fail-closed before acceptance, ServiceTerm serialization, same-IP de-duplication, schema19 concurrent-session authority, and no peer/session leakage on rejected admission.

## Task13

Remote publication remains partial. Reconstruct/recover the complete current-main exact-session-kill integration and rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence. Exact tuple identity, sibling survival, no credential revoke, no Caddy reload/restart, idempotent kill, authorization isolation and exactly-once final accounting are mandatory.

## Task16

Keep schema21 behind stable schema20. Prove exact 30-day retention, tenant RLS, trusted peer/accounting facts, final-accounting synchronization, maintenance-only purge, bounded reads/pagination, checksum coherence and rollback safety before merge/DONE.

## Worker capacity

Three SentinelX hosts are connected but the Free plan permits one active host. `pv-primary` is currently active; worker hosts return `upgrade_required`, so persistent worker reports cannot be freshly inspected and no new worker process can be started. Repository/CI work continues independently. To resume parallel worker execution, make a worker host active by disconnecting an unused connected host or changing the plan.

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

1. Follow PR #54 exact-head PG18 CI and fix only reproduced defects.
2. Merge Task15 only after all three exact-head gates are green.
3. Run exact-main CI after merge.
4. Before schema20 Production deployment, create and verify fresh encrypted backup + rollback snapshot.
5. Resume Task13 and Task16 worker-backed reconciliation when worker execution returns; do not invent progress while hosts are inaccessible.
6. Keep canonical docs/evidence synchronized only from verified GitHub and Production truth.
