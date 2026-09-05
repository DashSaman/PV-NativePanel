# PVNaive — Canonical Project Status

Last updated: 2026-09-05 22:39 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at inspection: `a0def2cfc1b5489cb9e821ee8d059773169e845d`.
- Current `main` combined status: no status rows returned; post-merge CI for this exact head is not credited. No commit workflow runs were returned for this exact head.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub exact-head CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy are SUCCESS. Fresh local focused validation remains supplemental. Real HTTP/1.1 + HTTP/2 rehearsal remains blocked during PostgreSQL setup at `unrecognized parameter "security_invoker"`.
- Task16: draft PR #81, current GitHub head `b96c65903e5fc314284ea777ceea236913a03842`; dedicated PostgreSQL18 evidence is historical, while fresh exact-head repository-wide all-green evidence and metadata reconciliation are not complete.
- Docs-only PRs remain open and stale relative to current main.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Only one SentinelX host is connected (`TrPaqet`). A fresh read-only project snapshot shows the Task13 checkout `/opt/pvnaive-task13-run11` is detached at exact head `3fc14825e1b1`, dirty=false, ahead=0, behind=0. A fresh rehearsal attempt still stopped before assertions at the PostgreSQL setup incompatibility above. No Production health pass was obtained because `pv-primary` was not available.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Connected SentinelX hosts: `TrPaqet` only.
- Clean Task13 checkout: `/opt/pvnaive-task13-run11`, detached at exact head `3fc14825...`, dirty=false, ahead=0, behind=0.
- Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts.
- No worker completion is creditable because the full HTTP/1.1 + HTTP/2 rehearsal and exact-head promotion proof are still absent.

## This run — 2026-09-05 22:39 Asia/Tehran

- Verified current main, open PRs, main status/workflow presence, Task13/Task16 metadata, one connected SentinelX host, clean Task13 checkout, and persistent handoff state.
- Re-ran the real Task13 rehearsal on the clean exact-head checkout; it failed before assertions with `psql: ERROR: unrecognized parameter "security_invoker"` (return code 3).
- No worker completion was reconciled or credited.
- No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation performed.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Fix or isolate the Task13 rehearsal compatibility issue without weakening security semantics, then run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Obtain a clean exact-head Task16 checkout, reconcile generic schema21 fixtures without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
