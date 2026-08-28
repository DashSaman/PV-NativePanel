# PVNaive Sanaei-style Customer Service Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the low-level Naive Runtime credential screen into an Owner-friendly customer service flow with numeric quota, three expiry modes, one-time Naive delivery, revocable subscription URL, local QR, and a safe path for first-successful-connection activation.

**Architecture:** Keep the existing business user/service-term layer separate from Runtime credentials. Create the customer/service term and Runtime credential in one guarded saga using the existing Runtime Agent validate -> backup -> reload-only -> verify/rollback boundary. Subscription tokens are hash-only at rest, and first-use activation is idempotent and can only be triggered by a local trusted Runtime event, never by a subscription fetch or panel view.

**Tech Stack:** Go, PostgreSQL 18, `database/sql`, existing `runtimecred` + Runtime Agent, React/TypeScript/Vitest, browser-local QR rendering, existing PVNaive auth/CSRF/RLS/idempotency patterns.

**Spec:** `docs/superpowers/specs/2026-08-29-sanaei-style-customer-service-ui-design.md`

## Global Constraints

- Product name remains `PVNaive`; do not rename the repository.
- Business user != Runtime credential.
- Default validity mode is `on_first_successful_connection`.
- Supported validity modes are exactly: `on_creation`, `on_first_successful_connection`, `fixed_expiry`.
- Numeric quota uses the Sanaei-style binary unit contract: `1 GB = 1,073,741,824 bytes`; unlimited is `quota_bytes = NULL`.
- Numeric duration is a positive integer number of days.
- Fixed manual expiry is an absolute RFC3339 timestamp and must be in the future.
- A subscription fetch, panel view, failed authentication, Caddy reload, or config generation must never start a first-use term.
- Exact traffic usage remains unavailable until PVN-045..047 proves authenticated per-credential accounting. Do not fabricate used/remaining traffic.
- Runtime secrets and raw subscription tokens must never be stored in logs/audit or returned from ordinary list/detail APIs.
- QR generation is local; no third-party QR API.
- Existing Caddy apply semantics remain validate -> exact backup -> reload-only -> verify -> rollback.

---

## File Structure

- `internal/customer/validity.go` — request-level validity parsing, GB-to-bytes conversion, initial term timing, first-use activation rules.
- `internal/customer/validity_test.go` — pure tests for the three expiry modes and conversion contract.
- `internal/customer/store.go` — direct-customer, service-term, binding, subscription-token persistence under existing signed DB context.
- `internal/customer/store_test.go` — query/contract tests.
- `internal/customer/service.go` — customer creation saga, binding, first-use activation, subscription token lifecycle.
- `internal/customer/service_test.go` — orchestration and rollback tests.
- `internal/runtimecred/service.go` — add explicit safe abort/rollback and secret resolution for trusted subscription rendering.
- `internal/runtimecred/service_test.go` — mutation abort and trusted secret tests.
- `db/migrations/0006_customer_subscription_tokens.up.sql` / `.down.sql` — hash-only revocable subscription tokens.
- `tests/db/customer_subscription_migration_test.sh` — PostgreSQL migration/RLS/hash-only contract.
- `internal/subscription/service.go` — token lookup, active service validation, Naive text rendering.
- `internal/subscription/service_test.go` — token and rendering tests.
- `internal/httpapi/customers.go` — Owner customer APIs.
- `internal/httpapi/customers_test.go` — auth/CSRF/idempotency/secret-leak tests.
- `internal/httpapi/subscription.go` — public opaque-token subscription endpoint.
- `internal/httpapi/subscription_test.go` — public subscription behavior and no-activation tests.
- `internal/httpapi/routes.go`, `internal/httpapi/server.go`, `cmd/pvnaive/main.go` — wiring.
- `internal/runtimeevent/firstuse.go` — trusted local first-success event DTO + idempotent handler boundary.
- `internal/runtimeevent/firstuse_test.go` — rejects failed/non-CONNECT/untrusted events and deduplicates successful events.
- `web/src/customers.ts` — typed API client and request normalization.
- `web/src/customers.test.ts` — request-shape tests, including stale hidden-value prevention.
- `web/src/Customers.tsx` — Sanaei-style create/list/delivery UI.
- `web/src/customers.css` — compact responsive table/form/modal styles.
- `web/src/LocalQr.tsx` — local QR component.
- `web/src/App.tsx`, `web/src/routes.ts`, `web/src/routes.test.ts` — navigation.
- `web/package.json`, `web/package-lock.json` — pinned local QR dependency.

