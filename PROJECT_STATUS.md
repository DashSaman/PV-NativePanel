# PVNaive — Canonical Project Status

Last updated: 2026-08-29

> This file is the canonical **development** status for the active feature branch. Production truth must still be cross-checked against `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, and production evidence before any live mutation.

## Project

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. The active development slice moves beyond raw Runtime credentials into an Owner-facing direct-customer workflow while preserving the existing Runtime safety boundary.

## Repository state

- Repository: `DashSaman/PV-NativePanel`
- Default branch: `main`
- Active feature branch: `s05-sanaei-customer-flow`
- Draft PR: `#6` — `S05: Sanaei-style customer service flow`
- PR base: `s05-user-quota-design`
- Parent S05 draft: PR `#5`
- `main` remains the source of production evidence; do not force-reset or overwrite it.
- Detailed continuation notes: `docs/S05_HANDOFF.md`.

## Production state

This branch is a development branch, not proof of deployment. Existing production evidence on `main` remains authoritative for the live host. No statement in this file means the S05 customer flow has already been deployed to production.

## Implemented S04R foundation

The branch retains the tested S04R Runtime foundation:

- AES-GCM Runtime secret envelope and SHA-256 secret fingerprinting;
- byte-preserving Naive/Caddy credential parser and renderer;
- fixed Unix-socket Runtime Agent without arbitrary shell/path/service API;
- expected-SHA Caddy mutation with validate, exact backup, reload-only postflight, verification and rollback;
- Runtime credential PostgreSQL store and desired/apply/applied revision saga;
- Owner-only Runtime API with CSRF, idempotency and optimistic revision checks;
- `/panel/#/runtime/naive` advanced Runtime UI;
- one-time generated password delivery;
- stable Runtime credential UUIDs across business bindings.

## S05 direct-customer flow implemented on this branch

The primary Owner workflow is `/panel/#/customers`. It follows a Sanaei/3x-ui-like operator experience without collapsing commercial state into the Runtime credential.

Implemented:

- username creation;
- generated secure password or custom password;
- numeric binary-GB quota or unlimited (`quota_bytes = NULL`);
- validity from creation;
- validity from first successful connection;
- fixed manual expiry;
- business `User` plus immutable `ServiceTerm` snapshot;
- stable binding to the actual Runtime credential UUID;
- one-time password and direct Naive URI delivery when available;
- revocable opaque Subscription URL;
- browser-local QR generation without third-party QR service;
- safe customer list projection without password, ciphertext or raw token leakage;
- Subscription reissue with previous active token revocation;
- idempotency claim before Subscription rotation so request retry cannot rotate twice;
- explicit Naive public destination configuration through `PVNAIVE_NAIVE_PUBLIC_HOST` rather than request Host header;
- responsive desktop/mobile customer UI;
- raw Runtime credential management retained separately as the advanced screen.

## Subscription security boundary

Migration `0006_direct_subscription_tokens` stores only a SHA-256 token digest plus a short non-secret prefix. The raw 256-bit token exists only for one-time delivery. The public Subscription resolver checks live user/service/binding/Runtime state before decrypting the Runtime secret internally to render a `naive+https://...` entry.

Subscription fetches do **not** start first-use validity.

## First-successful-connection boundary

`internal/runtimeevent` accepts only an authenticated trusted `CONNECT` event carrying the stable Runtime credential UUID and trusted observation timestamp. `customer.ActivateFirstUse` then performs an atomic compare-and-set from `pending` to `active`, calculates expiry and synchronizes active Subscription expiry. Duplicate trusted events are harmless.

The branch does **not** claim that the pinned live Caddy/Naive path already produces this trusted event. Producer instrumentation remains a separate proof gate. Until it is proven end-to-end, production operators should use `on_creation` or `fixed_expiry` when automatic first-use activation is required immediately.

## Exact usage / quota boundary

Configured quota is implemented as commercial service state, but exact byte accounting is still capability-gated.

Until PVN-045..049 pass:

- `usage_capability.available=false`;
- reason is `exact_accounting_not_proven`;
- used/remaining traffic is not fabricated;
- UI does not display fake `0 used` or `quota remaining` values;
- hard byte-quota enforcement remains disabled.

This boundary is intentional and must not be bypassed for UI completeness.

## Deployment precondition introduced by S05

When Runtime/customer Subscription services are enabled, the API requires:

```text
PVNAIVE_NAIVE_PUBLIC_HOST=<real-naive-host-or-host:port>
```

Do not include a scheme or path. This value must be supplied by deployment configuration; it is not inferred from an HTTP Host header.

## Verification state

The S05 feature-code checkpoint `26873435d0460d0297900629ece6a7fc93553c0a` is verified green by GitHub Actions CI run `33222914088` (run #570).

Verified PASS scope:

- Go formatting, vet, full Go tests, Runtime Agent safety rehearsal;
- Web tests and production build;
- PostgreSQL 18 schema/migration, health, backup/restore, collision and S05 migration contracts through schema v6;
- pinned Naive Caddy multi-auth proof;
- S04 authentication rehearsal;
- full S04R Runtime rehearsal;
- production bundle contract, archive checksum and artifact upload.

Bundle evidence for that checkpoint:

- artifact ID: `9705854213`;
- bundle: `PVNaive-S04R-26873435d046.tar.gz`;
- bundle SHA-256: `646b911394d1a373c70c6cca6d6c12816bd76f2afff1d6f8fa70b3988be5ccd7`.

This is a development/release-candidate proof, not production deployment evidence. A fresh CI must still pass on any later branch head (including documentation-only finalization commits) before that later head is called final.

## Remaining blockers outside this feature slice

1. exact per-credential byte accounting and reconciliation proof (`PVN-045+`);
2. hard quota enforcement only after that proof;
3. trusted Runtime producer instrumentation for `on_first_successful_connection` before claiming live automatic first-use activation;
4. generic fresh-server installer/release lifecycle;
5. production rollout and evidence capture on `main` as a separate controlled operation.

## Read first for continuation

1. `docs/S05_HANDOFF.md`
2. `docs/superpowers/specs/2026-08-29-sanaei-style-customer-service-ui-design.md`
3. `docs/superpowers/plans/2026-08-29-sanaei-customer-service-flow.md`
4. `docs/ARCHITECTURE_FA.md`
5. `docs/DECISIONS_FA.md`
6. `docs/API_FA.md`
7. `ROADMAP.md`
8. before any live change: `main:CONTINUE_HERE.md`, `main:ops/S04_LIVE_STATE.md`, newest `main:ops/evidence/*`
