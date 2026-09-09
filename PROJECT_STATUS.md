# PVNaive — Canonical Project Status

Last updated: 2026-09-09 15:39 Asia/Tehran

## Verified GitHub state
- Authoritative `main` currently resolves to `82d0f0156ea7b763c28f66d7faa13f32418fdf43`; this cycle makes documentation-only changes and integrates no runtime code.
- PR #64 Task13: OPEN / DRAFT, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; CI, Exact Accounting, and Pinned Forwardproxy are SUCCESS. Fresh real HTTP/1.1 + HTTP/2 rehearsal is still mandatory before merge or promotion.
- PR #81 Task16: OPEN / DRAFT, current API head `3c4310335ab4907d28bac995bba1be3545e14f6e`. Exact-head runs on that SHA: Task16 Schema21 TDD `33678134359` SUCCESS; WS1 Exact Accounting `33678134326` SUCCESS; WS1 Pinned Forwardproxy `33678134350` SUCCESS; repository-wide CI `33678134360` FAILURE. Do not credit this PR green. Latest known failure path is generic latest-schema fixture/DB validation; preserve schema20-specific Task15 fixtures.
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
- Confirmed Task16 exact-head CI state from GitHub workflow runs; normal CI remains red.
- Reconciled canonical documentation to the latest verified state and prepared independent next actions for Task16, Task13, Karing, and review lanes.

## Next executable gates
1. Task16: fix only the remaining generic latest-schema fixture/DB expectation on a clean branch, preserve Task15 schema20 fixtures, then run normal CI + Task16 TDD + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
