# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `8b8549b8d1431dfa8858c207e55b9a52eaa2c4e8`, verified merge of PR #60 / documentation-state reconciliation.
- Exact-main CI run `33500952494`: **SUCCESS**.
- Open roadmap PRs: none. Old draft #4 remains unrelated to current roadmap and awaits a real Karing client smoke.
- Production: schema **20**, deployed runtime source remains Task15 main `26aa74dddfd23535e45837f21531cf67ea2fd238`; later PRs changed documentation/evidence only.
- Last successful fresh Production check confirmed all four PVNaive services active and readiness `ready=true`, `db=ok`, `schema=ok`.
- A new read-only Production probe in the latest run was blocked by SentinelX `upgrade_required` because the single active slot had moved away from `pv-primary`; no Production mutation was attempted.
- Task15 fresh encrypted DB/config backup and rollback snapshot remain retained and verified from the guarded schema20 rollout.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / active-worker schema20 reconstruction underway; complete integration, exact publication and fresh real HTTP/1.1 + HTTP/2 rehearsal still required**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**. Caller-controlled retention/pagination must not permit bypassing the exact 30-day retention or bounded read contract; RED tests and server-side enforcement remain required.

## Worker access

Four SentinelX hosts are connected, while the Free plan permits one active host. In the latest run `pv-primary`, `pv-worker-main`, and `host_311ff...` returned `upgrade_required`, but `ubuntu-4gb-hel1-1` was executable with about 32% RAM used. A clean PVNaive repository was cloned there at exact main `8b8549b8...` and is now the development lane. Do not use Production as a development or PostgreSQL test worker.

## Task13

Do not merge the partial publication branch `lead/task13-kill-session-publish-20260901`. On the active worker, historical Task13 session-control protocol/client and exact-tuple registry primitives were reapplied on exact schema20 main. Focused `go test -race ./internal/sessionkill ./internal/sessioncontrol -count=1` passed, but the historical partial registry has no committed registry test and the data-plane/API/UI integration is still absent. Treat this only as recovered building blocks, not a completed candidate.

Preserve exact tuple identity (runtime credential + node + boot + session), sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once final accounting. Add RED tests before new integration code. Then rerun full Go/race/Web/HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16

Reconcile schema21 directly on top of current schema20 main. Required proof: exact server-enforced 30-day retention, hard-bounded pagination/read paths, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PG18 proof and safe rollback. Caller-supplied values must not extend history visibility beyond 30 days or bypass read bounds.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Continue Task13 surgical integration on `ubuntu-4gb-hel1-1` from exact main, TDD-first for registry/data-plane/API/UI behavior.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and a fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Publish only the exact verified Task13 tree, require green exact-head CI, then deploy with fresh backup/rollback and postflight if all gates remain green.
4. Fix Task16 retention/pagination bypass with RED tests first, then rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before any schema21 PR.
5. Keep `PROJECT_STATUS.md`, `HANDOFF.md` and evidence synchronized only from verified repository/Production truth.
