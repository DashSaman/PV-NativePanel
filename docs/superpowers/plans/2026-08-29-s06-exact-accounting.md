# S06 exact Naive accounting + persistent subscription implementation plan

> **Execution:** Follow `superpowers:test-driven-development` for every production behavior change. Work only on isolated branch `s06-exact-accounting`. Production rollout remains blocked until the earlier live-pilot dependency is reconciled and an S06-specific preflight passes.

**Goal:** Implement exact per-Runtime-credential Naive payload accounting, hard finite quota enforcement with bounded overshoot, recoverable encrypted active Subscription tokens for read-only QR/subscription viewing, and the Owner-facing used/remaining/expiry actions required by `OWNER_REQUIREMENTS.md`, without changing Runtime credential identity or fabricating historical usage.

**Architecture:** PostgreSQL schema 7 owns synchronized accounting policy/counters and idempotent connection sequences. The root Runtime Agent exposes fixed AF_UNIX accounting authorize/delta/health endpoints and is the only data-plane policy bridge. A minimal patch against pinned `klzgrad/forwardproxy@d62c80d3dd2c706b6b87579844d2397bddd18317` instruments successful stream writes and trusted authenticated username→Runtime UUID mapping; no client-supplied accounting identity is trusted. The management API projects verified usage and decrypts current active Subscription token only for authenticated Owner read operations. S06 has a new release path because changing the Caddy binary requires one controlled restart plus exact rollback; the existing S05 no-Caddy-restart contract remains unchanged.

**Approved design:**
- `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`
- `docs/superpowers/specs/2026-08-29-naive-usage-accounting-forwardproxy-addendum.md`
- `OWNER_REQUIREMENTS.md`

**Tasks:** `PVN-045..049`, with subscription-read work also advancing the existing S05/PVN-052/055 behavior. Do not mark any roadmap task DONE until final review/evidence.

## Global constraints

- No production mutation in this implementation phase.
- No OS/package upgrade.
- No approximate access-log/process-level accounting.
- Existing Runtime username/password/UUID are invariant across quota/expiry changes and quota depletion.
- Legacy adopted managed accounts start exact usage at zero; no historical reconstruction.
- `usage_capability.available` stays false until deterministic end-to-end traffic proof succeeds.
- Failed auth / failed CONNECT count zero.
- Quota depletion is not Runtime credential revocation.
- Public Subscription resolution continues to use SHA-256 token lookup.
- Persistent token retrieval stores only AES-256-GCM ciphertext + 12-byte nonce + key-id; no raw token plaintext at rest/log/audit.
- View QR / Copy Subscription / Copy Direct Link / View Details are read-only and never rotate token/password.
- `Reissue Subscription` and `Rotate Password` remain separate explicit actions.
- Runtime Agent remains fixed-capability and AF_UNIX only; no arbitrary path/URL/service/shell surface.
- The S06 custom Caddy build pins Caddy v2.11.2 and exact forwardproxy commit; no moving tags/branches.
- First-use validity may consume the same trusted successful CONNECT identity later, but do not claim production first-use proof until its own live test passes.

---

## Task 1 — Schema 7 accounting + encrypted recoverable Subscription token

**RED first**

Files:
- Create `tests/db/accounting_migration_test.sh`
- Modify `.github/workflows/ci.yml` to run it before migration exists.

The new test must initially fail because schema 7 objects are absent. It must require:

1. migration `0007_exact_accounting.up.sql` / down migration and checksums;
2. `direct_subscription_tokens` gains nullable `token_ciphertext`, `token_nonce`, `token_encryption_key_id` so pre-v7 active tokens remain valid but explicitly unrecoverable;
3. `usage_counters` one authoritative row per active managed service-term/runtime binding, with upload/download BIGINT nonnegative and zero backfill;
4. `usage_connection_sequences` keyed by runtime credential + connection UUID with monotonic last sequence;
5. a synchronized accounting policy projection keyed by stable runtime credential UUID containing user/service/runtime state, quota, expiry and counters without requiring browser RLS context;
6. narrow SECURITY DEFINER functions for authorize/current totals and atomic idempotent delta application;
7. duplicate sequence = idempotent no-op; sequence gap = fail closed; negative delta / BIGINT overflow = rejected;
8. quota and expiry policy derive effective allow/deny from state, not a single enabled boolean;
9. unmanaged Runtime credential returns tracked=false and is never inserted into usage counters;
10. rollback 7→6 removes only schema-7 additions and leaves existing S05 customer/runtime/token state intact.

**GREEN**

Files:
- Create `db/migrations/0007_exact_accounting.up.sql`
- Create `db/migrations/0007_exact_accounting.down.sql`
- Modify `db/migrations/SHA256SUMS`
- Modify `scripts/db/set-expected-schema-version.sh` ceiling to 7
- Update legacy version ceiling tests to reject 8, not 7.

Verification:

```bash
bash tests/db/accounting_migration_test.sh
bash tests/db/rollback_chain_test.sh
go test ./...
```

