# PVNaive — Canonical Handoff

Last updated: 2026-09-07 17:45 Asia/Tehran

## Current truth
- GitHub `main` was verified at `8d70a6ec92dac5b4117cb23701be1ac24d7c67ae`; this run adds documentation-only reconciliation commits. Exact-head combined status for the verified `main` was empty, so no post-merge CI green is claimed.
- #64 Task13 OPEN/DRAFT/mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; focused tests are supplemental; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; latest dedicated evidence is insufficient for promotion because repository-wide CI remains blocked by the generic schema-version expectation `schema version=21, want=20`.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; import/parse/connect/cleanup smoke remains pending.

## Production
Historical read-only health evidence exists for `pv-primary`, but no fresh command-level audit was available in this run. No fresh backup, rollback, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
No fresh completion receipt tied to current PR heads was found. Historical worker-only/stale/dirty work remains uncredited. Use disposable credentials and isolated canaries. Promotion requires every exact-head gate green, a fresh encrypted backup, independent rollback state, provenance and postflight verification.

## Actions in this run
- Re-verified current `main`, open PRs #64/#81/#4, exact-head status availability, canonical docs, and persistent-report search results.
- Confirmed no validated worker completion was available for integration.
- Updated `PROJECT_STATUS.md`, `CONTINUE_HERE.md`, and this handoff with the actual current main and evidence boundaries.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16/schema lane: correct only the remaining generic latest-schema expectation(s) that still require 20; preserve Task15 schema20 fixtures; rerun normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA and publish logs/run IDs.
2. Task13/protocol lane: exact-head isolated HTTP/1.1 + HTTP/2 rehearsal with target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting; do not bypass tooling blockers.
3. Karing lane: real client import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, platform/client version and redacted logs.
4. Independent review lane: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production lane: read-only health first; only after all release gates are green, create fresh encrypted backup and independent rollback state, then staged promotion and postflight.

Human blocker: compatible executable worker capacity and a connected Production audit/deploy lane are still missing. Until available, no Task13 live rehearsal or Production promotion can be freshly completed, and Task16 remains blocked by the generic schema21 fixture mismatch.