# PVNaive — Canonical Project Status

Last updated: 2026-09-12 15:39 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `3456f28c2f5c6680accbca694038e0b159b4c4c0`.
- This cycle re-verified repository metadata, open PRs, CI visibility, and persistent project reports. Documentation reconciliation is the only change made in this cycle; no runtime, schema, credential, Caddy, Production, backup, rollback, merge, or deployment state was changed.
- Current-tip combined CI lookup returned no published statuses for `3456f28c2f5c6680accbca694038e0b159b4c4c0`; current-tip CI is not claimed green.
- The prior validated runtime integration remains Task16/schema21, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT at exact head `3fc14825e1b164bad558decaef47f56b792e81af`, based on stale `0b921abe9b2bd1d827023f494fda11a407fe34d3`; focused tests are supplemental and the fresh real HTTP/1.1 + HTTP/2 rehearsal plus exact accounting proof remain mandatory.
- PR #4 / Karing remains OPEN/DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100.
- Historical, stale, dirty, mixed-head, unpushed, or absent-CI evidence remains uncredited.
- Persistent reports describe a constrained one-active-host environment. TrPaqet remains the identified isolated Task13 rehearsal lane; connected-but-inactive handles are not treated as executable without a current clean-worktree receipt.

## Production truth
- No connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged promotion, or postflight evidence was available in this cycle.
- External Production health was not overclaimed; Production was not touched.
- Promotion remains gated by read-only audit → fresh encrypted backup → exact SHA lock → staged promotion → health/postflight → rollback readiness.

## Next executable gates
1. Task13: rebase/republish from current `main`, rerun exact-head CI/accounting/forwardproxy gates, then execute the isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current `main` if needed, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Security/accounting review: inspect merged Task16/schema21 for RLS fail-closed behavior, privilege separation, retention/purge safety, trusted session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. CI: obtain published same-head green CI for the current `main` tip and any candidate integration head.
5. Production: only after runtime gates are complete, perform read-only audit, fresh encrypted backup, independent rollback snapshot, staged promotion, health/postflight, and retained rollback.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, or historical Production records.
