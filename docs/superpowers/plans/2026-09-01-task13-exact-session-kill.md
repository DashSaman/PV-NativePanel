# Task13 Exact Session Kill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Disconnect exactly one live Naive CONNECT identified by runtime-credential/node/boot/session without revoking credentials, restarting/reloading Caddy, breaking Task14/15 limits, or duplicating final accounting.

**Architecture:** Keep the management API separated from Caddy by a local Unix-domain control socket. Caddy owns an in-process exact-tuple registry whose entries close only the selected live stream; the existing stream teardown remains solely responsible for final accounting. API authorization is added only after the data-plane boundary is proven.

**Tech Stack:** Go, pinned Caddy forwardproxy overlay, Unix-domain sockets, HTTP control protocol, GitHub Actions, React/Vitest.

**Spec:** `docs/superpowers/specs/2026-08-29-naive-usage-accounting-forwardproxy-addendum.md`

## Global Constraints

- Preserve Task14 concurrent-session-limit semantics.
- Preserve Task15 trusted `RemoteAddr`/`ClientIP` unique-IP semantics.
- Never revoke a credential to kill one session.
- Never restart/reload Caddy to kill one session.
- Final accounting remains exactly once through the existing close path.
- Production is not a development or PostgreSQL test lane.

---

### Task 1: Forwardproxy live registry and local control listener

**Files:**
- Create/modify: `third_party/forwardproxy/patches/0002-pvnaive-session-control.patch`
- Modify: `scripts/build/build-pinned-accounting-caddy.sh`

**Interfaces:**
- Consumes: existing `pvnaiveAccountingSession` runtime/node/boot/session identity and the existing stream teardown/final-accounting path.
- Produces: local `POST /v1/sessions/kill` endpoint returning `{found,killed}` and exact stream cancellation.

- [ ] **Step 1:** Add tests proving exact tuple kill closes one registered stream, sibling survives, forged tuple cannot match, repeat kill is idempotent, and unregister removes the tuple.
- [ ] **Step 2:** Run the pinned forwardproxy workflow and verify RED because registry/control wiring does not exist.
- [ ] **Step 3:** Implement the minimal registry and local handler/listener; register only after accounting open/peer succeeds and unregister on stream teardown.
- [ ] **Step 4:** Run forwardproxy tests with race coverage; verify GREEN and preserve existing `RemoteAddr`/ClientIP tests.
- [ ] **Step 5:** Run reproducible Caddy build and exact-accounting gates.
- [ ] **Step 6:** Commit/publish only the validated data-plane increment.

### Task 2: API ownership/RBAC/CSRF boundary

**Files:**
- Modify: `cmd/pvnaive/main.go`
- Modify: `internal/httpapi/customer_sessions.go`
- Modify: `internal/httpapi/customer_extra_routes.go`
- Modify: `internal/httpapi/server.go`
- Test: `internal/httpapi/*session*_test.go`

**Interfaces:**
- Consumes: `sessioncontrol.Client.Kill(context.Context, sessionkill.Key)`.
- Produces: authenticated DELETE session endpoint using the exact tuple loaded through the caller's existing tenant/role-scoped transaction.

- [ ] **Step 1:** Add failing route/RBAC/IDOR/CSRF tests including forged user/session combinations.
- [ ] **Step 2:** Verify RED for missing route behavior.
- [ ] **Step 3:** Implement minimal route and dependency wiring; no credential mutation.
- [ ] **Step 4:** Run focused tests, race tests, then full Go tests.
- [ ] **Step 5:** Publish only after exact-head gates are green.

### Task 3: UI and exact live rehearsal

**Files:**
- Modify: `web/src/productApi.ts`
- Modify: `web/src/productApi.test.ts`
- Modify: `web/src/ProductCustomers.tsx`
- Add/update: Task13 stage rehearsal/evidence under `tests/stages/` and `ops/evidence/`.

**Interfaces:**
- Consumes: API DELETE endpoint.
- Produces: explicit per-session kill action and release proof.

- [ ] **Step 1:** Add failing API/UI tests for exact-session kill and confirmation copy.
- [ ] **Step 2:** Implement minimal UI/API client and verify Web tests/build.
- [ ] **Step 3:** Run full Go/race/Web/pinned-forwardproxy/reproducible-Caddy gates.
- [ ] **Step 4:** Rehearse real HTTP/1.1 and HTTP/2: target dies, sibling survives, forged tuple fails, repeat kill idempotent, credential remains usable, final accounting exactly once.
- [ ] **Step 5:** Merge only the exact verified tree; before any Production mutation create a fresh encrypted backup and rollback snapshot, then run postflight gates.