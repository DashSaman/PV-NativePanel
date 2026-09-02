# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `6e7e14aabe22719099443d80d65798bb53b2769c`.
- Task13: draft PR #64, branch `lead/task13-reconstruct-62573fee`, exact head `c79f5385b7a751c30282948423b8c34d1ba89deb`.
- Fresh compare: Task13 is 44 commits ahead / 0 behind current main; merge-base is exact current main.
- Exact-head Task13 workflows: CI `33612706915` **SUCCESS**; WS1 Exact Accounting `33612706979` **SUCCESS**; WS1 Pinned Forwardproxy `33612706951` **SUCCESS**.
- Task13 exact live-session kill: **IN PROGRESS / draft / sole remaining merge gate is fresh real HTTP1+HTTP2 live proof**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task16 bounded IP/session history: **IN PROGRESS / draft PR #81 / issue #79 / schema21 TDD RED established**.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh read-only observation this cycle on `pv-primary`:

- `pvnaive-api`: **active**;
- `caddy-naive`: **active**;
- `pvnaive-runtime-agent`: **active**;
- `pvnaive-telemetry-agent`: **active**;
- `GET http://127.0.0.1:8080/api/v1/health/ready`: `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `GET http://127.0.0.1:8080/api/v1/health/live`: `service=pvnaive-api`, `status=ok`;
- 45-minute journal scan: zero `panic`, `fatal`, `schema mismatch` or `segmentation fault` matches.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Task13 — exact live-session kill

All exact-head GitHub gates are green on `c79f5385...`. The sole merge gate is a fresh real HTTP/1.1 + HTTP/2 rehearsal on that exact head proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` remains assigned to that rehearsal. It is connected but current SentinelX plan capacity returns `upgrade_required` while Production Primary owns the single active-host slot. Production must not be used as the rehearsal lane.

Only after live proof may Task13 be merged. Deployment additionally requires a fresh encrypted Production backup + rollback snapshot, exact R1 artifact provenance and complete postflight.

## Task16 — bounded IP/session history / schema21

Issue #79 and draft PR #81 are the active execution ledger. Genuine TDD RED evidence exists: `tests/db/ip_session_history_contract_test.sh` was authored before implementation and failed because the schema21 migration pair was absent.

Fresh review found a coverage gap: `.github/workflows/ci.yml` syntax-checks all `tests/db/*.sh`, but its PostgreSQL18 database job explicitly invokes selected tests and does **not** execute `tests/db/ip_session_history_contract_test.sh`. Therefore PR #81's current generic green CI is not Task16 GREEN evidence.

Schema21 must derive history only from trusted schema17 peer/accounting lineage (`direct_naive_accounting_session_peers`), preserve forced tenant RLS and exact accounting semantics, enforce an exact 30-day retention boundary and hard server-side maximum 500 reads, provide maintenance-safe purge behavior, update migration checksums, and prove migration + rollback on PostgreSQL 18. The new integration test must be explicitly wired into CI before promotion.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override the canonical files.

## Worker / orchestration capacity

- `pv-primary`: currently executable; **Production-only**.
- `TrPaqet`: connected; assignment = Task13 final HTTP1/HTTP2 rehearsal when executable.
- `pv-worker-main`: assignment when executable = Task16 schema21 TDD/PostgreSQL18 lane.

Do not move development/testing onto Production.

## Immediate execution order

1. Keep #64 draft despite all exact-head GitHub gates being green.
2. Run the final exact-head Task13 live rehearsal on the first executable development Worker.
3. If and only if that proof passes, merge Task13, build exact R1 artifact, take fresh encrypted Production backup + rollback snapshot, deploy and postflight.
4. In parallel, continue Task16 with a real schema21 PostgreSQL18 integration test wired into CI, then implement only enough migration/query/purge behavior to turn that test green without weakening trusted accounting/session semantics.