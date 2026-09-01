# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `a29b5ef434a72004af80cf489f47fffe0b0a03a8` (verified merge of PR #61; documentation/state reconciliation only).
- Exact-main CI for this main was already verified green before the current run; no new main mutation occurred in this run.
- Open roadmap PR: **draft PR #62**, `lead/task13-reconstruct-a29b5ef`, exact published head `2e0f485d61f2dd70647b6f626b1f8a18178336d7`. Old draft #4 remains unrelated to the current roadmap and awaits a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #62**. Primitive registry/client plus a TDD-first local control handler are now published; forwardproxy/Unix-listener wiring, API authorization/ownership/CSRF, UI action, exact final accounting and real HTTP1+HTTP2 rehearsal are still required.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production runtime remains the guarded Task15 rollout from source `26aa74dddfd23535e45837f21531cf67ea2fd238`; later merges through PR #61 changed repository documentation/evidence only.

The last successful fresh read-only verification confirmed all four core services active, `/api/v1/health/ready` reporting `ready=true`, `db=ok`, `schema=ok`, and schema20 Task15 provenance.

A fresh read-only probe in the current run could not execute because `pv-primary` is connected/healthy but inactive under the SentinelX Free-plan one-active-host limit (`4` connected hosts). The probe returned `upgrade_required`. No Production mutation, restart, reload or migration was attempted.

## Task13 — exact live-session kill

The stale branch `lead/task13-kill-session-publish-20260901` must not be merged wholesale. The active reconstruction branch is draft PR #62 on current schema20 main.

Current verified/published Task13 scope:

- exact runtime-credential/node/boot/session registry primitive;
- sibling survival, forged-tuple rejection, repeat-kill idempotence and unregister semantics;
- local session-control protocol/client;
- initial CI formatting blocker was isolated to gofmt and corrected without behavior change;
- TDD-first local control handler was added after observing RED `undefined: NewHandler`;
- fresh isolated Worker verification with a clean Go 1.25.0 toolchain: focused handler tests PASS, `go test -race ./internal/sessionkill ./internal/sessioncontrol -count=1` PASS, full `go test ./... -count=1` PASS, `git diff --check` PASS;
- handler rejects incomplete tuples before touching live state and exact kill preserves siblings.

Exact-head CI for `2e0f485...` is currently running. PR #62 remains draft and must not merge until forwardproxy/Unix-listener ownership, API RBAC/ownership/CSRF, UI action, exact final accounting, full gates and fresh real HTTP/1.1 + HTTP/2 kill rehearsal are green on the exact published tree.

## Task16 — bounded IP/session history / schema21

Schema20 is stable, but Task16 is not mergeable. Retention/pagination behavior must be server-enforced rather than caller-extensible: a caller must not be able to request history older than the exact 30-day contract or exceed bounded pagination/read limits. Required proof remains RED tests for oversized retention/page requests, minimal server-side constants/clamps, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety.

## Worker / orchestration capacity

Four SentinelX hosts are connected and the current Free plan permits only one active host. In this run `TrPaqet` (`host_311ff...`) is executable and was used for isolated Task13 verification and a clean Task16 inspection workspace. `pv-primary` is connected/healthy but inactive with `upgrade_required`; Production remains excluded from development/testing.

True parallel worker capacity still requires human action: disconnect unused connected hosts or change the SentinelX active-host limit.

## Immediate execution order

1. Finish Task13 surgical integration on draft PR #62: forwardproxy registration/cancel path + Unix listener, then API RBAC/ownership/CSRF and UI action, each TDD-first.
2. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
3. Merge only the exact verified Task13 tree; Production rollout then requires fresh encrypted backup + rollback snapshot and postflight evidence.
4. In parallel when a second worker becomes active, implement Task16 RED tests for >30-day retention and oversized pagination before any schema21 code is accepted.
5. Keep old draft PR #4 out of the roadmap unless a real Karing client smoke explicitly revives it.
