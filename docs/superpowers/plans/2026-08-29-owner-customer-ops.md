# Owner Customer Operations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the Owner-observed customer-management gaps without weakening PVNaive's Runtime safety model, starting with a complete P0 customer action surface and then exact-accounting proof gates.

**Architecture:** Keep Business User/ServiceTerm state separate from Naive Runtime credentials. Read-only customer delivery actions must never mutate Runtime or commercial state; subscription rotation and password rotation remain explicit, independent mutations. Reversible subscription delivery is implemented with authenticated Owner retrieval of an encrypted-at-rest active token while public resolution continues to validate the SHA-256 digest. Runtime-changing customer lifecycle actions reuse the existing expected-revision/validate/backup/reload/rollback Runtime service rather than adding a second mutation path.

**Tech Stack:** Go, PostgreSQL 18, React/TypeScript/Vite/Vitest, Caddy/Naive Runtime Agent, GitHub Actions.

**Spec:** `OWNER_REQUIREMENTS.md`

## Global Constraints

- Product name is `PVNaive`.
- Do not force-push/reset `main` or overwrite production evidence.
- `OWNER_REQUIREMENTS.md` wins over older customer UX handoffs.
- No fake usage, remaining traffic, online state, device state, speed, or hard quota before Runtime proof.
- Preserve fixed Unix-socket Runtime Agent, expected-current-SHA, exact backup, `caddy validate`, reload-only, postflight verification, rollback, CSRF, idempotency, RLS, tenant isolation, secret redaction, and stable Runtime UUIDs.
- Viewing Details/QR/Subscription/Direct Link is read-only.
- Subscription reissue must not rotate the Runtime password.
- Password rotation must be an explicit independent action.
- Customer Delete uses safe revoke/soft-delete semantics, not ordinary physical row destruction.

---

### Task 1: Customer read-only delivery projection

**Files:**
- Modify: `db/migrations/*`
- Modify: `internal/customer/models.go`
- Modify: `internal/customer/list_subscription.go`
- Modify: `internal/customer/store.go`
- Modify: `internal/customer/service.go`
- Modify: `internal/httpapi/customer_management.go`
- Modify: `internal/httpapi/routes.go`
- Test: `internal/customer/*_test.go`
- Test: `internal/httpapi/*_test.go`
- Test: `tests/db/*`

**Interfaces:**
- Produces authenticated `GET /api/v1/customers/{id}/subscription` returning the current active `subscription_path` without token rotation or password/runtime mutation.
- Public resolver remains hash-based.

- [ ] Write failing service/store/API tests proving read-only retrieval returns the same active subscription token and does not rotate it.
- [ ] Add encrypted-at-rest token recovery fields/projection using the existing 32-byte application Runtime key with explicit key id; retain token hash for public lookup.
- [ ] Add authenticated Owner retrieval service/store/handler.
- [ ] Add migration tests and rollback coverage.
- [ ] Run Go, DB and HTTP tests.

### Task 2: Customer details, Edit and View QR UX

**Files:**
- Modify: `web/src/customers.ts`
- Modify: `web/src/customers.test.ts`
- Modify: `web/src/Customers.tsx`
- Modify: `web/src/customers.css`

**Interfaces:**
- Consumes `GET /api/v1/customers/{id}/subscription`.
- Produces distinct `Details`, `Edit`, `View QR`, `Copy Subscription`, and `Reissue Subscription` actions.

- [ ] Write failing web API tests for read-only current-subscription fetch.
- [ ] Add client function/type.
- [ ] Add Details modal and a QR modal that fetches current active subscription read-only.
- [ ] Rename the current mutation action to explicit `Reissue Subscription` and add destructive warning copy.
- [ ] Keep Edit obvious and preserve Runtime/password on service-only edits.
- [ ] Run Vitest and production build.

### Task 3: Customer lifecycle API and UI

**Files:**
- Create/Modify focused files under `internal/customer/` and `internal/httpapi/`.
- Modify: `internal/httpapi/routes.go`
- Modify: `web/src/customers.ts`
- Modify: `web/src/Customers.tsx`
- Test corresponding Go/HTTP/Web files.

