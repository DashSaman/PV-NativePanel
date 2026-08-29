# تصمیم‌های معماری

آخرین reconciliation: 2026-08-30

## ADR-001 — Wire protocol اختصاصی نمی‌سازیم

**تصمیم:** NaiveProxy استاندارد و کلاینت‌های سازگار حفظ می‌شوند.

**دلیل:** تغییر wire behavior ریسک fingerprint تازه، ناسازگاری client و بار نگهداری همگام با Chromium را بالا می‌برد. اختصاصی‌بودن در مدیریت، حسابداری و عملیات خواهد بود.

## ADR-002 — نسخهٔ اول کاملاً Standalone است

پروژه ابتدا روی یک سرور خارج و بدون پنل مرکزی کار می‌کند. Web UI/Manager محلی نباید availability دیتاپلین را کنترل کند. Fleet/Multi-node بعد از کامل‌شدن standalone P0/P1 می‌آید.

## ADR-003 — سرور و پهنای‌باند ایران خارج از Scope فعلی است

هیچ gateway، tunnel یا fallback ایران در MVP ساخته نمی‌شود. این موضوع اگر لازم شد بعداً به‌عنوان integration جدا بررسی می‌شود.

## ADR-004 — قابلیت اتصال آینده، بدون وابستگی امروز

boundaryهای Runtime Adapter و Desired/Applied Revision حفظ می‌شوند تا بعداً Agent/Controller اضافه شود؛ اما پیاده‌سازی standalone به Controller وابسته نیست.

## ADR-005 — PostgreSQL مبنای Production است

PostgreSQL مبنای Production است. SQLite فقط برای PoC یا نصب کوچک قابل بررسی است و نباید مسیر migration را بشکند.

Production audit مورخ 2026-08-30 schema 11 را روی PostgreSQL فعال نشان داد.

## ADR-006 — حسابداری exact قبل از enforcement باید اثبات شود

**تصمیم تاریخی:** hard quota نباید تا قبل از اثبات byte counter per-credential زیر restart/reconnect/multiplex/failure فعال شود.

**وضعیت جدید 2026-08-30:** feasibility/core این ADR توسط WS1 بسته شده است. Current main/Production دارای exact direct-Naive successful-write accounting، stable Runtime UUID identity، dedicated Telemetry Agent/socket، append-only/idempotent event ingest، boot/session/sequence/cumulative semantics و ServiceTerm isolation است.

**گیت باقی‌مانده:** legacy/adopted baseline truth، reset semantics و controlled Production acceptance برای quota race/exhaustion/reload/restart/reconnect. بنابراین accounting core را دوباره از صفر نمی‌سازیم، ولی enforcement را هم بدون acceptance evidence کامل Done اعلام نمی‌کنیم.

## ADR-007 — Cloudflare فقط برای control surfaces

صفحه مدیریت و Subscription می‌توانند پشت Cloudflare باشند. data-plane Naive به‌صورت پیش‌فرض DNS-only است. CDN/WARP/proxy-mode integrationهای Hiddify الزام معماری Naive نیستند و فقط در صورت نیاز مشخص Product بررسی می‌شوند.

## ADR-008 — secret قابل بازیابی در log نیست

credential plaintext فقط هنگام صدور/rotation و delivery لازم دیده می‌شود؛ در ذخیره‌سازی از hash یا envelope encryption مناسب استفاده می‌شود. raw Subscription token نیز one-time delivery است و DB digest غیرقابل‌بازیابی نگه می‌دارد.

## ADR-009 — UI شبیه Sanaei، مدل داده credential-centric نمی‌شود

**تصمیم:** تجربه Owner برای ساخت اکانت می‌تواند ساده و شبیه 3x-ui باشد، اما `User / ServiceTerm / RuntimeCredential / SubscriptionToken` چهار مفهوم جدا باقی می‌مانند.

**نتیجه:** quota، مدت و renewal به ServiceTerm تعلق دارند و rename/rotate رمز Runtime هویت تجاری کاربر را عوض نمی‌کند. Raw Runtime manager فقط سطح پیشرفته است؛ مسیر اصلی پنل `/panel/#/customers` است.

