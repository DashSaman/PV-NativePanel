# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## وضعیت

نام محصول **PVNative** و برند مادر **PVNETWORK** تثبیت شد. اسکلت Go API و React/TypeScript UI ایجاد شده، اما هنوز Runtime، DB، auth و business logic واقعی ندارد و Production-ready نیست.

## تغییرات موجود

- اسناد Brand، Product، API و Security
- SVG اختصاصی طلایی PVNative
- Route Registry مرکزی Backend
- صفحات UI و permission manifest
- health endpoint و security headers
- fail-closed برای routeهای مدیریتی
- تست route/public allowlist/header
- GitHub Actions برای Go و Web

## Scope تثبیت‌شده

- Standalone-first روی سرور خارج
- بدون ایران و بدون پنل مرکزی در MVP
- Naive استاندارد
- اتصال Controller/Node بعداً و اختیاری
- accounting per-credential یک PoC blocker

## وضعیت تست

- تست‌های Backend و Frontend در Repository و Workflow تعریف شده‌اند.
- محیط scratch فعلی Go toolchain ندارد؛ اجرای محلی Go با `go: command not found` متوقف شد.
- GitHub در زمان این به‌روزرسانی هنوز status check قابل مشاهده‌ای برای commit آخر برنگرداند.
- نباید تست‌ها را Passed فرض کرد تا Workflow واقعی سبز شود.
- lockfile وب هنوز تولید نشده؛ direct dependencyها دقیق pin شده‌اند و CI موقتاً از `npm install --ignore-scripts` استفاده می‌کند. پس از اولین resolution بازبینی‌شده باید `package-lock.json` commit و CI به `npm ci` تبدیل شود.

## کار بعدی دقیق

1. اجرای CI و رفع هر خطای build/test
2. ثبت lockfile بازبینی‌شده
3. Research Snapshot با commit/tag دقیق upstream
4. PoC accounting برای Caddy forwardproxy و sing-box Naive
5. ثبت نتیجه در `docs/POC_ACCOUNTING_FA.md`
6. انتخاب Runtime Adapter
7. سپس schema PostgreSQL، migration و auth/session امن

## خطوط قرمز مرحله بعد

- قبل از PoC، quota واقعی پیاده‌سازی یا تبلیغ نشود.
- auth نمایشی، token ثابت یا default password اضافه نشود.
- Runtime config با shell interpolation ساخته نشود.
- secret در log/API response ذخیره نشود.
- endpoint جدید بدون Access در Route Registry اضافه نشود.

## اطلاعات مورد نیاز از مالک هنگام Pilot

مشخصات سرور PoC، client matrix، سیاست quota/reset، همزمانی/device و تعداد کاربران Pilot.
