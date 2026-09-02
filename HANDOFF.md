# PVNaive — Canonical Handoff

Last updated: 2026-09-03

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, exact GitHub refs, newest CI evidence and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Release truth

- Current `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Task13: draft PR #64, exact head `3fc14825e1b164bad558decaef47f56b792e81af`; GitHub gates green, real HTTP/1.1 + HTTP/2 proof pending.
- Task16: draft PR #81, exact head `3c4310335ab4907d28bac995bba1be3545e14f6e`; TDD/Exact Accounting/Pinned Forwardproxy green, repository-wide CI red at `RLS coverage check failed: 43/42`.
- Task12, Task14, Task15 and Task35 security fixes remain done only where already in main/Production evidence supports them.

## Production evidence

Fresh read-only HTTP probes on `pv-primary` returned:

- `/api/v1/health/ready`: HTTP 200, `db=ok`, `schema=ok`, `ready=true`, `status=ready`;
- `/api/v1/health/live`: HTTP 200, `service=pvnaive-api`, `status=ok`.

Docker inventory was not readable from the SentinelX execution context because the Docker socket denied access. Do not extend the claim beyond these HTTP probes. No Production mutation, restart, reload, migration, DB write, credential change, backup mutation or rollback mutation occurred.

## Promotion gates

Do not merge #64 or #81 on partial evidence. Task13 needs a fresh real protocol rehearsal proving target-only termination, sibling survival, forged-tuple rejection, idempotency, credential survival, no kill-triggered Caddy restart/reload and exactly-once final accounting. Task16 needs full exact-head CI green and explicit PostgreSQL18 coverage.

Any deployment requires fresh encrypted backup + rollback state, exact artifact provenance, controlled apply and postflight verification.

## Persistent reports / assignments

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` are historical S04-era ledgers and do not override current GitHub/Production evidence.

- `TrPaqet`: Task13 rehearsal lane.
- `pv-worker-main`: Task16 latest-schema CI fixture correction and full exact-head rerun when executable.
- `pv-primary`: Production-only audit/backup/rollback/deploy lane.

## Rules

No force-push/reset of main, no secrets in Git/chat/CI/evidence, no fabricated usage/IP/session facts, no Production-as-test-database, and no feature marked DONE without fresh verification.
