# WS4 — Operations / Observability / Notifications / Release / Fleet

Status: final integration and production verification in progress

## Starting point

- Starting `main` SHA: `28639c6fc93e19f98faab623607c32ca8a1b8436`
- Branch: `parallel/ws4-ops-observability`
- Original draft PR: #14
- Standalone operation remains authoritative; no controller/fleet service is required.

## Phase A checklist

| Area | State | Evidence / semantics |
| --- | --- | --- |
| CPU/RAM/disk/load/uptime/network | implemented | Linux `/proc` + `statfs`; bounded reads; tested collector |
| RX/TX live rate | implemented | server timestamp + monotonic counter delta; counter rollback/reset makes rate unavailable rather than fabricated |
| API/DB/Runtime dependency summary | implemented | `/api/v1/system/status`; DB ping + runtime-agent health with bounded timeouts |
| Monitoring UI | implemented | authenticated dashboard polls real status every 5s; server computes rates; 24-sample local sparklines |
| WS1 usage/online dependency | capability-gated | UI explicitly says unavailable; no synthetic usage/online count |
| Structured request logs | implemented | request ID, method, redacted path, status, duration, trusted client IP |
| Secret redaction | implemented/tested | password/token/bearer/DSN/Naive URI/subscription-path patterns plus structured sensitive-field redaction |
| HTTP abuse control | implemented/tested | server-side fixed window; login 12/min/IP, subscription delivery 180/min/IP, other API 900/min/IP; XFF trusted only from loopback reverse proxy |
| Safe diagnostics | implemented | `pvnaive doctor` PASS/WARN/FAIL + redacted diagnostic bundle script |
| DB backup | retained/hardened | existing age-encrypted custom pg_dump, checksums, ownership/ACL metadata and archive parse validation |
| Config backup | implemented | sensitive config streams directly from tar into age; no plaintext config archive is persisted |
| Scheduled backup retention | implemented | daily systemd timer; bounded PVNaive-only retention root |
| Restore drill | implemented | existing checksum/ownership/ACL/RLS verification reused; weekly disposable DB drill wrapper |
| Generic R1 build/deploy/rollback | implemented | checksummed bundle, mandatory pre-deploy encrypted DB backup, versioned web release, rollback, Caddy SHA/PID/NRestarts invariant |
| SBOM/provenance | implemented | source commit, Go module list, npm dependency JSON, full SHA256 manifest |
| Artifact signing | capability-gated | no signing key is embedded or invented; release manifest states unsigned when no signing key is supplied |
| OpenAPI | implemented/tested | `/api/v1/openapi.json` generated only from released/ready endpoints plus system status; scaffold endpoints excluded |
| Load rehearsal | implemented | bounded local control-plane HTTP rehearsal; explicitly not presented as a capacity ceiling benchmark |
| UI error fallback | implemented | React ErrorBoundary hides raw exceptions and provides safe reload/doctor guidance |
| Notification engine | foundation implemented/tested | in-app/Telegram channel abstractions, retry/backoff/dedupe, secret sanitization, quota events disabled without exact WS1 accounting |
| Operational event wiring | deferred | persistent notification rules/store and complete runtime/DB/backup event producer wiring are not required for the current deployment and are not falsely marked complete |
| Generic clean-server bootstrap | deferred | current production host already has the verified S02–S06 baseline; replacing those hardened bootstrap stages with a generic fresh-host installer is separate work and is not exercised during this release |

## Metrics contract

`observability.Collector` owns sampling state. Browser refresh intervals do not determine the reported network rate. A rate exists only when two server samples have increasing timestamps and non-decreasing byte counters. The endpoint reports `traffic_semantics=server_counter_delta`.

The system endpoint reports real host metrics and dependency health. It does not infer exact customer traffic, customer online state or session count. Those remain unavailable until WS1 exposes proved capability.

## Doctor / diagnostics contract

`pvnaive doctor [--json]` checks:

- PostgreSQL, Caddy Naive, Runtime Agent and API service state;
- Runtime Unix socket type/permissions;
- API live/ready endpoints and loopback listener;
- disk thresholds;
- latest encrypted backup freshness;
- authentication/runtime key file modes.

Any FAIL returns a non-zero exit code. Output passes through the same redaction boundary. `scripts/ops/diagnostic-bundle.sh` includes only service metadata, redacted recent journals, resource summaries, doctor output and environment *key names*. It never copies raw env values or key files.

## Backup / restore / release contract

Daily backup is encryption-first and uses existing `age` recipient/key material. Database backup preserves ownership and ACLs and is checksum-verified. Configuration tar data is piped directly into `age`. Weekly restore drills restore into a new `pvnaive_restore_test_*` database, validate schema/ownership/ACLs/RLS signing key and then remove the disposable database.

The R1 release builder executes Go formatting/vet/tests and Web tests/build, emits source provenance and SBOM material, builds static Go binaries, creates a complete SHA256 manifest and packages an R1 tarball. Deployment validates that manifest, requires live schema 8, takes an encrypted DB backup, stages web assets in versioned directories, replaces only PVNaive binaries/units, restarts PVNaive API/Runtime only and proves that Caddy configuration hash, PID and restart count did not change. An explicit rollback command restores the captured binary/unit/web pointers.

## Fleet Phase B

Fleet is intentionally a tested architecture foundation only. `internal/fleet` provides:

- stable Node ID/UUID field;
- node health/version/runtime state;
- desired/applied revision and drift state;
- capacity weight;
- maintenance/draining state;
- assignment-aware delete protection;
- Controller↔Node mTLS/short-lived certificate/replay-boundary design;
- explicit proof that standalone mode does not require a controller.

No controller or node registration service is started by R1.

## Verification performed before integration

On isolated server workspace `/opt/pvnaive-ws4-review`, using portable Go 1.25.0 and Node 24.8.0 without modifying the production toolchain:

- `gofmt -l .`: clean at verified head;
- `go vet ./...`: PASS;
- `go test ./...`: PASS including observability, notification, ops and fleet tests;
- `npm test`: PASS, 12 files / 36 tests;
- `npm run build`: PASS.

GitHub CI and production deployment evidence are recorded after the final branch head is frozen.

## Remaining non-blocking work

- persistent notification preferences/history and full event-producer wiring;
- a replacement generic clean-server installer for the already-verified historical S02–S06 bootstrap chain;
- optional artifact signing once an operator-managed signing key is provisioned;
- full fleet controller/node agent, canary orchestration and failover are future phases.

These items are intentionally not represented as already shipped functionality.

READY_FOR_INTEGRATION: NO — waiting for final GitHub CI and production deployment verification.
