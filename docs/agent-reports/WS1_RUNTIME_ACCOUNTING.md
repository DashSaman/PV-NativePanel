# WS1 Runtime / Exact Accounting / First Connect / Hard Quota

Last updated: 2026-08-29

## Work identity

- AGENT: WS1 Runtime/Accounting Lead
- TASK-ID: PVN-045..PVN-051 (primary), with first-successful-CONNECT integration for existing customer activation
- GOAL: exact Runtime-UUID-bound Direct Naive accounting, restart/reconnect safety, real presence/session projection, trusted first successful CONNECT, and hard shared quota enforcement
- STARTING_MAIN_SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- BRANCH: `parallel/ws1-runtime-accounting`
- PR: `#12` (Draft while TDD/integration is in progress)

## Existing partial work discovered

- Open draft PR #8 / branch `s06-exact-accounting` exists.
- It is diverged from current `main`: 44 commits ahead, 134 commits behind; merge-base `e5cf47afe6b1ea3672187e7afb17f44efeedfc3f`.
- Useful accounting work there is limited to an older `0007_exact_accounting` migration/test/plan. Current `main` already owns schema versions 0007 and 0008, so that migration MUST NOT be copied under its old version.
- The older branch does not contain the required forwardproxy/runtime telemetry implementation.
- Main already contains `internal/accountingboundary/boundary_test.go` and the pinned-forwardproxy accounting design addendum proving an outer handler cannot infer successful upload writes.

## Architecture

1. Exact bytes are observed inside the pinned `klzgrad/forwardproxy` forwarding primitive, after trusted Basic authentication and at actual successful destination writes.
2. Runtime credential UUID is the billing identity; username is diagnostics only.
3. Caddy/forwardproxy uses a dedicated telemetry-only Unix socket. It never receives access to the privileged Runtime Agent management socket.
4. Telemetry identity is Runtime UUID + node ID + boot ID + session ID, with a session-wide monotonic sequence and cumulative upload/download counters.
5. Same sequence + same payload is idempotent. Same-sequence conflict, sequence gap/out-of-order and counter regression fail closed.
6. Missing a trusted final cumulative counter never causes estimated bytes. The affected accounting state becomes incomplete.
7. Presence is derived from accepted telemetry with a stale timeout. Stale unclosed sessions are offline and make accounting incomplete.
8. Usage/quota is keyed by immutable ServiceTerm, so renewal/new terms do not inherit prior-term usage.
9. First-use activation is triggered only by the accepted first successful authenticated CONNECT telemetry event.
10. Telemetry health is part of `accounting_complete`; hard-quota exactness is never claimed while telemetry is unhealthy.
11. Concurrent finite-quota traffic will use an atomic shared ServiceTerm reservation before each bounded payload write, followed by exact cumulative commit after the successful write. Pending/unknown reservations are never guessed or silently released.

## Exact byte semantics

- Source of truth: actual successful writes in the pinned forwardproxy tunnel data path.
- Upload: client payload bytes successfully written to the target connection.
- Download: target payload bytes successfully written to the client response/tunnel.
- Read bytes, access-log body sizes, Caddy logs, padding and framing bytes are not billable payload.
- Cumulative counters are reported; the ledger derives exact deltas transactionally.

## TDD evidence

### Event/session protocol

RED:
- CI run `#737` / Actions run `33266024593`.
- `gofmt` passed.
- `go vet` failed at `internal/telemetry/event_test.go:17:69` with `undefined: Event`.

GREEN:
- Implementation commit `d905a60612ef97d359aee950afa14d5ed66e76c9`.
- CI later confirmed `gofmt`, `go vet`, `go test ./...`, and runtime-agent safety rehearsal all passing.

Covered semantics:
- cumulative counters -> exact deltas
- exact duplicate -> zero-delta idempotent replay
- same-sequence changed payload -> conflict
- sequence gap/out-of-order -> rejected
- counter regression -> rejected
- reconnect -> distinct session ID
- restart -> distinct boot ID
- accepted final event closes session without inventing bytes

### Presence/shared quota read model

RED:
- Commit `30b0c2332215991d9fc3bef4c4d2bddec36e2772`.
- CI run `#744` / Actions run `33266159363`.
- `gofmt` passed.
- `go vet` failed at `internal/telemetry/projection_test.go:14:16` with `undefined: BuildReadModel`.

GREEN:
- Implementation commit `e3c24a8adc33931c98a772940ce1722f1773af2e`.
- CI run `#748` / Actions run `33266207254`: Go, Web, PostgreSQL 18 database, full rehearsal and bundle jobs all completed successfully.

