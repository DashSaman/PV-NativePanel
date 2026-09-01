# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `6e58111665993e6e62c2d4e364a476d20ceb4896` (PR #69 documentation reconciliation).
- Exact-main push CI run `33550756339`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact head `233fb1a3cc1b926e709b2035e777c9594fae2ef1`.
- Exact-head #64 workflows were restarted for `233fb1a...`; CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy were **in progress** at this checkpoint. Do not reuse green results from a prior head.
- Task13 dedicated `pvnaive-session-control` permission boundary remains in place: API receives only that supplementary group; Caddy keeps telemetry access separately and additionally receives session-control; the Caddy-owned socket resolves/chowns to the dedicated GID then applies `0660`, failing closed.
- Same-schema release logic takes the mandatory backup before creating the dedicated group and removes a newly-created group during failed-release rollback.
- New Task13 Web/UI increment is published: exact session DELETE sends only `userId/sessionId` in the path with CSRF and no caller tuple/body; successful kill reloads trusted active-session state; the Sessions modal exposes a per-session `قطع نشست` action with explicit confirmation that password/subscription remain unchanged.
- Fresh Worker verification for the exact published tree: targeted Web 10/10 PASS, full Web 19 files / 64 tests PASS, production Web build PASS, focused Go race PASS, full Go PASS, `TASK13_SESSION_CONTROL_PERMISSIONS=PASSED`, pinned-forwardproxy Task13 race PASS, and `git diff --check` PASS.
- Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` inspected on the Worker are stale S04-era ledgers from Aug 27 and contain no newer Task13/Task16 completion; they are historical evidence only.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production remains on the guarded Task15/schema20 rollout; no Task13 code has been deployed.

A fresh read-only Production probe was attempted in this checkpoint, but `pv-primary` returned `upgrade_required` because the SentinelX Free plan permits one active host while three are connected and the active slot is on development Worker `TrPaqet`. Therefore no fresh Production-health PASS is claimed here.

No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 now contains exact-tuple registry/client primitives, live CONNECT registration, reload-safe Unix session-control lifecycle, ownership-checked API kill, narrow dedicated socket permissions, backup-before-permission release safety, and the exact-session Web/UI action.

Still required before Task13 can merge: DB-integrated handler-level ownership/IDOR/CSRF failure-path proof; final release packaging/install/rollback proof; full final-tree Go/race/Web/pinned/reproducible-Caddy gates; and fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival and exactly-once final accounting.

## Task16 — bounded IP/session history / schema21

No fresh current-main Task16 implementation is credited. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh SentinelX listing shows **three connected hosts**: `TrPaqet`, `pv-worker-main`, and `pv-primary`; the Free plan permits one active host.

- Active/executable now: `TrPaqet`, development lane for Task13.
- `pv-primary`: Production-only; fresh probe is blocked by `upgrade_required` while the development Worker owns the slot.
- `pv-worker-main`: connected/healthy but fresh execution also returns `upgrade_required`.
- Task16 remains assigned to the next independently executable development Worker.

True parallel Task13 + Task16 execution, or simultaneous development plus Production probing, requires human action: disconnect/switch hosts or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft until exact-head workflows and all remaining Task13 gates complete.
2. Continue Task13 on `TrPaqet` with DB-integrated ownership/IDOR/CSRF failure-path proof and release packaging/install/rollback verification.
3. Run full exact-tree gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
4. Merge/deploy Task13 only after the exact verified tree is green and fresh encrypted Production backup + rollback snapshot + postflight access are available.
5. Start Task16 on the next independently executable Worker with RED retention/pagination tests; never use Production for schema21 development.
