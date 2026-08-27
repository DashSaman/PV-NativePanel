# Agent Handoff

آخرین به‌روزرسانی: 2026-08-27

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

تست محلی frontend (۸ تست + build)، Bash syntax، checksum، diff check، Go 1.24.4 formatting/vet/test سبز است. GitHub Actions runهای `33031844663` و `33032317065` پیش از تخصیص runner شکست خوردند: هر سه Job در هر دو run، `runner_id=0` و `steps=[]` داشتند و هیچ تستی اجرا نشد. PostgreSQL server محلی نیز به‌علت محدودیت OS-user محیط configure نشد؛ integration دیتابیس فقط وقتی PASSED است که Stage واقعی migration+RLS+rollback+backup+restore را کامل کند. Commitهای S03: `45ba5c4`، `fbf21e8` و preflight fix `133d36b`. lockfile ساخته و `npm ci` فعال شده است.

## آخرین تلاش S03 — 2026-08-27 02:27 UTC

bootstrap مربوط به Commit `6d4e5ce` پیش از اجرای Stage با `base64: error: invalid input` متوقف شد. terminal log نشان داد payload بزرگ copy/paste با bundle دارای SHA-256 `b9199c30eff3df4c71ade1c7deb642a9ac00780ce5f8437e08641a76a59495cd` یکسان نیست. `set -Eeuo pipefail` مانع ادامه شد؛ PostgreSQL، Caddy/NaiveProxy، SSH و UFW تغییری نکردند و rollback لازم نبود. S03 همچنان `NEXT` است. دفعه بعد فایل bundle باید upload و با launcher کوتاه SHA-gated اجرا شود؛ از تکرار payload بزرگ در clipboard خودداری کن.

bundle نهایی دوباره extract و با source مقایسه شد؛ syntax اسکریپت‌ها و هر دو checksum Migration پاس شدند. یک اجرای اولیه checksum از cwd اشتباه فقط `no file was verified` داد و با اجرای manifest از directory صحیح اصلاح شد؛ این مورد تغییری روی سرور نداشت.

تلاش دوم در `2026-08-27 02:52:22 UTC` انتقال bundle را با SHA صحیح و هر دو checksum Migration پاس کرد، اما preflight با syntax error عبارت awk بررسی پورت متوقف شد و پیام اشتباه port 22 داد؛ همان SSH session نشان می‌دهد پورت باز بوده است. اجرا پیش از APT و هر mutation دیتابیس متوقف شد و Caddy/NaiveProxy، SSH و UFW تغییر نکردند. `die()` نیز به‌علت استفاده از `exit 1`، `ERR` trap مرکزی را اجرا نکرد؛ هر دو نقص باید همراه regression test اصلاح شوند. S03 همچنان `NEXT` است.

اصلاح انجام شد: predicate پورت به helper تست‌پذیر منتقل و `die()` به return غیرصفر تغییر کرد. regression test واقعی IPv4/IPv6، پورت غایب/نامعتبر و `ERR` trap پاس شد؛ shell/SQL checksum و frontend ۸ test/build نیز پاس شدند. test اولیه trap به‌علت semantics شرط `if` شکست خورد و harness با shell مستقل اصلاح شد؛ فرمان validation نیز fail-fast شد. Go در Runtime فعلی در PATH نبود، اما هیچ فایل Go تغییر نکرده و اجرای کامل قبلی Go 1.24.4 پاس است. bundle جدید هنوز باید ساخته و روی سرور اجرا شود؛ S03 همچنان `NEXT` است.

hardening چندعاملی بعد از `133d36b` نیز کامل شد: rollback با نگهبان root `BASHPID` دقیقاً یک‌بار اجرا می‌شود؛ signal و rollback-step failure گزارش می‌شوند؛ marker موفقیت فقط پس از final gate و اتمیک نوشته می‌شود؛ marker قبلی در verify failure حذف نمی‌شود؛ health oneshot از `Result=success` سنجیده می‌شود؛ و provenance marker پایدار اجازه retry cluster متعلق به S03 را می‌دهد ولی cluster ناشناس را رد می‌کند. backup/restore کاملاً streaming و بدون dump plaintext است، storage root و DB port fail-closed هستند و تست‌های جدید پاس شدند. candidate با SHA `f9f0d57f...` منسوخ است و نباید استفاده شود. دو failure محیطی npm با اجرای مستقیم binaryهای موجود دور زده شد و ۸ test/build مجدداً پاس شد؛ Go در Runtime جاری موجود نیست و هیچ Go file تغییر نکرد. ممیزی نهایی agentها blocker کدی دیگری پیدا نکرد. Commit کامل hardening: `ded5af275d8e0000de25ce97d1de268fe54f58f8`. S03 هنوز `NEXT` است و هیچ اقدام جدیدی روی سرور نشده است.

CI run `33073904109` روی head مستندات `95a2689a0770e71db16dbd11bb436e9a3e6d92ab` دوباره پیش از runner شکست خورد؛ `go/web/database` همگی `runner_id=0` و `steps=[]` داشتند، پس هیچ تست CI اجرا نشد. bundle قطعی `pvnaive-s03-95a2689.tar.gz` با SHA-256 `9decbd705f548160343bdc41894b66ece86d53634e7fbf3719bac02f09be2b47` و اندازه 20358 byte، inventory ۱۳فایلی، مقایسه byte-for-byte، syntax، mode و migration checksums را پاس کرده است. check اولیه mode به‌علت glob منبع، S02 خارج از bundle را اشتباه طلب کرد و fail-fast شد؛ check صریح اصلاح و `BUNDLE_VERIFICATION=PASSED` شد. فایل باید دقیقاً به `/root/pvnaive-s03-95a2689.tar.gz` upload شود و فقط launcher نهایی S03 اجرا گردد. Stage همچنان `NEXT` است.

## کار بعدی دقیق

1. ساخت و upload bundle جدید fixشده با SHA-256 تازه
2. اجرای فقط `scripts/stages/S03-database.sh` روی `testAmir5-3` با launcher کوتاه و ثبت خروجی کامل
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
