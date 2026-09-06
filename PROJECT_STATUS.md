# PVNaive — Canonical Project Status

Last updated: 2026-09-06 06:43 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at inspection: `453c2fa7159014b8ffcac5545e43da8b6c736c9d`.
- Exact `main` combined status: no status rows returned; no commit workflow runs returned. Post-merge CI is not credited.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; documented exact-head CI / Exact Accounting / Pinned Forwardproxy are successful, but the required fresh real HTTP/1.1 + HTTP/2 rehearsal is incomplete. PR base/body metadata still references older main state.
- Task16: draft PR #81, head `b96c65903e5fc314284ea777ceea236913a03842`; PG18 TDD, Exact Accounting and Pinned Forwardproxy are documented successful, while repository-wide CI previously failed on the database job and PR metadata/body is stale. No fresh exact-head all-green proof was observed in this inspection.
- PR #4 (Karing export) remains open/draft pending one real Karing client smoke; no Production mutation is implied.
- Documentation-only PRs remain open/stale and are not promotion authority.

## Production truth

- No fresh command-level Production audit was available in this run; no Production health pass is claimed.
- No Task13 or schema21 code is authorized as deployed based on current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- No fresh worker completion receipt was available for reconciliation in this run.
- Historical reports remain non-authoritative without exact GitHub corroboration and fresh receipts.
- Latest corroborated lane plan: `TrPaqet` for Task13 development/rehearsal; PostgreSQL18-capable worker for Task16; independent worker for regression/security review; `pv-primary` Production-only when an executable Production slot is available.

## This run — 2026-09-06 06:43 Asia/Tehran

- Verified current `main`, open PRs, exact-head status presence and current GitHub evidence.
- Confirmed no fresh CI evidence for current `main` exact head.
- Reconciled that no worker completion can be credited from historical reports.
- Refreshed canonical status without integrating unvalidated runtime/schema work.
- No merge, deploy, migration, restart/reload, DB write, credential, backup or rollback mutation performed.

## Immediate execution order

1. Keep #64, #81 and #4 draft / DO NOT MERGE until their specific evidence gates are complete.
2. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Task16: fix/reconcile generic schema21 fixtures on a clean branch, preserve schema20-specific Task15 fixtures, align PR metadata and rerun all required gates on one exact SHA.
4. PR #4: obtain one real Karing client smoke before considering readiness.
5. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
