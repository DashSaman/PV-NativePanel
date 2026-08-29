# S06 Owner Customer Operations — Release & Production Evidence

Date: 2026-08-29
Branch: `s06-owner-customer-ops-work`
Production source commit: `e8f9fe5add4caafd154bd0155c7538bd41c65d0f`
Exact-source CI run: `33260443223` / run #720
Exact-source CI result: **PASS**
Production verification result: **PASS**

> The documentation commit that records this evidence is newer than the production source commit above. Production binaries and web assets were built from exactly `e8f9fe5add4caafd154bd0155c7538bd41c65d0f`; the newer commit changes documentation only.

## Exact-source CI gates

The exact production source commit passed all GitHub Actions gates before deployment:

- Go formatting: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS
- Runtime Agent safety rehearsal: PASS
- Web tests: PASS
- Web production build: PASS
- PostgreSQL 18 migration/health/backup/restore gates: PASS
- pinned forwardproxy accounting-boundary proof: PASS
- pinned Naive Caddy multi-`basic_auth` proof: PASS
- S04 authentication rehearsal: PASS
- S04R full runtime rehearsal: PASS
- S06 production bundle build: PASS
- S06 production bundle contract: PASS
- production archive checksum verification: PASS
- artifact upload: PASS

## Exact production artifact

GitHub Actions artifact ID: `9717124061`
Artifact name: `PVNaive-S06-Owner-e8f9fe5add4caafd154bd0155c7538bd41c65d0f`
Artifact ZIP size: `8488776` bytes
Artifact ZIP SHA-256 reported by GitHub and independently verified on the production host:

`dd9841cc91685a6a4f0219b375c08b42f3e9b59ce8457093a71b10612d1c4326`

Inner archive:

`PVNaive-S06-Owner-e8f9fe5add4c.tar.gz`

Archive SHA-256:

`f6a00abf897fc92b3bc11e30c4fd8956ae6ca813f4c9fa2e698ff25938bdd104`

The extracted bundle passed strict `SHA256SUMS` verification before the production upgrade.

## Production preflight

Target host: `testAmir5-3`

The exact S06 preflight ran before mutation and returned `PREFLIGHT_RESULT=PASS`.

Verified baseline:

- Caddyfile SHA-256: `be970b8478b995cac3f522f4f57ad4d2db5f794cb0fd1c267b99cc8724874955`
- Caddy MainPID: `1045`
- Caddy `NRestarts`: `0`
- Caddy validation: PASS
- pinned `http.handlers.forward_proxy` module: PASS
- `caddy-naive.service`: active
- `pvnaive-api.service`: active and ready
- `pvnaive-runtime-agent.service`: active and healthy through its Unix socket
- Runtime key baseline: PASS
- database schema: `7`
- configured expected schema: `7`
- target schema: `8`
- Naive public host: configured and verified
- TCP listeners 22/80/443: present
- SSH service: active
- production `/panel/`: HTTP 200
- public `/`: HTTP 200

## Backup and production upgrade

The verified bundle was executed with the exact Caddy SHA captured by preflight.

Upgrade output returned:

- `S06_RESULT=PASSED`
- `SOURCE_COMMIT=e8f9fe5add4caafd154bd0155c7538bd41c65d0f`
- `SCHEMA_VERSION=8`
- `CADDY_ACTION=none`

Encrypted database backup created before migration:

`/var/backups/pvnaive/database/20260829T153952Z-449015-H8O4gf/pvnaive.dump.age`

The backup is non-empty and has a valid `age-encryption.org/v1` envelope.

Installed web releases:

- `/opt/pvnaive/web/releases/20260829T153952Z-e8f9fe5add4c`
- `/var/www/pvnaive-preview/releases/20260829T153952Z-e8f9fe5add4c`

Migration `0008_subscription_profile_projection.up.sql` was applied. Its database-recorded checksum exactly matches the release file:

`e899e76420aa49a043c929f8737bbc8df54ef6cb4abb2dc903405f85fbf52ac8`

## Independent production postflight

A separate verification was executed after the upgrade, outside the upgrade script, and returned:

`PRODUCTION_VERIFY=PASS`

It independently verified:

- production database schema: `8`
- `/etc/pvnaive/db.env` expected schema: `8`
- API readiness: PASS
- Runtime Agent Unix-socket health: PASS
- Caddy service active: PASS
- production `/panel/`: HTTP 200
- public `/`: HTTP 200
- served panel JS and CSS are from the `e8f9fe5add4c` release
- new Owner UI markers are present in the assets served by production
- `/opt/pvnaive/web/current` points at the `e8f9fe5add4c` release
- `/var/www/pvnaive-preview/current` points at the `e8f9fe5add4c` release
- encrypted backup exists and is valid
- schema-8 migration checksum matches the release bundle

### Caddy invariants after deployment

- Caddyfile SHA-256: `be970b8478b995cac3f522f4f57ad4d2db5f794cb0fd1c267b99cc8724874955` — unchanged
- Caddy MainPID: `1045` — unchanged
- Caddy `NRestarts`: `0` — unchanged
- Caddy action performed by S06: `none`

Therefore this production release did not replace, reload or restart the Naive Caddy service and did not mutate the public camouflage root.

## S06 customer/product behavior deployed

Production now includes:

- unified directory for managed customers and existing Runtime accounts
- adoption of existing Runtime accounts without changing their existing Runtime UUID/username/password
- visible Create, Edit and Details workflows
- search/filter/sort/pagination and selection/bulk ergonomics
- numeric GB quota or unlimited policy
- Set Total and Add Volume operations
- creation-time, first-successful-connection and fixed/manual validity policies
- Extend Validity operation
- Suspend / Resume / safe Revoke-Delete
- password rotation independent from Subscription rotation
- read-only current Subscription/QR retrieval for recoverable tokens
- explicit separate Subscription reissue
- local QR generation without sending sensitive URLs to a third-party QR service
- branded browser Subscription/status page while compatible raw subscription clients continue to receive the direct Naive URI
- Owner dashboard KPI/status/expiry visualizations based only on data the system can currently prove

## Capability gates intentionally still closed

This release remains intentionally truthful about telemetry/enforcement that the pinned standard Naive/Caddy Runtime cannot yet prove:

- exact per-user upload/download accounting: **NOT PROVEN / NOT ENABLED**
- used/remaining traffic values: **NOT FABRICATED**
- hard quota enforcement based on exact traffic accounting: **NOT ENABLED**
- traffic reset/periodic reset: **NOT ENABLED**
- live online/session state: **NOT CLAIMED WITHOUT EVIDENCE**
- trusted Runtime producer for first-successful-CONNECT activation: **NOT PROVEN / NOT ENABLED**

The next engineering phase may implement those only after the pinned forwardproxy/Naive Runtime emits authenticated, restart-safe evidence through the Runtime Agent boundary.
