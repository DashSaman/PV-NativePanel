# PVNaive — Canonical Project Status

Last updated: 2026-09-02

This file is current repository + Production truth. Historical stage notes and worker reports are evidence only; exact GitHub state, exact-head CI and fresh observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3` (PR #83 merged).
- Task13: draft PR #64, branch `lead/task13-reconstruct-62573fee`, exact head `3fc14825e1b164bad558decaef47f56b792e81af`.
- Exact-head Task13 workflows: CI `33623363327` **SUCCESS**; WS1 Exact Accounting `33623363299` **SUCCESS**; WS1 Pinned Forwardproxy `33623363389` **SUCCESS**.
- Fresh exact-head worker validation: `git diff --check`, focused race tests for sessionkill/sessioncontrol/httpapi, forwardproxy session-control and permission contracts all **PASS**. This does not replace the real protocol rehearsal.
- Task13: **IN PROGRESS / draft / sole remaining merge gate is fresh real HTTP1+HTTP2 live proof**.
- Task12 active-session management: **DONE / Production**, schema17.
- Task14 concurrent-session limit: **DONE / Production**, schema19.
- Task15 simultaneous unique-IP limit: **DONE / Production**, schema20.
- Security Task35 / BUG-001/002/003: **DONE in main**.
- Task16 bounded IP/session history: **IN PROGRESS / draft PR #81 / issue #79 / schema21 implementation + PG18 proof in progress**.
- Task16 current head: `b96c65903e5fc314284ea777ceea236913a03842`.
- On prior Task16 head `a366c359...`, Task16 Schema21 TDD, Exact Accounting and Pinned Forwardproxy were green while generic CI failed in the database job. The failure exposed stale generic newest-schema=20 fixtures, not a reason to weaken migration safety.
- Stale generic fixtures now advanced to schema21: auth refresh, DB health, backup/restore, periodic reset executor. Schema20-specific Task15 fixtures remain intentionally schema20-pinned.

No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

This cycle `pv-primary` is connected but not executable because the SentinelX Free plan permits only one active host and the active slot is currently on development worker `TrPaqet`. Therefore no fresh command-level Production health claim is made in this cycle.

No Production mutation, deploy, restart, reload, migration, DB write or credential change was performed.

## Task13 — exact live-session kill

All exact-head GitHub gates and fresh focused worker tests are green on `3fc14825...`. The sole merge gate remains a fresh real HTTP/1.1 + HTTP/2 rehearsal on that exact tree proving target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

`TrPaqet` is currently executable and owns this rehearsal. Historical real-protocol evidence is supporting evidence only; it does not satisfy the fresh exact-head gate.

Only after live proof may Task13 be merged. Deployment additionally requires a fresh Production read-only audit, encrypted backup + rollback snapshot, exact R1 artifact provenance and complete postflight.

## Task16 — bounded IP/session history / schema21

Issue #79 and draft PR #81 are the active execution ledger. Genuine TDD RED preceded implementation. Dedicated PostgreSQL18 evidence has already proven the schema21 migration contract on an earlier implementation head, including forced RLS, maintenance/app privilege separation, bounded reads, confirmation-gated purge and clean empty rollback to schema20.

The current branch is being reconciled against repository-wide database contracts. Generic tests that migrate to the newest repository schema must now expect 21; intentionally frozen schema20 fixtures must remain pinned. Final promotion requires the exact current head to pass normal CI, Task16 Schema21 TDD, WS1 Exact Accounting and WS1 Pinned Forwardproxy.

Schema21 must derive history only from trusted peer/accounting lineage, preserve exact accounting semantics, enforce the exact 30-day boundary and hard server-side maximum 500 reads, keep destructive purge explicit and maintenance-only, and never accept client-provided IP/session truth.

## Persistent worker reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers last updated 2026-08-27 and contain no newer Task13/Task16 completion. They do not override the canonical files.

## Worker / orchestration capacity

- `TrPaqet`: currently executable development worker; assignment = Task13 final HTTP1/HTTP2 rehearsal, then independent Task16 verification if capacity remains.
- `pv-worker-main`: connected but inactive under one-active-host limit; assignment when executable = Task16 schema21 CI/PG18 lane.
- `pv-primary`: connected but inactive under one-active-host limit; **Production-only** and must be activated for fresh pre-deploy audit/backup/deployment.

Do not move development/testing onto Production.

## Immediate execution order

1. Keep #64 draft despite exact-head GitHub and focused worker gates being green.
2. Complete the final real HTTP1/HTTP2 Task13 rehearsal on exact `3fc14825...`.
3. If and only if that proof passes, merge Task13; before deployment reactivate Production Primary, perform fresh health audit, build/verify exact R1 artifact, create encrypted backup + rollback snapshot, then deploy and postflight.
4. In parallel, drive Task16 current head to all-green repository + dedicated PG18 gates; do not deploy schema21 before exact-head gates and Production backup/rollback readiness.