# Operator Session Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a trustworthy operator-facing active-session list with peer IP and exact per-session bytes.

**Architecture:** Extend the existing direct accounting session ledger with a separate trusted peer table and one CONNECT-time Unix-socket peer registration call. Read active sessions through a tenant-scoped database function and expose them through the existing customer product API/UI.

**Tech Stack:** Go 1.25, PostgreSQL 18, Caddy forwardproxy overlay, React/TypeScript/Vitest.

**Spec:** `docs/superpowers/specs/2026-08-31-operator-session-management-design.md`

## Global Constraints

- Source IP comes only from Caddy `RemoteAddr`; never trust forwarding headers.
- Exact byte/quota accounting code paths are not weakened or made asynchronous.
- Peer registration is fail-closed and happens once per CONNECT before payload forwarding.
- No kill/concurrency/IP-limit/history capability is claimed in Task12.
- Production deployment requires exact-head CI, encrypted backup, schema17 migration, R1 deploy and independent postflight.

---

### Task 1: Schema17 trusted peer projection
**Files:** create `db/migrations/0017_operator_session_peers.up.sql`, `db/migrations/0017_operator_session_peers.down.sql`, `tests/db/operator_session_peers_migration_test.sh`; modify `db/migrations/SHA256SUMS` and latest-schema gates.
- [ ] Write PostgreSQL18 RED proving table/function absence.
- [ ] Add peer table, immutable same-session IP function, tenant-scoped active-session read function and RLS/grants.
- [ ] Prove same-IP replay, changed-IP rejection, active/stale/final/incomplete filtering and rollback fail-closed after peer data.
- [ ] Update schema/health/backup/lifecycle/S04 latest-version gates to 17 and CI DB sequence.

### Task 2: Telemetry session-peer protocol
**Files:** modify `internal/telemetry/socket.go`, `internal/telemetry/store.go`; create/modify focused tests.
- [ ] Add RED handler/store tests for strict peer request validation and SQL function call.
- [ ] Add `SessionPeerRequest`, `/v1/accounting/session-peer`, store method and validation.
- [ ] Keep management-shaped/untrusted requests out of the backend.

### Task 3: Pinned forwardproxy trusted peer hook
**Files:** modify `third_party/forwardproxy/overlay/pvnaive_accounting.go.src`, overlay tests and pinned patch/rehearsal contract.
- [ ] RED: spoofed `X-Forwarded-For` does not alter peer; `RemoteAddr` is normalized.
- [ ] RED: peer registration error fails the CONNECT before payload forwarding.
- [ ] Implement one peer-registration call after open event and before stream wrapping.
- [ ] Regenerate/apply the pinned patch and run pinned-forwardproxy boundary tests.

### Task 4: Customer service and API
**Files:** create `internal/customer/sessions.go`, `internal/customer/sessions_store.go`, `internal/httpapi/customer_sessions.go`; modify route definitions/dispatch and tests.
- [ ] RED service/store/API tests for customer-scoped active list.
- [ ] Implement typed `CustomerSession` projection and signed-transaction store query.
- [ ] Add `GET /api/v1/product/users/{id}/sessions` with existing RBAC/RLS middleware.
- [ ] Prove stale/final/incomplete/missing-peer rows are absent and exact bytes/IP are returned.

### Task 5: Product UI
**Files:** modify `web/src/productApi.ts`, `web/src/ProductCustomers.tsx`, related CSS/tests.
- [ ] RED API and source/UI tests for session modal.
- [ ] Add fetch type/function and `نشست‌های فعال` row action/modal.
- [ ] Render IP/node/connected/last activity/duration/upload/download and explicit empty state.

### Task 6: Release and Production proof
**Files:** update `docs/API_FA.md`, `PROJECT_STATUS.md`, `HANDOFF.md`, `AGENT_TASKS.md`, `KNOWN_ISSUES.md` only after evidence.
- [ ] Full gofmt/vet/test, Web tests/build, full PostgreSQL18 DB gate, pinned forwardproxy and R1 build.
- [ ] Commit, exact-tree publish, PR and required CI gates; merge only green exact head.
- [ ] Build merged-main R1; take encrypted schema16 backup; migrate 16→17; deploy same-schema R1.
- [ ] Independent postflight: source/schema/services/Caddy invariants and a controlled session canary proving IP + exact session bytes.
- [ ] Record Task10/11/12 evidence and then start Task13 real session kill.
