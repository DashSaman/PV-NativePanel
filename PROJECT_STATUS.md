# PVNaive — Canonical Project Status

Last updated: 2026-09-12 07:40 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `2118ee70031c40122421090ce871bef2608f24aa` (latest observed GitHub branch head in this cycle).
- No published combined CI status entries were returned for this exact SHA; no current-main green CI claim is made.
- Task16 / schema21 remains the last validated runtime integration, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8` after recorded exact-head gates.
- PR #64 / Task13 remains OPEN / DRAFT / non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #4 / Karing remains OPEN / DRAFT / mergeable at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100.
- Historical, stale, dirty, mixed-head, and unpushed output remains uncredited.
- TrPaqet remains the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Issue #99 is the independent security/accounting review lane; issue #100 is the Production read-only/rollback-readiness lane.

## Production truth
- No fresh connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged deploy, or postflight evidence was available.
- Production was not touched.
- Promotion remains gated by read-only audit, fresh backup, exact SHA lock, staged promotion, health/postflight, and rollback readiness.

## Actions this cycle
- Re-verified repository metadata, current `main`, open PRs, PR heads/bases, current-main CI visibility, PR4/PR64 status, and persistent report availability.
- Reviewed current PR evidence and confirmed it is not sufficient for merge: PR #4 still lacks independent Karing client evidence; PR #64 still lacks the required real protocol rehearsal.
- Reissued execution assignments for Task13, Karing, security/accounting review, and Production audit/rollback readiness.
- Updated canonical documentation only; no runtime or Production mutation.

## Next executable gates
1. Task13: rebase/republish from current main, then run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current main, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS fail-closed behavior, privilege separation, retention/purge safety, accounting/session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. CI: obtain a published workflow result for exact current main.
5. Production: read-only audit first; only after runtime evidence is complete create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, absent CI, or historical Production records.
