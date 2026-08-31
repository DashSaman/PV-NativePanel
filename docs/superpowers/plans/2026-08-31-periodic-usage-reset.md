# Periodic Usage Reset Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement restart-safe periodic traffic reset for `none/daily/weekly/monthly/yearly/custom` plans without losing accounting bytes or rotating Runtime/Subscription identity.

**Architecture:** Schema16 snapshots reset policy per ServiceTerm into an RLS-protected schedule row, uses UTC-only due calculation, persists next-due/cursor/retry state, and executes due resets atomically through a SECURITY DEFINER scheduler function. Scheduled execution reuses the exact-accounting reset core, records `reason=scheduled`, immutable attempts and audit history, and advances cadence from the actual successful reset time so downtime never causes retroactive byte loss. A small Go runner in the API polls the due-execution function; multiple instances are safe via row locks/SKIP LOCKED.

**Tech Stack:** Go 1.25, PostgreSQL 18, Bash migration gates, systemd-managed PVNaive API, Vitest/TypeScript for operator capability display.

**Spec:** `KNOWN_ISSUES.md` ACCOUNTING-003 and `docs/superpowers/plans/2026-08-30-production-parity-reconciliation.md` Task #9.

## Global Constraints

- Start from deployed main `1b2b1feeea054060c607106a65e98e15258a0289`, schema 15.
- UTC is the only scheduler timezone policy in this task; no host-local-time dependence.
- A late/restarted scheduler resets at the actual successful execution time; it MUST NOT retroactively zero bytes observed after the nominal due timestamp.
- Manual, bulk and scheduled reset must not rotate Runtime passwords, Runtime credential UUIDs or Subscription tokens.
- Pending first-use terms have no due reset until `starts_at` is proven.
- `quota_depleted` terms are eligible for a scheduled reset; successful reset may reactivate them using Task7 semantics.
- Terminal expired/ended/revoked terms are not reset.
- Multiple API instances must not double-run the same due period.
- Migration up is non-destructive and must pass the repository destructive-SQL scanner.
- Rollback is fail-closed while scheduled-reset history exists.

---

### Task 1: Schema16 reset schedule and UTC cadence math

**Files:**
- Create: `db/migrations/0016_periodic_usage_reset.up.sql`
- Create: `db/migrations/0016_periodic_usage_reset.down.sql`
- Create: `tests/db/periodic_usage_reset_migration_test.sh`
- Modify: `db/migrations/SHA256SUMS`
- Modify latest-schema gates under `tests/db/`, `tests/stages/`, and `.github/workflows/ci.yml`.

**Interfaces:**
- Produces `pvnaive.service_term_reset_schedules` with `strategy`, `custom_days`, `timezone_name='UTC'`, `anchor_at`, `next_due_at`, `last_attempt_at`, `last_completed_at`, `retry_after_at`, `last_error`, `consecutive_failures`.
- Produces immutable `pvnaive.scheduled_usage_reset_attempts` history.
- Produces `pvnaive.next_usage_reset_due(timestamptz,text,integer) -> timestamptz`.
- Produces trigger `pvnaive.init_service_term_reset_schedule()` on ServiceTerm insert and start activation.

- [ ] **Step 1: Write failing schema16 migration test**

```bash
[[ -f db/migrations/0016_periodic_usage_reset.up.sql ]]
# migrate 15 -> 16, assert both tables are FORCE RLS, UTC policy is enforced,
# daily/weekly/monthly/yearly/custom next-due math is deterministic, and rollback 16 -> 15 works only with empty history.
```

- [ ] **Step 2: Run RED**

Run: `bash tests/db/periodic_usage_reset_migration_test.sh`
Expected: FAIL because schema16 files/functions do not exist.

- [ ] **Step 3: Implement non-destructive schema16 foundation**

```sql
CREATE TABLE pvnaive.service_term_reset_schedules (...);
CREATE TABLE pvnaive.scheduled_usage_reset_attempts (...);
CREATE FUNCTION pvnaive.next_usage_reset_due(...);
CREATE FUNCTION pvnaive.init_service_term_reset_schedule() RETURNS trigger ...;
```

Backfill existing ServiceTerms from their current Plan reset policy because periodic execution did not exist before schema16; after backfill the schedule row is the frozen policy snapshot. Custom/non-plan terms default to `none`.

- [ ] **Step 4: Run GREEN and rollback proof**

Run: `bash tests/db/periodic_usage_reset_migration_test.sh`
Expected: `PVNAIVE_PERIODIC_USAGE_RESET_MIGRATION_TEST=PASSED`.

### Task 2: Shared exact reset core and scheduled DB executor

**Files:**
- Modify: `db/migrations/0016_periodic_usage_reset.up.sql`
- Modify: `db/migrations/0016_periodic_usage_reset.down.sql`
- Extend: `tests/db/periodic_usage_reset_migration_test.sh`

**Interfaces:**
- Produces private `pvnaive.direct_naive_accounting_reset_core(uuid,timestamptz,bigint)`; no PUBLIC/app EXECUTE.
- Existing `pvnaive.direct_naive_accounting_reset(...)` remains Owner-context-only and delegates to the core.
- Produces `pvnaive.run_due_scheduled_usage_resets(timestamptz,integer,bigint)` executable by `pvnaive_app` but capable only of persisted, actually-due schedules.
- Uses a fixed non-login system actor solely for scheduled audit/event attribution.

