# CONTINUE HERE — PVNaive

Last updated: 2026-09-01

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `6e58111665993e6e62c2d4e364a476d20ceb4896`.
- Exact-main push CI run `33550756339`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact published head `932a7f1b9f38c062559c870860959162901fb99b`.
- Exact-head workflows for `932a7f1b...`: CI `33554423163` **SUCCESS**, WS1 Exact Accounting `33554423088` **SUCCESS**, WS1 Pinned Forwardproxy `33554423118` **SUCCESS**.
- Task13 now contains exact tuple/live CONNECT registration, Unix control lifecycle, dedicated `pvnaive-session-control` socket permissions, trusted-tuple DELETE API, and per-session Web/UI kill with no credential mutation.
- A real release blocker was found and fixed TDD-first: R1 previously did not carry the patched Caddy binary/drop-in, so Task13 could not become live even from a green tree. R1 now packages the reproducible pinned Caddy candidate and provenance, validates it before mutation, backs up the existing Caddy binary/drop-in, activates the candidate with exactly one controlled release-time Caddy restart, and restores/restarts the prior Caddy on rollback.
- Local Worker proof: focused Go race PASS, full Go PASS, `TASK13_R1_RELEASE_CONTRACT=PASSED`, `TASK13_SESSION_CONTROL_PERMISSIONS=PASSED`, reproducible pinned Caddy build PASS, `TASK13_FORWARDPROXY_SESSION_CONTROL=PASSED`. Pinned candidate SHA: `0e44d42a63b5e1001b6c2410f6fa7108256aabb89dfd86cbb50334030bdddb0e`.
- GitHub CI for the exact head additionally passed PostgreSQL18 DB gates, Go, Web, rehearsal and R1 bundle jobs.
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

PR #64 contains the data-plane registry/control path, narrow socket permission model, ownership-checked API, UI exact-session action, and R1 packaging/install/rollback support for the patched reproducible Caddy candidate. All three exact-head GitHub workflows are green.

Next sequence: add PostgreSQL18 DB-integrated handler-level ownership/IDOR/CSRF failure-path proof; then perform a real HTTP/1.1 + HTTP/2 exact-kill rehearsal proving target-only kill, sibling survival, forged tuple rejection, repeated-kill idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once final accounting.

## Task16

No fresh current-main Task16 delta is credited. Keep it queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Worker access

Fresh SentinelX listing shows three connected hosts: `TrPaqet`, `pv-worker-main`, and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `TrPaqet`, Task13 development lane.
- `pv-primary`: Production-only, currently `upgrade_required` while the development Worker owns the slot.
- `pv-worker-main`: connected/healthy but inactive under the same limit; Task16 remains assigned there for the first independently executable slot.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Keep #64 draft despite green exact-head workflows; authorization and live protocol/accounting proof still gate merge.
2. Complete PostgreSQL18 ownership/IDOR/CSRF proof on `TrPaqet`.
3. Run the final HTTP1/HTTP2 exact-kill rehearsal with exactly-once accounting and no kill-triggered Caddy lifecycle action.
4. Only then merge; deploy only with fresh Production access, encrypted backup, rollback state and postflight verification.
5. Put Task16 on `pv-worker-main` as soon as an independent active slot exists.
