# PVNaive exact accounting — pinned forwardproxy fallback addendum

Date: 2026-08-29
Branch: `s05-sanaei-customer-flow`
Parent design: `docs/superpowers/specs/2026-08-29-naive-usage-accounting-design.md`
Roadmap: `PVN-045..049`
Status: **proof complete; fallback design pending explicit Owner approval before production implementation**

## Proof-gate result

The preferred “separate Caddy handler only” architecture is **insufficient for exact per-user upload accounting** on the pinned Naive path.

Pinned upstream:

- repository: `klzgrad/forwardproxy`
- commit: `d62c80d3dd2c706b6b87579844d2397bddd18317`
- Caddy release line: v2.11.2 / Naive release used by the existing Pilot

Automated evidence on CI run `33227417609`:

```text
FORWARDPROXY_UPSTREAM_COMMIT=d62c80d3dd2c706b6b87579844d2397bddd18317
FORWARDPROXY_AUTH_LINE=256
FORWARDPROXY_TARGET_DIAL_LINE=333
FORWARDPROXY_HTTP1_STREAM_LINE=347
FORWARDPROXY_HTTP23_STREAM_LINE=352
PINNED_FORWARDPROXY_BOUNDARY_PROOF=PASSED
```

The same run passed Go/Web/PostgreSQL/rehearsal/bundle gates. The Go proof fixture demonstrates a short remote write where a preceding handler has already read more bytes from the client than `forward_proxy` successfully writes to the remote peer. Counting `r.Body` reads would therefore over-count upload in a real partial-write/error case.

The pinned source also proves:

- successful `basic_auth` sets the trusted Caddy replacer identity `http.auth.user.id`;
- `forward_proxy` itself creates the private remote `targetConn` after authentication;
- HTTP/1 authenticated CONNECT gives that connection to `serveHijack`;
- HTTP/2/3 authenticated CONNECT gives it to `dualStream`;
- `flushingIoCopy` is the primitive that observes actual `dst.Write` return counts;
- H2/H3 Naive padding is added/removed inside this same primitive.

A handler before `forward_proxy` can wrap the request body and response writer but cannot see successful writes to the private remote connection. A handler after `forward_proxy` does not own authenticated CONNECT streaming because `forward_proxy` terminates that path itself. Exact two-direction billing must therefore instrument the pinned forwarding primitive.

## Approved implementation candidate after Owner review

Use a **minimal PVNaive patch against the exact pinned upstream commit**, not log scraping and not an unbounded independent fork.

The repository will keep a reviewable patch and build metadata, for example:

```text
caddy/accounting/UPSTREAM_COMMIT
caddy/accounting/forwardproxy-pvnaive-accounting.patch
caddy/accounting/README.md
scripts/build/build-pinned-accounting-caddy.sh
```

CI must fetch exactly `d62c80d3dd2c706b6b87579844d2397bddd18317`, verify `FETCH_HEAD`, apply the patch with zero fuzz/rejects, build Caddy v2.11.2, and prove the resulting binary/modules before any release artifact is accepted. Moving branches/tags and unpinned `latest` are forbidden.

The patch must preserve ordinary `forward_proxy` behavior and Naive wire compatibility. It may add only the accounting hooks/configuration required below.

## Trusted identity and Runtime UUID mapping

Do not trust a username, UUID, or accounting header supplied by the client.

The existing successful authentication path is the trust source. On valid Basic proxy authentication, upstream already sets `http.auth.user.id` from the credential that matched using constant-time credential comparison.

PVNaive additionally needs the stable Runtime credential UUID. The patched Caddyfile grammar will accept PVNaive-owned metadata under the existing `forward_proxy` block, with a one-to-one mapping for every active credential, conceptually:

```caddyfile
forward_proxy {
    basic_auth alice <secret>
    basic_auth bob <secret>

    pvnaive_accounting_socket /run/pvnaive/runtime-agent.sock
    pvnaive_runtime_id alice 11111111-1111-1111-1111-111111111111
    pvnaive_runtime_id bob   22222222-2222-2222-2222-222222222222
}
```

The example secrets above are placeholders only; real secrets must never be committed or logged.

`internal/naiveruntime` will render the UUID metadata from trusted DB Runtime credential state. Metadata is written for **all active Runtime credentials**, not only Customer-managed accounts. Therefore adopting an existing Runtime credential into Customer management still does not change its username/password/UUID and does not require a credential rotation. Future Runtime credential mutations update `basic_auth` and UUID metadata together under the existing expected-SHA/validate/reload/rollback safety boundary.

When accounting is configured, Caddy provisioning must fail closed if:

