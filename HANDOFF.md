# PVNaive — Canonical Handoff

Last updated: 2026-09-02 (automation verification)

## Current repository / release truth

- `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; exact GitHub gates green, but final real HTTP/1.1 + HTTP/2 live proof is still required.
- Task16: draft PR #81, exact head `096db51ac10ab25ea530af2c18ff0ef9d6a35a27`; Schema21 TDD and Exact Accounting green; CI database job fails on a generic fixture expecting schema20; Pinned Forwardproxy was re-run after a transient upstream failure.
- Production remains Task15/schema20. No Task13/schema21 code is deployed.

## Production safety

No deploy, restart, reload, migration, DB write, credential change, or backup/rollback mutation was performed in this cycle. Fresh Production health is not asserted because SentinelX execution routing was unavailable. Never use Production as a development or rehearsal lane.

## Active blockers

1. Task13 rehearsal cannot be credited until its preflight dependencies (`jq` and executable development tooling) are available.
2. Task16 requires the remaining generic schema21 fixture correction and a fresh all-green exact-head gate set.
3. SentinelX active-host capacity/routing must allow the appropriate development or Production-only lane before any live command or deploy.

## Assignments

- `TrPaqet`: final Task13 live rehearsal only.
- `pv-worker-main`: Task16 PostgreSQL18/CI correction and verification.
- `pv-primary`: Production audit, encrypted backup, rollback and deploy only.

## Rules

No force-push/reset of main; no secrets in Git/chat/CI; no fabricated accounting/session/IP facts; no feature marked DONE without fresh evidence; preserve exact accounting/session semantics under retry, race, kill and disconnect.