## ADR-010 — Subscription و QR secret-bearing فقط محلی و قابل لغو است

**تصمیم:** Subscription یک token تصادفی opaque دارد؛ token خام فقط one-time delivery است و DB فقط digest/prefix غیرحساس را نگه می‌دارد. QR در مرورگر ساخته می‌شود و هیچ سرویس QR ثالثی token را دریافت نمی‌کند.

**تصمیم تکمیلی:** host مقصد Naive از تنظیم صریح deployment می‌آید، نه از Host header درخواست. صدور مجدد Subscription token قبلی را revoke می‌کند و password rotation یک action جدا است.

**Current delivery contract:** `/sub/<token>` machine/client endpoint و `/s/<token>` human Account Page هستند. View/copy هر دو read-only هستند.

## ADR-011 — شروع «اولین اتصال» فقط از successful authenticated Runtime CONNECT

**تصمیم:** `on_first_successful_connection` با مشاهده QR، دریافت `/sub`، مشاهده `/s`، health check، Caddy reload، API request یا authentication ناموفق فعال نمی‌شود. فقط successful authenticated Naive CONNECT در trusted Runtime instrumentation اجازه activation دارد.

**وضعیت جدید 2026-08-30:** producer instrumentation و atomic compare-and-set activation core توسط WS1 merge شده‌اند و direct accounting/session telemetry روی Production live است.

**گیت باقی‌مانده:** قبل از اینکه feature Production-Done شود، controlled acceptance باید ثابت کند first connection، duplicate/reconnect، concurrent first connection، Runtime/server restart و Caddy reload semantics دقیقاً مطابق قرارداد هستند. بنابراین core موجود را rewrite نمی‌کنیم و در عین حال evidence gate را حذف نمی‌کنیم.

## ADR-012 — expiry مستقل است؛ hard quota فقط با exact accounting و shared budget

**تصمیم:** تعریف quota و expiry بخشی از commercial state است. used/remaining و enforcement فقط از exact accounting trusted می‌آیند؛ access-log estimate یا fake zero مجاز نیست.

**وضعیت جدید:** exact accounting و shared finite-quota reservation/settlement core وجود دارند. Remaining Production gate شامل simultaneous connections، race، exact exhaustion، reload/restart/reconnect، عدم negative remaining و عدم quota bypass است.

Manual Reset Usage و periodic traffic reset execution هنوز separate backlog هستند و نباید از وجود reset-strategy model به‌عنوان implemented enforcement نتیجه‌گیری شوند.

## ADR-013 — Route declaration implementation evidence نیست

**تصمیم:** `Routes` می‌تواند contractهای آینده را نگه دارد، اما Feature فقط وقتی implemented است که handler/store/schema/authorization/test و در صورت operator-facing بودن UI واقعی داشته باشد. Route بدون handler ممکن است `501 not_implemented` باشد و نباید در parity به‌عنوان DONE شمرده شود.

## ADR-014 — Production mutation باید transaction-like عملیاتی باشد

هر Runtime/Production mutation باید تا حد امکان این توالی را داشته باشد:

`preflight/read-only verify → backup → validate → apply bounded change → postflight → rollback on failure → evidence`

قبل از mutation Production، DB/config/Caddy/web/binary backup و rollback plan لازم است. هیچ feature rollout نباید user موجود را حذف یا password/token را بدون action صریح rotate کند.

## ADR-015 — Competitor parity برای behavior است، نه blind source copy

3x-ui (GPL-3.0)، PasarGuard (AGPL-3.0) و Hiddify Manager (GPL-3.0) برای UX/behavior/architecture reference استفاده می‌شوند. کد آنها بدون license-compatibility review وارد PVNaive نمی‌شود. OV-PvNetwork (MIT public snapshot) می‌تواند patternهای عملیاتی سازگار ارائه دهد، اما تفاوت معماری همچنان review/test می‌خواهد.
