# PVNaive — Canonical Handoff

Last updated: 2026-09-07 15:46 Asia/Tehran

## Current truth
- GitHub verified `main` at `f23d9bf00c4b94f397a87111d9d56ca0708b41cf`; exact-head combined status empty; no post-merge CI green is claimed for the docs-only lineage.
- #64 Task13 OPEN/DRAFT/mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused tests are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory. Tooling/worker-capacity blocker remains unresolved.
- #81 Task16 OPEN/DRAFT/mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated gates were green on prior exact heads, but fresh single-SHA four-gate proof is not verified. Preserve schema20-specific Task15 fixtures; stale generic schema21/RLS failures and dirty worker candidates remain uncredited.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; import/parse/connect/cleanup smoke remains pending.

## Production
Historical read-only health evidence exists for `pv-primary`, but no fresh command-level audit was executed in this run. No fresh backup, rollback, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
No fresh completion receipt tied to current PR heads was found. Historical `pv-worker-main` schema/auth-refresh work was stale/dirty and could not run required DB validation because `psql` was unavailable; it was not credited or integrated. Use disposable credentials and isolated canaries. Promotion requires every exact-head gate green, a fresh encrypted backup, independent rollback state, provenance and postflight verification.

## Actions in this run
- Re-verified current GitHub `main`, open PRs #64/#81/#4, exact-head status availability, and persistent coordinator/worker reports.
- Added fresh dispatch comments to all three active lanes.
- Updated `PROJECT_STATUS.md`, `CONTINUE_HERE.md`, and this handoff to the verified current state.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16/schema lane: start from current main; patch only generic latest-schema/schema21 expectations; preserve Task15 schema20 fixtures; rerun normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA and publish logs/run IDs.
2. Task13/protocol lane: exact-head isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting; do not bypass tooling blockers.
3. Karing lane: real client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, platform/client version and redacted logs.
4. Independent review lane: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production lane: read-only health first; only after all release gates are green, create fresh encrypted backup and independent rollback state, then staged promotion and postflight.

Human blocker: compatible executable worker capacity and a connected Production audit/deploy lane are still missing. Until available, no Task13 live rehearsal, Task16 fresh validation or Production promotion can be freshly completed.