---

### Task 1: Lock the validity and quota contract

**Files:**
- Create: `internal/customer/validity.go`
- Create: `internal/customer/validity_test.go`
- Modify: `internal/customer/policy.go`
- Modify: `internal/customer/policy_test.go`

**Interfaces:**
- Produces:
  - `const BytesPerCustomerGB int64 = 1073741824`
  - `type ValidityMode string`
  - `type ValidityInput struct { Mode ValidityMode; DurationDays int; ExpiresAt *time.Time }`
  - `func QuotaGBToBytes(*int64) (*int64, error)`
  - `func NormalizeValidity(ValidityInput, time.Time) (TermTiming, time.Duration, error)`
  - `func ActivateOnFirstUse(TermTiming, time.Duration, time.Time) (TermTiming, bool, error)`

- [ ] **Step 1: Write RED pure tests**

Add tests proving:

```go
func TestQuotaGBToBytesUsesBinaryGB(t *testing.T) {
    gb := int64(50)
    got, err := QuotaGBToBytes(&gb)
    if err != nil || got == nil || *got != 50*1073741824 { t.Fatal(got, err) }
}

func TestQuotaGBToBytesNilMeansUnlimited(t *testing.T) {
    got, err := QuotaGBToBytes(nil)
    if err != nil || got != nil { t.Fatal(got, err) }
}

func TestNormalizeValidityFirstUseStartsPending(t *testing.T) {}
func TestNormalizeValidityCreationStartsNow(t *testing.T) {}
func TestNormalizeValidityFixedExpiryPreservesRFC3339Instant(t *testing.T) {}
func TestNormalizeValidityRejectsPastFixedExpiry(t *testing.T) {}
func TestNormalizeValidityRejectsZeroDays(t *testing.T) {}
func TestActivateOnFirstUseIsIdempotent(t *testing.T) {}
```

- [ ] **Step 2: Run RED**

Run:

```bash
go test ./internal/customer -run 'Test(Quota|NormalizeValidity|ActivateOnFirstUse)' -count=1
```

Expected: FAIL because the new contract does not exist.

- [ ] **Step 3: Implement minimal validity code**

Use exact request modes:

```go
const (
    ValidityOnCreation ValidityMode = "on_creation"
    ValidityOnFirstSuccessfulConnection ValidityMode = "on_first_successful_connection"
    ValidityFixedExpiry ValidityMode = "fixed_expiry"
)
```

For fixed expiry, persist `starts_at=now`, `expires_at=<exact supplied instant>`, and derive a positive `duration_seconds` only as a compatibility snapshot; expiry remains authoritative.

For first-use, persist `starts_at=NULL`, `first_connected_at=NULL`, `expires_at=NULL`, `state=pending` and retain the requested duration for later activation.

- [ ] **Step 4: Run GREEN**

```bash
go test ./internal/customer -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/customer
git commit -m "feat(customer): define quota and validity contract"
```

---

### Task 2: Complete customer persistence and safe Runtime saga

**Files:**
- Create: `internal/customer/store.go`
- Create: `internal/customer/store_test.go`
- Create: `internal/customer/service.go`
- Create: `internal/customer/service_test.go`
- Modify: `internal/runtimecred/service.go`
- Modify: `internal/runtimecred/service_test.go`

**Interfaces:**
- Customer store methods:
  - `DirectTenantID(ctx, tx) (string, error)`
  - `CreateUserTx(ctx, tx, tenantID, username string) (User, error)`
  - `CreateServiceTermTx(ctx, tx, input CreateServiceTermRecord) (ServiceTerm, error)`
  - `BindRuntimeCredentialTx(ctx, tx, tenantID, userID, termID, runtimeCredentialID string) error`
  - `ListCustomers(ctx, tx, tenantID string) ([]CustomerView, error)`
  - `ActivateFirstUseTx(ctx, tx, runtimeCredentialID string, observedAt time.Time) (ServiceTerm, bool, error)`
- Runtime mutation:
  - `func (m *Mutation) AbortAndRollback(ctx context.Context, tx *sql.Tx) error`
- Customer service:
  - `func (s *Service) CreateCustomer(ctx context.Context, tx *sql.Tx, actorID, idempotencyKey string, input CreateCustomerInput) (*CreateCustomerMutation, error)`

