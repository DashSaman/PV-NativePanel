# WS1 Runtime / Exact Accounting / First Connect / Hard Quota

Last updated: 2026-08-29

## Work identity

- AGENT: WS1 Runtime/Accounting Lead
- TASK-ID: PVN-045..PVN-051 (primary), with first-successful-CONNECT integration for existing customer activation
- GOAL: exact Runtime-UUID-bound Direct Naive accounting, restart/reconnect safety, real presence/session projection, trusted first successful CONNECT, and hard shared quota enforcement
- STARTING_MAIN_SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- BRANCH: `parallel/ws1-runtime-accounting`
- PR: `#12` — Draft while TDD/integration is in progress
- BASELINE_GREEN_HEAD: `e3c24a8adc33931c98a772940ce1722f1773af2e`
- BASELINE_GREEN_CI: Actions run `33266207254` / CI #748

## Existing partial work discovered

- Open draft PR #8 / branch `s06-exact-accounting` exists.
- It is badly diverged from current `main` and cannot be safely continued as the WS1 integration lane.
- Useful old accounting concepts were reviewed, but released schema numbers are not reused/re-written.
- Current WS1 branch already contained valid event/session and presence/read-model work, so it was continued instead of rebuilt.
- Main already contains the pinned-forwardproxy boundary proof showing an outer Caddy handler cannot infer successful upload writes.

## Pinned upstream / provenance boundary

- Caddy release line: `v2.11.2` / Naive release used by the existing pilot.
- forwardproxy repository: `klzgrad/forwardproxy`.
- forwardproxy commit: `d62c80d3dd2c706b6b87579844d2397bddd18317`.
- Existing pinned release archive SHA256: `19eccb7321dd877a5fb4a3dba6ef1b745185188b616c96cc6201f1a1fc0380a8`.
- The pinned source proves trusted Basic auth precedes CONNECT target ownership and that exact successful payload writes are only visible inside `serveHijack` / `dualStream` / `flushingIoCopy`.
- No `latest` artifact or moving upstream ref will be used for the accounting build.

## Architecture

1. Exact bytes are observed inside the pinned `klzgrad/forwardproxy` forwarding primitive, after trusted Basic authentication and at actual successful destination writes.
2. Runtime Credential UUID is the billing identity. Username is diagnostics only.
3. The privileged management Runtime Agent remains on `/run/pvnaive/runtime-agent.sock` and retains inspect/validate/apply/rollback. Caddy is never given that socket.
4. Accounting has a separate fixed telemetry socket: `/run/pvnaive/accounting.sock`.
5. The telemetry process is non-root and exposes only four fixed routes: health, authorize, claim and event. It does not accept arbitrary file paths, service names or commands.
6. Telemetry identity is Runtime UUID + node ID + boot ID + session ID, with a session-wide monotonic sequence and cumulative upload/download counters.
7. Same sequence + identical event is idempotent. Same-sequence changed payload, sequence gaps, out-of-order events and counter regressions fail closed and degrade completeness.
8. Missing a trusted final cumulative counter never causes estimated usage. Unknown bytes are never guessed into the ledger.
9. Presence is derived from accepted telemetry with a stale timeout. A stale unclosed session is offline and makes accounting incomplete.
10. Usage/quota is keyed by immutable ServiceTerm, so a renewal/new ServiceTerm starts with its own zero projection and does not inherit prior-period usage.
11. First-use activation occurs only on the accepted sequence-1 authenticated CONNECT event after the producer has successfully authenticated and successfully opened the target connection.
12. Authorization, QR/subscription fetch, account-page access, health checks, failed authentication, failed target dial, Caddy reload and panel login cannot start first-use.
13. Finite quota uses a persistent shared ServiceTerm reservation before every bounded data-path write, followed by exact settlement from the successful write count.
14. Concurrent sessions therefore reserve from one shared budget under a DB lock rather than each seeing the same stale remaining value.
15. If telemetry dies after a reservation but before settlement, the reservation is not silently released and bytes are not guessed. The connection fails closed and accounting truthfulness is degraded instead of fabricating exactness.

