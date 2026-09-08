# PVNaive — Canonical Project Status

Last updated: 2026-09-08 09:38 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` ref currently returns `44b9c76b070b0da5677659ed2b90466b8f7f72cd`.
- This exact head is documentation-only. `fetch_commit_workflow_runs` returned no workflow runs; no green post-merge CI is claimed.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head status is empty and fresh current-main-derived real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; the documented candidate `b96c659...` is not the API head, so exact-head identity is unresolved. Preserve Task15 schema20-specific fixtures and change only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke is still required.
- PR #95 docs refresh remains OPEN / DRAFT and documentation-only; it is not Production evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to the current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive/upgrade-required while the slot is held elsewhere. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main` ref, open PR metadata, exact-head CI/run availability, and persistent worker/coordinator reports.
- Reconciled this status file to the actual current `main` SHA.
- Reviewed Task13, Task16, Karing, and docs-only PR claims; none was accepted as ready for merge or Production because required fresh exact-head/live evidence is missing.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task13: rebuild/reconcile from current main, then run fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: reconcile the branch to one exact current-main-derived head, run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on that same SHA, and preserve schema20-specific Task15 fixtures.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
