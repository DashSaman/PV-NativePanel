# PVNative

پروژهٔ اختصاصی **PVNETWORK** برای راه‌اندازی و مدیریت مستقل NaiveProxy روی سرور خارج، با امکان اتصال به پنل مرکزی در آینده.

> وضعیت فعلی: **اسکلت مهندسی اولیه — غیرقابل استفاده در Production**.

## Scope فعلی

- Standalone روی سرور خارج
- بدون سرور، تونل یا پهنای‌باند ایران
- NaiveProxy استاندارد
- مدیریت محلی کاربر، credential، حجم، انقضا و Subscription
- نصب آسان در مرحلهٔ بعد
- Controller/Node فقط به‌عنوان قابلیت آینده

## اسکلت موجود

- Go API با Route Registry مرکزی
- health endpoint و secure response headers
- fail-closed بودن تمام routeهای مدیریتی پیاده‌سازی‌نشده
- React/TypeScript UI با برند طلایی PVNative/PVNETWORK
- manifest صفحات و مجوز هر صفحه
- تست route تکراری، public allowlist، security headers و fail-closed
- CI برای Go test/vet/format و Web test/build

## صفحات طراحی‌شده

Dashboard، کاربران، جزئیات کاربر، Subscription، Runtime، حجم، وضعیت سیستم، Audit، امنیت و Backup.

صفحات Node، Fleet، Routing و Iran در MVP وجود ندارند.

## اسناد

- [هویت برند](docs/BRAND_FA.md)
- [مشخصات محصول و صفحات](docs/PRODUCT_SPEC_FA.md)
- [قرارداد API](docs/API_FA.md)
- [Security Policy](SECURITY.md)
- [معماری](docs/ARCHITECTURE_FA.md)
- [تحقیق تطبیقی](docs/RESEARCH_FA.md)
- [نقشه راه](docs/ROADMAP_FA.md)
- [راهنمای ایجنت‌ها](AGENTS.md)
- [وضعیت تحویل](AGENT_HANDOFF.md)

## مانع بعدی

قبل از پیاده‌سازی quota و فروش واقعی باید حسابداری byte برای هر credential زیر HTTP/2 multiplex، reconnect و restart اثبات شود. تا آن زمان endpointهای تجاری عمداً `501 Not Implemented` یا `401` برمی‌گردانند.
