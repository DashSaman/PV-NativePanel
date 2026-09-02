# CONTINUE HERE — PVNaive

Last updated: 2026-09-02

Start here after interruption. Historical worker/stage notes are evidence only. Always re-read exact GitHub `main`, open PRs, exact-head CI and fresh Production health before mutation.

## Current verified state

- `main`: `274979910e1845b3918105a7043f982c6a0a6e78`.
- Exact-main push CI run `33570049441`: **SUCCESS**.
- Current roadmap work: draft PR #64 (`lead/task13-reconstruct-62573fee`), exact published head `5bc42d8dedd682eaf560a99777b21b9e82062c79`.
- Exact-head workflows for `5bc42d8...` are historical green evidence: CI `33557107036` **SUCCESS**, WS1 Exact Accounting `33557107038` **SUCCESS**, WS1 Pinned Forwardproxy `33557107045` **SUCCESS**.
- #64 is not directly integrable: relative to current main it is 38 commits ahead / 26 behind, merge base `62573fee8b88e4f951224da10e6a26d5b5838a54`, and GitHub reports `mergeable=false`. Do not merge or force-update it before current-main reconstruction and fresh revalidation.
- Task13 validated scope includes exact tuple/live CONNECT registration, Unix control lifecycle, dedicated `pvnaive-session-control` socket permissions, trusted-tuple DELETE API, per-session Web/UI kill with no credential mutation, R1 patched-Caddy packaging/rollback, and PostgreSQL18 auth/tenant proof.
- Old draft #4 remains outside the current roadmap.
- No Task13 runtime/schema change has been deployed; Production remains on Task15/schema20.
- Fresh Production read-only probe at `2026-09-02T00:13:29Z`: all four PVNaive/Caddy services are active; `/api/v1/health/ready` reports `db=ok`, `schema=ok`, `ready=true`; `/api/v1/health/live` reports `pvnaive-api` status `ok`; the inspected previous 30-minute logs contain no panic/fatal/schema-mismatch match.
- No Production mutation/restart/reload/migration/DB write/credential change occurred.

## Task accounting

- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 security P0: **DONE in main**.
- Task13 exact-session kill: **IN PROGRESS / draft PR #64 / current-main reconstruction required**.
- Task16 bounded session/IP history: **IN PROGRESS / schema21 / NOT mergeable yet**.

## Task13 current checkpoint

PR #64 contains the validated historical Task13 data-plane registry/control path, narrow socket permission model, ownership-checked API, UI exact-session action, release support and DB/auth proof. Its published head is green, but it is based on an old merge base and is now `mergeable=false` against current main.

Next sequence for Task13:

1. Reconstruct the validated Task13 delta onto exact current `main` on an executable development Worker; do not wholesale merge historical history and do not use Production as a development lane.
2. Re-run focused/full local verification and exact-head GitHub CI/Exact Accounting/Pinned Forwardproxy.
3. Perform the fresh real HTTP/1.1 + HTTP/2 exact-kill rehearsal proving target-only kill, sibling survival, forged tuple rejection, repeated-kill idempotency, credential survival, no kill-request-triggered Caddy restart/reload, and exactly-once final accounting.
4. Only then merge; deploy only with fresh Production access, encrypted backup, rollback state and postflight verification.

## Task16

No fresh current-main Task16 delta is credited. Keep it queued for the next independently executable Worker. Required first step remains RED tests proving callers cannot request >30 days or oversized pagination, followed by server-side enforcement and schema21/RLS/PG18/rollback proof.

## Worker access

Fresh SentinelX listing shows three connected hosts: `TrPaqet`, `pv-worker-main`, and `pv-primary`. Only one can be active under the Free plan.

- Active/executable now: `pv-primary`, Production-only.
- `TrPaqet`: connected/healthy, but fresh execution returns `upgrade_required`.
- `pv-worker-main`: connected/healthy, but fresh execution returns `upgrade_required`.
- Task13 reconstruction remains assigned to the first executable development Worker; Task16 remains assigned to the next independent development Worker.

## Production deployment rules

Before every Production mutation: fresh encrypted DB/config backup + rollback snapshot, verify artifacts, apply only intended migration/release, verify readiness + Runtime/Telemetry/Caddy/customer/accounting invariants + exact release provenance, and roll back on any failed invariant. Never use Production as a test database.

## Next sequence

1. Keep #64 draft; it is diverged and `mergeable=false`.
2. Make a development Worker executable, reconstruct Task13 onto exact current main, and rerun all exact-tree gates.
3. Run the final HTTP1/HTTP2 exact-kill rehearsal with exactly-once accounting and no kill-triggered Caddy lifecycle action.
4. Only then merge/deploy with fresh backup/rollback/postflight proof.
5. Start Task16 on the next independently executable Worker with RED retention/pagination tests.