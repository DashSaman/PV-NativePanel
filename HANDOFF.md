# PVNaive — Canonical Handoff

Last updated: 2026-09-09 15:39 Asia/Tehran

## Current truth
- Latest verified `main` before this documentation update was `82d0f0156ea7b763c28f66d7faa13f32418fdf43`; this cycle adds documentation only. No runtime code was integrated.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still required.
- #81 Task16 remains OPEN/DRAFT at API head `3c4310335ab4907d28bac995bba1be3545e14f6e`. Exact-head Task16 TDD `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are SUCCESS; repository-wide CI `33678134360` is FAILURE. Do not merge.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains missing.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified current main, open PRs, current exact-head workflow runs, and persistent reports.
- Confirmed Task16 has three green dedicated gates and a red repository-wide CI run on the same API head.
- Updated canonical documentation on main; no unvalidated code was integrated.

## Next assignments
1. Task16: on a clean branch, correct only the remaining generic latest-schema fixture/DB expectation, preserve Task15 schema20 fixtures, and rerun all four gates on one exact SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 repository-wide CI closure, and connected Production audit/deploy access.
