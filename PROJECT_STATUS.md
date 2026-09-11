# PVNaive — Canonical Project Status

Last updated: 2026-09-11 09:43 Asia/Tehran

## Verified GitHub state
- `main` at inspection: `c24a569284b1902d3a7d291752a4e00682acf701`.
- `main` push CI run `34564969365` (run #1698) completed SUCCESS on 2026-09-11.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; no fresh same-head repository-wide four-gate closure is verified in this run.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical exact-head CI is SUCCESS, but independent real-client import/parse/connect/cleanup proof remains missing.
- PR #95 and other documentation PRs remain stale-base reconciliation branches; keep them uncredited unless rebased and freshly validated.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Fresh dispatches were posted this cycle to PR #81, PR #64, and PR #4 with exact-head acceptance criteria.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through connected tools in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: inspect the failing database job from the latest exact-head CI evidence, correct only generic latest-schema/RLS fixture expectations, preserve Task15 schema20-specific fixtures, publish one exact head, and rerun normal CI + Task16 Schema21 TDD + Exact Accounting + Pinned Forwardproxy.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