## Exact byte semantics

- Source of truth: actual successful writes in the pinned forwardproxy CONNECT tunnel data path.
- Upload: client application payload bytes successfully written to `targetConn`.
- Download: target application payload bytes successfully written into the client CONNECT stream.
- Request reads, access-log sizes, TLS/TCP overhead, Naive framing headers and random padding are not billable.
- HTTP/1 buffered pre-CONNECT bytes must use the actual `targetConn.Write` return count.
- HTTP/2/3 RemovePadding upload counts the unpadded payload successfully written to target.
- HTTP/2/3 AddPadding download excludes the 3-byte framing header and random padding; partial writes count only the successfully delivered payload portion.
- PostgreSQL receives cumulative counters; the append-only ledger persists exact deltas and rejects regressions/conflicts.

## Persistent reservation model

For tracked finite-quota traffic:

1. `authorize` resolves the Runtime UUID to the active binding/ServiceTerm and checks status/expiry/quota without starting first-use.
2. After target dial succeeds, sequence 1 / cumulative 0+0 opens the session and is the only first-use trigger.
3. Before a producer write, `claim` locks the ServiceTerm accounting projection and reserves `min(requested, quota-used-reserved)`.
4. The producer writes at most the granted payload.
5. The matching event with the same next sequence settles the claim using only successful payload bytes and releases any unused reservation.
6. A second session sees the reservation immediately and cannot spend it again.
7. A lost settlement remains a visible/incomplete condition instead of being converted into guessed usage.

## PostgreSQL schema 0009

Added:

- `pvnaive.direct_naive_accounting_terms`
  - exact upload/download aggregate per ServiceTerm
  - persistent `reserved_bytes`
  - `last_online`, `last_telemetry_at`, `accounting_complete`
- `pvnaive.direct_naive_accounting_sessions`
  - Runtime UUID + node/boot/session key
  - last sequence/cumulative counters/final/completeness
- `pvnaive.direct_naive_accounting_claims`
  - pre-write persistent reservation, direction, requested/reserved/settled bytes
- `pvnaive.direct_naive_accounting_events`
  - append-only immutable exact event/delta ledger
  - unique session sequence for idempotency/conflict detection
- Narrow `SECURITY DEFINER` functions:
  - `direct_naive_accounting_authorize`
  - `direct_naive_accounting_claim`
  - `direct_naive_accounting_ingest`
  - `direct_naive_accounting_read`
- Direct table privileges are withheld from `pvnaive_app`; only the narrow typed functions are executable.
- FORCE-RLS source tables are resolved through a private transaction-local signed system context. The helper functions themselves are revoked from PUBLIC and `pvnaive_app`; tenant IDs are never accepted from Caddy telemetry.
- Existing migrations 0001..0008 are untouched.

## Telemetry security boundary

Implemented repository-side components:

- `internal/telemetry/listener.go`
  - accepts only exact `/run/pvnaive/accounting.sock`
  - refuses path escape and refuses replacing a non-socket object
  - socket mode `0660`
- `internal/telemetry/socket.go`
  - strict JSON / unknown fields rejected
  - bounded request body
  - fixed routes only
  - management-shaped routes return 404
- `cmd/pvnaive-telemetry-agent/main.go`
  - refuses root execution
  - connects only to the existing loopback PostgreSQL/pvnaive_app boundary
  - no shell/service/config execution path
- `ops/systemd/pvnaive-telemetry-agent.service`
  - `User=pvnaive`, no capabilities, `NoNewPrivileges`, strict filesystem/kernel hardening
  - only `/run/pvnaive` writable

## TDD evidence

### Event/session protocol

RED:
- CI run #737 / Actions `33266024593`; compile failed on missing `Event` as intended.

GREEN:
- implementation `d905a60612ef97d359aee950afa14d5ed66e76c9`.

Covered:
- cumulative -> delta
- exact duplicate idempotency
- same-sequence conflict
- sequence gap/out-of-order
- counter regression
- restart via new boot ID
- reconnect via new session ID
- final event without invented bytes

### Presence/shared quota read model

