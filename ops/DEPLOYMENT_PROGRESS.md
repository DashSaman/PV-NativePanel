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
| S03-DATABASE | NEXT | طراحی و اجرای PostgreSQL schema/migration | کد و تست محلی آماده؛ اجرای سرور فقط پس از سبزشدن CI مجاز است |
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

`S03-DATABASE` اکنون `NEXT` است. پس از سبزشدن CI، فقط بلوک کامل S03 اجرا و تمام خروجی آن در همین فایل ثبت شود. تا قبل از خروجی `S03_RESULT=PASSED`، S04 همچنان `BLOCKED` است.

## آماده‌سازی کد S03 — 2026-08-27

### اقدام و فایل‌ها

- Schema واقعی PostgreSQL نسخه 0001 برای ۲۷ جدول، ۲ view، RLS، ledger، outbox، runtime، audit/log/backup ساخته شد.
- `db/migrations/`، runner غیرمخرب، rollback gated، backup/restore age، health check و unit/timer سخت‌سازی‌شده اضافه شد.
- `scripts/stages/S03-database.sh` با preflight، backup قبل تغییر، SCRAM، loopback-only، secret تصادفی، migration، backup، restore drill و rollback failure-path آماده شد.
- context دیتابیس Go فقط `*sql.Tx` می‌پذیرد و tenant را از session hash معتبر bind می‌کند.
- `tests/db/migration_test.sh` و `tests/db/backup_restore_test.sh` و Job دیتابیس CI اضافه شدند.
- `web/package-lock.json` ثبت و CI frontend از `npm ci` استفاده می‌کند.
- نام‌های اجرایی/UI/docs باقیمانده به PVNaive اصلاح و cmd قدیمی `pvnative-api` حذف شد؛ نام Repository تغییر نکرد.

### خطاهای کشف‌شده و اصلاح

1. restore drill اولیه archive داخل directory با mode 0700 را به `pg_restore` تحت OS user دیگر می‌داد؛ علت permission traversal بود. archive اکنون از stdin بازشده توسط caller منتقل می‌شود.
2. flagهای rollback پس از بعضی mutationها set می‌شدند؛ failure می‌توانست HBA، role، secret، unit یا release نیمه‌ساخته باقی بگذارد. flagها پیش از mutation و cleanup restore-test DB اضافه شدند.
3. helper اولیه Go با interface از نظر type امکان دریافت `*sql.DB` داشت؛ API عمومی اکنون فقط `*sql.Tx` می‌پذیرد.
4. `auth_sessions.actor_id` می‌توانست بدون تطابق tenant تغییر کند؛ trigger و join دفاعی جلوی privilege escalation reseller→owner را می‌گیرند.
5. runner فایل Migration خارج manifest، gap نسخه و destructive SQL در میانه خط را کامل رد نمی‌کرد؛ preflight کامل قبل از اولین apply اضافه شد.

### تست محلی

- `bash -n scripts/db/*.sh scripts/stages/*.sh tests/db/*.sh`: PASSED
- `sha256sum --check --strict db/migrations/SHA256SUMS`: PASSED
- `git diff --check`: PASSED
- `npm ci --ignore-scripts --no-audit`: PASSED، 96 package
- `npm test`: PASSED، ۲ فایل و ۸ تست
- `npm run build`: PASSED، Vite production build
- Go و PostgreSQL integration در Runtime محلی موجود نبودند؛ نتیجه نهایی آن‌ها باید از GitHub Actions ثبت شود و تا آن زمان کد S03 روی سرور اجرا نمی‌شود.

### وضعیت استقرار و Rollback

- هیچ دستور S03 روی `testAmir5-3` اجرا نشده است.
- PostgreSQL نصب نشده و Caddy/NaiveProxy/SSH/Firewall تغییر نکرده‌اند.
- Rollback پیش‌بینی‌شده در ابتدای Stage چاپ می‌شود؛ packageها در failure برای inspection باقی می‌مانند، اما DB/role/secret/unit/release جدید حذف و config قبلی PostgreSQL restore می‌شود.
- Commit پیاده‌سازی: `45ba5c4c0c061a392a4e118ef99ae517d6eaead4` (`feat: implement guarded S03 PostgreSQL foundation`).

### خطای CI و تست جایگزین — 2026-08-27 01:59 UTC

- GitHub Actions run: `33031844663`، run number 61، conclusion=`failure`.
- هر سه Job با نام‌های `database`، `web` و `go` دارای `runner_id=0`، `runner_name=""` و `steps=[]` بودند؛ هیچ runner، step یا log کد ایجاد نشد. علت در سطح تخصیص runner/زیرساخت GitHub است و build/test failure محسوب نمی‌شود.
- برای جایگزینی، Go `1.24.4` رسمی Ubuntu در محیط محلی نصب و suite واقعی اجرا شد.
- اجرای اول Go دو failure واقعی نشان داد: فایل‌های قدیمی gofmt نبودند و `subscriptions.usage` در public-route allowlist تست نبود.
- اصلاح: `gofmt` روی تمام Go sourceها و افزودن route مورد انتظار به allowlist.
- اجرای دوم: `gofmt -l` خالی، `go vet ./...` PASSED و `go test ./...` برای `internal/database`، `internal/httpapi` و `internal/protocol` PASSED.
- تلاش نصب PostgreSQL server محلی به‌دلیل محدودیت user namespace محیط در configure پکیج `ssl-cert` با خطای `Cannot open audit interface` و failure ساخت OS group متوقف شد. از دورزدن permission خودداری شد؛ بنابراین database integration باید در Stage واقعی اجرا و PASSED شود.
- فایل‌های اصلاح‌شده: `cmd/pvnaive/main.go`، `internal/httpapi/routes.go`، `internal/httpapi/server_test.go`، `internal/protocol/adapter.go` و `internal/protocol/registry_test.go`.
- Commit اصلاح Go/ثبت failure: `fbf21e8fdc38a8c50bbc5704e407e9df4414171f` (`fix: format Go and align public route test`).
- GitHub Actions run دوم: `33032317065`، run number 62؛ هر سه Job دوباره `runner_id=0` و `steps=[]` و conclusion=`failure` داشتند. تکرار مستقل تأیید می‌کند blocker مربوط به تخصیص runner است، نه workflow step یا کد.
- وضعیت Stage بدون تغییر: `S03-DATABASE=NEXT`؛ S03 هنوز اجرا یا PASSED نشده است.
