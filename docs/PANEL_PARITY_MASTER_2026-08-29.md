# PVNaive — Panel Parity Master / Gap Analysis

Last updated: 2026-08-29

این سند source-of-truth مقایسه‌ی محصول PVNaive با پنل‌های مرجع زیر است و برای تقسیم کار موازی Agentها استفاده می‌شود.

## Reference snapshots

- **PVNaive**: `DashSaman/PV-NativePanel`, `main` after S06 integration commit `22a1671f8779ee7b3e619d46de565d91da0373d3`.
- **3x-ui / Sanaei**: `MHSanaei/3x-ui`, latest stable `v3.7.0` (2026-08-24).
- **PasarGuard**: `PasarGuard/panel`, latest stable `v5.3.0` (2026-08-29).
- **Hiddify Manager**: `hiddify/Hiddify-Manager`, latest stable `v12.3.3` (2026-05-29), plus current official README/dev feature declarations.
- **PVNetwork OpenVPN**: `DashSaman/OV-PvNetwork`, public `1.0.0-rc1` + explicitly labeled production-derived feature matrix.

## Rules for this comparison

1. Feature parity does **not** mean protocol parity. PVNaive remains NaiveProxy-first.
2. A DB column, hidden endpoint, route stub, or UI placeholder does not count as implemented.
3. Usage/remaining/online/device/speed must never be fabricated.
4. Features that standard NaiveProxy cannot enforce reliably are `capability-gated`, not fake checkboxes.
5. GPL/AGPL reference projects are behavior/design references unless license compatibility is explicitly approved.
6. Production safety patterns from OV-PvNetwork may be reused where license and architecture permit.

## Executive result

PVNaive S06 is already strong in secure Owner auth, safe Runtime credential mutation, unified customer lifecycle, volume/validity management, Subscription token lifecycle, QR/direct-link delivery, encrypted DB backup foundation, and reload/rollback safety.

The largest remaining product gaps versus Sanaei/PasarGuard/Hiddify/OV-PvNetwork are:

1. **exact per-customer accounting + first-CONNECT + hard quota + presence/session evidence**;
2. **advanced customer/product ergonomics**: groups/tags/notes, plan presets, renewal/next-plan/on-hold, periodic reset, bulk preview/actions, richer filters/status chips;
3. **clean subscription/account-page delivery contract + compatibility lab + templates/i18n/announcements**;
4. **operations/observability**: server metrics, diagnostics/doctor, automatic backup schedule/restore drill, Telegram/notifications, OpenAPI, release hardening;
5. **RBAC/reseller product layer** beyond existing schema/auth foundation;
6. **multi-node/fleet** after standalone R1 is stable: node health, assignment, reconciler, drain/canary/failover/auto-deploy.

---

# 1. Customer and account lifecycle

| Feature | PVNaive main | 3x-ui | PasarGuard | Hiddify | OV-PvNetwork | Decision / Gap |
|---|---|---|---|---|---|---|
| Create customer/account | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Adopt/import existing account | ✅ | N/A model | N/A model | N/A model | integration pattern | Keep; important PVNaive differentiator |
| Edit customer/service | ✅ S06 | ✅ | ✅ | ✅ | ✅ | Keep; improve form ergonomics |
| Suspend / Resume | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Revoke / safe delete | ✅ | ✅ | ✅ | ✅ | ✅ | Keep audit + soft-delete default |
| Rename username | ✅ where safe | ✅ | ✅ | varied | ✅ patterns | Keep optimistic revision/runtime safety |
| Explicit password rotation | ✅ | ✅ | ✅ | ✅ | ✅ | Keep separate from Subscription reissue |
| Traffic quota / unlimited | ✅ configured, enforcement proof-gated | ✅ | ✅ | ✅ | ✅ | **P0 runtime enforcement gap** |
| Add volume | ✅ | operator pattern | ✅ | ✅ | ✅ | Keep |
| Replace/set total volume | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Expiry from creation | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Expiry from first successful connection | receiver/business logic exists; live producer proof missing | analogous | on-hold patterns | ✅ time policies | N/A | **P0 exact CONNECT producer** |
| Fixed/manual expiry | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Extend by N days | ✅ | ✅ | ✅ | ✅ | ✅ | Keep |
| Unlimited/no-expiry | ✅ policy-supported | ✅ | ✅ | ✅ | ✅ | Keep |
| Usage reset | ⛔ proof-gated | ✅ | ✅ | ✅ | ✅ | After exact accounting only |
| Periodic quota reset | ❌ | ✅ | ✅ daily/weekly/etc | ✅ | partial patterns | **P1 after accounting** |
| Renewal | partial/service term model | partial | ✅ | ✅ | partial | **P1** |
| Next plan | ❌ product UX | some patterns | ✅ strong | varied | N/A | **P1** |
| On-hold / pending-first-use UX | partial pending state | ⚠️ | ✅ | ✅ | N/A | **P1 status UX** |
| Service history / renewals history | partial data foundations | limited | ✅ patterns | partial | audit patterns | **P1** |
| Notes/comments | ❌ | partial | ✅/common model | ✅ patterns | operational metadata | **P1** |
| Groups/tags | ❌ | inbound/client organization | ✅ strong groups | admin/user grouping patterns | node assignment model | **P1** |
| Plan presets (30/50/80/100 GB etc.) | ❌ | templates/manual | group/profile patterns | packages/profiles | assignment/profile patterns | **P1 high-value PVNetwork UX** |

