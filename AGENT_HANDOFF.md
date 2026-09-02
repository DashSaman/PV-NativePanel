# Agent Handoff — PVNaive

> **CURRENT-STATE OVERRIDE — 2026-09-02:** The remainder of this file is preserved as historical S04 evidence from 2026-08-27 and is **not** the current execution checkpoint. Before any action, read `PROJECT_STATUS.md`, `CONTINUE_HERE.md`, and `HANDOFF.md`, then verify exact GitHub `main`, open PRs/CI and fresh Production health. Current roadmap state: Task13 = draft PR #64 at exact head `c79f5385b7a751c30282948423b8c34d1ba89deb`, all exact-head GitHub gates green, final real HTTP1/HTTP2 proof still pending; Task16 = draft PR #81 / issue #79 with schema21 TDD RED active. Production remains on Task15/schema20 and must not be used as a development lane. Do not treat the historical `s04-auth` / PR #2 instructions below as current work.

آخرین به‌روزرسانی تاریخی: 2026-08-27 23:50 UTC

این بخش و ادامه فایل برای حفظ evidence تاریخی S04 نگه‌داری می‌شود. برای continuation فعلی، canonical files بالا اولویت دارند.

## خلاصه فوری تاریخی

- Repository: `DashSaman/PV-NativePanel`
- نام صحیح محصول: `PVNaive`
- نام قدیمی/اشتباه: `PVNative`
- نام Repository فعلاً قدیمی مانده و نباید بدون هماهنگی rename شود.
- branch توسعه در checkpoint تاریخی: `s04-auth`
- PR تاریخی: `#2` با عنوان `S04-AUTH: production authentication foundation`
- base تاریخی: `main`
- S00 تا S03 در آن checkpoint: PASSED
- S04-AUTH در آن checkpoint: در حال استقرار زنده

## قانون کار با سرور

کاربر دستورات را دستی به‌صورت root روی `testAmir5-3` اجرا می‌کند و خروجی کامل را paste می‌کند. Agent باید هر بار فقط یک مرحله/دستور واضح بدهد و بعد از دیدن خروجی تصمیم بگیرد. از heredoc/base64 بسیار بزرگ برای انتقال فایل استفاده نکن؛ artifact واقعی upload شود و SHA-256 آن verify شود.

هیچ Stage تا زمانی که خود Stage و یک postflight مستقل هر دو PASS نشده‌اند، `PASSED` اعلام نشود.

## سرور هدف تاریخی

- Host: `testAmir5-3`
- IPv4: `91.107.182.147`
- IPv6: `2a01:4f8:c010:37ee::1`
- OS: Ubuntu 26.04 LTS
- Kernel: `7.0.0-30-generic`
- RAM: حدود 3.7 GiB
- Domain: `namir.softarg.ir`
- Caddy: v2.11.2 + `forward_proxy`
- Service: `caddy-naive.service`
- Caddyfile: `/etc/caddy/Caddyfile`
- Caddyfile SHA-256 ثابت و مورد انتظار تاریخی:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- PostgreSQL: `18/main`, port `5432`, فقط `127.0.0.1` و `::1`

## Stage ledger تاریخی

| Stage | Status در 2026-08-27 | توضیح |
|---|---|---|
| S00-NAMING | PASSED | نام محصول به PVNaive اصلاح شده |
| S01-PREFLIGHT | PASSED | DNS/TLS/Caddy/ports/capacity بررسی شد |
| S02-FOUNDATION | PASSED | user/directory/backup پایه ساخته شد |
| S03-DATABASE | PASSED | PostgreSQL 18 + schema v1 + RLS + backup/restore/health + postflight مستقل |
| S04-AUTH | IN PROGRESS | checkpoint تاریخی؛ دیگر current roadmap نیست |
| S05-S10 | BLOCKED | وضعیت تاریخی؛ برای current state به canonical files رجوع شود |

## S02/S03 historical evidence

- S02 PASSED: `2026-08-26 20:18:57 UTC`
- S03 marker: `/opt/pvnaive/S03_DATABASE.json`
- PostgreSQL: `18/main`, loopback-only
- DB: `pvnaive`
- roles: `pvnaive_owner`, `pvnaive_app`
- schema version نهایی S03: `1`
- encrypted backup + restore drill: PASSED
- Caddy/SSH/Firewall در آن مراحل تغییر نکرد.

Historical final output:

```text
S03_POSTFLIGHT=PASSED
S03_DATABASE=PASSED
NEXT_STAGE=S04-AUTH
```

## Historical S04 architecture/evidence

این بخش فقط برای forensic/history نگه داشته شده است. Authentication foundation شامل Argon2id، opaque session token، CSRF، RBAC، TOTP، recovery-code hash، lockout، session rotation/revoke/reuse detection، `/api/v1/me` و owner bootstrap بدون default password بود. برای وضعیت فعلی implementation/deployment به canonical files و GitHub exact state رجوع شود.
