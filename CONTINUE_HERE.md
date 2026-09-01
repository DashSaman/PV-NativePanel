# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `392a153f14dc311f6d1dffe60e6b4d4da5f8cb17`.
- Prior exact-main CI on `cce50c4b...`: **SUCCESS**; no commit-associated workflow run has surfaced yet for the new documentation merge SHA, so do not invent one.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head gates: CI **SUCCESS**, Exact Accounting **SUCCESS**, Pinned Forwardproxy **SUCCESS**.
- Old draft #4 remains outside the current roadmap.
- No Task13 runtime/schema change has been deployed; Production remains on Task15/schema20.
- Fresh Production read-only probe: all four PVNaive services active and readiness on `127.0.0.1:8080` returned `db=ok`, `schema=ok`, `ready=true`.
- Journal scanning and deploy-source provenance could not be freshly re-proven due permissions; do not promote those missing observations into PASS.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**.

## Task13 current checkpoint

PR #64 contains exact-tuple primitives/live CONNECT registration plus a tested Unix-domain control-server primitive. Its exact published head has all three GitHub gates green.

Validated behavior includes strict/bounded local JSON, exact tuple kill, `0660` socket mode, stale-socket handling, owned-socket cleanup, sibling survival, forged-tuple rejection, idempotent repeat kill, unregister semantics, and preservation of the existing accounting teardown path.

Next TDD sequence: wire the Caddy-owned listener lifecycle so reload/unload cannot steal/delete a successor socket; prove a narrow local group/permission model for API access without broadening accounting access; then API RBAC/ownership/IDOR/CSRF; UI; full gates; finally real HTTP/1.1 + HTTP/2 exact-kill rehearsal with exactly-once final accounting.

## Task16

No fresh current-main Task16 delta is credited. Keep it queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Worker access

Fresh SentinelX listing shows two connected hosts: `TrPaqet` and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `pv-primary`, used only for read-only Production verification.
- `TrPaqet`: connected/healthy but currently `upgrade_required`, so development is blocked while Primary owns the slot.
- No second executable Worker exists for Task16 in parallel.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Keep #64 draft even though its exact-head gates are green; lifecycle/API/UI/rehearsal remain incomplete.
2. Switch/free the development slot back to `TrPaqet` when continuing implementation, without using Production as a development lane.
3. Continue Task13: reload-safe Caddy listener lifecycle + narrow socket permission proof, then API authorization boundary and UI with TDD.
4. Run full exact-tree gates and fresh HTTP1/HTTP2 rehearsal.
5. Only then merge; deploy only with fresh Production backup/rollback/postflight access.
6. Put Task16 on the next independent Worker as soon as capacity exists.