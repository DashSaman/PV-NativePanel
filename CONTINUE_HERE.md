# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `c6228290937a18c2dbe4ee06f966dc4636521d57` after merge of documentation PR #55.
- PR #55 head passed CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy before merge.
- Open roadmap PR: **#54 Task15 schema20**. Old draft PR #4 is unrelated to current roadmap.
- Production: schema **19**, deployed source `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`.
- Production services `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: active in the fresh read-only audit.
- Canonical API health routes are `/api/v1/health/live` and `/api/v1/health/ready`, both HTTP 200. Legacy `/healthz` and `/readyz` return 404 on the deployed binary and must not be used as current health evidence.
- DB health script with the Production DB environment returns `PVNAIVE_DB_HEALTH=OK`, schema19, and secret/MFA direct SELECT denial.
- No Production mutation was made in the latest audit.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / complete current-main source still not republished/reverified**.
- Task15 simultaneous unique-IP limit: **IN PROGRESS / PR #54 / schema20 / not merged or deployed**.
- Task16 bounded session/IP history: **IN PROGRESS behind schema20 / not merged or deployed**.

## Task15 exact current blocker

Original PR #54 head `31fd2caf...` passed Go/Web and WS1 Pinned Forwardproxy but failed PostgreSQL18 CI before reaching the unique-IP migration test because `tests/stages/S04_db_env_version_test.sh` still declared schema20 unsupported.

The branch was advanced to `ece028cb9122131f0b362474609ddd9f69701ced` with the schema-version contract corrected: schema20 is accepted and schema21 rejected. Wait for exact-head CI and inspect the next reproduced PostgreSQL18 failure. Prior isolated PG18 debugging found malformed UUID fixtures and an incorrect `first_connected_at` expectation for an `on_creation` term; fix those only if the exact current head reproduces them. Do not weaken accounting semantics to satisfy a fixture.

Required merge gates for Task15: **CI + WS1 Exact Accounting + WS1 Pinned Forwardproxy all green on the exact published head**.

## Task13

Do not merge the partial publication branch. Preserve exact tuple identity (runtime credential + node + boot + session), sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once final accounting. Reconstruct/recover missing current-main integration and rerun the previous full Go/race/Web/HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16

Keep schema21 behind stable schema20. Required proof: exact 30-day retention, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, bounded pagination/read paths, coherent migration/checksums and safe rollback.

## Worker access

Three SentinelX hosts are connected but only one is active under the current plan. `pv-primary` is presently executable; both worker hosts return `upgrade_required`. Therefore persistent worker reports cannot be freshly inspected and no new parallel worker process can be started until a worker host becomes active. Repository and CI work can continue independently.

## Production deployment rules

No schema20/21 deployment until the exact change is merged into eligible `main`. Before every Production mutation: fresh encrypted DB backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Follow PR #54 exact-head CI and fix only reproduced PG18 defects.
2. Merge Task15 only when all three required gates are green.
3. Run exact-main CI after merge.
4. Create and verify fresh Production backup/rollback state before any schema20 deployment.
5. Resume Task13/Task16 worker-backed reconciliation as soon as worker execution is available; do not invent worker progress while inaccessible.
6. Update `PROJECT_STATUS.md`, evidence and handoff only from verified repository/Production truth.
