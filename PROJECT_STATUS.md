# PVNaive — Canonical Project Status

Last updated: 2026-09-07 12:40 Asia/Tehran

## Verified state
- Current `main` tip from GitHub: `3a97c9a79aecd5998388cb23c047550eb181537d` (`docs: update canonical handoff with verified GitHub state`). Exact-head combined status is empty; no post-merge CI green is claimed for this docs-only head.
- PR #64 Task13: OPEN / DRAFT / mergeable=false, head `3fc14825e1b164bad558decaef47f56b792e81af`; current combined status is empty. Historical exact-head gates/focused tests are supplemental only; fresh real HTTP/1.1 + HTTP/2 rehearsal remains required.
- PR #81 Task16: OPEN / DRAFT / mergeable=false, current head `b96c65903e5fc314284ea777ceea236913a03842`; historical PG18 and fixture-repair evidence is mixed-head/supplemental, not current exact-head proof. Preserve schema20-specific Task15 fixtures.
- PR #4 Karing: OPEN / DRAFT / mergeable=true, head `2501e39dc39e14063b6a501bc96b77bbfcae7384`; historical CI is not real-client proof; reproducible real-client import/parse/connect/cleanup smoke remains missing.
- Stale documentation PRs remain non-canonical and were not merged.

## Production truth
No connected Production command/deployment lane or fresh command-level audit was available to this run. No fresh health pass, encrypted backup preflight, independent rollback snapshot, deploy or postflight is claimed. No merge, deploy, migration, restart/reload, DB write, credential mutation, backup mutation or rollback mutation occurred.

## Worker truth
Persistent coordinator/worker reports were searched again. No fresh completion receipt tied to the current PR heads was found. Historical, dirty or stale worker output is not credited and was not integrated. Persistent reports describe limited/one-active-host development capacity and prohibit using Production services as a test lane.

## Actions in this run
- Re-verified repository metadata, current `main`, open PR #64/#81/#4, exact-head combined-status availability and persistent coordinator/worker reports.
- Confirmed no validated worker completion was available for integration.
- Corrected canonical status to the actual GitHub `main` SHA `3a97c9a79aecd5998388cb23c047550eb181537d`.
- Dispatched fresh execution instructions to Task13, Task16 and Karing lanes.
- No runtime/schema/Production change was integrated.

## Next gates
1. Task16: reconstruct or update from latest exact `main`, then prove normal CI + Task16 PostgreSQL18 + Exact Accounting + Pinned Forwardproxy on one SHA.
2. Task13: fresh isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only kill, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once accounting.
3. Karing: reproducible real-client import/parse/connect/cleanup smoke with disposable credentials and redacted logs.
4. Independent review: RLS, privilege separation, retention/purge safety, accounting/session lineage and secret redaction.
5. Production promotion only after all exact-head gates pass, fresh encrypted backup and independent rollback state exist, then staged deploy and postflight verification.

Never claim completion from stale reports, older heads, partial evidence or dirty worktrees.
