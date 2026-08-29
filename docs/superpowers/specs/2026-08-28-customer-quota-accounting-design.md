# PVNaive Customer Lifecycle, Exact Usage, Quota and Subscription Design

Date: 2026-08-28
Branch: `s05-user-quota-design` (branched from `s04-auth`)
Status: approved in chat; written spec created for final user review before implementation planning

## Work record

```text
AGENT: Agent-ARCH/REVIEW
TASK-ID: PVN-037, PVN-038, PVN-040, PVN-044, PVN-045..PVN-055
GOAL: turn a raw Naive Runtime credential into a real customer service with quota, expiry, exact usage, subscription link and QR without fabricating unsupported accounting
FILES: design/spec only in this work unit
DEPENDENCIES: PVN-029 live Pilot evidence; PVN-030..PVN-036 security/closure chain; existing S04R Runtime Agent and credential saga
```

## 1. Goal

PVNaive must evolve from the current Owner-only Naive Runtime credential manager into a production customer-management plane where an operator can create a real customer service such as:

```text
test1 -> 50 GB -> 30 days -> one Naive service -> subscription link + QR
```

The management plane must show and enforce:

- total traffic quota;
- upload, download and total consumed traffic;
- remaining traffic and percentage used;
- service start time, first-use time and expiry time;
- remaining days/hours;
- quota reset policy;
- active / suspended / expired / depleted / revoked state;
- renewable service terms;
- one or more Naive Runtime credentials bound to the business user;
- copy-ready Naive link;
- subscription URL;
- QR code generated locally from the secret-bearing URI/token;
- safe renewal, volume increase, reset-usage, suspend/resume and revoke operations.

A service whose quota is exhausted or commercial term expires must be denied at the Runtime layer without making the Web/API availability path part of existing data-plane sessions.

The primary correctness rule is: **no estimated traffic may be displayed or enforced as exact/billable usage.**

## 2. Scope and roadmap mapping

This design does not invent new permanent task IDs. It maps the approved product behavior onto the existing canonical roadmap:

- `PVN-037` — User CRUD + lifecycle state machine
- `PVN-038` — Plans + quota policy lifecycle
- `PVN-040` — User-bound Naive credential lifecycle
- `PVN-044` — Users/plans/resellers production UI
- `PVN-045` — exact per-credential accounting feasibility PoC
- `PVN-046` — restart/reconnect/double-count protection
- `PVN-047` — H2 multiplex/concurrent credential behavior
- `PVN-048` — usage delta collector + append-only ledger/reconciliation
- `PVN-049` — quota/reset enforcement
- `PVN-050` — concurrency/device capability proof only if technically proven
- `PVN-051` — full Naive Runtime adapter/capabilities/status
- `PVN-052` — subscription token lifecycle + renderer
- `PVN-053` — subscription info/usage page/API
- `PVN-054` — client compatibility matrix
- `PVN-055` — QR/templates/update compatibility

The existing dependency chain remains authoritative. In particular, production customer lifecycle work must not bypass `PVN-030..PVN-036` merely because this design is ready.

## 3. Current baseline

The existing S04R implementation already provides the important runtime safety boundary:

```text
Browser Owner
  -> unprivileged PVNaive API
  -> PostgreSQL desired state
  -> /run/pvnaive/runtime-agent.sock
  -> privileged Runtime Agent
  -> exact Caddy credential transform
  -> validate
  -> exact backup
  -> reload-only
  -> verify
  -> applied revision or exact rollback
```

This design preserves that architecture.

The current `pvnaive.naive_runtime_credentials` records are Runtime credentials, not commercial users. They must not be retroactively treated as users. A new business-layer User/Service model will reference or own Runtime credential bindings explicitly.

## 4. Design principles

