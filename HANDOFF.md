# PVNaive — Canonical Handoff

Last updated: 2026-09-06 18:42 Asia/Tehran

## Current truth
- `main` start: `e76147775e9c99f878e23bfc252cfa0b5a45efac`; exact-head combined status empty.
- #64 Task13 DRAFT `3fc14825e1b164bad558decaef47f56b792e81af`; fresh HTTP/1.1 + HTTP/2 rehearsal required.
- #81 Task16 DRAFT `3c4310335ab4907d28bac995bba1be3545e14f6e`; PG18/Exact Accounting/Pinned Forwardproxy succeeded, normal CI failed at `periodic_usage_reset_executor_test.sh` (`schema version=21, want=20`); failed jobs re-run and not yet credited.
- #4 Karing DRAFT `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real Karing smoke required.

## Production
No fresh command-level audit was executable. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Persistent reports are historical without exact-head corroboration and fresh receipts. Do not integrate worker-only output. Promotion requires all exact-head gates green, fresh encrypted backup, rollback state, provenance and postflight verification.

## This run
- Inspected GitHub main, PRs, exact-head workflows/jobs/logs and persistent reports.
- Re-ran failed database jobs for CI run `33678134360`.
- Posted assignments to #81, #64 and #4; all remain DRAFT / DO NOT MERGE.
- Updated canonical status/continuation/handoff documentation; no runtime/schema work integrated.

## Next assignments
1. Task16: observe rerun; if needed fix only generic latest-schema expectations, preserve schema20 Task15 fixtures, then run four gates on one exact SHA.
2. Task13: fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with exact accounting/session receipt.
3. Karing: reproducible real client import/parse/connect/cleanup smoke.
4. Independent review: security, RLS, accounting and rollback.
5. Production-only: read-only health first; backup/rollback then promotion only after gates are green.
