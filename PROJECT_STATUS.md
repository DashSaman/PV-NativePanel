# PVNaive — Canonical Project Status

Last updated: 2026-09-01

This file is current repository + Production truth. Historical stage notes, worker reports and stale branches are evidence only; exact GitHub `main`, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Customer/service state, Runtime credentials, subscription delivery, exact direct-Naive accounting/session telemetry and privileged Runtime mutation stay separate. Never fabricate usage/online/IP/session history; never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, postflight verification and exact deployed provenance.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current main: `392a153f14dc311f6d1dffe60e6b4d4da5f8cb17`, documentation/state reconciliation only.
- Prior exact-main push CI on `cce50c4b198bd8ea449f385c1198fffbddfead8e` completed **SUCCESS**; the new merge SHA has not yet surfaced a commit-associated workflow run, so no green claim is made for it yet.
- Current roadmap PR: **draft PR #64**, `lead/task13-reconstruct-62573fee`, exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head #64 gates are now all **SUCCESS**: CI run `33537563177`, WS1 Exact Accounting `33537563178`, WS1 Pinned Forwardproxy `33537563183`.
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

Fresh read-only Production observation in this run:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` are all `active`.
- `GET http://127.0.0.1:8080/api/v1/health/ready` returned `{"db":"ok","ready":true,"schema":"ok","status":"ready"}`.
- Journal error scanning could not be completed because the SentinelX account lacks journal permissions; that missing observation is not treated as a pass.
- `/opt/pvnaive/deploy-src` provenance/process-environment inspection is permission-blocked in this run, so previously verified release provenance remains the latest valid provenance evidence.

No Production mutation, deployment, migration, restart, reload, DB write or credential change was performed.

## Task13 — exact live-session kill

Draft PR #64 now includes the exact-tuple registry/client, live CONNECT registration, and a validated Unix session-control server primitive.

Validated evidence for exact head `485657d...`:

- RED was observed specifically because `startPVNaiveSessionControlServer` did not exist;
- minimal Unix-domain control server was implemented with bounded/strict JSON, exact-tuple kill, socket mode `0660`, stale-socket handling, and owned-socket cleanup;
- a real HTTP request over a Unix socket proves target-only kill while the sibling remains untouched;
- explicit unregister-removes-tuple coverage is present;
- patched upstream forwardproxy `go test ./...` and `go test -race ./...` passed with Go 1.26.3;
- `tests/stages/Task13_forwardproxy_session_control_test.sh`, pinned forwardproxy boundary proof, repository Go tests, focused race tests and `git diff --check` passed;
- all three exact GitHub gates are green for the published head.

Still required before Task13 can merge: Caddy-owned reload-safe lifecycle for `/run/pvnaive/session-control.sock`; a narrow service/group permission model allowing only the intended local API across the `0660` socket boundary; API ownership/RBAC/IDOR/CSRF without credential mutation; UI action; and fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival and exactly-once final accounting.

## Task16 — bounded IP/session history / schema21

Persistent Worker evidence remains uncredited: there is no fresh current-main Task16 implementation to integrate. Required proof remains server-enforced exact 30-day retention and hard-bounded pagination/read limits, tenant RLS/authorization, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent schema21 migration/checksums, PostgreSQL18 behavior and rollback safety. RED tests for >30-day retention and oversized pagination must precede implementation.

## Worker / orchestration capacity

Fresh SentinelX listing shows **two connected hosts**: `TrPaqet` and `pv-primary`, while the Free plan permits one active host.

- The active slot is currently on `pv-primary`, which enabled the fresh read-only Production probe above.
- `TrPaqet` is connected/healthy but currently returns `upgrade_required`, so Task13 development cannot continue there simultaneously.
- Task16 has no independently executable Worker lane at this checkpoint.

True parallel Task13 + Task16 execution, or simultaneous development plus Production probing, requires human action: disconnect/switch the active host as needed or increase the SentinelX active-host limit.

## Immediate execution order

1. Keep PR #64 draft despite its green exact-head gates because functional integration remains incomplete.
2. When the development slot returns to `TrPaqet`, continue Task13 TDD-first with Caddy reload-safe session-control lifecycle and narrow Unix-socket permission proof.
3. Then implement API ownership/RBAC/IDOR/CSRF and UI exact-session kill.
4. Run full exact-tree gates plus fresh real HTTP1+HTTP2 exact-kill rehearsal proving sibling survival and exactly-once final accounting.
5. Merge/deploy Task13 only after the exact verified tree is green and a fresh encrypted Production backup + rollback snapshot + postflight are available.
6. Assign Task16 to the next independently executable Worker; do not move schema21 development onto Production.