# PVNaive — Canonical Project Status

Last updated: 2026-08-30

This file describes current repository + Production truth. Historical S04/S05 branch snapshots must not override this file.

## Product

PVNaive is PVNETWORK's standalone-first management plane for standard NaiveProxy. The current architecture deliberately separates:

1. commercial customer/service state (`User`, immutable `ServiceTerm`, plan/group/tag metadata),
2. Runtime Naive credential identity and secrets,
3. opaque Subscription/account-page delivery,
4. exact direct-Naive accounting/session telemetry,
5. privileged Runtime mutation through a narrow local agent.

Do not collapse these boundaries for UI convenience.

## Repository state

- Repository: `DashSaman/PV-NativePanel`
- Default branch: `main`
- Audited main: `a021aa4b62c35b775fb521d042b2f8e6dbde10b0`
- Current Lead reconciliation branch: `lead/parity-truth-2026-08-30`
- Current Lead PR: `#27` — truth/parity reconciliation
- Current schema head in main: `0011_customer_product_management`
- Full current parity matrix: `docs/PANEL_PARITY_MASTER_2026-08-30.md`

## Current Production truth — read-only audit 2026-08-30

Host/domain state was inspected without printing secrets.

- `pvnaive-api.service`: active
- `pvnaive-runtime-agent.service`: active
- `pvnaive-telemetry-agent.service`: active
- `caddy-naive.service`: active
- service restart counters observed at 0 for the current service lifetimes
- API listener: loopback `127.0.0.1:8080`
- local readiness: HTTP 200
- public `https://namir.softarg.ir/panel/`: HTTP 200
- public readiness: HTTP 200
- PostgreSQL database present, schema version **11**
- six active users
- six active Runtime credentials
- six ServiceTerms
- six active direct Subscription tokens
- six direct-accounting term projections; all six reported accounting complete at audit time
- live direct-accounting event/session data is present
- legacy `usage_ledger` is not the active direct-Naive billing path
- Caddy accounting socket and Runtime Agent socket permissions remain separated
- root filesystem was **79% used** at audit time
- backup files exist under the PVNaive backup root, but no PVNaive scheduled-backup systemd timer was observed

### Production provenance warning

`/opt/pvnaive/DEPLOYED_COMMIT` and the web release marker do not reliably identify the newest running binary/web state: marker SHAs/timestamps lag newer binary/web mtimes. Current behavior is healthy, but deployment provenance is not trustworthy enough for a Release Candidate. This must be fixed before final release signing/provenance work.

## Implemented and integrated capability

### Runtime credential management

- AES-GCM Runtime secret envelope and fingerprinting;
- stable Runtime credential UUIDs;
- import/create/update/password rotation/enable-disable/revoke lifecycle;
- fixed-capability Unix-socket Runtime Agent;
- expected-SHA Caddy mutation;
- validate → exact backup → apply → reload → postflight → rollback safety;
- desired/applied Runtime revision saga and reconciliation-required failure state;
- no arbitrary root shell/path/service API.

### Customer product management

Current main contains real customer/product implementations rather than route-only declarations:

- customer create and Runtime adoption;
- customer edit/service update;
- suspend/resume/revoke-safe-delete;
- quota and unlimited quota;
- add volume and set total volume;
- validity from creation, first successful connection and manual expiry;
- no-expiry and extend-days semantics;
- plan presets;
- renewal with new/immutable ServiceTerm semantics;
- Next Plan / On Hold foundations;
- groups, tags and notes;
- server-side search/filter/sort/pagination;
- bulk preview + idempotent execute for supported product/lifecycle actions;
- reseller/tenant-safe product data foundations.

### Subscription / customer delivery

- `/sub/<opaque-token>` is the machine/client endpoint;
- `/s/<opaque-token>` is the human Account Page;
- legacy API compatibility path remains machine-oriented;
- local QR generation;
- read-only Subscription view/copy does not rotate password/token;
- Subscription reissue and password rotation are separate explicit mutations;
- account/subscription paths are stable and token material is not listed in customer rows.

### Exact accounting / hard-quota core

The old `exact_accounting_not_proven` project status is obsolete.

Merged WS1 plus Production evidence provides:

