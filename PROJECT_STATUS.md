# PVNaive — Canonical Project Status

Last updated: 2026-09-07 22:39 Asia/Tehran

## Verified GitHub state
- Current `main` tip from GitHub: `45aa9b7beb6c5c1d3e31ff482314047616e66556` (latest ref verification). Exact-head combined status is empty; no post-merge CI green is claimed.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; historical CI/focused tests are supplemental only; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body references older heads and is not current proof. Current exact-head four-gate evidence is absent. Preserve Task15 schema20-specific fixtures and change only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental; reproducible real-client smoke remains missing.

## Worker / coordinator truth
- Persistent reports and latest PR receipts were rechecked. No fresh completion receipt tied to the current PR heads was found. Worker-only, stale, dirty, or mixed-head output is uncredited.
- The documented active executable slot remains TrPaqet; worker-local PostgreSQL 14.x is not PostgreSQL18 evidence. Tooling/capacity limits remain active.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified current `main` ref, open PR metadata, exact-head status availability, latest PR discussion state, and persistent-report truth.
- Posted fresh autonomous dispatch/reconciliation comments to PR #64, #81, and #4.
- Updated canonical project status, continue-here, and handoff files.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task16: create one clean current-main-derived head; reconcile branch/body/head; update only generic latest-schema expectations; run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on one exact SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
