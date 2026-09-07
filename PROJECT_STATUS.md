# PVNaive — Canonical Project Status

Last updated: 2026-09-07 19:40 Asia/Tehran

## Verified GitHub state
- Current `main` tip from GitHub: `59e4b6444abcf7964a4c2de56810e0e89c21ecdf` (docs-only reconciliation commit). Exact-head combined status is empty; no post-merge CI green is claimed.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`. Historical CI/focused tests are supplemental only; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body cites `b96c65903e5fc314284ea777ceea236913a03842` as intended exact head. This discrepancy is unresolved. Combined status for `b96c659...` is empty in the current connector view. Preserve Task15 schema20-specific fixtures and change only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental; reproducible real-client smoke remains missing.

## Worker / coordinator truth
- Persistent reports were searched. No fresh completion receipt tied to the current PR heads was found. Worker-only, stale, dirty, or mixed-head output is uncredited.
- TrPaqet remains the documented active executable slot. Worker-local PostgreSQL 14.x is not PostgreSQL18 evidence.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified current `main`, PR #64/#81/#4 metadata, exact-head status availability, and persistent reports.
- Added fresh reconciliation/dispatch comments to PR #64, #81, and #4.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task16: reconcile API head vs PR-body head, then produce one exact SHA with normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy all SUCCESS.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
