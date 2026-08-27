# S04-AUTH Design

Date: 2026-08-28
Stage: S04-AUTH
Status: design approved in chat; implementation pending

## Goal

Turn the current PVNaive HTTP scaffold into a real, localhost-bound authenticated management API with one-time owner bootstrap, Argon2id passwords, opaque browser sessions, CSRF protection, RBAC, TOTP MFA, recovery codes, audit events, and a minimal working login experience. S05+ must remain blocked until S04 production postflight passes.

## Existing constraints

- S03-DATABASE is PASSED on `testAmir5-3`.
- PostgreSQL is `18/main`, port `5432`, loopback-only.
- Current schema version is `1`.
- `actors.password_hash`, `actors.mfa_required`, `actors.status`, and `auth_sessions` already exist.
- `auth_sessions.token_hash` is SHA-256-sized and `pvnaive.set_request_context(bytea)` already binds signed transaction-local RLS context.
- HTTP routes already declare access classes: public, authenticated, reseller, admin, owner, auditor, operator.
- API default listener remains `127.0.0.1:8080` until reverse-proxy integration is validated.
- Caddy/NaiveProxy, SSH, and firewall must not be changed before the S04 localhost integration gate passes.
- No default password, JWT, plaintext session token in DB, TLS MITM, or secret-bearing logs.

## Security baseline

### Passwords

Use Argon2id via `golang.org/x/crypto/argon2`.

Parameters:
- memory: `19456 KiB` (19 MiB)
- iterations: `2`
- parallelism: `1`
- salt: `16` random bytes
- output: `32` bytes
- storage: PHC string

This follows the current OWASP minimum Argon2id recommendation. Password comparison must be constant-time. Unknown-email login must execute a dummy Argon2id verification to reduce user enumeration through timing.

### Session tokens

Use opaque random browser session tokens, not JWTs.

- raw token: 32 random bytes, base64url without padding
- DB: only `SHA-256(raw token)` in `auth_sessions.token_hash`
- cookie: `__Host-pvnaive_session`
- attributes: `Secure; HttpOnly; SameSite=Strict; Path=/`
- no `Domain` attribute
- idle session lifetime: 1 hour
- absolute session/family lifetime: 12 hours
- login sets `expires_at = min(now + 1 hour, absolute_expires_at)`
- refresh sets a new `expires_at = min(now + 1 hour, absolute_expires_at)` and never extends the absolute family lifetime
- successful login creates a fresh `refresh_family_id`
- refresh revokes the current row and inserts a new token in the same family
- reuse of an already-revoked refresh token marks reuse and revokes the whole active family
- logout revokes the current session
- password change or account disable revokes all active sessions for the actor

### CSRF

SameSite is defense in depth, not the only CSRF control.

Each session has a second 32-byte random CSRF token:
- DB stores only SHA-256 in a new `auth_sessions.csrf_token_hash`
- browser gets `__Host-pvnaive_csrf` with `Secure; SameSite=Strict; Path=/` and not HttpOnly
- state-changing requests must send `X-CSRF-Token` equal to the CSRF cookie
- server hashes the supplied value and compares it to the session-bound DB hash
- reject `Sec-Fetch-Site: cross-site`
- for unsafe methods, require same-origin `Origin` when present; if absent, fall back to strict CSRF token verification

### MFA

TOTP follows RFC 6238 behavior:
- HMAC-SHA1
- 30-second step
- 6 digits
- allow clock window `-1, 0, +1`
- persist `last_used_step` and reject replay of a previously consumed step

MFA secret is encrypted at rest with AES-256-GCM using a dedicated 32-byte key stored outside PostgreSQL at `/etc/pvnaive/auth.key`, mode `0640 root:pvnaive`. The backup age identity is never reused as an application encryption key.

Recovery codes:
- 10 codes on confirmation/regeneration
- each code has at least 128 bits of CSPRNG entropy
- displayed once
- only SHA-256 hashes stored because codes are high-entropy random values
- each code is one-time and atomically marked used

Owner bootstrap does not force MFA immediately, to avoid initial lockout. The owner can enroll TOTP after first login. `mfa_required=true` is enforced only when a confirmed TOTP factor exists or an administrator explicitly requires it after enrollment.

MFA removal is a sensitive operation: it requires the current password plus either a fresh TOTP code or an unused recovery code. Successful removal revokes all other active sessions for the actor and rotates the current session.

## Database migration 0002

Create `0002_auth_foundation.up.sql` and matching down migration.

### Actor login state

Add to `pvnaive.actors`:
- `failed_login_attempts smallint NOT NULL DEFAULT 0`
- `locked_until timestamptz`
- `password_changed_at timestamptz`

Lock policy:
- 5 consecutive failures -> 15 minute temporary lock
- successful login resets counter and lock
- existing `status='locked'` remains a manual administrative lock and is distinct from `locked_until`