**Interfaces:**
- `POST /api/v1/customers/{id}/suspend`
- `POST /api/v1/customers/{id}/resume`
- `DELETE /api/v1/customers/{id}` safe revoke
- Runtime mutation uses the bound Runtime credential revision, never blind writes.

- [ ] Write failing policy/service tests for suspend, resume and safe revoke.
- [ ] Implement transactional business state + Runtime state choreography with compensation on failure.
- [ ] Revoke active subscription on revoke; suspend/resume must preserve token identity unless policy requires otherwise.
- [ ] Add explicit UI actions and confirmations.
- [ ] Run failure-path rehearsals and tests.

### Task 4: Explicit password rotation

**Files:**
- Modify focused customer/runtime HTTP integration files.
- Modify web customer client/UI.
- Add tests.

**Interfaces:**
- Customer-level password rotation delegates to existing Runtime `Rotate` path with expected revision.
- Generated password is one-time response data only.

- [ ] Write failing tests proving QR/subscription operations never invoke password rotation.
- [ ] Add customer `rotate-password` action with generated/custom password support.
- [ ] Show generated password once after successful Runtime commit.
- [ ] Ensure subscription projection syncs updated Runtime secret without rotating subscription token.
- [ ] Run Go/Web tests.

### Task 5: Volume and validity operator ergonomics

**Files:**
- Modify customer service/update API and UI.
- Add tests.

**Interfaces:**
- Set total quota.
- Add volume delta.
- Unlimited quota.
- Set fixed expiry.
- Extend by N days.
- Preserve/choose first-successful-connection mode where valid.

- [ ] Write failing tests distinguishing add-volume from set-total-volume.
- [ ] Implement overflow-safe quota delta and extension semantics.
- [ ] Expose explicit UI controls without Runtime/password mutation.
- [ ] Run tests.

### Task 6: Search/filter/sort and bulk-safe customer table

**Files:**
- Modify customer list query/API/types/UI.
- Add tests.

- [ ] Add deterministic server-side search/sort/filter contracts for currently truthful fields only.
- [ ] Add pagination.
- [ ] Add checkbox selection and bulk preview.
- [ ] Add bulk lifecycle/volume/expiry actions only where single-user semantics are already proven.
- [ ] Run tests.

### Task 7: Exact Naive accounting proof gate

**Files:**
- Follow `docs/superpowers/plans/2026-08-29-naive-usage-accounting-proof-gate.md` and accounting specs.

- [ ] Prove a per-credential observable source for upload/download bytes against the pinned Naive/Caddy data path.
- [ ] Prove restart/reload/reboot and no-double-count behavior.
- [ ] Persist monotonic checkpoints/reconciliation evidence.
- [ ] Keep `usage_capability.available=false` until all proof gates pass.

### Task 8: Hard quota and usage reset after proof

- [ ] Add exact usage projection only after Task 7 evidence is green.
- [ ] Add hard quota enforcement with recovery after volume increase.
- [ ] Add audited/idempotent usage reset.
- [ ] Add reset strategies only where Runtime semantics are exact.

### Task 9: Trusted first-successful-connection producer

- [ ] Instrument/prove the pinned Runtime producer for authenticated successful CONNECT events.
- [ ] Verify panel reads, QR reads, subscription fetches, reloads, failed auth and health checks do not activate pending terms.
- [ ] Enable first-use production claim only after end-to-end evidence.

### Task 10: Release/production closure

- [ ] Fresh full CI on exact head.
- [ ] Update `OWNER_REQUIREMENTS.md`, `AGENT_TASKS.md`, `PROJECT_STATUS.md`, `docs/S05_HANDOFF.md`, `FEATURE_MATRIX.md`, `KNOWN_ISSUES.md`, `WORKLOG.md` with evidence.
- [ ] Build/verify release bundle.
- [ ] Perform controlled production rollout separately from development proof and record evidence on `main` without force operations.
