# PVNaive exact per-user usage accounting — design

Date: 2026-08-29
Branch: `s05-sanaei-customer-flow`
Baseline commit: `9212675ad6f40cbbaf066d1503d58dd57f0afb4f`
Status: design approved in chat; implementation must not begin until this written spec is reviewed.

## Goal

Add real, per-user NaiveProxy byte accounting and quota enforcement without inventing usage numbers and without breaking existing Runtime credentials.

The customer UI must ultimately show, for every managed account:

- total quota
- bytes/GB used
- bytes/GB remaining
- percent used
- expiry and remaining time
- subscription link
- QR code
- edit action
- effective state (active / expired / quota depleted / suspended / revoked)

For unlimited accounts, usage is still measured and displayed, but no byte quota is enforced.

## Current constraint

The pinned `forward_proxy` handler authenticates and then directly proxies CONNECT traffic with its own hijack/streaming path. PVNaive currently has no trusted per-user byte event source. Linux process-level counters are not usable because all users share the same Caddy process; access logs cannot prove exact tunneled upload/download for each authenticated user. Therefore the existing `usage_capability.available=false` behavior is correct and must remain until this feature is proven end to end.

## Chosen architecture

Build a small PVNaive Caddy module and compile it into the pinned Caddy v2.11.2 binary together with the existing `forward_proxy` module.

The module sits in the request handling chain where an authenticated user identity is available and wraps the CONNECT data path so it can count successful tunneled bytes in both directions. It must not derive identity from untrusted request text; identity must come from the authenticated Caddy request context / placeholder produced by the auth layer, or from an equivalently trusted PVNaive-owned auth handoff proven by tests.

The module reports monotonic per-connection deltas to a local Unix-socket endpoint owned by the Runtime Agent. No usage telemetry is sent over the public network.

### Why not patch upstream `forward_proxy`

We do not want a long-lived fork unless required. The preferred order is:

1. use a separate PVNaive module that composes with the pinned upstream module;
2. if Caddy's handler boundary cannot observe exact CONNECT tunnel bytes, create a minimal PVNaive-owned wrapper/fork around only the streaming primitive, keeping the upstream commit pinned and the patch as small/tested as possible;
3. never rely on log scraping or process-wide counters as an accounting source.

If step 1 proves technically impossible during the implementation spike, the work must stop and the spec must be updated before adopting step 2.

## Counting semantics

Usage is the sum of application payload bytes successfully transferred through the authenticated CONNECT tunnel in both directions:

`used_bytes += client_to_remote_bytes + remote_to_client_bytes`

Rules:

- Count only bytes successfully transferred, not attempted writes.
- Do not count TLS framing, TCP/IP overhead, Caddy access-log bytes, subscription downloads, panel/API traffic, failed authentication, or failed CONNECT attempts.
- Each connection has a unique event/connection ID.
- Runtime Agent accepts idempotent sequence-numbered deltas so reconnect/retry cannot double count.
- Database stores integer bytes using `BIGINT`.
- UI converts bytes to GiB for display consistently with existing quota code (`1 GiB = 1073741824 bytes`).
- Accounting starts at zero when an existing legacy credential is adopted into Customer management unless an explicitly supplied trusted historical baseline exists. We will not fabricate historical usage.

## Enforcement semantics

The authoritative enforcement decision belongs to PVNaive policy state, not the browser.

Before opening a new authenticated CONNECT tunnel, the module asks the local Runtime Agent whether the credential is allowed:

- revoked/suspended -> deny
- expired -> deny and mark effective state expired
- finite quota with `used_bytes >= quota_bytes` -> deny and mark quota depleted
- otherwise -> allow

During an active connection, exact hard-stop at the final byte is not guaranteed unless the stream wrapper checks allowance while forwarding chunks. The implementation should enforce at chunk boundaries with a bounded overshoot no larger than the chosen buffer/chunk size. The test suite must state and prove the bound. Target bound: <= 32 KiB aggregate overshoot per active connection.

When a quota is exhausted:

- close the active tunnel as soon as the bounded chunk check detects depletion;
- reject subsequent CONNECT requests;
- do **not** rotate/delete the credential;
- keep Subscription and QR stable;
- if quota or expiry is extended later, the same credential works again without a password change.

This differs deliberately from credential revocation.

## Runtime Agent additions

Add authenticated local Unix-socket endpoints, available only on `/run/pvnaive/runtime-agent.sock`:

- `POST /v1/accounting/authorize`
  - input: trusted runtime credential UUID / authenticated identity
  - output: allow/deny, quota state, expiry state, current used bytes
- `POST /v1/accounting/delta`
  - input: credential UUID, connection ID, sequence number, upload delta, download delta
  - output: accepted/idempotent + current totals + continue/stop
- `GET /v1/accounting/health`
  - exposes capability/probe state, never secrets

The existing Runtime key and Unix-socket ownership model remain the trust boundary.

The module must fail closed for finite-quota accounts if the accounting policy service is unavailable, because silently allowing traffic would make quotas untrustworthy. For unlimited accounts, the rollout policy may initially be fail-open only if explicitly configured; default production behavior should still be fail-closed until proven otherwise.

## Database design

Introduce schema version 7 with reversible migrations.

Preferred tables/fields:

### `usage_counters`

One authoritative row per managed service term / runtime credential binding:

- `service_term_id UUID PRIMARY KEY/FK`
- `runtime_credential_id UUID NOT NULL`
- `upload_bytes BIGINT NOT NULL DEFAULT 0 CHECK >= 0`
- `download_bytes BIGINT NOT NULL DEFAULT 0 CHECK >= 0`
- `updated_at TIMESTAMPTZ NOT NULL`

`used_bytes` is computed as upload + download; remaining is computed from quota.

