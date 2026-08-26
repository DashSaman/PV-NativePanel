# نقشه راه اجرایی

## فاز 0 — Research snapshot

- [x] بررسی اولیهٔ 3x-ui، PasarGuard، Marzban، Remnawave، NaiveProxy و OV-PvNetwork
- [x] تعریف معماری و ریسک اصلی accounting
- [ ] ثبت commit/tag دقیق upstreamها
- [ ] دریافت لینک سورس پنل Production فعلی PVNetwork، در صورت وجود
- [ ] benchmark پایه روی نود آزمایشی

## فاز 1 — PoCهای مسدودکننده

1. Runtime accounting per credential
2. apply/rollback اتمیک Caddy config
3. 400 اتصال همزمان و throughput هدف
4. restart/reconnect و جلوگیری از دوباره‌شماری
5. client matrix: Windows، Android/Karing، iOS، macOS
6. certificate/domain rotation
7. health/failure آزمایشی چهار نود

معیار عبور: خطای billing قابل اندازه‌گیری، بدون double count، بدون قطع گسترده در config apply.

## فاز 2 — Backend MVP

- PostgreSQL schema و migration
- Auth/RBAC/audit
- User/credential/quota/expiry
- Node enrollment و desired/applied revision
- Usage ledger و enforcement
- Host/pool/assignment
- Subscription token و renderer
- Job queue و alert

## فاز 3 — Agent و Runtime

- Agent service
- mTLS enrollment
- validate/stage/apply/verify/rollback
- health/capacity/usage stream
- drain/maintenance/canary
- local cache و offline behavior

## فاز 4 — UI و عملیات

- dashboard، users، nodes، hosts، groups
- bulk actions و import
- incident/fallback controls
- backup/restore
- Telegram/webhook
- Persian/English UI

## فاز 5 — Installer

- profileهای standalone/controller/node
- preflight غیرمخرب
- version pin و checksum
- idempotent upgrade
- backup قبل از migration
- uninstall محافظه‌کارانه
- smoke test و diagnostics bundle

## فاز 6 — مهاجرت Production

- pilot داخلی
- 5% canary
- 25% و 50%
- primary direct خارجی
- ایران فقط fallback
- rollback plan و بازهٔ مشاهده