- an active auth username has no Runtime UUID mapping;
- a mapping refers to an unknown/duplicate auth username;
- a UUID is malformed or duplicated;
- the accounting socket path differs from the fixed allowed Unix-socket path.

## Authorization path

For an authenticated CONNECT:

1. authenticate and obtain the trusted matched username;
2. map it to the configured stable Runtime UUID;
3. call Runtime Agent `POST /v1/accounting/authorize` over `/run/pvnaive/runtime-agent.sock` **before Fast-Open 200 and before target dial**;
4. only then open/stream the CONNECT tunnel if allowed.

Runtime Agent returns at minimum:

```text
runtime_credential_id
tracked
allowed
reason
quota_bytes | null
used_bytes
remaining_bytes | null
expires_at | null
```

If the Runtime UUID has no active Customer/service binding, authorize returns `tracked=false, allowed=true` while the agent is healthy. The proxy preserves that legacy Runtime credential and does not emit billing deltas for it. Once adopted, the same UUID becomes tracked from zero historical usage.

For the initial correctness rollout, accounting-agent communication failure is fail-closed for authenticated CONNECT while the accounting-enabled binary/config is active. This avoids silently bypassing finite quota. Availability relaxation for proven-unlimited/unmanaged cached policy is a separate future design and is not part of the first exact-accounting release.

## Exact payload counting inside the pinned forwarding primitive

`used_bytes = upload_payload_bytes + download_payload_bytes`.

Only successful payload bytes count.

### HTTP/1 CONNECT

`serveHijack` has two upload paths that must both be accounted:

- any bytes already buffered in `brw.Reader` and written to `targetConn` before the main copy loop;
- the later bidirectional `dualStream(..., padding=false)` traffic.

The current upstream pre-buffer write discards the returned byte count (`_, _ = targetConn.Write(rbuf)`). The PVNaive patch must record the actual `n` returned by the write and must not count bytes the remote write did not accept.

### HTTP/2/3 Naive CONNECT

The pinned code calls `dualStream(targetConn, r.Body, w, padding)`.

Client → remote with `RemovePadding` writes only the already-unpadded payload to `targetConn`, so successful upload payload is the actual `nw` returned by that target write.

Remote → client with `AddPadding` lays out the first padded frames as:

```text
3-byte Naive padding header | payload | random padding
```

Billing must exclude the header and random padding. For a successful/partial destination write of `nw` bytes with original payload length `payloadLen`, successful payload for that write is:

```text
clamp(nw - 3, 0, payloadLen)
```

for padded frames. For non-padded frames it is simply `nw`. Tests must cover writes that stop in the 3-byte header, inside payload, and inside trailing padding.

Do not use TLS/TCP byte counters, Caddy access-log sizes, or padded wire bytes as customer usage.

## Delta sequencing and bounded quota overshoot

Each CONNECT receives a random connection UUID. Deltas carry:

```text
runtime_credential_id
connection_id
sequence
upload_bytes
download_bytes
```

Sequence starts at 1 and increases monotonically for that `(runtime_credential_id, connection_id)` pair. Duplicate/retried sequence numbers must not add usage twice; a sequence gap fails closed.

For the correctness PoC, tracked traffic uses at most **16 KiB payload per directional forwarding iteration**, and the next iteration in that direction does not proceed until the local accounting delta is acknowledged. With at most one unacknowledged upload chunk plus one unacknowledged download chunk, the target aggregate in-flight overshoot bound is <= 32 KiB per active connection. Concurrent connections may each contribute their own bounded in-flight amount; PVN-046/047 must measure and document aggregate behavior before broad production enablement.

This synchronous correctness-first mode is not yet a performance claim. Throughput/CPU/agent/DB transaction cost must be benchmarked before release. Any later batching/lease optimization must preserve exact successful-byte semantics and a documented enforcement bound.

## Runtime Agent and database boundary

Runtime Agent remains the only local policy/accounting endpoint and remains unreachable over TCP.

Required endpoints:

- `POST /v1/accounting/authorize`
- `POST /v1/accounting/delta`
- `GET /v1/accounting/health`

Schema 7 will introduce an accounting projection/counter that can be used by narrowly scoped trusted DB functions without requiring browser/Owner RLS context on the data path. It must include enough synchronized policy state to authorize a Runtime UUID without trusting request text:

- tenant/user/service-term identity;
- Runtime credential UUID;
- user effective status;
- Runtime effective status;
- service state;
- quota bytes (`NULL` = unlimited);
- expiry;
- upload/download counters.

Existing managed bindings are backfilled with **zero** usage because historical traffic is not known. Existing Runtime credentials with no Customer binding remain untracked.

