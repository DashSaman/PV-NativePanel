# PVNaive — Canonical Handoff

Last updated: 2026-09-07 10:44 Asia/Tehran

## Current truth
- GitHub verified `main` at `5bd2420ea2fccb4b6e02f7dafded0a4d3dbd5f2d` before this refresh; this run adds documentation only. The resulting docs-only tip is `c2d5906b682ffca8f7cd57c2e6c49c83b970ea15`; no post-merge CI is credited for the new tip.
- #64 Task13 OPEN/DRAFT/mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; current combined status empty; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory.
- #81 Task16 OPEN/DRAFT/mergeable=false, current head `b96c65903e5fc314284ea777ceea236913a03842`; older dedicated PostgreSQL18 success and fixture corrections are supplemental only; no current single-SHA four-gate proof. Preserve schema20-specific Task15 fixtures.
- #4 Karing OPEN/DRAFT/mergeable=true; historical CI is not real-client proof; import/parse/connect/cleanup smoke remains pending.
- Stale documentation PRs remain non-canonical and were not merged.

## Production
No connected Production command/deploy lane or fresh command-level audit was available. Persistent evidence is bounded and historical/read-only only. No fresh health, backup, rollback or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
Worker output is creditable only when tied to an exact GitHub head and fresh receipt. Dirty/stale worktrees are not completion evidence. Do not integrate worker-only output. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, a fresh encrypted backup, independent rollback state, provenance and postflight verification.

## Actions in this run
- Re-verified current GitHub `main`, PR #64/#81/#4, exact-head combined-status availability and persistent coordinator/worker reports.
- Confirmed no validated worker completion was available for integration.
- Corrected canonical `PROJECT_STATUS.md`, `CONTINUE_HERE.md` and this handoff to the actual verified GitHub state.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task16/schema lane: from latest exact `main`, narrow-fix generic schema21/latest-schema expectations only, keep Task15 schema20 fixtures unchanged, then run normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13/protocol lane: exact-head checkout; fresh HTTP/1.1 + HTTP/2 rehearsal outside Production for target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing lane: real client import/parse/connect/cleanup smoke using disposable credentials and redacted logs.
4. Independent review lane: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production lane: read-only health first; only then backup/rollback and staged promotion after every release gate is green.

Human blocker: executable worker capacity with compatible Go/PostgreSQL18 tooling and a connected Production audit/deploy lane are still missing. Until available, no Task13 live rehearsal, Task16 repair validation or Production promotion can be freshly completed.
