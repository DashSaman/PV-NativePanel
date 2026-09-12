# PVNaive — Canonical Project Status

Last updated: 2026-09-12 10:38 Asia/Tehran

## Verified GitHub state
- Current authoritative `main` at verification start: `a3787fb42654b64cc2cec9d7248decaeb1289b77`.
- This cycle adds documentation-only reconciliation; no runtime, schema, credential, Caddy, Production, backup, rollback, merge, or deployment mutation.
- No published combined status entries or workflow runs were available for exact `a3787fb42654b64cc2cec9d7248decaeb1289b77`; no current-main green CI claim is made.
- Task16 / schema21 remains the last validated runtime integration, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8` after recorded exact-head gates.
- PR #64 / Task13 remains OPEN / DRAFT / non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; its base is stale and fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #4 / Karing remains OPEN / DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13, Karing, issue #99, or issue #100 in the persistent reports or current PR discussions.
- Historical, stale, dirty, mixed-head, and unpushed output remains uncredited.
- TrPaqet remains the only identified executable development slot; `pv-primary` and `pv-worker-main` are connected but inactive under the one-active-host constraint.
- Fresh dispatches are being reissued for all independent lanes; no completion evidence was received in this cycle.

## Production truth
- No fresh connected command-level Production audit, deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged deploy, or postflight evidence was available in this cycle.
- Production was not touched.
- Promotion remains gated by read-only audit, fresh backup, exact SHA lock, staged promotion, health/postflight, and rollback readiness.

## Actions this cycle
- Re-verified repository metadata, current `main`, open PRs, exact PR heads/bases, current-main CI visibility, PR #64/PR #4 evidence, and persistent report availability.
- Reconciled worker evidence conservatively; no stale or incomplete report was credited.
- Posted fresh exact-head dispatches to Task13, Karing, security/accounting review, and Production read-only audit lanes.
- Updated canonical documentation only; no runtime, schema, credential, Caddy, backup, rollback, merge, or Production mutation.

## Next executable gates
1. Task13: rebase/republish from current main, rerun exact-head CI/accounting/forwardproxy gates, then run isolated real HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current main if needed, then run real import/parse/connect/cleanup smoke with exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS fail-closed behavior, privilege separation, retention/purge safety, accounting/session lineage, commit-before-HTTP-success semantics, and secret redaction.
4. CI: obtain and verify a published workflow result for exact current main.
5. Production: read-only audit first; only after runtime evidence is complete create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, absent CI, or historical Production records.
