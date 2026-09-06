# PVNaive — Canonical Project Status

Last updated: 2026-09-06 08:42 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `81cc22e49d6b0b5f164f6689b4b35d5263c0b8be`.
- Exact `main` combined status: no status rows returned; exact-head workflow runs: none. Post-merge CI is not credited.
- Task13: draft PR #64, head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI/Exact Accounting/Pinned Forwardproxy runs remain historical-success evidence, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still incomplete. The PR body references older main state.
- Task16: draft PR #81, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; exact-head runs on 2026-09-05: Task16 Schema21 TDD `33678134359` SUCCESS, WS1 Exact Accounting `33678134326` SUCCESS, WS1 Pinned Forwardproxy `33678134350` SUCCESS, normal CI `33678134360` FAILURE only in `database` job. Failure is `ERROR: schema version=21, want=20` from `tests/db/periodic_usage_reset_executor_test.sh`; `go` and `web` jobs passed; `rehearsal` and `bundle` were skipped. Keep DO NOT MERGE.
- PR #4 (Karing export): draft, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; historical CI is green, but one real Karing client smoke is still required.
- Documentation-only PRs #85, #86, #87, #88, #89, #91, #92, #93, #94 and #95 remain open/stale and are not promotion authority.

## Production truth

- No fresh command-level Production audit was available in this run; no Production health pass is claimed.
- No Task13, Task16 or PR #4 code is authorized as deployed from current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Persistent reports were searched; they are historical unless corroborated by exact GitHub state and fresh receipts.
- No fresh worker completion receipt tied to the current heads was available for reconciliation.
- Bounded plan: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 generic fixture correction; independent worker → regression/security review; `pv-primary` → Production-only when executable access is available.

## This run — 2026-09-06 08:42 Asia/Tehran

- Verified repository metadata, current main commit, open PRs, exact-head status/workflow evidence and the latest Task16 database failure logs.
- Confirmed latest Task16 failure is a generic schema21 fixture expectation mismatch, not a PostgreSQL18 migration failure.
- Confirmed current main has no post-merge CI evidence.
- Reconciled that no historical worker completion can be credited.
- Refreshed canonical status without integrating unvalidated runtime/schema work.

## Immediate execution order

1. Keep #64, #81 and #4 draft / DO NOT MERGE until their specific evidence gates are complete.
2. Task16: update only the remaining generic latest-schema expectation in `tests/db/periodic_usage_reset_executor_test.sh`; preserve schema20-specific Task15 fixtures; rerun all four exact-head gates.
3. Task13: reconstruct validated work onto current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
4. PR #4: obtain one real Karing client smoke with version/platform/import/connect evidence.
5. Only after required evidence is green: obtain fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.