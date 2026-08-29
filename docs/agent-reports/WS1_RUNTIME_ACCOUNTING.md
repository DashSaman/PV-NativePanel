# WS1 Runtime / Exact Accounting / First Connect / Hard Quota

Last updated: 2026-08-29

## Work identity

- AGENT: WS1 Runtime/Accounting Lead
- TASK-ID: PVN-045..PVN-051 (primary), with first-successful-CONNECT integration for existing customer activation
- GOAL: exact Runtime-UUID-bound Direct Naive accounting, restart/reconnect safety, real presence/session projection, trusted first successful CONNECT, and hard shared quota enforcement
- STARTING_MAIN_SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- BRANCH: `parallel/ws1-runtime-accounting`
- STARTING_BRANCH_SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`

## Existing partial work discovered

- Open draft PR #8 / branch `s06-exact-accounting` exists.
- It is diverged from current `main`: 44 commits ahead, 134 commits behind; merge-base `e5cf47afe6b1ea3672187e7afb17f44efeedfc3f`.
- Useful accounting work there is limited to an older `0007_exact_accounting` migration/test/plan. Current `main` already owns schema versions 0007 and 0008, so that migration MUST NOT be copied under its old version.
- The older branch does not contain the required forwardproxy/runtime telemetry implementation.
- Main already contains `internal/accountingboundary/boundary_test.go` and the approved pinned-forwardproxy accounting design addendum proving an outer handler cannot infer successful upload writes.

## Architecture chosen

1. Exact bytes are observed inside the pinned `klzgrad/forwardproxy` forwarding primitive, after trusted Basic authentication and at actual successful destination writes.
2. Runtime credential UUID is configured alongside each active Basic credential and is the billing identity; username is diagnostics only.
3. Caddy/forwardproxy receives a dedicated telemetry-only Unix socket. It is separate from the privileged Runtime Agent management socket and exposes only fixed accounting operations.
4. Telemetry events use node ID + boot ID + session ID + monotonic sequence + cumulative upload/download counters + authenticated CONNECT marker.
5. Ingest persists append-only/idempotent samples and derives deltas transactionally. Duplicate same-sequence/same-payload is idempotent; same-sequence conflict, sequence regression/gap, and counter regression fail closed.
6. A lost final cumulative counter marks the affected accounting interval/session incomplete; no byte estimate is synthesized.
7. Presence is derived from accepted telemetry with explicit stale timeout; unclosed stale sessions become offline.
8. Quota is keyed by immutable ServiceTerm. Concurrent sessions share one term budget. New/renewed ServiceTerm starts a separate budget and does not inherit prior usage.
9. First-use activation is triggered only from an accepted authenticated successful-CONNECT event bound to Runtime UUID.
10. Telemetry health is part of accounting completeness. Hard quota cannot be claimed exact while the telemetry path is unhealthy.

## TDD state

- Existing proof: `internal/accountingboundary/boundary_test.go` demonstrates body reads != successful remote writes on partial write.
- Next RED set will cover event validation/idempotency/conflicts/regression/session staleness/shared quota before production code.

## Security boundary

- Existing privileged management socket remains unchanged for inspect/validate/apply/rollback.
- New telemetry socket is fixed-purpose, AF_UNIX only, strict JSON, bounded request size, no arbitrary path/service/command fields, and no route that can mutate Caddy/service state.
- No secret-bearing metadata is emitted in accounting events.

## Production

- No production mutation performed.
- Production rollout remains blocked until code + PostgreSQL 18 + pinned forwardproxy build/rehearsal are green and a production-safe staged preflight/backup/rollback path is available.

## Current status

STATUS: IN_PROGRESS

CHANGES:
- Audited canonical docs, latest main commits, open PRs, existing accounting proof, schema head, and stale exact-accounting branch.
- Created dedicated branch from current main.

FILES MODIFIED:
- `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md`

TESTS:
- No new test executed yet on this branch.

RESULT:
- Work can proceed without reusing stale branch history.

NEXT EXACT STEP:
- Add RED Go contracts for telemetry event validation, cumulative idempotency/conflict/regression semantics, reconnect/restart identity, stale sessions, first-CONNECT gating, and shared quota projection.

BLOCKERS:
- None for implementation.
- Local container cannot resolve github.com, so verification will use GitHub Actions on branch/PR commits rather than a local clone.

## Commits

- pending

## Remaining Tasks

- [ ] RED telemetry/accounting core tests
- [ ] core event/session/quota engine GREEN
- [ ] telemetry-only Unix socket server/client boundary
- [ ] PostgreSQL schema 0009 append-only/idempotent ledger + session/read model
- [ ] PostgreSQL 18 tests and migration checksum wiring
- [ ] first-successful-CONNECT bridge
- [ ] pinned forwardproxy patch + reproducible build/provenance/SHA contract
- [ ] exact padding/partial-write accounting tests
- [ ] hard shared quota enforcement tests
- [ ] reconnect/restart/incomplete-accounting semantics tests
- [ ] full CI regression
- [ ] production evidence if safe and available
- [ ] PR to main

READY_FOR_INTEGRATION: NO
