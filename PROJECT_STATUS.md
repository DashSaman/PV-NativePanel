# PVNaive — Canonical Project Status

Last updated: 2026-09-08 14:41 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` currently returns `3da457660b282826bb3b3720f546003ccbf44b45`.
- This head is documentation-only; combined status is empty and no green post-merge CI is claimed.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; CI, Exact Accounting, and Pinned Forwardproxy are SUCCESS on that exact head, but the required fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN / DRAFT / mergeable=true, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c65903e5fc314284ea777ceea236913a03842` has Task16 Schema21 TDD, Exact Accounting, and Pinned Forwardproxy SUCCESS, but repository-wide CI is FAILURE on that candidate. Exact-head identity remains unresolved; preserve Task15 schema20-specific fixtures.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke is still required.
- PR #95 and older docs-only PRs remain non-runtime evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to the current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required while the slot is held elsewhere. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main`, open PRs, exact-head workflow runs, and persistent reports.
- Confirmed Task13 exact-head CI/Accounting/Forwardproxy SUCCESS on `3fc14825...`.
- Confirmed Task16 split state: API head `3c431033...`; green specialized evidence on candidate `b96c659...`; repository-wide CI FAILURE on that candidate.
- Added a reconciliation/assignment comment to PR #64.
- Reconciled this status file to the actual current `main` SHA.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task13: reconstruct/rehearse on exact `3fc14825...` with HTTP/1.1 + HTTP/2 target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: reconcile the branch to one exact current-main-derived head, fix repository-wide CI, then rerun normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on that same SHA while preserving schema20-specific Task15 fixtures.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
