# PVNaive Customer Lifecycle Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first production-safe S05 slice: direct customer CRUD, reusable plans, service-term snapshots, and stable binding from a business service to an existing Naive Runtime credential, without claiming exact usage before the accounting PoC passes.

**Architecture:** Reuse the existing `pvnaive.users` and `pvnaive.plans` tables from migration 0001. Do not reuse the old reseller-only `subscriptions`/`credentials` path for Owner-direct customers because those relations are structurally tied to `pvnaive.resellers`; instead add the spec-approved `service_terms` and `user_runtime_credentials` relations. Keep Runtime mutation semantics inside the existing S04R runtime credential service; this slice only stores stable bindings to already-created Runtime credential UUIDs.

**Tech Stack:** Go 1.x, PostgreSQL 18, pgx, net/http, React/TypeScript/Vitest, existing signed RLS request context, existing S04R Runtime credential service.

**Spec:** `docs/superpowers/specs/2026-08-28-customer-quota-accounting-design.md`

## Global Constraints

- Business user is not a Runtime credential.
- No usage value may be described as exact/billable until PVN-045..047 proves the authenticated Runtime-path meter.
- Existing Caddy/Runtime Agent safe-apply semantics remain unchanged.
- Owner-direct customer rows use a dedicated `system` tenant; reseller purchase/credit tables remain untouched.
- User administrative states exposed by this slice are only `draft`, `active`, `suspended`, `revoked`; `expired` and `quota_depleted` are service-term/effective states, not reasons to overwrite user administration state.
- Default service start policy is `on_first_successful_connection`.
- One active primary Runtime credential binding per service term in R1.
- Username rename and password rotation must not change the stable Runtime credential UUID binding.
- No Caddy restart is introduced by this plan.

---

## File Structure

- `db/migrations/0004_customer_lifecycle_foundation.up.sql` — additive schema bridge for direct customers, service-term snapshots, and Runtime UUID binding.
- `db/migrations/0004_customer_lifecycle_foundation.down.sql` — rollback for only the additive S05 objects/columns.
- `db/migrations/SHA256SUMS` — checksums for migration 0004.
- `tests/db/customer_lifecycle_migration_test.sh` — PostgreSQL 18 migration/RLS/invariant proof.
- `internal/customer/types.go` — domain types and request DTOs.
- `internal/customer/policy.go` — state/start-policy validation and derived access state.
- `internal/customer/policy_test.go` — pure domain tests.
- `internal/customer/store.go` — pgx persistence under existing signed DB request context.
- `internal/customer/store_test.go` — query/contract tests that do not require Runtime mutation.
- `internal/customer/service.go` — orchestration for users, plans, terms, and stable credential binding.
- `internal/customer/service_test.go` — service behavior tests with narrow fake store.
- `internal/httpapi/customers.go` — Owner HTTP handlers and JSON contracts.
- `internal/httpapi/customers_test.go` — auth/CSRF/idempotency/If-Match behavior.
- `internal/httpapi/routes.go` — route registration only.
- `cmd/pvnaive/main.go` — customer store/service wiring.
- `web/src/customers.ts` — typed API client.
- `web/src/customers.test.ts` — frontend API contract tests.
- `web/src/Customers.tsx` — first customer/plan/service UI; usage explicitly unavailable until accounting capability exists.
- `web/src/routes.ts`, `web/src/App.tsx`, `web/src/styles.css` — navigation/route/style integration.

---

### Task 1: Add the additive customer-lifecycle migration

**Files:**
- Create: `tests/db/customer_lifecycle_migration_test.sh`
- Create: `db/migrations/0004_customer_lifecycle_foundation.up.sql`
- Create: `db/migrations/0004_customer_lifecycle_foundation.down.sql`
- Modify: `db/migrations/SHA256SUMS`

**Interfaces:**
- Produces: canonical Owner-direct tenant selected by `tenant_type='system' AND slug='direct'`.
- Produces: `pvnaive.service_terms(id, tenant_id, user_id, plan_id, quota_bytes, duration_seconds, reset_interval_seconds, start_policy, purchased_at, starts_at, first_connected_at, expires_at, ended_at, state, revision, created_at, updated_at)`.
- Produces: `pvnaive.user_runtime_credentials(user_id, service_term_id, runtime_credential_id, role, bound_at, unbound_at)`.
- Produces: `revision bigint` on `pvnaive.users` and `pvnaive.plans`.

