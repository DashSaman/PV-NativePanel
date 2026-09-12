# PVNaive — Canonical Handoff

Last updated: 2026-09-12 12:42 Asia/Tehran

## Current truth
- Current authoritative `main`: `4a29e823877bd1cdd67f9d4092a426fbedb1a06b` at inspection; this cycle adds documentation-only reconciliation.
- Exact current `main` has no published combined-status entries and no workflow runs in the connected GitHub view. Current-main CI is not claimed green.
- Task16/schema21 is merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`; historical receipts remain tied to their original heads.
- Task13 / PR #64 is OPEN/DRAFT/non-mergeable at `3fc14825e1b164bad558decaef47f56b792e81af`; fresh live HTTP/1.1 + HTTP/2 rehearsal and exact accounting proof are required.
- Karing / PR #4 is OPEN/DRAFT at `2501e39dc39e14063b6a501bc96b77bbfcae7384`; independent real-client smoke is required.

## Worker and Production
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100.
- Persistent reports describe a constrained worker environment; only current clean-worktree evidence is creditable.
- No connected Production command-level audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged deployment, or postflight evidence was available in this run.
- Production was not mutated.

## Next assignments
1. Task13: current-main-derived clean branch, exact-head CI, isolated HTTP/1.1 + HTTP/2 rehearsal, exact accounting receipt.
2. Karing: current-main-derived clean branch, real import/parse/connect/cleanup smoke, exact profile hash and redacted logs.
3. Security/accounting review (#99): merged Task16/schema21 RLS, privilege, retention/purge, lineage, commit-before-success, and redaction review.
4. Production lane (#100): read-only audit and rollback-readiness inventory only.
5. CI lane: publish and verify exact-current-main workflow results.

## Promotion rule
No merge or deployment from stale, mixed-head, dirty, partial, absent-CI, or historical evidence. If runtime gates become green, the sequence is read-only audit → fresh encrypted backup → exact SHA lock → staged promotion → health/postflight → rollback readiness.

## Human/infrastructure blockers
- Fresh Task13 live rehearsal environment and execution.
- Real Karing client for compatibility proof.
- Current-main CI publication for the latest docs checkpoint.
- Connected Production access for audit, backup, rollback, staged promotion, and postflight.
