# S04 runtime fix verification trigger

This commit exists to run the full CI suite with a non-bot actor after the one-shot runtime patch.

Runtime regressions under verification:

- API must derive its PostgreSQL DSN from `PVNAIVE_DB_*` and connect as `pvnaive_app` to `pvnaive`.
- S04 must atomically transition `/etc/pvnaive/db.env` expected schema from 1 to 2 and keep it consistent with rollback outcome.
- same-second encrypted backups must not collide.
- end-to-end rehearsal must use the same `PVNAIVE_DB_*` contract as the production systemd service.