A separate sequence table enforces idempotency for connection deltas. BIGINT overflow, negative deltas, malformed UUIDs, stale/gapped sequences and inconsistent policy state fail closed.

The exact PostgreSQL functions/table grants are defined test-first in the schema-7 implementation plan. Direct arbitrary SQL or root shell is not exposed through the Runtime Agent.

## Quota and expiry recovery semantics

Quota depletion is business/policy state, not credential revocation.

When `used_bytes >= quota_bytes`:

- current tracked tunnels stop within the documented bound;
- later CONNECT is denied;
- username/password/Runtime UUID stay unchanged;
- Subscription token stays unchanged;
- increasing quota above current usage can reactivate the same credential if no other blocking state exists.

Decreasing quota below already-used bytes immediately yields effective `quota_depleted` on the next policy evaluation. Expired/suspended/revoked states remain stronger blockers.

## Persistent Subscription and QR security correction

Schema 6 intentionally stores only a SHA-256 digest and non-secret prefix for each direct Subscription token. The raw 256-bit token is one-time delivery and cannot be reconstructed later. Therefore a persistent Owner “Subscription” / “QR” action is impossible without adding recoverable secret material.

Schema 7 will keep the public resolver unchanged (`SHA-256(raw token)` lookup) and add **authenticated encryption at rest** for Owner retrieval:

- `token_ciphertext BYTEA`
- `token_nonce BYTEA` exactly 12 bytes
- `token_encryption_key_id TEXT`

The existing 32-byte Runtime encryption key is used with the existing AES-256-GCM secret primitive. Raw token plaintext is never stored in PostgreSQL, logs, audit metadata, Caddyfile or browser local storage.

On create/adopt/rotate:

1. generate the existing 256-bit random token;
2. calculate/store its SHA-256 hash for public resolution;
3. AES-GCM encrypt the raw token and store ciphertext+nonce+key-id;
4. deliver the raw token to the Owner UI.

Owner-only `GET /api/v1/customers/{id}/subscription` decrypts the current active token and returns its subscription path without rotating it. QR rendering remains local in the browser. Explicit `Rotate link` is the only action that revokes the old token and generates a new one.

Pre-schema-7 active tokens have no recoverable raw material. They must **not** be silently replaced. UI reports that one explicit rotation is required before persistent view/QR becomes available for such a token. After that rotation, retrieval is stable.

## New release/deployment boundary

Do not weaken S05's proven invariant that its upgrade does not mutate/restart Caddy.

Accounting ships as a new S06-style release path with separate contracts:

```text
scripts/stages/S06-accounting-preflight.sh
scripts/stages/S06-accounting-upgrade.sh
scripts/release/build-s06-accounting-bundle.sh
tests/stages/S06_accounting_upgrade_contract_test.sh
tests/release/S06_accounting_bundle_contract_test.sh
```

The S06 bundle contains the exact custom Caddy binary + SHA256 and all schema/API/Runtime/web changes. Upgrade must:

1. verify bundle and pinned Caddy identity/modules;
2. take encrypted DB backup;
3. back up current Caddy binary/config/unit and API/Runtime/web state;
4. migrate schema 6→7;
5. install Runtime/API/web changes;
6. stage and validate the new Caddy binary/config before activation;
7. perform one controlled Caddy restart because the binary itself changes;
8. prove listeners, TLS, camouflage, panel/API, Runtime Agent, existing legacy credential and canary credential;
9. restore the exact prior binary/config/unit/schema on failure.

No OS/package upgrades are part of S06.

## Capability activation gate

`usage_capability.available` remains false through implementation and disposable CI rehearsal.

Production enablement requires a dedicated canary credential and deterministic traffic evidence:

- known upload payload matches DB upload exactly;
- known download payload matches DB download exactly;
- failed auth = 0 bytes;
- failed target CONNECT = 0 bytes;
- second user isolation = exact;
- HTTP/1 CONNECT = exact;
- HTTP/2 Naive padding path = exact application payload, excluding padding;
- duplicate delta retry = no double count;
- tiny quota stops within <=32 KiB per active connection;
- quota extension resumes the same credential;
- at least one pre-existing legacy credential remains compatible;
- Caddy/Runtime/API restart/recovery behavior passes PVN-046 before broad rollout.

Only after these gates may the API/UI report exact accounting as available. Production evidence, not the presence of code, changes the capability flag.

## Explicit non-goals for the first implementation

- no historical usage reconstruction;
- no TLS/TCP overhead billing;
- no H3 capability claim unless independently proven;
- no concurrency/device-limit claim;
- no approximate log-based fallback;
- no implicit Subscription token rotation;
- no production Caddy replacement before a checksum-verified S06 artifact and guarded preflight.