- [ ] **Step 1: Write RED saga tests**

Cover these exact behaviors:

```go
func TestCreateCustomerCreatesBusinessRowsAndRuntimeCredentialInOneSaga(t *testing.T) {}
func TestCreateCustomerBindsStableRuntimeCredentialUUID(t *testing.T) {}
func TestCreateCustomerRollsRuntimeBackWhenBusinessBindingFails(t *testing.T) {}
func TestCreateCustomerFirstUseTermRemainsPending(t *testing.T) {}
func TestCreateCustomerFixedExpiryKeepsExactExpiry(t *testing.T) {}
func TestCreateCustomerNeverReturnsStoredCiphertext(t *testing.T) {}
func TestRuntimeMutationAbortRollsBackAppliedRuntime(t *testing.T) {}
```

- [ ] **Step 2: Run RED**

```bash
go test ./internal/runtimecred ./internal/customer -count=1
```

- [ ] **Step 3: Implement `AbortAndRollback`**

Required semantics:

```text
if runtime has been applied but DB transaction is not committed:
  rollback Runtime Agent backup
  rollback SQL transaction
  clear generated password
  mark mutation terminal
```

If Runtime rollback fails, return `runtimecred.ErrReconciliationRequired`.

- [ ] **Step 4: Implement store/service**

Creation sequence in a single SQL transaction:

```text
resolve direct tenant
-> create business user
-> normalize quota/validity
-> create service term
-> runtimecred.Service.Create(...)
-> bind returned stable Runtime credential UUID to term
-> commit using runtime Mutation.CommitAndFinalize
```

If any step after Runtime apply fails, call `AbortAndRollback` before returning.

- [ ] **Step 5: Run GREEN**

```bash
go test ./internal/runtimecred ./internal/customer -count=1
go test ./...
```

- [ ] **Step 6: Commit**

```bash
git add internal/customer internal/runtimecred
git commit -m "feat(customer): create services through atomic runtime saga"
```

---

### Task 3: Add hash-only subscription tokens and renderer

**Files:**
- Create: `db/migrations/0006_customer_subscription_tokens.up.sql`
- Create: `db/migrations/0006_customer_subscription_tokens.down.sql`
- Modify: `db/migrations/SHA256SUMS`
- Create: `tests/db/customer_subscription_migration_test.sh`
- Create: `internal/subscription/service.go`
- Create: `internal/subscription/service_test.go`
- Modify: `internal/runtimecred/service.go`
- Modify: `internal/runtimecred/service_test.go`

**Interfaces:**
- DB table stores only SHA-256 token digest:

```sql
CREATE TABLE pvnaive.subscription_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  user_id uuid NOT NULL,
  service_term_id uuid NOT NULL,
  token_hash bytea NOT NULL CHECK (octet_length(token_hash)=32),
  status text NOT NULL CHECK (status IN ('active','revoked')),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  revoked_at timestamptz,
  UNIQUE (token_hash)
);
```

- `subscription.GenerateToken() (raw string, hash [32]byte, err error)`
- `subscription.Service.Resolve(ctx, rawToken, host string) (text string, error)`
- trusted Runtime helper returns a single active credential as `DesiredCredential`, not a JSON-serializable secret DTO.

- [ ] **Step 1: Write RED DB and service tests**

Assert:

```text
raw token is never stored
hash lookup resolves active term
revoked token returns not-found/gone behavior
expired/revoked/suspended service does not render config
renderer output is naive+https://<user>:<password>@<host>
ordinary customer JSON never includes token_hash/password/ciphertext
```

- [ ] **Step 2: Run RED**

```bash
bash tests/db/customer_subscription_migration_test.sh
go test ./internal/subscription ./internal/runtimecred -count=1
```

- [ ] **Step 3: Implement migration + token generation + renderer**

Use `crypto/rand` for 32 raw bytes, `base64.RawURLEncoding` for delivery, SHA-256 for persistence. Never log the raw token.

- [ ] **Step 4: Run GREEN**

```bash
bash tests/db/customer_subscription_migration_test.sh
go test ./internal/subscription ./internal/runtimecred -count=1
```

- [ ] **Step 5: Commit**

```bash
git add db/migrations tests/db internal/subscription internal/runtimecred
git commit -m "feat(subscription): add revocable hash-only customer tokens"
```

---

### Task 4: Expose Owner customer APIs and public subscription endpoint

