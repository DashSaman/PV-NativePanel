# PVNaive — Canonical Project Status

Last updated: 2026-09-05 21:39 Asia/Tehran

This file records verified repository and Production truth. Historical worker/stage notes are evidence only; exact GitHub state, exact-head CI and fresh Production observations override them.

## Product / safety invariants

PVNaive remains standalone-first. Never fabricate usage/online/IP/session history. Never rotate credentials/tokens from read-only flows. Production changes require a fresh encrypted backup + rollback state, intended migrations only, exact artifact provenance and postflight verification.

## Repository truth

- Repository: `DashSaman/PV-NativePanel`.
- Current `main`: `a0def2cfc1b5489cb9e821ee8d059773169e845d` (docs-only handoff refresh).
- Current `main` combined status: no status rows returned; post-merge CI for this exact head is not credited.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI `33623363327`, WS1 Exact Accounting `33623363299`, and WS1 Pinned Forwardproxy `33623363389` are SUCCESS. Focused worker validation is supplemental only. Fresh real HTTP/1.1 + HTTP/2 rehearsal remains incomplete and previously stopped before assertions at PostgreSQL setup with `unrecognized parameter "security_invoker"`.
- Task16: draft PR #81, current GitHub head `b96c65903e5fc314284ea777ceea236913a03842`; dedicated PostgreSQL18 evidence is historical, while fresh exact-head repository-wide all-green evidence and metadata reconciliation are not complete.
- Docs-only PRs remain open and stale relative to current main.

## Production truth

Production remains on Task15/schema20; no Task13 or schema21 code has been deployed.

Only one SentinelX host is connected (`TrPaqet`). Read-only host inspection showed `docker` and `nginx` inactive, `sentinelx-cloud-core` active, loopback PostgreSQL on 127.0.0.1:5432, and several other listeners; this host snapshot is not a Production health pass and no external HTTPS health claim is made.

No deploy, migration, restart, reload, DB write, credential change, backup mutation or rollback mutation was performed.

## Persistent worker state

- Connected SentinelX hosts: `TrPaqet` only.
- Clean Task13 checkout: `/opt/pvnaive-task13-run11`, detached at exact head `3fc14825...`, dirty=false, ahead=0, behind=0.
- Persistent `AGENT_HANDOFF.md` is historical (last updated 2026-08-27) and conflicts with current GitHub roadmap; it is not used as promotion authority.
- No worker completion is creditable because the full HTTP/1.1 + HTTP/2 rehearsal and exact-head promotion proof are still absent.

## This run — 2026-09-05 21:39 Asia/Tehran

- Verified repository metadata, current main, open PRs, exact-head CI for Task13, one connected worker host, clean Task13 checkout, and persistent handoff files.
- Confirmed Task13 exact-head CI/WS1 gates are green but not sufficient for merge without live protocol rehearsal.
- Confirmed current main has no commit status rows and no fresh post-merge CI credit.
- Performed read-only host/service/port inspection only.
- No merge, deploy, migration, restart/reload, DB write, credential change, backup mutation or rollback mutation performed.

## Immediate execution order

1. Keep #64 and #81 draft / DO NOT MERGE.
2. Fix or isolate the Task13 rehearsal compatibility issue without weakening security semantics, then run fresh HTTP/1.1 + HTTP/2 proof outside Production.
3. Obtain a clean exact-head Task16 checkout, reconcile generic schema21 fixtures without modifying schema20-specific Task15 fixtures, and rerun all Task16 gates on one published SHA.
4. Only if Task13 live proof and Task16 full gates pass: create a fresh encrypted Production backup + rollback state, then consider promotion.

Never claim completion from stale worker reports or partial evidence. No merge/deploy until all exact-head gates and safety prerequisites are green.
