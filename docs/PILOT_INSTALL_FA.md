# PVNaive — S05 Pilot Install Runbook — Historical Snapshot

آخرین reconciliation: 2026-08-30

> **HISTORICAL / DO NOT RUN ON CURRENT PRODUCTION**
>
> این Runbook برای ارتقای قدیمی S05 روی `testAmir5-3` نوشته شده بود. Production فعلی از آن مرحله عبور کرده و در audit مورخ 2026-08-30 روی PostgreSQL schema **11** با API، Runtime Agent، Telemetry Agent، Caddy/Naive و exact accounting فعال مشاهده شد.
>
> اجرای `S05-preflight.sh` / `S05-upgrade.sh` یا artifact قدیمی S05 روی Production فعلی می‌تواند downgrade/state drift ایجاد کند. برای current Production از این فایل به‌عنوان دستور نصب استفاده نکن.

## Current safe rule

قبل از هر mutation روی Production:

1. آخرین `main` و آخرین evidence را دوباره fetch/audit کن.
2. schema، service state، Caddy، Runtime Agent، Telemetry Agent، accounting sockets و web release را read-only بررسی کن.
3. DB backup بگیر.
4. config backup بگیر.
5. Caddy backup بگیر.
6. web backup بگیر.
7. binary backup بگیر.
8. rollback plan دقیق داشته باش.
9. فقط artifact version-pinned مربوط به **همان exact green HEAD** را استفاده کن.
10. validate → stage/apply → verify → rollback-on-failure را رعایت کن.
11. هیچ secret/password/token/key را در shell transcript، Git، Chat یا CI log منتشر نکن.

## Current Production baseline from 2026-08-30 read-only audit

- panel: `https://namir.softarg.ir/panel/` → HTTP 200؛
- API readiness → HTTP 200؛
- API listener → loopback `127.0.0.1:8080`؛
- `pvnaive-api.service` active؛
- `pvnaive-runtime-agent.service` active؛
- `pvnaive-telemetry-agent.service` active؛
- `caddy-naive.service` active؛
- PostgreSQL schema = 11؛
- live direct-accounting event/session data موجود؛
- root filesystem در لحظه audit حدود 79% مصرف داشت؛
- backup files موجود بودند، اما scheduled PVNaive backup timer مشاهده نشد؛
- deployment markerها با binary/web mtime جدیدتر همگام نبودند و برای Release provenance کافی نیستند.

این baseline فقط evidence همان audit است؛ قبل از mutation باید دوباره اندازه‌گیری شود.

## Current deployment documentation

برای current state ابتدا بخوان:

1. `PROJECT_STATUS.md`
2. `HANDOFF.md`
3. `CONTINUE_HERE.md`
4. `KNOWN_ISSUES.md`
5. `ROADMAP.md`
6. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
7. جدیدترین `ops/evidence/*`

Fresh installer/versioned upgrade/rollback عمومی هنوز جزو backlog Production-Ready است. تا وقتی آن lifecycle ساخته و روی clean Ubuntu آزمایش نشده، هیچ runbook مرحله‌ای قدیمی را به‌عنوان one-line installer فعلی معرفی نکن.

## Historical recovery

نسخه کامل Runbook قدیمی S05 در Git history در `main@a021aa4b62c35b775fb521d042b2f8e6dbde10b0` و commitهای پیش از reconciliation محفوظ است و فقط برای forensic/history باید به آن مراجعه شود.
