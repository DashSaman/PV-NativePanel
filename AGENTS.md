# AGENTS.md — دستور کار اجباری ایجنت‌ها

## مأموریت

PV NativePanel یک control plane برای NaiveProxy استاندارد است که هم standalone و هم controller/node را پشتیبانی می‌کند. هدف فعلی ساخت محصول Production برای حدود 400 اتصال همزمان، چهار نود خارجی و fallback محدود ایران است.

## قبل از هر تغییر

1. `README.fa.md`، `docs/ARCHITECTURE_FA.md`، `docs/DECISIONS_FA.md`، `docs/ROADMAP_FA.md` و `AGENT_HANDOFF.md` را کامل بخوان.
2. وضعیت repo، branch و تغییرات موجود را بررسی کن.
3. تغییرات کاربر یا ایجنت دیگر را overwrite نکن.
4. ادعای «آماده Production» یا «ضدفیلتر تضمینی» نکن.
5. اگر تصمیم معماری عوض شد، DECISIONS و HANDOFF را همان commit به‌روز کن.

## خطوط قرمز

- wire protocol Naive را بدون ADR، benchmark و client compatibility fork نکن.
- accounting را از access log تخمینی برای billing نهایی استفاده نکن.
- secret واقعی، IP حساس، password، token یا subscription مشتری را commit نکن.
- data plane را به availability پنل وابسته نکن.
- SSH را در installer نبند.
- destructive migration/uninstall بدون backup و تأیید نساز.
- Cloudflare orange proxy را پیش‌فرض data plane معرفی نکن.
- از `latest` بدون version pin در Production استفاده نکن.

## ترتیب پیاده‌سازی

1. PoC accounting و atomic config
2. schema/API tests
3. Agent reconcile و rollback
4. Subscription و assignment
5. UI
6. installer
7. canary migration

## تعریف Done

- test خودکار و failure-path وجود دارد
- migration/rollback مستند است
- metrics و log فاقد secret است
- docs فارسی با رفتار واقعی یکی است
- HANDOFF شامل آخرین commit، کار انجام‌شده، ریسک و قدم بعدی به‌روز شده است

## استاندارد commit

commit کوچک و موضوعی؛ پیام انگلیسی imperative. تغییر schema همراه migration و test. تغییر API همراه contract. تغییر runtime همراه benchmark و rollback.

## وقتی ادامهٔ کار مبهم است

به ترتیب ROADMAP جلو برو. اولویت فعلی همیشه در AGENT_HANDOFF است. اگر دادهٔ لازم نیست، فرض پنهان نساز؛ blocker را ثبت کن.
