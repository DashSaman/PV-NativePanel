# PVNaive — Canonical Handoff

Last updated: 2026-09-07 19:40 Asia/Tehran

## Current truth
- GitHub `main` verified at `59e4b6444abcf7964a4c2de56810e0e89c21ecdf` before this documentation update; this run added only a docs reconciliation commit. Exact-head combined status is absent; no post-merge CI green is claimed.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal is mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body cites `b96c65903e5fc314284ea777ceea236913a03842`, creating an unresolved head discrepancy. Combined status for `b96c659...` is empty.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client smoke remains pending.

## Worker and Production
- No fresh exact-head completion receipt was found in persistent reports. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- No fresh command-level Production audit, backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified current main, PR metadata, exact-head status availability, and persistent reports.
- Posted fresh reconciliation/dispatch comments to #64, #81, and #4.
- Updated canonical project status and this handoff.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16: reconcile branch/API head mismatch, then run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on one exact SHA; preserve Task15 schema20 fixtures.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
5. Production lane: read-only audit first; after all gates are green, create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: connected Production audit/deploy access, compatible worker/tooling capacity for Task13, a real Karing client, and Task16 head reconciliation.