- [ ] **Step 1: Write the failing migration contract test**

Create `tests/db/customer_lifecycle_migration_test.sh` following the existing disposable PostgreSQL test harness. Assert after migrating through 0004:

```sql
SELECT count(*) = 1
FROM pvnaive.tenants
WHERE tenant_type = 'system' AND slug = 'direct' AND status = 'active';

SELECT to_regclass('pvnaive.service_terms') IS NOT NULL;
SELECT to_regclass('pvnaive.user_runtime_credentials') IS NOT NULL;

SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema='pvnaive' AND table_name='users' AND column_name='revision'
);
SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema='pvnaive' AND table_name='plans' AND column_name='revision'
);
```

Also insert one direct user, one plan, one `on_first_successful_connection` term, and one Runtime credential; prove that one active primary binding succeeds and a second active primary binding for the same term fails with a uniqueness violation.

- [ ] **Step 2: Run the new DB test and verify RED**

Run:

```bash
bash tests/db/customer_lifecycle_migration_test.sh
```

Expected: FAIL because migration 0004/tables/columns do not exist.

- [ ] **Step 3: Implement migration 0004**

The up migration must be additive and transactional. Core SQL shape:

```sql
-- pvnaive:migration-version 0004
-- pvnaive:migration-name customer_lifecycle_foundation
-- pvnaive:transactional true
-- pvnaive:destructive false
SET LOCAL ROLE pvnaive_owner;

INSERT INTO pvnaive.tenants (tenant_type, slug, display_name, status)
SELECT 'system', 'direct', 'PVNaive Direct', 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM pvnaive.tenants WHERE lower(slug) = 'direct'
);

ALTER TABLE pvnaive.users
  ADD COLUMN revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0);
ALTER TABLE pvnaive.plans
  ADD COLUMN revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0);

CREATE TABLE pvnaive.service_terms (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES pvnaive.tenants(id) ON DELETE RESTRICT,
  user_id uuid NOT NULL,
  plan_id uuid REFERENCES pvnaive.plans(id) ON DELETE RESTRICT,
  quota_bytes bigint CHECK (quota_bytes IS NULL OR quota_bytes > 0),
  duration_seconds bigint NOT NULL CHECK (duration_seconds > 0),
  reset_interval_seconds bigint CHECK (reset_interval_seconds IS NULL OR reset_interval_seconds > 0),
  start_policy text NOT NULL CHECK (start_policy IN ('on_creation','on_first_successful_connection','fixed_timestamp')),
  purchased_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  starts_at timestamptz,
  first_connected_at timestamptz,
  expires_at timestamptz,
  ended_at timestamptz,
  state text NOT NULL CHECK (state IN ('pending','active','expired','quota_depleted','ended','revoked')),
  revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (id, tenant_id),
  UNIQUE (id, tenant_id, user_id),
  FOREIGN KEY (user_id, tenant_id) REFERENCES pvnaive.users(id, tenant_id) ON DELETE RESTRICT,
  CHECK ((starts_at IS NULL AND expires_at IS NULL) OR (starts_at IS NOT NULL AND expires_at > starts_at))
);

CREATE TABLE pvnaive.user_runtime_credentials (
  user_id uuid NOT NULL,
  service_term_id uuid NOT NULL,
  runtime_credential_id uuid NOT NULL REFERENCES pvnaive.naive_runtime_credentials(id) ON DELETE RESTRICT,
  role text NOT NULL DEFAULT 'primary' CHECK (role = 'primary'),
  bound_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  unbound_at timestamptz,
  PRIMARY KEY (service_term_id, runtime_credential_id),
  FOREIGN KEY (service_term_id, user_id) REFERENCES pvnaive.service_terms(id, user_id) ON DELETE RESTRICT
);
CREATE UNIQUE INDEX user_runtime_credentials_one_active_primary_uidx
  ON pvnaive.user_runtime_credentials(service_term_id)
  WHERE role='primary' AND unbound_at IS NULL;
CREATE UNIQUE INDEX user_runtime_credentials_runtime_active_uidx
  ON pvnaive.user_runtime_credentials(runtime_credential_id)
  WHERE unbound_at IS NULL;
```

Add RLS to both new tables using `pvnaive.has_tenant_access(tenant_id, false)`; for the binding table, add a stored `tenant_id` column so RLS does not depend on a join. Grant only `SELECT, INSERT, UPDATE` to `pvnaive_app`; revoke DELETE so binding history is soft-unbound rather than deleted.

