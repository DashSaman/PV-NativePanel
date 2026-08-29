# S06 Owner Customer Operations — Release Evidence

Date: 2026-08-29
Branch: `s06-owner-customer-ops-work`
Verified source commit: `b6b4de93e54a326b3674fad9197deba36b7b7af0`
CI run: `33255917733` / run #712
CI result: **PASS**

## Verified gates

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

## Production bundle

Archive: `PVNaive-S06-Owner-b6b4de93e54a.tar.gz`
Archive SHA-256: `6fd264d248469f2ee8636f99dadbd4170d2100644e4dcd29ccd727c6fe4ff773`
GitHub Actions artifact ID: `9715818739`
Artifact name: `PVNaive-S06-Owner-b6b4de93e54a326b3674fad9197deba36b7b7af0`
Artifact ZIP digest reported by GitHub: `sha256:fefb12b12ab943cfdbf35320c65d0055d5e59d6eb4e51f1d8652bfcc48b6cc7c`
Retention expiry: 2026-09-12

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

## Live deployment status

**NOT DEPLOYED YET.**

At verification time the connected SentinelX host list was empty, so no production host could be inspected or mutated from this session. Live installation must only proceed after the target server is connected and `scripts/stages/S06-owner-preflight.sh` passes. The production upgrade must use the exact archive and SHA-256 above, capture the expected current Caddy SHA from preflight, take the encrypted database backup, run `S06-owner-upgrade.sh`, and then record live post-deployment evidence.