1. **Business user != Runtime credential.** A customer/service has lifecycle, quota and expiry. A credential is only an authentication mechanism attached to that service.
2. **Exact accounting before quota enforcement.** If the meter is not proven, quota fields stay capability-unavailable.
3. **Append-only usage truth.** Aggregates are rebuildable; resets do not delete historical usage.
4. **Fail closed for commercial access, not for management-plane outages.** Existing already-authorized data-plane sessions must not depend on panel/API uptime, but new/continued access after a proven quota/expiry transition must be enforced by Runtime state.
5. **No secret-bearing third-party QR service.** QR is generated locally in the browser or server from a token/URI; the secret is never sent to an external QR API.
6. **No wire-protocol change for accounting.** Instrumentation may observe authenticated request/stream byte counts but must not modify Naive protocol semantics, padding, TLS, H2/H3 behavior or probe-resistance behavior.
7. **Capability-first UI.** If exact usage, sessions, device limits or speed limits are not proven, the UI must say unavailable rather than display fake values.

## 5. Accounting approaches considered

### Approach A — Caddy access-log estimation

Use request logs and response sizes to estimate per-user traffic.

Rejected because:

- CONNECT tunnels and multiplexed streams do not map cleanly to ordinary HTTP request sizes;
- payload, transport and proxy-layer bytes may be counted inconsistently;
- reconnect/retry behavior can double count;
- it cannot support a defensible exact quota/billing claim.

### Approach B — external socket/network meter

Use nftables/eBPF/interface counters or process-level traffic accounting.

Rejected as the primary per-customer source because multiple credentials share the same Caddy process, TLS listener, IP and HTTP/2/HTTP/3 transport. Network counters can measure server traffic but cannot reliably attribute it to an authenticated username without introducing another identity-aware layer.

### Approach C — authenticated Runtime-path meter

Instrument the exact Naive/Caddy forward-proxy path after authentication and around the actual proxied byte streams.

Chosen, subject to `PVN-045..047` proof.

Why:

- username identity is known at authentication time;
- byte reads/writes can be associated with the authenticated credential before aggregation;
- CONNECT/body/response stream handling can be measured at the point where the proxy actually transfers application bytes;
- HTTP/2 multiplexing can be attributed per request/stream while credentials are still known;
- the design can expose monotonic counters without logging destinations or payloads.

The current upstream/forked forwardproxy implementation already authenticates requests and supports HTTP/1.1, HTTP/2 and HTTP/3-style proxy operation. PVNaive must prove the exact installed Naive Caddy fork's code path and behavior before production use.

## 6. Accounting PoC contract (`PVN-045`)

Before any schema or UI is allowed to label traffic as exact, build a disposable accounting probe against the exact pinned Naive Caddy source/binary lineage.

The probe must answer:

- where the authenticated username is available in the request/stream lifecycle;
- which byte boundaries are counted for upload and download;
- whether CONNECT tunnels and ordinary proxy requests use the same meter path;
- whether HTTP/2 concurrent streams remain correctly attributed;
- how HTTP/3, if enabled in the exact build, behaves;
- whether padding/transport overhead is counted or excluded;
- whether counters can be monotonic across application-level requests;
- whether runtime reload causes active streams to restart/recount;
- whether abrupt disconnect/retry can double count;
- maximum observed error versus a controlled known-byte transfer.

### Accuracy acceptance gate

For controlled payload tests, define one explicit billing unit: **proxied application bytes** as observed at the authenticated forward-proxy transfer boundary, not TLS/IP overhead.

Acceptance requires:

- deterministic definition documented;
- per-credential upload/download attribution;
- controlled known-byte transfer error <= 0.5% or exact byte equality where implementation permits;
- no double counting across reconnect/retry tests;
- aggregate sum consistency;
- no destination URL/host storage required for accounting.

If this cannot be proven, `PVN-045` remains blocked and quota enforcement is not implemented.

## 7. Instrumented Runtime module

If the exact installed forwardproxy module does not expose a safe metrics hook, PVNaive may maintain a narrowly scoped patched module/fork for accounting instrumentation.

Constraints:

- no changes to authentication semantics except exposing the matched username internally;
- no changes to TLS, HTTP/2, HTTP/3, padding, probe resistance, routing or camouflage;
- no arbitrary callback URLs;
- no per-destination activity logging;
- no plaintext passwords in metrics labels/logs;
- usernames must not be emitted to public Prometheus-style endpoints;
- metrics transport remains local-only.

Preferred export interface:

```text
/run/pvnaive/usage.sock
```

