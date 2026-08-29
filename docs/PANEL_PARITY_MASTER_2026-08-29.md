# PVNaive — Panel Parity Master / Gap Analysis — 2026-08-29 Snapshot

> **HISTORICAL / SUPERSEDED**
>
> این فایل دیگر source-of-truth فعلی نیست. Snapshot مورخ 2026-08-29 قبل از merge شدن کامل WS1/WS2/WS3 و قبل از Production schema 11 تهیه شده بود و به همین دلیل چند قابلیت را Missing نشان می‌داد که اکنون واقعاً در `main` و Production وجود دارند.
>
> Matrix جاری و evidence-backed را از این فایل بخوان:
>
> `docs/PANEL_PARITY_MASTER_2026-08-30.md`

## چرا این snapshot superseded شد؟

در نسخه 2026-08-29 موارد زیر هنوز به‌عنوان Gap اصلی ثبت شده بودند:

- exact per-customer accounting؛
- trusted first-CONNECT producer؛
- hard-quota core؛
- groups/tags/plans/renewal/search/bulk؛
- deterministic `/sub` و human `/s` delivery؛
- schema جدید customer product management.

بعد از آن، PRهای WS1/WS2/WS3 و follow-upهای Production merge شدند. Audit مورخ 2026-08-30 نشان داد:

- Production schema = 11؛
- exact direct-Naive accounting event/session data زنده است؛
- تمام 6 ServiceTerm audited دارای accounting projection کامل بودند؛
- customer CRUD/product management و plans/groups/tags/renewal/search/bulk روی main وجود دارند؛
- `/sub/<token>` و `/s/<token>` و QR محلی وجود دارند؛
- first-CONNECT producer core و hard-quota reservation/settlement core merge شده‌اند.

بنابراین استفاده از جدول قدیمی 2026-08-29 برای تعیین TODO باعث duplicate work و regression می‌شود.

## چیزهایی از این snapshot که هنوز معتبرند

این اصول همچنان لازم‌الاجرا هستند:

1. Feature parity به معنی protocol parity نیست؛ PVNaive همچنان NaiveProxy-first است.
2. DB column، route stub یا UI placeholder به معنی implemented نیست.
3. Usage/remaining/online/device/speed هرگز نباید fabricated باشند.
4. قابلیت غیرقابل‌اثبات باید capability-gated باشد.
5. 3x-ui / PasarGuard / Hiddify با Licenseهای GPL/AGPL صرفاً behavior/UX/architecture reference هستند مگر اینکه compatibility صریحاً تأیید شود.
6. الگوهای عملیاتی OV-PvNetwork فقط در صورت سازگاری معماری و License قابل reuse هستند.
7. standalone correctness قبل از fleet/multi-node قرار دارد.

## Historical recovery

جزئیات کامل Matrix قدیمی برای forensic/history در Git history در `main@a021aa4b62c35b775fb521d042b2f8e6dbde10b0` و commitهای قبل از reconciliation باقی می‌ماند. برای تصمیم اجرایی جدید از نسخه 2026-08-30 استفاده کن.
