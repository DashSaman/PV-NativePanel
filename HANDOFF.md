# PVNaive — Canonical Handoff

Last updated: 2026-09-09 04:42 Asia/Tehran

## Current truth
- Verified GitHub `main` is `8ac4f3e8935c24cb95085a6b04fffb6ce1b2cea4`; main docs CI `34284651112` succeeded, but it does not prove live rehearsal or Production health.
- #64 Task13 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; focused gates are green, fresh HTTP/1.1 + HTTP/2 rehearsal is still mandatory.
- #81 Task16 remains OPEN/DRAFT at `3c4310335ab4907d28bac995bba1be3545e14f6e`. Exact-head specialized gates are green: Task16 TDD `33678134359`, Exact Accounting `33678134326`, Pinned Forwardproxy `33678134350`. Normal CI `33678134360` is red in database job `101509296474` because `tests/db/periodic_usage_reset_executor_test.sh` still expects schema20 after migration to schema21. Do not credit this head as fully green.
- #4 Karing remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains pending.

## Worker and Production
- No fresh exact-head worker completion receipt was found. Historical notes identify TrPaqet as the active executable slot; other workers are inactive or upgrade-required. This is not a fresh Production audit.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, rollback snapshot, staged deploy, or postflight was available. No Production mutation occurred.
- Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, independent rollback state, and truthful accounting/session lineage.

## Actions in this cycle
- Re-verified current main, open PRs, exact-head workflow runs/jobs/logs, and persistent report state.
- Confirmed Task16 specialized gates green but repository-wide CI red only on the remaining generic schema20 expectation in `periodic_usage_reset_executor_test.sh`.
- Updated `PROJECT_STATUS.md` and this handoff with exact run IDs and blocker detail.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16: correct only the generic latest-schema expectation in `tests/db/periodic_usage_reset_executor_test.sh`; preserve Task15 schema20-specific fixtures; rerun all four gates on the resulting exact SHA.
2. Task13: run isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all exact-head gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 exact-head CI closure, and connected Production audit/deploy access.