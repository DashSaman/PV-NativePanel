# تصمیم‌های معماری

## ADR-001 — Wire protocol اختصاصی نمی‌سازیم

**تصمیم:** NaiveProxy استاندارد و کلاینت‌های سازگار حفظ می‌شوند.

**دلیل:** تغییر wire behavior ریسک fingerprint تازه، ناسازگاری client و بار نگهداری همگام با Chromium را بالا می‌برد. اختصاصی‌بودن در مدیریت، حسابداری و عملیات خواهد بود.

## ADR-002 — نسخهٔ اول کاملاً Standalone است

پروژه ابتدا روی یک سرور خارج و بدون پنل مرکزی کار می‌کند. Web UI/Manager محلی نباید availability دیتاپلین را کنترل کند.

## ADR-003 — سرور و پهنای‌باند ایران خارج از Scope فعلی است

هیچ gateway، tunnel یا fallback ایران در MVP ساخته نمی‌شود. این موضوع اگر لازم شد بعداً به‌عنوان integration جدا بررسی می‌شود.

## ADR-004 — قابلیت اتصال آینده، بدون وابستگی امروز

boundaryهای Runtime Adapter و Desired/Applied Revision حفظ می‌شوند تا بعداً Agent/Controller اضافه شود؛ اما پیاده‌سازی MVP به Controller وابسته نیست.

## ADR-005 — انتخاب DB با benchmark

PostgreSQL مبنای Production است. SQLite فقط برای PoC یا نصب کوچک بررسی می‌شود و نباید مسیر migration را ببندد.

## ADR-006 — حسابداری PoC blocker است

تا زمانی که byte counter per-credential زیر restart، reconnect، multiplex و failure آزمون نشده، quota آماده علامت نمی‌خورد.

## ADR-007 — Cloudflare فقط برای control surfaces

صفحه مدیریت و Subscription می‌توانند پشت Cloudflare باشند. data-plane Naive به‌صورت پیش‌فرض DNS-only است.

## ADR-008 — secret قابل بازیابی در log نیست

credential plaintext فقط هنگام صدور/rotation و delivery لازم دیده می‌شود؛ در ذخیره‌سازی از hash یا envelope encryption مناسب استفاده می‌شود.

## ADR-009 — UI شبیه Sanaei، مدل داده شبیه credential-centric نمی‌شود

**تصمیم:** تجربه Owner برای ساخت اکانت می‌تواند ساده و شبیه 3x-ui باشد، اما `User / ServiceTerm / RuntimeCredential / SubscriptionToken` چهار مفهوم جدا باقی می‌مانند.

**نتیجه:** quota، مدت و renewal به service term تعلق دارند و rename/rotate رمز Runtime هویت تجاری کاربر را عوض نمی‌کند. Raw Runtime manager فقط سطح پیشرفته است؛ مسیر اصلی پنل `/panel/#/customers` است.

## ADR-010 — Subscription و QR secret-bearing فقط محلی و قابل لغو است

**تصمیم:** Subscription یک token تصادفی ۲۵۶ بیتی opaque دارد؛ token خام فقط one-time delivery است و DB فقط hash را نگه می‌دارد. QR در مرورگر ساخته می‌شود و هیچ سرویس QR ثالثی token را دریافت نمی‌کند.

**تصمیم تکمیلی:** host مقصد Naive از تنظیم صریح `PVNAIVE_NAIVE_PUBLIC_HOST` می‌آید، نه از Host header درخواست. صدور مجدد Subscription token قبلی را atomically revoke می‌کند و با idempotency ledger از rotate دوباره در retry جلوگیری می‌شود.

## ADR-011 — شروع «اولین اتصال» فقط از رویداد احراز‌شده Runtime

**تصمیم:** `on_first_successful_connection` با مشاهده پنل، دریافت Subscription، reload Caddy یا authentication ناموفق فعال نمی‌شود. فقط یک event داخلی با `authenticated CONNECT + runtime_credential_id + observed_at` اجازه ورود به compare-and-set activation را دارد.

**وضعیت Production:** boundary و atomic activation پیاده شده‌اند، اما تا زمانی که producer این event در مسیر pinned Naive/Caddy با تست packet/runtime اثبات نشود، producer به‌عنوان live capability اعلام نمی‌شود. در این وضعیت termهای first-use باید `pending` بمانند؛ راه‌حل موقت production استفاده از `on_creation` یا `fixed_expiry` است.

## ADR-012 — expiry قابل اعمال است، hard quota هنوز capability-gated است

تعریف quota و نمایش حجم خریداری‌شده مستقل از Usage Accounting است. expiry و state تجاری می‌توانند مدیریت شوند، اما عدد used/remaining و قطع خودکار بر اساس byte quota تا پاس‌شدن PVN-045..049 فعال نمی‌شوند. UI در این فاصله صریحاً `exact_accounting_not_proven` نشان می‌دهد و مقدار ساختگی صفر تولید نمی‌کند.
