# تحقیق بازخورد کاربران و شکاف‌های محصول

آخرین بازبینی: 2026-08-26

## روش و سطح اطمینان

منابع اصلی، مستندات، Issue و Discussion قابل ردیابی پروژه‌های 3x-ui، Marzban، PasarGuard، Remnawave، Hiddify Manager، X-UI forks و NaiveProxy هستند. ویدئوهای عمومی YouTube فقط برای کشف جریان کار و اصطلاحات استفاده می‌شوند؛ وقتی صفحهٔ Comment قابل بازیابی کامل نبوده، ادعای «بررسی همه کامنت‌ها» نمی‌شود.

## خواسته‌های پرتکرار که باید در PVNative باشند

| نیاز | تصمیم PVNative |
|---|---|
| Bulk create/edit/disable/delete | Job آسنکرون، preview و audit؛ حذف بدون تأیید دو مرحله‌ای ممنوع |
| User template | Plan/Template نسخه‌دار؛ تغییر Template کاربران قبلی را ناخواسته تغییر نمی‌دهد |
| Next plan / auto-renew | Renewal policy جدا با idempotency key |
| Backup و restore ساده | backup کافی نیست؛ restore drill، verify و rollback اجباری |
| Update آسان | stable/canary channel، نسخه pin، backup قبل از update و rollback |
| Subscription پایدار | token قابل rotate/revoke، بدون افشای ID داخلی |
| Client deep-link | renderer جدا برای Karing، v2rayN، Clash، sing-box و کلاینت‌های بعدی |
| Device/concurrency limit | Session/Credential؛ IP به‌تنهایی شناسه دستگاه نیست |
| Speed limit | فقط وقتی Runtime Adapter آن را واقعاً پشتیبانی کند |
| Note و tag | searchable، RBAC-aware و بدون نمایش ناخواسته در Subscription |
| Notification | outbox پایدار، retry محدود، dedup و failover |
| مشاهده سلامت | Runtime، حساب، quota و presence چهار state جدا |
| Mobile admin | bottom navigation، drawer، جدول به card و عملیات لمسی 44px |
| Reseller | quota delegation، price isolation و audit؛ در MVP غیرفعال |
| Billing | ledger append-only؛ counter خام Runtime منبع نهایی حسابداری نیست |
| Protocol بیشتر | Adapter capability-based؛ هیچ شرط Naive/Xray در مدل User |

## کمبودهایی که معمولاً دیر دیده می‌شوند

1. restore آزمایش‌نشده؛
2. migration دیتابیس بدون dry-run؛
3. write amplification در ثبت usage؛
4. time-zone و resetهای تکراری؛
5. clock skew میان Runtime و پنل؛
6. tokenهای Subscription غیرقابل revoke؛
7. race میان تمدید، reset و قطع کاربر؛
8. قابلیت UI که Runtime واقعاً ندارد؛
9. صفحه موبایل که فقط کوچک شده و قابل‌عملیات نیست؛
10. export/erase داده و retention؛
11. accessibility، keyboard navigation و reduced motion؛
12. safe mode هنگام خرابی Runtime؛
13. config diff، validate و atomic apply؛
14. disaster recovery با RPO/RTO مشخص؛
15. تشخیص drift میان DB، Runtime و فایل تنظیمات.

## اولویت پیشنهادی

P0: accounting، auth/MFA، atomic config، backup/restore، audit و installer امن.

P1: CRUD واقعی، plan/template، subscription renderer، responsive UI و notification outbox.

P2: adapter دوم، reseller، billing integration و fleet/controller اختیاری.

## منابع قابل ردیابی

- Marzban #986: Bulk action، template و کنترل بهتر inbound
- Marzban #111: قالب ساخت کاربر
- Marzban #1001: update channel، update checker و rollback
- Marzban #1940: restore یک‌فرمانی
- NaiveProxy README و issue #76/#72: padding مذاکره‌شده و تصادفی
- Remnawave docs/repositories: تفکیک backend، frontend و subscription page

این فایل snapshot پژوهش است، نه تضمین پیاده‌سازی. هر قابلیت باید در ROADMAP و تست پذیرش وارد شود.
