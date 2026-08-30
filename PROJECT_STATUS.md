# PVNaive — Canonical Project Status

Last updated: 2026-08-30

This file describes current repository + Production truth. Historical S04/S05 snapshots and stale PR branches must not override it.

## Product / architecture invariants

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. Keep these boundaries separate:

1. commercial customer/service state (`User`, immutable `ServiceTerm`, plans/groups/tags),
2. Runtime Naive credentials/secrets,
3. `/sub/<opaque-token>` machine delivery and `/s/<opaque-token>` human account page,
4. exact direct-Naive accounting/session telemetry,
5. privileged Runtime mutation through the narrow local Runtime Agent.

Do not rotate Runtime credentials or Subscription tokens when merely editing quota/expiry or viewing Subscription/QR.

## Repository state

- Repository: `DashSaman/PV-NativePanel`
- Product: **PVNaive**
- Default branch: `main`
- Current implementation / Production release: `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`
- Current schema head: `0011_customer_product_management` / schema 11
- Parity reference: `docs/PANEL_PARITY_MASTER_2026-08-30.md`
- Master Tasks #1-#4: **DONE**
- Next task: **#5 Legacy/adopted accounting baseline truth**

## Current Production truth — verified 2026-08-30

No secret-bearing values were printed.

- host: `testAmir5-3`
- public panel: `https://namir.softarg.ir/panel/`
- `pvnaive-api.service`: active, observed restart count 0
- `pvnaive-runtime-agent.service`: active, observed restart count 0
- `pvnaive-telemetry-agent.service`: active, observed restart count 0
- `caddy-naive.service`: active, observed restart count 0
- PostgreSQL: active
- API listener: loopback `127.0.0.1:8080`
- local readiness: healthy
- Runtime health: healthy
- Telemetry accounting socket health: healthy
- PostgreSQL schema: **11**
- six active users
- six active Naive Runtime credentials
- six active ServiceTerms
- six active direct Subscription tokens
- backup and restore-drill timers: active/enabled
- public panel HTTP 200
- public readiness HTTP 200
- `pvnaive doctor`: 14 PASS / 1 disk WARN / 0 FAIL
- real systemd restore drill: schema/ownership/ACL/signing-key checks PASS
- bounded loopback rehearsal: 100/100 success, 0 failures; this is not a capacity proof
- Caddy config SHA/PID/restart count unchanged through WS4 deployment
- `/opt/pvnaive/release/CURRENT` and legacy deployment marker point to `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb`
- Caddy serves the current panel release from `/var/www/pvnaive-preview/current`
- `/opt/pvnaive/web/current` and `/opt/pvnaive/db/current` point to the same final release generation
- root filesystem is around 79% used and intentionally remains a Doctor warning

Fresh encrypted config/database snapshots and release rollback directories were created before each Production mutation.

## Already integrated on main

### Runtime / customer / delivery

- AES-GCM Runtime secret envelope and stable Runtime UUID identity;
- safe Runtime Agent validate → backup → apply → verify → rollback path;
- customer create/adopt/edit/suspend/resume/revoke-safe-delete;
- quota/unlimited/add/set volume;
- creation/first-CONNECT/manual validity, no-expiry and extend-days;
- plans/groups/tags/notes/renewal/new immutable ServiceTerm;
- search/filter/sort/pagination and supported idempotent bulk operations;
- `/sub` machine endpoint, `/s` human account page, local QR;
- Subscription reissue and password rotation remain separate explicit actions.

### Exact accounting core

- successful authenticated Naive CONNECT accounting boundary;
- dedicated Telemetry Agent/accounting Unix socket;
- append-only idempotent boot/session/sequence/cumulative events;
- duplicate/conflict/gap/counter-regression handling;
- restart-safe ServiceTerm-isolated usage projection;
- trusted first-successful-CONNECT activation producer;
- session/presence projection;
- finite-quota reservation/settlement core.

Do not rewrite this core. Remaining P0 accounting work starts with legacy/adopted baseline truth.

### Operations / observability / release foundations — Task #4 DONE

Task #4 manually reconciled useful stale PR #16 ideas against the newer code and deployed them to Production:

- Linux CPU/RAM/disk/load/uptime/network metrics with server-side counter-delta rates;
- structured redacted logging;
- request IDs, bounded per-IP rate limiting and trusted-proxy handling;
- ready-route OpenAPI endpoint;
- `/api/v1/system/status` with API/DB/Runtime/Telemetry dependencies;
- `pvnaive doctor` and redacted diagnostic support bundle;
- encrypted scheduled backup + isolated restore drill + active systemd timers;
- release builder with dynamic schema discovery, checksums/basic SBOM/source provenance;
- same-schema guarded deploy + telemetry-aware rollback;
- Production-layout-aware web/DB symlink handling and deployment markers;
- bounded loopback control-plane rehearsal;
- notification retry/dedupe/redaction and secure Telegram transport foundations;
- standalone-safe fleet model/drift/delete-guard foundation;
- live owner System Dashboard;
- safe React ErrorBoundary.

Final WS4 evidence:

- PR #30 exact head `09b085a877e52fa02c095799359b6b9e89bb3492`: CI #1065, Exact Accounting #181, Pinned Forwardproxy #165 — SUCCESS;
- merge `c717d162a7e9b2e31fb5822b6b16c27ad048cbbd` deployed with fresh encrypted backups and rollback state;
- postflight exposed Doctor key-mode and restore-validation false negatives while customer service/accounting remained healthy;
- PR #31 exact head `b740352012fd9646c25d4c70c83f64f2f86ce029`: CI #1070, Exact Accounting #185, Pinned Forwardproxy #169 — SUCCESS;
- final hotfix merge `e9cce65d3fe8d82100b6bbb7e1231d07dc997edb` rebuilt, backed up, deployed and independently postflight-verified.

## Task #5 target — legacy/adopted accounting baseline truth

The six existing Production accounts and all future adopted accounts must not be assigned fabricated historical usage.

Required rules:

- numeric historical baseline only when authoritative pre-adoption usage can be proven;
- otherwise explicit Unknown/unavailable state, never implicit zero;
- exact post-adoption direct Naive usage stays separate and trustworthy;
- baseline + direct usage may be combined only when epoch/provenance proves the periods do not overlap;
- adoption must preserve enough provenance and boundary information to prevent double-counting;
- reads, quota/expiry edits, QR, Subscription view/reissue and password rotation cannot silently rewrite baseline history;
- existing Production users are classified from real server/database evidence only.

Task #5 requires RED→GREEN unit/integration/database tests, exact-head CI/Exact Accounting/Pinned Forwardproxy, and a backed-up Production rollout if schema/runtime semantics change.

## Important incomplete work after Task #5

- complete `/s` accounting/presence projection;
- Manual Reset Usage / Bulk Reset Usage / periodic restart-safe reset scheduler;
- controlled Production hard-quota race/restart proof;
- controlled first-successful-CONNECT Production proof;
- operator session list/kill/concurrent/unique-IP limits/history;
- trustworthy HWID/device identity PoC;
- real per-user speed-limit PoC/enforcement;
- full reseller CRUD/wallet/ledger/restrictions;
- customer history + Audit Explorer;
- notification preferences/history/rule builder and configured Telegram workflow;
- historical metrics/log UIs beyond the live System Dashboard;
- fleet controller/multi-node operations/failover/smart selection;
- clean-server installer/upgrade/uninstall lifecycle;
- Karing multi-OS compatibility campaign;
- 50/100/200/400+ capacity campaign;
- supply-chain SAST/dependency scan/signing policy beyond current checksum/SBOM foundation.

## Confirmed P0 security defects still open

1. refresh-token reuse-family path is blocked too early by `BeginAuthenticated`'s `revoked_at IS NULL` selection;
2. generic authenticated handlers can write success before final transaction commit is known and commit errors are ignored;
3. `/health/ready` still lacks the required bounded DB/schema readiness probe.

Do not remove these until their Owner-mandated security stage has tests and Production-safe proof.

## Current execution order

1. Task #5 — legacy/adopted accounting baseline truth;
2. Task #6 — `/s` accounting/presence completion;
3. Tasks #7-#9 — reset semantics;
4. Tasks #10-#11 — hard-quota and first-CONNECT controlled Production proofs;
5. remaining session/reseller/security/fleet/installer/client/capacity/release gates in `ROADMAP.md` order.

## Read next

1. `OWNER_REQUIREMENTS.md`
2. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
3. `HANDOFF.md`
4. `KNOWN_ISSUES.md`
5. `ROADMAP.md`
6. `AGENT_TASKS.md`
7. `WORKLOG.md`
8. before any Production mutation, re-check live state and fresh backup/rollback evidence.
