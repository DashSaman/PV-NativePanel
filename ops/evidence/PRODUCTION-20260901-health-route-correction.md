# Production health-route correction — 2026-09-01

This is a read-only evidence correction. No Production application, database, Caddy, service, schema, backup, credential or release mutation was performed.

A follow-up probe after the earlier audit verified that the deployed schema19/source `0645b2e4758d3cc6c197dd9dba9127e8de983d6c` binary does **not** expose `/healthz` or `/readyz`; both return HTTP 404. The canonical deployed routes, confirmed both by the running API journal and fresh direct probes, are:

- `GET http://127.0.0.1:8080/api/v1/health/live` → HTTP 200, `{"service":"pvnaive-api","status":"ok"}`;
- `GET http://127.0.0.1:8080/api/v1/health/ready` → HTTP 200, readiness reports DB/schema ready.

The same follow-up verified `pvnaive-api.service`, `caddy-naive.service`, `pvnaive-runtime-agent.service`, and `pvnaive-telemetry-agent.service` active; API on `127.0.0.1:8080`; Caddy on public 80/443; DB health script with `/etc/pvnaive/db.env` returned `PVNAIVE_DB_HEALTH=OK`, `PVNAIVE_SCHEMA_VERSION=19`, and direct secret/MFA SELECT denied.

Therefore the earlier evidence line claiming `/healthz` and `/readyz` returned 200 is superseded by this correction. The 404 results are route mismatch, not evidence of an outage. Production remains schema19/source `0645b2e...` and was not mutated during either audit.
