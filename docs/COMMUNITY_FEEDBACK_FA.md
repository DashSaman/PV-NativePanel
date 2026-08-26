# بازخورد جامعه و پیشنهادهای پرتکرار

تاریخ بررسی: 2026-08-26

## محدودیت منبع

GitHub Issues/README و مستندات رسمی قابل ردیابی بررسی شدند. جست‌وجوی YouTube ویدئوهای آموزشی فارسی درباره PasarGuard، پنل‌های محدودیت حجم، نصب/backup/firewall و ربات فروش را پیدا کرد، اما YouTube دریافت صفحه و کامنت‌ها را throttle کرد. بنابراین هیچ ادعایی درباره «تمام کامنت‌ها» نداریم و یافته‌های YouTube فقط سیگنال موضوعی‌اند.

## پنل‌های اضافه‌شده به مقایسه

- Hiddify Manager: مدیریت دامنه‌های connection/subscription، چند transport/protocol، UX کاربر نهایی و mobile management
- S-UI: Sing-box، multi-protocol/client/inbound، routing UI، status، subscription و Dark/Light
- Xboard/Xboard-Node: مدل تجاری panel/node، speed/device/alive-IP، push+poll و چند kernel
- V2Board family: plan/order/payment/reseller؛ ولی سطح حریم خصوصی و نگهداری forkها باید جدا ارزیابی شود
- 2s-ui: مشتق Sing-box با multi-node و protocolهای جدید؛ مرجع تحقیق، نه dependency بدون audit

## درخواست‌های پرتکرار و تصمیم PVNative

| درخواست جامعه | شاهد نمونه | تصمیم |
|---|---|---|
| Renew هنگام depletion | 3x-ui #6314/#6298 | Next Plan event-driven با سقف renewal |
| Stable subscription token | PasarGuard #811 | یک token فعال پیش‌فرض، rotate/revoke صریح |
| Custom subscription slug | PasarGuard #792 | slug نمایشی جدا از secret token |
| App deep-link | 3x-ui #4653/#4977 | renderer per-client و capability matrix |
| HWID/device limit | 3x-ui #5170/#6234 | اختیاری؛ credential-first، bulk و reset |
| Per-user speed | 3x-ui #5195 | capability؛ فقط اگر runtime enforcement معتبر |
| Protocol UI بدون raw JSON | 3x-ui #3901 | schema-driven protocol form |
| چند routing identity در یک sub | 3x-ui #4776 | ProfileVariant جدا از User/Credential |
| مدیریت HTTP/Mixed مثل بقیه | 3x-ui #4852/#6261 | protocol adapter با accounting capability |
| Port conflict validation | 3x-ui #3904 | preflight global listener registry |
| Notification failover | 3x-ui #6327 | notifier chain + retry/dead-letter |
| Restore یک‌فرمانی | Marzban #1940 | installer: verify-only/restore/rollback |
| Login bot protection | Marzban #2012 | rate limit، delay، optional CAPTCHA |
| Note در جدول | Marzban #2059 | ستون/زیرمتن قابل انتخاب |
| حذف حساب‌های مرده | Marzban #1771 | retention policy + dry-run؛ بدون hard delete فوری |
| IP sharing detection | Marzban #1996 | detect-only ابتدا؛ CGNAT-aware |
| RBAC دقیق | Remnawave #332 | action + object scope، deny-by-default |
| Subscription backend enforcement | Remnawave #448 | access mode واقعی، نه صرفاً hide UI |
| WAL/write amplification | PasarGuard #786 | delta batch، bucket و idempotent flush |

## مواردی که عمداً کورکورانه اضافه نمی‌شوند

- CAPTCHA شخص ثالث به‌صورت اجباری
- بستن کاربر صرفاً با تغییر IP موبایل
- ادعای device limit قطعی با User-Agent
- plugin باینری ناشناس داخل process پنل
- raw JSON آزاد بدون validation
- Delete-all یک‌کلیکی بدون dry-run/backup/typed confirmation
- protocol جدید بدون accounting و compatibility declaration
- ترافیک مصنوعی برای سایت عمومی

## اولویت پیشنهادی

### P0

Accounting، Auth/RBAC، stable token، backup/restore، runtime atomic apply، mobile user table، audit.

### P1

Next Plan، deep-link، bulk، saved filters، rate policy، notification chain، domain separation.

### P2

Protocol SDK، multi-runtime، speed/device capabilities، billing/reseller integration، controller/node.

### P3

Marketplace/plugin sidecar، payment، Telegram sales automation و advanced routing.
