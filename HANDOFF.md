# PVNaive — Canonical Handoff

Last updated: 2026-09-05 22:39 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` after this run's docs refresh: `4ab8c6068df2738d4a775d114f9c75da0ef0bae4`.
- No combined status rows or workflow runs were returned for the exact docs head; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates are green, but the fresh real rehearsal failed before assertions because PostgreSQL rejected `security_invoker`.
- Task16: draft #81, current GitHub head `b96c65903e5fc314284ea777ceea236913a03842`; fresh exact-head repository-wide green evidence and metadata reconciliation remain pending.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Production state

Only one SentinelX host (`TrPaqet`) was connected during this run. Fresh read-only inspection confirmed a clean Task13 checkout and a rehearsal failure in the development lane; it did not produce a Production health pass. `pv-primary` was not available.

No Production mutation, restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until a fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not deploy without fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current Task13 rehearsal blocker: `psql: ERROR: unrecognized parameter "security_invoker"` (return code 3).

## Persistent reports / worker capacity

Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts. One active SentinelX slot was available; Task13 lane used `/opt/pvnaive-task13-run11` and was clean at exact head `3fc14825...`. There was no active Production lane in this run.

## Next assignments

- Task13: isolate the rehearsal compatibility issue without weakening RLS/security semantics, then rerun full HTTP/1.1 + HTTP/2 proof outside Production.
- Task16: use a clean checkout from current main, reconcile only generic schema21 fixtures, preserve schema20 Task15 fixtures, align PR metadata, and rerun all four gates on one SHA.
- Production: remain read-only until both task lanes are fully green; then perform fresh encrypted backup and rollback preflight before any promotion.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.