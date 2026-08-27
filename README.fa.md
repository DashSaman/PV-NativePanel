# PVNaive

پروژه اختصاصی **PVNETWORK** برای مدیریت NaiveProxy استاندارد، با معماری قابل‌گسترش برای Runtimeهای آینده.

> وضعیت فعلی: **Engineering preview؛ هنوز برای Production آماده نیست.**

## محدوده محصول

- Standalone روی سرور خارج؛ بدون وابستگی به Controller
- کاربر تک‌اتصاله، چنداتصاله و نامحدود
- حجم، انقضا، reset، خرید، اولین اتصال، credential و session
- صفحه Subscription با تاریخ خرید، روز باقی‌مانده و مصرف
- نقش owner/admin/reseller/operator/auditor با tenant isolation
- اعتبار و ledger نماینده
- اعلان 80/95/100 درصد حجم و 7/3/1 روز تا انقضا
- Dark/Light/System و Responsive دسکتاپ/موبایل
- audit، logs، diagnostics و backup/restore
- سایت عمومی Static و Content Pack قابل تعویض
- Runtime Adapter برای افزودن پروتکل‌های آینده

## وضعیت کد

API health قابل اجرا و Route Registry ثبت شده است، اما Auth، PostgreSQL، Accounting و عملیات تجاری هنوز پیاده‌سازی نشده‌اند. endpointهای ناقص fail-closed هستند و نصب Production تا سبز شدن gateها ممنوع است.

## سرور Pilot شناخته‌شده

`testAmir5-3`، IP `91.107.182.147`، Ubuntu 26.04، دامنه `namir.softarg.ir`، Caddy سفارشی v2.11.2 با forward_proxy روی 443. اسکریپت `scripts/preflight-testamir5-3.sh` فقط وضعیت را می‌خواند و هیچ تغییری ایجاد نمی‌کند.

## اسناد

- [مشخصات محصول](docs/PRODUCT_SPEC_FA.md)
- [معماری توسعه‌پذیر](docs/EXTENSIBILITY_FA.md)
- [طراحی PostgreSQL و S03](docs/DATABASE_FA.md)
- [ممیزی پنل‌ها](docs/PANEL_DEEP_AUDIT_FA.md)
- [UI/UX](docs/UI_UX_SPEC_FA.md)
- [امنیت](SECURITY.md)
- [Roadmap](docs/ROADMAP_FA.md)
- [Easy Installer](docs/EASY_INSTALLER_FA.md)
- [Handoff](AGENT_HANDOFF.md)

PVNaive TLS MITM نمی‌کند. Accounting تخمینی مبنای فروش نیست و ترافیک ساختگی تولید نمی‌شود.
