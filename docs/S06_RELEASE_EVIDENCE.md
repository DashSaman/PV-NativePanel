# S06 Owner Customer Operations — Release & Production Evidence

Date: 2026-08-29
Branch: `s06-owner-customer-ops-work`
Production source commit: `ee749b81f327c6130543b7593b8b1c067115b04d`
Exact-source CI run: `33256059786` / run #713
Exact-source CI result: **PASS**

> This document was updated after production deployment. The documentation-only commit that records this evidence is newer than the production source commit above; the binaries deployed to production are still exactly from `ee749b81f327c6130543b7593b8b1c067115b04d`.

## Verified CI gates for the production source

- Go formatting: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS
- Runtime Agent safety rehearsal: PASS
- Web tests: PASS
- Web production build: PASS
- PostgreSQL 18 migration/health/backup/restore contracts: PASS
- Public homepage contract: PASS
- Public portal routes contract: PASS
- S03/S04/S04R/S05/S06 contracts: PASS
- Pinned forwardproxy accounting-boundary proof: PASS
- Pinned Naive Caddy multi-`basic_auth` proof: PASS
- S04 authentication rehearsal: PASS
- S04R full runtime compatibility rehearsal on schema 7: PASS
- S06 production bundle build: PASS
- S06 production bundle contract: PASS
- Production archive SHA-256 verification: PASS
- GitHub artifact upload: PASS

## Exact production bundle

Archive: `PVNaive-S06-Owner-ee749b81f327.tar.gz`
Archive SHA-256: `98fcfafb105b545e51fc85ec8c16062984e02ce81ce39719d3e11dec364d4883`
GitHub Actions artifact ID: `9715863419`
Artifact name: `PVNaive-S06-Owner-ee749b81f327c6130543b7593b8b1c067115b04d`
Artifact ZIP size: `7551704` bytes
Artifact ZIP digest reported by GitHub: `sha256:f762dde5509e5c68f35146444c3fb82e8fb7ef2b3cae4606c2ab941b20d4c68e`
Retention expiry: 2026-09-12

The ZIP digest was verified again on the production host before extraction. The inner tar archive SHA-256 was then verified against the exact value above, followed by a strict `SHA256SUMS` check for the extracted release bundle and a `RELEASE.json` source-commit/schema check.

## S06 behavior included

- visible customer Edit and Details UX
- read-only retrieval of the current active Subscription/QR after recoverable token material exists
- encrypted-at-rest Subscription token recovery while public lookup remains hash-based
- explicit Subscription reissue separated from password rotation
- explicit password rotation separated from Subscription token rotation
- safe Suspend / Resume / Revoke-Delete operations through the Runtime safety boundary
- Add Volume delta operation
- Extend Validity operation
- existing account adoption without changing the existing username/password/Runtime UUID
- search/filter/sort/pagination/selection/bulk customer UI
- exact-accounting capability remains gated and no fake usage/remaining/online data is exposed

## Safety boundary

This S06 release intentionally declares:

- `usage_accounting_proven=false`
- `hard_quota_enforcement_enabled=false`
- `first_success_connect_producer_proven=false`
- `caddy_installer_mutation=false`
- `public_root_mutation=false`

S06 upgrades schema 6 -> 7 and must preserve the existing Caddyfile SHA, Caddy MainPID and Caddy restart count. It does not replace or restart Caddy.

## Live production deployment

**DEPLOYED AND INDEPENDENTLY VERIFIED — PASS.**

Target host: `testAmir5-3`
Deployment source: `ee749b81f327c6130543b7593b8b1c067115b04d`

### Preflight

The exact `scripts/stages/S06-owner-preflight.sh` from the production source commit ran on the connected production host before mutation and returned `PREFLIGHT_RESULT=PASS`.

Verified baseline included:

- Caddy configuration validation: PASS
- pinned `http.handlers.forward_proxy` module present: PASS
- `caddy-naive.service`: active
- Caddy MainPID captured: `1045`
- Caddy restart count captured: `0`
- `pvnaive-api.service`: active and ready
- `pvnaive-runtime-agent.service`: active and healthy through its Unix socket
- Runtime key baseline: PASS
- database schema: `6`
- expected database schema: `6`
- Naive public host configured
- TCP listeners 22/80/443 present
- SSH service active
- `/panel/`: HTTP 200 over local TLS resolution
- public `/`: HTTP 200 over local TLS resolution

### Backup and upgrade

The official bundle `scripts/stages/S06-owner-upgrade.sh` was executed with the Caddyfile SHA captured by preflight.

Upgrade evidence:

- encrypted PostgreSQL backup created and verified before migration
- migration `6 -> 7`: PASS
- installed API binary matches the verified release bundle byte-for-byte
- installed Runtime Agent binary matches the verified release bundle byte-for-byte
- installed password utility matches the verified release bundle byte-for-byte
- web release switched to the `ee749b81f327` release
- preview release switched to the `ee749b81f327` release
- `S06_RESULT=PASSED`

### Caddy invariants

After upgrade:

- Caddyfile SHA: unchanged
- Caddy MainPID: `1045` (unchanged)
- Caddy `NRestarts`: `0` (unchanged)
- Caddy action performed by S06: `none`

Therefore S06 did not replace, restart, or mutate the production Caddy configuration.

### Independent postflight

A separate postflight, outside the upgrade script, was executed immediately afterward and returned `POSTFLIGHT_RESULT=PASS`.

It independently verified:

- database schema: `7`
- configured expected schema: `7`
- API service active: PASS
- API readiness: PASS
- Runtime Agent active: PASS
- Runtime Agent Unix-socket health: PASS
- Caddy service active: PASS
- Caddy SHA unchanged: PASS
- Caddy PID unchanged: PASS
- Caddy restart count unchanged: PASS
- production `/panel/`: HTTP 200
- public `/`: HTTP 200
- web release points at `ee749b81f327`: PASS
- preview release points at `ee749b81f327`: PASS
- installed binaries match the verified bundle: PASS
- encrypted backup file exists and is non-empty: PASS
- recent error-priority journal entries for API: none
- recent error-priority journal entries for Runtime Agent: none
- recent error-priority journal entries for Caddy: none

## Accounting / first-use status after deployment

The production deployment intentionally remains truthful about unproven capabilities:

- exact per-auth usage accounting: **NOT PROVEN / NOT ENABLED**
- hard quota enforcement based on exact accounting: **NOT ENABLED**
- trusted live producer for first-successful-CONNECT events: **NOT PROVEN / NOT ENABLED**

The UI must continue to present usage as unavailable rather than inventing zero usage, remaining volume, online state, or other unproven telemetry.
