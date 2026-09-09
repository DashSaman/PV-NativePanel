# PVNaive — Canonical Project Status

Last updated: 2026-09-09 21:43 Asia/Tehran

## Verified GitHub state
- Authoritative `main` currently resolves to `0b921abe9b2bd1d827023f494fda11a407fe34d3`; latest commit is documentation-only. CI run `33623286003` for this exact SHA completed SUCCESS. No runtime code was integrated in this cycle.
- PR #64 Task13: OPEN / DRAFT, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; CI `33623363327`, Exact Accounting `33623363299`, and Pinned Forwardproxy `33623363389` are SUCCESS. Fresh real HTTP/1.1 + HTTP/2 rehearsal is still mandatory before merge or promotion.
- PR #81 Task16: OPEN / DRAFT, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 TDD `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are SUCCESS; repository-wide CI `33678134360` is FAILURE in database job `102372426913` due to `periodic_usage_reset_executor_test.sh` expecting schema 20 after schema21. Failed jobs were re-run in this cycle; result is pending and no green credit is assigned yet.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client import/parse/connect/cleanup smoke remains required.

## Worker / coordinator truth
- Persistent-report search found no fresh completion receipt tied to the current Task13/Task16/Karing heads. Worker-only, stale, dirty, mixed-head, and historical evidence remains uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this cycle
- Re-verified current `main`, open PRs, exact-head workflow evidence, and persistent coordinator/worker reports.
- Re-ran failed jobs for Task16 repository-wide CI run `33678134360`; final result is pending.
- Added fresh dispatch instructions for Task16, Task13, and Karing lanes.
- Reconciled canonical documentation to current verified state.

## Next executable gates
1. Task16: observe rerun; if still failing, correct only generic latest-schema fixture/DB expectation on a clean branch, preserve Task15 schema20 fixtures, then run all four gates on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.