- [ ] **Step 1: Add RED cases for due/deferred/restart/concurrency semantics**

```sql
-- due daily schedule => exactly one reason='scheduled' reset event
-- second run at same instant => zero duplicate success
-- telemetry-stale/reservation-pending => deferred attempt, cursor not advanced
-- later safe retry => success, next_due based on actual execution time
-- manual reset after schedule creation => schedule reconciles to manual last_reset_at
```

- [ ] **Step 2: Run RED**

Run: `bash tests/db/periodic_usage_reset_migration_test.sh`
Expected: FAIL because executor/core do not exist.

- [ ] **Step 3: Implement executor**

Use `FOR UPDATE OF schedule SKIP LOCKED`, reconcile any newer `direct_naive_accounting_terms.last_reset_at`, call the private reset core with actual `p_now`, append mutation/reset/audit rows atomically on success, keep `next_due_at` unchanged on transient defer, and set `retry_after_at = p_now + interval '1 minute'`.

- [ ] **Step 4: Run GREEN**

Run: `bash tests/db/periodic_usage_reset_migration_test.sh`
Expected: all exactly-once/defer/restart cases PASS.

### Task 3: Go scheduler runner and API lifecycle integration

**Files:**
- Create: `internal/resetscheduler/runner.go`
- Create: `internal/resetscheduler/runner_test.go`
- Modify: `cmd/pvnaive/main.go`
- Modify: `cmd/pvnaive/main_test.go`

**Interfaces:**
- `resetscheduler.New(db *sql.DB, interval time.Duration, limit int) (*Runner,error)`
- `(*Runner).RunOnce(ctx context.Context) (BatchResult,error)`
- `(*Runner).Run(ctx context.Context, logf func(string,...any))`
- Default production cadence: 30 seconds, batch limit 50.

- [ ] **Step 1: Write RED unit tests**

```go
func TestRunOnceCallsDueResetFunctionAndAggregatesResults(t *testing.T) {}
func TestRunContinuesAfterTransientDatabaseErrorAndStopsOnContext(t *testing.T) {}
```

- [ ] **Step 2: Run RED**

Run: `go test ./internal/resetscheduler ./cmd/pvnaive -count=1`
Expected: FAIL because package/runner wiring does not exist.

- [ ] **Step 3: Implement minimal runner and start it from API run context**

The runner performs no shell/service mutation and never handles raw Runtime/Subscription secrets.

- [ ] **Step 4: Run GREEN**

Run: `go test ./internal/resetscheduler ./cmd/pvnaive -count=1`
Expected: PASS.

### Task 4: Product capability truth and operator contract

**Files:**
- Modify: `internal/customer/product_catalog_store.go` or service projection to mark reset enforcement available.
- Modify: `internal/httpapi/customer_product.go`
- Modify: `web/src/productApi.ts` and relevant tests only if current UI needs capability truth.
- Modify: `docs/API_FA.md`, `KNOWN_ISSUES.md`, `PROJECT_STATUS.md`, `HANDOFF.md` after verification.

**Interfaces:**
- Plan/API projection reports `reset_enforcement_available=true` only once Task9 implementation is present.
- Existing reset strategy values remain unchanged.

- [ ] **Step 1: Add RED capability test**
- [ ] **Step 2: Run focused RED**
- [ ] **Step 3: Implement capability truth**
- [ ] **Step 4: Run focused GREEN**

### Task 5: Full verification and release gates

**Files:** no new production behavior.

- [ ] **Step 1: Run formatting/vet/all Go tests**

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
```

- [ ] **Step 2: Run Web tests/build**

```bash
(cd web && npm test && npm run build)
```

- [ ] **Step 3: Run full PostgreSQL18 CI DB sequence including periodic reset gate**
- [ ] **Step 4: Verify migration SHA256 manifest and pinned forwardproxy boundary**
- [ ] **Step 5: Commit only after all fresh verification is green**

### Task 6: PR, merge and controlled Production rollout

**Files:** release evidence/doc status as needed.

- [ ] **Step 1: Publish exact tested tree; verify remote tree SHA**
- [ ] **Step 2: Require `CI`, `WS1 Exact Accounting`, `WS1 Pinned Forwardproxy` success on exact PR head**
- [ ] **Step 3: Merge with expected-head lock; require main push CI success**
- [ ] **Step 4: Build exact merged R1 schema16 artifact and verify inner/outer checksums**
- [ ] **Step 5: Take encrypted Production schema15 config+DB backup**
- [ ] **Step 6: Stop management API only, migrate 15→16, set expected16, deploy exact R1 with outer rollback**
- [ ] **Step 7: Independent postflight: counts, RLS, scheduler runner health/log evidence, Caddy invariant, public panel/root, runtime/accounting sockets**
- [ ] **Step 8: Prove scheduler on Production without resetting real customer traffic by using a transaction-safe/read-only due query or an isolated disposable fixture; do not mutate existing customer usage merely for acceptance.**

## Self-review

- Spec coverage: restart safety, persisted cursor, UTC policy, idempotent/exactly-once execution, audit/history and identity invariants are each assigned to a task.
- No retroactive boundary reset is permitted; this avoids dropping bytes accumulated during downtime.
- Manual/bulk resets reconcile the periodic cursor so a manual reset cannot be followed by an immediate stale scheduled reset.
- Type/function names above are the canonical Task9 interfaces.
- Placeholder scan is clean; no unfinished implementation markers are intentionally left in this plan.
