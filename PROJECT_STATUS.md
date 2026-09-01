# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `26aa74dddfd23535e45837f21531cf67ea2fd238` (verified merge of PR #54, Task15 schema20).
- Exact-main CI run `33471780919`: **SUCCESS**.
- Open PRs: old draft #4 only; it is unrelated to the current roadmap and remains unmerged pending a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / current-main recovery + publication incomplete**.
- Task16 bounded IP/session history: **IN PROGRESS / now unblocked by stable schema20 but not merged/deployed**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Fresh Production truth

Fresh guarded rollout on 2026-09-01 deployed exact main `26aa74dddfd23535e45837f21531cf67ea2fd238` and schema20.

- `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, `pvnaive-telemetry-agent.service`: **active**.
- API listener: `127.0.0.1:8080`; Caddy: public `:80` / `:443`.
- `GET /api/v1/health/live`: HTTP **200**.
- `GET /api/v1/health/ready`: HTTP **200**, DB/schema ready.
- public SNI-correct panel probe: HTTP **200**.
- public SNI-correct API readiness probe: HTTP **200**.
- database schema: **20**; `/etc/pvnaive/db.env` expected schema: **20**.
- release metadata source commit: `26aa74dddfd23535e45837f21531cf67ea2fd238`.
- Caddy: `v2.11.2`, pinned `http.handlers.forward_proxy` module present and config validation passes.

Task15 rollout used a fresh encrypted DB/config backup and a separately checksummed rollback snapshot before migration. Exact evidence is in `ops/evidence/TASK15-20260901-schema20-production-pass.md`.

Important health-route correction remains: `/healthz` and `/readyz` are not canonical on this deployed API. Use `/api/v1/health/live` and `/api/v1/health/ready`.

## Task13 — exact live-session kill

Task13 remains the highest-priority unfinished runtime/session item. Do not merge the old partial publication branch. Required invariants: exact runtime-credential/node/boot/session tuple, sibling survival, no credential revoke, no Caddy reload/restart, idempotent repeated kill, tenant/role isolation and exactly-once normal final accounting. Reconstruct/recover against current schema20 main and rerun full Go/race/Web/real HTTP1+HTTP2/pinned-forwardproxy/reproducible-Caddy evidence before publication.

## Task16 — bounded IP/session history / schema21

Schema20 is now stable in Production, so Task16 is no longer blocked by Task15 ordering. Required proof remains exact 30-day retention, tenant-scoped RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, bounded pagination/read paths, coherent schema21 migration/checksums and safe rollback. It is not yet merged or deployed.

## Worker / orchestration capacity

Three SentinelX hosts remain connected but the current plan permits only one active host. `pv-primary` is executable; the two worker hosts remain unavailable for command execution while Production is the active host. Therefore persistent worker reports cannot be freshly trusted until a worker becomes executable. Repository/CI work can continue independently.

Human action is needed only to restore parallel worker capacity: disconnect an unused connected host so a worker can become active, or change the SentinelX host limit.

## Immediate execution order

1. Publish this verified Task15 Production evidence/doc reconciliation and merge only if documentation CI is green.
2. Reconcile/recover Task13 against exact schema20 main and republish only after full exact-session-kill gates pass.
3. Rebase/reconcile Task16 onto schema20 main, rerun PG18 + Go/Web/RLS/retention/purge/rollback gates, then publish schema21 candidate.
4. Keep old draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
5. Before every future Production mutation, create and verify a fresh encrypted backup + rollback snapshot and retain exact provenance.
