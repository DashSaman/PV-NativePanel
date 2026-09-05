# PVNaive — Canonical Project Status

Last updated: 2026-09-05 18:42 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `441ea962f457d2d16f488a64ca892500957c72fb`.
- No workflow run or commit status is associated with that exact main SHA; post-merge CI for that head is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; focused tests are supplemental only. Fresh protocol rehearsal is blocked before assertions by PostgreSQL incompatibility: `unrecognized parameter "security_invoker"`.
- Task16: draft PR #81, current GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body still documents stale historical head `b96c659...`, and base metadata is stale. Dedicated evidence is historical; exact-head repository-wide promotion evidence is not complete.
- Docs-only PRs remain open and stale relative to current main.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Only one SentinelX host is currently connected (`TrPaqet`). No fresh command-level Production audit was executed in this run and no Production health claim is credited. No external HTTPS health claim is made.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Connected SentinelX hosts: `TrPaqet` only in this run.
- Clean Task13 rehearsal checkout: `/opt/pvnaive-task13-run11`, detached at exact head `3fc14825...`, clean worktree.
- Validation: `git diff --check` passed; discovered shell scripts passed `bash -n`.
- Focused Task13 tests and `go test -race ./internal/sessionkill ./internal/sessioncontrol ./internal/httpapi` previously passed with Go 1.26.3.
- Fresh full rehearsal rerun on this host stopped at PostgreSQL setup with `security_invoker` unsupported; no protocol assertions or accounting proof were reached.
- No worker completion is creditable because the full HTTP/1.1 + HTTP/2 rehearsal and exact-head CI proof are still absent.
- Persistent handoff on the checkout is historical and must not drive promotion decisions.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Fix or isolate the Task13 rehearsal environment incompatibility without weakening security semantics, then run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Obtain a clean exact-head Task16 checkout, reconcile the generic schema21-aware fixture path without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

## This run — 2026-09-05 18:42 Asia/Tehran

- Verified `main=441ea962...`; no exact-head workflow run/status.
- Verified open PR inventory; #64 and #81 remain draft/non-mergeable for truthful reasons.
- Verified one connected SentinelX host (`TrPaqet`) and clean Task13 checkout at exact head `3fc14825...`.
- Re-ran the real Task13 rehearsal; it failed before assertions at PostgreSQL setup with `ERROR: unrecognized parameter "security_invoker"` and return code 3.
- Added a fresh checkpoint to PR #64 documenting the failure and next action.
- No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation performed.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.