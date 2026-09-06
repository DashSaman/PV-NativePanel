# PVNaive — Canonical Handoff

Last updated: 2026-09-06 20:38 Asia/Tehran

## Current truth
- `main` inspected at `7dfc283a8cbc2f837bdae8683b2566f9dcf31d46`; exact-head workflow lookup returned no runs for this docs-only head; do not credit post-merge CI for docs-only updates.
- #64 Task13 OPEN/DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; exact CI/WS1 gates green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 PG18 `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are green; repository CI `33678134360` failed in `database` at `periodic_usage_reset_executor_test.sh` with `schema version=21, want=20`. Job logs show migration, health, backup/restore and related gates passed before the generic latest-schema fixture assertion. Older or partial green runs are not sufficient.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI green, real Karing smoke still required.

## Production
No fresh command-level Production health audit or executable backup/rollback/deploy lane was available. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Persistent reports remain historical without exact-head corroboration and fresh receipts. Do not integrate worker-only output. Do not print or copy secrets. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, a fresh encrypted backup, an independent rollback state, provenance and postflight verification.

## This run
- Verified current `main`, open PRs, exact-head workflows and Task16 database failure.
- Read the database job log and isolated the remaining failure to the generic periodic-usage schema expectation after preceding gates passed.
- Reconciled persistent coordinator/worker reports without crediting stale completion claims.
- Added fresh bounded assignments to Task16, Task13 and Karing; all remain DRAFT/DO NOT MERGE.
- Refreshed canonical status/continuation/handoff docs on `main`.
- No runtime/schema work integrated.

## Next assignments
1. Task16: make the minimal generic latest-schema fixture correction on a new exact head, preserve Task15 schema20 fixtures, and run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on that one SHA.
2. Task13: fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with exact accounting/session receipt.
3. Karing: reproducible real client import/parse/connect/cleanup smoke.
4. Independent review: security, RLS, truthful session/accounting semantics, rollback and secret redaction.
5. Production-only lane: read-only health first; backup/rollback then staged promotion only after all release gates are green.
