# Logging، Diagnostics و Privacy

## تفکیک داده‌ها

### Audit log

چه کسی، چه عملی، روی چه objectی، قبل/بعدِ redact‌شده، نتیجه، زمان، request_id و source IP مدیریتی. Append-only و owner/auditor-only.

### Application log

خطا، warning و lifecycle سرویس با structured JSON. شامل stack داخلی در فایل امن، ولی پاسخ API فقط error code و request_id.

### Runtime log

شروع/توقف، config revision، validate/apply/rollback، certificate و counter collector. Credential و Authorization header همیشه redact می‌شوند.

### Access log سایت عمومی

روش، route class، status، bytes، latency و User-Agent خلاصه‌شده. Query، Cookie، Authorization و token Subscription ثبت نمی‌شوند.

### Usage ledger

رکورد billing و مستقل از logها. پاک‌کردن log نباید حجم را تغییر دهد.

## Domain Activity

Naive forward proxy می‌تواند مقصد CONNECT مانند hostname:port را ببیند؛ برای HTTPS مسیر صفحه، query و محتوای داخل TLS را نمی‌بیند و PVNaive نباید TLS MITM انجام دهد.

این قابلیت:

- پیش‌فرض خاموش
- فقط با مجوز Owner
- banner واضح درباره ریسک privacy
- فقط hostname نرمال‌شده، port، first_seen، last_seen، connection count و bytes
- بدون URL path، query، DNS payload یا محتوا
- retention پیش‌فرض 24 ساعت، حداکثر پیشنهادی 7 روز
- رمزنگاری at-rest
- دسترسی جداگانه `diagnostics:destinations:read`
- هر مشاهده و export در Audit
- امکان حذف فوری و pause collector
- collector failure هیچ اثری بر Runtime ندارد

برای گزارش بلندمدت، hostname خام نگهداری نشود؛ category یا keyed hash دوره‌ای کافی است.

## پنل Logs

صفحات:

- `/logs/application`
- `/logs/runtime`
- `/logs/security`
- `/diagnostics/requests/{requestId}`
- `/diagnostics/domain-activity`

فیلتر level/service/time/request_id/code، live tail محدود، download diagnostics bundle با redaction preview.

## Retention

- Audit: 180 روز یا سیاست کسب‌وکار
- Security events: 90 روز
- Application/Runtime: 14 تا 30 روز با rotation
- Public access: 7 روز
- Domain Activity: خاموش؛ در حالت روشن 24 ساعت
- Usage ledger: مطابق دوره مالی و backup

Disk quota، compression و حذف oldest-first لازم است. پرشدن disk نباید دیتاپلین را متوقف کند؛ alert و reserve space الزامی است.

