# PVNaive — Canonical Project Status

Last updated: 2026-09-10 21:40 Asia/Tehran

## Verified GitHub state
- `main` at inspection start: `c588f41caec8a7579267e332da488d31bc0c5185`.
- This cycle performed verification and documentation reconciliation only; no runtime, schema, credential, Caddy, backup, rollback, or Production mutation was integrated.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused checks are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; repository-wide CI run `33678134360` remains FAILED in the database job; web and Go passed, rehearsal/bundle skipped. No four-gate same-head closure is credited.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; exact-head CI run `33209239812` / run number `402` is SUCCESS across web, Go, DB, S04R rehearsal and production bundle, but real Karing-client import/parse/connect/cleanup proof is still missing.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts and are not current truth without exact-base validation.

## CI truth
- Current `main` is documentation-only; no post-merge workflow run is exposed for this head.
- Task16 database failure remains the active merge blocker; dedicated Task16 checks are not enough without repository-wide green CI on the same published head.
- PR #4 repository CI is green on its exact head, but compatibility evidence is incomplete.
- No safe merge gate is green in this run.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable slot; other lanes are inactive or upgrade-required under the connected one-active-host constraint.
- Fresh dispatch comments from the prior cycle remain the latest verified assignments: Task16 `5622504366`, Task13 `5622505075`, Karing `5622505912`.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through the connected state.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: fix only generic latest-schema/RLS fixture expectations, preserve Task15 schema20-specific fixtures, and rerun normal CI + Task16 Schema21 TDD + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.