## Required additions

- PlanPreset entity + Owner CRUD + apply-on-create/apply-on-renew.
- Note/comment and tags/groups without coupling commercial state to Runtime credential.
- Renewal and `next_plan` semantics with immutable ServiceTerm history.
- `pending/on-hold` UX separated from account lifecycle.
- Periodic reset strategies only after accounting is exact.

---

# 2. Customer list, search, bulk and operator ergonomics

| Feature | PVNaive | 3x-ui | PasarGuard | Hiddify | Gap |
|---|---|---|---|---|---|
| Unified customer list | ✅ | ✅ | ✅ | ✅ | Keep |
| Search username | ✅ | ✅ | ✅ | ✅ | Keep |
| Search ID/token prefix safely | partial | partial | strong filters | partial | P1 |
| Status filter | ✅/basic | ✅ | ✅ + status/filter chips | ✅ | improve chips/multi-filter |
| Expiry filter/range | partial | ✅ | ✅ | ✅ | P1 |
| Quota/usage filter | proof-gated | ✅ | ✅ | ✅ | after accounting |
| Last-online filter | ❌ | ✅ | ✅ | ✅ | after presence proof |
| Sort username/date/expiry | ✅/basic | ✅ | ✅ | ✅ | complete + URL persistence |
| Sort usage/last-online | ❌ proof-gated | ✅ | ✅ | ✅ | after evidence |
| Pagination/page size | ✅ | ✅ | ✅ | ✅ | Keep |
| Select visible columns | ❌ | partial | strong table UX | partial | P1 |
| Bulk enable/disable | ❌/limited | ✅ | ✅ | admin tools | P1 |
| Bulk extend/add volume | ❌ | partial | ✅ | partial | P1 |
| Bulk revoke/delete | ❌ | ✅ | ✅ | partial | P1 with preview |
| Bulk plan/group assignment | ❌ | partial | ✅ | partial | P1 |
| Dry-run / preview before destructive bulk action | ❌ | not universal | partial | not universal | **PVNaive differentiator** |
| Command palette / keyboard navigation | ❌ | ❓ | ✅ v5.3.0 | ❓ | P2 |
| Theme palette/density | basic dark/light | dark/light | ✅ v5.3.0 | themed UI | P2 |
| Mobile action ergonomics | ✅ baseline | ✅ | ✅ v5.3.0 improvements | ✅ | continue QA |

---

# 3. Accounting, quota, first-use and presence

This is the highest-risk technical gap. No UI parity claim is allowed until the data path is proven.

