# Production fresh read-only audit — 2026-09-01

Audit type: **read-only**. No application, database, Caddy, service, schema, backup, release or credential mutation was performed.

## Access state

SentinelX execution access to `pv-primary` became available again during this coordinator run. The previous `upgrade_required` blocker is therefore not currently active for read-only commands.

## Fresh observed Production state

- Host: `pvnativepanel` / `pv-primary` control target.
- Running application process listens on `127.0.0.1:8080`.
- Naive/Caddy process listens on public `:80` and `:443`.
- Actual service units are:
  - `pvnaive-api.service` — **active**
  - `caddy-naive.service` — **active**
  - `pvnaive-runtime-agent.service` — **active**
  - `pvnaive-telemetry-agent.service` — **active**
- `GET http://127.0.0.1:8080/healthz` — HTTP **200**, body `ok`.
- `GET http://127.0.0.1:8080/readyz` — HTTP **200**, readiness checks report database/runtime-agent/telemetry-agent ready.
- Direct current DB health script `/opt/pvnaive/db/current/scripts/db/health.sh` — `PVNAIVE_DB_HEALTH=OK`.
- Database `max(schema_migrations.version)` — **19**.
- Running application environment advertises `PVNAIVE_EXPECTED_SCHEMA_VERSION=19`.
- Release metadata source commit — `0645b2e3a70a0a209f471953211bbdf079a2dd07`.
- Release metadata schema version — **19**.
- No warning-or-higher journal messages were observed in the preceding 30 minutes for the checked application/runtime/telemetry/Caddy units.

## systemd naming / historical failure note

A first generic probe checked legacy names `pvnaive.service` and `caddy.service`, which are inactive because Production uses `pvnaive-api.service` and `caddy-naive.service`. This is not an outage.

`pvnaive-db-health.service` is currently shown failed from a prior timer invocation. The same current DB health script was executed read-only during this audit and returned `PVNAIVE_DB_HEALTH=OK`; therefore the historical failed unit state is not used as evidence of a current database failure. It should be investigated/cleared separately only through a guarded maintenance change if needed.

## Deployment decision

**No deployment performed.** Production remains on schema19/source `0645b2e...` because Task13 is still under source recovery/reverification and Task15 schema20 does not yet have a green exact-head PostgreSQL18 gate. Task16 is also not validated.

Any future deploy still requires: exact eligible merged main; fresh encrypted backup and rollback snapshot; intended migrations only; readiness/runtime/telemetry/Caddy/customer/accounting postflight; exact deployed provenance; rollback on any failed invariant.
