# نقشه راه اجرایی

## فاز 0 — Research snapshot

- [x] بررسی اولیهٔ 3x-ui، PasarGuard، Marzban، Remnawave، NaiveProxy و OV-PvNetwork
- [x] انتخاب Standalone خارج به‌عنوان Scope نسخهٔ اول
- [x] حذف ایران و پنل مرکزی از Scope فعلی
- [ ] ثبت commit/tag دقیق upstreamها
- [ ] benchmark پایه روی یک سرور آزمایشی خارج

## فاز 1 — PoCهای مسدودکننده

1. accounting مصرف per credential
2. apply/rollback اتمیک Runtime config
3. restart/reconnect و جلوگیری از double count
4. تست H2 multiplex و چند credential
5. client matrix: Windows، Android/Karing، iOS و macOS
6. certificate/domain rotation
7. تست تدریجی تا 400 اتصال همزمان و throughput هدف

معیار عبور: خطای billing اندازه‌گیری‌شده، بدون double count و بدون قطعی گسترده هنگام اعمال config.

## فاز 2 — Standalone Core

- schema و migration
- user/credential/quota/expiry
- Runtime Adapter
- Usage ledger و enforcement محلی
- Subscription token و renderer
- audit، health و backup
- API versioned

## فاز 3 — Standalone UI و عملیات

- dashboard و users
- bulk actions و import
- runtime/certificate management
- usage reports
- backup/restore
- alert و diagnostics
- رابط فارسی/انگلیسی

## فاز 4 — Installer مستقل

- profile `standalone`
- preflight غیرمخرب
- version pin و checksum
- upgrade idempotent
- backup قبل از migration
- rollback و uninstall محافظه‌کارانه
- smoke test و diagnostics bundle

## فاز 5 — Pilot Production

- تست داخلی
- canary محدود
- افزایش تدریجی کاربران
- بررسی مصرف، CPU/RAM، reconnect و کیفیت clientها
- rollback plan

## فاز 6 — اتصال اختیاری به پنل مرکزی؛ بعداً

- Agent outbound-only
- enrollment و mTLS
- desired/applied revision از راه دور
- multi-node، host pool و assignment
- حفظ standalone در قطع Controller

سناریوی ایران در این Roadmap وجود ندارد و در صورت نیاز باید سند و فاز جدا داشته باشد.