Commit checkpoint: `feat(db): add schema7 exact accounting foundation`

---

## Task 2 — Recoverable encrypted active Subscription token, read-only retrieval

**RED first**

Files:
- Add customer/subscription service tests for encrypted raw-token persistence.
- Add API test for `GET /api/v1/customers/{id}/subscription`.
- Add web client test proving read-only GET has no CSRF/idempotency mutation headers and no rotation call.

Required behavior:

- create/adopt/reissue generate existing 256-bit token and hash;
- encrypt raw token using existing runtime AES-GCM primitive and persist ciphertext/nonce/key-id in the same transaction;
- pre-v7 token with null encrypted fields returns an explicit `subscription_reissue_required`, not silent rotation;
- authenticated Owner GET decrypts current active token and returns `subscription_path`; normal request does not update token/hash/password/quota/expiry;
- explicit reissue still revokes previous active token and creates a new encrypted one;
- list APIs never expose raw token/ciphertext/nonce.

**GREEN files likely:**
- `internal/customer/*subscription*`
- `internal/subscription/*`
- `internal/httpapi/customer_management.go` or a focused new handler file
- `internal/httpapi/routes.go`
- `web/src/customers.ts`
- tests beside each package.

Verification: targeted Go/Web tests + full Go/Web gates.

Commit checkpoint: `feat(subscription): persist encrypted active token for readonly retrieval`

---

## Task 3 — Runtime Agent accounting policy service over AF_UNIX

**RED first**

Files:
- Add `internal/accounting` store/service tests against PostgreSQL behavior.
- Add `internal/runtimeagent` protocol/server/client tests before handler implementation.
- Add `cmd/pvnaive-runtime-agent` startup tests/config contract requiring loopback PostgreSQL environment only.

Endpoints:

- `POST /v1/accounting/authorize`
- `POST /v1/accounting/delta`
- `GET /v1/accounting/health`

Contracts:

- input identity is stable Runtime UUID only; Caddy supplies it after trusted auth mapping;
- authorize returns tracked, allowed, reason, quota/used/remaining/expiry;
- unmanaged credential while service healthy => tracked=false, allowed=true;
- tracked suspended/revoked/expired/depleted => deny;
- delta requires valid UUIDs, sequence >=1, nonnegative directional deltas, bounded request size;
- duplicate sequence returns accepted/idempotent without increment;
- gap sequence fails closed;
- accepted delta atomically increments counters and returns `continue` based on post-delta policy;
- database/service unavailable => operation failure; accounting-enabled Caddy treats it fail closed;
- health proves DB/schema 7 functions, never secrets.

Security:

- keep socket path fixed;
- no arbitrary DSN from request;
- systemd gives Runtime Agent only the fixed DB environment it needs;
- add least-privilege DB role/grant contract if the existing root process should not connect as superuser.

Verification: Go unit/integration + PostgreSQL CI + Runtime Agent rehearsal.

Commit checkpoint: `feat(accounting): add unix runtime policy and delta service`

---

## Task 4 — Runtime UUID metadata in byte-preserving Caddy renderer

**RED first**

Add parser/renderer tests requiring accounting metadata to be managed atomically with active `basic_auth` entries while preserving unrelated Caddyfile bytes.

Behavior:

- each active Runtime credential emits exactly one `pvnaive_runtime_id <username> <uuid>`;
- fixed `pvnaive_accounting_socket /run/pvnaive/runtime-agent.sock` emitted once when accounting mode is enabled;
- duplicate/missing/malformed mappings fail closed;
- disabled/revoked credentials are absent from active auth + mapping output;
- zero-change render remains byte-identical;
- import of current legacy Caddyfile remains possible before accounting mode activation.

Files:
- `internal/naiveruntime/caddyfile.go`
- parser/renderer tests
- Runtime Agent desired-state DTO/operator tests as needed.

Commit checkpoint: `feat(runtime): render trusted accounting identity metadata`

---

## Task 5 — Minimal pinned forwardproxy accounting patch + deterministic Caddy integration proof

**Supply chain files:**
- Create `caddy/accounting/UPSTREAM_COMMIT`
- Create `caddy/accounting/forwardproxy-pvnaive-accounting.patch`
- Create `caddy/accounting/README.md`
- Create `scripts/build/build-pinned-accounting-caddy.sh`
- Create integration proof scripts/tests.

**TDD/proof order:**

1. CI contract references absent patch/build files => observe RED.
2. Add patch and build script pinned to exact upstream commit and Caddy v2.11.2.
3. Patch must fail provisioning unless accounting socket/mappings are consistent.
4. Authorization occurs after successful Basic auth but before Fast-Open 200 and target dial.
5. HTTP/1 pre-buffer write counts actual successful `n` only.
6. dual-stream upload counts only successful unpadded bytes written to target.
7. H2 padded download excludes 3-byte header and random padding; partial writes in header/payload/padding are tested.
8. connection UUID + monotonic sequence emitted to Runtime Agent.
9. tracked forwarding iteration max payload 16 KiB/direction and waits for delta ack, giving target <=32 KiB aggregate in-flight overshoot/connection.
10. stop immediately when delta response says continue=false.
11. agent error for accounting-enabled connection => deny/terminate fail closed.
12. unmanaged tracked=false credentials remain compatible but do not emit billing deltas.

