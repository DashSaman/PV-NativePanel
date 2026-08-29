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

## Task 1 — Schema 7 accounting + encrypted recoverable Subscription token

RED first: wire `tests/db/accounting_migration_test.sh` into PostgreSQL CI before it exists; observe a targeted missing-test failure. Then create the test while migrations are absent and observe schema-7-specific RED.

Schema 7 must add nullable `token_ciphertext`, `token_nonce`, `token_encryption_key_id` to `direct_subscription_tokens`; authoritative nonnegative BIGINT upload/download counters; connection sequence idempotency keyed by Runtime credential + connection UUID; a trusted accounting policy projection keyed by stable Runtime UUID; narrow SECURITY DEFINER authorize/current-totals/delta functions; duplicate sequence no-op; sequence gap/negative delta/overflow rejection; correct quota/expiry state; unmanaged Runtime tracked=false without fabricated counter rows; reversible 7→6 rollback preserving S05 state. Update expected-schema ceiling to 7 and legacy ceiling tests to reject 8.

Checkpoint: `feat(db): add schema7 exact accounting foundation`.

## Task 2 — Recoverable encrypted active Subscription token, read-only retrieval

TDD service/API/web behavior so create/adopt/reissue encrypt the current raw token with the existing runtime AES-GCM key while public resolution continues by SHA-256 hash. Owner-only `GET /api/v1/customers/{id}/subscription` decrypts current active token without rotation. Pre-v7 tokens with null recoverable material return explicit `subscription_reissue_required`. List APIs never expose raw token/ciphertext/nonce. Reissue remains explicit and separate from password rotation.

Checkpoint: `feat(subscription): persist encrypted active token for readonly retrieval`.

## Task 3 — Runtime Agent accounting policy service over AF_UNIX

TDD `POST /v1/accounting/authorize`, `POST /v1/accounting/delta`, `GET /v1/accounting/health`. Identity input is stable Runtime UUID after trusted Caddy auth mapping. Unmanaged+healthy => tracked=false/allowed=true; tracked suspended/revoked/expired/depleted => deny. Delta is strict/idempotent/atomic; gap fails closed; DB/service failure fails closed for accounting-enabled Caddy. Socket/DB configuration stays fixed-capability and non-request-controlled.

Checkpoint: `feat(accounting): add unix runtime policy and delta service`.

## Task 4 — Runtime UUID metadata in byte-preserving Caddy renderer

TDD renderer support for exactly one `pvnaive_runtime_id <username> <uuid>` per active credential and one fixed accounting socket directive in accounting mode. Duplicate/missing/malformed mappings fail closed; disabled/revoked identities are absent; zero-change remains byte-identical; legacy import remains possible before accounting activation.

Checkpoint: `feat(runtime): render trusted accounting identity metadata`.

## Task 5 — Minimal pinned forwardproxy accounting patch + deterministic Caddy integration proof

Pin Caddy v2.11.2 and forwardproxy `d62c80d3dd2c706b6b87579844d2397bddd18317`; store a small reviewable patch and reproducible build. Authorization occurs after successful Basic auth but before Fast-Open 200/target dial. Count only actual successful writes, including HTTP/1 prebuffer; exclude H2 Naive framing/padding. Emit Runtime UUID + connection UUID + monotonic sequence. Use max 16 KiB directional forwarding chunks with synchronous delta acknowledgement to target <=32 KiB aggregate active-connection overshoot. Accounting-agent failure is fail-closed. Prove HTTP/1 and HTTP/2; do not claim HTTP/3 yet.

Checkpoint: `feat(caddy): instrument pinned naive stream accounting`.

## Task 6 — Customer usage projection, effective quota state, Owner UI actions

TDD verified fields: upload/download/used/remaining/percent/accounting_updated_at/effective state. Quota increase can resume same credential if otherwise eligible; lowering below current usage depletes immediately; unlimited still measures but does not byte-deplete. UI must expose Total/Used/Remaining/progress/expiry plus Edit, read-only Details/QR/Copy Subscription/Copy Direct Link and separate explicit Reissue. No fake online/device/speed values.

Checkpoint: `feat(customers): show exact usage and stable subscription actions`.

## Task 7 — S06 release, rehearsal, guarded upgrade/rollback

Do not weaken S05 scripts. Add S06-specific preflight/upgrade/bundle/contracts/rehearsal. Bundle exact custom Caddy binary+SHA. Upgrade takes encrypted DB and exact Caddy/config/unit/API/runtime/web backups, migrates 6→7, stages/validates custom Caddy, performs one controlled restart because the binary changes, verifies TLS/listeners/camouflage/panel/API/runtime/legacy credential/canary accounting, and restores exact previous binary/config/schema on failure. No package/OS upgrade.

Checkpoint: `feat(release): add guarded s06 accounting rollout`.

## Task 8 — Verification, docs, review, capability gate

Fresh exact-HEAD verification must include formatting, vet, all Go tests, Web tests/build, pinned-forwardproxy proof, PostgreSQL tests, full disposable S06 rehearsal and S06 bundle. Only then update roadmap/worklog/status/known issues/feature matrix/owner requirements/handoff. `usage_capability.available=true` is forbidden until deterministic end-to-end accounting canary proof passes. Production remains a hard stop until live state is re-read and S06 read-only preflight proves the actual starting state.