**Files:**
- Create: `internal/httpapi/customers.go`
- Create: `internal/httpapi/customers_test.go`
- Create: `internal/httpapi/subscription.go`
- Create: `internal/httpapi/subscription_test.go`
- Modify: `internal/httpapi/routes.go`
- Modify: `internal/httpapi/server.go`
- Modify: `cmd/pvnaive/main.go`

**Interfaces:**
- `GET /api/v1/customers`
- `POST /api/v1/customers`
- `POST /api/v1/customers/{id}/subscription/rotate`
- `GET /sub/{opaque-token}`

Create request:

```json
{
  "username": "customer1",
  "password": "",
  "generate_password": true,
  "quota_gb": 50,
  "validity": {
    "mode": "on_first_successful_connection",
    "duration_days": 30
  }
}
```

Fixed-expiry request:

```json
{
  "username": "customer2",
  "generate_password": true,
  "quota_gb": null,
  "validity": {
    "mode": "fixed_expiry",
    "expires_at": "2026-10-15T23:59:00+03:30"
  }
}
```

Create response may contain `generated_password` and raw subscription URL exactly once. List/detail responses never do.

- [ ] **Step 1: Write RED endpoint tests**

Require Owner auth, CSRF and `Idempotency-Key` for mutations. Assert 401/403/400/409 behavior follows existing runtime routes. Assert `/sub/...` requires no panel session but invalid/revoked token does not disclose whether a username exists.

- [ ] **Step 2: Run RED**

```bash
go test ./internal/httpapi -run 'Test(Customer|Subscription)' -count=1
```

- [ ] **Step 3: Implement handlers and wiring**

Set public subscription responses to:

```text
Content-Type: text/plain; charset=utf-8
Cache-Control: no-store
X-Content-Type-Options: nosniff
```

A subscription GET must never call first-use activation.

- [ ] **Step 4: Run GREEN**

```bash
go test ./internal/httpapi -count=1
go test ./...
```

- [ ] **Step 5: Commit**

```bash
git add internal/httpapi cmd/pvnaive
git commit -m "feat(api): expose customer creation and subscription delivery"
```

---

### Task 5: Build the Sanaei-style customer UI and local QR

**Files:**
- Create: `web/src/customers.ts`
- Create: `web/src/customers.test.ts`
- Create: `web/src/Customers.tsx`
- Create: `web/src/customers.css`
- Create: `web/src/LocalQr.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/routes.ts`
- Modify: `web/src/routes.test.ts`
- Modify: `web/package.json`
- Modify: `web/package-lock.json`

**Interfaces:**
- Form state uses one discriminated validity object; hidden stale fields are not submitted.
- Default form:

```ts
{
  generate_password: true,
  quota_gb: 50,
  unlimited: false,
  validity_mode: 'on_first_successful_connection',
  duration_days: 30
}
```

- QR is generated locally from the subscription URL using a pinned package, with no HTTP call.

- [ ] **Step 1: Add RED frontend contract tests**

Tests must prove:

```text
50 GB submits numeric 50
unlimited submits quota_gb:null
switching duration -> fixed expiry removes duration_days
switching fixed expiry -> first-use removes expires_at
manual date submits timezone-aware ISO string
first-use is default
usage columns display capability unavailable, not 0 B
QR component has no network call
```

- [ ] **Step 2: Run RED**

```bash
cd web
npm test -- --run
```

- [ ] **Step 3: Add pinned QR package and implement UI**

UI order:

```text
نام کاربری
رمز عبور [تولید خودکار]
حجم [number GB] [نامحدود]
نوع اعتبار [از ساخت | از اولین اتصال | تاریخ دستی]
مدت روز / تاریخ‌وساعت
ساخت و اعمال اکانت
```

The customer table shows quota, expiry, remaining time, start mode, subscription/QR and actions. Used/remaining traffic shows `در دسترس نیست — حسابداری دقیق هنوز تأیید نشده` until capability becomes true.

Delivery modal shows username, one-time generated password, Naive URI, subscription URL, quota, validity, copy buttons and local QR.

- [ ] **Step 4: Run GREEN and build**

```bash
cd web
npm test -- --run
npm run build
```

- [ ] **Step 5: Commit**

```bash
git add web
git commit -m "feat(web): add Sanaei-style customer service workflow"
```

---

### Task 6: Implement idempotent first-successful-connection activation boundary

