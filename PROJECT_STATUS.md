# PVNaive — Canonical Project Status

Last updated: 2026-09-09 04:42 Asia/Tehran

## Verified GitHub state
- Authoritative GitHub `main` currently returns `8ac4f3e8935c24cb95085a6b04fffb6ce1b2cea4` (docs-only). Combined status is empty; do not claim post-merge CI from status alone.
- Main docs push CI run `34284651112` completed `success` on 2026-09-08 22:16Z; this proves only the docs commit's CI, not live rehearsal or Production health.
- PR #64 Task13: OPEN / DRAFT, API head `3fc14825e1b164bad558decaef47f56b792e81af`; focused exact-head CI/Exact Accounting/Pinned Forwardproxy evidence is green, but fresh real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT, API head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current exact-head runs on that SHA: Task16 Schema21 TDD `33678134359` SUCCESS, WS1 Exact Accounting `33678134326` SUCCESS, WS1 Pinned Forwardproxy `33678134350` SUCCESS, normal CI `33678134360` FAILURE in database job `101509296474`.
- Current Task16 CI failure is in generic latest-schema coverage: after migration/health/backup/restore and other gates pass, `tests/db/periodic_usage_reset_executor_test.sh` still asserts `schema version=20` while the migrated database is schema21. This is not a Production failure and is not credited as a green exact-head gate.
- PR #4 Karing: OPEN / DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; PR text reports CI/rehearsal/bundle success, but reproducible real-client Karing import/parse/connect/cleanup smoke is still required.
- PR #95 and older docs-only PRs are non-runtime evidence.

## Worker / coordinator truth
- Persistent-report search did not produce a fresh completion receipt tied to the current PR heads. Worker-only, stale, dirty, mixed-head, or historical evidence remains uncredited.
- Historical notes identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required. This is not a fresh command-level Production audit.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup preflight, independent rollback snapshot, staged deploy, or postflight was available in this run.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production must not be used as a test lane. Promotion requires exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this run
- Re-verified authoritative `main`, open PRs, exact-head CI runs/jobs/logs, and persistent reports.
- Confirmed Task16 exact-head specialized gates are green, while normal CI is still blocked by the remaining generic schema20 assertion in `periodic_usage_reset_executor_test.sh`.
- Confirmed the CI database log also shows expected safety/negative-path checks passing (append-only DELETE rejection, invalid refresh rejection, privilege/RLS checks) before the final schema mismatch.
- Updated canonical status to record the new exact run IDs and the concrete blocker.
- No runtime/schema/Production change was integrated.

## Next executable gates
1. Task16: update only the generic latest-schema expectation in `tests/db/periodic_usage_reset_executor_test.sh` (preserve Task15 schema20-specific fixtures), then rerun normal CI, Task16 Schema21 TDD, Exact Accounting, and Pinned Forwardproxy on one exact SHA.
2. Task13: execute isolated HTTP/1.1 + HTTP/2 rehearsal on exact `3fc14825...` with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.