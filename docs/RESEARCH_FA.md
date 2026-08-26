# تحقیق تطبیقی پنل‌ها برای PV NativePanel

تاریخ بررسی: 2026-08-26

## دامنه و محدودیت بررسی

کد و مستندات رسمی 3x-ui، PasarGuard Panel/Node، Marzban، Remnawave، NaiveProxy و ریپازیتوری در دسترس `DashSaman/OV-PvNetwork` بررسی شد. سورس پنل Production نشان‌داده‌شده در تصویر (3X/PVNetwork) در میان ریپازیتوری‌های قابل دسترس پیدا نشد؛ بنابراین دربارهٔ آن فقط از قابلیت‌های قابل مشاهده و مستندات OV-PvNetwork استفاده شده و ادعای ممیزی سورس Production نداریم.

## جمع‌بندی اجرایی

ساخت پنل Naive منطقی است، ولی «اختصاصی بودن پنل» جلوی شناسایی پروتکل را تضمین نمی‌کند. مزیت واقعی از این موارد می‌آید:

1. استفاده از Naive استاندارد و به‌روز با fingerprint شبکهٔ Chromium؛
2. جداسازی دامنه و IP پنل/Subscription از دیتاپلین؛
3. چند Host و چند Node با سلامت‌سنجی و rollout کنترل‌شده؛
4. حسابداری مستقل و دقیق؛
5. fallback واقعی، نه چند لینک بی‌خبر از سلامت؛
6. rotation امن دامنه/credential بدون قطع کاربران.

## الگوهای قابل اقتباس

### 3x-ui (نسخه فعلی main)

نقاط ارزشمند:

- Master می‌تواند پنل‌های دیگر را با API token یا mTLS مدیریت کند.
- heartbeat، وضعیت online/offline، منابع و ترافیک نودها
- هویت پایدار نود و نمایش transitive nodeها
- Managed Host با override آدرس، پورت، SNI، Host، path، ALPN، fingerprint و scope نود/فرمت
- outbound routing، route test، connectivity test، balancer و observatory

درس: مدل `Node + ManagedHost + Health` را می‌گیریم، اما Node نباید یک پنل کامل دیگر باشد؛ Agent کوچک‌تر، idempotent و outbound-only بهتر است.

### PasarGuard

نقاط ارزشمند:

- REST API، Multi-node، RBAC و چند ادمین
- مدل quota/expiry/periodic reset و چند پروتکل برای کاربر
- jobهای node checker، node stats، record usages و reset usage
- Panel و Node جدا؛ Node با Xray API آمار را جمع می‌کند
- DBهای مختلف و نصب اسکریپتی

درس: حسابداری باید job/ledger مستقل و idempotent داشته باشد. مدل «عدد آخر counter» به تنهایی در restart/reset خطا می‌دهد.

### Marzban

نقاط ارزشمند:

- معماری Python/React، REST API، Multi-node
- subscription چندفرمتی، محدودیت حجم و تاریخ، CLI و Telegram
- تفکیک منطقی user/inbound/node

درس: API-first و subscription renderer باید از روز اول جزو هسته باشند، نه افزونهٔ پایانی.

### Remnawave

نقاط ارزشمند:

- تمرکز جدی روی Node، Host، routing، subscription page و عملیات
- تفکیک panel/node و مستندات عملیاتی مناسب

درس: Host یک موجودیت مستقل و versioned باشد؛ تغییر endpoint نباید نیازمند بازنویسی همهٔ کاربران باشد.

### PVNetwork / OV-PvNetwork

الگوهای مستندشده در ریپازیتوری قابل دسترس:

- controller/agent، desired-state reconciliation
- استقرار خودکار نود، چرخهٔ امن edit/delete
- داشبورد تجمیعی realtime
- drain/canary/maintenance
- معنی جهت ترافیک: Node TX = دانلود کاربر و Node RX = آپلود کاربر
- ترجیح نرخ‌سنجی server-side برای جلوگیری از spike ناشی از jitter مرورگر

درس: reconcile و lifecycle عملیات باید از نسخهٔ اول طراحی شود؛ اجرای دستور SSH موردی جای Agent را نمی‌گیرد.

## NaiveProxy و محدودیت حیاتی آمار

NaiveProxy رسمی از network stack کرومیوم، HTTP/2 multiplexing، application fronting و padding استفاده می‌کند و توصیه می‌کند client همیشه هم‌نسخهٔ Chrome نگه داشته شود. سرور مرجع Caddy forwardproxy است.

اما Caddy/forwardproxy به شکل آماده API حسابداری حجمی قابل اعتماد «برای هر credential» مشابه Xray ارائه نمی‌کند. بنابراین یکی از این مسیرها باید با benchmark و threat-model انتخاب شود:

- **A — Runtime instrumentation (ترجیح پژوهشی):** fork کوچک و قابل نگهداری از forwardproxy برای counter بایت موفق هر credential و export امن delta
- **B — eBPF/nftables:** مناسب per-IP/port، اما برای چند کاربر پشت یک listener معمولاً نسبت‌دادن دقیق به credential ممکن نیست
- **C — access log estimation:** برای billing دقیق پذیرفتنی نیست و ریسک privacy دارد
- **D — sing-box Naive inbound:** نیازمند اثبات semantics آمار per-user و تطابق fingerprint/رفتار با upstream قبل از انتخاب

تا زمانی که PoC حسابداری آزمون نشود، پروژه نباید ادعای quota دقیق داشته باشد.

## Cloudflare

Orange-cloud برای دیتاپلین forward proxy استاندارد راه‌حل عمومی و تضمین‌شده‌ای نیست. دامنه‌های Naive باید پیش‌فرض DNS-only باشند. Cloudflare می‌تواند برای پنل، صفحه Subscription و API عمومی استفاده شود؛ هر CDN transport باید جداگانه و مطابق شرایط سرویس آزموده شود.

## منابع اولیه

- [NaiveProxy upstream](https://github.com/klzgrad/naiveproxy)
- [3x-ui multi-node](https://github.com/MHSanaei/3x-ui/blob/main/docs/content/docs/en/operations/multi-node.mdx)
- [3x-ui routing](https://github.com/MHSanaei/3x-ui/blob/main/docs/content/docs/en/operations/outbounds-routing.mdx)
- [PasarGuard panel](https://github.com/PasarGuard/panel)
- [PasarGuard node](https://github.com/PasarGuard/node)
- [Marzban](https://github.com/Gozargah/Marzban)
- [Remnawave](https://github.com/remnawave/panel)
- [OV-PvNetwork](https://github.com/DashSaman/OV-PvNetwork)

قبل از شروع implementation، commit/tag دقیق هر upstream در یک Research Snapshot ثبت می‌شود تا نتایج بازتولیدپذیر باشند.
