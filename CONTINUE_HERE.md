# CONTINUE HERE — PVNaive

Last updated: 2026-08-30

If a Chat/Agent session is interrupted, start here. The old S04/S05 instructions are superseded.

## First read

1. `OWNER_REQUIREMENTS.md`
2. `PROJECT_STATUS.md`
3. `HANDOFF.md`
4. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
5. `KNOWN_ISSUES.md`
6. `ROADMAP.md`
7. `AGENT_TASKS.md`
8. `WORKLOG.md`
9. newest `ops/evidence/*` before any Production mutation

## One-line current state

`main@a021aa4b62c35b775fb521d042b2f8e6dbde10b0` already contains customer/product management, `/sub` + `/s` + local QR, exact direct-Naive accounting, trusted first-CONNECT telemetry and hard-quota core. Production is running schema 11 with active API/Runtime/Telemetry/Caddy and live accounting data. The immediate task is **truth/CI reconciliation, not rebuilding those features**.

Always re-fetch latest `main`; the SHA above is only the audit starting point.

## Active branch / PR

- branch: `lead/parity-truth-2026-08-30`
- PR: `#27` — Lead production truth and panel parity reconciliation
- plan: `docs/superpowers/plans/2026-08-30-production-parity-reconciliation.md`

The PR is documentation/truth work only. It must not mutate Production.

## Current Production facts from read-only audit

- panel: `https://namir.softarg.ir/panel/` returned 200;
- public readiness returned 200;
- API is loopback-only on `127.0.0.1:8080`;
- `pvnaive-api`, Runtime Agent, Telemetry Agent and Caddy services are active;
- PostgreSQL schema is 11;
- six active users/runtime credentials/ServiceTerms and six active Subscription tokens exist;
- all six direct-accounting term projections reported complete at audit time;
- live accounting events/session rows are present;
- root filesystem was 79% used;
- backup files exist, but no PVNaive scheduled-backup timer was observed;
- deployment marker files lag the actual newer binary/web timestamps and are not sufficient release provenance.

No secret/token/password/key is recorded here.

## Current CI fact

PR #26's final bot commit recorded failed CI/WS1 runs, but the immediately preceding human commit and previous bot commit passed all three workflows. Do not call this a product regression without reproduction.

PR #27 was opened from exact current main to reproduce the baseline. During the current reconciliation, WS1 Exact Accounting has already succeeded on a refreshed PR head; CI and pinned-forwardproxy results must be checked on the final exact head before merge.

## Current confirmed blockers/gaps

### Before feature expansion

1. finish current truth/parity/status reconciliation;
2. get exact-head PR #27 CI green;
3. merge the reconciliation;
4. extract useful PR #16 work on a fresh branch, never blind merge.

### Next product work after PR #16 reconciliation

1. legacy/adopted Runtime accounting baseline truth;
2. `/s` exact accounting/presence projection;
3. Manual Reset Usage;
4. Bulk Reset Usage;
5. periodic reset scheduler/cursor/exactly-once execution;
6. hard-quota controlled Production proof;
7. first-successful-CONNECT controlled Production proof;
8. sessions/kill/concurrent/IP/history;
9. HWID/speed PoCs;
10. reseller/RBAC/wallet/ledger/restrictions;
11. remaining ordered Owner backlog.

## Do not claim these are done

- Manual/Bulk Reset Usage;
- periodic reset execution;
- session kill or limits;
- HWID/speed limit;
- full reseller wallet/ledger/restrictions;
- audit explorer/customer history;
- notifications/Telegram;
- real system monitoring/log UIs/doctor;
- scheduled backup on Production;
- OpenAPI on current main;
- multi-node/fleet;
- clean one-line installer;
- real Karing compatibility matrix;
- 400-concurrent capacity proof;
- supply-chain/release signing gates.

## Current P0 security defects — still open

- refresh-token reuse-family handling is bypassed by the `revoked_at IS NULL` authenticated lookup;
- generic HTTP middleware can write success before durable transaction commit result is known;
- readiness is not DB/schema-backed.

See `KNOWN_ISSUES.md` for exact evidence.

## Production mutation rule

Before any mutation:

1. fetch latest main + exact deployment evidence;
2. confirm current service/database/Caddy state;
3. DB backup;
4. config backup;
5. Caddy backup;
6. web backup;
7. binary backup;
8. explicit rollback plan;
9. validate/stage;
10. apply one bounded change;
11. postflight and rollback on failure.

Never print or commit raw secrets.