or the existing Runtime Agent socket through a typed read-only usage operation if privilege/isolation analysis shows that reusing the socket is cleaner.

The meter exposes monotonic per-credential counters:

```json
{
  "generation": "opaque-boot-generation",
  "credentials": [
    {
      "runtime_credential_id": "uuid-or-safe-binding-id",
      "username_fingerprint": "safe-id",
      "upload_bytes": 123,
      "download_bytes": 456
    }
  ]
}
```

A stable credential ID is preferable to username as the durable accounting key because usernames may be renamed.

## 8. Runtime credential identity mapping

The meter must not use mutable username as the permanent business key.

Introduce a stable association:

```text
business user/service
    -> service credential binding
        -> runtime credential UUID
            -> current username + encrypted secret
```

Rename changes the username but not the Runtime credential UUID or usage ownership.

Rotate changes the secret but not usage ownership.

Revoke ends authorization but retains historical usage attribution.

## 9. Business data model

The exact migration number is assigned only during implementation after reconciling branch history, but the conceptual model is fixed.

### `pvnaive.users`

Commercial/customer identity record:

- `id uuid`
- `tenant_id uuid`
- `display_name text`
- optional safe customer reference fields
- lifecycle state (`draft`, `active`, `suspended`, `revoked`)
- timestamps
- optimistic revision

No Runtime secret is stored here.

### `pvnaive.plans`

Reusable policy template:

- `id uuid`
- `tenant_id uuid`
- `name`
- `quota_bytes bigint NULL` (`NULL` = unlimited if product permits)
- `duration_seconds bigint NULL`
- `reset_policy`
- optional capability-gated concurrency/device/speed fields that remain disabled until proven
- enabled/archived state
- revision/timestamps

### `pvnaive.service_terms`

A purchased/assigned service instance. This prevents historical plan changes from silently rewriting active customers.

- `id uuid`
- `user_id uuid`
- snapshot of quota/duration/reset policy applied to this term
- `purchased_at`
- `starts_at`
- `first_connected_at NULL`
- `expires_at NULL`
- `ended_at NULL`
- lifecycle state
- revision

A plan update affects new/renewed terms, not historical term semantics unless an explicit migration operation is executed.

### `pvnaive.user_runtime_credentials`

Binding table:

- `user_id`
- `service_term_id`
- `runtime_credential_id`
- role (`primary` initially; future aliases optional)
- bind/unbind timestamps

R1 should default to one active Runtime credential per customer service unless a real requirement proves multiple are needed.

### `pvnaive.usage_events`

Append-only normalized usage delta ledger:

- `id uuid`
- `runtime_credential_id`
- `service_term_id`
- `source_generation`
- `source_sequence`
- `observed_at`
- `upload_delta_bytes`
- `download_delta_bytes`
- `collector_instance_id`
- uniqueness over source generation/sequence/credential to make ingestion idempotent

No destination hostname, URL, request path or payload metadata is stored.

### `pvnaive.usage_aggregates`

Rebuildable acceleration table/materialization:

- current-cycle upload/download/total
- lifetime upload/download/total
- last source counter snapshot
- reset epoch/cycle ID
- last reconciled timestamp

This is not the sole source of truth.

### `pvnaive.quota_events`

Append-only business control events:

- reset
- quota increase/decrease
- renewal
- depletion
- expiry
- administrative override

Reset creates a new accounting cycle. It never deletes old usage rows.

## 10. Service state model

Business lifecycle and derived commercial state are separate.

Administrative state:

```text
draft -> active <-> suspended -> revoked
```

Derived access state:

```text
active
expired
quota_depleted
suspended
revoked
```

A user may be administratively active but commercially depleted. UI must show the reason explicitly rather than collapse every state into one boolean.

Effective access is allowed only when all required gates are true:

```text
admin_state == active
AND term_started
AND not_expired
AND not_quota_depleted
AND runtime_credential_not_revoked
```

## 11. Start and expiry semantics

Do not conflate purchase time, creation time and first connection.

Each term explicitly chooses one start policy:

- `on_creation`
- `on_first_successful_connection`
- `fixed_timestamp`

Default recommendation for customer plans: `on_first_successful_connection`, with an optional maximum activation window later if needed.

