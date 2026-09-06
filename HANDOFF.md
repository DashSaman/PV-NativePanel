# PVNaive — Canonical Handoff

Last updated: 2026-09-06 22:38 Asia/Tehran

## Current truth
- `main` exact head before this docs-only update: `70a42c7de2913face8a6d07e8c96a37dbc75e57b`; no workflow runs or status rows were visible for that head, so post-merge CI is not credited.
- #64 Task13 OPEN/DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; recorded focused gates are not a substitute for fresh real HTTP/1.1 + HTTP/2 rehearsal.
- #81 Task16 OPEN/DRAFT, reported mergeable but with stale base/body claims; head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated Task16/Accounting/Pinned evidence is historical and repository-wide CI remains blocked by generic schema21/latest-schema fixture/RLS expectation drift.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; historical CI green, reproducible real Karing smoke required.

## Production
Persistent reports provide bounded historical/read-only health observations, but no fresh command-level Production audit, encrypted backup preflight, independent rollback snapshot, deploy or postflight was executable in this run. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Worker reports remain historical unless matched to an exact GitHub head and fresh receipt. Dirty/stale worker trees are not completion evidence. Do not integrate worker-only output. Never print or copy secrets. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, fresh encrypted backup, independent rollback state, provenance and postflight verification.

## This run
- Re-verified `main`, open PRs #64/#81/#4, exact-head status/workflow presence, and persistent coordinator/worker reports.
- Confirmed no current exact-head all-four-green Task16 receipt, no fresh Task13 protocol rehearsal, and no real Karing smoke.
- Updated `PROJECT_STATUS.md`, `CONTINUE_HERE.md`, and this handoff on `main` as documentation-only commits.
- No runtime/schema work integrated and no Production mutation performed.

## Next assignments
1. Task16: clean branch from the latest exact `main`, narrow-fix only generic schema21/latest-schema expectations, preserve schema20-specific Task15 fixtures, align PR metadata, then run normal CI + Task16 PG18 + Exact Accounting + Pinned Forwardproxy on one exact SHA.
2. Task13: clean exact-head reconstruction and fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with exact accounting/session receipt.
3. Karing: reproducible real client import/parse/connect/cleanup smoke.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production-only lane: read-only health first; backup/rollback then staged promotion only after all release gates are green.
