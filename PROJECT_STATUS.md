# PVNaive — Canonical Project Status

Last updated: 2026-09-08 22:39 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` currently returns `478638bdd491173fb8bfd33f075eb157bc49ac3b`.
- This head is documentation-only; combined status is empty and the PR-triggered workflow-runs endpoint returned no runs. No green post-merge CI is claimed for this head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI `33623363327`, WS1 Exact Accounting `33623363299`, and WS1 Pinned Forwardproxy `33623363389` are green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains missing.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; documented candidate `b96c65903e5fc314284ea777ceea236913a03842` has specialized gates green, but repository-wide CI previously failed database coverage with `ERROR: RLS coverage check failed: 43/42`. Exact-head promotion remains blocked until one clean head fixes the mismatch and all four gates pass on that same SHA; preserve Task15 schema20-specific fixtures.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; reproducible real-client smoke is still required.
- PR #95 and older docs-only PRs remain non-runtime evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main`, open PRs, current combined status, and persistent reports.
- Confirmed `main=478638bdd491173fb8bfd33f075eb157bc49ac3b` has no current status checks or PR-triggered workflow runs.
- Reconciled this file to the actual current `main` SHA and recorded current exact-head blockers.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task16: create one clean current-main-derived head, fix the repository-wide RLS coverage mismatch (`43/42`), preserve Task15 schema20-specific fixtures, then rerun normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on that same SHA.
2. Task13: reconstruct/rehearse on exact `3fc14825...` with HTTP/1.1 + HTTP/2 target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