`first_connected_at` may only be set from a proven authenticated usage/session signal, not merely because the Owner created the user.

Expiry is computed from immutable term policy or stored explicitly and updated through audited renewal.

## 12. Usage collection (`PVN-046..048`)

A local collector polls/receives monotonic Runtime counters and converts them into idempotent deltas.

Flow:

```text
instrumented forwardproxy counters
  -> local usage collector
  -> compare generation + previous counter
  -> append usage_events deltas transactionally
  -> update rebuildable aggregate
  -> evaluate threshold/quota state
```

### Generation handling

Runtime counter reset is expected on process restart/reload depending on implementation. Every source lifecycle has an opaque `generation` ID.

If generation changes:

- never subtract a larger old counter from a new smaller one;
- begin a fresh source snapshot;
- append only bytes observed from the new generation;
- preserve old ledger totals.

### Double-count protection

Each observation has source sequence or snapshot identity. Database uniqueness makes retries idempotent.

Collector crash after DB commit but before local acknowledgment must replay safely without increasing total twice.

### Counter regression

A monotonic counter decrease within the same generation is a reconciliation error. Do not guess. Mark accounting degraded and block strict quota decisions that would depend on uncertain bytes until reconciled.

## 13. Quota enforcement (`PVN-049`)

Quota is enforced from durable, reconciled usage state.

### Normal path

1. collector appends new usage delta;
2. aggregate total reaches/exceeds quota;
3. DB transitions service to `quota_depleted` through an idempotent control event;
4. desired Runtime state disables/removes that customer's active credential from the applied active set;
5. existing Runtime Agent uses expected SHA -> exact backup -> validate -> install -> reload-only -> verify -> rollback on failure;
6. enforcement revision is recorded.

### Race tolerance

A hard byte-perfect stop at exactly 50,000,000,000 bytes is unrealistic because bytes are observed in chunks and already-in-flight streams may continue briefly.

The product must define a bounded enforcement overshoot target during PoC/load tests. Example acceptance target for R1:

- collector/enforcer interval <= 2 seconds under healthy load;
- overshoot is measured and published in internal capacity evidence;
- no undercounting/double counting;
- UI labels quota as exact consumed bytes but acknowledges enforcement can occur after a small in-flight overshoot.

Do not promise a zero-byte overshoot unless tests prove it.

### Management-plane outage

The collector/enforcer must run locally as a service independent from browser availability. If the Web UI is down but API/DB/runtime services remain healthy, enforcement continues.

If DB/accounting becomes unavailable, fail behavior must be explicit and conservative. Recommended R1 policy:

- already-established data-plane is not immediately killed by panel failure;
- new commercial state mutations are blocked;
- accounting state is marked degraded;
- no fake remaining quota is displayed;
- optional emergency policy can fail closed for new sessions only after measured operational validation.

## 14. Expiry enforcement

A local scheduler derives expiry transitions from authoritative DB time/state.

On expiry:

- append an expiry event once;
- transition effective state to expired;
- disable the user-bound Runtime credential using the same validated Runtime saga;
- preserve the credential row and historical usage for renewal/audit.

Renewal may reactivate the same Runtime credential or rotate it according to Owner-selected policy. Default: preserve credential on simple renewal unless compromised/security policy requires rotate.

## 15. Reset Usage behavior

`Reset Usage` means **start a new usage cycle**, not erase usage history.

Operation:

1. write quota reset event;
2. close prior cycle;
3. set current-cycle aggregate to zero using a new cycle ID;
4. preserve lifetime total and all usage_events;
5. if service was depleted only because of quota, re-enable it through the Runtime saga after reset commits.

UI shows both current-cycle usage and lifetime usage where useful.

## 16. Add volume and renewal

### Add volume

Owner can increase quota for the current term through an audited quota adjustment.

Example:

```text
50 GB original + 20 GB adjustment = 70 GB effective quota
```

Do not rewrite historical usage.

### Renewal

Renewal creates a new service term or schedules one after current term depending on product policy.

Recommended R1 choices:

- `renew_now`: begins new term immediately, resetting cycle according to plan;
- `extend_expiry`: preserves current cycle usage/quota and extends time;
- `add_quota_only`: keeps expiry and only changes volume.

