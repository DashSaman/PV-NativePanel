# S04 independent postflight — role-format harness false negative

Timestamp: `2026-08-28T00:41:21Z`
Host: `testAmir5-3`

## Result

The fresh independent S04 postflight stopped in the database-role assertion, but the stop was caused by the postflight harness expecting PostgreSQL booleans as `t/f` while the live query rendered them as `true/false`.

Observed live role rows:

```text
pvnaive_app|true|false|false|false|false|false|false
pvnaive_owner|false|false|false|false|false|false|false
```

These rows match the intended privilege contract semantically:

- `pvnaive_app`: LOGIN=true; SUPERUSER/CREATEDB/CREATEROLE/INHERIT/REPLICATION/BYPASSRLS=false.
- `pvnaive_owner`: LOGIN=false and all listed elevated attributes=false.

The failing harness expected:

```text
pvnaive_app|t|f|f|f|f|f|f
pvnaive_owner|f|f|f|f|f|f|f
```

Therefore this was a **postflight test-format bug**, not a server privilege failure.

## What had already passed in this run before the false negative

- S04 marker contract.
- Installed API binary, password helper, systemd unit, web release, and auth key metadata.
- PostgreSQL schema version 2.
- Exact migration 0002 identity:
  `0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- The observed live role values themselves are the intended restricted values.

The command explicitly reported `NO_CONFIGURATION_CHANGES_MADE=true`; it was read-only.

## Current live state remains

- `/opt/pvnaive/db/current` is already the repaired schema2 immutable release `0002-84bb735877d5`.
- `DB_TIMER_S04_AWARE=true` was established by the prior live repair.
- API remains healthy on `127.0.0.1:8080`.
- schema remains 2.
- Caddy/SSH/firewall remain unchanged.

## Exact next action

Rerun the same independent postflight with the role assertion corrected to the live PostgreSQL rendering:

```text
pvnaive_app|true|false|false|false|false|false|false
pvnaive_owner|false|false|false|false|false|false|false
```

No production mutation is required for this correction. Only after the corrected postflight reaches `S04_POSTFLIGHT=PASSED` may Owner bootstrap begin.
