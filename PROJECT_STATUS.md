# PVNaive — Canonical Project Status

Last updated: 2026-09-08 02:39 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` ref currently returns `31304a6cdf7f6b5772448386d36d1b201acb53b8`.
- Exact-head combined status for `main` is empty; no green post-merge CI is claimed for this docs-only head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, API head `3fc14825e1b164bad558decaef47f56b792e81af`, base SHA `0b921abe9b2bd1d827023f494fda11a407fe34d3` is stale versus current main. Historical exact-head gates/focused tests are supplemental; fresh current-main-derived real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; body cites older implementation heads including `b96c659...`; current exact-head four-gate proof is absent. Preserve Task15 schema20-specific fixtures and change only generic latest-schema expectations.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI/curl evidence is supplemental; reproducible real-client smoke remains missing.
- Open docs PRs #92–#95 are documentation-only historical refreshes with stale bases; none is treated as production evidence.

## Worker / coordinator truth
- Persistent-report search returned no fresh completion receipt tied to the current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence is uncredited.
- Historical notes identify TrPaqet as the active executable development slot and report `upgrade_required` on inactive development workers while the slot is held by Production Primary. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy, or postflight was available in this run. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main` ref, open PR metadata, exact-head status, and persistent worker/coordinator reports.
- Reconciled canonical status to the actual current `main` SHA.
- Posted fresh execution assignments to Task13, Task16, and Karing lanes.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task13: rebuild/reconcile from current main, then run fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Task16: create one clean current-main-derived head; update only generic latest-schema expectations; run normal CI + PostgreSQL18 + WS1 Exact Accounting + WS1 Pinned Forwardproxy on one exact SHA.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
