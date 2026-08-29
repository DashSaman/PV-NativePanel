# WS2 — Customer Product Management

Date: 2026-08-29
Branch: `parallel/ws2-customer-product`
PR: #13

## Scope delivered

WS2 implements the Customer Management / Plans / Renewal / Search / Bulk / Groups / Tags / Reseller / RBAC workstream without replacing the existing working runtime credential lifecycle.

Delivered product behavior includes:

- customer product model with independent lifecycle, commercial, presence, quota, and runtime status dimensions;
- plan presets with quota, validity/start policy, no-expiry, reset metadata, default group, tags, enable/sort fields;
- renew-current, renew-with-plan/custom snapshot, and scheduled Next Plan semantics;
- customer groups, tags, notes, on-hold state, assigned actor metadata;
- database-backed search/filter/sort/pagination for customer lists;
- bulk dry-run/preview and idempotent execution records;
- database-atomic commercial bulk actions and per-customer safe runtime bulk coordination;
- Owner/Admin/Reseller action matrix and reseller-safe HTTP aliases;
- tenant-isolated RLS plus database trigger guards preventing cross-tenant Group/Tag/Plan associations;
- safe set-volume/add-volume/extend-validity operations without implicit password or subscription rotation;
- explicit subscription reissue/password-rotation separation.

## Security and isolation

The implementation preserves signed request context and PostgreSQL RLS as the tenant boundary. Operation tenant resolution uses `pvnaive.current_tenant_id()` from the signed request context. Customer↔Group, Customer↔Tag, Plan↔Group, and Plan↔Tag references are additionally validated by database guards so guessed UUIDs cannot create cross-tenant associations.

Reseller permissions deliberately exclude Owner-only destructive administration and direct runtime adoption. Safe Delete/Revoke and security-sensitive runtime changes continue through the existing runtime coordinator/idempotency model.

No usage value is fabricated when exact accounting proof is unavailable. WS1 accounting is integrated independently and supplies the exact-accounting path.

## WS1 integration

`main` advanced during WS2 implementation with WS1 exact accounting. Both workstreams originally used migration number 0009. The integration was reconciled without overwriting WS1:

- `0009_direct_naive_exact_accounting` remains WS1 accounting;
- WS2 Customer Product is `0010_customer_product_management`;
- combined latest schema is 10;
- migration manifest, health, backup/restore, rollback, lifecycle, and schema-range contracts were advanced accordingly.

The resulting combined tree retains WS1 telemetry/accounting code and WS2 product-management code.

## Verification performed

Fresh combined-tree verification before GitHub integration:

- `gofmt`: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS, including WS1 telemetry and WS2 customer packages
- Web tests: 12 test files / 38 tests PASS
- Web production build: PASS
- shell syntax for DB/stage test scripts: PASS

GitHub Actions is the authoritative final PostgreSQL 18/rehearsal/bundle gate and must be green on the final PR head before merge/deployment.

## Deployment contract

Production is not mutated by this workstream until the final GitHub gate is green. The production upgrade must:

1. take the repository-supported encrypted pre-migration database backup;
2. preserve Caddy/Naive runtime state and existing credentials;
3. deploy the full merged `main`, including WS1 telemetry/accounting and WS2 customer product code;
4. migrate PostgreSQL from the installed schema to latest schema 10 through the normal contiguous migration runner;
5. promote the immutable schema-10 DB tooling release;
6. deploy/restart the required PVNaive services without exposing secret material;
7. verify schema 10, DB health/RLS, API live/ready, runtime/telemetry services, Caddy, and panel assets.

## Remaining work

Only integration/release gates remain: final GitHub Actions green result, PR Ready + merge, and guarded production deployment/postflight verification.