| Feature | PVNaive main | References | Required |
|---|---|---|---|
| Exact upload/download per Runtime UUID | ❌ production proof | 3x-ui/PasarGuard use runtime/core counters; OV has OpenVPN-specific accounting | P0 |
| Cumulative counter protocol | ❌ main | restart-safe pattern required | P0 |
| boot/session/sequence identity | ❌ main | required to avoid double count | P0 |
| Append-only usage ledger | schema foundations exist but direct Naive production path incomplete | OV/accounting patterns | P0 |
| Restart/reconnect reconciliation | ❌ | OV desired-state/recovery patterns | P0 |
| First successful authenticated CONNECT event | receiver/business state exists; producer unproven | runtime instrumentation required | P0 |
| Used/remaining | intentionally unavailable | all reference panels display it | only after proof |
| Hard byte quota | intentionally unavailable | 3x-ui/PasarGuard/Hiddify enforce | only after exact counter proof |
| Depleted state automation | partial commercial state | references support | after hard quota |
| Online/offline | intentionally unavailable | 3x-ui/PasarGuard/Hiddify support | trusted session evidence |
| Last online | ❌ | supported by references | trusted session evidence |
| Current sessions | ❌ | supported by references | capability-gated |
| Session disconnect | ❌ | reference panels often expose controls | only if enforceable |
| Concurrent session limit | ❌ | 3x-ui/PasarGuard support patterns | capability-gated P2 |
| IP limit | ❌ | 3x-ui Fail2ban/IP limit | only with reliable identity |
| HWID/device limit | ❌ | PasarGuard | likely N/A unless a trustworthy client identity exists |
| Speed limit | ❌ | PasarGuard/other panels | capability-gated |

## Non-negotiable enforcement rule

If telemetry/accounting becomes unavailable, PVNaive must not continue claiming exact enforcement. The runtime contract must define fail-open/fail-closed behavior explicitly, expose `accounting_complete`, and never silently invent zero usage.

---

# 4. Subscription, account page, QR and client compatibility

| Feature | PVNaive main | 3x-ui | PasarGuard | Hiddify | Gap |
|---|---|---|---|---|---|
| Opaque revocable Subscription token | ✅ | ✅ | ✅ | ✅ | Keep |
| Read existing Subscription without rotation | ✅ S06 | ✅ | ✅ | ✅ | Keep |
| Explicit reissue | ✅ | ✅ patterns | ✅ | ✅ | Keep |
| Direct `naive+https://` | ✅ | N/A | N/A | N/A | Keep |
| QR local generation | ✅ | ✅ | ✅ | ✅ | Keep local/no third party |
| Branded account/status page | ✅ | custom templates | subscription UI | dedicated user pages | improve |
| Machine subscription endpoint independent of browser headers | ❌ on `main`; current handler branches on `Accept: text/html` | robust dedicated endpoints common | required P0 |
| Separate human account-page endpoint | ❌ on `main` | templates/user pages | required P0 |
| Multiple output formats | Naive-only by design | many formats | V2Ray/Clash/Meta | many configs | only add formats that make sense for Naive clients |
| Client compatibility matrix | partial/unit only | broad | broad | dedicated client | P0/P1 real lab |
| Karing smoke | needed per release | external client | external client | Hiddify app | P0 |
| V2Box / NekoBox / sing-box compatible delivery where Naive is supported | not fully proven | N/A | N/A | app ecosystem | lab-driven |
| Subscription page templates | hard-coded branded page | ✅ customizable | rich frontend | dedicated pages | P1 |
| Announcements/dynamic links | ❌ | partial | ✅ dynamic announcement URL in v5.3.0 | rich user page | P2 |
| i18n user page | partial/Farsi-focused | 13 languages | multi-language | multi-language | P1/P2 |

## Required endpoint contract

- `/sub/<opaque-token>`: always machine/client output; no `Accept` sniffing.
- `/s/<opaque-token>`: always human HTML account page.
- legacy `/api/v1/subscriptions/<token>` may remain machine-only for compatibility.
- View/copy is read-only; reissue is explicit mutation.

---

# 5. Auth, RBAC, reseller and audit

