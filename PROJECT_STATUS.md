# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `b5f9466c1464afa9bc3183418aaf8e124b890563` (PR #70 documentation reconciliation).
- Exact-main push CI run `33556193807`: **SUCCESS**.
- Current roadmap PR: **draft PR #64**, branch `lead/task13-reconstruct-62573fee`, exact published head `5bc42d8dedd682eaf560a99777b21b9e82062c79`.
- Exact-head #64 workflows on `5bc42d8...`: CI run `33557107036` **SUCCESS**, WS1 Exact Accounting run `33557107038` **SUCCESS**, WS1 Pinned Forwardproxy run `33557107045` **SUCCESS**.
- Task13 now includes exact tuple registry/client primitives, live CONNECT registration after accounting-open/trusted-peer success, reload-safe Caddy-owned Unix listener, a dedicated `pvnaive-session-control` group and `0660` socket, ownership-checked DELETE API, and per-session Web/UI kill without credential mutation.
- New TDD-first release work closes a real activation gap: R1 now packages the exact reproducible pinned Task13 Caddy binary plus provenance and the Caddy systemd drop-in; deploy validates the candidate, takes the mandatory encrypted backup before mutation, installs it with exactly one controlled binary-swap Caddy restart, and rollback restores the previous Caddy binary/drop-in state and reactivates it.
- Local Worker proof for this increment: focused Go race PASS, full Go PASS, `TASK13_R1_RELEASE_CONTRACT=PASSED`, `TASK13_SESSION_CONTROL_PERMISSIONS=PASSED`, reproducible pinned Caddy build PASS, and `TASK13_FORWARDPROXY_SESSION_CONTROL=PASSED`. The pinned Caddy binary SHA is `0e44d42a63b5e1001b6c2410f6fa7108256aabb89dfd86cbb50334030bdddb0e`.
- New TDD-first DB/auth proof: the CI-contract test first failed RED because no Task13 API-kill rehearsal was wired; it now passes GREEN. The new PostgreSQL18 rehearsal proves missing CSRF and cross-tenant IDOR attempts cause zero session-control side effects, while an owned session emits exactly one trusted full-tuple kill and leaves the credential active. Worker-local Go/race and pinned Caddy proofs pass; PostgreSQL18 execution is delegated to exact-head GitHub CI because this Worker has no Docker/PG18 lane.
- The prior Task13 head passed PostgreSQL18 database gates, Go/vet/tests, Web tests/build, S04/S04R rehearsal and the R1 bundle job. The new `5bc42d8...` head is fully green; its PostgreSQL18 database, Go, Web, full rehearsal, and bundle jobs all completed successfully.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task13 exact live-session kill: **IN PROGRESS / draft PR #64**.
- Task16 bounded IP/session history: **IN PROGRESS / schema21 design gate remains failed until retention and pagination are server-bounded**.

No task becomes DONE from a local candidate, historical worker report or partial branch alone.

## Production truth

Production remains on the guarded Task15/schema20 rollout; no Task13 code has been deployed.

A fresh read-only Production probe was attempted in this checkpoint, but `pv-primary` returned `upgrade_required`: the SentinelX Free plan permits one active host while three are connected and the active slot is on development Worker `TrPaqet`. Therefore no fresh Production-health PASS is claimed here.

No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 now contains the data-plane registry/control path, dedicated socket permission boundary, ownership-checked API, Web/UI kill action, and R1 packaging/install/rollback support for the patched reproducible Caddy binary and drop-in. The new exact head is fully green, including the PostgreSQL18 DB/auth rehearsal.

The DB-integrated handler ownership/IDOR/CSRF lane passed exact-head PostgreSQL18 CI. Still required before Task13 can merge: fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no Caddy lifecycle action caused by a kill request, and exactly-once final accounting. Only after those gates may the final R1 artifact be backed up, deployed and postflight-verified on Production.

## Task16 — bounded IP/session history / schema21

No fresh current-main Task16 implementation is credited. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh SentinelX listing shows **three connected hosts**: `TrPaqet`, `pv-worker-main`, and `pv-primary`; the Free plan permits one active host.

- Active/executable now: `TrPaqet`, development lane for Task13.
- `pv-primary`: Production-only; fresh probe is blocked by `upgrade_required` while the development Worker owns the slot.
- `pv-worker-main`: connected/healthy but inactive under the same one-active-host limit.
- Task16 remains assigned to the next independently executable development Worker.

True parallel Task13 + Task16 execution, or simultaneous development plus Production probing, requires human action: disconnect/switch hosts or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft despite the exact head being fully green; the real live protocol/accounting proof still gates merge.
2. Run fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting with no kill-triggered Caddy restart/reload.
3. Merge/deploy Task13 only after the exact verified tree is green and fresh encrypted Production backup + rollback snapshot + postflight access are available.
4. Start Task16 on the next independently executable Worker with RED retention/pagination tests; never use Production for schema21 development.
