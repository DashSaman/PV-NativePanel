# Easy Installer — مشخصات اجرایی

> هنوز اسکریپت Production منتشر نشده است. این سند معیار پذیرش Installer آینده است.

## تجربه هدف

```bash
curl -fsSL https://RELEASE_HOST/install.sh -o /tmp/pvnative-install.sh
curl -fsSL https://RELEASE_HOST/install.sh.sha256 -o /tmp/pvnative-install.sh.sha256
sha256sum -c /tmp/pvnative-install.sh.sha256
sudo bash /tmp/pvnative-install.sh install
```

اجرای مستقیم `curl | bash` در مستند رسمی پیشنهاد نمی‌شود؛ ابتدا checksum/signature بررسی می‌شود.

## Wizard

Installer این موارد را مرحله‌ای می‌پرسد:

1. Domain عمومی سایت
2. Domain یا access policy پنل مدیریت
3. ایمیل TLS
4. DB profile
5. مسیر backup
6. timezone و واحد حجم
7. فعال‌سازی firewall فقط پس از نمایش diff
8. ساخت Owner با setup link یک‌بارمصرف؛ بدون default password

## Preflight

- OS/version/architecture
- RAM/CPU/disk و inode
- clock/NTP
- DNS A/AAAA
- اشغال پورت‌های 80/443 و پورت مدیریت
- outbound connectivity لازم
- وضعیت UFW/nftables
- Docker/Podman یا systemd conflict
- SELinux/AppArmor
- backup target
- وجود installation قبلی

Fail شدن preflight قبل از mutation پایان می‌یابد.

## اجزای نصب

- `pvnative-api`
- `pvnative-web`
- `pvnative-runtime`
- PostgreSQL یا اتصال خارجی
- Caddy سفارشی version-pinned
- static public site
- logrotate/journald limits
- backup timer
- health/doctor command

## مدیریت

```bash
pvnative status
pvnative doctor
pvnative logs api
pvnative logs runtime
pvnative backup
pvnative restore --verify-only FILE
pvnative update --check
pvnative update VERSION
pvnative rollback
pvnative uninstall
```

## Atomic install/update

download → checksum/signature → unpack staging → migrate check → backup → validate config → switch release symlink → health check → commit. در شکست، symlink و DB در صورت سازگاری rollback می‌شوند. Migration غیرقابل برگشت بدون تأیید و maintenance window ممنوع است.

## حفاظت SSH و شبکه

- Installer هیچ rule مربوط به SSH را حذف نمی‌کند.
- قبل از firewall change، پورت SSH فعلی از socket/config تشخیص داده و allow می‌شود.
- session فعال SSH تا پایان smoke test حفظ می‌شود.
- تغییر default route، policy routing یا interface در MVP وجود ندارد.
- site/Naive روی 443 بدون دست‌زدن به مسیر مدیریت نصب می‌شوند.

## Idempotency

اجرای دوباره نباید Owner، secret، certificate، DB یا log policy را بدون درخواست rotate کند. هر step دارای detect/apply/verify است.

## Secret

- تولید با CSPRNG
- permission حداقل
- عدم چاپ در terminal history/log
- setup token یک‌بارمصرف و کوتاه‌عمر
- backup رمزنگاری‌شده
- نمایش redacted در doctor bundle

## Smoke test

- public site پاسخ 200
- health live/ready
- panel از policy مجاز قابل دسترس
- subscription headerهای no-store/noindex
- Runtime config valid
- probe نامعتبر وب عادی می‌بیند
- credential آزمایشی فقط در mode test
- restart و last-known-good
- backup verify

## Uninstall

پیش‌فرض binary/service را متوقف می‌کند ولی DB، backup، certificate و public files را نگه می‌دارد. حذف داده فقط با flag صریح، نمایش مسیر دقیق و تأیید دومرحله‌ای.