| Feature | PVNaive | References | Gap |
|---|---|---|---|
| Secure Owner login | ✅ | all | keep |
| TOTP MFA | ✅ | 3x-ui/Hiddify/etc | keep |
| Session revoke/list | ✅ | references | keep |
| Abuse/rate-limit hardening | partial | 3x-ui Fail2ban patterns | P0/P1 security |
| RBAC roles in auth/schema | foundation | PasarGuard strong RBAC; Hiddify multiple admin privileges | product UI/API missing |
| Reseller scoped customer visibility | DB/RLS foundation | PasarGuard/admin ownership patterns | P1 |
| Action-level permissions | partial | PasarGuard | P1 |
| Reseller plan restrictions | ❌ | commercial/operator patterns | P1 |
| Credit ledger UI/ops | schema foundation | reseller panels | P1 if business model enabled |
| Customer owner/admin attribution | partial | PasarGuard owner/admin filtering | P1 |
| Audit explorer UI | backend foundations | Remnawave/ops-grade panels | P1 |

---

# 6. Dashboard, monitoring, notification and diagnostics

| Feature | PVNaive | 3x-ui | PasarGuard | Hiddify | OV-PvNetwork | Gap |
|---|---|---|---|---|---|---|
| Customer counts/status KPIs | ✅ baseline | ✅ | ✅ | ✅ | ✅ | enhance |
| Real traffic graphs | ❌ proof-gated | ✅ | ✅ | ✅ | ✅ | after accounting |
| CPU/RAM/disk | ❌ product UI | ✅ | ✅ | status tooling | ✅ | P1 |
| Network RX/TX/rate | ❌ | ✅ | ✅ | monitoring | ✅ strong server-side rate semantics | P1 |
| Runtime health | partial | ✅ | ✅ | ✅ | ✅ | unify |
| Node health | future | ✅ multi-node | ✅ | architecture varies | ✅ | fleet phase |
| Telegram bot | ❌ | ✅ | ✅ | ✅ | monitoring/Telegram derived | P1 |
| Expiry/quota notifications | schema only | ✅ patterns | ✅ | ✅ | alerts | P1 |
| Runtime-down / backup-failed alerts | ❌ delivery | monitoring | monitoring | monitoring | ✅ patterns | P1 |
| Audit/log viewer | ❌ UI | logs | admin logs | logs | doctor/monitoring | P1 |
| Diagnostic bundle with redaction | ❌ | partial | partial | support tools | doctor patterns | P1 |
| Error boundary/server fallback UI | basic | partial | ✅ v5.3.0 | mature UI | N/A | P1 web resilience |

---

# 7. Install, update, backup, release and API

| Feature | PVNaive | 3x-ui | PasarGuard | Hiddify | OV-PvNetwork | Gap |
|---|---|---|---|---|---|---|
| Fresh one-line installer | not final generic | ✅ | ✅ | ✅ quick install | ✅ | P0 release |
| Non-interactive/cloud-init install | ❌ | ✅ v3.7.0 | scripts/container patterns | automation | partial | P1 |
| Version-pinned dependencies | ✅ design | release-based | release-based | update system | ✅ strong | keep |
| Upgrade lifecycle | staged releases | ✅ | ✅ | automatic update | ✅ transactional pattern | complete generic |
| Rollback | strong stage/runtime rollback | installer-level varied | varied | backup/update patterns | ✅ strong | generic product command |
| Scheduled backup | manual/encrypted foundation | backup support | DB options | ✅ every 6h | ✅ | P1 schedule |
| Restore drill | ✅ historical | varied | varied | backup restore | ✅ | automate recurring drill |
| `doctor` command | ❌ final CLI | x-ui CLI/system menu | CLI | support scripts | ✅ `ovpv doctor` | P1 |
| Log rotation/support bundle | ❌ | ✅ logs | ✅ | ✅ | ✅ | P1 |
| REST API | partial business API | ✅ | ✅ | APIs/admin | panel API | finish |
| Swagger/OpenAPI | ❌ public docs | ✅ | ✅ | API docs patterns | ❓ | P1 |
| Webhooks | ❌ | partial | partial | integrations | external hooks | P2 |
| SBOM/SAST/secret scan/signing | incomplete | mixed | mixed | mixed | lifecycle foundation | P0/P1 supply-chain |
| Load/capacity evidence | ❌ final | project-specific | project-specific | project-specific | project-specific | P0 release evidence |

---

# 8. Multi-node / fleet

PVNaive standalone R1 must not be delayed for fleet, but fleet remains a major parity gap.

