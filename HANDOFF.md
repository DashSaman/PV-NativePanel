# PVNaive — Canonical Handoff

Last updated: 2026-09-06 14:42 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` at start of this run: `072d6fb9ddd74fa916516fe757aecb2ae116a36b`; current documentation refresh commits this run: status `1448469372eb4862ba8de27d4ffa3e2cf7fdc322`, continuation `611d3de4b41ae9e030968f38c65cf0d711fd0f44`.
- No exact-main status/workflow evidence was returned; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending.
- Task16: draft #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated runs are green but normal CI `33678134360` failed. Failed jobs were re-run during this run and remain uncredited pending completion.
- PR #4: draft Karing export, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; keep pending until one reproducible real Karing client smoke is captured.
- Documentation-only PRs #91–#95 remain open/stale and are not promotion authority; canonical docs were refreshed directly on `main`.

## Production state

No fresh command-level Production audit was executable in this run. No Production health pass is claimed and no Production mutation occurred.

No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA; failed-job rerun must be observed as successful.
- Do not merge #4 until real Karing client smoke evidence exists.
- Do not deploy without a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current blockers are missing fresh worker receipts, stale PR metadata/heads, limited executable worker capacity and missing executable Production probe in this run.

## Persistent reports / worker capacity

Persistent coordinator/worker reports were searched. They remain historical unless corroborated by exact GitHub state and fresh receipts. Current bounded assignments: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 CI rerun/fix verification; independent worker → regression/security review; `pv-primary` → Production-only when an executable slot is available.

## This run — 2026-09-06 14:42 Asia/Tehran

- Re-verified current main history, open PRs, Task16 exact-head workflow state and persistent reports.
- Re-ran only the failed Task16 CI jobs (`33678134360`); pending result is not credited.
- Confirmed no new validated worker completion or fresh Production receipt.
- Refreshed canonical status/continuation/handoff docs on `main`; no unvalidated runtime/schema work integrated.
- No merge, deploy, migration, restart/reload, DB write, credential, backup or rollback mutation performed.

## Next assignments

- Task16: inspect the failed-job rerun; if still failing, fix only the generic latest-schema/RLS expectation, preserve schema20-specific Task15 fixtures, align PR metadata and rerun all four gates on one exact published SHA.
- Task13: reconstruct onto current main and run fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with full protocol/accounting receipt.
- PR #4: run a real Karing client smoke and attach reproducible evidence.
- Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
- Production-only lane: perform fresh read-only health; only after all gates pass, create fresh encrypted backup + rollback snapshot and then consider deployment.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
