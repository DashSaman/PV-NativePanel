# Naive Runtime Credential Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let an authenticated PVNaive Owner safely import the existing NaiveProxy credential set, add/rename/rotate/disable/revoke credentials from `/runtime/naive`, and apply the resulting Caddy `forward_proxy` credential set with validate → exact backup → reload-only → verify → rollback semantics.

**Architecture:** Keep the existing API unprivileged. PostgreSQL stores desired credential/runtime revision state; a dedicated `pvnaive-runtime-agent` exposes a narrow JSON API over `/run/pvnaive/runtime-agent.sock` and is the only component allowed to inspect/transform `/etc/caddy/Caddyfile`, run `/usr/local/bin/caddy validate`, reload `caddy-naive.service`, and restore agent-created backups. Runtime mutations are sagas: DB state is staged, the agent applies Caddy, DB commit happens before success is written to HTTP, and an agent rollback is mandatory if the DB commit fails after a live apply.

**Tech Stack:** Go 1.25, PostgreSQL 18, `database/sql` + pgx v5, AES-256-GCM, SHA-256, Unix-domain HTTP/JSON, systemd, Caddy v2.11.2 + `forward_proxy`, React/TypeScript/Vite, GitHub Actions.

**Spec:** `docs/superpowers/specs/2026-08-28-naive-runtime-credentials-design.md`

## Global Constraints

- Work on `s04-auth`; never implement this feature on `main`.
- Product name is `PVNaive`; repository name remains unchanged.
- Current production schema is v2; this extension introduces migration `0003` and does not mark S06 complete.
- Existing Caddy data-plane behavior, `probe_resistance`, camouflage root, panel route, SSH and firewall must be preserved.
- Caddy changes use `systemctl reload caddy-naive.service` only; never restart.
- API remains `127.0.0.1:8080`; runtime agent uses Unix socket only and has no TCP listener.
- Runtime secret key is `/etc/pvnaive/runtime.key`, exactly 32 raw bytes, `root:pvnaive 0640`; never reuse `auth.key` or age backup identity.
- No plaintext runtime password, encrypted secret bytes, raw Caddyfile or `probe_resistance` secret in API GET responses, logs, audit records or repository evidence.
- Existing global `pvnaive.credentials` must not be populated with fake reseller/user/subscription rows.
- Mutating Naive credentials is Owner-only for this extension.
- At least one active Naive credential must remain.
- `DELETE` is soft revoke; rows are retained for audit/rollback.
- Every production-code task follows RED → observed failure → minimal GREEN → full relevant test pass.
- No production deployment until CI is fully green and a disposable rehearsal passes.

---

### Task 1: Migration 0003 and owner-only runtime credential state

**Files:**
- Create: `db/migrations/0003_naive_runtime_credentials.up.sql`
- Create: `db/migrations/0003_naive_runtime_credentials.down.sql`
- Modify: `db/migrations/SHA256SUMS`
- Create: `tests/db/naive_runtime_migration_test.sh`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Produces table `pvnaive.naive_runtime_credentials` with columns `id`, `username`, encrypted secret fields, status/origin/audit actor fields, timestamps and `revision`.
- Extends `pvnaive.runtime_revisions` with nullable `idempotency_key text` and a scope/protocol/idempotency unique index.
- The new credential table uses FORCE RLS with an Owner-only policy based on `pvnaive.current_actor_role() = 'owner'`.
- `pvnaive_app` receives only the table privileges needed after signed request context binding; no pre-auth SECURITY DEFINER credential accessor is added.

- [ ] **Step 1: Write failing migration contract test**

Create `tests/db/naive_runtime_migration_test.sh` that provisions disposable PostgreSQL 18 roles/database, runs current migrations, and asserts schema version `3`, table/column/check/index existence, Owner-context access, Admin-context denial, last-active protection support at the application constraint layer, runtime revision idempotency uniqueness, and one-step destructive rollback returns schema to v2 while preserving v2 auth tables.