Covered semantics:
- shared ServiceTerm budget across simultaneous sessions
- online/offline and session count
- stale timeout
- final session offline but complete
- telemetry failure -> `accounting_complete=false`
- finite/depleted/unlimited quota states
- new ServiceTerm mismatch prevents accidental usage carry
- overflow/invalid projection rejected

## Security boundary

- Existing privileged management socket remains unchanged for inspect/validate/apply/rollback.
- New telemetry socket will be AF_UNIX only, fixed-purpose, strict JSON and bounded request size.
- It will not accept arbitrary path, service name, command or management mutation fields.
- Caddy will have no route to apply config, rollback, execute commands or select arbitrary services/files.
- No password/secret material is emitted in accounting telemetry.

## PostgreSQL plan

Current schema head before WS1: `8`.

WS1 will add a new `0009` migration; released migrations are not rewritten.

Planned DB-owned objects/semantics:
- append-only event ledger
- per-session cumulative state
- per-ServiceTerm exact aggregate usage
- shared finite-quota reservation ledger
- first successful CONNECT activation bridge
- stale/incomplete accounting projection
- narrow authorize/reserve/ingest/read functions
- direct table DML withheld from the application role; normal access goes through narrow functions

## Production

- No production mutation performed.
- Production is not being used as a raw test environment.
- Rollout remains blocked until PostgreSQL 18, telemetry socket, pinned forwardproxy patch/build, exact byte tests and full regression are green.
- If safe Production access becomes available later, mutation requires read-only preflight, versions/schema/SHA capture, encrypted backup, rollback point, staged deployment and postflight.

## Current status

STATUS: IN_PROGRESS

CHANGES:
- Audited canonical docs, latest main commits, open PRs, schema head, runtime code and stale exact-accounting branch.
- Created dedicated branch and Draft PR #12.
- Added implementation plan.
- Added exact telemetry event/session state machine using TDD.
- Added presence/shared-quota read model using TDD.

FILES ADDED/MODIFIED SO FAR:
- `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md`
- `docs/superpowers/plans/2026-08-29-ws1-runtime-accounting.md`
- `internal/telemetry/event.go`
- `internal/telemetry/event_test.go`
- `internal/telemetry/projection.go`
- `internal/telemetry/projection_test.go`

TESTS:
- Event/session RED -> GREEN recorded in CI.
- Presence/shared-quota RED -> GREEN recorded in CI.
- Full existing repository regression at `e3c24a8...` is green, including PostgreSQL 18 and pinned forwardproxy boundary rehearsal.

NEXT EXACT STEP:
- Add PostgreSQL 18 RED migration test for schema 0009 covering append-only/idempotent events, shared concurrent reservations, first successful CONNECT, duplicate/conflict/gap/regression, restart/reconnect/stale semantics, incomplete-accounting truthfulness and ServiceTerm renewal isolation. Then implement migration 0009 to GREEN.

BLOCKERS:
- None for implementation.
- Local container cannot fetch current repo/toolchain dependencies from GitHub; authoritative full verification is being captured in GitHub Actions.

## Commits

- `f0ad99e97a4ecde2ef347d6510ca8e10fbd9f2a6` — initialize WS1 report
- `4e001009b4eb01fe1f99fb724fbec52c779e7776` — implementation plan
- `0d9c87373ca8057abbd79abf1a0cf4c03956f3dc` — gofmt-clean event/session RED tests
- `d905a60612ef97d359aee950afa14d5ed66e76c9` — cumulative telemetry state GREEN
- `30b0c2332215991d9fc3bef4c4d2bddec36e2772` — presence/shared quota RED tests
- `e3c24a8adc33931c98a772940ce1722f1773af2e` — presence/shared quota GREEN

## Remaining Tasks

- [x] RED telemetry/accounting event/session tests
- [x] event/session core GREEN
- [x] RED presence/shared quota tests
- [x] pure read-model GREEN
- [ ] PostgreSQL schema 0009 append-only/idempotent ledger + reservations/session/read model
- [ ] PostgreSQL 18 runtime-accounting migration tests and checksum wiring
- [ ] telemetry-only Unix socket server/client boundary
- [ ] first-successful-CONNECT database bridge
- [ ] pinned forwardproxy patch + reproducible build/provenance/SHA contract
- [ ] exact HTTP/1 and padded HTTP/2/3 byte-accounting tests
- [ ] hard shared quota enforcement integration tests
- [ ] reconnect/restart/incomplete-accounting integration tests
- [ ] full final CI regression at final HEAD
- [ ] production evidence if safe access exists and rollout is actually performed
- [ ] mark PR ready only when integration evidence is complete

READY_FOR_INTEGRATION: NO
