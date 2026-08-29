# PVNaive feature parity review — 2026-08-29

> **Canonical update:** این فایل یک snapshot قبلی از بررسی همان روز است. برای مقایسه‌ی فعلی و جامع 3x-ui/Sanaei، PasarGuard v5.3.0، Hiddify Manager v12.3.3 و OV-PvNetwork و همچنین backlog چهار-lane، از `docs/PANEL_PARITY_MASTER_2026-08-29.md` استفاده کنید. Promptهای اجرای موازی در `docs/PARALLEL_WORKSTREAM_PROMPTS_2026-08-29.md` هستند.
>
> نکته: توضیح قدیمی پایین درباره‌ی یک URL مشترک browser/raw یک gap شناخته‌شده‌ی `main` است؛ قرارداد هدف جدید `/sub/<token>` برای machine و `/s/<token>` برای human است.

این سند مقایسه‌ی عملی PVNaive با پنل‌های مرجع مدیریت پراکسی است. هدف کپی‌کردن 3x-ui/PasarGuard نیست؛ هدف این است که UX مدیریت مشتری در PVNaive از نظر عملیات روزمره عقب‌تر نباشد، در حالی که Runtime اصلی همچنان NaiveProxy و قواعد امنیتی PVNaive حفظ شوند.

## منابع مرجع

- 3x-ui official repository: https://github.com/3x-ui/3x-ui
- MHSanaei/3x-ui wiki: https://github.com/MHSanaei/3x-ui/wiki/Home
- PasarGuard panel: https://github.com/PasarGuard/panel
- Marzban upstream/family reference: https://github.com/Gozargah/Marzban

طبق مستندات رسمی 3x-ui، قابلیت‌های کلیدی شامل مدیریت per-client، quota/expiry، online/IP limit، share/QR/subscription، traffic statistics/reset، multi-node، Telegram و REST API است. PasarGuard نیز traffic/expiry، periodic limits، HWID/device limits، QR/subscription، monitoring، Telegram و RBAC را به‌عنوان قابلیت‌های اصلی اعلام می‌کند.

## وضعیت فعلی PVNaive

| قابلیت | PVNaive | وضعیت نسبت به پنل‌های مرجع |
|---|---|---|
| لیست واحد همه اکانت‌ها | ✅ | در این release اکانت‌های Runtime قدیمی و مشتریان جدید در یک directory واحد هستند |
| Search / filter / sort / pagination | ✅ | هم‌سطح نیاز عملی روزمره |
| Create customer | ✅ | حجم، اعتبار، password خودکار/دستی |
| Adopt existing Runtime account | ✅ | بدون تغییر Username/Password فعلی |
| Edit total quota / validity | ✅ | Set total و سه سیاست شروع اعتبار |
| Add volume / extend days | ✅ | عملیات مستقل از Runtime mutation |
| Suspend / Resume | ✅ | با حفظ Runtime UUID |
| Safe delete / revoke | ✅ | history-preserving؛ حذف فیزیکی عادی نیست |
| Password rotation مستقل | ✅ | Subscription token را rotate نمی‌کند |
| Read-only current Subscription / QR | ✅ | View/Copy هیچ mutationی ندارد |
| Explicit subscription reissue | ✅ | operation جدا و آگاهانه |
| Branded subscription status page | ✅ | browser HTML + QR + quota + expiry؛ raw clients همان URI را می‌گیرند |
| Local QR generation | ✅ | هیچ token/URL حساسی به سرویس ثالث ارسال نمی‌شود |
| Dashboard KPI / status chart / expiry chart | ✅ | از داده واقعی business/runtime ساخته می‌شود |
| Exact per-user traffic used/remaining | ⛔ Proof-gated | هنوز منبع packet-level معتبر از pinned Naive/Caddy نداریم |
| Hard quota enforcement | ⛔ Proof-gated | تا exact accounting اثبات نشود فعال نمی‌شود |
| Traffic reset / periodic reset | ⛔ Proof-gated | وابسته به accounting واقعی |
| First-use automatic activation | 🟡 Receiver ready | producer معتبر CONNECT-success هنوز live-proven نیست |
| Live online status / sessions | ⛔ Not proven | نباید از request/sub fetch استنتاج شود |
| Per-client IP limit | ⛔ Not implemented | نیازمند enforceable runtime evidence |
| HWID/device limit | ⛔ Not implemented | Naive استاندارد چنین identity قابل اتکایی ارائه نمی‌دهد |
| Multi-node central management | 🟡 Planned | runtime safety model آماده‌ی توسعه است، fleet UI هنوز کامل نیست |
| Telegram bot / notifications | 🟡 Planned | notification schema وجود دارد؛ delivery UX مرحله بعد |
| Reseller / RBAC UI | 🟡 Foundation | schema/auth roles وجود دارد؛ reseller operations UI کامل نشده |
| REST API | 🟡 Partial | API عملیاتی وجود دارد؛ public OpenAPI/Swagger هنوز کامل نشده |
| Backup / rollback safety | ✅ | encrypted DB backup + migration/checksum/rollback gates |
| Dark/light theme | ✅ | theme پایه موجود است؛ owner UI با هر دو theme سازگار طراحی شده |
| Multi-language UI | 🟡 Partial | محصول فعلاً فارسی‌محور است؛ i18n کامل مانند 3x-ui هنوز انجام نشده |
| Multi-protocol | N/A by product design | PVNaive عمداً NaiveProxy-first است؛ breadth پروتکل هدف این محصول نیست |

