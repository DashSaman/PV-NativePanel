# S05 Handoff — Sanaei-style Naive customer flow

Last updated: 2026-08-29

## Where to continue

- Repository: `DashSaman/PV-NativePanel`
- Product name: `PVNaive`
- Feature branch: `s05-sanaei-customer-flow`
- Draft PR: `#6` — `S05: Sanaei-style customer service flow`
- PR base: `s05-user-quota-design`
- Parent S05 draft: PR `#5`
- Do not force-push or overwrite `main` production evidence.

## Scope completed in this branch

The Owner workflow is no longer credential-only. `/panel/#/customers` is the primary direct-customer page and `/panel/#/runtime/naive` remains the advanced raw Runtime manager.

Implemented customer creation flow:

- username;
- generated secure password or custom password;
- numeric GB quota or unlimited;
- validity from creation, first successful connection, or manual fixed expiry;
- business `User` + immutable `ServiceTerm` snapshot;
- stable binding to the real Runtime credential UUID;
- one-time generated password delivery;
- one-time opaque Subscription token delivery;
- copy-ready direct Naive URI when the password is available at creation;
- local browser-generated QR for the Subscription URL;
- safe customer list projection without password/token leakage;
- subscription reissue: previous active token revoked, replacement token shown once;
- idempotency claim before subscription rotation so network retry cannot rotate twice.

## Subscription security

Migration `0006_direct_subscription_tokens` adds the revocable direct Subscription projection.

- raw token: 256-bit CSPRNG, never persisted;
- persisted identity: SHA-256 hash + short non-secret prefix;
- public resolver joins live user/service/binding/runtime state;
- Runtime password stays encrypted at rest and is decrypted only inside the renderer;
- public URL host is not trusted as the Naive destination;
- `PVNAIVE_NAIVE_PUBLIC_HOST` must be explicitly configured when customer Subscription is enabled;
- QR is generated locally; no third-party QR endpoint receives a secret-bearing URL.

## Usage/quota boundary

`quota_bytes` is a commercial limit snapshot, not proof of consumption.

Until PVN-045..049 exact accounting gates pass:

- `usage_capability.available=false`;
- reason is `exact_accounting_not_proven`;
- UI displays usage/remaining as unavailable;
- UI must never synthesize `0 used` or `quota remaining`;
- hard byte-quota enforcement must remain disabled.

This is intentional and must not be bypassed for visual completeness.

## First-successful-connection boundary

`internal/runtimeevent` defines the only accepted first-use event contract:

- authenticated event;
- method `CONNECT`;
- stable Runtime credential UUID;
- trusted observed timestamp.

`customer.ActivateFirstUse` performs an atomic compare-and-set from pending first-use term to active and synchronizes the active Subscription expiry. Duplicate events are harmless.

**Important:** this branch does not claim that the pinned Caddy/Naive Runtime already emits that trusted event. Subscription fetch, panel view, reload and failed authentication explicitly do not count. Until producer instrumentation is proven end-to-end, production can safely use `on_creation` or `fixed_expiry`; first-use terms stay pending.

## Relevant files

- `web/src/Customers.tsx`
- `web/src/customers.ts`
- `web/src/qr.ts`
- `web/src/customers.css`
- `internal/customer/service.go`
- `internal/customer/store.go`
- `internal/customer/list_subscription.go`
- `internal/customer/firstuse.go`
- `internal/customer/firstuse_store.go`
- `internal/runtimeevent/firstuse.go`
- `internal/subscription/service.go`
- `internal/subscription/store.go`
- `internal/httpapi/customer.go`
- `internal/httpapi/customer_management.go`
- `internal/httpapi/subscription.go`
- `db/migrations/0004_customer_lifecycle_foundation.*.sql`
- `db/migrations/0005_customer_mutation_idempotency.*.sql`
- `db/migrations/0006_direct_subscription_tokens.*.sql`
- `docs/superpowers/specs/2026-08-29-sanaei-style-customer-service-ui-design.md`
- `docs/superpowers/plans/2026-08-29-sanaei-customer-service-flow.md`

## Deployment precondition added by S05

When Runtime/customer services are enabled the API startup now requires:

```text
PVNAIVE_NAIVE_PUBLIC_HOST=<real-naive-host-or-host:port>
```

Do not include `https://` or a path. Do not hardcode the panel Host header as the data-plane host.

## Verification rule

Before calling this feature complete, require one fresh CI on the final branch head with:

- Go formatting/vet/tests: PASS;
- Web tests/build: PASS;
- PostgreSQL 18 migration/backup/rollback tests: PASS;
- downstream rehearsal/bundle jobs: PASS when workflow conditions run them.

If any job is red, fix the root cause and update this handoff with the final run ID rather than claiming completion.
