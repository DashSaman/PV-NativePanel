# S04 independent postflight — core passed, periodic health release blocked

Timestamp: `2026-08-28T00:20:41Z`
Host: `testAmir5-3`

## Result

- `S04_POSTFLIGHT_CORE=PASSED`
- `S04_POSTFLIGHT=BLOCKED`
- S04 must NOT be marked PASSED yet.
- Owner bootstrap must wait until the periodic DB health release is aligned and the independent postflight is rerun successfully.

Exact blocker:

```text
BLOCKER=periodic pvnaive-db-health.service still uses the S03-era health release and does not verify S04 MFA secret tables
ACTION_REQUIRED=upgrade the immutable DB health release before marking S04 PASSED
```

## Verified live state

The postflight independently verified all of the following:

- `/opt/pvnaive/S04_AUTH.json` marker contract passed.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- Encrypted schema-v2 rollback backup: `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age`.
- Installed binaries, web release and API systemd unit matched the pinned deployed S04 bundle.
- `/etc/pvnaive/auth.key` metadata was exactly `root|pvnaive|640|32`.
- PostgreSQL schema version was 2.
- Migration 0002 identity was exactly:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- `pvnaive_app` and `pvnaive_owner` retained the intended restricted role attributes.
- Fresh S04-aware health directly from `/opt/pvnaive/auth/current/scripts/db/health.sh` passed:
  - `PVNAIVE_DB_HEALTH=OK`
  - `PVNAIVE_SCHEMA_VERSION=2`
  - `PVNAIVE_DB_USER=pvnaive_app`
  - signing-key direct SELECT denied
  - MFA secret-table direct SELECT denied
- Existing systemd periodic health execution returned `Result=success` and `ExecMainStatus=0`, but its immutable script release was still S03-era; `DB_TIMER_S04_AWARE=false`.
- `pvnaive-api.service` was enabled/active, ran as `pvnaive:pvnaive`, had zero restarts and listened only on `127.0.0.1:8080`.
- API liveness and readiness passed.
- Rollback backup checksum and encrypted archive parse passed.
- Caddy remained active; Caddyfile SHA-256 remained exactly:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- SSH remained active.

## Root cause

`pvnaive-db-health.service` executes:

`/opt/pvnaive/db/current/scripts/db/health.sh`

`/opt/pvnaive/db/current` still selects the immutable S03 database tooling release (`0001-...`). S04 installed its schema-2-aware health script under the auth release, so a direct S04 health check succeeds, but the periodic timer still runs the older immutable DB release.

The correct fix is NOT to weaken the health check and NOT to alter the timer unit. Promote `/opt/pvnaive/db/current` atomically to a new immutable schema-2 database tooling release while preserving the S03 release for rollback.

Expected new release ID from migration 0002 checksum:

`0002-84bb735877d5`

## Repository fix in progress

Active branch: `s04-auth`
PR: `#2`

TDD evidence:

1. `tests/stages/S04_db_release_promotion_test.sh` added.
2. RED CI run `33129595441`: missing `scripts/db/promote-release.sh`.
3. `scripts/db/promote-release.sh` implemented at commit `d8c4751b77e59e3c2cdcad2e55e34729c9e51403`; helper promotion behavior then passed.
4. Test strengthened to require S04 Stage wiring.
5. RED CI run `33129769272`: `ERROR: S04 stage does not require the DB release promotion helper`.
6. Guarded one-shot Stage patch succeeded and produced commit `708a4e7fd71011e5b21f136ae7305612f295a258`.
7. Temporary one-shot workflow was removed in normal-user commit `20ed774d06969a3f4c301fd6072a4db83fcffcca` so final CI runs on a clean branch head.
8. Final CI run for the clean head: `33130012929` (verification pending at the moment this evidence note was written).

## Exact continuation rule

Do not touch the live server until final CI on the clean `s04-auth` head is green. Then perform one atomic live DB tooling release promotion, require `pvnaive-db-health.service` success using the new release, preserve API/Caddy/SSH/firewall invariants, and rerun the independent S04 postflight. Only a postflight with `DB_TIMER_S04_AWARE=true` may advance to Owner bootstrap.
