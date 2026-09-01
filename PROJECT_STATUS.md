# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `cce50c4b198bd8ea449f385c1198fffbddfead8e`, documentation/state reconciliation only.
- Exact-main push CI run `33530087359` for this SHA completed **SUCCESS**.
- Current roadmap PR: **draft PR #64**, `lead/task13-reconstruct-62573fee`, exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head #64 gates at this checkpoint: WS1 Exact Accounting **SUCCESS**, WS1 Pinned Forwardproxy **SUCCESS**, CI still **IN PROGRESS**. Do not merge while incomplete even after CI turns green.
- Old draft #4 remains unrelated to the current roadmap and awaits a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production runtime remains the guarded Task15 schema20 rollout from previously verified source `26aa74dddfd23535e45837f21531cf67ea2fd238`; later repository changes have not been deployed.

A fresh Production probe was attempted in this run, but `pv-primary` is connected yet non-executable because the SentinelX Free plan permits one active host and the active slot is currently assigned to the development Worker. Therefore this run does **not** upgrade prior health/provenance evidence into a fresh observation.

No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed in this run. Previous verified schema20 health/provenance remains the latest valid Production evidence until `pv-primary` becomes executable again.

## Task13 — exact live-session kill

Draft PR #64 now includes a newly validated/published Unix session-control server primitive in addition to the earlier exact-tuple registry/client and live CONNECT registration.

Fresh TDD/Worker evidence for head `485657d...`:

- RED was observed specifically because `startPVNaiveSessionControlServer` did not exist;
- minimal Unix-domain control server was then implemented with bounded/strict JSON, exact-tuple kill, socket mode `0660`, stale-socket handling, and owned-socket cleanup;
- a real HTTP request over a Unix socket proves target-only kill while the sibling remains untouched;
- explicit unregister-removes-tuple coverage is present;
- patched upstream forwardproxy `go test ./...` and `go test -race ./...` passed with Go 1.26.3;
- `tests/stages/Task13_forwardproxy_session_control_test.sh` passed;
- pinned forwardproxy boundary proof passed;
- repository `go test ./... -count=1`, focused race tests and `git diff --check` passed.

A combined local full-gate/reproducible-Caddy invocation exceeded the Worker execution ceiling, so that timeout is **not** counted as a PASS. Exact GitHub gates remain authoritative for the published head.

Still required before Task13 can merge: Caddy-owned reload-safe lifecycle for `/run/pvnaive/session-control.sock`; a narrow service/group permission model that lets only the intended local API cross the `0660` socket boundary without broadening access to accounting; API ownership/RBAC/IDOR/CSRF; UI action; full exact-head gates; and fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival and exactly-once final accounting.

## Task16 — bounded IP/session history / schema21

Persistent Worker workspaces were reconciled again. The current Task16 run workspace is clean on an older main snapshot and contains no new diff/completion to integrate. Task16 therefore remains uncredited.

Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh host listing now shows **two connected SentinelX hosts**: `TrPaqet` and `pv-primary`. The Free plan permits one active host.

- `TrPaqet` is currently executable and assigned to Task13 development/verification.
- `pv-primary` is connected/healthy but a fresh command returns `upgrade_required`; it cannot be used concurrently for fresh Production probing.
- Historical Task13 Worker workspaces were reconciled: the old untracked local session-control server files are already represented by current PR #64 work and are not a separate completion.
- Task16 has no fresh completed Worker delta to integrate.

True parallel Task13 + Task16 execution, or simultaneous development plus fresh Production probing, requires human action: disconnect/switch the active host when needed or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft while exact-head CI finishes and while lifecycle/API/UI/rehearsal work is incomplete.
2. Continue Task13 on the active Worker with TDD-first Caddy reload-safe session-control lifecycle and narrow Unix-socket permission proof.
3. Then implement API ownership/RBAC/IDOR/CSRF and UI exact-session kill.
4. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
5. Merge/deploy Task13 only after the exact verified tree is green and a fresh encrypted Production backup + rollback snapshot + postflight are available.
6. Assign Task16 to the next independently executable Worker; do not move its schema21 development onto Production.
