# PVNaive — Canonical Handoff

Last updated: 2026-09-06 13:41 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` at start of this run: `0aac4359993f5e56011da029ba28b62a7d968bd9`; canonical refresh commits in this run: `6d49bf14e6804abfdc5489d0fba845bb18fe1e26` and `099adcc7bfb2e0e312082919e40f6176c63b86f0`.
- No combined status rows or workflow runs were returned for the exact main head; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending and PR base metadata is stale.
- Task16: draft #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; dedicated historical gates exist, but no fresh exact-head full-green proof is present and PR base/body metadata are stale.
- PR #4: draft Karing export, exact head `2501e39dc39e14063b6a501bc96b77bbfcae7384`, base `s04-auth`; keep pending until one reproducible real Karing client smoke is captured.
- Documentation-only PRs #91–#95 remain open/stale and are not promotion authority; canonical docs were refreshed directly on `main` in this run.

## Production state

No fresh command-level Production audit was executable in this run. No Production health pass is claimed and no Production mutation occurred.

No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not merge #4 until real Karing client smoke evidence exists.
- Do not deploy without a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current blockers are missing fresh worker receipts, stale PR metadata/heads, limited executable worker capacity and missing executable Production probe in this run.

## Persistent reports / worker capacity

Persistent coordinator/worker reports were searched. They remain historical unless corroborated by exact GitHub state and fresh receipts. Current bounded assignments: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 generic fixture correction; independent worker → regression/security review; `pv-primary` → Production-only when an executable slot is available.

## This run — 2026-09-06 13:41 Asia/Tehran

- Re-verified current repository branch ref, open PRs #4/#64/#81 plus docs PRs #91–#95, exact main status/workflow responses and persistent reports.
- Confirmed no new validated worker completion or fresh Production receipt.
- Refreshed canonical status/continuation/handoff docs on `main`; no unvalidated runtime/schema work integrated.
- No merge, deploy, migration, restart/reload, DB write, credential, backup or rollback mutation performed.

## Next assignments

- Task16: on a clean PostgreSQL18-capable lane, reconcile to current main, update only generic schema21/latest-schema expectations, preserve schema20-specific Task15 fixtures, align PR metadata and rerun all four gates on one exact published SHA.
- Task13: on a clean compatible development lane, reconstruct onto current main and run fresh HTTP/1.1 + HTTP/2 rehearsal outside Production with full protocol/accounting receipt.
- PR #4: run a real Karing client smoke and attach reproducible evidence.
- Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
- Production-only lane: perform fresh read-only health; only after all gates pass, create fresh encrypted backup + rollback snapshot and then consider deployment.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
