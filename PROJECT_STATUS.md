# PVNaive — Canonical Project Status

Last updated: 2026-09-05 12:43 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at inspection start: `fa0c7ae8ed11bf8cbfa5b8489f8265b5badf8b8d`.
- `main` has no visible workflow runs or status rows; post-merge CI is not proven.
- Task13: draft PR #64, exact GitHub head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact GitHub head `b96c65903e5fc314284ea777ceea236913a03842`; Task16 Schema21 TDD, WS1 Exact Accounting and WS1 Pinned Forwardproxy are SUCCESS, while generic CI run `33626300697` failed in `database` and was re-run during this turn; fresh rerun result is pending.
- No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

No fresh command-level Production audit was credited in this run because `pv-primary` was inactive under the SentinelX Free-plan one-active-host limit. Therefore no new readiness/liveness/service-state claim is made here.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Three SentinelX hosts are connected: `TrPaqet`, `pv-primary`, and `pv-worker-main`.
- The only executable lane available in this run was `pv-worker-main`; the other two hosts were blocked by the one-active-host plan limit.
- The persistent checkout at `/workspace/pvnaive-main` is dirty and stale: branch `main`, HEAD `d8c85225ab87`, 11 unstaged changes, 3 untracked files, and 149 commits behind the expected remote state.
- Fresh validation on that checkout: `git diff --check` PASS; shell syntax scan FAIL at `tests/db/auth_refresh_reuse_test.sh:51` with `unexpected EOF while looking for matching ')'`.
- No fresh worker completion was creditable.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Consume the fresh generic-CI rerun result on PR #81; if it fails, repair only schema21-aware generic fixtures on a clean exact-head branch, preserving Task15 schema20 fixtures.
3. Obtain a clean exact-head development checkout and run the final exact-head Task13 HTTP/1.1 + HTTP/2 rehearsal outside Production with PostgreSQL 18 and compatible Go/jq tooling.
4. Re-run all four Task16 promotion gates on the same exact SHA; do not mix historical receipts.
5. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.
