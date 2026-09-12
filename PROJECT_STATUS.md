# PVNaive — Canonical Project Status

Last updated: 2026-09-12 20:39 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `a1300961b1d364a812433af0f2cc25114053a66f`.
- Exact-main CI lookup for this SHA returned no published combined statuses or workflow runs; CI is therefore **not claimed green** for this checkpoint.
- The last validated runtime integration remains Task16/schema21, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`, based on stale `0b921abe9b2bd1d827023f494fda11a407fe34d3`; focused gates are supplemental and the fresh real HTTP/1.1 + HTTP/2 rehearsal plus exact accounting proof remain mandatory.
- PR #4 / Karing remains OPEN/DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.
- Documentation-only PRs #95 and earlier remain stale relative to current `main` and are not treated as validated runtime or Production changes.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100 in this inspection.
- Historical, stale, dirty, mixed-head, unpushed, or absent-CI evidence remains uncredited.
- Persistent reports describe a multi-server development pool and identify TrPaqet as the isolated Task13 rehearsal lane, but no current clean-worktree completion receipt was available.

## Production truth
- No connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged promotion, or postflight evidence was available in this cycle.
- External Production health was not overclaimed; Production was not touched.
- Promotion remains gated by read-only audit → fresh encrypted backup → exact SHA lock → staged promotion → health/postflight → rollback readiness.

## Next executable gates
1. Task13: rebase/republish from current `main`, rerun exact-head CI/accounting/forwardproxy gates, then execute the isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current `main` if needed, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Security/accounting review: inspect merged Task16/schema21 for RLS fail-closed behavior, privilege separation, retention/purge safety, trusted session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. Production: only after runtime gates are complete, perform read-only audit, fresh encrypted backup, independent rollback snapshot, staged promotion, health/postflight, and retained rollback.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, or historical Production records.