### Session CSRF binding

Add to `pvnaive.auth_sessions`:
- `csrf_token_hash bytea NOT NULL DEFAULT gen_random_bytes(32)` with exact 32-byte check
- `absolute_expires_at timestamptz NOT NULL DEFAULT (clock_timestamp() + interval '12 hours')`
- check `expires_at <= absolute_expires_at`

Application inserts explicit values; defaults exist only to migrate safely.

### TOTP table

Create `pvnaive.actor_totp_factors`:
- `actor_id uuid PRIMARY KEY REFERENCES pvnaive.actors(id) ON DELETE CASCADE`
- `secret_ciphertext bytea NOT NULL`
- `secret_nonce bytea NOT NULL CHECK octet_length(secret_nonce)=12`
- `encryption_key_id text NOT NULL`
- `last_used_step bigint`
- `confirmed_at timestamptz`
- `created_at timestamptz NOT NULL DEFAULT clock_timestamp()`
- `updated_at timestamptz NOT NULL DEFAULT clock_timestamp()`

### Recovery codes

Create `pvnaive.actor_mfa_recovery_codes`:
- `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
- `actor_id uuid NOT NULL REFERENCES pvnaive.actors(id) ON DELETE CASCADE`
- `code_hash bytea NOT NULL CHECK octet_length(code_hash)=32`
- `used_at timestamptz`
- `created_at timestamptz NOT NULL DEFAULT clock_timestamp()`
- unique `(actor_id, code_hash)`

Direct table access to both MFA tables is revoked from `PUBLIC` and `pvnaive_app`. Authentication-specific SECURITY DEFINER functions expose only the minimal operations needed by the API.

## Authentication database boundary

Public login cannot query `actors` directly because the application role is subject to RLS before a request context exists. Migration 0002 therefore adds narrowly-scoped SECURITY DEFINER functions owned by `pvnaive_owner`, with fixed `search_path = pg_catalog, pvnaive, public` and `EXECUTE` granted only to `pvnaive_app`.

Required functions:

- `pvnaive.auth_lookup_actor(text)` -> actor ID, tenant ID, role, password hash, MFA flags/status, temporary lock timestamp
- `pvnaive.auth_record_login_failure(uuid)` -> increments failure state and applies temporary lock
- `pvnaive.auth_record_login_success(uuid)` -> clears temporary failure state and records `last_login_at`
- `pvnaive.auth_create_session(uuid, bytea, bytea, uuid, bytea, timestamptz, timestamptz)` -> creates a scoped session
- `pvnaive.auth_rotate_session(bytea, bytea, bytea, bytea, timestamptz)` -> atomic rotation / family-reuse detection
- `pvnaive.auth_revoke_session(bytea)` -> revoke current token
- `pvnaive.auth_revoke_actor_sessions(uuid)` -> revoke all actor sessions
- MFA fetch/store/confirm/consume helpers that never expose the database signing key and always verify actor identity/scope internally

The existing `pvnaive.set_request_context()` remains the only way normal authenticated handlers establish RLS identity.

## HTTP request transaction model

`NewServer` becomes dependency-injected rather than a static scaffold.

Authenticated request flow:
1. read session cookie
2. SHA-256 token
3. begin `*sql.Tx`
4. call `database.BindRequestContext(ctx, tx, hash)`
5. load principal from signed DB context/session
6. enforce route RBAC
7. enforce CSRF for unsafe methods
8. attach `*sql.Tx` and principal to request context
9. run handler
10. commit on success; rollback on failure/panic

This transaction boundary is required so S05 CRUD can reuse the exact signed RLS context without passing tenant IDs from HTTP input.

## RBAC

Access hierarchy is explicit, not inferred from PostgreSQL RLS:

- `Public`: no authentication
- `Authenticated`: any active authenticated actor
- `Owner`: owner only
- `Admin`: owner or admin
- `Operator`: owner or operator
- `Auditor`: owner or auditor
- `Reseller`: owner or reseller

Admin/operator/auditor are intentionally not interchangeable even though database RLS allows global tenant visibility for management roles. HTTP authorization remains mandatory.

## API behavior in S04

Make these routes ready:
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/me`
- `GET /api/v1/me/sessions`
- `DELETE /api/v1/me/sessions/{id}`
- `POST /api/v1/me/mfa/totp/enroll`
- `POST /api/v1/me/mfa/totp/confirm`
- `DELETE /api/v1/me/mfa/totp`

Login request:
```json
{"email":"owner@example.com","password":"...","totp_code":"optional"}
```

Unknown email, wrong password, locked account and wrong MFA all return a generic authentication failure to the client. Internal audit reason codes remain specific.

