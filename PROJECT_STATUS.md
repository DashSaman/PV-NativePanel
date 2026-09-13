# PVNaive — Canonical Project Status

Last updated: 2026-09-13 04:37 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel`; verified pre-documentation `main`: `2b460e471807f3bfdb68f3b4a2e6a39db3d1133d`.
- No commit-specific workflow runs are published for that docs-only main tip; do not claim the docs lineage CI-green from absent status data.
- Task16/schema21 remains the last validated merged runtime integration: PR #81 merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- Compare `7efa359c...` → `2b460e47...` shows only `PROJECT_STATUS.md`, `HANDOFF.md`, and `CONTINUE_HERE.md` changed, so the reviewed schema21/accounting runtime code is still current on main.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; current-main reconstruction plus fresh real pinned-Caddy HTTP/1.1 + HTTP/2 kill/accounting rehearsal remain mandatory.
- PR #101 / Karing is OPEN/DRAFT/mergeable at exact head `6168c8445ce5b9358c9cbae12be98951d3153845`. Exact-head workflows are now all SUCCESS: CI `34727433471`, WS1 Exact Accounting `34727433445`, WS1 Pinned Forwardproxy `34727433444`. UI copy-action wiring and independent real Karing import/parse/connect/cleanup smoke remain outstanding, so do not merge.
- Legacy PR #4 remains historical until PR #101 fully supersedes its UI/client handoff behavior.

## Security/accounting reconciliation
- Independent issue #99 is COMPLETED/CLOSED this cycle with no patch required.
- Review confirmed schema21 ENABLE+FORCE RLS, bounded app read, owner-only materialization, hard `p_limit` 1..500, exact 30-day retention boundary, trusted finalized accounting + exact peer/service-term lineage, explicit confirmation-gated maintenance purge, strict older-than-30-days deletion, and rollback refusal when retained history exists.
- Exact reviewed head `904e17c...` has SUCCESS on Task16 Schema21 TDD `34587866720`, WS1 Exact Accounting `34587866787`, WS1 Pinned Forwardproxy `34587866795`, and normal CI `34587866885`.

## Worker / coordinator reconciliation
- No fresh executable completion receipt arrived for Task13, Karing real-client smoke, or Production issue #100 after the previous dispatches.
- Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are still 2026-08-27 S04-era historical ledgers; where they conflict with fresh exact-SHA GitHub evidence, current canonical files and exact-head evidence win.
- TrPaqet remains the persisted isolated Task13 rehearsal lane. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence.

## Production truth
- No fresh connected command-level Production receipt exists. Deployed SHA/schema, services/listeners, Caddy state, session-control socket, backup freshness, rollback snapshot and postflight prerequisites are not re-verified.
- Production was not mutated: no deploy, restart/reload, DB/schema change, credential change, Caddy change, backup mutation or rollback mutation.
- Promotion sequence remains: read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified main, open PRs, PR #101 exact-head workflows, persistent S04-era reports, and issue #99/#100 evidence.
- Completed and closed independent security/accounting review #99 with exact-lineage and CI evidence.
- Confirmed PR #101 exact head is fully green at repository workflow level; kept it DRAFT because UI wiring + real Karing client smoke are still required.
- Canonical docs refreshed; remaining executable lanes are Task13, Karing client handoff, and Production read-only audit.

## Next executable gates
1. PR #101: add current-main UI copy action without overwriting newer `RuntimeNaive.tsx`; preserve TDD discipline, rerun exact-head CI/accounting/pinned-forwardproxy, then independent real Karing import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version and redacted logs. Only then close PR #4 and consider merge.
2. Task13 PR #64: reconstruct validated delta on current main, rerun exact-head gates, then isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Issue #100: connected read-only Production audit first. No backup/deploy mutation until runtime gates are green.
