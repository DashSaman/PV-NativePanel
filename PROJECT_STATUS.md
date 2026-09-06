# PVNaive — Canonical Project Status

Last updated: 2026-09-07 01:39 Asia/Tehran

## Verified state
- `main` exact head: `fe8af216fff15aa36fab51c36e40eae87467bace` (`docs: refresh canonical status after current production probe`). `fetch_commit_workflow_runs` for this exact head returned no PR-triggered workflow runs; post-merge CI is therefore unproven for this docs-only head.
- Open PR inventory includes stale documentation reconciliation PRs (#85/#86/#87/#88/#89/#91/#92/#93/#94/#95); none was merged in this run because their bases/claims are stale relative to current `main`.
- PR #64 Task13: OPEN, DRAFT, `mergeable=false`, head `3fc14825e1b164bad558decaef47f56b792e81af`; historical exact-head focused/CI gates are supplemental, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN, DRAFT, `mergeable=false`, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; prior dedicated gates are historical and the PR body/base metadata are stale. No current exact-head all-four-green receipt exists.
- PR #4 Karing: OPEN, DRAFT, `mergeable=true`, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI `33209239812` is SUCCESS, but reproducible real Karing client smoke is missing.

## Production truth
Available persistent evidence contains bounded read-only observations (including prior healthy API readiness/liveness and SNI-correct HTTPS probes), but no fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were rechecked. No fresh completion receipt tied to the current PR heads was found. Historical/dirty/stale worker trees are not credited; worker-only output was not integrated. The one-active-host/limited development capacity constraint remains the dominant execution blocker.

## This run
- Re-verified repository metadata, exact `main`, open PRs, current PR heads/bases/bodies, available CI evidence, and persistent coordinator/worker reports.
- Rejected stale docs PRs and historical worker claims as promotion evidence.
- Refreshed this canonical status file; documentation-only change.

## Next gates
1. Task16: create one clean current-head branch; narrow-fix only generic schema21/latest-schema expectations; preserve schema20-specific Task15 fixtures; align PR metadata; run normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: obtain fresh real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: obtain reproducible real client import/parse/connect/cleanup evidence with disposable credentials.
4. Independent review: inspect RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Only after all required gates pass: fresh encrypted Production backup, independent rollback state, staged deploy, and postflight verification.

Never claim completion from stale reports, older heads, partial evidence, or dirty worktrees.