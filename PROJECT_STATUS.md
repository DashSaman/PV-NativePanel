# PVNaive — Canonical Project Status

Last updated: 2026-09-13 06:39 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel`; verified pre-documentation `main`: `8fcd6b4f05c717e739e91836f540fcb1d39295c6`.
- No pull-request workflow runs are published for that docs-only main tip; do not claim the docs lineage CI-green from absent data.
- Task16/schema21 remains the last validated merged runtime integration: PR #81 merged as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN/DRAFT/non-mergeable at stale exact head `3fc14825e1b164bad558decaef47f56b792e81af`; current-main reconstruction plus fresh real pinned-Caddy HTTP/1.1 + HTTP/2 kill/accounting rehearsal remain mandatory.
- PR #101 / Karing is the canonical reconstruction lane at exact head `216d53670066033403fe95f61b0402bb710186a3`.
- Legacy PR #4 was closed without merge as superseded by #101; retain only as historical evidence.

## Karing verification
- TDD trail remains valid: RED `0db472d960e0004c8c60cfc3294c35b3f5e64ee1`; UI implementation `1996ed732a3551125d521c29c865f616b49a5dc6`; browser-safe test correction `216d53670066033403fe95f61b0402bb710186a3`.
- Exact-head repository gates on `216d5367...` are now all SUCCESS: CI `34732580376`, WS1 Exact Accounting `34732580468`, and WS1 Pinned Forwardproxy `34732580377`.
- PR #101 remains DRAFT because an independent real Karing import/parse/connect/cleanup smoke is still mandatory before merge. Use disposable credentials, record client/platform/version and exact generated-profile SHA-256, redact secrets, and prove cleanup/revoke.

## Security/accounting reconciliation
- Independent issue #99 is COMPLETED/CLOSED with no patch required.
- Schema21 review remains accepted: forced RLS, bounded owner-only reads/materialization, finalized trusted accounting lineage, exact 30-day retention, explicit confirmation-gated purge, and rollback refusal while retained history exists.

## Worker / coordinator reconciliation
- No fresh executable completion receipt arrived for Task13, real Karing client smoke, or Production issue #100 after the previous dispatches.
- Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain 2026-08-27 S04-era historical ledgers; where they conflict with fresh exact-SHA GitHub evidence, current canonical files and exact-head evidence win.
- TrPaqet remains the persisted isolated Task13 rehearsal lane. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence.

## Production truth
- Issue #100 still contains read-only dispatches rather than a returned connected command-level audit receipt. Deployed SHA/schema, services/listeners, Caddy state, session-control socket, backup freshness, rollback snapshot and postflight prerequisites remain unverified this cycle.
- Production was not mutated: no deploy, restart/reload, DB/schema change, credential change, Caddy change, backup mutation or rollback mutation.
- Promotion sequence remains: read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified current main, all open PRs, exact-head PR #101 workflows, Task13, issue #100, and the persistent coordinator/deployment reports.
- Confirmed all three exact-head repository gates are green for PR #101.
- Closed legacy Karing PR #4 without merge because PR #101 is the canonical current-main reconstruction.
- Refreshed canonical docs and re-dispatched remaining independent gates.

## Next executable gates
1. PR #101: run independent real Karing import/parse/connect/cleanup smoke on exact `216d5367...`; if successful, reconcile against latest main, keep exact-head gates green, then review for merge.
2. Task13 PR #64: reconstruct validated delta on current main, rerun exact-head gates, then isolated HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no restart/reload and exactly-once final accounting.
3. Issue #100: connected read-only Production audit first. No backup/deploy mutation until runtime gates are green.
