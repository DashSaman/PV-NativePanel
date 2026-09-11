# PVNaive — Canonical Project Status

Last updated: 2026-09-11 13:38 Asia/Tehran

## Verified GitHub state
- `main` at inspection: `848154d01130fcef3f310ac17c0106f71a2a5f96`.
- Combined status for this docs-only checkpoint is empty; no fresh post-update CI is claimed.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head moved after a minimal fixture repair to `904e17c4a013e3adb5fb349c70f254ab59c925f8`; prior exact-head gates were 3 green + repository CI failure. The failure was confirmed in `tests/db/customer_lifecycle_migration_test.sh` asserting schema 20 after schema21 migration. The repair changes only the generic latest-schema expectation to 21 and adds rollback 21→20 coverage; Task15 schema20-specific fixtures remain pinned to 20. Fresh CI rerun is pending.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical exact-head CI is SUCCESS, but independent real-client import/parse/connect/cleanup proof remains missing.
- PR #95 and other documentation-only PRs remain stale-base reconciliation attempts; keep them uncredited unless rebased and freshly validated.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13 or Karing.
- Task16 has a new branch commit `904e17c4a013e3adb5fb349c70f254ab59c925f8` for the minimal generic fixture repair; it is not credited as complete until fresh same-head CI and dedicated gates succeed.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Fresh exact-head dispatch comments were posted this cycle to PR #81 (`5632895387`), PR #64 (`5632896115`), and PR #4 (`5632896690`).

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through connected tools in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: wait for fresh CI on `904e17c4...`; if green, verify all four gates on the same head and only then consider merge.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