When MFA is configured and `totp_code` is absent, return `401` with stable code `mfa_required`; no temporary session is created. The client resubmits password + TOTP. This avoids an extra pre-auth challenge-token subsystem in S04.

`DELETE /api/v1/me/mfa/totp` requires JSON containing `password` and either `totp_code` or `recovery_code`; a session cookie alone is insufficient.

## Owner bootstrap

Create `scripts/auth/bootstrap-owner.sh`.

Requirements:
- must run as root on the target host
- refuses if any owner already exists
- prompts email, display name, password and password confirmation from TTY
- password never appears in argv, environment, history or log
- uses the compiled PVNaive password hasher
- performs the one-time INSERT through local PostgreSQL administration (`runuser -u postgres -- psql`) and immediately verifies exactly one active owner
- does not create a default password
- never prints the password or PHC hash

## Minimal panel/login UI

S04 includes a minimal functional browser login flow so the panel is not API-only:
- unauthenticated view: email/password form and conditional TOTP field
- authenticated shell: calls `/api/v1/me`
- logout works
- CSRF cookie/header wiring exists for unsafe API calls
- no user/reseller/runtime business controls are enabled before their stages

The existing visual shell is preserved; S04 should not redesign unrelated UI.

## Service and deployment

Add `pvnaive-api.service`:
- `User=pvnaive`
- listen only `127.0.0.1:8080`
- read `/etc/pvnaive/db.env`
- read `/etc/pvnaive/auth.key`
- hardening: `NoNewPrivileges`, `PrivateTmp`, `ProtectSystem=strict`, `ProtectHome=true`, explicit read/write paths only
- restart on failure with sane backoff

S04 stage order:
1. verify S03 marker and source commit
2. backup DB/config/Caddy/service state
3. disposable PostgreSQL migration/auth integration rehearsal
4. apply migration 0002 to production
5. generate auth encryption key if absent
6. build/install API and web assets
7. install/start API service on localhost
8. local API health/auth integration test
9. interactive owner bootstrap as a separate explicit operator step
10. login/session/MFA smoke test
11. only after local gate passes, stage Caddy reverse-proxy changes with `caddy validate`, reload, external HTTPS smoke test, and automatic rollback on failure
12. independent postflight

Caddy must not be restarted; use validated reload only. Existing NaiveProxy forwarding behavior must remain unchanged.

## Caddy exposure

Expose management UI/API on the existing TLS host without taking over NaiveProxy paths:
- panel: `/panel/*`
- API: `/api/*`
- localhost upstream: `127.0.0.1:8080`

The exact Caddy matcher must be derived from the live Caddyfile during S04 preflight and must preserve the existing forward_proxy/site behavior. No Caddy mutation is allowed before a tested backup and a validated candidate configuration exist.

## Audit events

Record at least:
- login success/failure
- temporary lock applied
- logout
- session refresh
- refresh-token reuse detection/family revocation
- session revoke
- MFA enrollment/confirmation/removal
- recovery-code consumption
- owner bootstrap

Never log passwords, raw session/CSRF tokens, TOTP secrets, TOTP codes, recovery codes, or MFA encryption keys.

## Tests / acceptance gates

S04 is PASSED only if all are true:

1. migration 0002 apply/reapply/rollback and checksum gates pass on PostgreSQL 18
2. Argon2id hash/verify tests pass and parameters are encoded/validated
3. login has generic failure behavior and dummy-hash timing path
4. session DB stores only hashes; raw token never appears in DB/logs
5. refresh rotates token; old-token reuse revokes family
6. CSRF rejects missing/mismatch/cross-site unsafe requests
7. RBAC matrix is fully table-tested
8. TOTP valid/window/replay tests pass
9. recovery code is one-time
10. owner bootstrap refuses a second owner and never logs password/hash
11. authenticated request binds RLS in a transaction and cross-tenant access still fails
12. API process runs as `pvnaive` on `127.0.0.1:8080` only
13. minimal login UI works through HTTPS
14. Caddy/SSH/firewall/NaiveProxy invariants remain intact
15. independent production postflight passes

## Dependencies

- Go 1.24 module remains the project baseline.
- `golang.org/x/crypto` is added for Argon2id; current module release checked during design was `v0.55.0`.
- PostgreSQL access uses `github.com/jackc/pgx/v5/stdlib`; v5 is the stable major and current 2026 changelog includes v5.10.0 hardening. Pin the exact version in `go.mod/go.sum` during implementation and run `govulncheck`/CI before merge.
- TOTP and AES-GCM use Go standard library to minimize dependency surface.

## Non-goals for S04

- user/reseller CRUD business logic (S05)
- runtime Caddy credential/accounting adapter (S06)
- subscription renderer (S07)
- notification delivery (S08)
- final installer/upgrade framework (S09)
- pilot/load rollout (S10)
