# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## وضعیت

Repository دارای تحقیق و معماری اولیه است. هیچ backend، frontend، runtime adapter یا installer قابل اجرا هنوز وجود ندارد.

## آخرین تغییر Scope

مالک پروژه تصمیم گرفت:

- فعلاً پنل مرکزی کنار گذاشته شود.
- سرور، تونل و پهنای‌باند ایران کاملاً از Scope فعلی حذف شود.
- نسخهٔ اول یک پروژهٔ مستقل NaiveProxy روی سرور خارج باشد.
- فقط امکان اتصال اختیاری به پنل مرکزی برای آینده در معماری پیش‌بینی شود.

## تصمیم تثبیت‌شده

- Naive استاندارد
- Standalone-first
- مدیریت و enforcement محلی
- اتصال مستقیم کاربران به سرور خارج
- accounting per-credential یک PoC blocker
- اتصال Controller/Node بعداً و اختیاری
- هیچ fallback ایران در MVP

## کار بعدی دقیق

یک Research Snapshot و PoC حسابداری بازتولیدپذیر بساز:

1. commit SHA/tag منابع upstream را ثبت کن.
2. semantics counter در Caddy forwardproxy fork و sing-box Naive را از کد مشخص کن.
3. harness آزمایش userهای مجزا، H2 multiplex، reconnect، restart و counter reset را طراحی کن.
4. نتیجه را در `docs/POC_ACCOUNTING_FA.md` ثبت کن.
5. قبل از انتخاب Runtime، دقت آمار، fingerprint، performance و هزینه نگهداری را مقایسه کن.

## اطلاعات مورد نیاز از مالک در زمان تست

- مشخصات سرور خارجی PoC
- ماتریس clientهای واقعی و درصد هر پلتفرم
- سیاست quota/reset
- سیاست همزمانی/device
- تعداد کاربران pilot

اطلاعات پنل PVNetwork فعلی برای MVP مستقل blocker نیست.

## ریسک‌های باز

- نبود counter رسمی دقیق per credential در Runtime مرجع
- سازگاری Ubuntu 26.04 با dependencyهای انتخابی
- rotation دامنه/IP بدون session storm
- عملکرد clientهای iOS
- رسیدن یک سرور به ظرفیت هدف

## قاعدهٔ تحویل

هر ایجنت پس از کار باید این فایل را به‌روز کند: تغییرات، تست‌ها، blocker و قدم بعدی دقیق.
