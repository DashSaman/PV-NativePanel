# PVNaive — Canonical Project Status

Last updated: 2026-09-05 14:41 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `4c2a2b2ec22ed5464dcde3362200bfc977cdda65`.
- Main CI: workflow run `33959857930` (`CI`, push on main) completed `success` on 2026-09-05.
- Task13: draft PR #64, exact GitHub head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e` while the PR body still names stale `b96c659...`; dedicated Task16/WS1 receipts are historical evidence, and the PR remains non-mergeable pending exact-head reconciliation and fresh repository-wide gates.
- No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

No fresh command-level Production audit was credited in this run because `pv-primary` is connected but inactive under the SentinelX Free-plan one-active-host limit. Therefore no new readiness/liveness/service-state claim is made here.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Three SentinelX hosts are connected: `TrPaqet`, `pv-primary`, and `pv-worker-main`.
- The active executable lane in this run was `pv-worker-main`; `pv-primary` is inactive under the one-active-host plan limit.
- The persistent checkout at `/workspace/pvnaive-main` is dirty and stale: branch `main`, HEAD `d8c85225ab87`, 11 unstaged changes, 3 untracked files, and 149 commits behind the expected remote state.
- Read-only validation on that checkout: `git diff --check` passed; `bash -n tests/db/auth_refresh_reuse_test.sh` failed at line 51 with `unexpected EOF while looking for matching ')'`.
- No fresh worker completion was creditable from that checkout.
- Persistent `HANDOFF.md` is stale (last updated 2026-08-30) and conflicts with current GitHub truth; it remains evidence only and must not drive promotion decisions.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Obtain a clean exact-head development checkout for Task16; reconcile the generic schema21-aware fixture path without modifying Task15 schema20-specific fixtures, then rerun all Task16 gates on one published SHA.
3. Obtain a clean exact-head development checkout for Task13 and run the final real HTTP/1.1 + HTTP/2 rehearsal outside Production with PostgreSQL 18 and compatible Go/jq tooling.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

## This run — 2026-09-05 14:41 Asia/Tehran

- Verified `main=4c2a2b2ec22ed5464dcde3362200bfc977cdda65` and main CI run `33959857930=success`.
- Verified open PR inventory includes #64, #81, docs-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85, and older PR #4.
- Verified PR #81 head/base/body mismatch and PR #64 live rehearsal blocker.
- Verified worker capacity: 3 connected, 1 active; active lane `pv-worker-main`; persistent checkout remains dirty/stale.
- Verified fresh worker validation: `git diff --check=PASS`; `bash -n tests/db/auth_refresh_reuse_test.sh=FAIL` at line 51.
- No merges, deployments or other Production mutations performed.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
