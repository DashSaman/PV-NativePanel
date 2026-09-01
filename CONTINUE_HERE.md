# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `1aea1961515ab86428231499490202af4aef5e97`.
- Exact-main push CI run `33541904912`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact head `9a863258455473605a370f7ad4964043a0df92a1`.
- New exact-head workflows for the API publication are running; do not inherit prior-head green checks.
- Task13 `0003` lifecycle patch is byte-identical to the Worker-tested file (Git blob `b5889058caf4312df5508655193a6275c4ae5a1e`).
- New API increment: ownership-checked exact-session kill is wired from the authenticated customer/RLS read model; client-supplied tuple forgery is not accepted; credential mutation remains false. Worker RED→GREEN, focused race tests, full `go test ./...`, and `git diff --check` passed.
- Old draft #4 remains outside the current roadmap.
- No Task13 runtime/schema change has been deployed; Production remains on Task15/schema20.
- Fresh Production probing is currently blocked because the one SentinelX active slot is on `TrPaqet`; do not represent the prior health probe as fresh.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**.

## Task13 current checkpoint

PR #64 contains exact-tuple primitives/live CONNECT registration, Unix-domain control server, reload-safe lifecycle, and the first ownership-checked API kill increment. The API resolves the exact runtime-credential/node/boot/session tuple from the trusted active-session read model rather than accepting tuple fields from the request.

Worker gates passed for the new API increment: behavioral RED, focused route/tuple tests, `go test -race ./internal/httpapi ./internal/sessioncontrol ./internal/sessionkill -count=1`, full `go test ./... -count=1`, and `git diff --check`. Exact-head GitHub workflows are running separately and must be evaluated on this new head.

Next sequence: prove a narrow local service/group permission model for API access to the `0660` socket without broadening accounting access; then strengthen handler-level ownership/IDOR/CSRF failure tests; UI; full final-tree gates; finally real HTTP/1.1 + HTTP/2 exact-kill rehearsal with exactly-once final accounting.

## Task16

No fresh current-main Task16 delta is credited. Keep it queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Worker access

Fresh SentinelX listing shows two connected hosts: `TrPaqet` and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `TrPaqet`, development lane.
- `pv-primary`: Production-only and currently `upgrade_required` while the development Worker owns the slot.
- No second executable Worker exists for Task16 in parallel.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Keep #64 draft; wait only for its exact-head checks while continuing independent Task13 work on the Worker.
2. Prove socket permission/service-group boundary, then stronger handler-level authorization/ownership/CSRF/IDOR tests and UI with TDD.
3. Run full exact-tree gates and fresh HTTP1/HTTP2 rehearsal.
4. Only then merge; deploy only with fresh Production backup/rollback/postflight access.
5. Put Task16 on the next independent Worker as soon as capacity exists.