**Files:**
- Create: `internal/runtimeevent/firstuse.go`
- Create: `internal/runtimeevent/firstuse_test.go`
- Modify: `internal/customer/store.go`
- Modify: `internal/customer/service.go`
- Modify: `internal/customer/service_test.go`
- Modify: `cmd/pvnaive/main.go`

**Interfaces:**
- Trusted local event:

```go
type FirstUseEvent struct {
    RuntimeCredentialID string    `json:"runtime_credential_id"`
    ObservedAt          time.Time `json:"observed_at"`
    EventID             string    `json:"event_id"`
    Authenticated       bool      `json:"authenticated"`
    Method              string    `json:"method"`
}
```

- Only `Authenticated=true` and `Method=CONNECT` may activate.
- Store update is compare-and-set:

```sql
UPDATE pvnaive.service_terms st
SET first_connected_at=$observed,
    starts_at=$observed,
    expires_at=$observed + make_interval(secs => st.duration_seconds),
    state='active',
    revision=revision+1,
    updated_at=clock_timestamp()
FROM pvnaive.user_runtime_credentials urc
WHERE urc.runtime_credential_id=$runtime_id
  AND urc.service_term_id=st.id
  AND urc.unbound_at IS NULL
  AND st.start_policy='on_first_successful_connection'
  AND st.first_connected_at IS NULL;
```

- [ ] **Step 1: Write RED tests**

Prove concurrent/replayed event IDs cannot move the start time twice, failed auth does nothing, non-CONNECT does nothing, subscription GET does nothing.

- [ ] **Step 2: Run RED**

```bash
go test ./internal/runtimeevent ./internal/customer ./internal/httpapi -count=1
```

- [ ] **Step 3: Implement local-only event handler boundary**

The handler is not exposed through public HTTP routes. It is wired only to a Unix socket owned by the PVNaive runtime path, with strict filesystem permissions. Event processing failure must not affect the data-plane connection; it records degraded activation state for reconciliation instead.

- [ ] **Step 4: Run GREEN**

```bash
go test ./internal/runtimeevent ./internal/customer ./... -count=1
```

- [ ] **Step 5: Commit**

```bash
git add internal/runtimeevent internal/customer cmd/pvnaive
git commit -m "feat(runtime): add idempotent first-use activation boundary"
```

---

### Task 7: Final regression, security scan and documentation

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `docs/API_FA.md`
- Modify: `docs/DECISIONS_FA.md`
- Modify: `AGENT_HANDOFF.md`
- Modify: `PROJECT_STATUS.md`

**Interfaces:** none; this is the release gate.

- [ ] **Step 1: Add all new test commands to CI**

CI must run customer DB migration tests, Go customer/subscription/runtimeevent packages, full Go suite, frontend tests/build and existing S04R regression tests.

- [ ] **Step 2: Run complete local verification**

```bash
go test ./... -count=1
bash tests/db/customer_lifecycle_migration_test.sh
bash tests/db/customer_idempotency_migration_test.sh
bash tests/db/customer_subscription_migration_test.sh
bash tests/db/migration_test.sh
cd web && npm test -- --run && npm run build
```

- [ ] **Step 3: Secret-pattern scan**

Inspect the diff and tests for any accidental raw Runtime password, raw subscription token, `Proxy-Authorization`, ciphertext or nonce in logs/audit/ordinary JSON.

- [ ] **Step 4: Update docs with exact capability state**

Document that quota configuration and expiry modes are implemented. Do not mark exact usage/quota enforcement complete until PVN-045..049 pass. If Runtime first-use producer instrumentation is not yet proven on the pinned Caddy/forwardproxy build, mark the UI/API mode present but production activation gate blocked rather than pretending it is live.

- [ ] **Step 5: Commit**

```bash
git add .github docs AGENT_HANDOFF.md PROJECT_STATUS.md
git commit -m "docs: record customer service delivery capability and gates"
```

---

## Self-review

- Spec coverage: numeric quota, unlimited, all three validity modes, one-time password delivery, Naive link, subscription, local QR, first-use idempotency, no fake usage, and secret hygiene all map to explicit tasks.
- No hidden validity field can be authoritative alongside another mode.
- The Runtime mutation rollback gap is closed before composing business rows with Runtime state.
- Subscription fetch is explicitly separated from first-use activation.
- Exact usage/quota enforcement remains gated rather than fabricated.
- Production first-use is not claimed unless the pinned Caddy/forwardproxy producer emits the trusted local event and passes integration evidence.
