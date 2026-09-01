# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `d00c96be9dccbb1f5402a24a18beeef6910c78cd`, verified merge of PR #57 / Task15 Production evidence.
- Exact-main CI run `33476564850`: **SUCCESS**.
- Open PRs: old draft #4 only; unrelated to current roadmap.
- Production: schema **20**, deployed runtime source remains Task15 main `26aa74dddfd23535e45837f21531cf67ea2fd238`; PR #57 changed documentation/evidence only.
- Fresh read-only Production check on 2026-09-01: `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent` all active; `/api/v1/health/ready` returned `ready=true`, `db=ok`, `schema=ok`.
- Task15 fresh encrypted DB/config backup and rollback snapshot remain retained and verified from the guarded schema20 rollout.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / current-schema20 candidate requires complete publication plus fresh real HTTP/1.1 + HTTP/2 rehearsal before merge**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**. Review found a design blocker in the worker candidate: caller-controlled retention/pagination must not permit bypassing the exact 30-day retention or bounded read contract. Add RED tests for oversized retention/page requests and enforce server-side constants/clamps before acceptance.

## Worker access

Four SentinelX hosts are currently connected, but the Free plan permits only one active host. `pv-primary` is executable; `pv-worker-main` is connected/healthy but command execution returns `upgrade_required`. Do not run development or PostgreSQL proof on Production to work around this restriction. Persistent worker reports remain historical until a worker is executable again.

## Task13

Do not merge the partial publication branch. Preserve exact tuple identity (runtime credential + node + boot + session), sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once final accounting. Reconstruct/recover against current schema20 main and rerun full Go/race/Web/HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16

Reconcile schema21 directly on top of current schema20 main. Required proof: exact server-enforced 30-day retention, hard-bounded pagination/read paths, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PG18 proof and safe rollback. Caller-supplied values must not extend history visibility beyond 30 days or bypass read bounds.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Recover/reconcile Task13 against exact current schema20 main and republish only after all exact-session-kill gates pass, including fresh real HTTP/1.1 + HTTP/2 rehearsal.
2. Fix Task16 retention/pagination bypass with RED tests first, then rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before any schema21 PR.
3. Resume parallel worker execution when a worker host becomes active; do not invent worker progress while inaccessible.
4. Keep `PROJECT_STATUS.md`, `HANDOFF.md` and evidence synchronized only from verified repository/Production truth.
