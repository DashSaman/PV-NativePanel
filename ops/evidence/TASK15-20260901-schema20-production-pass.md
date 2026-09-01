# Task15 schema20 Production rollout — 2026-09-01

Status: **PASS / Production**.

## Repository gate

- Exact merged `main`: `26aa74dddfd23535e45837f21531cf67ea2fd238` (merge of PR #54).
- Exact-main GitHub CI run `33471780919`: **success**.
- PR #54 was merged only after its required CI, WS1 Exact Accounting and WS1 Pinned Forwardproxy gates were green.

## Fresh preflight

Immediately before mutation on `pv-primary`:

- `pvnaive-api`, `caddy-naive`, `pvnaive-runtime-agent`, `pvnaive-telemetry-agent`: active.
- `/api/v1/health/live`: HTTP 200.
- `/api/v1/health/ready`: HTTP 200 with DB/schema ready.
- Database schema: 19 (`pvnaive.schema_migrations` and `/etc/pvnaive/db.env`).
- Deployed release source: `0645b2e4758d3cc6c197dd9dba9127e8de983d6c`.

## Built artifacts

Artifacts were built from exact clean main `26aa74dd...` on the Production host using Go 1.25 and the repository-pinned Caddy build path. Pinned Forwardproxy/Caddy reproducibility proof passed.

- `pvnaive` SHA256: `38ad79cce5f39eb3b3cf29f47d754148f9afeac0774c7e864d2c2a75edf35855`
- `pvnaive-telemetry-agent` SHA256: `21e584c5dc79a926a68e0e05e1c104c48e9eaa082785ba54637ba6eb40fa6b4a`
- accounting Caddy SHA256: `8097c552bffc8bc2cf2bbaa01cd88783b082e10d9e7ebb367f5c2b303548b997`
- Caddy version: `v2.11.2`; `http.handlers.forward_proxy` present.

## Backup / rollback gate

The first direct backup attempt intentionally failed closed because it inherited `pvnaive_app`, which cannot lock protected secret tables. No DB/runtime mutation had occurred at that point.

The canonical scheduled-backup path was then used with local PostgreSQL socket + `postgres` OS/DB role, matching the installed Production backup service:

- encrypted config backup: `/var/backups/pvnaive/scheduled/20260901T051036Z-4013973/config.tar.age`
- encrypted DB backup: `/var/backups/pvnaive/database/20260901T051036Z-4013985-w4IxYu/pvnaive.dump.age`
- backup checksum verification: PASS
- rollback snapshot: `/var/backups/pvnaive/task15-schema20-20260901T051038Z`

The rollback snapshot contains pre-deploy API, telemetry-agent, Caddy binary, Caddyfile, DB env, release marker, web target and DB release target with SHA256 verification for copied files.

## Migration / deployment

- Immutable DB tooling release promoted to `0020-5df21d237574`.
- Migration `0020_unique_ip_limit.up.sql` applied exactly once.
- Final DB schema: 20.
- `/etc/pvnaive/db.env`: `PVNAIVE_EXPECTED_SCHEMA_VERSION=20`.
- Migration record checksum: `5df21d237574b6ea8010f18ab346c24f1b41e40e0f8bda174594adc05639c385`.
- `plans.unique_ip_limit` and `service_terms.unique_ip_limit` both present.
- API and telemetry binaries atomically replaced.
- Panel web release selected at `/var/www/pvnaive-preview/releases/task15-26aa74dddfd2`.
- Candidate Caddy validated against the live Caddyfile before swap; accounting Caddy was atomically replaced and `caddy-naive` received one explicit controlled service restart. No Caddy config change was required.
- Release marker now records source `26aa74dddfd23535e45837f21531cf67ea2fd238`, base schema19, schema20.

## Postflight

- all four Production services active;
- local liveness/readiness: PASS;
- public SNI-correct panel request: HTTP 200;
- public SNI-correct API readiness request: HTTP 200;
- DB schema and expected schema: 20/20;
- Caddy config validation: PASS;
- pinned Forwardproxy module present;
- recent service logs: no panic/fatal/permission/schema-mismatch failures.

A local `https://127.0.0.1` panel probe failed TLS because it did not send the production SNI; this was a probe error, not an outage. It was immediately repeated correctly with `--resolve namir.softarg.ir:443:127.0.0.1` and returned HTTP 200.

## Result

Task15 simultaneous unique-IP limit is **DONE / Production / schema20**. Production was not used as a test database; PG18 concurrency/migration proof came from CI before merge. Rollback material remains retained.
