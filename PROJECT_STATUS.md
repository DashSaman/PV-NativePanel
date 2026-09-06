# PVNaive — Canonical Project Status

Last updated: 2026-09-06 03:38 Asia/Tehran

This file records verified repository truth and bounded Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main` at inspection: `6108a43a9c4c2297eb1cde7606ba05a67261876e`.
- Current `main` combined status: no status rows returned; no commit workflow runs were returned for this exact head. Post-merge CI is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; published CI / WS1 Exact Accounting / WS1 Pinned Forwardproxy evidence is green, but the required fresh real HTTP/1.1 + HTTP/2 rehearsal is not complete.
- Task16: draft PR #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR base metadata still points at stale `0b921abe...`, so exact-head repository-wide all-green evidence and metadata reconciliation are not complete.
- Documentation-only PRs remain open and stale; they are not promotion authority.

## Production truth

Production remains on Task15/schema20 by the latest corroborated repository record; no Task13 or schema21 code has been deployed.

No fresh command-level Production audit was available in this run; no Production health pass is claimed.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Latest corroborated worker note: `TrPaqet` is the single active development slot for Task13; `pv-primary` and `pv-worker-main` are inactive under the one-active-host limit.
- Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts.
- No worker completion is creditable because Task13 live protocol proof and Task16 exact-head full-gate proof are still absent.

## This run — 2026-09-06 03:38 Asia/Tehran

- Verified current `main` ref, current open PRs, exact-head CI/status presence, and persistent coordinator/worker reports.
- Reconciled current truth: `main=6108a43...`; #64 and #81 remain draft / DO NOT MERGE; no fresh Production health pass is claimed.
- No worker completion was reconciled or credited.
- No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation performed.
- Canonical status refreshed to this exact head and inspection time.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. On the next executable development slot, reconstruct Task13 onto exact current main and run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Independently obtain a clean exact-head Task16 checkout, reconcile generic schema21 fixtures without modifying schema20-specific Task15 fixtures, align PR base metadata, and rerun all Task16 gates on one SHA.
4. Only if Task13 live proof and Task16 full gates pass: obtain a fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
