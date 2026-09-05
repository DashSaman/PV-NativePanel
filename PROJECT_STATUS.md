# PVNaive — Canonical Project Status

Last updated: 2026-09-05 17:49 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` before this documentation commit: `e16592496a26a275148a7ff66b42d2d0328f94e3`.
- No workflow run or commit status was associated with that exact main SHA; post-merge CI for that head is not credited.
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
- After switching to the installed compatible Go 1.26.3 at `/opt/pvnetwork/go1.26.3/bin/go`, focused Task13 tests passed: `TASK13_FORWARDPROXY_SESSION_CONTROL=PASSED` and `TASK13_SESSION_CONTROL_PERMISSIONS=PASSED`; `go test -race ./internal/sessionkill ./internal/sessioncontrol ./internal/httpapi` also passed.
- No fresh worker completion is currently creditable because no full Task13 HTTP/1.1 + HTTP/2 protocol rehearsal or exact-head CI proof exists.
- Persistent `AGENT_HANDOFF.md` on the Task13 checkout is historical (last update 2026-08-27) and must not drive promotion decisions.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Run the fresh real HTTP/1.1 + HTTP/2 rehearsal outside Production on the clean Task13 checkout, then attach complete receipts.
3. Obtain a clean exact-head Task16 checkout, reconcile the generic schema21-aware fixture path without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

## This run — 2026-09-05 17:49 Asia/Tehran

- Verified `main=e16592496a26a275148a7ff66b42d2d0328f94e3`; no exact-head workflow run/status.
- Verified open PR inventory: #81, #64, docs PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 and older #4.
- Verified clean Task13 checkout at exact head `3fc14825...`; shell syntax and `git diff --check` passed.
- Resolved the earlier toolchain blocker by using Go 1.26.3 from `/opt/pvnetwork/go1.26.3/bin/go` in the isolated worker checkout.
- Focused Task13 shell tests and race tests passed; the real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Verified only `TrPaqet` is connected through SentinelX; no fresh Production command-level audit or mutation was performed.
- No merges, deployments or other Production mutations performed.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
