# PVNaive — Canonical Project Status

Last updated: 2026-09-09 12:41 Asia/Tehran

## Verified GitHub state
- Authoritative `main` currently resolves to `d774557cb36e1a4b43d65030022d53a5f18d6785` (`docs: refresh continue-here after targeted Task16 database rerun`). Latest visible main changes are documentation-only; no runtime code was integrated in this cycle.
- PR #64 Task13: OPEN / DRAFT, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; focused CI, Exact Accounting, and Pinned Forwardproxy are green. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory before merge or promotion.
- PR #81 Task16: OPEN / DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`. Exact-head runs on that SHA: Task16 Schema21 TDD `33678134359` SUCCESS, Exact Accounting `33678134326` SUCCESS, Pinned Forwardproxy `33678134350` SUCCESS, normal CI `33678134360` FAILURE.
- Normal CI job breakdown: database `102372426913` FAILURE; go `102372428869` SUCCESS; web `102372454448` SUCCESS; rehearsal and bundle skipped. The database failure is not credited as green; no new rerun result was available in this cycle.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client import/parse/connect/cleanup smoke is still required.

## Worker / coordinator truth
- Persistent-report search yielded no fresh completion receipt tied to current Task13/Task16/Karing heads. Worker-only, stale, dirty, mixed-head, and historical evidence remains uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this cycle
- Re-verified current `main`, open PRs, exact-head workflow runs, and normal-CI job breakdown.
- Confirmed Task16 remains blocked on normal-CI database job `102372426913`; no unvalidated work was integrated.
- Reconciled canonical status to the latest verified GitHub state.
- Dispatched independent next actions to Task16, Task13, Karing, and review lanes via GitHub comments.

## Next executable gates
1. Task16: create one clean exact-head repair branch if the database failure persists; fix only generic latest-schema fixtures, preserve Task15 schema20 fixtures, then run all four gates on one SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
