# PVNaive — Feature Matrix and Gap Analysis

Last updated: 2026-08-30

This file is the canonical **short-form** capability truth. The full 120-feature competitor matrix is:

`docs/PANEL_PARITY_MASTER_2026-08-30.md`

A route declaration, schema column/table, disabled UI control or old workstream claim does not count as an implemented feature. Production-facing claims require current evidence where practical.

## Current audited snapshots

- PVNaive: `main@a021aa4b62c35b775fb521d042b2f8e6dbde10b0`
- 3x-ui: `f727d04f6522bb94a8fb52e8352fdcafb51c11e1` / v3.7.0
- PasarGuard: `aebf7256927710329d380d67ce96224f287ae5f6` / v5.3.0
- Hiddify Manager: `dev@a99c811aa63fe908f1e06607b81f475b502ebf07`; stable v12.3.3
- OV-PvNetwork: `5b6a578bfe7733ebc67c08d9c431da6e32ac7ced` / public v1.0.0-rc1 snapshot

GPL/AGPL competitors are behavior/architecture/UX references only unless a later explicit compatibility review permits source reuse.

## Current PVNaive truth

Legend: `DONE`, `PARTIAL`, `BLOCKED`, `OPTIONAL`, `N/A`.

| Area | Status | Evidence / remaining gate |
|---|---|---|
| Owner authentication/session/CSRF/TOTP | PARTIAL | real auth exists; refresh-family reuse, commit-before-success, DB readiness, abuse controls and recovery-login policy remain open |
| Raw Naive Runtime credential lifecycle | DONE | import/create/update/rotate/disable/revoke, expected-SHA apply/rollback and privileged Runtime Agent are integrated |
| Customer CRUD / adopt / edit / suspend / resume / revoke | DONE | current customer handlers, service and UI are present |
| Plans | DONE | schema 11 + customer product service/UI |
| Groups / tags / notes | DONE | customer product model/store/API/UI |
| Renewal / immutable ServiceTerm | DONE | renewal service preserves history by new term rather than rewriting old usage |
| Next Plan / On Hold | PARTIAL | semantics exist; operator history/workflow needs final polish/proof |
| Add volume / set volume / unlimited | DONE | single-customer and database bulk operations exist |
| Expiry / no-expiry / creation validity / manual expiry / extend days | DONE | customer validity/service APIs + tests |
| First-successful-CONNECT validity | PARTIAL | trusted WS1 producer/core integrated and Production has active accounting; controlled first-transition Production proof still required |
| Search / filter / sort / pagination | DONE | server-side customer-product query path |
| Selectable columns / URL-persisted filters | BLOCKED | final table UX backlog |
| Bulk operations | DONE | preview + idempotent execute for supported lifecycle/product actions; Reset Usage is not included yet |
| Subscription machine endpoint `/sub/<token>` | DONE | stable machine endpoint |
| Human account page `/s/<token>` | DONE | stable HTML endpoint |
| Local QR | DONE | no third-party secret-bearing QR endpoint |
| Subscription reissue | DONE | explicit, separate from password rotation |
| Password rotation | DONE | explicit one-time secret delivery; separate from subscription token |
| Exact direct-Naive byte accounting | DONE | schema 9+ direct accounting, pinned forwardproxy instrumentation, telemetry agent; Production has complete accounting term projections and live event/session data |
| Restart-safe accounting | DONE | boot/session/sequence/cumulative semantics and idempotent ingest |
| Hard quota core | PARTIAL | shared reservation/settlement is integrated; controlled concurrent exhaustion/reload/restart/reconnect Production proof remains |
| Online / last online | PARTIAL | trusted session/presence evidence exists, but customer read models/UI are not consistently wired to it |
| Operator session list / kill | BLOCKED | accounting sessions are present; operator-facing session-control capability is not ready |
| Concurrent/IP limits | BLOCKED | requires real enforcement and race tests |
| HWID/device limit | BLOCKED | requires trustworthy identity PoC; never fabricate HWID |
| Per-user speed limit | BLOCKED | requires data-plane enforcement PoC; hide until real |
| Reseller / RBAC | PARTIAL | tenant/RLS and product foundations exist; full reseller CRUD/wallet/ledger/restriction UX/API not ready |
| Tenant isolation | DONE | WS2 RLS/trigger/actor-alias foundation; full route-wide IDOR matrix still remains security work |
| Audit | PARTIAL | audit events exist; customer history + Audit Explorer UI not ready |
| Notifications / Telegram | PARTIAL | schema/foundation exists; delivery engine, preferences/history/rules/Telegram not integrated into current main |
| Dashboard | PARTIAL | customer dashboard exists; real CPU/RAM/disk/network/history remains |
| Application/runtime/security logs | BLOCKED | useful implementation exists only on stale PR #16 and must be safely extracted |
| Diagnostics / Doctor | BLOCKED | useful PR #16 units require fresh integration/testing |
| Backup | PARTIAL | encrypted/manual backup and restore test foundations exist; scheduled product backup is not active on Production |
| Scheduled backup | BLOCKED | no PVNaive scheduled-backup timer observed during 2026-08-30 Production audit |
| OpenAPI / Swagger | BLOCKED | candidate exists on PR #16; not in current main |
| General API rate limiting | PARTIAL | auth/foundation only; PR #16 has broader candidate implementation |
| Webhooks | BLOCKED | only after stable event contracts |
| Multi-node/fleet | BLOCKED | standalone-first; PR #16 has model foundation only |
| Fresh one-line installer | BLOCKED | stage-specific install/upgrade scripts are not a generic clean Ubuntu installer |
| Generic upgrade/rollback | PARTIAL | strong stage/runtime rollback patterns exist; generic lifecycle incomplete |
| Uninstall / auto-update | BLOCKED | later lifecycle work |
| Karing compatibility | PARTIAL | machine/direct delivery exists; real current Karing OS/client matrix and refresh/reissue acceptance remain required |
| 400-concurrent capacity proof | BLOCKED | bounded rehearsals are not the required 50/100/200/400+ campaign |
| SBOM/SAST/secret/dependency scanning/release signing | BLOCKED | supply-chain gate not yet production-ready |

## Production truth captured 2026-08-30

Read-only audit found:

- API, Runtime Agent, Telemetry Agent and Caddy active;
- API loopback-only on `127.0.0.1:8080`;
- public panel and readiness HTTP 200;
- PostgreSQL schema version 11;
- six active customers/runtime credentials/service terms and six active subscription tokens;
- six direct-accounting term projections, all complete at audit time;
- live append-only direct accounting events and session records present;
- no PVNaive scheduled-backup timer observed;
- root filesystem 79% used at audit time;
- Production deployment marker files are stale/inconsistent with newer binary/web mtimes, so release provenance must be hardened.

No secret, password, raw Subscription token or encryption key is recorded in this file.

## Immediate required execution order

1. Merge truth/parity reconciliation after exact-head CI.
2. Reconcile useful PR #16 units without blind merge.
3. Legacy/adopted-account accounting baseline truth.
4. `/s` accounting/presence truth.
5. Manual Reset Usage.
6. Bulk Reset Usage.
7. Restart-safe periodic traffic reset.
8. Hard quota Production proof.
9. First successful CONNECT Production proof.
10. Sessions / disconnect / concurrency / IP limit / history.
11. HWID and speed-limit PoCs.
12. Reseller/RBAC/wallet/ledger/restrictions.
13. Customer history/audit, notifications, monitoring/logs/doctor, backup/restore, API/security, fleet, installer, client matrix, load/capacity, final bulk/UI/docs.

The target remains **correct, evidence-backed NaiveProxy operations**, not feature-count parity.