These must be separate operations to avoid hidden billing semantics.

## 17. User-bound Runtime credential lifecycle (`PVN-040`)

Creating a customer follows a saga:

```text
create user/term desired state
-> create/bind Runtime credential
-> Runtime validate/apply
-> finalize user active state
-> return one-time secret/subscription handoff
```

If Runtime apply fails, business activation must not report success.

Generated password is returned once only after both Runtime and DB finalization succeed, using the existing S04R secret-return contract.

Credential rename/rotate/revoke stays on the existing Runtime machinery but authorization is initiated from the User page and records the user/service binding in audit metadata.

## 18. Plans (`PVN-038`)

R1 Plan fields:

- name;
- quota: fixed bytes or unlimited;
- duration: fixed duration or no expiry;
- start policy;
- reset policy: none / interval / calendar only after exact semantics are tested;
- enabled/archived.

Do not expose speed, device or concurrent-session limit fields as functional until `PVN-050` proves enforcement. If schema anticipates them, capability flags must keep UI disabled/hidden.

## 19. API contract

Exact endpoint naming may follow existing route registry, but implemented contracts are:

### Users

```text
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{id}
PATCH  /api/v1/users/{id}
POST   /api/v1/users/{id}/suspend
POST   /api/v1/users/{id}/resume
POST   /api/v1/users/{id}/revoke
POST   /api/v1/users/{id}/renew
POST   /api/v1/users/{id}/reset-usage
POST   /api/v1/users/{id}/quota-adjustments
```

### Credentials

```text
GET    /api/v1/users/{id}/credentials
POST   /api/v1/users/{id}/credentials
POST   /api/v1/users/{id}/credentials/{credentialId}/rotate
DELETE /api/v1/users/{id}/credentials/{credentialId}
```

### Usage

```text
GET /api/v1/usage/summary
GET /api/v1/usage/users/{id}
GET /api/v1/usage/reconciliation
```

### Plans

```text
GET  /api/v1/plans
POST /api/v1/plans
PATCH /api/v1/plans/{id}
```

All mutations require authenticated authorization, CSRF, idempotency key and optimistic revision where applicable.

Reseller tenant isolation is enforced in backend/RLS, never only by UI filters.

## 20. Subscription lifecycle (`PVN-052`)

A customer receives a long-entropy opaque subscription token.

Storage:

- only SHA-256 or stronger one-way token hash in DB;
- plaintext token is returned once at creation/rotation;
- revoke/rotate/expiry supported;
- token does not encode user ID, tenant ID or secret data.

Public URL example:

```text
https://namir.softarg.ir/s/<opaque-token>
```

The page/API uses `Cache-Control: no-store` and avoids internal identifiers.

It shows:

- service display name;
- effective status;
- total quota;
- upload/download/total used;
- remaining bytes and percentage;
- purchase/start/first connection/expiry times;
- remaining time;
- next reset time when applicable;
- last data update time;
- copy/import actions.

## 21. Naive config and QR (`PVN-052/055`)

The customer's direct configuration remains:

```text
naive+https://USERNAME:PASSWORD@HOST:443
```

For secret handling:

- the Owner gets a direct secret once after credential creation/rotation;
- the public subscription endpoint may render a current client configuration only for a valid opaque subscription token;
- responses are `no-store`;
- no direct credential is embedded in analytics logs or server access-log query strings.

### QR

QR is generated **locally**, never by a remote QR image service.

Preferred behavior:

- subscription page can generate a QR for the direct Naive URI in browser memory;
- Owner user-detail page can generate/copy QR after an explicit reveal action;
- QR component never persists image bytes server-side;
- QR secret value is not logged;
- user can copy URI and scan QR;
- representative Karing/compatible client import is tested under `PVN-054/055`.

If a JavaScript QR dependency is used, it must be pinned and included in supply-chain review; a remote CDN dependency is not acceptable for R1.

## 22. UI design (`PVN-044/053`)

### `/users`

Table/cards:

- username/display name;
- effective status;
- plan;
- used / quota;
- remaining volume;
- expiry / remaining time;
- last usage update;
- quick suspend/resume.

Filtering:

