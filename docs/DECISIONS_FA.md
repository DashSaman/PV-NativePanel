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
