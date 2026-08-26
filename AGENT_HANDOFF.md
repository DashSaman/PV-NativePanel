# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## قانون ادامه استقرار

قبل از هر اقدام روی سرور، `ops/DEPLOYMENT_PROGRESS.md` را بخوان. فقط Stage دارای وضعیت `NEXT` اجرا شود و خروجی/خطا پیش از حرکت ثبت گردد. وضعیت فعلی: `S01-PREFLIGHT=PASSED` و `S02-FOUNDATION=NEXT`.

## وضعیت

PVNaive دارای specification عمیق، اسکلت Go/React، سایت عمومی Static، مدل وضعیت کاربر، Routeهای logs/diagnostics و boundary چندپروتکلی است. Runtime، DB، auth و business logic هنوز پیاده‌سازی نشده‌اند؛ Production-ready نیست.

## آخرین تغییرات

- اصلاح نام محصول از PVNative به PVNaive
- افزودن Routeهای نمایندگی، پلن، تمدید، اعلان و Subscription usage
- افزودن preflight فقط‌خواندنی برای testAmir5-3
- ایجاد executable اولیه cmd/pvnaive؛ هنوز Production-ready نیست

- تحقیق تکمیلی بازخوردهای قابل‌ردیابی 3x-ui، Marzban، Remnawave و NaiveProxy
- ثبت محدودیت بازیابی YouTube Comments؛ ادعای خواندن کامل کامنت‌ها ممنوع
- Protocol Adapter capability-based؛ مدل User مستقل از Naive/Xray
- قابلیت‌های accounting/session/speed/device/atomic reload/padding به‌صورت صریح
- Responsive shell با sidebar دسکتاپ، bottom navigation موبایل، touch target و reduced motion
- Content Pack schema برای تعویض موضوع سایت عمومی بدون تغییر binary
- سند Random Traffic: استفاده از padding رسمی Runtime، منع chaff جعلی پیش‌فرض
- Account status، Presence، Quota state و Runtime health جدا
- تک/چندکاربره با concurrency_limit
- Dark/Light/System theme
- Domain Activity owner-only، disabled-by-default و بدون path/query
- سایت عمومی Static و policy دانلود امن
- specification Easy Installer و restore/rollback requirement

## تصمیم سایت عمومی

سایت عمومی باید واقعی، Static، جدا از پنل و data plane و قابل تعویض با Content Pack باشد. هیچ موضوع سیاسی/رسانه‌ای در کد و installer هاردکد نشود. محتوای سیاسی درخواستی فقط به شکل pack اختیاری و با Asset دارای مجوز قابل ساخت است. بازدید یا ترافیک ساختگی و ادعای 2TB ممنوع است.

## تصمیم Random Data

NaiveProxy مرجع padding مذاکره‌شده و pseudo-random دارد. تزریق بایت یا درخواست رندوم مستقل پذیرفته نیست مگر PoC آزمایشگاهی، feature flag، budget، kill switch و benchmark نشان دهد سازگاری/accounting/fingerprint را خراب نمی‌کند. پیش‌فرض همیشه خاموش است.

## Scope

Standalone خارج، بدون ایران و بدون Controller در MVP. Controller/Node آینده اختیاری است. Naive اولین Adapter است؛ پروتکل‌های دیگر باید قرارداد docs/EXTENSIBILITY_FA.md را طی کنند.

## وضعیت تست

تست‌ها و CI تعریف شده‌اند، اما Passed فرض نشوند تا workflow سبز و lockfile بازبینی‌شده ثبت شود. تغییر adapter جدید باید با fake adapter و capability tests پوشش داده شود.

## کار بعدی دقیق

1. بررسی نتیجه GitHub Actions و اصلاح CI/lockfile
2. fake adapter و capability/UI visibility tests
3. PoC accounting بین Caddy forwardproxy و sing-box Naive
4. انتخاب Runtime Adapter
5. PostgreSQL schema: users، credentials، sessions، usage ledger، reset events، audit و logs metadata
6. Auth/session/MFA امن
7. User CRUD و UI واقعی responsive
8. content-pack loader + schema validation
9. Installer امضاشده، restore drill و Pilot

## الزامات غیرقابل حذف

- status حساب و online یکی نشوند
- depleted/expired با متن و رنگ/آیکن جدا
- UI قابلیت unsupported را نمایش ندهد
- Domain Activity پیش‌فرض خاموش و بدون TLS MITM
- optional collector نباید data plane را متوقف کند
- سایت عمومی ترافیک مصنوعی نسازد
- content topic در محصول hard-code نشود
- هیچ default password/token
- endpoint بدون Access ممنوع
- log بدون secret/token/query
- config apply باید validate/stage/atomic/rollback داشته باشد