RED:
- commit `30b0c2332215991d9fc3bef4c4d2bddec36e2772`
- CI #744 / Actions `33266159363`; missing `BuildReadModel` as intended.

GREEN:
- commit `e3c24a8adc33931c98a772940ce1722f1773af2e`
- CI #748 / Actions `33266207254`: Go, Web, PostgreSQL 18 database, rehearsal and bundle all GREEN.

Covered:
- online/offline/current session count
- stale timeout
- telemetry failure -> `accounting_complete=false`
- finite/depleted/unlimited quota projection
- ServiceTerm mismatch prevents period carry-over

### Telemetry socket + concurrent budget RED

RED commits:
- `a1939c51d96f3a36ab23069aa1bb692b7df6e795` — fixed telemetry socket isolation tests
- `7919690701a1a4f9dc1b7d98fc5e3fedfa744e4e` — concurrent shared-budget tests

The RED tests referenced not-yet-existing `NewTelemetryHandler` / `NewSharedQuotaBudget` and therefore intentionally could not compile before implementation.

Implementation:
- `057503dfd8bf9c507b971f3646a37674c5797be5` — telemetry-only handler
- `d7d770963ce34af46339b2cc49353dd6837dec36` — atomic shared-budget unit primitive
- `821dfc977ab65d456b201f84ab00ee61a3462fa3` — fixed authorize/claim/event routes and strict protocol

### PostgreSQL 18 accounting RED

- `62c14ff0d504baf9630725c1948df56d6c36c57e` adds `tests/db/direct_naive_accounting_migration_test.sh`.
- Test requires schema 9 and proves:
  - authorize does not start first-use
  - sequence-1 successful CONNECT starts first-use exactly
  - exact duplicate idempotency
  - same-sequence conflict
  - two sessions share a 100-byte budget as 80 + 20, never 80 + 80
  - settlement stores exactly 100 bytes and zero reservation
  - quota state becomes depleted
  - event ledger is immutable
- `81efd28cb95021b9cebef94eeed125257b013b87` adds a migration manifest test that prints exact SHA256 values until 0009 is wired into `SHA256SUMS`.
- Migration implementation is committed but authoritative PG18 GREEN is still pending CI execution/fixes.

## Files added/modified in current checkpoint

Existing WS1 core:
- `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md`
- `docs/superpowers/plans/2026-08-29-ws1-runtime-accounting.md`
- `internal/telemetry/event.go`
- `internal/telemetry/event_test.go`
- `internal/telemetry/projection.go`
- `internal/telemetry/projection_test.go`

New checkpoint files:
- `internal/telemetry/socket.go`
- `internal/telemetry/socket_test.go`
- `internal/telemetry/quota.go`
- `internal/telemetry/quota_test.go`
- `internal/telemetry/store.go`
- `internal/telemetry/listener.go`
- `internal/telemetry/migration_manifest_test.go`
- `cmd/pvnaive-telemetry-agent/main.go`
- `ops/systemd/pvnaive-telemetry-agent.service`
- `db/migrations/0009_direct_naive_exact_accounting.up.sql`
- `db/migrations/0009_direct_naive_exact_accounting.down.sql`
- `tests/db/direct_naive_accounting_migration_test.sh`

## Commits

Existing WS1:
- `f0ad99e97a4ecde2ef347d6510ca8e10fbd9f2a6` — initialize WS1 report
- `4e001009b4eb01fe1f99fb724fbec52c779e7776` — implementation plan
- `0d9c87373ca8057abbd79abf1a0cf4c03956f3dc` — event/session RED
- `d905a60612ef97d359aee950afa14d5ed66e76c9` — event/session GREEN
- `30b0c2332215991d9fc3bef4c4d2bddec36e2772` — presence/read-model RED
- `e3c24a8adc33931c98a772940ce1722f1773af2e` — presence/read-model GREEN

