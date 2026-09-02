# PVNaive — Canonical Project Status

Last verified: 2026-09-02

This file records only exact GitHub and fresh read-only Production observations. Historical worker reports are evidence only.

## Repository

- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13 draft PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; all three exact-head GitHub gates are green, but final real HTTP/1.1 + HTTP/2 rehearsal remains mandatory before merge.
- Task16 draft PR #81 latest branch commit: `3c4310335ab4907d28bac995bba1be3545e14f6e`; previous exact head `b96c659...` had TDD, Exact Accounting and Pinned Forwardproxy green, while repository-wide CI failed at schema21 RLS coverage `43/42`.
- Task16 branch-only fix now makes the health contract expect 43 RLS tables for `expected_version >= 21`; schema17–20 expectations remain unchanged. Fresh CI gates are queued and not yet credited.

## Production

Fresh read-only probe on `pv-primary` succeeded:

- `pvnaive-api` readiness: `db=ok`, `schema=ok`, `ready=true`, `status=ready`.
- liveness: `service=pvnaive-api`, `status=ok`.
- `docker` active; `nginx` inactive.
- 45-minute journal scan had no `panic`, `fatal`, `schema mismatch`, or `segmentation fault` matches.

No Production deploy, restart, reload, migration, DB write, credential change, backup mutation, or rollback mutation was performed.

## Safety invariants

Never derive accounting/session/IP truth from client headers or untrusted input. Production changes require fresh encrypted backup, rollback state, exact artifact provenance, and postflight verification. Development/test work must not run on Production.

## Worker assignments / blockers

- `pv-primary`: Production-only audit/backup/rollback/deploy lane.
- `TrPaqet`: Task13 final live rehearsal lane; currently blocked by missing rehearsal tooling/plan capacity.
- `pv-worker-main`: Task16 exact-head CI and PostgreSQL18 lane; latest branch fix is queued for fresh verification.

Persistent `AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers and do not override this file.