- active;
- suspended;
- expired;
- depleted;
- revoked;
- near quota threshold;
- near expiry.

### `/users/:id`

Primary summary cards:

- Status
- Plan
- Quota
- Used
- Remaining
- Upload
- Download
- Expiry
- Remaining time
- Current credential state

Actions:

- add volume;
- renew;
- reset usage;
- suspend/resume;
- rotate password;
- revoke credential/service;
- copy link;
- show QR;
- rotate subscription token.

Destructive or service-interrupting actions require confirmation.

### Dashboard

Only real metrics:

- active users;
- suspended/expired/depleted counts;
- aggregate exact usage only after accounting capability is healthy;
- users near quota/expiry;
- Runtime/accounting health;
- no fabricated online/session count before capability proof.

## 23. Customer status and error messages

The API/UI must distinguish:

- `administratively_suspended`
- `expired`
- `quota_depleted`
- `runtime_unavailable`
- `accounting_degraded`
- `credential_revoked`

A customer must not be told merely "disabled" when the actual reason is quota or expiry.

## 24. Reconciliation and operational safety

Usage accuracy is operationally critical.

A reconciliation job compares:

- latest Runtime monotonic counters;
- last collector snapshots;
- append-only ledger totals;
- current aggregate totals;
- term/cycle boundaries.

Any impossible regression or gap marks accounting state degraded and creates an audit/operational alert.

Quota enforcement must not silently proceed from corrupted/uncertain counters.

## 25. Privacy contract

Store only what is necessary for quota/accounting:

- credential stable ID;
- upload/download bytes;
- timestamps/generation/sequence;
- business user/service binding.

Do not store by default:

- destination domain;
- destination IP history;
- URL/path;
- DNS history;
- payload content;
- SNI browsing history.

Operational logs redact subscription tokens, passwords and secret-bearing URIs.

## 26. Authorization

Initial production roles:

- Owner: full users/plans/runtime/quota/subscription control;
- Admin: users/plans/subscriptions, but no privileged security secrets;
- Reseller: only users and plans permitted under own tenant/credit boundary;
- Operator: operational status only as defined by later matrix;
- Auditor: usage/audit read-only where authorized.

Every business row carries tenant ownership or derives it unambiguously. IDOR tests are mandatory.

## 27. Failure semantics

### DB commit fails before Runtime apply

No Runtime mutation; return failure.

### Runtime apply fails

Existing Runtime Agent restores exact last-known-good config; business activation/mutation remains uncommitted/failed.

### Runtime apply succeeds but DB finalization fails

Use existing compensation rule: restore pre-apply Runtime backup. If compensation fails, enter `runtime_reconciliation_required` and do not report success.

### Usage collector unavailable

Do not invent usage. Mark usage stale/degraded and expose last successful observation timestamp.

### Counter source corrupted/regresses

Stop exact quota decisions for affected scope until reconciled; preserve last durable ledger.

### Subscription page unavailable

Existing Naive credential/data plane continues; customer may still use previously imported direct config until quota/expiry Runtime state changes.

## 28. Testing strategy

Every production task follows RED -> observed failure -> minimal GREEN -> full relevant suite.

### Database

- migrations/up/down/checksums;
- RLS/tenant isolation;
- user state machine;
- plan snapshot semantics;
- append-only ledger;
- idempotent usage delta uniqueness;
- reset cycle without historical deletion;
- renewal/quota adjustment;
- subscription token hash/revoke/rotate.

### Accounting PoC

- controlled known upload/download sizes;
- CONNECT tunnel;
- H1;
- H2 multiple concurrent streams;
- H3 only if exact production build enables/supports it;
- reconnect;
- abrupt disconnect;
- Runtime reload;
- Caddy process restart;
- collector restart between observation and ack;
- duplicate observation replay;
- counter-generation change;
- same-generation counter regression fail-closed.

### Quota enforcement

- under quota remains active;
- threshold crossing creates one depletion event;
- credential removed/disabled via guarded Runtime saga;
- add-volume restores access;
- reset usage restores access when appropriate;
- expiry independently blocks access;
- simultaneous usage updates do not produce duplicate transitions;
- Runtime apply failure does not falsely mark enforcement applied.

