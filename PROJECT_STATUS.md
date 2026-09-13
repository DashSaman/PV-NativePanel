# PVNaive — Canonical Project Status

Last updated: 2026-09-13 05:41 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel`; verified pre-documentation `main`: `180ffcf98f939cc6c7d0b2278f3d9b6c2d8dbc5c`.
- No combined commit statuses are published for that docs-only main tip; do not claim the docs lineage CI-green from absent status data.
- Task16/schema21 remains the last validated merged runtime integration: PR #81 merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; current-main reconstruction plus fresh real pinned-Caddy HTTP/1.1 + HTTP/2 kill/accounting rehearsal remain mandatory.
- PR #101 / Karing is the canonical reconstruction lane. It advanced from exact head `6168c8445ce5b9358c9cbae12be98951d3153845` to `216d53670066033403fe95f61b0402bb710186a3` this cycle.
- Legacy PR #4 remains historical; do not merge it while #101 is the canonical lane.

## Karing TDD / verification this cycle
- RED commit `0db472d960e0004c8c60cfc3294c35b3f5e64ee1` added only the UI wiring regression test. CI web job failed exactly at `npm test`, proving the missing wiring behavior.
- Minimal implementation commit `1996ed732a3551125d521c29c865f616b49a5dc6` wired `buildKaringSingBoxProfile` into current `RuntimeNaive.tsx`, preserved the raw Naive URI action, and exposed distinct `کپی کانفیگ Karing` / `کپی لینک Naive` actions.
- That implementation made `npm test` green, but `npm run build` failed because the first regression fixture imported `node:fs`, which is not part of the browser TypeScript build environment.
- Test-only correction commit `216d53670066033403fe95f61b0402bb710186a3` removed the Node-only fixture dependency while preserving the same behavior assertion. On exact head `216d5367...`, the CI web job has already passed both `npm test` and `npm run build`; full CI, WS1 Exact Accounting, and WS1 Pinned Forwardproxy were still running at the last check, so do not yet call the whole PR green.
- Independent real Karing import/parse/connect/cleanup smoke is still mandatory before merge. Use disposable credentials, record client/platform/version and exact generated-profile hash, and redact secrets.

## Security/accounting reconciliation
- Independent issue #99 is COMPLETED/CLOSED with no patch required.
- Review confirmed schema21 ENABLE+FORCE RLS, bounded app read, owner-only materialization, hard `p_limit` 1..500, exact 30-day retention boundary, trusted finalized accounting + exact peer/service-term lineage, explicit confirmation-gated maintenance purge, strict older-than-30-days deletion, and rollback refusal when retained history exists.

## Worker / coordinator reconciliation
- No fresh executable completion receipt arrived for Task13, real Karing client smoke, or Production issue #100 after the previous dispatches.
- Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain 2026-08-27 S04-era historical ledgers; where they conflict with fresh exact-SHA GitHub evidence, current canonical files and exact-head evidence win.
- TrPaqet remains the persisted isolated Task13 rehearsal lane. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence.

## Production truth
- Issue #100 still contains dispatches rather than a returned connected command-level audit receipt. Deployed SHA/schema, services/listeners, Caddy state, session-control socket, backup freshness, rollback snapshot and postflight prerequisites remain unverified.
- Production was not mutated: no deploy, restart/reload, DB/schema change, credential change, Caddy change, backup mutation or rollback mutation.
- Promotion sequence remains: read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified main, open PRs, exact Karing branch state, Task13, current-main status, issue #100 comments, and the persistent S04-era reports.
- Advanced PR #101 with a real RED→implementation→build-fix TDD cycle while keeping it DRAFT and Production-isolated.
- Canonical docs refreshed; remaining executable lanes are full exact-head Karing workflows + real-client smoke, Task13 rehearsal/reconstruction, and Production read-only audit.

## Next executable gates
1. PR #101: wait for exact-head CI/accounting/pinned-forwardproxy completion; if all green, run independent real Karing import/parse/connect/cleanup smoke. Merge only when that evidence is attached and the PR is still clean against current main.
2. Task13 PR #64: reconstruct validated delta on current main, rerun exact-head gates, then isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Issue #100: connected read-only Production audit first. No backup/deploy mutation until runtime gates are green.
