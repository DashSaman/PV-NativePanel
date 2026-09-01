# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `8b8549b8d1431dfa8858c207e55b9a52eaa2c4e8` (verified merge of PR #60; documentation/state reconciliation only).
- Exact-main CI run `33500952494`: **SUCCESS**.
- Open roadmap PRs: none. Old draft #4 is unrelated to the current roadmap and remains unmerged pending a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / active-worker schema20 reconstruction underway; complete integration/publication/rehearsal still required**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 / design gate failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production runtime remains the guarded Task15 rollout from source `26aa74dddfd23535e45837f21531cf67ea2fd238`; later merges through PR #60 changed repository documentation/evidence only.

Last successful fresh read-only verification confirmed:

- `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`: **active**.
- `GET /api/v1/health/ready`: body reports `ready=true`, `db=ok`, `schema=ok`.
- Production remains on the verified schema20 Task15 rollout.
- Task15 backup and rollback evidence remains in `ops/evidence/TASK15-20260901-schema20-production-pass.md` and `/var/backups/pvnaive`.

A new Production read-only probe in the latest orchestration run could not execute because the SentinelX Free-plan active slot had moved away from `pv-primary`; the host remained connected/healthy but returned `upgrade_required`. No Production mutation was attempted.

## Task13 — exact live-session kill

Task13 remains the highest-priority unfinished runtime/session item. The old GitHub publication branch `lead/task13-kill-session-publish-20260901` remains partial and stale; it must not be merged wholesale.

An executable worker was recovered on `ubuntu-4gb-hel1-1` while `pv-primary` and `pv-worker-main` were inactive under the one-active-host plan limit. A clean repository was cloned at exact main `8b8549b8...`. The historical Task13 primitives were reapplied locally on top of schema20: exact session-control protocol/client and exact-tuple in-process registry. Focused `go test -race ./internal/sessionkill ./internal/sessioncontrol -count=1` passed; `sessionkill` still has no committed registry test in that historical partial candidate, so this is not sufficient for publication.

Required invariants remain exact runtime-credential/node/boot/session tuple identity, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting. The remaining work is surgical integration with current forwardproxy/accounting, API authorization/ownership, UI action and fresh real HTTP1+HTTP2 rehearsal before publication.

## Task16 — bounded IP/session history / schema21

Schema20 is stable, but Task16 is not mergeable. Retention/pagination behavior must be server-enforced rather than caller-extensible: a caller must not be able to request history older than the exact 30-day contract or exceed bounded pagination/read limits. Required next proof is RED tests for oversized retention/page requests, minimal server-side constants/clamps, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker / orchestration capacity

Four SentinelX hosts are connected and the current Free plan permits only one active host. In the latest run, `pv-primary`, `pv-worker-main`, and `host_311ff...` all returned `upgrade_required`, while `ubuntu-4gb-hel1-1` was executable with about 32% RAM used. That host is now the safe development lane for Task13 reconstruction; Production remains excluded from development/testing.

Human action is still needed to restore true parallel worker capacity: disconnect unused connected hosts so multiple intended workers can be selected over time, or change the SentinelX host limit.

## Immediate execution order

1. Finish surgical Task13 integration on the active worker against exact schema20 main; add RED tests first for any new integration behavior, then run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 rehearsal.
2. Publish Task13 only as the exact verified tree and merge only after exact-head CI is green; Production rollout then requires fresh encrypted backup + rollback snapshot and postflight evidence.
3. Repair Task16 retention/pagination bypass with RED tests first; then rerun PG18 + Go/Web/RLS/retention/purge/rollback gates before schema21 publication.
4. Keep old draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
