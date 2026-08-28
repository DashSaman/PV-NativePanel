# S04-AUTH Live State — testAmir5-3

Last updated: 2026-08-28 00:43 UTC after final corrected independent localhost postflight PASSED

> Authoritative fast continuation file. Read `CONTINUE_HERE.md`, newest S04 evidence, `AGENT_HANDOFF.md`, and `ops/DEPLOYMENT_PROGRESS.md` before acting on the live host.

## Current state

- `S00-S03=PASSED`.
- `S04-AUTH=IN PROGRESS`, not official PASSED yet because public Caddy exposure/external postflight have not run.
- S04 API/auth/web localhost deployment is installed and healthy on `testAmir5-3`.
- Periodic DB health release drift was repaired live; `/opt/pvnaive/db/current` selects schema2 immutable release `0002-84bb735877d5`.
- Final corrected independent read-only postflight at `2026-08-28T00:43:44Z` returned:
  - `S04_POSTFLIGHT_CORE=PASSED`
  - `DATABASE_ROLE_CONTRACT=PASSED`
  - `DB_TIMER_S04_AWARE=true`
  - `API_POSTFLIGHT=PASSED`
  - `ROLLBACK_BACKUP=PASSED`
  - `INFRASTRUCTURE_POSTFLIGHT=PASSED`
  - `S04_POSTFLIGHT=PASSED`
  - `NO_CONFIGURATION_CHANGES_MADE=true`
  - `NEXT=BOOTSTRAP_REAL_OWNER`
- The next permitted action is the one-time real Owner bootstrap.
- Do not expose `/panel/` or `/api/` through Caddy until Owner creation and localhost real-auth verification pass.

## Final independently verified live facts

- PostgreSQL 18 schema: `2`.
- Migration 0002 SHA-256: `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- DB role contract:
  - `pvnaive_app|1|0|0|0|0|0|0`
  - `pvnaive_owner|0|0|0|0|0|0|0`
- `/etc/pvnaive/db.env` expects schema `2`.
- S04 marker: `/opt/pvnaive/S04_AUTH.json`.
- Auth release: `/opt/pvnaive/auth/releases/20260828T001418Z`.
- Web release: `/opt/pvnaive/web/releases/20260828T001418Z`.
- DB current release: `/opt/pvnaive/db/releases/0002-84bb735877d5`.
- Old S03 immutable DB release preserved: `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- Periodic DB health returns schema2, `pvnaive_app`, signing secret denial and MFA secret-table denial.
- `pvnaive-api.service` enabled/active as `pvnaive:pvnaive`, `NRestarts=0`.
- API listener only `127.0.0.1:8080`.
- API live: `{"service":"pvnaive-api","status":"ok"}`.
- API ready: `{"ready":true,"status":"ready"}`.
- Encrypted rollback backup: `/var/backups/pvnaive/database/20260828T001418Z-96157-k53f6h/pvnaive.dump.age`; checksum and decrypt/archive parse passed; no plaintext dump/sql sibling detected.
- PostgreSQL listeners only `127.0.0.1:5432` and `[::1]:5432`.
- Caddy active; Caddyfile SHA unchanged: `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; `caddy validate` passed. Existing formatting warning is non-blocking.
- SSH active as `ssh.service`.

Newest evidence: `ops/evidence/S04-20260828T004344Z-final-independent-postflight-pass.md`.

## Installed artifact / repo repair provenance

- Installed API/auth/web artifact source: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`.
- Installed artifact CI: `33128780602`.
- Archive SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`.
- DB release repair branch head verified by final CI: `20ed774d06969a3f4c301fd6072a4db83fcffcca`.
- Repair CI `33130012929`: Go/Web/PostgreSQL18/rehearsal/bundle all SUCCESS.
- Never reuse old `b4803e27...` bundle.

## Owner bootstrap contract

Use installed script:

`/opt/pvnaive/auth/current/scripts/auth/bootstrap-owner.sh`

It is root-only, exact-host, schema-v2, TTY-only, and refuses when any Owner already exists. It prompts for email, display name, password and confirmation on `/dev/tty`; password is hidden, never placed in argv/env/log, and is hashed using `/opt/pvnaive/bin/pvnaive-password`. Expected success output:

```text
PVNAIVE_OWNER_BOOTSTRAP=PASSED
OWNER_EMAIL=<normalized-email>
```

After Owner creation, independently verify one active Owner (without exposing hash), then run localhost real login/session/me/CSRF logout/revocation. Only after this passes may Caddy exposure be attempted. Only after controlled Caddy reload plus external postflight may `S04-AUTH` be marked PASSED and `S05-USERS` become NEXT.
