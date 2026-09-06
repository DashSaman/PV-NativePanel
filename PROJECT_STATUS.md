# PVNaive — Canonical Project Status

Last updated: 2026-09-06 14:42 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at start of this run: `072d6fb9ddd74fa916516fe757aecb2ae116a36b` (verified from GitHub recent commits; docs-only head).
- Exact `main` combined status/workflow evidence was not returned in this run; post-merge CI is not credited.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal remains incomplete; PR metadata is stale.
- Task16: draft PR #81, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated runs `33678134359`, `33678134326`, `33678134350` are SUCCESS, but CI run `33678134360` is FAILURE. Failed jobs were re-run from GitHub during this run; result is pending and is not yet credited.
- PR #4 (Karing export): draft, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; no reproducible real Karing client smoke evidence is attached.
- Documentation-only PRs #91–#95 remain open/stale and are not promotion authority; canonical docs are updated directly on `main`.

## Production truth

- No fresh command-level Production audit was executable in this run. Historical read-only evidence is not re-credited as a fresh pass.
- No Task13, Task16 or PR #4 code is authorized as deployed from current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Persistent reports were searched; they are historical unless corroborated by exact GitHub state and fresh receipts.
- No fresh worker completion receipt tied to the current heads was available for reconciliation.
- Current bounded assignments: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 CI failure/fix verification; independent worker → regression/security review; `pv-primary` → Production-only when executable access is available.
- Persistent evidence records a one-active-host SentinelX limit; connected workers can be inactive and therefore cannot be treated as executable.

## This run — 2026-09-06 14:42 Asia/Tehran

- Re-verified current main history, open PRs, Task16 exact-head workflow state and persistent coordinator/worker reports.
- Re-ran only the failed Task16 CI jobs (`33678134360`) from GitHub; pending result is not credited until completion is observed.
- Confirmed no new validated worker completion or fresh Production receipt.
- Updated canonical documentation directly on `main`; no runtime/schema work integrated.
- No merge/deploy/migration/restart/reload/DB write/credential/backup/rollback mutation performed.

## Immediate execution order

1. Observe the Task16 failed-job rerun and inspect the database job result/log before any merge decision.
2. Keep #64, #81 and #4 draft / DO NOT MERGE until their specific evidence gates are complete.
3. Task16: if the rerun fails, fix only the failing generic latest-schema/RLS expectation, preserve schema20-specific Task15 fixtures, and rerun all four gates on one exact published SHA.
4. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
5. PR #4: obtain one reproducible real Karing client smoke with version/platform/import/connect evidence.
6. Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
7. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
