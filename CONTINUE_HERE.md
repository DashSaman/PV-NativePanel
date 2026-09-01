# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `cce50c4b198bd8ea449f385c1198fffbddfead8e`.
- Exact-main push CI run `33530087359`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head gates at latest observation: Exact Accounting **SUCCESS**, Pinned Forwardproxy **SUCCESS**, CI **IN PROGRESS**.
- Old draft #4 remains outside the current roadmap.
- No runtime/schema change from this run has been deployed; Production remains on the previously verified Task15/schema20 rollout.
- Fresh Production probing was attempted but blocked because `pv-primary` is non-active under the one-active-host SentinelX Free-plan limit. Do not restate prior health as fresh evidence.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**.

## Task13 current checkpoint

PR #64 now contains the earlier exact-tuple primitives/live CONNECT registration plus a tested Unix-domain control-server primitive.

Fresh RED→GREEN evidence on the active Worker:

- RED failed on missing `startPVNaiveSessionControlServer`;
- GREEN adds bounded/strict JSON handling, exact tuple kill, mode `0660`, stale-socket handling and owned-socket cleanup;
- a real HTTP request over the Unix socket proves exact target kill while sibling remains open;
- unregister-removes-tuple is covered;
- patched forwardproxy normal/race tests PASS;
- Task13 forwardproxy stage PASS;
- pinned forwardproxy boundary PASS;
- repository Go tests, focused race tests and `git diff --check` PASS.

The combined local reproducible-Caddy/full-gate command timed out and is not credited. Use exact GitHub workflows for published-head validation.

Next TDD sequence: wire the Caddy-owned listener lifecycle so reload/unload cannot steal/delete a successor socket; prove the narrow local group/permission model for API access without broadening accounting access; then API RBAC/ownership/IDOR/CSRF; UI; full gates; finally real HTTP/1.1 + HTTP/2 exact-kill rehearsal with exactly-once final accounting.

## Task16

Fresh persistent-workspace reconciliation found no new Task16 delta: its current run workspace is clean on an older main snapshot. Keep Task16 uncredited and queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination.

## Worker access

Fresh SentinelX listing shows two connected hosts: `TrPaqet` and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `TrPaqet` → assigned Task13.
- `pv-primary`: connected but `upgrade_required` on fresh command, so no fresh Production probe this run.
- No second executable Worker exists for Task16 in parallel.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Review remaining exact-head CI for PR #64; keep it draft regardless because functional integration is still incomplete.
2. Continue Task13 on `TrPaqet`: reload-safe Caddy listener lifecycle + narrow socket permission proof.
3. Implement API authorization boundary and UI with TDD.
4. Run full exact-tree gates and fresh HTTP1/HTTP2 rehearsal.
5. Only then merge; deploy only with fresh Production backup/rollback/postflight access.
6. Put Task16 on the next independent Worker as soon as capacity exists.
