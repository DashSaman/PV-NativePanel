# PVNaive — Canonical Project Status

Last updated: 2026-09-10 08:39 Asia/Tehran

## Verified GitHub state
- `main` at inspection start: `0b9cd38c482a8ffe4ce9cafda8a26be199ea9bcf`; this cycle did not modify runtime code.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; existing CI, Exact Accounting, and Pinned Forwardproxy evidence remain SUCCESS; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT; GitHub PR/branch ref currently resolves to published head `3c4310335ab4907d28bac995bba1be3545e14f6e`. The PR body references later worker commits including `b96c659...`, but that SHA is not the current published branch ref and cannot be credited as PR-head truth. Exact workflow query for `b96c659...` shows Task16 TDD, WS1 Exact Accounting, and WS1 Pinned Forwardproxy SUCCESS, while CI run `33626300697` is FAILURE; failed jobs were rerun this cycle. Do not merge until the published PR head itself has all four required gates green on one exact SHA.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; PR body reports CI #402 success including full S04R rehearsal and production bundle; independent real-client smoke remains required.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt tied to the currently published Task13, Task16, or Karing heads. Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Reports continue to identify TrPaqet as the active executable development slot; other workers are inactive or upgrade-required.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available in this cycle.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Actions in this cycle
- Re-verified repository metadata, current `main`, open PRs, exact published PR heads, exact-head workflow evidence, and persistent reports.
- Reran failed jobs for Task16 CI run `33626300697`.
- Posted fresh exact-head execution assignments to PR #81 Task16, PR #64 Task13, and PR #4 Karing.
- No runtime work was integrated because no validated exact-head completion receipt was available.

## Next executable gates
1. Task16: observe the rerun on the currently published head; if red, repair only generic latest-schema/RLS fixture expectations; preserve Task15 schema20 fixtures; rerun all four gates on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal covering target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