Key assertions:

```bash
[[ "${schema_version}" == "3" ]]
[[ "${contract}" == "true|true|true|true|true|true" || "${contract}" == "t|t|t|t|t|t" ]]
[[ "${admin_select_rc}" -ne 0 ]]
[[ "${remaining}" == "true|2" || "${remaining}" == "t|2" ]]
```

- [ ] **Step 2: Wire the new test into database CI and verify RED**

Add `bash tests/db/naive_runtime_migration_test.sh` after `auth_migration_test.sh` in `.github/workflows/ci.yml`.

Run through PR CI. Expected: database job FAILS because migration 0003/table does not exist; Go/Web baseline remain green.

- [ ] **Step 3: Implement migration 0003**

Use this exact data shape:

```sql
CREATE TABLE pvnaive.naive_runtime_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username text NOT NULL CHECK (length(username) BETWEEN 1 AND 64),
    secret_hash bytea NOT NULL CHECK (octet_length(secret_hash) = 32),
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 16),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) = 12),
    encryption_key_id text NOT NULL CHECK (length(encryption_key_id) BETWEEN 1 AND 160),
    status text NOT NULL CHECK (status IN ('active','disabled','revoked')),
    origin text NOT NULL CHECK (origin IN ('imported','panel')),
    created_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    updated_by_actor_id uuid REFERENCES pvnaive.actors(id) ON DELETE RESTRICT,
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    rotated_at timestamptz,
    revoked_at timestamptz
);
CREATE UNIQUE INDEX naive_runtime_credentials_username_uidx
    ON pvnaive.naive_runtime_credentials (username);
```

Add Owner-only FORCE RLS, add nullable runtime revision `idempotency_key`, and create a unique index over `COALESCE(tenant_id, zero-uuid), protocol_id, idempotency_key` where non-null. Down migration drops the index/column/table only.

- [ ] **Step 4: Update migration checksums and verify GREEN**

Run full database job. Expected: all existing migration/auth/backup tests plus `PVNAIVE_NAIVE_RUNTIME_MIGRATION_TEST=PASSED`.

- [ ] **Step 5: Commit**

```bash
git add db/migrations tests/db/naive_runtime_migration_test.sh .github/workflows/ci.yml
git commit -m "feat: add naive runtime credential schema"
```

---

### Task 2: Runtime secret envelope and conservative credential policy

**Files:**
- Create: `internal/runtimecred/secret.go`
- Create: `internal/runtimecred/secret_test.go`
- Create: `internal/runtimecred/policy.go`
- Create: `internal/runtimecred/policy_test.go`
- Create: `internal/runtimecred/types.go`

**Interfaces:**
- `EncryptSecret(key, plaintext []byte) (ciphertext, nonce []byte, err error)`
- `DecryptSecret(key, nonce, ciphertext []byte) ([]byte, error)`
- `HashSecret([]byte) [32]byte`
- `GeneratePassword() (string, error)` using 24 CSPRNG bytes, base64url without padding.
- `ValidateUsername(string) error`
- `ValidatePassword(string, imported bool) error`
- `Credential`, `DesiredCredential`, `DesiredState` domain types contain no JSON tag that exposes ciphertext/plaintext to browser DTOs.

- [ ] **Step 1: Write failing crypto and policy tests**