## تصمیم UX این release

1. برچسب و باکس `Legacy Runtime` حذف شد. تفاوت قدیمی/جدید برای Owner یک مفهوم داخلی است و نباید layout را دو تکه کند.
2. هر Runtime account بدون business term در همان جدول با وضعیت «نیازمند تنظیم» دیده می‌شود و با یک action به customer management وارد می‌شود.
3. دکمه‌های ردیف به سه action اصلی `QR/لینک`، `ویرایش` و `جزئیات` محدود شدند؛ عملیات حساس/کم‌استفاده در منوی بیشتر قرار گرفتند.
4. Dashboard از KPI و نمودارهای قابل اثبات استفاده می‌کند: account state، expiry horizon و configured quota. نمودار traffic جعلی اضافه نشده است.
5. Subscription URL دو رفتار سازگار دارد: مرورگر صفحه‌ی status برنددار می‌بیند؛ clientهایی مانند Karing با raw request همان Naive URI را دریافت می‌کنند.

## Gapهای اولویت‌دار بعدی

### P0 — قبل از ادعای traffic/quota واقعی

- patch حداقلی و reviewشده برای pinned `caddy2-naive/forwardproxy` جهت emission شمارنده upload/download به Runtime Agent.
- cumulative counter + runtime boot ID + authenticated Unix socket/datagram ingestion.
- اثبات reconnect/restart semantics و جلوگیری از double count.
- فقط بعد از آن: used/remaining، progress bar واقعی، hard quota، reset strategy و depletion automation.

### P1 — parity عملیاتی با 3x-ui/PasarGuard

- online/presence فقط از session evidence معتبر.
- Telegram notifications برای expiry/quota/runtime health.
- reseller management UI و scoped RBAC.
- multi-node/fleet dashboard و node health.
- audit explorer و diagnostics در UI.
- OpenAPI/Swagger و webhooks.
- saved plan presets/defaults برای ساخت سریع سرویس‌های 30/50/80/100GB.

### P2 — فقط در صورت enforceability

- concurrent-session limits.
- IP/device/HWID constraints.
- speed policy.
- periodic quota reset.

## اصل غیرقابل نقض

PVNaive نباید برای رسیدن ظاهری به feature parity، `used=0`، `online=false` یا remaining ساختگی تولید کند. هر metric عملیاتی باید producer، persistence semantics، failure behavior و enforcement path قابل اثبات داشته باشد.
