# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `5f7fd64f34d951ec7f9c16907123ddd659484515`.
- Exact-main push CI run `33548065775`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact head `8bb5804545cd977ec1e01f6331d40d1aa9148279`.
- Exact-head workflows were restarted for `8bb5804...`; WS1 Exact Accounting is **SUCCESS** and the other exact-head gates were still running at this checkpoint.
- New Task13 permission increment creates a dedicated `pvnaive-session-control` system group, distinct from `pvnaive-telemetry`; API gets only session-control access, while Caddy keeps telemetry plus the new group. The Caddy-owned `0660` socket resolves/chowns to the dedicated GID and fails closed if permission setup cannot be proven.
- Worker TDD/evidence passed: real RED permission failures before implementation, permission contract PASS, pinned forwardproxy race PASS, focused session/API race PASS, full Go PASS, diff check PASS.
- Old draft #4 remains outside the current roadmap.
- No Task13 runtime/schema change has been deployed; Production remains on Task15/schema20.
- Fresh Production probing was attempted but `pv-primary` returned `upgrade_required` while `TrPaqet` owns the single active SentinelX slot. Do not represent prior health as fresh.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**.

## Task13 current checkpoint

PR #64 contains exact-tuple primitives/live CONNECT registration, Unix-domain control server, reload-safe lifecycle, ownership-checked API kill, and a narrow dedicated socket permission model.

Next sequence: strengthen handler-level ownership/IDOR/CSRF failure tests; add UI exact-session kill; prove release packaging/install/rollback for the Caddy drop-in/binary and dedicated group; run full exact-tree gates; finally perform real HTTP/1.1 + HTTP/2 exact-kill rehearsal with exactly-once final accounting.

## Task16

No fresh current-main Task16 delta is credited. Keep it queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Worker access

Fresh SentinelX listing shows three connected hosts: `TrPaqet`, `pv-worker-main`, and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `TrPaqet`, Task13 development lane.
- `pv-primary`: Production-only, currently `upgrade_required` while the development Worker owns the slot.
- `pv-worker-main`: connected/healthy but also `upgrade_required`.
- No second executable Worker exists for Task16 in parallel.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Keep #64 draft while its exact-head checks complete and independent work continues on `TrPaqet`.
2. Add handler authorization/ownership/CSRF/IDOR failure tests, UI, and release packaging/rollback proof with TDD.
3. Run full exact-tree gates and fresh HTTP1/HTTP2 rehearsal.
4. Only then merge; deploy only with fresh Production backup/rollback/postflight access.
5. Put Task16 on the next independent Worker as soon as capacity exists.
