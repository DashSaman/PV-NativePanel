# PVNaive — Canonical Project Status

Last updated: 2026-09-12 06:41 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `08abdabb52d72708413d6291c415b184aff7b8eb` (documentation-only reconciliation from verified tree `2239ca96c76fba55786b4e3f55711e9e91f86b4f`).
- No published CI result is available for the exact current docs checkpoint; no fresh green CI claim is made.
- Task16 / schema21 remains the last validated runtime integration, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8` after recorded exact-head gates.
- PR #64 / Task13 remains OPEN / DRAFT at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #4 / Karing remains OPEN / DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13 or Karing.
- Historical, stale, dirty, mixed-head, and unpushed output remains uncredited.
- TrPaqet remains the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Issue #99 is the independent review lane; issue #100 is the Production read-only lane.

## Production truth
- No fresh connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged deploy, or postflight evidence was available.
- Production was not touched.
- Promotion remains gated by read-only audit, backup, exact SHA lock, staged promotion, health/postflight, and rollback readiness.

## Actions this cycle
- Re-verified GitHub, open PRs, exact heads/bases, CI visibility, persistent reports, and available Production evidence.
- Reviewed PR #64 and PR #4; no validated runtime work was available to integrate.
- Reconciled canonical documentation and reissued exact-head worker assignments.

## Next executable gates
1. Task13: rebase/republish from current main, then run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current main, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction.
4. CI: obtain a published workflow result for exact current main.
5. Production: read-only audit first; only after runtime evidence is complete create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, absent CI, or historical Production records.