Cover AES-GCM roundtrip, tamper/wrong-key failure, nonce uniqueness, 32-byte key enforcement, generated password entropy shape, username injection rejection (`:`, whitespace, quotes, braces, backslash, newline), new-password minimum 14 bytes, and imported-password compatibility path that preserves existing live value while still rejecting control/newline/NUL.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/runtimecred`. Expected: FAIL because package/functions do not exist.

- [ ] **Step 3: Implement minimal crypto/policy code**

Use Go stdlib `crypto/aes`, `cipher.NewGCM`, `crypto/rand`, `crypto/sha256`, `encoding/base64`, and explicit byte/rune validation. Do not add dependencies.

- [ ] **Step 4: Verify GREEN and full Go suite**

Run `go test ./internal/runtimecred && go test ./...`.

- [ ] **Step 5: Commit**

```bash
git add internal/runtimecred
git commit -m "feat: add runtime credential secret handling"
```

---

### Task 3: Byte-preserving Caddy forward_proxy inspection and transformation

**Files:**
- Create: `internal/naiveruntime/caddyfile.go`
- Create: `internal/naiveruntime/caddyfile_test.go`
- Create: `internal/naiveruntime/testdata/live_like.Caddyfile`

**Interfaces:**
- `InspectCaddyfile(input []byte) (Inspection, error)`
- `RenderCredentials(input []byte, credentials []runtimecred.DesiredCredential) ([]byte, error)`
- `Inspection` exposes usernames and safe structural metadata; an internal import-only field may carry parsed passwords but must not have JSON tags.
- Parser supports one `forward_proxy {}` block only, tracks braces and Caddyfile quoted tokens safely enough to avoid matching nested comments/strings, and fails closed on ambiguity.

- [ ] **Step 1: Write failing parser/transform tests**

Tests must prove: zero/multiple block rejection; repeated `basic_auth` parsing; preservation of `hide_ip`, `hide_via`, `probe_resistance`, panel/API route and camouflage bytes; only the credential directive span changes; duplicate usernames rejected; rendered output is deterministic; unsafe username/password values cannot inject new Caddy directives.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/naiveruntime`. Expected: FAIL because parser/renderer is missing.

- [ ] **Step 3: Implement parser and renderer**

Implement a small lexical scanner for comments, quoted strings, escapes and brace depth. Do not build a general Caddy parser. Locate the one `forward_proxy` block and replace only contiguous `basic_auth` directive lines, preserving all unrelated bytes exactly.

- [ ] **Step 4: Verify GREEN**

Run `go test ./internal/naiveruntime && go test ./...`.

- [ ] **Step 5: Commit**

```bash
git add internal/naiveruntime
git commit -m "feat: add safe naive caddy transformer"
```

---

### Task 4: Unix-socket Runtime Agent protocol and fixed capability boundary

**Files:**
- Create: `internal/runtimeagent/protocol.go`
- Create: `internal/runtimeagent/server.go`
- Create: `internal/runtimeagent/client.go`
- Create: `internal/runtimeagent/server_test.go`
- Create: `internal/runtimeagent/client_test.go`
- Create: `cmd/pvnaive-runtime-agent/main.go`
- Create: `ops/systemd/pvnaive-runtime-agent.service`

**Interfaces:**
- Unix socket: `/run/pvnaive/runtime-agent.sock`.
- JSON operations: `/v1/health`, `/v1/inspect`, `/v1/validate`, `/v1/apply`, `/v1/rollback`.
- `Client.Inspect(ctx)`, `Validate(ctx, request)`, `Apply(ctx, request)`, `Rollback(ctx, request)`.
- Requests carry only typed credential sets, expected Caddy SHA and agent-issued backup IDs; no arbitrary command/path/service fields exist.

- [ ] **Step 1: Write failing protocol boundary tests**

