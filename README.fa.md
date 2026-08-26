# PV NativePanel

کنترل‌پلین اختصاصی برای مدیریت **NaiveProxy** در حالت تک‌سرور و چندنود.

> وضعیت فعلی: **مرحلهٔ صفر — تحقیق و طراحی معماری**. هنوز نسخهٔ قابل نصب یا Production منتشر نشده است.

## هدف

- اتصال مستقیم کاربران به نودهای خارج در حالت عادی
- نگه‌داشتن ورودی ایران فقط به‌عنوان fallback محدود
- مدیریت متمرکز کاربر، حجم، انقضا، نود، Host، Subscription و مانیتورینگ
- استفاده از NaiveProxy استاندارد؛ بدون ساخت Wire Protocol اختصاصی و ناسازگار
- نصب یک‌فرمانی در آینده برای حالت `standalone` و `controller/node`

## تصمیم اولیه

مسیر اصلی، Naive روی HTTPS/HTTP2 و پورت 443 است. کنترل‌پلین و دیتاپلین از هم جدا می‌شوند:

- **Controller:** API، پنل، دیتابیس، حسابداری، Subscription و Scheduler
- **Node Agent:** اعمال desired state، health، آمار و کنترل lifecycle
- **Naive Runtime:** Caddy forwardproxy/Naive استاندارد
- **Fallback:** ورودی ایران با ظرفیت 100Mbps فقط در شرایط اختلال
- **Compatibility:** خروجی اختیاری Xray/VLESS برای مشتریان قدیمی، نه مسیر اصلی

جزئیات در [معماری](docs/ARCHITECTURE_FA.md)، [تحقیق تطبیقی](docs/RESEARCH_FA.md) و [نقشه راه](docs/ROADMAP_FA.md) آمده است.

## مقیاس هدف فعلی

- حدود 400 اتصال همزمان
- حداکثر حدود 2TB مصرف روزانه (میانگین تقریبی 185Mbps)
- چهار نود خارجی در شروع
- یک مسیر ایران 100Mbps برای fallback تنزل‌یافته

## هشدار ظرفیت

100Mbps حداکثر نظری حدود 1.08TB در روز است و برای بار 2TB/day جایگزین کامل نیست. با 316 کاربر همزمان، سهم متوسط هر کاربر حدود 316Kbps می‌شود؛ بنابراین fallback باید degraded mode، صف‌بندی یا محدودیت سرعت داشته باشد.

## منابع و ادامه کار

- [ماتریس قابلیت‌ها](docs/FEATURE_MATRIX_FA.md)
- [تصمیم‌های معماری](docs/DECISIONS_FA.md)
- [قرارداد نصب آسان](docs/INSTALLER_CONTRACT_FA.md)
- [راهنمای ایجنت‌ها](AGENTS.md)
- [وضعیت تحویل](AGENT_HANDOFF.md)
