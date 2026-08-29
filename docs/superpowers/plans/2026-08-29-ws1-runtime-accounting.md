# WS1 Runtime Accounting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: use TDD for each production behavior; this plan is executed inline on `parallel/ws1-runtime-accounting` per the Owner's autonomous-work instruction.

**Goal:** Build the exact authenticated-CONNECT telemetry/accounting path required for honest usage, first-use, presence/session state and hard shared ServiceTerm quota.

**Architecture:** Patch the exact pinned forwardproxy commit only at the authenticated forwarding primitive so only successful payload writes are counted. Emit cumulative session counters over a dedicated telemetry-only Unix socket to an unprivileged telemetry agent. PostgreSQL 18 owns the append-only/idempotent ledger, session projection, completeness flag, first-connect transition, and shared ServiceTerm quota/read model.

**Tech Stack:** Go 1.25, PostgreSQL 18, Caddy v2.11.2-naive, pinned `klzgrad/forwardproxy@d62c80d3dd2c706b6b87579844d2397bddd18317`, AF_UNIX HTTP/JSON, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-08-29-naive-usage-accounting-forwardproxy-addendum.md` plus `docs/PARALLEL_WORKSTREAM_PROMPTS_2026-08-29.md` WS1 contract.

## Global Constraints

- Stable Runtime credential UUID is identity; username is diagnostics only.
- No access-log or handler-read estimation.
- No released migration rewrite; current schema head is 0008, so WS1 uses 0009+.
- Telemetry socket cannot expose Runtime management operations.
- Duplicate exact event is idempotent; conflicting sequence/counter regression/gaps fail closed.
- Missing final counter marks accounting incomplete; never estimate bytes.
- First-use starts only on accepted authenticated CONNECT telemetry.
- Presence uses stale timeout.
- Finite concurrent sessions share one ServiceTerm quota budget.
- No `latest` dependencies or moving upstream refs.
- Production remains unchanged until full CI/rehearsal and guarded rollout evidence.

---

### Task 1: Telemetry event/state contracts

**Files:**
- Create: `internal/telemetry/event.go`
- Create: `internal/telemetry/event_test.go`

**Produces:** strict `Event`, `Delta`, `SessionState`, validation/apply semantics.

- [ ] Write failing tests for malformed UUID/identity/time/auth marker, duplicate, conflict, gap, out-of-order sequence, counter regression, reconnect session isolation and final completeness.
- [ ] Observe RED in CI.
- [ ] Implement minimum state machine.
- [ ] Observe GREEN.

### Task 2: Presence and shared quota projection

**Files:**
- Create: `internal/telemetry/projection.go`
- Create: `internal/telemetry/projection_test.go`

**Produces:** `ReadModel`, stale session semantics, shared ServiceTerm quota calculation.

- [ ] RED tests for online/offline, LastOnline, session count, stale timeout, unlimited quota, finite quota, depleted state, renewal/new term isolation and concurrent-session shared budget.
- [ ] Implement minimal projection and keep tests green.

### Task 3: PostgreSQL 18 append-only ledger

**Files:**
- Create: `db/migrations/0009_runtime_accounting.up.sql`
- Create: `db/migrations/0009_runtime_accounting.down.sql`
- Modify: `db/migrations/SHA256SUMS`
- Create: `tests/db/runtime_accounting_migration_test.sh`
- Modify: `.github/workflows/ci.yml`

**Produces:** immutable telemetry sample ledger, session state, term usage, trusted ingest/authorize/read functions and completeness state.

- [ ] RED DB test against schema 8.
- [ ] Add 0009 migration with strict constraints/functions.
- [ ] Test duplicate/conflict/gap/regression, first-use, stale sessions, shared quota and term isolation on PostgreSQL 18.

### Task 4: Telemetry-only Unix socket agent

**Files:**
- Create: `internal/telemetry/server.go`
- Create: `internal/telemetry/server_test.go`
- Create: `internal/telemetry/postgres.go`
- Create: `cmd/pvnaive-telemetry-agent/main.go`

**Produces:** fixed `/v1/accounting/authorize`, `/v1/accounting/event`, `/v1/accounting/health` only; AF_UNIX listener with strict JSON and no management capability.

- [ ] RED route/security tests including forbidden management paths/arbitrary fields.
- [ ] Implement handler/store and process startup.
- [ ] Add build/rehearsal wiring.

### Task 5: Pinned forwardproxy patch and exact byte semantics

**Files:**
- Create: `third_party/forwardproxy/UPSTREAM_COMMIT`
- Create: `third_party/forwardproxy/forwardproxy-pvnaive-accounting.patch`
- Create: `third_party/forwardproxy/README.md`
- Create: `scripts/build/build-pinned-accounting-caddy.sh`
- Create: `tests/accounting/forwardproxy_accounting_patch_test.sh`
- Modify: `.github/workflows/ci.yml`

**Produces:** reproducible patch/build contract pinned to commit and Caddy release; exact successful payload byte counting for HTTP/1 and padded HTTP/2/3 path.

- [ ] RED static/build contract first.
- [ ] Add patch with trusted username→Runtime UUID mapping and telemetry client.
- [ ] Verify patch applies with zero fuzz, upstream commit matches, Caddy build/modules succeed, and padding partial-write tests pass.

### Task 6: Runtime metadata rendering and quota enforcement evidence

**Files:**
- Modify only narrow `internal/naiveruntime/**` / runtime credential adapter code as required.
- Add targeted tests.

- [ ] RED proves every active credential has exactly one UUID mapping when telemetry enabled.
- [ ] Render metadata atomically with existing credential mutation/rollback path.
- [ ] Prove quota denial never removes the last Basic auth credential.

### Task 7: Full verification, report, PR

- [ ] `gofmt` gate.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] PostgreSQL 18 migration/rehearsal.
- [ ] forwardproxy patch/build tests.
- [ ] full repository CI at exact HEAD.
- [ ] Update `docs/agent-reports/WS1_RUNTIME_ACCOUNTING.md` with RED/GREEN/CI evidence and exact remaining blockers.
- [ ] Open PR to `main`.
