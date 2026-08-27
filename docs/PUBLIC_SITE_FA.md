# سایت عمومی هم‌میزبان با PVNaive

## هدف

روی همان دامنه/پورت 443 یک وب‌سایت عمومی واقعی، سالم و قابل استفاده نمایش داده می‌شود. هدف masquerade جعلی یا تولید ترافیک مصنوعی نیست؛ سایت باید محتوای مشروع داشته باشد و نگهداری شود.

## انتخاب

**Static site تولیدشده و سرو‌شده توسط Caddy** به‌جای WordPress:

- سطح حمله کمتر
- بدون PHP/MySQL و plugin
- مصرف RAM کم
- backup و deploy ساده
- cache، Range request و checksum برای فایل‌ها
- بدون وابستگی دیتاپلین به CMS

اگر بعداً تیم محتوا واقعاً CMS لازم داشت، CMS روی origin جدا و با hardening قرار می‌گیرد.

## Routeها

| Route | رفتار |
|---|---|
| `/` | صفحه اصلی عمومی |
| `/about` | درباره سرویس/پروژه عمومی |
| `/status` | وضعیت عمومی بدون IP و جزئیات زیرساخت |
| `/downloads` | فهرست فایل‌های عمومی مجاز |
| `/downloads/{file}` | دانلود با Range، ETag و checksum |
| `/privacy` | Privacy |
| `/terms` | Terms |
| `/support` | راه ارتباط عمومی |
| `/s/{token}` | صفحه Subscription، noindex/no-store |
| `/robots.txt` | crawl policy |
| `/favicon.ico` | برند PVNaive |

Admin panel باید روی hostname یا path غیرعمومی جدا با authentication باشد. نام و لینک پنل در سایت عمومی منتشر نمی‌شود.

## فایل‌های دانلود

فقط فایل‌هایی که مالک حق انتشارشان را دارد؛ مانند client guide، checksum، فایل نمونه یا نرم‌افزار آزاد با رعایت license. برای هر فایل:

- نام، version، size، MIME
- SHA-256
- license/source
- تاریخ انتشار
- malware scan نتیجه‌دار
- allowlist extension
- خارج از web root آپلود و پس از validation منتشر شود

آپلود عمومی وجود ندارد. directory listing خاموش است.

## واقع‌گرایی و ترافیک

ترافیک مصنوعی، refresh loop یا دانلود اجباری ساخته نمی‌شود. سایت باید طبیعی باشد: محتوای واقعی، status، docs و download مشروع. ادعای اینکه وجود سایت به‌تنهایی مانع فیلترشدن است قابل تضمین نیست.

## Caddy routing

درخواست معتبر Forward Proxy توسط ماژول Naive پردازش می‌شود؛ درخواست معمولی به file server می‌رود. config باید validate و atomic apply شود. سایت حتی در خطای Manager با last-known-good بالا می‌ماند.

