# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `1aea1961515ab86428231499490202af4aef5e97` (PR #67 documentation reconciliation).
- Exact-main push CI run `33541904912`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact head `1e6974a8098a1162d4a5ca6ebc1d89206c02ddd9`.
- Exact-head #64 workflows for this new lifecycle publication are running; do not reuse green results from the prior head.
- Published Task13 lifecycle patch blob is byte-identical to the Worker-tested file: `b5889058caf4312df5508655193a6275c4ae5a1e`.
- Old draft #4 remains unrelated to the current roadmap and awaits a real Karing client smoke.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production remains on the guarded Task15/schema20 rollout; no Task13 code has been deployed.

A fresh read-only Production probe could not be executed in this checkpoint because the single SentinelX execution slot is currently on development Worker `TrPaqet`; `pv-primary` returns `upgrade_required`. Therefore no new Production-health PASS is claimed here. The last fresh successful health observation remains historical evidence only until Primary is executable again.

No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 now contains the exact-tuple registry/client, live CONNECT registration, Unix session-control server primitive, and a reload-safe Caddy-owned lifecycle.

Latest lifecycle evidence on the Worker:

- TDD RED failed on the missing lifecycle acquire/wiring functions before implementation;
- only accounting-enabled forwardproxy handlers acquire `/run/pvnaive/session-control.sock`; ordinary non-accounting handlers do not;
- overlapping old/new Caddy configs share one exact-session registry and socket through a ref-counted lease;
- predecessor cleanup cannot remove a successor socket; final lease cleanup removes the owned socket;
- patched pinned forwardproxy `go test -race ./... -count=1` passed with isolated Go 1.26.3;
- `tests/stages/Task13_forwardproxy_session_control_test.sh` passed;
- pinned forwardproxy boundary proof passed;
- reproducible-Caddy build contract passed;
- `git diff --check` passed;
- the published `0003` patch exactly reproduces the validated lifecycle source from the existing `0002` state and its Git blob SHA matches the Worker file.

Still required before Task13 can merge: narrow Production service/group permission proof for the `0660` socket; API ownership/RBAC/IDOR/CSRF without credential mutation; UI action; full final-tree Go/race/Web/pinned/reproducible-Caddy gates; and fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival and exactly-once final accounting.

## Task16 — bounded IP/session history / schema21

Persistent Worker evidence remains uncredited: there is no fresh current-main Task16 implementation to integrate. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh SentinelX listing shows **two connected hosts**: `TrPaqet` and `pv-primary`, while the Free plan permits one active host.

- Active/executable now: `TrPaqet`, reserved for development work.
- `pv-primary`: Production-only and currently blocked from execution by the one-active-host limit.
- Task16 has no independently executable second Worker lane at this checkpoint.

True parallel Task13 + Task16 execution, or simultaneous development plus Production probing, requires human action: switch/disconnect hosts as needed or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft while its new exact-head workflows run and while API/UI/rehearsal remain incomplete.
2. Continue Task13 on `TrPaqet` with the narrow socket service/group permission proof, then API ownership/RBAC/IDOR/CSRF and UI.
3. Run full exact-tree gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
4. Merge/deploy Task13 only after the exact verified tree is green and fresh encrypted Production backup + rollback snapshot + postflight access are available.
5. Assign Task16 to the next independently executable Worker; do not move schema21 development onto Production.
