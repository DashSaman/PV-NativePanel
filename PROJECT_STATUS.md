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
- Current main before WS4 merge: `1ced722f9ee46fc4fb1a005ae8653f52339786b0`
- Active branch: `lead/ws4-safe-extract-2026-08-30`
- Active PR: `#29` — WS4 safe extraction of useful stale PR #16 units
- Current schema head: `0011_customer_product_management` / schema 11
- Parity reference: `docs/PANEL_PARITY_MASTER_2026-08-30.md`

Tasks #1-#3 truth/parity reconciliation are merged. Task #4 is code-complete on PR #29 but is **not Production-complete until PR merge + backed-up Production deployment + smoke verification**.

## Current Production truth — read-only audit 2026-08-30

No secret-bearing values were printed.

- host: `testAmir5-3`
- public panel: `https://namir.softarg.ir/panel/`
- `pvnaive-api.service`: active
- `pvnaive-runtime-agent.service`: active
- `pvnaive-telemetry-agent.service`: active
- `caddy-naive.service`: active
- observed restart counters: 0
- API listener: loopback `127.0.0.1:8080`
- local readiness: healthy
- Runtime health: healthy
- Telemetry accounting socket health: healthy
- PostgreSQL schema: **11**
- six active users
- six active Naive Runtime credentials
- six ServiceTerms
- six active direct Subscription tokens
- direct accounting events/sessions are live
- root filesystem observed at about 79% used
- backup age recipient/private key exist with restricted permissions
- Caddy serves the panel from `/var/www/pvnaive-preview/current`
- `/opt/pvnaive/web/current` and `/opt/pvnaive/db/current` also exist as release symlinks
- scheduled PVNaive backup/restore timers were not active at the pre-WS4 audit

### Production provenance gap before WS4 deployment

Legacy `/opt/pvnaive/DEPLOYED_COMMIT` and `/opt/pvnaive/DEPLOYED_WEB_RELEASE` lag the newest running binary/web state. WS4 release tooling now backs up, updates and rollback-restores these markers together with `/opt/pvnaive/release/CURRENT`, but Production remains unchanged until the final PR #29 merge is deployed.

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

Do not rewrite this core. Remaining P0 work starts with legacy/adopted baseline truth.

## Task #4 — WS4 safe extraction on PR #29

PR #16 was **not** merged/cherry-picked wholesale. Useful units were manually reimplemented/reconciled against current main with RED→GREEN tests.

Implemented on `lead/ws4-safe-extract-2026-08-30`:

- Linux CPU/RAM/disk/load/uptime/network metrics with server-side counter-delta rates;
- structured application logging + sensitive-key/text redaction;
- request IDs, bounded per-IP rate limiting and trusted-proxy handling;
- ready-route OpenAPI endpoint;
- `/api/v1/system/status` with API/DB/Runtime/Telemetry dependencies;
- `pvnaive doctor` and redacted diagnostic support bundle;
- encrypted scheduled backup + isolated restore-drill scripts and systemd timers;
- release builder with dynamic schema discovery, checksums/SBOM/source provenance;
- same-schema guarded deploy + telemetry-aware rollback;
- Production-layout-aware dual web symlink handling for `/opt/pvnaive/web/current` and `/var/www/pvnaive-preview/current`;
- legacy/new deployment marker backup/update/restore;
- bounded loopback control-plane load rehearsal (explicitly not a capacity ceiling);
- notification retry/dedupe/redaction foundation;
- Telegram transport foundation without token leakage;
- fleet model/drift/delete-guard foundation only — no fake multi-node readiness;
- live owner System Dashboard using real server metrics/dependencies;
- safe React ErrorBoundary that never renders raw exceptions.

The notification/fleet packages are foundations only and are not wired as fake completed product surfaces.

### Exact-head verification before final docs

Unit #7 implementation head `6a61512449884432477b1c107de6ac80e9e0d69c` passed:

- CI #1060 — SUCCESS, including Web tests/build, Go vet/tests, PostgreSQL gates, full S04R rehearsal and production bundle;
- WS1 Exact Accounting #176 — SUCCESS;
- WS1 Pinned Forwardproxy #160 — SUCCESS.

After this documentation update, all required workflows must pass again on the **final exact PR head** before merge.

## Important incomplete work

- legacy/adopted accounting baseline truth;
- complete `/s` accounting/presence projection;
- Manual Reset Usage / Bulk Reset Usage / periodic restart-safe reset scheduler;
- controlled Production hard-quota race/restart proof;
- controlled first-successful-CONNECT Production proof;
- operator session list/kill/concurrent/unique-IP limits/history;
- trustworthy HWID/device identity PoC;
- real per-user speed-limit PoC/enforcement;
- full reseller CRUD/wallet/ledger/restrictions;
- customer history + Audit Explorer;
- notification preferences/history/rule builder and actual configured Telegram product workflow;
- historical metrics/log UIs beyond the live System Dashboard;
- fleet controller/multi-node operations/failover/smart selection;
- clean-server installer/upgrade/uninstall lifecycle;
- Karing multi-OS compatibility campaign;
- 50/100/200/400+ capacity campaign;
- supply-chain SAST/dependency scan/signing policy beyond the current release checksum/SBOM foundation.

## Confirmed P0 security defects still open

1. refresh-token reuse-family path is blocked too early by `BeginAuthenticated`'s `revoked_at IS NULL` selection;
2. generic authenticated handlers can write success before final transaction commit is known and commit errors are ignored;
3. `/health/ready` still lacks the required bounded DB/schema readiness probe.

Do not remove these until their Owner-mandated security stage has tests and Production-safe proof.

## Current execution order

1. finish Task #4 final docs/exact-head CI/review;
2. merge PR #29;
3. take fresh Production backups, deploy exact merged commit using the tested release tooling, verify API/Runtime/Telemetry/panel/Caddy invariants/timers/markers and rollback readiness;
4. then Task #5: legacy/adopted accounting baseline truth;
5. `/s` accounting/presence;
6. reset semantics;
7. hard-quota and first-CONNECT controlled Production proofs;
8. remaining session/reseller/security/fleet/installer/client/capacity/release gates in `ROADMAP.md` order.

## Read next

1. `OWNER_REQUIREMENTS.md`
2. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
3. `HANDOFF.md`
4. `KNOWN_ISSUES.md`
5. `ROADMAP.md`
6. `AGENT_TASKS.md`
7. `WORKLOG.md`
8. before any Production mutation, re-check live state and fresh backup/rollback evidence.
