# PVNaive — Canonical Handoff

Last updated: 2026-09-06 04:40 Asia/Tehran

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub `main`, open PRs, newest evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main` at inspection: `46d906d8ab9428d0a6b9c106f9b510e081268406`.
- No combined status rows or workflow runs were returned for the exact inspected main head; post-merge CI is not credited.
- Task13: draft #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head CI / Exact Accounting / Pinned Forwardproxy pass, but fresh real HTTP/1.1 + HTTP/2 rehearsal is still pending.
- Task16: draft #81, exact head `b96c65903e5fc314284ea777ceea236913a03842`; PG18 TDD, Exact Accounting and Pinned Forwardproxy pass, while repository-wide CI `33626300697` fails. PR body/base metadata is stale and must be reconciled.
- Production remains on Task15/schema20; no Task13 or schema21 code is deployed.

## Production state

No fresh command-level Production audit was obtained in this run. No Production health pass is claimed and no Production mutation occurred.

No restart, reload, migration, DB write, credential rotation, backup mutation, rollback mutation or deployment occurred.

## Gates and blockers

- Do not merge #64 until fresh exact-head HTTP/1.1 + HTTP/2 rehearsal proves target-only kill, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered restart/reload and exactly-once accounting.
- Do not merge #81 until generic CI, Task16 PG18, Exact Accounting and Pinned Forwardproxy all pass on one exact published SHA, with schema20-specific fixtures preserved.
- Do not deploy without a fresh encrypted backup, rollback state, exact artifact provenance and postflight verification.
- Current blocker is execution capacity/access: no fresh Production probe or active worker receipt was available in this run.

## Persistent reports / worker capacity

Persistent coordinator/worker reports are historical unless corroborated by exact GitHub state and fresh receipts. The latest corroborated plan keeps `TrPaqet` for Task13 development and `pv-primary` Production-only when the single executable slot is available.

## This run — 2026-09-06 04:40 Asia/Tehran

- Verified current main ref `46d906d8...`, open PRs, exact-head CI/status presence, and persistent coordinator/worker reports.
- Reconciled Task16 CI: three exact-head gates pass; repository-wide CI fails; no completion credited.
- Refreshed canonical status and continuation docs; no runtime code integrated.
- No merge/deploy/migration/restart/reload/DB write/credential/backup/rollback mutation performed.

## Next assignments

- Task13: reconstruct the validated delta onto exact current main and run fresh HTTP/1.1 + HTTP/2 rehearsal outside Production.
- Task16: fix generic schema21 fixture drift on a clean branch, preserve schema20 Task15 fixtures, align PR base/head metadata and rerun all four gates on one SHA.
- Worker lanes when available: TrPaqet → Task13 protocol rehearsal; worker with PostgreSQL18 → Task16 CI/fixture reconciliation; independent worker → full regression/static/security review; Production-only lane → read-only health then backup/rollback preflight only after gates pass.

Keep truthful accounting/session semantics under retry, race, kill and disconnect.
