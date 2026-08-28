# S04 live DB health release promotion — PASSED

Timestamp: `2026-08-28T00:37:59Z`
Host: `testAmir5-3`

## Result

The bounded live repair for the final S04 postflight blocker completed successfully.

```text
S04_DB_HEALTH_RELEASE_PROMOTION=PASSED
DB_RELEASE_OLD=/opt/pvnaive/db/releases/0001-7f66adefd8f0
DB_RELEASE_NEW=/opt/pvnaive/db/releases/0002-84bb735877d5
DB_CURRENT=/opt/pvnaive/db/releases/0002-84bb735877d5
DB_TIMER_S04_AWARE=true
SCHEMA_CHANGED=false
API_CHANGED=false
CADDY_CHANGED=false
SSH_CHANGED=false
FIREWALL_CHANGED=false
NEXT=RERUN_S04_INDEPENDENT_POSTFLIGHT
```

S04 is still **IN PROGRESS** until the independent postflight is rerun and returns `S04_POSTFLIGHT=PASSED`. Do not bootstrap the real Owner yet.

## Live preflight evidence

Before the repair:

- `/opt/pvnaive/db/current` selected `/opt/pvnaive/db/releases/0001-7f66adefd8f0`.
- PostgreSQL schema version was exactly `2`.
- migration 0002 checksum was exactly:
  `84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886`.
- Caddyfile SHA-256 was unchanged:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- API listener was only `127.0.0.1:8080`.
- live preflight passed.

## Source-integrity evidence

The repair reused the already-installed pinned S04 migration/DB scripts and downloaded only the promotion helper pinned to clean repair head:

`20ed774d06969a3f4c301fd6072a4db83fcffcca`

The downloaded helper Git blob SHA matched exactly:

`0f83469e8f7928d8dbc58d1984fb236552a97e29`

All migration checksum verification passed before promotion.

## Atomic immutable release promotion

Promotion output:

```text
PVNAIVE_DB_RELEASE_PROMOTION=PASSED
PVNAIVE_DB_RELEASE_ID=0002-84bb735877d5
PVNAIVE_DB_RELEASE_OLD=/opt/pvnaive/db/releases/0001-7f66adefd8f0
PVNAIVE_DB_RELEASE_NEW=/opt/pvnaive/db/releases/0002-84bb735877d5
DB_CURRENT_AFTER=/opt/pvnaive/db/releases/0002-84bb735877d5
IMMUTABLE_DB_RELEASE=PASSED
```

The old S03 immutable release remains present for rollback.

## Periodic DB health gate

The real systemd health path now selects the S04-aware schema2 release.

```text
DB_HEALTH_RESULT=success
DB_HEALTH_EXEC_STATUS=0
PVNAIVE_DB_HEALTH=OK
PVNAIVE_SCHEMA_VERSION=2
PVNAIVE_DB_USER=pvnaive_app
PVNAIVE_DB_SERVER_ADDRESS=127.0.0.1
PVNAIVE_DB_SERVER_PORT=5432
PVNAIVE_DB_CLIENT_ADDRESS=127.0.0.1
PVNAIVE_SECRET_DIRECT_SELECT=DENIED
PVNAIVE_MFA_DIRECT_SELECT=DENIED
DB_TIMER_S04_AWARE=true
```

This closes the exact blocker found by the prior independent postflight.

## Final infrastructure invariants

After promotion:

- API listener remained only `127.0.0.1:8080`.
- API liveness returned `{"service":"pvnaive-api","status":"ok"}`.
- API readiness returned `{"ready":true,"status":"ready"}`.
- Caddyfile SHA-256 remained exactly `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`.
- `caddy validate` completed successfully; its formatting warning is non-blocking and no Caddyfile mutation occurred.
- SSH unit remained `ssh.service` and active.
- schema remained exactly `2`.
- no schema/API/Caddy/SSH/firewall mutation was performed by the repair beyond the intended DB tooling symlink promotion.

## Repository repair verification

The promotion helper and Stage wiring had already passed the clean-head final CI run:

- repair head: `20ed774d06969a3f4c301fd6072a4db83fcffcca`
- CI run: `33130012929`
- Go: SUCCESS
- Web: SUCCESS
- PostgreSQL18 regression suite: SUCCESS
- end-to-end S04 rehearsal: SUCCESS
- production bundle: SUCCESS

## Exact next action

Rerun an **independent** S04 postflight. It must independently verify at minimum:

- marker + schema2 + exact migration identity;
- `/opt/pvnaive/db/current` -> `0002-84bb735877d5`;
- selected health script is S04/MFA-aware;
- health service/timer success;
- API active and loopback-only;
- liveness/readiness;
- valid encrypted rollback backup;
- Caddy SHA unchanged and Caddy active;
- SSH active;
- `DB_TIMER_S04_AWARE=true`.

Only if that independent postflight returns `S04_POSTFLIGHT=PASSED` may the real Owner bootstrap begin.
