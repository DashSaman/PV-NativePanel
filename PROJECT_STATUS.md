# PVNaive — Canonical Project Status

Last updated: 2026-09-05 17:45 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `e16592496a26a275148a7ff66b42d2d0328f94e3`.
- No workflow run or commit status is currently associated with this exact main SHA; post-merge CI for this head is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; PR body/base still reference stale history. Dedicated receipts are historical and the PR remains non-mergeable pending exact-head reconciliation and fresh repository-wide gates.
- Docs-only PRs remain open and stale relative to current main; no docs PR was merged in this run.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Only one SentinelX host is currently connected (`TrPaqet`). Therefore no fresh command-level Production audit was executed in this run and no Production health claim is credited. No external HTTPS health claim is made.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Connected SentinelX hosts: `TrPaqet` only in this run; prior connected-worker inventory is historical until re-verified.
- Clean Task13 rehearsal checkout: `/opt/pvnaive-task13-run11`, detached at exact head `3fc14825e1b164bad558decaef47f56b792e81af`, clean worktree.
- Validation on that checkout: `git diff --check` passed; all discovered shell scripts passed `bash -n`.
- Focused Task13 shell tests were attempted but blocked by the repository's vendored forwardproxy `go.mod`: local Go rejects `go 1.21.0` and the `toolchain` directive (`invalid go version '1.21.0'` / `unknown directive: toolchain`). This is an environment/toolchain compatibility blocker, not a passing test.
- No fresh worker completion is currently creditable because no full Task13 protocol rehearsal or exact-head CI proof exists.
- Persistent `AGENT_HANDOFF.md` on the Task13 checkout is historical (last update 2026-08-27) and must not drive promotion decisions.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Resolve the isolated Task13 toolchain mismatch without changing production semantics, then run the focused tests and the real HTTP/1.1 + HTTP/2 rehearsal outside Production.
3. Obtain a clean exact-head Task16 checkout, reconcile the generic schema21-aware fixture path without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

## This run — 2026-09-05 17:45 Asia/Tehran

- Verified `main=e16592496a26a275148a7ff66b42d2d0328f94e3`; no exact-head workflow run/status.
- Verified open PR inventory: #81, #64, docs PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 and older #4.
- Verified clean Task13 checkout at exact head `3fc14825...`; shell syntax and `git diff --check` passed.
- Attempted focused Task13 shell tests; blocked by vendored forwardproxy module requiring a newer Go toolchain syntax than the installed local toolchain accepts.
- Verified only `TrPaqet` is currently connected through SentinelX; no fresh Production command-level audit or mutation was performed.
- No merges, deployments or other Production mutations performed.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
