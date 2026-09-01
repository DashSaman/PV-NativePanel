# PVNaive — Canonical Handoff

Last updated: 2026-09-01

Resume from this file plus `CONTINUE_HERE.md`, `PROJECT_STATUS.md`, newest `ops/evidence/*`, exact GitHub `main`, open PRs and fresh Production health. Older stage/worker checkpoints are historical evidence.

## Repository / release truth

- Current `main`: `1aea1961515ab86428231499490202af4aef5e97`.
- Exact-main push CI `33541904912`: **SUCCESS**.
- Open roadmap PR: draft #64, exact head `9a863258455473605a370f7ad4964043a0df92a1`.
- New exact-head #64 workflows for the API publication are running; prior-head green results are not reused.
- Published lifecycle patch Git blob `b5889058caf4312df5508655193a6275c4ae5a1e` exactly matches the Worker-tested file.
- New API increment wires ownership-checked exact-session kill through the authenticated customer/RLS read model; exact tuple fields come from trusted session state, not request data. Worker RED→GREEN, focused race tests, full Go suite and diff check passed.
- Old draft #4 is not current roadmap work.
- Production remains on Task15/schema20; no Task13 code has been deployed.
- Task12: **DONE / Production**, schema17.
- Task14: **DONE / Production**, schema19.
- Task15: **DONE / Production**, schema20.
- Task35 BUG-001/002/003: **DONE in main**.
- Task13: **IN PROGRESS / PR #64 draft**.
- Task16: **IN PROGRESS / schema21 design gate blocked until retention/pagination are server-bounded**.

## Production state

No fresh Production probe is credited in this checkpoint. The sole SentinelX active slot is currently assigned to development Worker `TrPaqet`, so `pv-primary` cannot be queried without moving capacity. This does not invalidate prior health evidence, but it is not fresh enough to authorize a deployment.

No Production mutation, restart, reload, migration, credential rotation, DB write or deployment occurred.

## Task13

PR #64 now contains:

- exact full-tuple registry: runtime credential + node + boot + session;
- sibling survival, forged tuple rejection, repeated-kill idempotence and unregister semantics;
- local session-control protocol/client;
- live CONNECT registration after accounting open + trusted peer recording, with teardown unregister;
- Unix-domain control-server primitive with bounded/strict JSON, exact kill, `0660` socket mode, stale-socket handling and owned-socket cleanup;
- reload-safe Caddy lifecycle in `0003`: accounting-only acquire, shared registry/socket across overlapping configs, predecessor-safe cleanup, final-lease cleanup, `caddy.CleanerUpper` integration;
- ownership-checked API kill route: user/session IDs enter via the route, but exact runtime credential/node/boot/session tuple is selected from the authenticated customer/RLS active-session read model; no credential mutation is performed.

New API validation passed on `TrPaqet` with isolated Go 1.26.3: clean behavioral RED, focused route/tuple tests, `go test -race ./internal/httpapi ./internal/sessioncontrol ./internal/sessionkill -count=1`, full `go test ./... -count=1`, and `git diff --check`.

PR #64 must remain draft until all of the following are proven: narrow Unix-socket ownership/group model allowing the intended API but not widening accounting/local access; stronger handler-level ownership/IDOR/CSRF failure tests; UI kill action; full exact-tree gates after those changes; real HTTP1/HTTP2 target-only kill; sibling survival; forged tuple rejection; repeated-kill idempotency; credential survival; exactly-once normal final accounting.

## Task16

No fresh current-main Task16 completion is credited. First implementation step remains RED tests for retention >30 days and oversized pagination, followed by minimal server-side constants/clamps, tenant RLS, trusted peer/accounting facts only, final-accounting synchronization, maintenance-only purge, coherent migration/checksums, PostgreSQL18 proof and rollback safety.

## Worker capacity / assignments

Fresh host listing: two connected hosts, `TrPaqet` and `pv-primary`; Free plan allows one active host.

- `TrPaqet`: active development Worker; assigned Task13 permission/API lane.
- `pv-primary`: Production-only; currently non-executable while the development slot is active.
- Task16: assigned to the next independently executable development Worker; currently no second slot exists.

Human action is required only for true parallelism or simultaneous fresh Production probing: switch/disconnect hosts as needed or increase the SentinelX active-host limit. Do not move development/testing onto Production.

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

1. Keep #64 draft while exact-head checks complete.
2. Continue Task13 on `TrPaqet`: narrow socket permission/service-group proof, then stronger handler-level ownership/IDOR/CSRF failure tests and UI action with TDD.
3. Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates and fresh real HTTP1+HTTP2 rehearsal.
4. Merge only the exact verified Task13 tree; deploy only after fresh encrypted backup + rollback snapshot and Production postflight access.
5. Execute Task16 on the next independent Worker when capacity exists.
