# PVNaive — Canonical Project Status

Last updated: 2026-09-05 10:40 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `58719b1f4c662b75a2db132bea6fa048838fdff6` (`docs: refresh continuation handoff for automation turn 11`).
- `main` has no visible workflow runs or status rows; post-merge CI is not proven.
- Task13: draft PR #64, exact GitHub head `3fc14825e1b164bad558decaef47f56b792e81af`; historical focused gates are supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains pending.
- Task16: draft PR #81, exact GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body/base remain stale. Exact-head workflow snapshot: Task16 Schema21 TDD SUCCESS, WS1 Exact Accounting SUCCESS, WS1 Pinned Forwardproxy SUCCESS, generic CI FAILURE (`33678134360`). Failed generic CI jobs were re-run during this turn; result is pending.
- No task becomes DONE from a local candidate, historical report or partial branch alone.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

No fresh command-level Production audit was credited in this run because `pv-primary` was inactive under the SentinelX Free-plan one-active-host limit. Therefore no new readiness/liveness/service-state claim is made here.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Three SentinelX hosts are connected: `TrPaqet`, `pv-primary`, and `pv-worker-main`.
- The active executable lane in this run was `pv-worker-main`; `TrPaqet` and `pv-primary` were inactive under the one-active-host limit.
- The persistent checkout at `/workspace/pvnaive-main` is dirty and stale: branch `main`, HEAD `d8c85225ab87`, 11 unstaged changes, 3 untracked files, and 149 commits behind the expected remote state.
- No fresh worker completion was creditable.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Wait for the fresh generic-CI rerun result on PR #81; if it fails, repair only schema21-aware generic fixtures on a clean exact-head branch, preserving Task15 schema20 fixtures.
3. Obtain a clean exact-head development checkout and run the final exact-head Task13 HTTP/1.1 + HTTP/2 rehearsal outside Production with PostgreSQL 18 and compatible Go/jq tooling.
4. Re-run all four Task16 promotion gates on the same exact SHA; do not mix historical receipts.
5. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.
