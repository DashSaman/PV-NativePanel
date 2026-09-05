# PVNaive — Canonical Handoff

Last updated: 2026-09-05 23:38 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` after this run's docs refresh: `c0891228e15dfab380b5f574dbe7240a0b67b8ae`.
- No combined status rows or workflow runs were returned for the exact docs head; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; published task gates are green, but the fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending.
- Task16: draft #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; fresh exact-head repository-wide green evidence and metadata reconciliation remain pending.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Production state

No fresh command-level Production audit was obtained in this run. No Production health pass is claimed and no Production mutation occurred.

No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until a fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not deploy without a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current execution blocker is capacity/access: no new verified development or Production probe was available in this run.

## Persistent reports / worker capacity

Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts. The latest corroborated plan keeps `TrPaqet` for Task13 development and `pv-primary` Production-only when the single executable slot is available.

## Next assignments

- Task13: on the next executable development slot, reconstruct the validated delta onto exact current main and run the fresh HTTP/1.1 + HTTP/2 rehearsal outside Production.
- Task16: independently obtain a clean checkout, reconcile only generic schema21 fixtures, preserve schema20 Task15 fixtures, align PR metadata and rerun all four gates on one SHA.
- Production: remain read-only until both task lanes are fully green; then perform fresh encrypted backup and rollback preflight before any promotion.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
