# PVNaive — Canonical Project Status

Last updated: 2026-09-10 12:41 Asia/Tehran

## Verified GitHub state
- `main` at inspection start: `1dbb26ca5deabac5a00f24975a26d47c7fd647d8`; this cycle adds documentation only.
- PR #64 Task13: OPEN / DRAFT, exact published head `3fc14825e1b164bad558decaef47f56b792e81af`; existing focused evidence is supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #81 Task16: OPEN / DRAFT, exact published head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current exact-head workflow evidence remains not green as a four-gate set because repository-wide CI failed on the database path; no merge credit.
- PR #4 Karing: OPEN / DRAFT, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; CI/body evidence reports success, but independent real-client import/parse/connect/cleanup smoke remains required.
- Documentation-only PRs #95/#94/#93/#92/#91/#89/#88/#87/#86/#85 remain stale-base or historical reconciliation attempts.

## CI truth
- Task16 exact head `b96c65903e5fc314284ea777ceea236913a03842`: Schema21 TDD `33626300588` SUCCESS, WS1 Exact Accounting `33626300594` SUCCESS, WS1 Pinned Forwardproxy `33626300589` SUCCESS, repository-wide CI `33626300697` FAILURE.
- No current post-merge CI result is claimed for docs-only `main` `1dbb26ca...`.
- Historical green checks are not transferable across different PR heads.

## Worker / coordinator truth
- Persistent-report search found no fresh exact-head completion receipt for Task13, Task16, or Karing.
- New dispatch comments posted this cycle: Task16 `5616113876`; Task13 `5616116250`; Karing `5616117008`.
- Historical worker-only, stale, dirty, mixed-head, and unpushed output remains uncredited.

## Production truth
- No fresh command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight was available.
- No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation, or rollback mutation occurred.
- Production is not a test lane. Promotion requires green exact-head gates, fresh encrypted backup, independent rollback state, staged deploy, and postflight verification.

## Next executable gates
1. Task16: repair remaining generic latest-schema fixture expectations only; preserve Task15 schema20-specific fixtures; rerun all four gates on one exact SHA.
2. Task13: isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
3. Karing: real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
5. Production promotion only after all exact-head gates pass and fresh backup/rollback evidence exists.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
