# PVNaive

<p align="center">
  <img src="web/public/private-network.webp" alt="Private Network" width="120" />
</p>

**PVNaive** is a production-grade management panel for **NaiveProxy**, built as a single Go binary with an embedded Caddy forward-proxy, PostgreSQL 18, and a React (Vite + TypeScript) control panel served under `/panel/`. It manages customers, quota/validity, subscriptions with QR delivery, live server telemetry, a node pool, and role-based administration — designed, documented and tested for public deployment.

> Status: **live at `https://naive.softarg.ir/panel/`** · Go + Web test suites green · schema v33

<p align="center">
  <img src="docs/screenshots/login.png" width="45%" alt="Login" />
  <img src="docs/screenshots/dashboard.png" width="45%" alt="Dashboard" />
</p>

---

## Table of contents

1. [Feature overview](#feature-overview)
2. [Quick start (Docker)](#quick-start-docker)
3. [Panel guide — every page and button](#panel-guide)
4. [Subscriptions & QR delivery](#subscriptions--qr-delivery)
5. [Connecting clients (Karing, v2rayNG, …)](#connecting-clients)
6. [Architecture](#architecture)
7. [Security model](#security-model)
8. [Tests & deployment](#tests--deployment)
9. [Roadmap — done vs. remaining](#roadmap)
10. [Design credits](#design-credits)

---

## Feature overview

- **Native NaiveProxy**: Caddy `forward_proxy` on :443 with `probe_resistance`, `hide_ip`, `hide_via`; per-customer HTTP basic credentials reconciled straight from the database.
- **Customer lifecycle**: create, edit, suspend/resume, revoke (safe soft-delete), volume add/set, extend, renew, next-plan, reset usage — all explicit, audited, capability-gated.
- **Dual QR delivery**: every customer gets a **subscription QR** (auto-updating client profile) **and** a **direct `naive+https://` QR** (manual, subscription-free import).
- **Bilingual account page** (`/s/<token>`): usage, quota, expiry, both QRs, and a complete per-client connection guide in Persian and English.
- **Content negotiation**: `/sub/<token>` renders raw `naive+https://`, sing-box JSON (Karing), Clash/Mihomo YAML, Hiddify and base64 v2ray lists from the same token via User-Agent detection (or `?family=` override).
- **Live telemetry**: SSE stream (`/api/v1/system/stream`, 1 s ticks) with CPU/RAM/disk gauges, live network rates from server-side counter deltas, load/uptime — polling fallback keeps data honest if the stream drops.
- **Node pool**: enroll additional nodes with tokens, publish signed manifests, drift/maintenance view (R5 mTLS fleet).
- **Roles**: owner / admin / reseller / operator / auditor with tenant isolation and reseller credit ledger.
- **Security**: `__Host-` session cookies, CSRF, optional TOTP MFA, AES-GCM encrypted subscription secrets, append-only accounting ledgers, audit events, per-request security logs.
- **Theme**: "Private Gold on Midnight Glass" — dark glassmorphism (Vazirmatn + JetBrains Mono), system-following light/dark, RTL-first.

---

## Quick start (Docker)

```bash
git clone https://github.com/DashSaman/PV-NativePanel.git
cd PV-NativePanel/docker

PVNAIVE_DOMAIN=naive.example.com \
PVNAIVE_PROXY_USER=pvbootstrap \
PVNAIVE_PROXY_PASSWORD='pick-a-strong-secret' \
PVNAIVE_OWNER_EMAIL=admin@example.com \
PVNAIVE_OWNER_PASSWORD='pick-an-owner-password' \
bash install.sh
```

- Panel: `https://<your-domain>/panel/`
- Account page: `https://<your-domain>/s/<token>`
- Machine subscription: `https://<your-domain>/sub/<token>`
- The persisted Caddyfile lives in `./data/caddy/Caddyfile` and survives recreation; hashed panel assets are served `immutable`, `index.html` is `no-cache` so updates reach users immediately.
- `PVNAIVE_NAIVE_PUBLIC_HOST` (host:443) controls the host used inside every subscription/direct link and QR payload.

---

## Panel guide

### ورود / Login — `#/login`

| Element | What it does |
| --- | --- |
| Email + Password fields | Hidden until hovered/focused ("stealth" reveal) — shoulder-surfing protection. |
| ورود امن | Issues `__Host-pvnaive_session` (HttpOnly, Secure) + CSRF cookie. |
| Brand tile | Gold Private Network mark, shared with the favicon. |

### داشبورد / Dashboard — `/`

- **KPI cards** — کل کاربران، کاربران فعال، پلن‌ها، نیازمند توجه. Each card is a real count from the database; the active/attention cards show share-of-total bars.
- **توزیع سرویس‌ها (donut)** — active / pending-first-connection / suspended / ended split, with percentage legend. Pure SVG, no chart library.
- **۷ روز آینده / ۳۰ روز آینده** — accounts expiring in the next 7/30 days.
- **مانیتورینگ زنده سرور** — live console fed by the SSE stream:
  - **زنده · هر ۱ ثانیه** pill — green = streaming, amber = polling fallback (5 s) if the stream drops; auto-retries every 10 s.
  - **Telemetry / Runtime / DB / API: OK** chips — server dependency health.
  - **Gauges** — پردازنده / حافظه / دیسک with threshold colors (green < 75 %, amber < 90 %, red ≥ 90 %).
  - **آپ‌تایم سرور** + Load 1/5/15 + interface + traffic semantics, all in tabular mono numerals.
  - **ترافیک شبکه** — live RX/TX area chart (90-sample window) with a "now" cursor; rates are computed server-side from counter deltas, never guessed in the browser.
- **پشتیبانی / خروج امن / مدیریت روزمره** — support shortcut, secure logout, and daily-management jump links.

<p align="center"><img src="docs/screenshots/dashboard.png" width="80%" alt="Dashboard" /></p>
<p align="center"><img src="docs/screenshots/monitoring.png" width="80%" alt="Live monitoring console" /></p>

### کاربران / Customers — `#/customers`

- **کاربر جدید** — create account: username, حجمی/نامحدود quota, validity mode (**از همین حالا / از اولین اتصال موفق / تاریخ دستی**), auto-generated password (shown once).
- **Search + فیلترهای پیشرفته** — by plan, status, expiry range, unlimited volume/expiry, reseller, group, tag.
- **Bulk bar** — appears when rows are selected: تعلیق، فعال‌سازی، Revoke، افزایش حجم، Set حجم، تمدید، اعمال پلن، تغییر گروه/تگ، Reissue subscription، Reset مصرف.
- **Row actions (••• menu)** — per customer:
  - **ویرایش مشخصات** — edit metadata/display profile (read-only-safe).
  - **تمدید سرویس** — extend by N days or renew from plan.
  - **اشتراک و QR** — opens the [dual-QR delivery dialog](#subscriptions--qr-delivery); strictly read-only.
  - **تغییر رمز** — explicit password rotation (auto-generate or custom ≥ 12 chars); never a side effect of viewing.
  - **صدور لینک جدید** — explicit subscription reissue with warning; invalidates the old token only.
  - **Reset مصرف** — resets usage after exact accounting is proven for the term.
  - **نشست‌های فعال** — live sessions per customer with per-session kill (exact-session kill control).
  - **لغو حساب** — safe revoke: revokes runtime credential, keeps history, fully audited.

<p align="center"><img src="docs/screenshots/customers.png" width="80%" alt="Customers" /></p>

### پلن‌ها و دسته‌بندی / Catalog — `#/catalog`

- Tabs for **پلن‌ها / گروه‌ها / تگ‌ها**; the plan editor appears only when **پلن جدید** is requested (no clutter).
- Plan fields: quota GB, duration days, unlimited variants, **اتصال همزمان (concurrency limit)** enforcement, next-plan chaining, on-hold policy.

<p align="center"><img src="docs/screenshots/catalog.png" width="80%" alt="Plan catalog" /></p>

### سیستم / Runtime — `#/runtime/naive`

- Live view of every runtime credential: origin (panel/import), status, revision.
- **Import** existing naive credentials, **create**, and **rotate password**; every change produces a runtime **revision** you can **apply/validate/rollback**.
- Reconciliation guarantee: the Caddy `basic_auth` block always mirrors database truth.

<p align="center"><img src="docs/screenshots/runtime.png" width="80%" alt="Runtime credentials" /></p>

### استخر گره‌ها / Pool — `#/pool`

- Stats row (total / in-sync / pending / draining), node table with health, capacity weight and last-seen.
- **Enrollment tokens** (shown once), signed manifest per node, **maintenance mode**, revision publishing with drift explanation.

<p align="center"><img src="docs/screenshots/pool.png" width="80%" alt="Node pool" /></p>

### امنیت و حساب / Security — `#/settings/security`

- Change own password, enroll/remove TOTP MFA (with recovery codes), view & revoke active sessions, panel-access policy.

<p align="center"><img src="docs/screenshots/security.png" width="80%" alt="Security settings" /></p>

---

## Subscriptions & QR delivery

Every customer carries one subscription token rendered three ways:

| Endpoint | Audience | Behaviour |
| --- | --- | --- |
| `/sub/<token>` | Machine clients | Content-negotiated: `naive+https://` (default), sing-box JSON (Karing UA), Clash/Mihomo YAML, Hiddify, base64 v2ray. `?family=` overrides. |
| `/s/<token>` | Humans | Read-only account page: usage, quota, expiry, both QRs, full connection guide (FA/EN switch). |
| Panel dialog | Owner | Same data, dual QR, copy buttons — opening it never mutates anything. |

<p align="center"><img src="docs/screenshots/qr-delivery.png" width="75%" alt="Dual QR delivery" /></p>

**Two independent connection paths — both QR-only:**

1. **Subscription QR** — client keeps a live profile; quota/password/token changes propagate on refresh.
2. **Direct Naive QR** (`naive+https://user:pass@host:443`) — manual import with zero subscription; ideal for quick tests.

<p align="center"><img src="docs/screenshots/account-page.png" width="80%" alt="Account page with dual QR" /></p>

---

## Connecting clients

The account page ships a complete bilingual guide. Summary:

| Client | Platform | Fastest import |
| --- | --- | --- |
| **Karing** (recommended) | Android / iOS / Windows / macOS | “+” → Scan QR → scan the subscription QR |
| v2rayNG | Android | Subscriptions → + → paste sub URL |
| NekoBox / NekoRay | Android / PC | Group → new subscription, or clipboard import of `naive+https://` |
| Hiddify | Android / iOS / PC | + → Add from clipboard / scan QR |
| sing-box | PC | fetch `/sub/<token>?family=singbox` |
| Clash Meta / Mihomo | PC | fetch `/sub/<token>?family=clash` |

Troubleshooting tips (rendered on the page): status flips to **فعال / آنلاین** after a successful connect; refresh the subscription after quota/password changes; depleted/expired accounts cannot connect; if all else fails use the Direct Naive QR.

<p align="center"><img src="docs/screenshots/account-guide.png" width="80%" alt="Step-by-step connection guide on the account page" /></p>

---

## Architecture

```
┌────────────────────────── pvnaive container ──────────────────────────┐
│  React panel (/panel)   Go API (:8080)   Caddy (:80/:443)             │
│   ├─ Vite build          ├─ /api/v1/*       ├─ TLS (Let's Encrypt)    │
│   ├─ RTL glass theme     ├─ SSE stream      ├─ forward_proxy (naive)  │
│   └─ local QR encoder    ├─ RBAC roles      ├─ /sub, /s, /panel       │
│                          ├─ accounting      └─ accounting socket ◄────┤
│                          └─ fleet/pool mTLS        exact byte counters│
│  PostgreSQL 18 (schema v33, forward-only migrations, encrypted backups)│
└────────────────────────────────────────────────────────────────────────┘
```

- **Accounting** is append-only: per-credential upload/download, restart-safe baselines, audited resets. Hard quota enforcement is gated behind exact-accounting proof; until proven, the UI shows an explicit *unavailable* state instead of inventing numbers.
- **First-use validity** starts only on a proven authenticated CONNECT — never on panel views, QR reads, or subscription fetches.
- **Steering / cover** (R6, default-off) and **fleet mTLS pull** (R5) are wired and feature-flagged.

---

## Security model

- Sessions: `__Host-pvnaive_session` (HttpOnly/Secure/SameSite) + CSRF token; hashed user-agent binding; MFA (TOTP + recovery codes).
- Secrets: runtime passwords AES-GCM encrypted at rest (`runtime-v1` key); subscription tokens stored as SHA-256 hashes; generated passwords shown exactly once.
- Read-only operations (view QR, copy links, details) are provably non-mutating — contract-tested.
- Audit trail for every lifecycle action; security log with per-request IDs; diagnostics bundles.
- The `/s/` page is private-by-obscurity with `noindex` and a public-sharing warning.

---

## Tests & deployment

```bash
# Go (unit + contract + subscription rendering + concurrency)
go test ./...

# Web (115 tests: geometry, API contracts, QR encoding, UI contracts)
cd web && npx vitest run

# Production build (Go binary + web dist + pinned accounting Caddy)
bash /root/pvnaive-fin3-build.sh   # clones GitHub main, builds image pvnaive:repo-fin3-<ts>
bash /root/pvnaive-fin3-deploy.sh  # DB dump → image swap → health checks
```

Deployment invariants: DB backup before every deploy; schema migrations forward-only; runtime credentials reconciled from DB truth on boot; health probe `GET /api/v1/health/ready`.

---

## Roadmap

### Shipped ✅

| Area | Delivered |
| --- | --- |
| Domain | `naive.softarg.ir` live (panel + subscriptions + TLS), IP fallback `45.141.148.59.nip.io` |
| Theme | R10 "Private Gold on Midnight Glass": gold brand from Private Network logo, frosted glass, gold nav/buttons |
| Chart typography | Latin-digit tabular mono on every gauge/axis/badge (fixes broken Persian-digit axis rendering) |
| Dual QR | Subscription QR + direct `naive+https://` QR, prominent in panel and on `/s/` page |
| Account page | Full FA/EN per-client connection guide + troubleshooting |
| Cache hygiene | `index.html` no-cache, hashed assets immutable — no stale bundles after deploys |
| Data hygiene | All test users/plans purged (35 users, 16 Test10GB plans, accounting rows); single clean `demo` customer |
| Docs | This bilingual README with screenshots for every page |

### In progress / next 🔜

| Area | Remaining |
| --- | --- |
| Named-client evidence (#120/#101) | Real disposable Karing import → parse → CONNECT acceptance on exact main |
| Pool PKI (#114) | Certificate overlap/rotation + revocation/replay fail-closed proofs |
| R6-FLIP (#115) | Cover/persona rehearsal before enabling by default |
| Production promotion (#100) | Fresh audit + backup/rollback gates on the trusted primary |
| Reseller storefront | Public purchase flow on top of the reseller ledger |

---

## Design credits

Visual methodology (glass surfaces, gold accent system, typography pairing, chart color separation) follows the open **[ui-ux-pro-max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill)** design-system dataset — Glassmorphism + Modern Dark profiles — adapted to Persian RTL with Vazirmatn and JetBrains Mono.
