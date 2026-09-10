# PVNaive — Canonical Project Status

Last updated: 2026-09-10 23:40 Asia/Tehran

## Verified GitHub state
- `main` at inspection start and end: `f0a0157868dfd12ff45ede8be7c0f88911ac1c3c` before this docs-only update.
- Current `main` push CI run `34519124546` / run 1667: SUCCESS.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated checks are historically green, but repository-wide four-gate closure is not currently verified on one fresh published head.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; exact-head CI run `33209239812` / run 402 is SUCCESS across web, Go, DB, rehearsal and production bundle, but real Karing-client import/parse/connect/cleanup proof remains missing.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts and are not current truth without exact-base validation.

## CI truth
- Current `main` documentation-only push CI is green.
- Task16 remains blocked on fresh same-head repository-wide closure, with the known database-fixture/RLS expectation issue not yet independently cleared.
- No runtime merge gate is green for promotion in this cycle.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Fresh assignments were re-issued for Task16, Task13, Karing, independent review, and the Production audit lane; no completion receipt was available to reconcile.

## Production truth
- Public panel health could not be independently verified through the connected web path, and no fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through connected tools.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: correct only generic latest-schema/RLS fixture expectations, preserve Task15 schema20-specific fixtures, and rerun normal CI + Task16 Schema21 TDD + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
