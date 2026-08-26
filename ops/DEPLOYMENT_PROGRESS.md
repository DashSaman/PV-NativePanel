# PVNaive Deployment Progress — testAmir5-3

این فایل منبع اصلی وضعیت استقرار است و بعد از هر خروجی سرور باید به‌روزرسانی شود.

## قواعد ادامه کار

1. هر مرحله یک ID یکتا دارد.
2. قبل از مرحله بعد، خروجی مرحله فعلی ثبت می‌شود.
3. خطا پاک یا پنهان نمی‌شود؛ علت و راه‌حل کنار آن نوشته می‌شود.
4. هیچ تغییر Caddy، Firewall، SSH یا DB بدون Backup و Rollback انجام نمی‌شود.
5. اگر چت قطع شد، ابتدا این فایل و `AGENT_HANDOFF.md` خوانده شود.
6. فقط مرحله‌ای که وضعیت آن `NEXT` است اجرا شود.

## سرور هدف

| مورد | مقدار |
|---|---|
| Host | `testAmir5-3` |
| IPv4 | `91.107.182.147` |
| IPv6 | `2a01:4f8:c010:37ee::1` |
| OS | Ubuntu 26.04 LTS |
| Kernel | 7.0.0-29-generic |
| RAM | 3.7 GiB |
| Disk available | حدود 34 GiB |
| Domain | `namir.softarg.ir` |
| Runtime | Caddy v2.11.2 + forward_proxy |
| Service | `caddy-naive.service` |
| Caddyfile | `/etc/caddy/Caddyfile` |
| Caddyfile SHA-256 | `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1` |

## Stage ledger

| ID | وضعیت | شرح | نتیجه |
|---|---|---|---|
| S00-NAMING | PASSED | اصلاح PVNative به PVNaive در محصول | module، UI، API service و docs اصلاح شدند؛ نام Repository هنوز قدیمی است |
| S01-PREFLIGHT | PASSED | بررسی فقط‌خواندنی سرور | DNS، TLS، Caddy، ports، firewall و capacity سالم |
| S02-FOUNDATION | PASSED | Backup محلی و ساخت directory/user پایه | backup و checksum سالم؛ user/directoryها ساخته شدند |
| S03-DATABASE | NEXT | طراحی و اجرای PostgreSQL schema/migration | قبل از اجرا باید کد، rollback و تست آماده شود |
| S04-AUTH | BLOCKED | bootstrap owner، session، MFA و RBAC | منتظر S03 |
| S05-USERS | BLOCKED | User/Plan/Reseller CRUD | منتظر S04 |
| S06-RUNTIME | BLOCKED | Atomic Caddy adapter و accounting PoC | منتظر S05 |
| S07-SUBS | BLOCKED | Subscription page و renderer | منتظر S06 |
| S08-NOTIFY | BLOCKED | اعلان حجم/انقضا و outbox | منتظر S07 |
| S09-INSTALLER | BLOCKED | installer/upgrade/rollback نهایی | پس از gateهای قبلی |
| S10-PILOT | BLOCKED | canary روی همین سرور | پس از تست کامل |

## گزارش S01-PREFLIGHT — 2026-08-26 20:12 UTC

### موفق

- SSH نهایی روی port 22 برقرار شد.
- `namir.softarg.ir → 91.107.182.147`.
- Caddy روی 80 و 443 گوش می‌دهد.
- moduleهای `http.handlers.forward_proxy` و `file_server` موجودند.
- `caddy-naive.service` فعال است.
- Caddyfile معتبر است.
- TLS خودکار Let’s Encrypt موفق است.
- UFW فقط 22/80/443 را باز گذاشته است.
- warning/error اخیر Caddy وجود ندارد.
- 34GiB disk و 3.3GiB memory available است.

### خطاها و هشدارها

- دو بار `Access denied` قبل از ورود موفق SSH: احتمال password اشتباه اولیه؛ blocker نیست.
- `Caddyfile input is not formatted`: فقط formatting؛ config valid است. فعلاً تغییر داده نشود.
- Swap صفر است: برای Pilot blocker نیست؛ قبل از PostgreSQL و load test دوباره ارزیابی شود.
- 30 update موجود است که 23 مورد security است: قبل از Production باید maintenance window و reboot impact مشخص شود.

## گزارش S02-FOUNDATION — 2026-08-26 20:18:57 UTC

- نتیجه: `PASSED`
- Backup: `/var/backups/pvnaive/20260826T201857Z`
- checksum فایل‌های Caddyfile، service و public-site: همگی `OK`
- user: `pvnaive` با UID 995 و GID 982
- `/opt/pvnaive`: `750 root:pvnaive`
- data/secrets: `700 pvnaive:pvnaive`
- `/etc/pvnaive`: `750 root:pvnaive`
- Caddy پس از مرحله همچنان `active`
- portهای 22/80/443 بدون تغییر
- Caddy restart/reload، Firewall/SSH change و package install انجام نشد.
- هشدار formatting قبلی Caddyfile تکرار شد؛ config معتبر است و تغییری داده نشد.

## مرحله بعد

`S03-DATABASE` اکنون `NEXT` است، اما تا زمانی که schema، migration، rollback، secret handling و تست در Repository آماده و CI بررسی نشده، هیچ دستور نصب PostgreSQL روی سرور صادر نشود.
