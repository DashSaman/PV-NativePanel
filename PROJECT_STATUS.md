# PVNaive — Canonical Project Status

Last updated: 2026-09-09 05:39 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` latest verified commit: `99d1cb2baf0bff8426827df2bf06c0830b19ba17` (`docs: update continue-here with exact Task16 CI blocker`). This is documentation-only; no fresh post-merge CI is claimed for this exact head.
- PR #64 Task13: OPEN / DRAFT, API head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI `33623363327`, Exact Accounting `33623363299`, and Pinned Forwardproxy `33623363389` are SUCCESS. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory before merge.
- PR #81 Task16: OPEN / DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; candidate implementation head `b96c65903e5fc314284ea777ceea236913a03842` has Task16 Schema21 TDD `33626300588`, Exact Accounting `33626300594`, and Pinned Forwardproxy `33626300589` SUCCESS, but normal CI `33626300697` is FAILURE in the database job. Exact-head identity is still not reconciled because the PR API head and candidate SHA differ.
- Task16 database failure is the generic latest-schema expectation in `tests/db/periodic_usage_reset_executor_test.sh` (`schema version=21, want=20`) after migration/health/backup/restore and safety checks pass. Do not credit as green until the fix is on one exact published SHA and all four gates rerun.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI `33209239812` is SUCCESS, but reproducible real-client Karing import/parse/connect/cleanup smoke is still required.
- Older docs-only PRs are non-runtime evidence.

## Worker / coordinator truth
- Persistent-report search did not yield a fresh completion receipt tied to the current Task13/Task16/Karing heads. Worker-only, stale, dirty, mixed-head, or historical evidence remains uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires green exact-head gates, a fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this cycle
- Re-verified repository metadata, latest `main`, open PRs, exact-head workflow runs, and persistent coordinator/worker reports.
- Confirmed Task13 exact-head focused gates are green but real protocol rehearsal is still missing.
- Confirmed Task16 specialized gates are green while normal CI remains blocked by the generic schema20 assertion and exact-head drift.
- Confirmed Karing CI is green but real-client smoke is still absent.
- Updated canonical status to record the current main and exact blockers.
- Posted fresh lane dispatches on PRs #64, #81, and #4; no code/runtime/Production mutation was performed.

## Next executable gates
1. Task16: repair only the generic latest-schema expectation, preserve Task15 schema20-specific fixtures, reconcile the branch to current `main`, then rerun normal CI, Task16 Schema21 TDD, Exact Accounting, and Pinned Forwardproxy on one exact SHA.
2. Task13: run isolated HTTP/1.1 + HTTP/2 rehearsal on exact `3fc14825...` proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
