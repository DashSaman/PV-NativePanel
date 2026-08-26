# PVNative

پروژهٔ اختصاصی **PVNETWORK** برای مدیریت مستقل NaiveProxy روی سرور خارج، همراه یک سایت عمومی واقعی و امکان اتصال به پنل مرکزی در آینده.

> وضعیت: **اسکلت و specification مهندسی — غیرقابل استفاده در Production**.

## Scope

- Standalone خارج؛ بدون ایران در MVP
- NaiveProxy استاندارد
- کاربر تک‌اتصاله، چنداتصاله یا نامحدود با policy
- حجم، انقضا، reset دوره‌ای، on-hold، credential و session
- Online/Idle/Offline مستقل از Active/Expired/Depleted
- Dark/Light/System
- logs، audit، diagnostics و Domain Activity اختیاری
- سایت عمومی Static و دانلودهای مجاز
- Easy Installer پس از تکمیل Runtime و accounting

## اسکلت موجود

Go API با Route Registry و fail-closed، React/TypeScript UI، state model کاربر، Theme tokenها، سایت عمومی، تست مجوز routeها و security headers، و CI.

## اسناد اصلی

- [ممیزی عمیق پنل‌ها](docs/PANEL_DEEP_AUDIT_FA.md)
- [UI/UX و تمام وضعیت‌ها](docs/UI_UX_SPEC_FA.md)
- [Logging و Privacy](docs/OBSERVABILITY_PRIVACY_FA.md)
- [سایت عمومی](docs/PUBLIC_SITE_FA.md)
- [Easy Installer](docs/EASY_INSTALLER_FA.md)
- [مشخصات محصول](docs/PRODUCT_SPEC_FA.md)
- [قرارداد API](docs/API_FA.md)
- [Security Policy](SECURITY.md)
- [معماری](docs/ARCHITECTURE_FA.md)
- [نقشه راه](docs/ROADMAP_FA.md)
- [راهنمای ایجنت‌ها](AGENTS.md)
- [وضعیت ادامه](AGENT_HANDOFF.md)

## محدودیت صریح

Domain Activity نمی‌تواند صفحه و Path داخل HTTPS را ببیند؛ فقط مقصد CONNECT مثل hostname:port قابل مشاهده است. PVNative TLS MITM نمی‌کند. این collector پیش‌فرض خاموش و مستقل از دیتاپلین است.

قبل از quota واقعی، حسابداری byte هر credential زیر HTTP/2 multiplex، reconnect و restart باید اثبات شود. endpointهای تجاری تا آن زمان عمداً بسته‌اند.
