# سایت عمومی، Content Pack و Random Traffic

## تصمیم

سایت عمومی از پنل، Subscription و data plane جداست. محتوای آن با Content Pack نسخه‌دار تعویض می‌شود و هیچ موضوعی در binary، دیتابیس یا installer هاردکد نیست.

هر بسته می‌تواند title، locale، navigation، article، video، gallery، downloads، legal notice و branding override داشته باشد. Asset عمومی باید مجاز، دارای منبع/مجوز، SHA-256، نوع MIME مشخص و اسکن بدافزار باشد.

موضوع سیاسی درخواستی فقط می‌تواند یک Content Pack اختیاری باشد؛ پیش‌فرض پروژه خنثی است. تولید ترافیک جعلی، آمار بازدید ساختگی، ادعای روزانه ۲ ترابایت یا محتوای فریبنده خودکار جزو PVNaive نیست.

## Random data

NaiveProxy در پیاده‌سازی مرجع برای CONNECT/HTTP2 padding مذاکره‌شده و طول‌های شبه‌تصادفی دارد. بنابراین:

- بایت رندوم مستقل به stream اضافه نشود؛ کلاینت آن را payload می‌بیند یا اتصال خراب می‌شود.
- درخواست‌های پوششی زمان‌بندی‌شده و دائمی پیش‌فرض نباشند؛ هزینه، quota و fingerprint را بدتر می‌کنند.
- padding فقط از capability رسمی Runtime Adapter کنترل شود.
- هر آزمایش traffic shaping باید feature-flag، سقف پهنای‌باند، benchmark A/B، kill switch و accounting جدا داشته باشد.
- هیچ ادعای «غیرقابل‌شناسایی» ثبت نمی‌شود.

## Routing پیشنهادی

- `panel.example`: پنل مدیریتی، IP allowlist/VPN admin در حالت production
- `sub.example`: Subscription با rate limit و token
- `www.example`: سایت Static عمومی
- `proxy.example`: Naive data plane، DNS only

امکان هم‌دامنه‌شدن صرفاً بعد از PoC و بررسی تداخل Caddy/Naive مجاز است. public site نباید سلامت data plane را تعیین کند.

