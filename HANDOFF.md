# PVNaive — Canonical Handoff

Last updated: 2026-09-06 06:43 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` at inspection: `453c2fa7159014b8ffcac5545e43da8b6c736c9d`.
- No combined status rows or workflow runs were returned for that exact main head; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; documented exact-head/focused gates pass, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending. PR base/body metadata is stale.
- Task16: draft #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; documented PG18/Exact Accounting/Pinned Forwardproxy gates pass, but repository-wide database CI previously failed and no fresh all-green proof was observed.
- PR #4: draft Karing export; keep pending until one real Karing client smoke is captured.
- Documentation-only PRs are not promotion authority.

## Production state

No fresh command-level Production audit was obtained in this run. No Production health pass is claimed and no Production mutation occurred.

No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not merge #4 until real Karing client smoke evidence exists.
- Do not deploy without a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current blocker is missing fresh worker receipts and missing executable Production probe in this run.

## Persistent reports / worker capacity

Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts. Latest corroborated plan: TrPaqet → Task13 rehearsal; PostgreSQL18-capable worker → Task16 fixture/CI reconciliation; independent worker → regression/security review; pv-primary → Production-only when the executable slot is available.

## This run — 2026-09-06 06:43 Asia/Tehran

- Verified current main ref, open PRs, exact-head status presence and current GitHub evidence.
- Confirmed no fresh CI evidence for the exact inspected main head.
- Reconciled that no worker completion can be credited from historical reports.
- Refreshed canonical status and continuation documentation; no unvalidated runtime/schema work integrated.
- No merge, deploy, migration, restart/reload, DB write, credential, backup or rollback mutation performed.

## Next assignments

- Task13: reconstruct validated delta onto current main and run fresh HTTP/1.1 + HTTP/2 rehearsal outside Production.
- Task16: fix generic schema21 fixture drift on a clean branch, preserve schema20 Task15 fixtures, align PR metadata and rerun all four gates on one SHA.
- PR #4: run a real Karing client smoke and attach reproducible evidence.
- Independent review: inspect Task13/Task16 diffs for security, accounting, RLS and rollback regressions.
- Production-only lane: read-only health first; only after all gates pass, create fresh encrypted backup + rollback snapshot and then consider deployment.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
