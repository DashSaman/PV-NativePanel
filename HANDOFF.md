# PVNaive — Canonical Handoff

Last updated: 2026-09-06 21:43 Asia/Tehran

## Current truth
- `main` exact head is `91330fa83a55f7067a06bf7e9652b4b29a36824d`; this run's follow-up commits are documentation-only and no post-merge CI is credited for them.
- #64 Task13 OPEN/DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; exact CI/WS1 gates green in recorded evidence, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 PG18 `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are green; repository CI `33678134360` failed in `database` at `periodic_usage_reset_executor_test.sh` with `schema version=21, want=20`. Older or partial green runs are not sufficient.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI green, real Karing smoke still required.

## Production
No fresh command-level Production health audit or executable backup/rollback/deploy lane was available. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Persistent reports remain historical without exact-head corroboration and fresh receipts. Do not integrate worker-only output. Do not print or copy secrets. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, a fresh encrypted backup, an independent rollback state, provenance and postflight verification.

## This run
- Re-verified current `main`, open PRs, exact-head workflows and Task16 database failure.
- Rechecked persistent coordinator/worker reports; no fresh completion receipt tied to current PR heads was found.
- Added fresh bounded assignments to Task16, Task13 and Karing; all remain DRAFT/DO NOT MERGE.
- Refreshed canonical status/continuation/handoff docs on `main`.
- No runtime/schema work integrated.

## Next assignments
1. Task16: make the minimal generic latest-schema fixture correction on a new exact head, preserve Task15 schema20 fixtures, and run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on that one SHA.
2. Task13: fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with exact accounting/session receipt.
3. Karing: reproducible real client import/parse/connect/cleanup smoke.
4. Independent review: security, RLS, truthful session/accounting semantics, rollback and secret redaction.
5. Production-only lane: read-only health first; backup/rollback then staged promotion only after all release gates are green.
