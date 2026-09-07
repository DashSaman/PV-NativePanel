# PVNaive — Canonical Handoff

Last updated: 2026-09-07 14:38 Asia/Tehran

## Current truth
- GitHub verified `main` before this docs refresh at `c75e98bb58586454810840666ce42307a9b7a173`; exact-head combined status empty; no post-merge CI green is claimed for the docs-only lineage.
- #64 Task13 OPEN/DRAFT/mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal remains mandatory. TrPaqet is the only executable rehearsal slot; host-local Go 1.18.1 and missing `jq` remain unresolved.
- #81 Task16 OPEN/DRAFT/mergeable=false, head `3c4310335ab4907d28bac995bba1be3545e14f6e`; current exact-head four-gate proof is not verified. Preserve schema20-specific Task15 fixtures.
- #4 Karing OPEN/DRAFT/mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; import/parse/connect/cleanup smoke remains pending.

## Production
No connected Production command/deploy lane or fresh command-level audit was available. No fresh health, backup, rollback or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation occurred.

## Worker/release rules
No fresh completion receipt tied to the current PR heads was found in persistent report search. Dirty/stale worker output is not completion evidence. Use disposable credentials and isolated canaries. Promotion requires all exact-head gates green, a fresh encrypted backup, independent rollback state, provenance and postflight verification.

## Actions in this run
- Re-verified current GitHub `main`, open PRs #64/#81/#4, exact-head status availability and persistent coordinator/worker reports.
- Added fresh dispatch comments to all three active lanes.
- Corrected canonical status, continue-here and handoff documents to the verified GitHub state.
- No runtime/schema/Production change was integrated.

## Next assignments
1. Task13/protocol lane: exact-head checkout; fresh isolated HTTP/1.1 + HTTP/2 rehearsal for target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting. Resolve tooling via an approved compatible worker/image; do not bypass proofs.
2. Task16/schema lane: rebuild/update from latest exact `main`, narrow-fix generic schema21/latest-schema expectations only, keep Task15 schema20 fixtures unchanged, then run normal CI + PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
3. Karing lane: real client import/parse/connect/cleanup smoke with disposable credentials and redacted logs.
4. Independent review lane: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production lane: read-only health first; only then backup/rollback and staged promotion after every release gate is green.

Human blocker: executable worker capacity with compatible tooling, plus a connected Production audit/deploy lane, are still missing. Until available, no Task13 live rehearsal, Task16 fresh validation or Production promotion can be freshly completed.
