# مشخصات محصول Standalone

## نقش‌ها

- `owner`: همه دسترسی‌ها، تنظیم امنیت و backup
- `admin`: کاربران، credential، subscription و گزارش
- `operator`: مشاهده و عملیات محدود؛ بدون secret و تنظیم امنیت
- `auditor`: فقط مشاهده audit و گزارش

## صفحات نسخه اول

| مسیر UI | صفحه | مجوز |
|---|---|---|
| `/login` | ورود | عمومی |
| `/` | Dashboard | authenticated |
| `/users` | کاربران، جستجو و bulk actions | admin |
| `/users/:id` | جزئیات، حجم، انقضا و credential | admin |
| `/subscriptions` | tokenها و preview خروجی | admin |
| `/runtime` | وضعیت Naive، config revision و rollback | owner/admin |
| `/usage` | نمودار مصرف و reconciliation | admin/auditor |
| `/system` | CPU/RAM/disk/TLS/version | operator+ |
| `/audit` | رویدادهای امنیتی و مدیریتی | owner/auditor |
| `/settings/security` | MFA/session/password policy | owner |
| `/settings/backup` | backup و restore test | owner |
| `/setup` | bootstrap اولیه، فقط قبل از ساخت owner | loopback/setup token |

صفحات Node/Fleet/Routing/Iran در MVP وجود ندارند.

## چرخه کاربر

`draft → active → suspended → expired/depleted → revoked`

- تغییر وضعیت و دلیل در Audit ثبت می‌شود.
- حذف سخت کاربر در UI وجود ندارد؛ ابتدا revoke و سپس retention policy.
- secret credential بعد از صدور کامل دوباره نمایش داده نمی‌شود؛ rotation، secret تازه صادر می‌کند.
- quota reset یک event مستقل است و ledger قبلی را پاک نمی‌کند.

## حجم

- ورودی و خروجی جدا ذخیره می‌شوند.
- UI مجموع را با مبنای واحد صریح نمایش می‌دهد.
- billing بر اساس usage ledger است، نه جمع access log.
- هر delta دارای boot_id، sequence و بازه زمانی است.
- وضعیت reconciliation و آخرین counter معتبر در صفحه Usage دیده می‌شود.
- quota enforcement محلی و fail-closed برای کاربر depleted است؛ خرابی UI سرویس کاربران سالم را قطع نمی‌کند.

## Subscription

- token تصادفی با entropy کافی و hash ذخیره‌شده
- revoke/rotate/expiry
- پاسخ با `Cache-Control: no-store`
- عدم قرار دادن username داخلی یا اطلاعات شخصی در remark
- خروجی Naive استاندارد و config مناسب clientهای تأییدشده
