# PVNaive — Canonical Handoff

Last updated: 2026-09-09 14:38 Asia/Tehran

## Current truth
- Latest verified `main` before this documentation update was `82d0f0156ea7b763c28f66d7faa13f32418fdf43`; this cycle adds documentation only. No runtime code was integrated.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI, Exact Accounting, and Pinned Forwardproxy are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still required.
- #81 Task16 remains OPEN/DRAFT at API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; recorded candidate `b96c65903e5fc314284ea777ceea236913a03842` is not the API head. Dedicated gates are green on the recorded candidate; normal CI `33626300697` previously failed and its failed jobs were re-run in this cycle, with result not yet verified.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains missing.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified current main, open PRs, exact-head workflow evidence, and persistent reports.
- Re-ran failed jobs for Task16 normal CI run `33626300697`; no result was credited before verification.
- Updated canonical documentation on main; latest cycle commit is `cd3e3caf0b3df64dbbcc848bbd670e5ceb9ecf30` for `PROJECT_STATUS.md`.
- No unvalidated code was integrated; no merge or deploy was performed.

## Next assignments
1. Task16: verify rerun; if database CI remains red, reconcile one exact implementation SHA, repair only generic latest-schema fixtures on a clean branch, preserve Task15 schema20 fixtures, and run all four gates on that same SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 database CI closure, and connected Production audit/deploy access.