Add a trigger or check function that permits a term plan only when `plans.tenant_id IS NULL OR plans.tenant_id = NEW.tenant_id`.

- [ ] **Step 4: Run migration tests and existing DB suite**

Run:

```bash
bash tests/db/customer_lifecycle_migration_test.sh
bash tests/db/migration_test.sh
bash tests/db/auth_migration_test.sh
bash tests/db/naive_runtime_migration_test.sh
```

Expected: PASS.

- [ ] **Step 5: Update migration checksums and commit**

```bash
sha256sum db/migrations/0004_customer_lifecycle_foundation.up.sql db/migrations/0004_customer_lifecycle_foundation.down.sql
# update db/migrations/SHA256SUMS using the repository's existing sorted format
git add db/migrations tests/db/customer_lifecycle_migration_test.sh
git commit -m "feat(db): add customer lifecycle foundation"
```

---

### Task 2: Implement customer domain policy and persistence

**Files:**
- Create: `internal/customer/types.go`
- Create: `internal/customer/policy.go`
- Create: `internal/customer/policy_test.go`
- Create: `internal/customer/store.go`
- Create: `internal/customer/store_test.go`

**Interfaces:**
- Produces: `type User`, `type Plan`, `type ServiceTerm`, `type RuntimeBinding`.
- Produces: `EffectiveAccess(userState string, termState string, startsAt, expiresAt *time.Time, now time.Time) string`.
- Produces store methods `DirectTenantID`, `ListUsers`, `CreateUser`, `UpdateUserState`, `ListPlans`, `CreatePlan`, `CreateServiceTerm`, `BindRuntimeCredential`, `GetCustomerDetail`.

- [ ] **Step 1: Write pure RED policy tests**

Test these exact rules:

```go
func TestEffectiveAccessRequiresActiveAdminAndStartedTerm(t *testing.T) {}
func TestEffectiveAccessReportsSuspendedBeforeCommercialState(t *testing.T) {}
func TestEffectiveAccessReportsExpiredWhenNowAtOrAfterExpiry(t *testing.T) {}
func TestUserAdminStateRejectsExpiredAndDepleted(t *testing.T) {}
func TestOnFirstConnectionTermStartsPendingWithoutExpiry(t *testing.T) {}
```

- [ ] **Step 2: Run only customer package tests and verify RED**

```bash
go test ./internal/customer -run Test -count=1
```

Expected: package/functions missing.

- [ ] **Step 3: Implement minimal domain types/policy**

Use explicit string enums; do not introduce fake usage fields. A customer detail may contain:

```go
type UsageCapability struct {
    Available bool   `json:"available"`
    Reason    string `json:"reason,omitempty"`
}
```

For this plan always return `Available:false` and `Reason:"exact_accounting_not_proven"`.

- [ ] **Step 4: Write store contract tests**

Use the existing pgx/context pattern. Verify every query executes inside the existing signed request context and that optimistic updates use `WHERE id=$1 AND revision=$2` then increment revision.

- [ ] **Step 5: Implement store and run tests**

```bash
go test ./internal/customer -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/customer
git commit -m "feat(customer): add lifecycle domain and store"
```

---

### Task 3: Add customer service orchestration and Runtime UUID binding

**Files:**
- Create: `internal/customer/service.go`
- Create: `internal/customer/service_test.go`

**Interfaces:**
- Consumes: store interfaces from Task 2.
- Produces: `CreateCustomer`, `CreatePlan`, `AssignPlan`, `SetUserState`, `BindExistingRuntimeCredential`.

- [ ] **Step 1: Write RED service tests**

Cover:

```go
func TestCreateCustomerUsesDirectSystemTenant(t *testing.T) {}
func TestAssignPlanSnapshotsQuotaAndDuration(t *testing.T) {}
func TestAssignPlanDefaultsToFirstConnectionStart(t *testing.T) {}
func TestBindRejectsRevokedOrDisabledRuntimeCredential(t *testing.T) {}
func TestBindPreservesRuntimeCredentialUUIDAcrossUsernameRename(t *testing.T) {}
```

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/customer -run 'Test(CreateCustomer|AssignPlan|Bind)' -count=1
```

- [ ] **Step 3: Implement minimal service**

`AssignPlan` copies plan quota/duration/reset values into `service_terms`; later edits to the plan must not rewrite an existing term. `BindExistingRuntimeCredential` accepts only an active Runtime credential and stores its UUID; it never decrypts or returns the Runtime password.

- [ ] **Step 4: Verify GREEN and full Go tests**

```bash
go test ./internal/customer -count=1
go test ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/customer
git commit -m "feat(customer): orchestrate plans terms and runtime bindings"
```

---

### Task 4: Expose guarded Owner API

**Files:**
- Create: `internal/httpapi/customers.go`
- Create: `internal/httpapi/customers_test.go`
- Modify: `internal/httpapi/routes.go`
- Modify: `cmd/pvnaive/main.go`

**Interfaces:**
- `GET /api/v1/customers`
- `POST /api/v1/customers`
- `PATCH /api/v1/customers/{id}`
- `GET /api/v1/plans`
- `POST /api/v1/plans`
- `POST /api/v1/customers/{id}/service-terms`
- `POST /api/v1/service-terms/{id}/runtime-binding`

All mutations require authenticated Owner, CSRF, `Idempotency-Key`; revision-changing PATCH requires `If-Match`.

- [ ] **Step 1: Write RED endpoint tests**

Assert unauthorized=401, non-Owner=403, missing CSRF=403, missing idempotency=400, stale revision=409. Assert JSON never includes Runtime password/secret fields.

- [ ] **Step 2: Verify RED**

```bash
go test ./internal/httpapi -run Customer -count=1
```

- [ ] **Step 3: Implement handlers and wiring**

Return `usage_capability` as unavailable; do not return `used_bytes`, `remaining_bytes`, or percentages as if they were measured.

- [ ] **Step 4: Verify GREEN and all Go tests**

```bash
go test ./internal/httpapi -count=1
go test ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/httpapi cmd/pvnaive/main.go
git commit -m "feat(api): expose customer lifecycle endpoints"
```

---

### Task 5: Add first customer management UI without fake usage

**Files:**
- Create: `web/src/customers.ts`
- Create: `web/src/customers.test.ts`
- Create: `web/src/Customers.tsx`
- Modify: `web/src/routes.ts`
- Modify: `web/src/routes.test.ts`
- Modify: `web/src/App.tsx`
- Modify: `web/src/styles.css`

**Interfaces:**
- Customer create: display name + username.
- Plan create: name/code + quota bytes + duration + optional reset interval.
- Assign plan to customer creates a service term with default `on_first_successful_connection`.
- Bind existing active Runtime credential by UUID/selection.
- Usage panel shows `Exact usage: unavailable — accounting proof pending` until capability is true.

- [ ] **Step 1: Write RED frontend API/route tests**

Test typed request paths, CSRF/idempotency headers, and route `/panel/#/customers`.

- [ ] **Step 2: Verify RED**

```bash
cd web && npm test -- --run
```

- [ ] **Step 3: Implement minimal API client and page**

Do not add QR/subscription rendering yet; those belong to the later subscription plan. Do not show numeric used/remaining traffic.

- [ ] **Step 4: Verify frontend tests/build**

```bash
cd web
npm test -- --run
npm run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src
git commit -m "feat(web): add customer plan and service management"
```

---

### Task 6: Full regression gate and handoff to accounting PoC

**Files:**
- Modify only if evidence/docs need status recording: `PROJECT_STATUS.md`, `WORKLOG.md`, `AGENT_HANDOFF.md`.

- [ ] **Step 1: Run full repository regression**

```bash
go test ./...
cd web && npm test -- --run && npm run build
cd ..
bash tests/db/customer_lifecycle_migration_test.sh
bash tests/db/migration_test.sh
bash tests/db/auth_migration_test.sh
bash tests/db/naive_runtime_migration_test.sh
bash tests/stages/S04R_full_rehearsal.sh
```

Expected: all PASS; no Caddy production mutation.

- [ ] **Step 2: Update status docs with evidence only after green**

Record that PVN-037/038 foundation and business-to-runtime UUID binding are implemented, while exact traffic remains blocked behind PVN-045..047.

- [ ] **Step 3: Commit evidence/docs**

```bash
git add PROJECT_STATUS.md WORKLOG.md AGENT_HANDOFF.md
git commit -m "docs: record customer lifecycle foundation evidence"
```

- [ ] **Step 4: Open/refresh a draft PR**

Keep production deployment gated. The next separate implementation plan is the exact-accounting PoC (`PVN-045..047`); quota enforcement must not begin before that PoC passes.
