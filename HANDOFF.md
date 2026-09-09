# PVNaive — Canonical Handoff

Last updated: 2026-09-09 10:40 Asia/Tehran

## Current truth
- Latest visible `main` head is `b38af6c612a0900f72fb809bb94baffff0e3d8f7`; this cycle added a docs-only status reconciliation as commit `ecaf9d5b7490664cda1384281adad6c394b96658`. No runtime code has been integrated.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; focused exact-head gates are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still required.
- #81 Task16 remains OPEN/DRAFT at API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; TDD/Accounting/Forwardproxy runs `33678134359`/`33678134326`/`33678134350` are green, normal CI `33678134360` is red in database job `102358503557`.
- Database failure is precise: `tests/db/periodic_usage_reset_executor_test.sh` reports `ERROR: schema version=21, want=20` after PostgreSQL 18 migration and safety gates pass. Targeted rerun of job `102358503557` was requested; final result was not observable at handoff, so it is not credited.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains missing.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified main, open PRs, exact-head workflow runs/jobs, and failure logs.
- Confirmed the exact Task16 generic-fixture mismatch from the database log.
- Re-ran only the failed Task16 database job; no green credit is assigned until the new attempt is observed.
- Updated canonical documentation. No unvalidated code was integrated; no merge or deploy was performed.

## Next assignments
1. Task16: observe targeted rerun; if red, repair only generic latest-schema fixtures on a clean exact-head branch, preserve Task15 schema20 fixtures, then run all four gates on one SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 exact-head CI closure, and connected Production audit/deploy access.
