# PVNaive — Canonical Project Status

Last updated: 2026-09-12 00:41 Asia/Tehran

## Verified GitHub state
- Authoritative `main` at inspection: `39914e4f63466dd2a04a75fcf1ac7f05ba255649`.
- `main` combined status query returned zero published status entries for this docs checkpoint; no fresh green CI claim is made.
- Task16 / schema21 remains the last validated runtime integration: merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8` after four required exact-head gates passed.
- PR #64 / Task13 remains OPEN / DRAFT at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; branch/base is stale relative to current `main`; fresh real HTTP/1.1 + HTTP/2 rehearsal is still mandatory.
- PR #4 / Karing remains OPEN / DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; base is `s04-auth`, not current `main`; independent real-client smoke is still mandatory.
- PRs #85–#95 are stale-base or documentation reconciliation attempts and are not current runtime truth.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13 or Karing in the persistent reports/comments inspected this cycle.
- Historical, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent capacity notes identify TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.

## Production truth
- No fresh connected command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deploy, or postflight evidence was available through connected tools in this cycle.
- No Production mutation occurred.
- Promotion remains gated by read-only audit, fresh encrypted backup, exact SHA lock, staged promotion, health/postflight, and rollback readiness.

## Next executable gates
1. Task13: rebase/republish from current `main`, then run isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current `main`, then run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction after Task16 merge.
4. Production: read-only audit, then backup/rollback/staged promotion only after runtime evidence is complete.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, or dirty worktrees.
