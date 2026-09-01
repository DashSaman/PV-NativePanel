# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `62573fee8b88e4f951224da10e6a26d5b5838a54`, verified merge of PR #63 (documentation/state reconciliation only).
- Exact-main CI run `33520634435` for this SHA completed **SUCCESS**.
- Current roadmap PR: **draft PR #64**, `lead/task13-reconstruct-62573fee`, exact head `263dc3e34982c739363822c7ac2cc643af7408c2`.
- Prior draft PR #62 is closed as superseded because it had diverged from current main; its validated seven-file primitive/control delta was republished cleanly in PR #64.
- Old draft #4 remains unrelated to the current roadmap and awaits a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**. Exact-tuple registry/client plus local control handler are published on exact current main; forwardproxy/Unix-listener wiring, API authorization/ownership/CSRF, UI action, exact final accounting and real HTTP1+HTTP2 rehearsal are still required.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production runtime remains the guarded Task15 rollout from source `26aa74dddfd23535e45837f21531cf67ea2fd238`; later repository merges through PR #63 changed documentation/evidence only.

Fresh read-only verification in the current run succeeded on `pv-primary`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: **active**;
- `GET /api/v1/health/ready`: `ready=true`, `db=ok`, `schema=ok`;
- no Production mutation, restart, reload, deployment or migration was performed.

The latest probe did not re-read the API process environment/provenance because the read-only script exited before that section; therefore previous verified schema20/provenance evidence remains authoritative for those details rather than being restated as fresh evidence.

## Task13 — exact live-session kill

The stale branch `lead/task13-kill-session-publish-20260901` and superseded PR #62 must not be merged. Active reconstruction is draft PR #64, created directly from exact current main `62573fee...`.

Current published scope in PR #64 is exactly seven files:

- exact runtime-credential/node/boot/session registry primitive;
- sibling survival, forged-tuple rejection, repeat-kill idempotence and unregister semantics;
- local session-control protocol/client;
- local control HTTP handler rejecting incomplete tuples before touching live state.

PR #64 starts directly from current main and is currently 0 commits behind it. Its three GitHub workflows started on exact head `263dc3e...`: CI and Exact Accounting were queued and Pinned Forwardproxy was in progress at the latest observation. Do not merge until all are green and the remaining Task13 integration/rehearsal gates are completed.

Mandatory next Task13 proof remains: TDD-first forwardproxy registration/cancel wiring + Unix listener ownership; preserve Task14/15 response/finalization and trusted `RemoteAddr`/`ClientIP` semantics; add API RBAC/ownership/CSRF and UI action; then prove real HTTP/1.1 + HTTP/2 exact kill, sibling survival, forged-tuple rejection, idempotency, credential survival and exactly-once final accounting on the exact published tree.

## Task16 — bounded IP/session history / schema21

Schema20 is stable, but Task16 is not mergeable. Retention/pagination behavior must be server-enforced rather than caller-extensible: a caller must not be able to request history older than the exact 30-day contract or exceed bounded pagination/read limits. Required proof remains RED tests for oversized retention/page requests, minimal server-side constants/clamps, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker / orchestration capacity

Fresh host listing shows **three connected SentinelX hosts**: `TrPaqet`, `pv-worker-main`, and `pv-primary`. Free-plan capacity allows only one active host. Fresh execution attempts on `TrPaqet` and `pv-worker-main` returned `upgrade_required`; `pv-primary` is the active slot and is reserved for Production read-only/guarded operations, not development.

Therefore no safe development worker is currently executable. True parallel worker capacity requires human action: disconnect unused connected hosts or increase the active-host limit.

## Immediate execution order

1. Keep PR #64 draft while its exact-head workflows complete.
2. When a development worker becomes executable, continue Task13 surgically on PR #64 with TDD-first forwardproxy/listener wiring, then API authorization and UI.
3. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
4. Merge/deploy Task13 only after all exact-tree gates are green and a fresh encrypted Production backup + rollback snapshot + postflight are ready.
5. In parallel when another development worker is available, start Task16 with RED tests for >30-day retention and oversized pagination before any schema21 code is accepted.
6. Keep old draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