Current continuation:
- `a1939c51d96f3a36ab23069aa1bb692b7df6e795` — telemetry socket RED
- `7919690701a1a4f9dc1b7d98fc5e3fedfa744e4e` — concurrent shared quota RED
- `057503dfd8bf9c507b971f3646a37674c5797be5` — telemetry-only handler
- `d7d770963ce34af46339b2cc49353dd6837dec36` — atomic shared budget
- `62c14ff0d504baf9630725c1948df56d6c36c57e` — PostgreSQL accounting RED test
- `76ffc8533dacafa1b697d1c24d00eea949f5227c` — migration 0009 exact ledger
- `fd80401d304cbf4bb4790ba2919558db7169b084` — migration 0009 rollback
- `81efd28cb95021b9cebef94eeed125257b013b87` — migration manifest SHA contract
- `9b5881d6c703c74fd640819bcbdc3428620a2771` — narrow PostgreSQL store
- `821dfc977ab65d456b201f84ab00ee61a3462fa3` — authorize/claim/event socket protocol
- `8bea55a0fc61f2f527c071ce8d7267389dccb4dd` — fixed Unix listener
- `7c29234ae62959e7d2534a3bee872b4fc700e979` — non-root telemetry agent
- `041dd9c8df58d05c0862c6a3e264c89d6b7f0551` — hardened telemetry systemd unit

## CI / verification

- Baseline full regression at `e3c24a8...`: GREEN, Actions `33266207254`.
- Connector-authored continuation commits have not automatically produced a new Actions run yet.
- Planned safe trigger: reopen Draft PR #12 so the `pull_request` workflow evaluates the current head without merging or touching Production.
- New CI must validate format, vet, Go tests, PostgreSQL 18 migration/function behavior, forwardproxy patch/tests/build and full repository regression before integration readiness.

## Production

- No Production mutation has been performed.
- No Production system is being used as a raw test environment.
- Safe Production access is not available through the current tool context, so there is no production evidence to claim.
- If deployment is later performed, the required preflight/backup/rollback/postflight sequence from the Owner prompt remains mandatory.

## Remaining Tasks

- [x] canonical repo/branch/PR/CI audit
- [x] event/session cumulative semantics
- [x] online/session/read-model semantics
- [x] telemetry-only handler boundary
- [x] separate non-root telemetry agent/socket implementation
- [x] schema 0009 initial append-only ledger/session/projection/reservation implementation
- [x] first-successful-CONNECT activation in schema 0009 event path
- [x] persistent shared ServiceTerm quota reservation model
- [ ] add exact 0009 SHA256 entries to migration manifest from authoritative run
- [ ] make PostgreSQL 18 schema/function behavior GREEN
- [ ] wire direct accounting migration test explicitly into CI
- [ ] pinned forwardproxy patch inside repository
- [ ] trusted Runtime UUID mapping in patched forwardproxy config
- [ ] authorize before CONNECT response / sequence-1 after successful target dial
- [ ] exact HTTP/1 prebuffer + normal stream successful-write accounting
- [ ] exact padded HTTP/2/3 payload accounting tests
- [ ] claim -> write -> settle hard quota enforcement in forwardproxy
- [ ] producer telemetry-failure fail-closed tests
- [ ] pinned reproducible Caddy v2.11.2 build + patch SHA256 + binary SHA256 + provenance
- [ ] build telemetry agent in CI/rehearsal
- [ ] restart/reconnect/stale/incomplete integration evidence against PostgreSQL
- [ ] full final CI regression at final HEAD
- [ ] Production evidence only if secure production access is actually available
- [ ] final PR body/report update and mark ready for review

## Blockers

- Local container cannot fetch the private repository/toolchain from GitHub due DNS/network limitations; authoritative execution is therefore through GitHub Actions.
- Current Connector commits did not auto-trigger Actions. This is operational, not an implementation blocker; Draft PR #12 can be safely reopened to create a pull-request workflow run.

## Next Exact Step

1. Trigger CI for the current Draft PR head.
2. Capture exact migration SHA values / Go compile errors / PostgreSQL 18 errors from the run and fix to GREEN.
3. Then implement and test the pinned forwardproxy patch and reproducible Caddy build without changing Production.

STATUS: IN_PROGRESS
READY_FOR_INTEGRATION: NO
