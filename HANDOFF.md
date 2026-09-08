# PVNaive — Canonical Handoff

Last updated: 2026-09-09 00:42 Asia/Tehran

## Current truth
- Verified GitHub `main` is `86bdd05eadbfa256b4a274416228a5d408e3addd` at inspection; this cycle added documentation-only reconciliation commit `f1d12fc7a529f69b979e94e608dae8f177102219`. No green post-merge CI is claimed for docs-only updates; combined status is empty.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI/Exact Accounting/Pinned Forwardproxy are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, documented candidate `b96c65903e5fc314284ea777ceea236913a03842`; specialized gates are green, while repository-wide CI database job `101289670458` failed in run `33626300697` and was re-run this cycle. The rerun result is pending verification; no promotion is allowed until all four gates are green on one exact SHA.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke remains pending.

## Worker and Production
- Persistent-report search found no fresh exact-head completion receipt. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot and other workers as inactive/upgrade-required; this is not a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this cycle
- Re-verified repository metadata, authoritative main ref, open PRs, exact-head workflow evidence, and persistent reports.
- Re-ran only the failed Task16 database job `101289670458` from run `33626300697`; result must be checked before crediting.
- Reconciled this handoff to current observed GitHub state. No runtime/schema/Production change was integrated.

## Next assignments
1. Task16: verify the database-job rerun; if green, prove all four gates on exact candidate `b96c659...`; otherwise inspect the new failure and repair on a clean current-main-derived head, preserving Task15 schema20 fixtures.
2. Task13: execute isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected worker execution for live rehearsal, a real Karing client, Task16 exact-head CI closure, and connected Production audit/deploy access.
