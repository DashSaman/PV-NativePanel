# PVNaive — Canonical Project Status

Last updated: 2026-09-06 05:38 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at inspection: `ed428b69150e6e85e21aed3a93011d5e7a7e3f3f`.
- Current `main` combined status: no status rows returned; no commit workflow runs were returned for this exact head. Post-merge CI is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI / Exact Accounting / Pinned Forwardproxy are successful, but the required fresh real HTTP/1.1 + HTTP/2 rehearsal is not complete.
- Task16: draft PR #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; PG18 TDD, Exact Accounting and Pinned Forwardproxy are documented successful, while repository-wide CI remains failing on the database job (`33626300697`). Branch body/base metadata is stale and requires reconciliation before promotion.
- Documentation-only PRs remain open/stale and are not promotion authority.

## Production truth

- No fresh command-level Production audit was available in this run; no Production health pass is claimed.
- No Task13 or schema21 code is authorized as deployed based on current evidence.
- No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- No fresh worker completion receipt was available for reconciliation in this run.
- Historical reports remain non-authoritative without exact GitHub corroboration and fresh receipts.
- Latest corroborated lane plan: `TrPaqet` for Task13 development/rehearsal; `pv-primary` Production-only when an executable Production slot is available.

## This run — 2026-09-06 05:38 Asia/Tehran

- Verified current `main` ref, open PRs, exact-head status presence, and available GitHub evidence.
- Reconciled Task16 truth: three exact-head gates are documented successful; repository-wide CI/database gate is failing, so no completion credit.
- No worker completion was credited from historical reports.
- No merge, deploy, migration, restart/reload, DB write, credential, backup or rollback mutation performed.
- Canonical status refreshed to the exact inspected main and PR heads.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Task13: reconstruct validated work onto exact current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Task16: fix generic schema21 fixture drift on a clean branch, preserve schema20-specific Task15 fixtures, align PR metadata, and rerun repository-wide CI plus the other three gates on one exact SHA.
4. Only if Task13 live proof and Task16 full gates pass: obtain a fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.