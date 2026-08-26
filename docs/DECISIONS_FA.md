# تصمیم‌های معماری

## ADR-001 — Wire protocol اختصاصی نمی‌سازیم

**تصمیم:** NaiveProxy استاندارد و کلاینت‌های سازگار حفظ می‌شوند.

**دلیل:** fork کردن wire behavior ریسک fingerprint تازه، ناسازگاری client و بار نگهداری همگام با Chromium را بالا می‌برد. اختصاصی‌بودن در control plane، policy و عملیات خواهد بود.

## ADR-002 — Controller در مسیر دیتا نیست

قطع پنل نباید دیتاپلین را قطع کند. Node با last-known-good کار می‌کند و enforcement quota محلی دارد.

## ADR-003 — Agent کوچک به‌جای panel-to-panel

هر نود یک Agent خروجی‌محور با mTLS دارد. این مدل سطح حمله و هزینهٔ upgrade را نسبت به نصب پنل کامل روی هر نود کم می‌کند.

## ADR-004 — PostgreSQL مبنای Production

SQLite فقط development یا standalone کوچک. برای ledger، lock، job و گزارش حجمی production از PostgreSQL استفاده می‌شود.

## ADR-005 — حسابداری PoC blocker است

تا زمانی که byte counter per-credential زیر restart، reconnect، multiplex و failure آزمون نشده، quota به‌عنوان قابلیت آماده علامت نمی‌خورد.

## ADR-006 — ایران fallback تنزل‌یافته است

100Mbps جایگزین کامل بار روزانهٔ هدف نیست. ایران فقط با rate cap، TTL و بازگشت تدریجی فعال می‌شود.

## ADR-007 — Cloudflare فقط برای control surfaces

Panel/subscription می‌توانند پشت Cloudflare باشند. data-plane Naive به‌صورت پیش‌فرض DNS-only است.

## ADR-008 — secrets قابل بازیابی در log نیستند

credential plaintext فقط هنگام صدور/rotation و delivery لازم دیده می‌شود؛ در DB تا حد ممکن hash یا envelope encryption استفاده می‌شود.
