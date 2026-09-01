# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `26aa74dddfd23535e45837f21531cf67ea2fd238`, verified merge of PR #54 / Task15.
- Exact-main CI run `33471780919`: **SUCCESS**.
- Open PRs: old draft #4 only; unrelated to current roadmap.
- Production: schema **20**, deployed source `26aa74dddfd23535e45837f21531cf67ea2fd238`.
- Production services `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: active.
- Canonical API health routes `/api/v1/health/live` and `/api/v1/health/ready`: HTTP 200.
- Public SNI-correct panel and API readiness probes: HTTP 200.
- Task15 fresh encrypted DB/config backup and rollback snapshot were created and verified before mutation.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / complete current-schema20 source still must be recovered/published/reverified**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / now unblocked by stable schema20 but not merged/deployed**.

## Task15 Production evidence

See `ops/evidence/TASK15-20260901-schema20-production-pass.md`.

Key facts:

- immutable DB release `0020-5df21d237574` promoted;
- `0020_unique_ip_limit.up.sql` applied exactly once;
- DB and expected schema both 20;
- API + telemetry + web + pinned accounting Caddy deployed from exact main;
- Caddy candidate validated before swap and received one explicit controlled service restart;
- postflight services, health, public panel/API, schema, release provenance and module checks passed;
- rollback material remains retained under `/var/backups/pvnaive`.

A direct backup attempt with `pvnaive_app` failed closed before any mutation because protected tables are intentionally unreadable. The canonical scheduled-backup path using local postgres role then succeeded and was verified. Do not weaken DB privileges to make backups work.

## Task13

Do not merge the partial publication branch. Preserve exact tuple identity (runtime credential + node + boot + session), sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once final accounting. Reconstruct/recover against current schema20 main and rerun full Go/race/Web/HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16

Reconcile schema21 directly on top of current schema20 main. Required proof: exact 30-day retention, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, bounded pagination/read paths, coherent migration/checksums, PG18 proof and safe rollback.

## Worker access

Three SentinelX hosts are connected but only one is active under the current plan. `pv-primary` is executable; worker hosts remain unavailable for command execution. Persistent worker reports cannot be freshly trusted until a worker host becomes executable.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Merge the Task15 Production evidence/docs reconciliation only after its documentation CI is green.
2. Recover/reconcile Task13 against exact current schema20 main and republish only after all exact-session-kill gates pass.
3. Rebase/reconcile Task16 onto schema20 and rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before any schema21 PR.
4. Resume parallel worker execution when a worker host becomes active; do not invent worker progress while inaccessible.
5. Keep `PROJECT_STATUS.md`, `HANDOFF.md` and evidence synchronized only from verified repository/Production truth.
