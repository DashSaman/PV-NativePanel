# CONTINUE HERE — PVNaive

If a Chat/Agent session was interrupted, start here.

## Read in this order

1. `ops/evidence/S04-20260828T003759Z-db-health-release-promotion-pass.md` — newest live evidence; periodic DB health blocker is now repaired.
2. `ops/S04_LIVE_STATE.md` — authoritative current live state and safety invariants.
3. `ops/evidence/S04-20260828T002041Z-postflight-core-pass-health-release-blocked.md` — prior independent postflight that found the blocker plus TDD repair history.
4. `AGENT_HANDOFF.md` — broader project history and non-negotiable constraints.
5. `ops/DEPLOYMENT_PROGRESS.md` — official stage ledger.
6. Active implementation branch: `s04-auth`; open draft PR: `#2`.

## Current one-line state

`S00-S03=PASSED`; `S04-AUTH=IN PROGRESS`. The S04 API/auth/web Stage is installed and healthy on `testAmir5-3`. The only blocker from the first independent postflight has now been repaired: `/opt/pvnaive/db/current` selects schema2 immutable release `0002-84bb735877d5`, the real periodic DB health path verifies MFA secret-table denial, and live repair output returned `DB_TIMER_S04_AWARE=true`. API/Caddy/SSH/firewall remained unchanged and schema stayed 2.

**S04 is still NOT PASSED until a fresh independent postflight returns `S04_POSTFLIGHT=PASSED`. Do not bootstrap Owner yet.**

## Live repair evidence

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
NEXT=RERUN_S04_INDEPENDENT_POSTFLIGHT
```

Periodic health returned:

```text
DB_HEALTH_RESULT=success
DB_HEALTH_EXEC_STATUS=0
PVNAIVE_DB_HEALTH=OK
PVNAIVE_SCHEMA_VERSION=2
PVNAIVE_DB_USER=pvnaive_app
PVNAIVE_SECRET_DIRECT_SELECT=DENIED
PVNAIVE_MFA_DIRECT_SELECT=DENIED
```

API remained only on `127.0.0.1:8080`; liveness/readiness passed; Caddy SHA stayed exactly `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`; SSH remained active.

## Repository repair is green

- branch repair head: `20ed774d06969a3f4c301fd6072a4db83fcffcca`
- final CI run: `33130012929` — SUCCESS
- Go: SUCCESS
- Web: SUCCESS
- PostgreSQL18 regression suite: SUCCESS
- end-to-end S04 rehearsal: SUCCESS
- production bundle: SUCCESS

The live repair used the already-installed pinned S04 DB source plus the exact commit-pinned `promote-release.sh`, whose Git blob SHA matched `0f83469e8f7928d8dbc58d1984fb236552a97e29` before execution.

## Live S04 artifact already installed

- source commit: `11c54dc1faae99a1491c750b30db9faa44a0c3ae`
- CI run: `33128780602`
- artifact ID: `9669443464`
- archive: `PVNaive-S04-11c54dc1faae.tar.gz`
- SHA-256: `52acdde2bff6777abeb31c081c86e018d1e7f5f0cdb39f8eb4151efbba2820fc`

Never reuse the older `b4803e27...` bundle.

## Exact next action

Run a fresh **independent S04 postflight** on `testAmir5-3`. It must independently verify marker/schema/migration identity, selected schema2 DB release, MFA-aware periodic health, health service/timer state, API loopback-only + live/ready, encrypted rollback backup validity, Caddy SHA/active state, SSH active state, and `DB_TIMER_S04_AWARE=true`.

Only after that command itself returns `S04_POSTFLIGHT=PASSED` may the real Owner be bootstrapped. After Owner localhost login/session/logout passes, the later external exposure/Caddy gate can proceed; only after the external postflight should the ledger advance to `S04-AUTH=PASSED` and `S05-USERS=NEXT`.
