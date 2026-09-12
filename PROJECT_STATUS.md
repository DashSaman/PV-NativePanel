# PVNaive — Canonical Project Status

Last updated: 2026-09-12 05:41 Asia/Tehran

## Verified GitHub state
- Current authoritative `main`: `f763fe4a6072ab33d1571292b32640349023221e`.
- Current-main combined status and workflow lookup show no published CI result for this exact docs checkpoint. No fresh green CI claim is made.
- Task16 / schema21 remains the last validated runtime integration, merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8` after recorded exact-head gates.
- PR #64 / Task13 remains OPEN / DRAFT at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; stale relative to current `main`; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- PR #4 / Karing remains OPEN / DRAFT at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, based on `s04-auth`; independent real-client smoke remains mandatory.
- Documentation PRs #85–#95 remain stale-base/reconciliation attempts and are not current runtime truth.

## Worker / coordinator truth
- No fresh exact-head completion receipt was found for Task13 or Karing in persistent reports or current PR discussions.
- Historical, stale, dirty, mixed-head, and unpushed output remains uncredited.
- Persistent capacity notes identify TrPaqet as the active executable development slot; other lanes are inactive or upgrade-required under the one-active-host constraint.
- Independent review lane remains open in issue #99; Production read-only lane remains open in issue #100.

## Production truth
- Available persistent material contains historical/read-only health and backup evidence, but no fresh connected command-level Production audit for this cycle.
- No current deployed SHA/schema verification, fresh encrypted backup, independent rollback snapshot, staged deploy, or postflight evidence was available through connected tools.
- Production was not touched.
- Promotion remains gated by read-only audit, fresh encrypted backup, exact SHA lock, staged promotion, health/postflight, and rollback readiness.

## Actions this cycle
- Re-verified repository metadata, current `main`, open PRs, exact PR heads/bases, CI visibility, persistent reports, and available Production evidence.
- Reviewed current PR #64 and PR #4 state; no validated exact-head completion evidence was available to integrate.
- Reaffirmed worker assignments on PR #64, PR #4, issue #99, and issue #100.
- Reconciled canonical documentation only; no runtime merge or Production deployment was justified.

## Next executable gates
1. Task13: rebase/republish from current `main`, then run isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload, and exactly-once accounting.
2. Karing: rebase/republish from current `main`, then run real import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version, and redacted logs.
3. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage, and secret redaction after Task16 merge.
4. CI: obtain a published workflow result for exact `main` `f763fe4a6072ab33d1571292b32640349023221e`.
5. Production: read-only audit first; only after runtime evidence is complete create encrypted backup + independent rollback snapshot, then staged promotion and postflight.

Never claim completion from stale reports, older heads, partial evidence, mixed-head proofs, dirty worktrees, absent CI, or historical Production records.
