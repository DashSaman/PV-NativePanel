# PVNaive — Canonical Project Status

Last updated: 2026-09-13 01:41 Asia/Tehran

## Verified GitHub state
- Repository `DashSaman/PV-NativePanel` is accessible; `main` is the default branch.
- Verified pre-documentation `main`: `655526a5a572f8d62bc33319d7f1e2c9705bd57b`.
- Exact-main combined status is `pending` with zero published statuses; commit-specific workflow lookup returns no workflow runs. This checkpoint is **not claimed CI-green**.
- Last validated runtime integration remains Task16/schema21: PR #81 merged successfully as `7efa359ccc5745c548cda9590bc5c516e9d5aa9e` from exact head `904e17c4a013e3adb5fb349c70f254ab59c925f8`.
- PR #64 / Task13 remains OPEN, DRAFT and non-mergeable at exact head `3fc14825e1b164bad558decaef47f56b792e81af`; its body records historical exact-head greens, but a current-main reconstruction and fresh real pinned-Caddy HTTP/1.1 + HTTP/2 kill/accounting rehearsal are still mandatory.
- PR #4 / Karing remains OPEN/DRAFT/mergeable at exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, still based on `s04-auth`; a current-main-derived head plus independent real Karing smoke remain mandatory.

## Worker / coordinator reconciliation
- PR #64, PR #4, issue #99 and issue #100 were re-read after the previous dispatches. No newer exact-head completion receipt exists.
- Issue #79 was closed as completed because its Task16/schema21 implementation is already merged and remains the last validated runtime integration.
- Duplicate follow-up issues #97 and #98 were closed as superseded/not-planned; canonical worker lanes are issue #99 (security/accounting) and issue #100 (Production read-only audit).
- Persistent historical `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are materially stale (2026-08-27, S04-era) and must not override this canonical status, `HANDOFF.md`, `CONTINUE_HERE.md`, or fresh exact-SHA evidence.
- TrPaqet remains the persisted isolated Task13 rehearsal lane. Never credit stale, dirty, mixed-head, unpushed, historical, or absent-CI evidence.

## Production truth
- No connected command-level Production tool/evidence is available in this run. Issue #100 contains dispatches but no fresh audit receipt after the latest assignment.
- Therefore deployed SHA/schema, current services/listeners, Caddy binary/state, fresh encrypted backup, independent rollback snapshot, staged promotion and postflight are **not re-verified**.
- Production was not mutated. No restart/reload, DB/schema change, credential change, Caddy change, backup mutation, rollback mutation or deploy was performed.
- Promotion sequence remains: read-only audit → fresh encrypted backup → independent rollback snapshot → exact deploy-SHA lock → staged promotion → health/postflight → retained rollback.

## Actions completed this cycle
- Re-verified exact `main`, combined status, commit workflow visibility, all open PRs, actionable worker issues, and persistent handoff/progress records.
- Reconciled Task16 as completed and removed duplicate issue lanes (#79 completed; #97/#98 superseded).
- Refreshed canonical status/handoff/continue documentation.
- Re-dispatched all active independent lanes against the latest main lineage.

## Next executable gates
1. Task13: reconstruct/rebase validated delta onto current `main` without rewriting validated history, publish a new exact head, rerun exact-head CI/accounting/pinned-forwardproxy/focused gates, then run the isolated real HTTP/1.1 + HTTP/2 rehearsal proving target-only termination, sibling survival, forged-tuple rejection, repeat-kill idempotency, credential survival, no Caddy restart/reload, and exactly-once final accounting.
2. Karing: reconcile export change onto current `main`, rerun exact-head CI, then run a real client import/parse/connect/cleanup smoke using disposable credentials and a non-Production target; record exact profile hash, client/platform/version and redacted logs.
3. Security/accounting: issue #99 must return clean-worktree exact-SHA PASS/FAIL evidence for RLS fail-closed behavior, privilege separation, bounded retention/purge safety, trusted lineage, commit-before-HTTP-success semantics and redaction.
4. Production: issue #100 must first return a connected read-only audit. Only after all runtime gates are green may backup/rollback creation and staged deployment proceed.
