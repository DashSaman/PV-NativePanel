# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## وضعیت

PVNative دارای specification عمیق، اسکلت Go/React، سایت عمومی Static، مدل وضعیت کاربر و routeهای logs/diagnostics است. Runtime، DB، auth و business logic هنوز پیاده‌سازی نشده‌اند؛ Production-ready نیست.

## آخرین تغییرات

- ممیزی pinned از 3x-ui، PasarGuard، Marzban، Remnawave و OV-PvNetwork
- Account status، Presence، Quota state و Runtime health جدا شدند
- تک/چندکاربره با concurrency_limit
- رنگ و threshold حجم و تست state
- Dark/Light/System token
- Application/Runtime/Security logs و diagnostics routes
- Domain Activity owner-only، disabled-by-default و بدون path/query
- سایت عمومی Static و policy دانلود امن
- specification کامل Easy Installer
- تست public allowlist و owner-only بودن Domain Activity

## Scope

Standalone خارج، بدون ایران و بدون Controller در MVP. Controller/Node آینده اختیاری است.

## وضعیت تست

تست‌ها و CI تعریف شده‌اند، اما محیط scratch Go ندارد و GitHub status قابل مشاهده برنگردانده است. Passed فرض نشود. lockfile وب نیز باید پس از resolution بازبینی‌شده commit و CI به npm ci تغییر کند.

## کار بعدی دقیق

1. سبزکردن CI و lockfile
2. تکمیل Research Snapshot
3. PoC accounting بین Caddy forwardproxy و sing-box Naive
4. انتخاب Runtime Adapter
5. PostgreSQL schema: users، credentials، sessions، usage ledger، reset events، audit و logs metadata
6. Auth/session/MFA امن
7. سپس User CRUD و UI واقعی
8. در پایان Installer امضاشده و Pilot

## الزامات غیرقابل حذف

- status حساب و online یکی نشوند
- depleted/expired با متن و رنگ/آیکن جدا
- Domain Activity پیش‌فرض خاموش
- عدم TLS MITM
- optional collector نباید data plane را متوقف کند
- سایت عمومی ترافیک مصنوعی نسازد
- هیچ default password/token
- endpoint بدون Access ممنوع
- log بدون secret/token/query
