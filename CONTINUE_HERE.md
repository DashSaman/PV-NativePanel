# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `62573fee8b88e4f951224da10e6a26d5b5838a54`, verified merge of PR #63 / documentation-state reconciliation.
- Exact-main CI run `33520634435`: **SUCCESS**.
- No runtime/schema change has landed on main since Task15; Production runtime remains the guarded schema20 Task15 rollout.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact head `263dc3e34982c739363822c7ac2cc643af7408c2`.
- PR #62 is closed as superseded because its branch diverged from current main; its validated seven-file Task13 primitive/control delta was republished cleanly in PR #64.
- Old draft #4 remains unrelated to the roadmap.
- Fresh current-run Production audit: all four PVNaive services active; readiness `ready=true`, `db=ok`, `schema=ok`; no Production mutation/restart/reload/migration was performed.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64**. Exact-tuple registry/client plus local control handler are published on current main; forwardproxy/Unix listener integration, API RBAC/ownership/CSRF, UI action, exact final accounting and real HTTP1+HTTP2 rehearsal remain.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**. Caller-controlled retention/pagination must not bypass exact 30-day retention or bounded reads.

## Worker access

Fresh SentinelX host listing shows three connected hosts: `TrPaqet`, `pv-worker-main`, `pv-primary`. Free-plan capacity permits only one active host. Fresh execution on both development workers returned `upgrade_required`; `pv-primary` is active and must remain Production-only for guarded/read-only operations. Do not use Production as a development or PostgreSQL test worker.

## Task13

Do not merge stale `lead/task13-kill-session-publish-20260901`, closed PR #62, or draft PR #64 while incomplete.

PR #64 starts directly from exact current main and carries exactly seven validated primitive/control files:

- exact runtime credential + node + boot + session registry;
- sibling survival, forged tuple rejection, idempotent repeated kill, unregister semantics;
- local session-control protocol/client;
- local control handler that rejects incomplete tuples before touching live state.

At latest observation, exact-head workflows for `263dc3e...` had started: CI and Exact Accounting queued, Pinned Forwardproxy in progress. Continue only after reviewing those results.

Next implementation sequence is TDD-first: forwardproxy registration/cancel wiring + Unix listener ownership; preserve Task14/15 response/finalization and trusted `RemoteAddr`/`ClientIP` unique-IP semantics; then API RBAC/ownership/CSRF and UI action. Final proof must include exact HTTP/1.1 + HTTP/2 kill, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival and exactly-once final accounting.

## Task16

Reconcile schema21 directly on current schema20 main. Required proof: exact server-enforced 30-day retention, hard-bounded pagination/read paths, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PG18 proof and safe rollback. Add RED tests for >30-day and oversized-page requests before implementation.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Review all exact-head gates on PR #64; keep draft while incomplete.
2. When a development worker becomes executable, continue Task13 TDD-first on PR #64.
3. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 exact-kill rehearsal.
4. Merge/deploy only the exact verified Task13 tree after fresh backup/rollback/postflight gates.
5. When another worker is available, begin Task16 RED tests and minimal server-side enforcement.
6. Keep `PROJECT_STATUS.md`, `HANDOFF.md` and evidence synchronized only from verified repository/Production truth.