- pinned forwardproxy instrumentation at the successful authenticated Naive CONNECT write boundary;
- Runtime credential UUID billing identity;
- dedicated telemetry Unix socket;
- append-only/idempotent event ingest;
- boot/session/sequence/cumulative counter semantics;
- duplicate/conflict/gap/counter-regression handling;
- restart-safe accounting projection;
- ServiceTerm-isolated usage;
- trusted first-CONNECT activation producer;
- session/presence projection;
- shared finite-quota reservation/settlement core.

What remains is not another accounting rewrite. Remaining P0 work is legacy/adopted baseline truth, UI/read-model completion, reset semantics, and controlled Production acceptance tests for hard quota and first-CONNECT race/restart behavior.

## Important features that are NOT complete

Do not infer completion from `Routes` or schema tables.

- Manual Reset Usage: not implemented as a ready customer capability.
- Bulk Reset Usage: not implemented.
- Periodic reset execution: plan model exists; restart-safe scheduler/cursor/exactly-once execution does not.
- Operator-facing active session list / kill session: not ready.
- Concurrent session/IP limits: not implemented/enforced.
- trustworthy HWID limit: no proven identity source; no fake HWID is allowed.
- per-user speed limit: no proven enforcement path; no fake control is allowed.
- reseller CRUD/wallet/ledger/plan restrictions: foundations exist, product workflow incomplete.
- customer history + Audit Explorer UI: incomplete.
- notification engine/Telegram/preferences/history/rule UI: incomplete in current main.
- real CPU/RAM/disk/network historical dashboard/log UI/doctor/support bundle: not integrated into current main; useful candidates exist in stale PR #16.
- scheduled encrypted backup: not active on Production.
- OpenAPI/Swagger: not in current main; PR #16 candidate.
- multi-node/fleet/failover/smart node: future after standalone correctness.
- generic clean-server one-line installer/upgrade/uninstall lifecycle: incomplete.
- real Karing client compatibility matrix: pending.
- 50/100/200/400+ concurrent capacity campaign: pending.
- SBOM/SAST/secret/dependency scanning/release signing/provenance: pending.

## Current confirmed P0 security defects

These were re-checked against current main on 2026-08-30 and remain open:

1. **Refresh-token reuse-family path** — `refresh` calls `BeginAuthenticated`, while `BeginAuthenticated` requires `s.revoked_at IS NULL`; a reused already-rotated token can fail before `auth_rotate_session` reaches its intended reuse-family handling.
2. **Commit-before-success integrity** — authenticated middleware executes the handler and may write the HTTP response before calling `bound.Tx.Commit()`, and the commit result is currently ignored for generic authenticated handlers.
3. **DB-backed readiness** — `/health/ready` currently checks configured auth/MFA dependencies but does not perform the required bounded ongoing DB/schema readiness probe.

Fix these in the Owner-mandated security stage after the earlier accounting/session/reseller/ops sequence, unless a technical dependency requires an earlier minimal fix.

## CI state

PR #26's final bot-authored head recorded failed CI/WS1 workflow runs, but its immediately preceding human commit and prior bot commit passed all three workflows. A clean Lead PR #27 was opened from exact current main to reproduce the baseline rather than guessing.

At the time this status was written:

- PR #27 WS1 Exact Accounting run for the refreshed parity head had completed successfully;
- CI and pinned-forwardproxy runs were still in progress and must complete before the reconciliation head is called green;
- after the final documentation head, CI must be triggered again for that exact SHA.

## Current execution order

Follow the Owner master prompt without skipping numbered correctness gates:

1. finish audit + current competitor parity + documentation truth;
2. safely reconcile useful PR #16 units;
3. legacy accounting baseline;
4. `/s` accounting/presence;
5. Manual Reset Usage, Bulk Reset Usage, periodic resets;
6. hard quota Production proof;
7. first successful CONNECT Production proof;
8. sessions/kill/concurrent/IP/history;
9. HWID/speed PoCs;
10. reseller/RBAC/wallet/ledger/restrictions;
11. history/audit, notifications/Telegram, monitoring/logs/doctor, backup/restore, API/security;
12. fleet/multi-node;
13. installer/upgrade/rollback;
14. client compatibility;
15. load/capacity;
16. final bulk/search/UI/docs/clean-install/Production smoke/RC.

## Read next

1. `OWNER_REQUIREMENTS.md`
2. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
3. `HANDOFF.md`
4. `KNOWN_ISSUES.md`
5. `ROADMAP.md`
6. `AGENT_TASKS.md`
7. `WORKLOG.md`
8. before any Production mutation: current `ops/evidence/*`, live service state, current backups and rollback plan.
