# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## قانون ادامه استقرار

قبل از هر اقدام روی سرور، `ops/DEPLOYMENT_PROGRESS.md` را بخوان. فقط Stage دارای وضعیت `NEXT` اجرا شود و خروجی/خطا پیش از حرکت ثبت گردد. وضعیت فعلی: `S01-PREFLIGHT=PASSED`، `S02-FOUNDATION=PASSED` و `S03-DATABASE=NEXT`.

## آخرین استقرار سرور

S02 در `testAmir5-3` موفق شد. Backup سالم: `/var/backups/pvnaive/20260826T201857Z`. Caddy/SSH/Firewall تغییر نکردند. برای جزئیات `ops/DEPLOYMENT_PROGRESS.md` مرجع اصلی است.

## وضعیت

PVNaive دارای specification عمیق، اسکلت Go/React، سایت عمومی Static، مدل وضعیت کاربر، Routeهای logs/diagnostics و boundary چندپروتکلی است. Schema و ابزارهای امن S03 PostgreSQL آماده‌اند، اما هنوز روی سرور اجرا نشده‌اند. Runtime، auth و business logic پیاده‌سازی نشده‌اند؛ Production-ready نیست.

## آخرین تغییرات

- آماده‌سازی کامل S03: schema/migration/rollback/backup/restore/health/systemd/Stage script
- RLS امضاشده با session hash و اجبار `sql.Tx` در Backend boundary
- تست privilege escalation، tenant isolation، ledger، migration safety و backup/restore
- اصلاح آخرین نام‌های executable/UI/docs از PVNative به PVNaive؛ Repository بدون تغییر
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

تست محلی frontend (۸ تست + build)، Bash syntax، checksum، diff check، Go 1.24.4 formatting/vet/test سبز است. GitHub Actions run `33031844663` پیش از تخصیص runner شکست خورد: هر سه Job `runner_id=0` و `steps=[]` داشتند و هیچ تستی اجرا نشد. PostgreSQL server محلی نیز به‌علت محدودیت OS-user محیط configure نشد؛ integration دیتابیس فقط وقتی PASSED است که Stage واقعی migration+RLS+rollback+backup+restore را کامل کند. lockfile ساخته و `npm ci` فعال شده است.

## کار بعدی دقیق

1. بررسی GitHub Actions مربوط به Commit S03 و اصلاح هر failure
2. اجرای فقط `scripts/stages/S03-database.sh` روی `testAmir5-3` و ثبت خروجی کامل
3. در صورت `S03_RESULT=PASSED` تغییر S03 به PASSED و S04-AUTH به NEXT؛ در غیر این صورت ثبت failure و حفظ S03=NEXT
4. سپس Auth/session/MFA امن
5. fake adapter و capability/UI visibility tests
6. PoC accounting بین Caddy forwardproxy و sing-box Naive
7. User CRUD و UI واقعی responsive
8. content-pack loader + schema validation
9. Installer امضاشده و Pilot

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
