# مشخصات محصول Standalone — PVNaive

## نقش‌ها و مرز نمایندگی

- `owner`: همه دسترسی‌ها، اعتبار نمایندگان، امنیت، Runtime و Backup
- `admin`: کاربران، نمایندگان، پلن، Subscription و گزارش
- `reseller`: فقط کاربران، پلن‌های مجاز و مصرف زیر tenant خودش
- `operator`: مشاهده و عملیات محدود؛ بدون secret و تنظیم امنیت
- `auditor`: فقط مشاهده audit و گزارش

هر رکورد تجاری `tenant_id` دارد. نماینده هرگز با ID قابل‌حدس یا فیلتر UI به اطلاعات نماینده دیگر دسترسی ندارد؛ isolation در query و policy backend اجباری است. اعتبار نماینده append-only ledger، سقف بدهی، هشدار کمبود اعتبار و audit دارد.

## صفحات نسخه اول

| مسیر UI | صفحه | مجوز |
|---|---|---|
| `/login` | ورود | عمومی |
| `/` | Dashboard | authenticated |
| `/users` | کاربران، جستجو و bulk actions | reseller+ |
| `/users/:id` | جزئیات، حجم، انقضا، خرید و credential | reseller+ |
| `/resellers` | نمایندگان و اعتبار | admin+ |
| `/plans` | پلن‌ها، قیمت و reset policy | reseller+ |
| `/subscriptions` | tokenها و preview خروجی | reseller+ |
| `/notifications` | قواعد و صف اعلان‌ها | admin+ |
| `/runtime` | وضعیت Naive، config revision و rollback | owner/admin |
| `/usage` | مصرف و reconciliation | admin/auditor |
| `/system` | CPU/RAM/disk/TLS/version | operator+ |
| `/audit` | رویدادهای امنیتی و مدیریتی | owner/auditor |
| `/settings/security` | MFA/session/password policy | owner |
| `/settings/backup` | backup و restore test | owner |
| `/setup` | bootstrap اولیه | loopback/setup token |

## صفحه Subscription کاربر

صفحه `/s/:token` باید بدون افشای اطلاعات داخلی نشان دهد:

- نام نمایشی سرویس و وضعیت فعال/تعلیق/منقضی/حجم‌تمام
- تاریخ خرید، تاریخ اولین اتصال و تاریخ انقضا
- روز و ساعت باقی‌مانده با timezone صریح
- حجم کل، مصرف‌شده، باقی‌مانده و درصد
- تاریخ reset بعدی
- تعداد دستگاه/اتصال مجاز
- آخرین به‌روزرسانی اطلاعات
- دکمه کپی لینک و deep-link کلاینت‌های پشتیبانی‌شده
- اعلان‌های قابل نمایش و راهنمای رفع خطا
- ETag و Cache-Control: no-store

## اعلان‌ها

قواعد پیش‌فرض قابل تغییر: حجم در 80% و 95%، تمام‌شدن حجم، 7/3/1 روز تا انقضا، انقضا، TLS نزدیک انقضا، Runtime down، backup failed و اعتبار کم نماینده. ارسال از Outbox با idempotency، dedup، retry محدود و ثبت نتیجه است. کانال‌ها در MVP: داخل پنل و Webhook؛ Telegram/Email بعد از secret management.

## چرخه کاربر و حجم

`draft → active → suspended → expired/depleted → revoked`. وضعیت حساب، حضور آنلاین، quota و Runtime جدا هستند. Upload/Download جدا و مجموع با واحد صریح نمایش داده می‌شود. reset یک event است و ledger قبلی حذف نمی‌شود. تاریخ خرید با شروع سرویس و اولین اتصال یکی فرض نمی‌شود.

## Subscription security

token تصادفی با entropy کافی و فقط hash ذخیره‌شده، rotate/revoke/expiry، rate limit و پاسخ no-store. username داخلی، tenant، قیمت خرید یا اطلاعات شخصی در remark قرار نمی‌گیرد.