CI must verify:

- patch applies with `git apply --check` and no fuzz/rejects;
- exact upstream SHA;
- resulting Caddy version v2.11.2;
- forward_proxy module present;
- PVNaive accounting capability marker/module/config behavior present;
- existing multiple-basic-auth proof still passes.

Deterministic traffic fixture must cover HTTP/1 and HTTP/2; do not claim HTTP/3 yet.

Commit checkpoint: `feat(caddy): instrument pinned naive stream accounting`

---

## Task 6 — Customer usage projection, effective quota state, Owner UI actions

**RED first**

Backend tests require managed customer list/details to expose verified accounting fields only when capability is enabled:

- upload_bytes
- download_bytes
- used_bytes
- remaining_bytes nullable for unlimited
- usage_percent nullable for unlimited
- accounting_updated_at
- effective service/quota state.

Until runtime capability probe is proven, the production/default capability response remains unavailable. Disposable integration can explicitly enable the proven fixture.

Quota edit behavior:

- increasing quota above current usage resumes same credential if no expiry/suspension/revocation blocker;
- decreasing quota below current usage makes effective depleted immediately;
- unlimited quota never enforces byte depletion but still displays exact measured use;
- quota/time edits do not rotate Runtime password or Subscription.

Owner UI (`/panel/#/customers`) must expose clearly:

- Total, Used, Remaining, progress %, Expiry/remaining time;
- Edit;
- read-only View Details;
- read-only View QR;
- read-only Copy Subscription;
- Copy Direct Link when recoverable token/config allows;
- explicit Reissue Subscription separately;
- keep any password-rotation action separate;
- no fake online/device/speed fields.

Pre-v7 tokens show a clear one-time `Reissue required to enable persistent QR` state rather than silently changing the link.

Verification: React tests + build + API tests; responsive layout regression.

Commit checkpoint: `feat(customers): show exact usage and stable subscription actions`

---

## Task 7 — S06 release, rehearsal, guarded upgrade/rollback

Do **not** modify S05 scripts to permit Caddy restart. Create separate S06 artifacts/contracts:

- `scripts/stages/S06-accounting-preflight.sh`
- `scripts/stages/S06-accounting-upgrade.sh`
- `scripts/release/build-s06-accounting-bundle.sh`
- `tests/stages/S06_accounting_upgrade_contract_test.sh`
- `tests/release/S06_accounting_bundle_contract_test.sh`
- full disposable S06 rehearsal.

Upgrade order:

1. verify exact source commit/bundle/internal checksums/custom Caddy SHA/modules;
2. read-only baseline: host, schema6 expected/live, API/Runtime/Caddy/SSH/listeners/TLS/panel/camouflage;
3. encrypted DB backup;
4. exact backups of current Caddy binary/config/unit + API/Runtime/web/env state;
5. migrate 6→7 and promote expected schema;
6. install Runtime/API/web/systemd changes;
7. stage custom Caddy binary and accounting Caddy candidate; validate before activation;
8. one controlled Caddy restart because binary changes;
9. verify exact binary/config SHA, service PID/restart expectations, 22/80/443/8080, TLS, camouflage, panel/API, runtime health;
10. run deterministic canary accounting test before capability activation;
11. preserve at least one legacy credential compatibility check;
12. any failure restores exact prior Caddy binary/config/unit and DB schema6 from validated backup/rollback path.

No broad production command is provided until this disposable rehearsal is green and the earlier production-state dependency has been re-read.

Commit checkpoint: `feat(release): add guarded s06 accounting rollout`

---

## Task 8 — Verification, docs, review, capability gate

Before any completion claim:

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
cd web && npm test && npm run build
bash tests/accounting/pinned_forwardproxy_boundary_test.sh
# all PostgreSQL migration/rehearsal/bundle gates
```

Required fresh CI on exact HEAD: go/web/database/rehearsal/S06 bundle all green.

Update only after evidence:
- `WORKLOG.md`
- `ROADMAP.md`
- `PROJECT_STATUS.md`
- `AGENT_TASKS.md`
- `KNOWN_ISSUES.md`
- `FEATURE_MATRIX.md`
- `OWNER_REQUIREMENTS.md`
- `docs/S05_HANDOFF.md` or new S06 handoff as appropriate.

`PVN-045` may move from TODO to IN_PROGRESS immediately on this branch. It may become DONE only after deterministic exact traffic proof. `PVN-046..049` transition independently only with their own restart/concurrency/ledger/quota evidence.

**Production hard stop:** even after implementation CI is green, do not mutate live Caddy/DB from this branch until current production truth is re-read and an S06 live preflight establishes the actual starting schema/Caddy state. The earlier roadmap dependency `PVN-029` remains a deployment dependency; isolated development is allowed, deployment is not.