Use a temp Unix socket and fake fixed operator dependency. Prove there is no TCP listener, malformed/oversized JSON is rejected, unknown fields are rejected, arbitrary path/service fields cannot be supplied, client uses Unix dialer, and response DTOs redact import secrets except on the explicit internal inspect response type.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/runtimeagent`. Expected: FAIL because package is missing.

- [ ] **Step 3: Implement protocol/server/client**

Use `net.Listen("unix", socket)`, `http.Server`, `http.Transport.DialContext`, strict JSON decoding and body limits. Main binary creates `/run/pvnaive`, removes stale socket only after verifying it is a socket, listens, then `chown root:pvnaive`/`chmod 0660`.

- [ ] **Step 4: Add hardened systemd unit**

Unit runs as root but narrows filesystem and address-family access; only AF_UNIX is required. It receives fixed paths via compiled constants rather than caller-controlled environment.

- [ ] **Step 5: Verify GREEN**

Run `go test ./internal/runtimeagent && go test ./...` and `systemd-analyze verify ops/systemd/pvnaive-runtime-agent.service` in an Ubuntu/systemd-capable test stage when available.

- [ ] **Step 6: Commit**

```bash
git add internal/runtimeagent cmd/pvnaive-runtime-agent ops/systemd/pvnaive-runtime-agent.service
git commit -m "feat: add narrow runtime agent socket API"
```

---

### Task 5: Privileged Caddy validate/apply/rollback implementation

**Files:**
- Create: `internal/runtimeagent/operator.go`
- Create: `internal/runtimeagent/operator_test.go`
- Create: `tests/stages/S04R_runtime_agent_rehearsal.sh`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- `Operator.Inspect`, `Validate`, `Apply`, `Rollback`, `Health` implement fixed live paths.
- Apply response returns safe `PreviousSHA256`, `AppliedSHA256`, `BackupID`, `MainPID`, `NRestarts`; never returns raw Caddyfile.
- Rollback accepts only opaque `BackupID` generated under the fixed backup root.

- [ ] **Step 1: Write failing operator tests**

Use temporary fake Caddyfile and injected command runner. Assert expected-current-SHA mismatch fails before write, exact backup is made, candidate validation precedes install, reload command is exactly `systemctl reload caddy-naive.service`, restart is never used, PID/restart changes fail the apply, failed postflight restores old bytes, and backup traversal IDs are rejected.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/runtimeagent -run Operator`. Expected: FAIL because operator is missing.

- [ ] **Step 3: Implement fixed operator**

Use atomic same-filesystem temp write + fsync + rename, file metadata preservation, SHA-256, fixed command argv without shell, bounded command contexts and exact backup directory creation with collision-safe names.

- [ ] **Step 4: Add disposable shell rehearsal**

Rehearsal uses fake `caddy` and `systemctl` executables in a controlled test harness to prove validate/apply/rollback ordering and that `restart` is rejected/not invoked. It must not require real production Caddy secrets.

- [ ] **Step 5: Verify GREEN and wire CI**

Run Go suite plus `bash tests/stages/S04R_runtime_agent_rehearsal.sh` in CI.

- [ ] **Step 6: Commit**

```bash
git add internal/runtimeagent tests/stages/S04R_runtime_agent_rehearsal.sh .github/workflows/ci.yml
git commit -m "feat: add atomic caddy runtime apply rollback"
```

---

### Task 6: PostgreSQL runtime credential store and revision saga

**Files:**
- Create: `internal/runtimecred/store.go`
- Create: `internal/runtimecred/store_test.go`
- Create: `internal/runtimecred/service.go`
- Create: `internal/runtimecred/service_test.go`

**Interfaces:**
- `Store.ListTx(ctx, tx)`, `CreateStagedTx`, `UpdateStagedTx`, `RotateStagedTx`, `RevokeStagedTx`, `CreateRuntimeRevisionTx`, `MarkRevisionAppliedTx`, `MarkRevisionFailedTx`.
- `Service` depends on `*sql.DB`, runtime key and `runtimeagent.Client` interface.
- Mutations receive authenticated `*sql.Tx`, actor ID, idempotency key and expected record/runtime revision.
- Service decrypts active secrets only to construct agent desired state and zeroes temporary plaintext buffers as soon as practical.

- [ ] **Step 1: Write failing store/service tests**

