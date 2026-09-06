# PVNaive — Canonical Project Status

Last updated: 2026-09-06 18:42 Asia/Tehran

## Verified state
- `main` at start: `e76147775e9c99f878e23bfc252cfa0b5a45efac`; no combined status rows for this exact head, so post-merge CI is not credited.
- PR #64 Task13: DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh HTTP/1.1 + HTTP/2 rehearsal pending.
- PR #81 Task16: DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 PG18, Exact Accounting and Pinned Forwardproxy succeeded; repository CI `33678134360` failed in `periodic_usage_reset_executor_test.sh` with `schema version=21, want=20`. Failed jobs were re-run; outcome pending and not credited.
- PR #4 Karing: DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI 402 succeeded; real Karing smoke pending.

## Production truth
No fresh command-level Production audit was executable. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent reports are historical unless corroborated by exact GitHub state and fresh receipts. No fresh completion receipt tied to the current PR heads was available. One-active-host limitations mean connected workers may be inactive.

## This run
- Inspected main, PRs, exact-head CI, failed database logs and persistent reports.
- Re-ran only failed jobs for CI run `33678134360`.
- Posted bounded assignments to #81, #64 and #4; all remain DRAFT / DO NOT MERGE.
- No runtime/schema work integrated.

## Next gates
1. Observe Task16 rerun; fix only generic latest-schema expectations if still failing, preserving schema20 Task15 fixtures, then run all four gates on one SHA.
2. Run fresh Task13 HTTP/1.1 + HTTP/2 rehearsal outside Production.
3. Obtain reproducible real Karing smoke.
4. Only after all gates pass, create fresh encrypted backup + rollback state and consider promotion.

Never claim completion from stale reports or partial evidence.
