# PVNaive — Canonical Project Status

Last updated: 2026-09-13 03:41 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel`; verified pre-documentation `main`: `7ab4d8a8bbb5af7ecef1135743bda40ef7bfa472`.
- Exact-main combined status is `pending` with zero published statuses and no commit-specific workflow runs; this checkpoint is not claimed CI-green.
- Last validated merged runtime integration remains Task16/schema21: PR #81 merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`. Current-main reconstruction plus fresh real pinned-Caddy HTTP/1.1 + HTTP/2 kill/accounting rehearsal remain mandatory.
- Legacy PR #4 / Karing remains OPEN/DRAFT/non-mergeable at `2501e39dc39e14063b6a501bc96b77bbfcae7384`.
- New PR #101 is the clean current-main Karing reconstruction lane. RED commit `95768324b5f5dd43ddaf153c0b115678fa24a49e` added only the regression test and GitHub Actions web job failed at `npm test` as expected. GREEN commit `6168c8445ce5b9358c9cbae12be98951d3153845` adds the minimal `buildKaringSingBoxProfile` implementation; its web job passes both `npm test` and `npm run build`. Full CI / accounting / pinned-forwardproxy runs are still completing, and UI copy-button wiring plus independent real Karing smoke remain outstanding. Keep PR #101 DRAFT.

## Worker / coordinator reconciliation
- PR #64, PR #4, issue #99 and issue #100 were re-read. No newer worker completion receipt exists after the prior dispatches.
- Issue #99 has 22 comments and was last updated by the 02:40 dispatch; no independent exact-main security/accounting receipt has arrived.
- Issue #100 has 21 comments and was last updated by the 02:40 read-only Production dispatch; no connected Production audit receipt has arrived.
- Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain 2026-08-27 S04-era historical records and must not override fresh exact-SHA evidence or these canonical files.
- TrPaqet remains the persisted isolated Task13 rehearsal lane. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence.

## Production truth
- No connected command-level Production evidence is available in this run. Deployed SHA/schema, services/listeners, Caddy state, backup freshness, rollback snapshot and postflight prerequisites are not re-verified.
- Production was not mutated: no deploy, restart/reload, DB/schema change, credential change, Caddy change, backup mutation or rollback mutation.
- Promotion sequence remains: read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified exact main, open PRs, CI visibility, issue #99/#100 state and persistent reports.
- Advanced Karing independently by creating clean draft PR #101 from current main.
- Proved RED on exact commit `95768324...`: web CI failed at the new missing-builder test.
- Added the minimal builder on exact commit `6168c844...`; web CI is GREEN for tests and build. No Production code outside the Karing helper/test lane was merged.
- Canonical docs refreshed and all blocked lanes re-dispatched.

## Next executable gates
1. PR #101: finish current-main UI wiring without overwriting newer `RuntimeNaive.tsx` behavior; require exact-head full CI/accounting/pinned-forwardproxy success, then independent real Karing import/parse/connect/cleanup smoke with disposable credentials, exact profile hash, client/platform/version and redacted logs. Only then supersede/close PR #4 and consider merge.
2. Task13 PR #64: reconstruct validated delta on current main, rerun exact-head gates, then isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Issue #99: independent clean-worktree review for RLS fail-closed behavior, privilege separation, retention/purge safety, trusted lineage, commit-before-success and redaction.
4. Issue #100: connected read-only Production audit first. No backup/deploy mutation until runtime gates are green.
