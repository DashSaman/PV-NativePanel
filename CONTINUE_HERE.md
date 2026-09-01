# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `a29b5ef434a72004af80cf489f47fffe0b0a03a8`, verified merge of PR #61 / documentation-state reconciliation.
- No runtime/schema change has landed on main since Task15; Production runtime provenance remains Task15 source `26aa74dddfd23535e45837f21531cf67ea2fd238`, schema20.
- Open roadmap work: draft PR #62 (`lead/task13-reconstruct-a29b5ef`), exact published head `2e0f485d61f2dd70647b6f626b1f8a18178336d7`. Old draft #4 remains unrelated to the roadmap.
- Last successful fresh Production check confirmed all four PVNaive services active and readiness `ready=true`, `db=ok`, `schema=ok`.
- Current-run read-only Production probe was blocked by SentinelX `upgrade_required`: `pv-primary` is connected/healthy but inactive because 4 hosts are connected under a one-active-host Free plan. No Production mutation was attempted.

## Task accounting

- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #62**. Exact-tuple registry/client plus a TDD-first local control handler are published; complete forwardproxy/Unix listener integration, API RBAC/ownership/CSRF, UI action, exact final accounting and real HTTP1+HTTP2 rehearsal are still required.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**. Caller-controlled retention/pagination must not bypass the exact 30-day retention or bounded read contract.

## Worker access

Four SentinelX hosts are connected while the Free plan permits one active host. In the current run `TrPaqet` (`host_311ff...`) is executable. A clean Go 1.25.0 toolchain was installed in an isolated worker path and used only for development verification. `pv-primary` remains connected/healthy but inactive; do not use Production as a development or PostgreSQL test worker.

## Task13

Do not merge the stale partial branch `lead/task13-kill-session-publish-20260901` and do not merge PR #62 while it is incomplete.

PR #62 currently contains:

- exact sessionkill registry semantics using runtime credential + node + boot + session tuple;
- sibling survival, forged tuple rejection, idempotent repeated kill and unregister tests;
- local session-control protocol/client;
- gofmt correction for the initial CI formatting-only failure;
- TDD-first local control HTTP handler. RED was explicitly observed as `undefined: NewHandler` before implementation; GREEN verification on the isolated Worker passed focused handler tests, race tests for `sessionkill` + `sessioncontrol`, full `go test ./... -count=1`, and `git diff --check`.

Exact-head CI for `2e0f485...` is running. Continue TDD-first with forwardproxy registration/cancel wiring + Unix socket listener, then API RBAC/ownership/CSRF and UI action. Preserve no credential revoke, no Caddy reload/restart, sibling survival, exact tuple identity, idempotent kill and exactly-once final accounting.

## Task16

A clean Task16 inspection workspace exists on the active worker at exact main. Reconcile schema21 directly on top of current schema20 main. Required proof: exact server-enforced 30-day retention, hard-bounded pagination/read paths, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PG18 proof and safe rollback. Caller-supplied values must not extend history visibility beyond 30 days or bypass read bounds.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify API readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Continue Task13 surgical integration on PR #62, TDD-first for forwardproxy/listener/API/UI behavior.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and a fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Publish/merge only the exact verified Task13 tree, then deploy only after fresh backup/rollback and postflight if all gates remain green.
4. When additional worker capacity is available, start Task16 with RED tests for >30-day retention and oversized pagination before server-side enforcement code.
5. Keep `PROJECT_STATUS.md`, `HANDOFF.md` and evidence synchronized only from verified repository/Production truth.