Reference capabilities worth adopting from 3x-ui, PasarGuard and especially OV-PvNetwork:

- node registry and health;
- stable node UUID and desired/applied revision;
- customer-to-node assignment;
- reconciliation after drift/outage;
- node capacity/weight;
- failover policy;
- maintenance mode;
- drain existing sessions / stop new assignments;
- canary rollout;
- auto-node bootstrap/deployment;
- safe node edit/delete with assignment awareness;
- fleet-wide aggregated metrics;
- per-node version/runtime status;
- node backup/upgrade sequencing.

Status: **P2 after standalone accounting + customer lifecycle + ops release are green.**

---

# 9. Hiddify-specific capabilities: adopt selectively, not blindly

Hiddify has valuable capabilities that are not automatically PVNaive requirements:

| Hiddify capability | PVNaive decision |
|---|---|
| Multiple domains | Future/P2 if needed for delivery/rotation |
| Automatic Cloudflare connection / Auto CDN IP | Not a default Naive data-plane requirement |
| Smart proxy domestic/filtered routing | Client/routing product, not core PVNaive R1 |
| WARP integration | Optional future outbound policy |
| DoH endpoint | Optional; not customer-management blocker |
| Telegram proxy | Out of scope for Naive-first R1 |
| Dedicated client application | Separate PVNetwork-Client project; integrate later |
| 20+ protocol breadth | Explicitly out of scope; PVNaive is Naive-first |

---

# 10. Priority backlog distilled from parity

## P0 — block final standalone R1

1. Split machine Subscription and human Account Page endpoints; eliminate `Accept`-header ambiguity.
2. Exact trusted Naive accounting: Runtime UUID + boot/session/sequence + cumulative counters + append-only persistence.
3. Restart/reconnect/double-count proof and accounting completeness state.
4. Trusted first successful CONNECT activation.
5. Hard quota and depleted state only after exact accounting proof.
6. Real client compatibility lab: Karing first, then other Naive-capable clients actually used by PVNetwork.
7. Fresh install/upgrade/rollback/backup/restore/release artifact and supply-chain gates.
8. Auth abuse/rate-limit and final security review.
9. Production load/capacity evidence.

## P1 — operator parity / production quality

1. Plan presets, renewal, next plan, on-hold and service history.
2. Groups/tags/notes/admin attribution.
3. Advanced filters, status chips, selectable columns, URL-persisted filters.
4. Bulk dry-run + execute for enable/disable/extend/volume/revoke/delete/plan/group.
5. Reseller/RBAC product UI and scoped operations.
6. CPU/RAM/disk/network monitoring, Runtime metrics, audit explorer and diagnostics.
7. Telegram notification engine and meaningful event rules.
8. Scheduled encrypted backup + restore verification.
9. OpenAPI/Swagger and stable API contract.
10. Subscription templates, i18n and resilient error states.

## P2 — after standalone R1

1. Multi-node/fleet controller/agent + reconciler.
2. Failover/drain/canary/auto-node deployment.
3. Speed/concurrency/IP limits only when enforceable.
4. HWID/device limit only if a trustworthy identity exists.
5. Command palette, theme palette/density, advanced UI polish.
6. Optional domains/WARP/DoH/client-app integrations.

---

# 11. Parallelization map

Implementation is divided into four non-overlapping workstreams. Detailed copy-paste prompts live in `docs/PARALLEL_WORKSTREAM_PROMPTS_2026-08-29.md`.

| Lane | Ownership | Primary gap groups |
|---|---|---|
| WS1 | Runtime / Accounting / Enforcement | P0 accounting, first-CONNECT, quota, presence foundation |
| WS2 | Customer Product / Plans / Bulk / RBAC | customer ergonomics, plan/renewal/groups/bulk/reseller |
| WS3 | Subscription / Account Page / Client Compatibility | `/sub` vs `/s`, QR/templates/i18n/client lab |
| WS4 | Operations / Observability / Notifications / Release / Fleet foundation | metrics, doctor, backup schedule, Telegram, OpenAPI, installer/release, fleet later |

Each lane must use a separate branch and a separate report file. Agents must not edit another lane's report or silently mark shared tasks DONE.
