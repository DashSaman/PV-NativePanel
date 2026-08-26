# قرارداد نصب آسان

این فایل مشخص می‌کند installer آینده چه تعهدی دارد. در وضعیت فعلی installer منتشر نشده است.

## رابط هدف

```bash
curl -fsSL https://example.invalid/install.sh | sudo bash -s -- standalone
curl -fsSL https://example.invalid/install.sh | sudo bash -s -- controller
curl -fsSL https://example.invalid/install.sh | sudo bash -s -- node --controller https://panel.example --token ONE_TIME_TOKEN
```

URL واقعی فقط پس از release امضاشده جایگزین می‌شود.

## الزامات

- پشتیبانی اولیه از Ubuntu LTS مشخص و pin‌شده؛ نه ادعای «همه نسخه‌ها»
- preflight برای CPU/RAM/disk/ports/DNS/time
- نمایش دقیق تغییرات firewall قبل از اعمال
- عدم بستن SSH
- package/release دارای checksum و signature
- نصب idempotent
- نگهداری config و secret در مسیر استاندارد با permission محدود
- systemd health و restart policy
- backup قبل از upgrade/migration
- rollback نسخه
- `doctor` برای DNS/TLS/port/runtime/controller/clock
- عدم چاپ token/password در خروجی
- uninstall بدون حذف خودکار DB/backup مگر با تأیید صریح

## خروجی نصب

- URL پنل یا شناسه نود
- وضعیت سرویس‌ها
- مسیر config و backup
- commandهای status/log/doctor/upgrade
- قدم بعدی امن

هر اسکریپتی که فقط Docker Compose را دانلود و اجرا کند، بدون preflight/rollback/verification، «نصب Production» محسوب نمی‌شود.
