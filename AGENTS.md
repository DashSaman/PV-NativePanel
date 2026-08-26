# AGENTS.md — دستور کار اجباری ایجنت‌ها

## مأموریت فعلی

PV NativePanel فعلاً یک پروژهٔ **Standalone-first** برای NaiveProxy استاندارد روی سرور خارج است. پنل مرکزی، چندنود و سناریوی ایران بخشی از MVP نیستند. فقط boundary لازم برای اتصال اختیاری آینده حفظ می‌شود.

## قبل از هر تغییر

1. `README.fa.md`، `docs/ARCHITECTURE_FA.md`، `docs/DECISIONS_FA.md`، `docs/ROADMAP_FA.md` و `AGENT_HANDOFF.md` را کامل بخوان.
2. وضعیت repo و تغییرات موجود را بررسی کن.
3. تغییرات دیگران را overwrite نکن.
4. ادعای «Production-ready» یا «ضدفیلتر تضمینی» نکن.
5. تغییر تصمیم معماری باید همراه به‌روزرسانی DECISIONS و HANDOFF باشد.

## خطوط قرمز

- MVP را به Controller، Node fleet یا ایران وابسته نکن.
- wire protocol Naive را بدون ADR، benchmark و client compatibility تغییر نده.
- access log تخمینی را مبنای billing نهایی نکن.
- secret، password، token یا subscription واقعی commit نکن.
- Web UI را در مسیر availability دیتا قرار نده.
- SSH را در installer نبند.
- migration یا uninstall مخرب بدون backup نساز.
- Cloudflare orange proxy را پیش‌فرض data plane معرفی نکن.
- از نسخهٔ `latest` بدون pin استفاده نکن.

## ترتیب پیاده‌سازی

1. Research snapshot
2. PoC accounting و atomic config
3. Standalone schema/API/Runtime Adapter
4. UI و Subscription
5. installer
6. pilot و benchmark
7. اتصال اختیاری پنل در فاز بعد

## تعریف Done

- test خودکار و failure-path وجود دارد
- migration و rollback مستند است
- metrics/log فاقد secret است
- مستند فارسی با رفتار واقعی یکی است
- HANDOFF بعد از هر تغییر مهم به‌روز شده است

## وقتی ادامهٔ کار مبهم است

طبق ROADMAP و سپس AGENT_HANDOFF جلو برو. اگر داده‌ای موجود نیست، blocker ثبت کن و فرض پنهان نساز.