### `usage_connection_sequences`

Idempotency ledger for active/recent connections:

- `connection_id UUID`
- `runtime_credential_id UUID`
- `last_sequence BIGINT`
- `updated_at TIMESTAMPTZ`
- PK `(connection_id, runtime_credential_id)`

If a compact event table is required for audit, it must have retention/compaction rules; raw per-chunk events must not grow without bound.

Concurrency updates use transactions and row locking / atomic increments. `BIGINT` overflow must be guarded.

## Customer API changes

Managed customer views gain:

- `usage_capability.available=true` only after the production accounting module self-test passes
- `upload_bytes`
- `download_bytes`
- `used_bytes`
- `remaining_bytes` (`null` for unlimited)
- `usage_percent` (`null` for unlimited)
- `accounting_updated_at`
- effective service state

Subscription/QR operations remain independent of quota edits.

Editing quota or expiry must not mutate Runtime username/password/UUID.

When quota is increased after depletion, service can return to active if not expired/suspended/revoked. Decreasing quota below current usage immediately produces quota-depleted state.

## Persistent Subscription + QR UI

For every managed account, the table gets fixed actions:

- `Subscription` — show/copy current link
- `QR` — render current subscription URL locally in the browser
- `Rotate link` — explicit destructive subscription-token rotation with confirmation
- `Edit volume/date`

Normal viewing of QR/subscription must not rotate the token.

Rows show:

- Username
- Status
- Total quota
- Used
- Remaining
- Usage % / progress bar
- Expiry / remaining days
- Subscription
- QR
- Edit

The one-time password behavior for newly created credentials remains unchanged.

## Caddy build and supply-chain rules

Do not replace production Caddy with an unpinned ad-hoc build.

The build must pin:

- Caddy v2.11.2
- the exact currently proven `forward_proxy` dependency/ref
- the PVNaive accounting module commit

CI must produce the Caddy binary and its SHA256 as a release artifact. Tests must prove required modules are present with `caddy list-modules`.

No package upgrades are part of this rollout.

## Tests — mandatory TDD gates

### Unit

- exact upload/download counting
- partial writes count only written bytes
- retries do not double count
- sequence/idempotency behavior
- quota remaining/percent calculations
- unlimited behavior
- expiry behavior
- state recovery after quota extension

### Caddy module integration

- authenticated credential maps to correct Runtime credential UUID
- two simultaneous users never share counters
- failed auth counts zero
- CONNECT failure counts zero payload
- HTTP/1 CONNECT
- HTTP/2 CONNECT used by Naive
- if HTTP/3 is supported in the pinned path, test it or explicitly keep it outside the supported accounting capability
- active finite-quota connection stops within the documented overshoot bound

### Database

- schema 6 -> 7 -> 6 rollback against PostgreSQL 18
- atomic concurrent increments
- no negative counters
- idempotent duplicate deltas
- backup/restore includes counters

### API/Web

- used/remaining/percent shown from server values
- unlimited remaining/percent are displayed correctly
- QR uses the stable existing Subscription path
- viewing QR does not rotate token
- edit quota/date leaves Runtime credential unchanged

### End-to-end rehearsal

Use a local upstream that sends/receives deterministic byte counts. Prove exact expected totals for one user and isolation with a second user.

## Production rollout

No production mutation until all CI gates are green and the artifact is independently checksummed.

Rollout stages:

1. Build/test custom Caddy artifact in CI.
2. Preflight current production: Caddy SHA/PID/restarts, listeners, API, Runtime, DB, panel, backup ability.
3. Encrypted DB backup and filesystem backup.
4. Install schema 7/API/Runtime/web first while keeping accounting capability disabled.
5. Install custom Caddy binary via guarded replacement with an immediate rollback copy.
6. Validate config and module list before service mutation.
7. Controlled Caddy restart is unavoidable for replacing the binary; record old PID/restarts and require service/port/Naive checks immediately after.
8. Enable accounting for one dedicated test credential only.
9. Send deterministic traffic and compare client bytes with server counter.
10. Verify an existing legacy credential still connects unchanged.
11. Verify finite quota enforcement with a deliberately tiny test quota.
12. Only then set production `usage_capability.available=true` and allow normal users to use quota enforcement.

If any gate fails, rollback the Caddy binary/config, API/Runtime/web release links, and DB schema using the verified encrypted backup/rollback chain.

## Production invariants

Must preserve:

- `namir.softarg.ir:443`
- existing Runtime credential usernames/passwords/UUIDs
- owner access
- SSH
- camouflage site
- existing panel routes
- subscription-token confidentiality
- Caddy version v2.11.2 behavior except for the explicitly added accounting module

## Known limitations / explicit non-goals

- Historical traffic before accounting is enabled cannot be reconstructed accurately and starts at zero unless a trusted baseline exists.
- Network-layer overhead is not part of customer quota; quota is proxied payload bytes.
- A bounded small overshoot at quota boundary is acceptable only if its maximum is documented and proven by tests.
- No fake usage values are allowed if telemetry is unavailable; capability must fall back to unavailable/error state.

## Acceptance criteria

The feature is complete only when production evidence proves all of the following for a managed test credential:

1. 100 MiB deterministic proxied traffic produces the expected per-user counter within the defined payload semantics.
2. A second user remains isolated at its own count.
3. QR and Subscription stay available without token rotation.
4. Quota exhaustion blocks further CONNECT traffic and terminates active traffic within the bounded overshoot.
5. Increasing quota re-enables the same credential without password/UUID change.
6. Existing legacy credentials remain usable after the Caddy binary change.
7. API/UI show total, used, remaining, percent and expiry from authoritative server state.
8. Rollback procedure is rehearsed before broad enablement.
