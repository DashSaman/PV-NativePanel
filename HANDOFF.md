# PVNaive — Canonical Handoff

Last updated: 2026-09-10 02:42 Asia/Tehran

## Current truth
- Latest verified `main` before this documentation reconciliation was `158b8a7efc70fa70332b4ba82d4d625278fa0628`; this cycle adds documentation only. Combined status for that exact head is empty; no fresh post-merge CI is claimed.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal is required.
- #81 Task16 remains OPEN/DRAFT at `3c4310335ab4907d28bac995bba1be3545e14f6e`; generic latest-schema fixture failures and intermittent pinned-build failures remain unresolved. Do not transfer green credit across heads.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains missing.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 are stale or based on older main snapshots and are not current truth without exact-base validation.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified current main, open PRs, exact-head status, and persistent reports.
- Posted fresh exact-head execution assignments to PRs #81, #64, and #4.
- Reconciled `PROJECT_STATUS.md` and this handoff to the authoritative current-main checkpoint.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.

## Next assignments
1. Task16: create/verify one clean branch head, correct only generic latest-schema fixture/DB expectations, preserve Task15 schema20 fixtures, and rerun all four gates on one exact SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 repository-wide CI closure, and connected Production audit/deploy access.
