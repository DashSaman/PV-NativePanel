# PVNaive — Canonical Project Status

Last updated: 2026-09-06 12:39 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `dba9bceea003ab0f6422944cb6e313477c9122c6` (verified from GitHub branch ref).
- Exact `main` combined status: no status rows returned; exact-head workflow runs: none. Post-merge CI is not credited.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused historical gates and local race/shell checks are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal is still incomplete; current PR metadata still references an older base SHA.
- Task16: draft PR #81, head `3c4310335ab4907d28bac995bba1be3545e14f6e`, stale base/body references. Dedicated historical gates exist, but no fresh exact-head full-green proof is present and combined status is empty.
- PR #4 (Karing export): draft, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; no reproducible real Karing client smoke evidence is attached.
- Documentation-only PRs #91–#95 remain open/stale and are not promotion authority; canonical docs are updated directly on `main` in this run.

## Production truth

- No fresh command-level Production audit was executable in this run. Historical read-only evidence in persistent reports/PR comments is not re-credited as a fresh pass here.
- No Task13, Task16 or PR #4 code is authorized as deployed from current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Persistent reports were searched; they are historical unless corroborated by exact GitHub state and fresh receipts.
- No fresh worker completion receipt tied to the current heads was available for reconciliation.
- Current bounded assignments: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 generic fixture correction; independent worker → regression/security review; `pv-primary` → Production-only when executable access is available.
- Persistent evidence also records a one-active-host SentinelX limit; connected workers can be inactive and therefore cannot be treated as executable.

## This run — 2026-09-06 12:39 Asia/Tehran

- Re-verified current repository branch ref, open PRs, exact main/PR status responses and persistent coordinator/worker reports.
- Confirmed no new validated worker completion or fresh Production receipt.
- Corrected canonical documentation to the verified `main` commit `dba9bce...`; no runtime/schema work integrated.
- No merge/deploy/migration/restart/reload/DB write/credential/backup/rollback mutation performed.

## Immediate execution order

1. Keep #64, #81 and #4 draft / DO NOT MERGE until their specific evidence gates are complete.
2. Task16: reconcile the branch to current main, fix only generic latest-schema/RLS expectations, preserve schema20-specific Task15 fixtures, and rerun all four exact-head gates.
3. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production using PostgreSQL18 and a compatible Go toolchain.
4. PR #4: obtain one reproducible real Karing client smoke with version/platform/import/connect evidence.
5. Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
6. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
