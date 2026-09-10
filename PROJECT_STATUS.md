# PVNaive — Canonical Project Status

Last updated: 2026-09-11 01:38 Asia/Tehran

## Verified GitHub state
- `main` at inspection: `49be527e0c63ed0295414a63213dd9071fbde87b`.
- Current main push workflow visibility for this SHA: no PR-triggered workflow runs returned by the connected GitHub view; no fresh post-update CI result is claimed.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; historical dedicated checks are not sufficient; fresh same-head repository-wide four-gate closure remains pending. Latest persistent evidence still identifies generic schema21 fixture mismatch as the database blocker.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; exact-head workflow `33209239812` / run 402 is SUCCESS, but independent real-client import/parse/connect/cleanup proof remains missing.
- Documentation-only PRs remain stale-base or historical reconciliation attempts and are not current truth without exact-base validation.

## CI truth
- No fresh green runtime merge gate was observed in this cycle.
- Do not reuse older green evidence after a head changes.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Fresh exact-head dispatches remain required for Task13, Task16, and Karing; no completion receipt was available to reconcile.

## Production truth
- Public panel health could not be independently verified through the connected web path.
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through connected tools.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: correct only generic latest-schema/RLS fixture expectations, preserve Task15 schema20-specific fixtures, and rerun normal CI + Task16 Schema21 TDD + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
