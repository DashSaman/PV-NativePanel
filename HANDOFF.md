# PVNaive — Canonical Handoff

Last updated: 2026-09-02 (automation verification)

## Repository truth

- `main`: `0b921abe9b2bd1d827023f494fda11a407fe34d3`.
- Main CI run `33623286003`: **SUCCESS**.
- Task13 draft PR #64 exact head: `3fc14825e1b164bad558decaef47f56b792e81af`; exact-head GitHub gates are green, but the live HTTP/1.1 + HTTP/2 rehearsal is still the merge gate.
- Task16 draft PR #81 remains **DO NOT MERGE** until repository-wide PostgreSQL18 CI is green and the Task16 contract is explicitly executed.

## Production truth

Production remains on Task15/schema20. `pv-primary` was connected but inactive during this run because the SentinelX Free plan allows only one active host. No fresh command-level Production health claim is made for this run. No Production deploy, restart, reload, migration, DB write, credential rotation or other mutation occurred.

## Task13 execution

`TrPaqet` was reconciled to exact head `3fc14825...`. The required rehearsal binaries were built with Go 1.26.3, then `tests/stages/Task13_api_session_kill_rehearsal.sh` was invoked. It stopped before test execution because the worker lacks `jq` (`ERROR: missing jq`). This is a blocker, not a pass; #64 remains draft.

Required proof remains: target-only termination, sibling survival, forged-tuple rejection, repeated-kill idempotency, credential survival, no kill-triggered Caddy restart/reload, and exactly-once final accounting.

## Task16 execution

Continue from issue #79 / PR #81. Preserve trusted schema17 peer/accounting lineage, tenant forced RLS, exact 30-day retention, hard server-side read maximum 500, explicit maintenance-only purge confirmation, migration checksums and disposable rollback. Do not accept client-provided IP/session facts.

## Persistent reports

`AGENT_HANDOFF.md` and `ops/DEPLOYMENT_PROGRESS.md` remain historical S04-era ledgers and are evidence only.

## Worker assignments

- `TrPaqet`: Task13 final rehearsal; blocked on missing `jq` on the currently executable slot.
- `pv-worker-main`: Task16 schema21/PG18 lane when its slot is executable.
- `pv-primary`: Production-only audit/backup/rollback/deploy lane; keep separate from development/testing.

## Safety rules

No force-push or history rewrite on `main`; no secrets in Git/chat/evidence; no fabricated usage/online/IP/session facts; no Production-as-test-database; no Production mutation without fresh encrypted backup, rollback state, exact artifact provenance, and postflight verification.