Cover Owner-context list/create; duplicate username; optimistic record revision conflict; idempotency replay returns the prior logical result; last-active disable/revoke rejected; soft revoke retained; generated password returned once; API-safe DTO never contains ciphertext; agent apply failure leaves DB mutation unapplied; agent success + DB commit failure invokes rollback.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/runtimecred`. Expected: FAIL for missing store/service behavior.

- [ ] **Step 3: Implement store and service**

Reuse existing `pvnaive.runtime_revisions`: store encrypted canonical desired-state JSON as `nonce || ciphertext` in `config_ciphertext`, SHA-256 of canonical plaintext in `config_checksum_sha256`, safe usernames/statuses plus backup/SHA metadata in `manifest`, and `idempotency_key` in the new column.

- [ ] **Step 4: Implement commit-compensation contract**

For mutation calls used by HTTP, service exposes a result with `CommitAndFinalize(ctx)` or equivalent: after successful agent apply + DB applied-state update, commit the bound transaction before any success response. If commit fails, call agent rollback using the returned backup ID and return a stable consistency error. Never report success before commit.

- [ ] **Step 5: Verify GREEN**

Run package tests and full Go suite.

- [ ] **Step 6: Commit**

```bash
git add internal/runtimecred
git commit -m "feat: add naive runtime credential saga"
```

---

### Task 7: Owner-only Naive Runtime HTTP API

**Files:**
- Modify: `internal/httpapi/routes.go`
- Modify: `internal/httpapi/server.go`
- Create: `internal/httpapi/runtime_naive.go`
- Create: `internal/httpapi/runtime_naive_test.go`
- Modify: `cmd/pvnaive/main.go`
- Modify: `ops/systemd/pvnaive-api.service`

**Interfaces:**
- `ServerConfig` gains `RuntimeService *runtimecred.Service`.
- Routes:
  - `GET /api/v1/runtime/naive`
  - `GET /api/v1/runtime/naive/credentials`
  - `GET /api/v1/runtime/naive/revisions`
  - `POST /api/v1/runtime/naive/credentials`
  - `PATCH /api/v1/runtime/naive/credentials/{id}`
  - `POST /api/v1/runtime/naive/credentials/{id}/rotate-password`
  - `DELETE /api/v1/runtime/naive/credentials/{id}`
  - `POST /api/v1/runtime/naive/revisions/{id}/rollback`
- Mutations are `Owner`; safe reads may be `Admin` where explicitly tested, but credentials listing defaults to Owner for this extension.

- [ ] **Step 1: Write failing route/handler tests**

Assert anonymous 401, Admin mutation 403, Owner mutation requires CSRF, unknown JSON fields 400, missing/short idempotency key 400, stale revision 409, generated-password response is `Cache-Control: no-store` and one-time only, DELETE maps to soft revoke, and no GET response contains secret/ciphertext fields.

- [ ] **Step 2: Verify RED**

Run `go test ./internal/httpapi`. Expected: FAIL because routes/service wiring are absent.

- [ ] **Step 3: Implement API handlers and path parsing**

Use strict JSON decoder, `Idempotency-Key` header, expected revision from `If-Match` or explicit JSON field with one canonical contract, existing CSRF middleware and authenticated transaction. Mutation handlers call service commit/finalization before writing success.

- [ ] **Step 4: Wire runtime key and Unix client in API main**

Read `/etc/pvnaive/runtime.key`, require 32 bytes, create Unix runtime-agent client for `/run/pvnaive/runtime-agent.sock`, construct store/service, and keep readiness semantics explicit: Auth readiness remains required; runtime status endpoint reports agent degradation rather than taking the whole auth API down.

- [ ] **Step 5: Update API unit access**

Grant read-only access to `/etc/pvnaive/runtime.key` and Unix socket path without giving API write access to Caddy or `/var/backups/pvnaive/caddy`.

- [ ] **Step 6: Verify GREEN**

Run `go test ./internal/httpapi ./cmd/pvnaive/... ./...`.

- [ ] **Step 7: Commit**

```bash
git add internal/httpapi cmd/pvnaive/main.go ops/systemd/pvnaive-api.service
git commit -m "feat: expose owner naive runtime API"
```

---

### Task 8: Initial live credential import bootstrap

**Files:**
- Create: `cmd/pvnaive-runtime-import/main.go`
- Create: `scripts/runtime/import-live-naive.sh`
- Create: `tests/runtime/import_live_naive_test.sh`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Root-only command inspects current Caddy through the local agent, encrypts each current credential with `/etc/pvnaive/runtime.key`, inserts `origin='imported'` rows and an initial applied `runtime_revisions` record, and refuses when DB state is already owned/imported.
- It never prints password, encrypted bytes, probe-resistance secret or raw Caddyfile.

- [ ] **Step 1: Write failing import tests**

Use fake agent and disposable DB. Assert first import succeeds, second import refuses, ambiguous/no credentials refuses, stdout/stderr contain usernames/count/SHA only, encrypted DB secret decrypts to the fixture value with the runtime key, and no fixture secret appears in logs/evidence output.

- [ ] **Step 2: Verify RED**

Run Go/shell test. Expected: FAIL because import command is absent.

- [ ] **Step 3: Implement root-only import**

Require EUID 0, exact host only in the production wrapper, use local DB admin boundary only as needed for bootstrap, and immediately prove rendered imported state validates without changing live Caddy.

- [ ] **Step 4: Verify GREEN and CI**

Run shell syntax, import test and full Go/database CI.

- [ ] **Step 5: Commit**

```bash
git add cmd/pvnaive-runtime-import scripts/runtime tests/runtime .github/workflows/ci.yml
git commit -m "feat: add guarded live naive credential import"
```

---

### Task 9: `/runtime/naive` panel UI

**Files:**
- Create: `web/src/runtimeNaive.ts`
- Create: `web/src/runtimeNaive.test.ts`
- Modify: `web/src/routes.ts`
- Modify: `web/src/routes.test.ts`
- Modify: `web/src/App.tsx`
- Modify: `web/src/styles.css`

**Interfaces:**
- Client functions: `getNaiveRuntime`, `listNaiveCredentials`, `createNaiveCredential`, `patchNaiveCredential`, `rotateNaivePassword`, `revokeNaiveCredential`, `rollbackNaiveRevision`.
- All unsafe calls read the existing CSRF cookie and send `X-CSRF-Token` plus `Idempotency-Key`.
- UI never requests existing password.

- [ ] **Step 1: Write failing web client/route tests**

Assert correct API paths/methods/headers, generated secret parsed only from create/rotate successful response, `runtime/naive` navigation is Owner-visible, and no password field is present in list DTOs.

- [ ] **Step 2: Verify RED**

Run `cd web && npm test`. Expected: FAIL for missing runtime client/UI route.

- [ ] **Step 3: Implement client and responsive UI**

Add Runtime summary and credential table with Add, Rename, Rotate, Generate, Enable/Disable, Revoke and Rollback actions. Generated password appears once with copy affordance. Disable/revoke last active returns explanatory backend error. Keep unavailable quota/session/accounting fields visibly unavailable rather than fake values.

- [ ] **Step 4: Verify GREEN and build**

Run `cd web && npm test && npm run build`.

- [ ] **Step 5: Commit**

```bash
git add web/src
git commit -m "feat: add naive runtime credential panel"
```

---

### Task 10: S04R production bundle, rehearsal and rollout gates

**Files:**
- Create: `scripts/stages/S04R-naive-credentials.sh`
- Create: `tests/stages/S04R_naive_credentials_rehearsal.sh`
- Modify: `scripts/release/build-s04-bundle.sh` or create a new `build-s04r-bundle.sh` if changing the frozen S04 artifact would make rollback ambiguous.
- Modify: `.github/workflows/ci.yml`
- Update after live evidence: `CONTINUE_HERE.md`, `ops/S04_LIVE_STATE.md`, `ops/DEPLOYMENT_PROGRESS.md`, `AGENT_HANDOFF.md`
- Create live evidence files under `ops/evidence/` only from observed output.

**Interfaces:**
- Bundle contains API, password helper, runtime agent, runtime import helper, web build, migrations 0001–0003, units and stage/rehearsal scripts with SHA256 manifest.
- Production stage is recovery-safe and does not modify Caddy credential content until import equivalence + exact live custom Caddy multi-credential validation pass.

- [ ] **Step 1: Write failing end-to-end rehearsal**

Disposable PostgreSQL 18 + fake Caddy/systemd harness: migrate to v3, start runtime agent on Unix socket, bootstrap owner, import one live-like credential, start API, authenticate Owner, list credential, add second generated credential, apply, verify candidate contains both credentials, revoke one, verify soft revoke and active-set render, force apply failure and prove exact rollback, and prove DB-commit failure invokes agent rollback.

- [ ] **Step 2: Verify RED**

Wire rehearsal after Go/Web/DB jobs; expected failure until stage/bundle wiring exists.

- [ ] **Step 3: Implement release/stage scripts**

Preflight checks current live Caddy SHA, exact host, schema v2, S04/API health, existing panel exposure and runtime key/agent paths. Stage backs up DB/config, migrates to v3, installs agent/key/API/web, performs guarded import/equivalence validation, and leaves live credentials unchanged after initial install.

- [ ] **Step 4: Verify exact custom Caddy multiple credential syntax before first mutation**

On `testAmir5-3`, use a no-install candidate copied from the exact live Caddyfile with an additional dummy `basic_auth` directive and run `/usr/local/bin/caddy validate --config <candidate> --adapter caddyfile`. Do not install candidate. Record only PASS/FAIL and safe structural evidence; do not record credentials or probe secret.

Upstream current `caddyserver/forwardproxy` supports repeated `basic_auth`, but production ownership still depends on this exact local-binary validation.

- [ ] **Step 5: Full CI verification**

Require fresh success for Go formatting/vet/tests, Web tests/build, PostgreSQL 18 database gates, S04 auth rehearsal, S04R runtime rehearsal and bundle checksum.

- [ ] **Step 6: Production deploy**

Deploy only the verified artifact by SHA. Run one server command at a time. Require runtime agent Unix socket only, schema v3, import equivalence, API health, panel health, Caddy active, unchanged PID/NRestarts, ports 22/80/443, API 127.0.0.1:8080 only, and camouflage root unchanged.

- [ ] **Step 7: Visual Owner smoke**

Owner opens `/panel/`, enters `/runtime/naive`, sees imported username (never password), creates a temporary second credential, verifies it works in a Naive-compatible client, then revokes the temporary credential and verifies the original remains working. Never paste runtime passwords into chat.

- [ ] **Step 8: Independent postflight and documentation**

Only after independent postflight, record `S04R-NAIVE-CREDENTIALS=PASSED`. Do not mark `S06-RUNTIME=PASSED`; S05 remains the next full product stage unless the official ledger is deliberately reordered with a separate ADR.

- [ ] **Step 9: Commit evidence/docs**

```bash
git add ops/evidence CONTINUE_HERE.md ops/S04_LIVE_STATE.md ops/DEPLOYMENT_PROGRESS.md AGENT_HANDOFF.md
git commit -m "docs: record S04R naive credential rollout evidence"
```

---

## Self-review result

- Spec coverage: migration, secret isolation, current credential import, multi-credential validation, narrow privileged agent, byte-preserving Caddy transform, validate/reload/rollback, optimistic concurrency, idempotency, last-active guard, soft revoke, DB-commit compensation, Owner API, one-time generated password, responsive UI, rehearsal and production gates are all mapped to tasks.
- Domain integrity: existing `pvnaive.credentials` is intentionally untouched until S05/S06 user/subscription mapping exists.
- Runtime revisions: existing `pvnaive.runtime_revisions` is reused; migration 0003 adds only durable idempotency support rather than creating a second revision system.
- No placeholder implementation steps remain; every production task has a failing-test gate before implementation.
