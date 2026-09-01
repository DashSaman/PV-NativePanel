# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `392a153f14dc311f6d1dffe60e6b4d4da5f8cb17`.
- Prior exact-main push CI on `cce50c4b...`: **SUCCESS**; the new documentation merge SHA has not yet surfaced a commit-associated workflow run, so no green claim is made for it.
- Open roadmap PR: draft #64, exact head `485657d5232da27cbcc4a2c5b8018a4c6b42d3e9`.
- Exact-head #64 status: CI **SUCCESS**, Exact Accounting **SUCCESS**, Pinned Forwardproxy **SUCCESS**.
- Old draft #4 is not current roadmap work.
- Production remains on Task15/schema20; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

Fresh read-only probe on `pv-primary` succeeded after the active SentinelX slot moved back to Primary:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, and `pvnaive-telemetry-agent` are all `active`;
- `GET http://127.0.0.1:8080/api/v1/health/ready` returned `db=ok`, `schema=ok`, `ready=true`;
- journal inspection is permission-blocked for the SentinelX account and is not credited as a fresh pass;
- `/opt/pvnaive/deploy-src` provenance/process-env inspection is also permission-blocked in this run, so earlier verified Task15 provenance remains the latest provenance evidence.

No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 now contains:

- exact full-tuple registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- live CONNECT registration after accounting open + trusted peer recording, with teardown unregister;
- Unix-domain control-server primitive with bounded/strict JSON, exact kill, `0660` socket mode, stale-socket handling and owned-socket cleanup.

The exact published head `485657d...` has all three GitHub gates green. This is validation of the current data-plane increment, not completion of Task13.

PR #64 must remain draft until all of the following are proven: reload-safe Caddy-owned listener lifecycle at `/run/pvnaive/session-control.sock`; narrow Unix-socket ownership/group model allowing the intended API but not widening accounting access; API RBAC/ownership/IDOR/CSRF without credential mutation; UI kill action; full exact-tree gates after those changes; real HTTP1/HTTP2 target-only kill; sibling survival; forged tuple rejection; repeated-kill idempotency; credential survival; exactly-once normal final accounting.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: two connected hosts, `TrPaqet` and `pv-primary`; Free plan allows one active host.

- `pv-primary`: active/executable now and reserved for Production read-only verification/deployment gates only.
- `TrPaqet`: connected/healthy but currently non-active (`upgrade_required`); assigned to resume Task13 lifecycle/permission/API work as soon as the development slot is available.
- Task16: assigned to the next independently executable development Worker; currently no second slot exists.

Human action is required only for true parallelism or to move execution back to the development Worker: disconnect/switch hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

## Current execution rules

- no force push/reset of main;
- no secrets in Git/chat/CI/evidence;
- no fake usage/online/IP/session facts;
- no Production-as-test-database;
- no Runtime mutation without validate → backup → apply → verify → rollback;
- no Production mutation without fresh backup/rollback state;
- no feature becomes DONE without fresh verification evidence;
- accounting/session truth must remain exact under retry, race, kill and disconnect.

## Exact next sequence

1. Keep #64 draft despite its green exact-head gates.
2. Resume Task13 on `TrPaqet` when that development slot is executable: Caddy reload-safe listener lifecycle and narrow permission boundary.
3. Add API RBAC/ownership/IDOR/CSRF and UI action with TDD.
4. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 rehearsal.
5. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and Production postflight access.
6. Execute Task16 on the next independent Worker when capacity exists.