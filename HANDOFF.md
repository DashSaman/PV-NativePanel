# PVNaive — Canonical Handoff

Last updated: 2026-09-07 06:39 Asia/Tehran

## Current truth
- GitHub recent history verifies `main` at `49fe611b848bcda7015c1193fe86b3360c0dd4fe` before this documentation refresh. This run is documentation-only; combined status for that exact head was empty and no post-merge CI is credited.
- #64 Task13 OPEN/DRAFT, head `3fc14825e1b164bad558decaef47f56b792e81af`; current combined status empty; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current combined status empty; older dedicated PostgreSQL18 success and stale-fixture corrections are supplemental only; no current single-SHA four-gate proof. Preserve schema20-specific Task15 fixtures.
- #4 Karing OPEN/DRAFT, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI success is not real-client proof; import/parse/connect/cleanup smoke remains pending.
- Stale docs PRs #85/#86/#87/#88/#89/#91/#92/#93/#94/#95 were not merged.

## Production
No connected Production command/deploy lane or fresh command-level audit was available. Persistent evidence is bounded and historical/read-only only. No fresh health, backup, rollback or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Worker output is creditable only when tied to an exact GitHub head and fresh receipt. Dirty/stale worktrees are not completion evidence. Do not integrate worker-only output. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, a fresh encrypted backup, independent rollback state, provenance and postflight verification.

## Actions in this run
- Re-verified latest GitHub `main`, PR heads/bodies, exact-head combined status availability and persistent coordinator/worker reports.
- Added exact next-step comments to PRs #81, #64 and #4.
- Updated canonical status documentation only; no runtime/schema change was integrated.

## Next assignments
1. Task16/schema lane: from latest exact `main`, narrow-fix generic schema21/latest-schema expectations only, keep Task15 schema20 fixtures unchanged, then run normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13/protocol lane: exact-head checkout; fresh HTTP/1.1 + HTTP/2 rehearsal outside Production for target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing lane: real client import/parse/connect/cleanup smoke using disposable credentials and redacted logs.
4. Independent review lane: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production lane: read-only health first; only then backup/rollback and staged promotion after every release gate is green.

Human blocker: an executable development worker slot with compatible Go/PostgreSQL18 tooling (or multi-host capacity) and a connected Production audit/deploy lane are still missing. Until available, no Task13 live rehearsal, Task16 repair validation or Production promotion can be freshly completed.
