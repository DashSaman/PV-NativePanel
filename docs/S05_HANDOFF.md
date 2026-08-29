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
- Runtime password stays encrypted at rest and is decrypted only inside the renderer;
- public URL host is not trusted as the Naive destination;
- `PVNAIVE_NAIVE_PUBLIC_HOST` must be explicitly configured when customer Subscription is enabled;
- QR is generated locally; no third-party QR endpoint receives a secret-bearing URL.

Management tables use FORCE RLS, so the public resolver deliberately does not bypass them. Instead migration 0006 maintains a narrow Subscription projection with SECURITY DEFINER trigger functions owned by `pvnaive_owner`:

- user state changes synchronize `user_state`;
- service-term state/expiry changes synchronize `service_state` and `expires_at`;
- active Runtime username/password-envelope rotation synchronizes the Subscription projection;
- disabling/revoking the Runtime credential revokes the active direct Subscription token rather than leaving a stale secret delivery path.

The resolver reads only this synchronized projection and requires active token/user/service state plus a non-expired term.

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

## Verified CI checkpoint

Feature-code checkpoint `26873435d0460d0297900629ece6a7fc93553c0a` passed CI run `33222914088` (run #570) on 2026-08-29.

Verified evidence from that run:

- Go formatting, `go vet ./...`, `go test ./...`, and Runtime Agent safety rehearsal: PASS;
- Web tests and production build: PASS (`22` tests across `6` files during bundle build);
- PostgreSQL 18 readiness, migration, health, backup/restore, backup collision, auth migration, Runtime migration, customer lifecycle migration, customer idempotency migration, and direct Subscription migration: PASS;
- exact pinned Naive Caddy multi-`basic_auth` proof: PASS;
- S04 authentication rehearsal: PASS;
- full S04R Runtime rehearsal: PASS;
- S04R production bundle contract and archive checksum: PASS;
- artifact upload: PASS, artifact ID `9705854213`;
- bundle: `PVNaive-S04R-26873435d046.tar.gz`;
- bundle SHA-256: `646b911394d1a373c70c6cca6d6c12816bd76f2afff1d6f8fa70b3988be5ccd7`.

Two CI-only regression defects were fixed on the way to this checkpoint without changing production evidence:

1. `tests/db/direct_subscription_migration_test.sh` had a PostgreSQL expression-precedence bug around `NOT (...) ||`; fixed by explicitly parenthesizing the boolean expression.
2. `tests/stages/S04R_full_rehearsal.sh` still expected schema version 5 and did not provide the new explicit S05 public-host configuration; it now expects schema 6 and uses the reserved test-only host `naive-rehearsal.example.invalid:443`.

This checkpoint proves the feature code/rehearsal/bundle path is green. It does **not** prove live deployment, exact per-credential traffic accounting, hard byte-quota enforcement, or a production trusted first-successful-CONNECT producer.

## Verification rule

Before calling a later branch head complete, require one fresh CI on that exact head with:

- Go formatting/vet/tests: PASS;
- Web tests/build: PASS;
- PostgreSQL 18 migration/backup/rollback tests: PASS;
- downstream rehearsal/bundle jobs: PASS when workflow conditions run them.

If any job is red, fix the root cause rather than claiming completion.
