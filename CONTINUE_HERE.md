# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T004121Z-postflight-role-format-harness-false-negative.md` — newest postflight evidence; the latest stop was a read-only test-format bug, not a server privilege failure.
2. `ops/evidence/S04-20260828T003759Z-db-health-release-promotion-pass.md` — live DB health release promotion passed.
3. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
4. `ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md` — earlier postflight that found the periodic health release blocker.
5. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
6. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
7. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. The S04 API/auth/web Stage is installed and healthy on `testAmir5-3`. The periodic DB health blocker is repaired: `/opt/pvnaive/db/current` selects schema2 immutable release `0002-84bb735877d5` and `DB_TIMER_S04_AWARE=true` was verified live. API/Caddy/SSH/firewall remained unchanged and schema stayed 2.

A fresh independent postflight at `2026-08-28T00:41:21Z` reached the DB role check after passing marker, installed artifact integrity, schema2 and exact migration identity. It then stopped only because the harness expected PostgreSQL booleans as `t/f`; the server returned `true/false`. The observed role rows are the intended restricted privilege values. The command was read-only and reported `NO_CONFIGURATION_CHANGES_MADE=true`.

**S04 is still NOT PASSED until the corrected independent postflight itself returns `S04_POSTFLIGHT=PASSED`. Do not bootstrap Owner yet.**

## Live role values from the latest postflight

```text
pvnaive_app|true|false|false|false|false|false|false
pvnaive_owner|false|false|false|false|false|false|false
```

These mean `pvnaive_app` has LOGIN only and no listed elevated privilege; `pvnaive_owner` has no LOGIN and no listed elevated privilege.

## Live repair evidence already passed

At `2026-08-28T00:37:59Z`:

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
```

Periodic health returned schema2, `pvnaive_app`, signing-secret denial and MFA-secret denial. API remained only on `127.0.0.1:8080`; liveness/readiness passed; Caddy SHA remained `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; SSH remained active.

## Repository repair is green

- branch repair head: `20ed774d06969a3f4c301fd6072a4db83fcffcca`
- final CI run: `33130012929` — SUCCESS
- Go: SUCCESS
- Web: SUCCESS
- PostgreSQL18 regression suite: SUCCESS
- end-to-end S04 rehearsal: SUCCESS
- production bundle: SUCCESS

## Live S04 artifact installed

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse the older `b4803e27...` bundle.

## Exact next action

Rerun the independent S04 postflight with its role assertions corrected to `true/false`. It must then continue through selected schema2 DB release, MFA-aware periodic health, API loopback-only + live/ready, encrypted rollback backup, Caddy SHA/validation, SSH and required network invariants.

Only after that corrected command returns all of:

```text
S04_POSTFLIGHT_CORE=PASSED
DB_TIMER_S04_AWARE=true
S04_POSTFLIGHT=PASSED
NEXT=BOOTSTRAP_REAL_OWNER
```

may the real Owner bootstrap begin. After Owner localhost login/session/logout passes, the later Caddy exposure gate can proceed; only after external postflight should the official ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
