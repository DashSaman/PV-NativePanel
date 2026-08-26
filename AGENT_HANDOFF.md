# Agent Handoff

آخرین به‌روزرسانی: 2026-08-26

## وضعیت

Repository از حالت خالی با اسناد تحقیق و معماری مقداردهی اولیه شد. هیچ backend، frontend، agent یا installer قابل اجرا هنوز وجود ندارد.

## تصمیم تثبیت‌شده

- Naive استاندارد مسیر اصلی
- Controller خارج از data path
- Agent کوچک outbound-only با mTLS
- standalone و multi-node از یک codebase
- چهار نود مستقیم خارج؛ ایران 100Mbps فقط degraded fallback
- PostgreSQL برای Production
- accounting per-credential یک PoC blocker

## کار بعدی دقیق

یک Research Snapshot بازتولیدپذیر بساز:

1. commit SHA/tag منابع upstream را ثبت کن.
2. semantics counter در Caddy forwardproxy fork و sing-box Naive را از کد مشخص کن.
3. یک harness آزمایش طراحی کن که userهای مجزا، H2 multiplex، reconnect، restart و counter reset را پوشش دهد.
4. نتیجه را در `docs/POC_ACCOUNTING_FA.md` ثبت کن.
5. قبل از انتخاب stack نهایی، benchmark و maintenance cost دو گزینه را مقایسه کن.

## اطلاعات مورد نیاز از مالک

- لینک/دسترسی سورس دقیق پنل Production فعلی PVNetwork، اگر غیر از OV-PvNetwork است
- نحوهٔ فعلی quota/reset/reseller
- ماتریس clientهای واقعی و درصد هر پلتفرم
- ظرفیت واقعی هر چهار نود و محدودیت ترافیک provider
- سیاست مطلوب همزمانی/device

## ریسک‌های باز

- نبود counter رسمی دقیق per credential در runtime مرجع
- رفتار fallback در clientهای مختلف
- سازگاری Ubuntu 26.04 با dependencyهای انتخابی
- ظرفیت 100Mbps ایران کمتر از بار متوسط
- rotation دامنه/IP بدون session storm

## قاعدهٔ تحویل

هر ایجنت پس از کار باید این فایل را به‌روز کند: چه چیزی تغییر کرد، چه چیزی تست شد، چه blockerی باقی است و قدم بعدی دقیق چیست.
