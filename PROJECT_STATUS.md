# PVNaive — Canonical Project Status

Last updated: 2026-09-12 12:42 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `4a29e823877bd1cdd67f9d4092a426fbedb1a06b`.
- This cycle changed no runtime, schema, credential, Caddy, Production, backup, rollback, merge, or deployment state.
- Exact current `main` has no published combined-status entries and no workflow runs in the connected GitHub view; current-main CI is therefore **not claimed green**.
- Task16/schema21 PR #81 is merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`. Treat the runtime change as merged, but do not transfer historical gate receipts to another head.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal and exact accounting proof are still mandatory.
- PR #4 / Karing remains OPEN/DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; independent real-client smoke remains mandatory.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100.
- Historical, stale, dirty, mixed-head, unpushed, or absent-CI evidence remains uncredited.
- Persistent capacity notes describe a constrained one-active-host environment. Do not assume a worker is executable unless a current clean-worktree receipt proves it.
- Next independent assignments are recorded on PRs/issues; worker identity is not invented where no live agent handle is available.

## Production truth
- No connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged promotion, or postflight evidence was available in this cycle.
- Production was not touched.
- Promotion remains gated by read-only audit → fresh encrypted backup → exact SHA lock → staged promotion → health/postflight → rollback readiness.

## Next executable gates
1. Task13: rebase/republish from current `main`, rerun exact-head CI/accounting/forwardproxy gates, then execute the isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current `main` if needed, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Security/accounting review: inspect merged Task16/schema21 for RLS fail-closed behavior, privilege separation, retention/purge safety, trusted session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. CI: obtain a published workflow result for the exact current `main` and for any candidate integration head.
5. Production: only after runtime gates are complete, perform read-only audit, fresh encrypted backup, independent rollback snapshot, staged promotion, health/postflight, and retained rollback.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, absent CI, or historical Production records.
