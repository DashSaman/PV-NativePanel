# PVNaive — Canonical Project Status

Last updated: 2026-09-08 00:41 Asia/Tehran

## Verified GitHub state
- GitHub default-branch ref endpoint currently returns `f1db7a894449a17598d7403d056d11815536bd78` for `main`.
- PR metadata is internally inconsistent with that ref: PR #64 and #81 bodies cite older `main` bases; the repository open-PR listing also includes docs PR #95 based on `a5d114c9...`. Treat this as a GitHub/API reconciliation blocker; do not infer a clean linear history from stale body text.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`; status endpoint reports `pending` with zero statuses. Historical focused/CI evidence is supplemental only; fresh current-main-derived real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; body references older implementation heads including `b96c659...`; current exact-head four-gate proof is absent. Preserve Task15 schema20-specific fixtures and change only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental; reproducible real-client smoke remains missing.
- PR #95 docs refresh: OPEN / DRAFT, head `f7f670dc48928527160b668e8a8f0f394f7fc2a5`, base recorded as `a5d114c9...`; documentation-only and not treated as production evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to the current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot and report `upgrade_required` on inactive development workers while the slot is held by Production Primary. This is not a fresh command-level Production audit.
- New dispatch/reconciliation comments were posted to PR #64, #81, and #4 in this run.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified repository metadata, `main` ref, open PRs, Task13 exact-head status, and persistent reports.
- Reconciled the GitHub/API base/head inconsistency as a blocker instead of crediting stale PR-body claims.
- Posted fresh autonomous assignment/reconciliation comments to PR #64, #81, and #4.
- Updated canonical project status, continue-here, and handoff files.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. GitHub reconciliation: confirm one authoritative current `main` SHA and update stale PR bodies/bases before any merge decision.
2. Task16: create one clean current-main-derived head; update only generic latest-schema expectations; run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on one exact SHA.
3. Task13: current-main-derived isolated protocol rehearsal for HTTP/1.1 + HTTP/2 proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
4. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
5. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
6. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
