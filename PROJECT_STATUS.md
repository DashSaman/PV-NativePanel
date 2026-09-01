# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main after verified documentation merge PR #55: `c6228290937a18c2dbe4ee06f966dc4636521d57`.
- PR #55 exact head `156f728f...` passed CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy before merge.
- Open roadmap PR: **#54 Task15 schema20**; old draft PR #4 remains unrelated and must not be merged as current work.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35: **DONE in main** (BUG-001/002/003 closed in repository truth).
- Task13 exact live-session kill: **IN PROGRESS / recovery + publication incomplete**.
- Task15 unique-IP limit: **IN PROGRESS / PR #54, not merged/deployed**.
- Task16 bounded IP/session history: **IN PROGRESS but integration remains ordered behind schema20**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Fresh Production truth

SentinelX execution access to `pv-primary` is currently available again. Fresh read-only checks on 2026-09-01 show:

- `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`: **active**;
- API listener: `127.0.0.1:8080`; Caddy: public `:80` / `:443`;
- canonical liveness endpoint `GET /api/v1/health/live`: HTTP **200**, service status `ok`;
- canonical readiness endpoint `GET /api/v1/health/ready`: HTTP **200**, DB/schema ready;
- DB health script with `/etc/pvnaive/db.env`: `PVNAIVE_DB_HEALTH=OK`, schema **19**, direct secret/MFA SELECT denied;
- release metadata source commit: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`;
- release metadata schema: **19**.

Important correction: `/healthz` and `/readyz` are not the canonical routes on this deployed binary and return 404. The current health routes are `/api/v1/health/live` and `/api/v1/health/ready`. A 404 on the legacy probes is not an outage.

No Production mutation was performed during this verification. Production remains schema19/source `0645b2e...`.

## Task13 — exact live-session kill

Remote publication branch `lead/task13-kill-session-publish-20260901` remains incomplete. Historical evidence proves a prior complete candidate passed strong exact-kill gates, but the complete current-main source must be recovered/reconstructed and reverified before merge. Invariants remain exact credential/node/boot/session identity, sibling survival, no credential revoke, no Caddy restart/reload, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting.

The persistent worker hosts are currently connected but unavailable for command execution under the SentinelX one-active-host limit while `pv-primary` is active. Do not claim fresh Task13 worker progress until access to a worker host is restored or the source is reconstructed through repository-backed evidence.

## Task15 — simultaneous unique-IP limit / schema20

PR #54 implements the schema20 candidate. Its original exact head `31fd2caf...` had Go and Web PASS and Pinned Forwardproxy PASS, but CI database and WS1 Exact Accounting PostgreSQL18 gates failed. The first exact failure was not the unique-IP SQL itself: `tests/stages/S04_db_env_version_test.sh` still treated schema20 as unsupported.

That contract is now corrected on PR #54 head `ece028cb9122131f0b362474609ddd9f69701ced`: schema20 is accepted and schema21 is rejected. This advances the gate to the next real PostgreSQL18 failure rather than weakening accounting semantics. PR #54 is still **NOT mergeable as DONE** until a fresh exact-head run proves CI + WS1 Exact Accounting + WS1 Pinned Forwardproxy all green.

Fresh PG18 debugging evidence also identified malformed UUID fixtures and an incorrect `first_connected_at` expectation for an `on_creation` service term. Those fixes must be applied only when reproduced on the current exact head; do not alter production accounting semantics to satisfy a bad fixture.

Task15 invariants remain trusted Caddy `RemoteAddr` only, fail-closed-before-acceptance, PostgreSQL service-term serialization, same-IP de-duplication, schema19 concurrent-session authority and no leaked peer/session rows on rejection.

## Task16 — bounded IP/session history / schema21

Task16 stays behind stable schema20. Required proof: exact 30-day retention, tenant-scoped RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, bounded pagination/read paths, coherent schema21 migration/checksums and safe rollback. No merge/deploy/DONE claim exists yet.

## Worker / orchestration capacity

Three SentinelX hosts are connected, but the Free plan permits one active host. `pv-primary` is the currently usable host; the two worker hosts return `upgrade_required`. Therefore no new worker-agent process was started in this run and existing persistent worker reports cannot be freshly inspected from those hosts. Repository/CI work continues independently. Human action is needed only if parallel worker execution is desired: disconnect an unused connected host so a worker can become active, or change the SentinelX plan.

## Immediate execution order

1. Let PR #54 rerun from head `ece028cb...`; inspect the next exact PostgreSQL18 failure and fix only reproduced test/implementation defects.
2. Merge Task15 only after exact-head CI + WS1 Exact Accounting + WS1 Pinned Forwardproxy are all green.
3. Do not deploy schema20 until it is merged into an eligible main and a fresh Production backup/rollback snapshot is created and verified.
4. Resume Task13 source recovery/reconstruction and Task16 reconciliation when worker access is available; continue repository-backed independent work meanwhile.
5. Keep old draft PR #4 out of the current roadmap unless a real Karing client smoke explicitly revives it.
6. After any eligible merge, run exact-main CI, then guarded Production deployment/postflight with rollback on any failed invariant.
