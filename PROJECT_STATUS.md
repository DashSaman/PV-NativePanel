# PVNaive — Canonical Project Status

Last updated: 2026-09-07 18:39 Asia/Tehran

## Verified state
- Current `main` tip from GitHub: `781f6b1149d1f15ba0e9496ff96c0445f1f401ff` (docs-only reconciliation commit). Exact-head combined status is empty; no post-merge CI green is claimed.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains required; focused/race evidence is supplemental only.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`. PR body references `b96c659...` as intended exact head; this discrepancy must be reconciled before any gate decision. Exact-head combined status for `b96c659...` is empty. Preserve Task15 schema20-specific fixtures and correct only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; real-client import/parse/connect/cleanup smoke remains missing.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
- Persistent coordinator/worker reports were searched. No fresh completion receipt tied to the current PR heads was found. TrPaqet remains the documented active executable slot; worker-local PostgreSQL 14.x is not PostgreSQL18 evidence. Worker-only, stale or dirty outputs remain uncredited.

## Actions in this run
- Re-verified GitHub main, PR #64/#81/#4, exact-head status availability, and persistent coordinator/worker reports.
- Added reconciliation/dispatch comments to PR #64, #81 and #4 with exact evidence requirements and safety boundaries.
- Updated this canonical status; no runtime/schema/Production change was integrated.

## Next gates
1. Task16: reconcile API head vs PR-body referenced head, then publish one exact SHA with normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy all SUCCESS.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.
