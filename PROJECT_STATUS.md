# PVNaive — Canonical Project Status

Last updated: 2026-09-10 16:40 Asia/Tehran

## Verified GitHub state
- `main` at inspection start: `8e4823543b5d43f6cc91bda5e0a19232fe8c01d3`; this cycle performs documentation reconciliation only.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused CI evidence is supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; PR body records prior dedicated green checks but repository-wide CI is not credited as a four-gate exact-head set because the database path previously failed. The connected status view currently exposes no status entries for this published head.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI/body evidence is not a substitute for independent real-client import/parse/connect/cleanup proof.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts and are not current truth without exact-base validation.

## CI truth
- No fresh post-merge CI result is claimed for docs-only `main` `8e4823543b5d43f6cc91bda5e0a19232fe8c01d3`.
- Historical Task16 dedicated runs and the repository-wide failure remain non-transferable across heads.
- No safe merge gate is green in this run.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- The latest persistent state identifies TrPaqet as the active executable slot; other lanes are inactive or upgrade-required.
- Fresh exact-head dispatches were posted this cycle: Task16 `PVNaive-orchestrator-2026-09-10-task16`; Task13 `PVNaive-orchestrator-2026-09-10-task13`; Karing `PVNaive-orchestrator-2026-09-10-karing`; independent review `PVNaive-orchestrator-2026-09-10-review`; read-only Production audit `PVNaive-orchestrator-2026-09-10-prod-audit`.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through the connected state.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: repair remaining generic latest-schema fixture expectations only; preserve Task15 schema20-specific fixtures; rerun all four gates on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
