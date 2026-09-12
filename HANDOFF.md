# PVNaive — Canonical Handoff

Last updated: 2026-09-12 06:41 Asia/Tehran

## Current truth
- Current authoritative `main`: `2239ca96c76fba55786b4e3f55711e9e91f86b4f` before this documentation refresh; the refresh itself is documentation-only.
- Current-main combined status is empty and the workflow lookup for the exact current SHA returned no runs; do not claim green.
- Task16 / schema21 is validated and merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` after recorded exact-head gates on `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- Task13 / PR #64 remains OPEN/DRAFT at `3fc14825e1b164bad558decaef47f56b792e81af`; stale versus current main; live HTTP/1.1 + HTTP/2 rehearsal is still required.
- Karing / PR #4 remains OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; base `s04-auth`; independent real-client smoke is still required.

## Worker and Production
- No fresh exact-head completion receipt was found for Task13 or Karing.
- TrPaqet remains the active executable slot; other worker lanes are inactive or upgrade-required under the one-active-host constraint.
- Issue #99 is the independent security/accounting review lane. Issue #100 is the Production read-only audit/rollback-readiness lane.
- No fresh connected command-level Production audit, deployed SHA/schema verification, encrypted backup, independent rollback snapshot, staged deployment, or postflight evidence was available. Production was not mutated.

## Actions this cycle
- Re-verified GitHub repository, exact main ref, open PRs, PR heads/bases, CI visibility, persistent reports, and available Production evidence.
- Reviewed PR #64 and PR #4 without integrating unvalidated work.
- Reconciled canonical documentation directly on `main` only.
- Reaffirmed task assignments to PR #64, PR #4, issue #99, and issue #100.

## Next assignments
1. Task13: rebase/republish from current main, then run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current main, then run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
3. Security/accounting lane: complete issue #99 on clean current-main worktree.
4. Production lane: complete issue #100 read-only audit and rollback-readiness inventory only; no mutation.
5. CI: obtain and verify a published workflow result for exact current main.

Human/infrastructure blockers: connected worker execution for the live rehearsal, a real Karing client, current-main CI publication, and connected Production audit/deploy access.
