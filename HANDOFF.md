# PVNaive — Canonical Handoff

Last updated: 2026-09-05 19:38 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` after this run's docs commits: `091a1b1dd7b3cbb908f2ecb5c68b65b24e0e4e0f`.
- The inspected pre-run main head had no workflow run or status row; post-merge CI is not credited for that head.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; focused tests and shell validation pass, but real rehearsal fails before assertions because PostgreSQL rejects `security_invoker`.
- Task16: draft #81, GitHub head `3c4310335ab4907d28bac995bba1be3545e14f6e`; body/base metadata is stale and fresh exact-head repository-wide green evidence is incomplete.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Production state

Only one SentinelX host (`TrPaqet`) was connected during this run. No fresh command-level Production audit was possible and no Production health claim is credited. No external HTTPS health claim is made.

No Production mutation, restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until a fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not deploy without fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current Task13 rehearsal blocker: `ERROR: unrecognized parameter "security_invoker"`.

## Persistent reports / worker capacity

Persistent worker reports are historical unless corroborated by exact GitHub state and fresh receipts. One active SentinelX slot was available; Task13 lane used `/opt/pvnaive-task13-run11` and was clean at exact head `3fc14825...`. There was no active Production lane in this run.

## Next assignments

- Task13: isolate the rehearsal compatibility issue without weakening RLS/security semantics, then rerun full HTTP/1.1 + HTTP/2 proof outside Production.
- Task16: use a clean checkout from current main, reconcile only generic schema21 fixtures, preserve schema20 Task15 fixtures, align PR metadata, and rerun all four gates on one SHA.
- Production: remain read-only until both task lanes are fully green; then perform fresh encrypted backup and rollback preflight before any promotion.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
