# PVNaive — Canonical Handoff

Last updated: 2026-09-08 00:41 Asia/Tehran

## Current truth
- GitHub default-branch ref endpoint returns `f1db7a894449a17598d7403d056d11815536bd78` for `main`.
- PR metadata/body is not fully aligned with that ref: #64/#81 cite older bases, and docs PR #95 records base `a5d114c9...`. Treat the discrepancy as an unresolved reconciliation blocker.
- #64 Task13 OPEN/DRAFT/mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; status endpoint is `pending` with zero statuses. Fresh current-main-derived HTTP/1.1 + HTTP/2 rehearsal is mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; body cites older implementation heads including `b96c659...`; current exact-head four-gate proof is absent.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client smoke remains pending.
- #95 docs refresh OPEN/DRAFT, head `f7f670dc48928527160b668e8a8f0f394f7fc2a5`; not production evidence.

## Worker and Production
- No fresh exact-head completion receipt was found in persistent reports or latest PR comments. Preserve disposable credentials, isolated canaries, redacted logs, backup-before-promotion, and independent rollback state.
- Historical notes identify TrPaqet as the active executable slot; this does not constitute a fresh command-level Production audit.
- No fresh command-level Production audit, encrypted backup, rollback, deploy, or postflight was available. No Production mutation occurred.

## Actions in this run
- Re-verified repository metadata, main ref, open PRs, exact-head status, and persistent reports.
- Posted fresh assignment/reconciliation comments to #64, #81, and #4.
- Updated canonical project status, continue-here, and this handoff.
- No runtime/schema/Production change was integrated.

## Next assignments
1. GitHub reconciliation: confirm one authoritative current `main` SHA and refresh stale PR body/base references before any merge decision.
2. Task16: start from authoritative current `main`, preserve Task15 schema20 fixtures, update only generic latest-schema expectations, and run normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
3. Task13: current-main-derived isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no restart/reload, and exactly-once accounting.
4. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
5. Independent review: RLS, privilege separation, retention/purge, accounting/session lineage, and secret redaction.
6. Production lane: read-only audit first; only after all gates are green create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Human blockers: authoritative main/PR reconciliation, connected Production audit/deploy access, compatible worker/tooling capacity for Task13, a real Karing client, and a clean current-main Task16 branch with four fresh gates.
