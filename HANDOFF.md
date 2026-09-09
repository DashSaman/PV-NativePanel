# PVNaive — Canonical Handoff

Last updated: 2026-09-09 05:39 Asia/Tehran

## Current truth
- Verified GitHub `main` latest commit is `99d1cb2baf0bff8426827df2bf06c0830b19ba17`; this cycle's status reconciliation is now recorded on top. No fresh post-merge CI is claimed for the docs-only head.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI `33623363327`, Exact Accounting `33623363299`, and Pinned Forwardproxy `33623363389` are SUCCESS. Fresh HTTP/1.1 + HTTP/2 rehearsal is still mandatory.
- #81 Task16 remains OPEN/DRAFT at API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; candidate `b96c65903e5fc314284ea777ceea236913a03842` has TDD `33626300588`, Exact Accounting `33626300594`, and Pinned Forwardproxy `33626300589` SUCCESS, but normal CI `33626300697` is red in the database job. The PR head/candidate drift is unresolved.
- Task16's concrete blocker is the generic `tests/db/periodic_usage_reset_executor_test.sh` expectation `schema version=21, want=20`; repair only the generic latest-schema assertion and keep Task15 schema20 fixtures pinned.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI `33209239812` is SUCCESS, but reproducible real-client smoke is pending.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified current main, open PRs, exact-head workflows, and persistent reports.
- Updated `PROJECT_STATUS.md` with the current GitHub and worker truth.
- Posted fresh execution dispatches to #81, #64, and #4; no unvalidated code was integrated.

## Next assignments
1. Task16: fix the generic latest-schema fixture on a clean branch from current `main`, reconcile exact head, and rerun all four gates on one SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 exact-head CI closure, and connected Production audit/deploy access.
