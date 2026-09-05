# PVNaive — Canonical Project Status

Last updated: 2026-09-05 15:42 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `423896d534e1c3830541fedd15a6c05fec401f59`.
- No workflow run is currently associated with this exact main SHA; post-merge CI for this head is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body/base still reference stale history. Dedicated receipts are historical and the PR remains non-mergeable pending exact-head reconciliation and fresh repository-wide gates.
- Docs-only PRs remain open and stale relative to current main; no docs PR was merged in this run.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Fresh command-level Production audit was not executed because the connected `pv-primary` host is inactive under the SentinelX Free-plan one-active-host limit. The latest visible fleet report (2026-09-05 15:43 UTC) shows `caddy-naive`, `postgresql@18-main`, `pvnaive-api`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` active on the Production host; this is inventory evidence only, not a fresh command-level health credit.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Three SentinelX hosts are connected: `TrPaqet`, `pv-primary`, and `pv-worker-main`.
- Active executable lane: `pv-worker-main`; `TrPaqet` and `pv-primary` are inactive under the one-active-host plan limit.
- Persistent checkout `/workspace/pvnaive-main` is dirty and stale: branch `main`, HEAD `d8c85225ab87`, 11 unstaged changes, 3 untracked files, and 149 commits behind expected remote state.
- Task13 candidate checkout `/workspace/task13-rebased-fce` is also not integration-clean: branch `lead/task13-real-kill-v3-20260831`, HEAD `fce39283c6449b0d1836757ee7caddb31fab9def`, 14 staged changes and 5 untracked paths.
- Fresh validation on the Task13 candidate: all discovered shell scripts passed `bash -n`; `git diff --check` passed. Focused race tests could not run because the default Go toolchain was absent; using `/workspace/tools/go/bin/go` (Go 1.25.14) with `-race` first failed because CGO was disabled, then the CGO-enabled run was started in background and has not yet returned. No completion is credited from this candidate.
- No fresh worker completion is currently creditable.
- Persistent `HANDOFF.md` remains historical evidence only and must not drive promotion decisions.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Finish the pending Task13 race-test job, then obtain a clean exact-head checkout and run the real HTTP/1.1 + HTTP/2 rehearsal outside Production.
3. Obtain a clean exact-head Task16 checkout, reconcile the generic schema21-aware fixture path without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

## This run — 2026-09-05 15:42 Asia/Tehran

- Verified `main=423896d534e1c3830541fedd15a6c05fec401f59`; no workflow run associated with this exact head.
- Verified open PR inventory: #64, #81, docs PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 and older #4.
- Verified Task13 candidate shell syntax and diff checks passed; focused race suite remains pending because toolchain/CGO execution is unresolved.
- Verified Production host inventory from latest visible fleet report, but no fresh command-level Production audit was credited because `pv-primary` is inactive.
- No merges, deployments or other Production mutations performed.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