### HTTP/API

- RBAC/RLS/IDOR;
- CSRF;
- idempotency;
- optimistic revision;
- strict JSON;
- no secret leaks;
- stable error reasons.

### Web

- user CRUD flows;
- quota/remaining calculations from server values;
- stale/degraded usage presentation;
- no fake online/device/speed metrics;
- QR generated locally;
- secret reveal/copy behavior;
- responsive desktop/mobile.

### Client lab

- Karing Windows/Android representative import;
- direct URI from QR and clipboard;
- password rotate invalidates old credential and new one works;
- quota depletion blocks access;
- add-volume/reset/renewal restores access according to policy.

## 29. Rollout order

The safe sequence is:

1. finish and record `PVN-029` Pilot evidence from the already-performed live credential/Karing smoke;
2. complete `PVN-030..PVN-036` P0/P1 auth/security closure according to the canonical dependency chain;
3. implement `PVN-037` user lifecycle;
4. implement `PVN-038` plan/term policy;
5. implement `PVN-045` accounting feasibility PoC before binding commercial quota behavior;
6. prove restart/reconnect and H2 behavior (`PVN-046/047`);
7. implement append-only collector/ledger/reconciliation (`PVN-048`);
8. implement quota/reset enforcement (`PVN-049`);
9. bind user lifecycle to Runtime credentials (`PVN-040`) using proven accounting/enforcement;
10. implement production user/plan UI (`PVN-044`);
11. implement subscription token/renderer (`PVN-052`);
12. implement subscription usage page (`PVN-053`);
13. execute client compatibility lab (`PVN-054`);
14. add local QR/templates only after client proof (`PVN-055`);
15. continue notifications, installer, disaster recovery, supply-chain and load/release gates.

Some code work can be parallelized later where file ownership is independent, but accounting proof gates quota claims and user-runtime commercial binding.

## 30. Production acceptance example

A representative final R1 smoke for one customer must prove the entire chain:

1. Owner creates plan `50GB / 30 days / first-use start`;
2. Owner creates customer `test1` from that plan;
3. Runtime credential is generated and applied safely;
4. Karing imports direct/subscription configuration and connects;
5. first successful authenticated traffic sets `first_connected_at` exactly once;
6. known test traffic increments upload/download/total for `test1` only;
7. UI/subscription page shows exact durable usage and remaining bytes;
8. password rotation disconnects old credential and new credential works;
9. quota crossing produces one depletion transition and prevents further authenticated use after bounded enforcement delay;
10. adding volume restores access without deleting usage history;
11. reset usage starts a new cycle and retains lifetime history;
12. expiry blocks access independently from quota;
13. renewal restores access according to explicit renewal mode;
14. QR imports correctly in a supported client;
15. Caddy changes remain validate/backup/reload-only/verify/rollback and never routine restart;
16. API/UI downtime does not rewrite or erase data-plane Runtime configuration.

Only after this evidence may PVNaive advertise exact traffic quota and usage for Naive customer services.

## 31. Explicit non-goals for this slice

- payment gateway;
- arbitrary Caddy editor;
- destination browsing analytics;
- random traffic/chaff generation;
- multi-node/fleet as an R1 dependency;
- unproven HWID/device/session/speed limits;
- estimated access-log billing;
- automatic deletion of historical usage/audit on reset;
- external QR generation service.

## 32. Architectural decision summary

Chosen architecture:

```text
Business User + Service Term + Plan
      |
      v
Stable Runtime Credential Binding
      |
      v
Instrumented authenticated Naive transfer counters
      |
      v
Local idempotent Usage Collector
      |
      v
Append-only Usage Ledger + Rebuildable Aggregates
      |
      +--> Quota/Expiry State Machine
      |       |
      |       v
      |   Existing guarded Runtime Agent
      |   (expected SHA / backup / validate / reload-only / rollback)
      |
      +--> User/Admin UI + Subscription page + local QR
```

The central gate is `PVN-045`: exact accounting is proven before quota is enabled. This keeps PVNaive honest, preserves the existing Naive data plane, and gives the requested Marzban/3x-ui-like customer experience without pretending unsupported capabilities exist.
