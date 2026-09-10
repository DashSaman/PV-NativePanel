# PVNaive — Canonical Project Status

Last updated: 2026-09-10 18:43 Asia/Tehran

## Verified GitHub state
- `main` at inspection start: `b6fbbde898405bb339d75fdb19b75aa3e2a29945`; this cycle performs documentation reconciliation only.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; focused CI evidence is supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; Task16 TDD `33678134359`, Exact Accounting `33678134326`, and Pinned Forwardproxy `33678134350` are SUCCESS; repository-wide CI `33678134360` remains FAILURE in the database job only. Failed jobs were re-run this cycle; no green credit is claimed until a successful same-head result is observed.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI/body evidence is not a substitute for independent real-client import/parse/connect/cleanup proof.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts and are not current truth without exact-base validation.

## CI truth
- Current `main` is a documentation-only head; no fresh post-merge workflow runs are exposed for this head.
- Task16 dedicated exact-head gates are green, but the repository-wide CI run remains failed in the database path; the failed jobs were re-run from `33678134360` during this cycle.
- No safe merge gate is green in this run.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent material identifies TrPaqet as the active executable slot; other lanes are inactive or upgrade-required under the connected one-active-host constraint.
- Fresh dispatch comments posted this cycle: Task16 comment `5620960409`, Task13 comment `5620961328`, Karing comment `5620962152`.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available through the connected state.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: observe the re-run of the failed CI jobs; if database remains red, correct only generic latest-schema fixture expectations, preserve Task15 schema20-specific fixtures, and rerun all four gates on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
