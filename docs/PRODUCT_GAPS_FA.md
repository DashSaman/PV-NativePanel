# ارزیابی کمبودها و پیشنهاد محصول

## کمبودهای فعلی Scaffold

1. Accounting اثبات‌نشده
2. Auth/MFA/RBAC واقعی ندارد
3. DB و migration ندارد
4. Runtime Adapter ندارد
5. UI فعلاً route/visual scaffold است
6. CI هنوز نتیجه تأییدشده ندارد
7. Subscription renderer/client matrix ندارد
8. Public site فقط صفحه پایه است
9. backup/restore و installer فقط contract هستند
10. notification و diagnostics collector پیاده نشده‌اند

## پیشنهادهای متمایزکننده PVNaive

### Reliability score برای هر User

به‌جای Online ساده: آخرین اتصال، accounting lag، failed auth، session churn و runtime state را خلاصه کند؛ بدون ذخیره مقصدهای مرور.

### Explainable status

روی badge کلیک شود و دقیق بگوید چرا کاربر قطع است: manual suspension، expiry، quota، concurrency، credential revoke یا runtime apply failure.

### Dry-run همه عملیات Bulk

قبل از Apply: تعداد affected، conflict، estimated runtime config changes و rollback plan.

### Subscription compatibility lab

Owner یک subscription را مقابل rendererهای Karing/Happ/Streisand/Shadowrocket/v2rayN preview و validate کند؛ secretها در screenshot/log redact شوند.

### Support bundle امن

با request_id، health، revision و errorها؛ قبل دانلود redaction preview و expiry خودکار bundle.

### Renewal queue

خرید/تمدید زودتر از پایان plan به Next Plan queue می‌رود و با policy `on_expiry`، `on_depletion` یا `threshold` فعال می‌شود؛ event idempotent.

### Public/Admin isolation

hostname، CSP، cookie scope، access log و release pipeline جدا؛ شکست سایت عمومی Manager را متوقف نکند و بالعکس.

### Capability-first UI

فیلد speed/device/session فقط وقتی Adapter واقعاً پشتیبانی می‌کند نمایش داده شود؛ UI قابلیت جعلی وعده ندهد.

## پیشنهاد عدم ساخت در MVP

Payment و ربات فروش، multi-node، routing پیچیده، protocol marketplace و Domain History بلندمدت. ابتدا هسته قابل اعتماد ساخته شود.

