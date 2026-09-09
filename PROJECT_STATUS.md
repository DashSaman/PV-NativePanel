# PVNaive — Canonical Project Status

Last updated: 2026-09-09 06:38 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` latest verified commit at inspection: `91a9c63d1cc02a4f172f5e241084f447b87920a1` (`docs: refresh continue-here with current exact-head blockers`). This is documentation-only; `fetch_commit_workflow_runs` returned no PR-triggered workflow runs for this exact head, so no post-merge CI is claimed.
- PR #64 Task13: OPEN / DRAFT, exact head `3fc14825e1b164bad558decaef47f56b792e81af`. Exact-head CI `33623363327`, WS1 Exact Accounting `33623363299`, and WS1 Pinned Forwardproxy `33623363389` are SUCCESS. Required fresh real HTTP/1.1 + HTTP/2 rehearsal is still missing.
- PR #81 Task16: OPEN / DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; candidate `b96c65903e5fc314284ea777ceea236913a03842`. Task16 Schema21 TDD `33626300588`, Exact Accounting `33626300594`, and Pinned Forwardproxy `33626300589` are SUCCESS; normal CI `33626300697` is FAILURE. Exact-head identity remains unresolved.
- Task16 blocker remains the generic latest-schema assertion in `tests/db/periodic_usage_reset_executor_test.sh` (`schema version=21, want=20`) after migration/health/backup/restore and safety checks pass. Preserve Task15 schema20-specific fixtures; do not credit green until one exact published SHA has all four gates green.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; PR body reports CI success, but reproducible real-client import/parse/connect/cleanup smoke is still required and not independently verified.
- Task12 branch `lead/task12-session-management-2026-08-31` exists at `3cd98a1bc1358fc3b58dd8642646da122cac84c6` and is not an open PR/merge candidate; do not treat it as current-main-integrated work.

## Worker / coordinator truth
- Persistent-report search yielded no fresh completion receipt tied to current Task13/Task16/Karing heads. Worker-only, stale, dirty, mixed-head, or historical evidence remains uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires green exact-head gates, a fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this cycle
- Re-verified repository metadata, current `main`, open PRs, exact-head workflow runs, Task12 branch state, and persistent coordinator/worker reports.
- Confirmed Task13 focused gates are green but live protocol rehearsal is still missing.
- Confirmed Task16 specialized gates are green while normal CI remains blocked by the generic schema20 assertion and API-head/candidate drift.
- Confirmed Karing PR remains unverified without a real-client smoke receipt.
- Updated canonical status with the current main and exact blockers.
- Posted fresh lane dispatches on PRs #64, #81, and #4; no code/runtime/Production mutation was performed.

## Next executable gates
1. Task16: repair only the generic latest-schema expectation on a clean branch from current `main`, preserve Task15 schema20-specific fixtures, reconcile exact head, then rerun normal CI, Task16 Schema21 TDD, Exact Accounting, and Pinned Forwardproxy on one SHA.
2. Task13: run isolated HTTP/1.1 + HTTP/2 rehearsal on exact `3fc14825...` proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
