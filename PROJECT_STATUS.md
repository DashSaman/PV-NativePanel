# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `d00c96be9dccbb1f5402a24a18beeef6910c78cd` (verified merge of PR #57; documentation/evidence reconciliation after Task15 Production rollout).
- Exact-main CI run `33476564850`: **SUCCESS**.
- Open PRs: old draft #4 only; it is unrelated to the current roadmap and remains unmerged pending a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / current-main recovery + complete publication/rehearsal incomplete**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 / design gate failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Fresh Production truth

Production runtime remains the guarded Task15 rollout from source `26aa74dddfd23535e45837f21531cf67ea2fd238`; PR #57 changed repository documentation/evidence only.

Fresh read-only verification on 2026-09-01 confirmed:

- `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`: **active**.
- `GET /api/v1/health/ready`: HTTP **200**, body reports `ready=true`, `db=ok`, `schema=ok`.
- database/expected schema remain **20** from the verified Task15 rollout.
- Task15 backup and rollback evidence remains in `ops/evidence/TASK15-20260901-schema20-production-pass.md` and `/var/backups/pvnaive`.

No Production mutation was required for this reconciliation.

## Task13 — exact live-session kill

Task13 remains the highest-priority unfinished runtime/session item. Do not merge the old partial publication branch. Required invariants: exact runtime-credential/node/boot/session tuple, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting. Reconstruct/recover against current schema20 main and rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16 — bounded IP/session history / schema21

Schema20 is stable, but Task16 is not mergeable. Review of the worker candidate found that retention/pagination behavior must be server-enforced rather than caller-extensible: a caller must not be able to request history older than the exact 30-day contract or exceed bounded pagination/read limits. Required next proof is RED tests for oversized retention/page requests, minimal server-side constants/clamps, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker / orchestration capacity

Four SentinelX hosts are connected but the current Free plan permits only one active host. `pv-primary` is executable. `pv-worker-main` is connected and healthy but returns `upgrade_required` for execution. Therefore persistent worker reports cannot be freshly trusted and Production must not be repurposed as a development/test worker.

Human action is needed only to restore parallel worker capacity: disconnect unused connected hosts so a worker can become active, or change the SentinelX host limit.

## Immediate execution order

1. Recover/reconcile Task13 against exact schema20 main and publish only after full exact-session-kill gates pass, including fresh real HTTP1+HTTP2 rehearsal.
2. Repair Task16 retention/pagination bypass with RED tests first; then rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before schema21 publication.
3. Keep old draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
4. Before every future Production mutation, create and verify a fresh encrypted backup + rollback snapshot and retain exact provenance.
