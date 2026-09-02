# Agent Handoff — PVNaive

آخرین به‌روزرسانی: 2026-08-27 23:50 UTC

این فایل مرجع اصلی ادامه کار برای هر Agent/Chat جدید است. قبل از هر اقدام روی سرور، این فایل و `ops/DEPLOYMENT_PROGRESS.md` را کامل بخوان. هیچ Stage را از روی حدس جلو نبر.

## خلاصه فوری برای Chat جدید

- Repository: `DashSaman/PV-NativePanel`
- نام صحیح محصول: `PVNaive`
- نام قدیمی/اشتباه: `PVNative`
- نام Repository فعلاً قدیمی مانده و نباید بدون هماهنگی rename شود.
- branch فعال توسعه: `s04-auth`
- PR فعال: `#2` با عنوان `S04-AUTH: production authentication foundation`
- base: `main`
- S00 تا S03: PASSED
- S04-AUTH: در حال استقرار زنده، هنوز PASSED نشده
- S05 تا S10: BLOCKED تا S04 کامل و postflight شود
- هدف فوری: علت واقعی `pvnaive-api.service is not active` را از systemd journal پیدا و رفع کن؛ سپس S04 recovery را دوباره اجرا کن.
- روی Caddy/NaiveProxy، SSH و Firewall تا پایان localhost gate هیچ تغییری نده.

## قانون کار با سرور

کاربر دستورات را دستی به‌صورت root روی `testAmir5-3` اجرا می‌کند و خروجی کامل را paste می‌کند. Agent باید هر بار فقط یک مرحله/دستور واضح بدهد و بعد از دیدن خروجی تصمیم بگیرد. از heredoc/base64 بسیار بزرگ برای انتقال فایل استفاده نکن؛ artifact واقعی upload شود و SHA-256 آن verify شود.

هیچ Stage تا زمانی که خود Stage و یک postflight مستقل هر دو PASS نشده‌اند، `PASSED` اعلام نشود.

## سرور هدف

- Host: `testAmir5-3`
- IPv4: `91.107.182.147`
- IPv6: `2a01:4f8:c010:37ee::1`
- OS: Ubuntu 26.04 LTS
- Kernel: `7.0.0-30-generic`
- RAM: حدود 3.7 GiB
- Domain: `namir.softarg.ir`
- Caddy: v2.11.2 + `forward_proxy`
- Service: `caddy-naive.service`
- Caddyfile: `/etc/caddy/Caddyfile`
- Caddyfile SHA-256 ثابت و مورد انتظار:
  `101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1`
- پورت‌های baseline: `22`, `80`, `443`
- PostgreSQL: `18/main`, port `5432`, فقط `127.0.0.1` و `::1`

## Stage ledger فعلی

| Stage | Status | توضیح |
|---|---|---|
| S00-NAMING | PASSED | نام محصول به PVNaive اصلاح شده |
| S01-PREFLIGHT | PASSED | DNS/TLS/Caddy/ports/capacity بررسی شد |
| S02-FOUNDATION | PASSED | user/directory/backup پایه ساخته شد |
| S03-DATABASE | PASSED | PostgreSQL 18 + schema v1 + RLS + backup/restore/health + postflight مستقل |
| S04-AUTH | IN PROGRESS | کد/CI تقریباً کامل؛ استقرار localhost درگیر service startup blocker |
| S05-USERS | BLOCKED | بعد از S04 |
| S06-RUNTIME | BLOCKED | بعد از S05 |
| S07-SUBS | BLOCKED | بعد از S06 |
| S08-NOTIFY | BLOCKED | بعد از S07 |
| S09-INSTALLER | BLOCKED | بعد از S08 |
| S10-PILOT | BLOCKED | بعد از همه gateها |

## S02 نهایی

- PASSED: `2026-08-26 20:18:57 UTC`
- Backup: `/var/backups/pvnaive/20260826T201857Z`
- user `pvnaive`: UID 995 / GID 982
- `/opt/pvnaive`: `750 root:pvnaive`
- data/secrets: `700 pvnaive:pvnaive`
- `/etc/pvnaive`: `750 root:pvnaive`
- Caddy/SSH/Firewall تغییر نکرد.
- marker: `/opt/pvnaive/FOUNDATION.json`

## S03 نهایی — PASSED

S03 production و postflight مستقل هر دو موفق شدند.

- S03 marker: `/opt/pvnaive/S03_DATABASE.json`
- PostgreSQL: `18/main`, port `5432`, loopback-only
- DB: `pvnaive`
- roles: `pvnaive_owner`, `pvnaive_app`
- schema version نهایی S03: `1`
- health با `pvnaive_app`: PASSED
- direct secret SELECT: DENIED
- encrypted backup: PASSED
- restore drill: PASSED
- systemd DB health timer: enabled + active
- Caddy checksum بعد از S03 بدون تغییر ماند.
- SSH/Firewall تغییر نکرد.

S03 postflight نهایی در خروجی واقعی داشت:

```text
S03_POSTFLIGHT=PASSED
S03_DATABASE=PASSED
NEXT_STAGE=S04-AUTH
NO_CONFIGURATION_CHANGES_MADE=true
```

## معماری S04-AUTH که تصویب و پیاده‌سازی شده

- Password hashing: Argon2id
- opaque session token؛ raw token در DB ذخیره نمی‌شود، فقط SHA-256
- secure browser cookie
- CSRF binding برای state-changing requests
- RBAC برای Owner/Admin/Operator/Auditor/Reseller
- TOTP MFA با AES-256-GCM برای secret
- recovery code hash-only
- MFA replay-step protection
- login lockout
- session rotation/revoke/reuse detection
- `/api/v1/me`
- owner bootstrap بدون default password
- owner bootstrap باید local/root/TTY باشد
- هیچ default password/token وجود ندارد
- auth audit بدون secret/token/password
- API ابتدا فقط روی `127.0.0.1:8080`
- Caddy exposure فقط بعد از localhost auth gate
- migration S04 = `0002_auth_foundation`
- rollback S04 باید v2→v1 باشد و S03 را سالم نگه دارد

## CI / توسعه S04

branch: `s04-auth`
PR: `#2`

یک build کاملاً سبز قبل از استقرار زنده روی commit زیر ثبت شد:

- source commit: `b4803e27af36bb35de33f7dcbe39750aeadc4146`
- GitHub Actions run: `33126898878`
- Go: PASSED
- Web tests/build: PASSED
- PostgreSQL 18: PASSED
- end-to-end auth rehearsal: PASSED
- bundle build: PASSED

rehearsal واقعی CI شامل PostgreSQL 18، migration v2، binary واقعی، login owner، session cookie، `/me`، CSRF logout و revoke session بود.

### Bundle دقیق استفاده‌شده روی سرور

Artifact GitHub:
`PVNaive-S04-b4803e27af36bb35de33f7dcbe39750aeadc4146`

فایل داخلی:
`PVNaive-S04-b4803e27af36.tar.gz`

SHA-256:
`c279c91f42ca7d0b91096c48875fa61be8e8c80c8effa826aa477314a9732147`

artifact روی سرور به‌صورت ZIP upload شد:
`/root/PVNaive-S04-b4803e27af36-artifact.zip`

extract امن انجام شد و tar داخلی byte/checksum verified شد.

bundle فعلی روی سرور:
`/root/pvnaive-s04-deploy-b4803e27af36/PVNaive-S04-b4803e27af36`

تمام `S04_SHA256SUMS` داخل bundle PASS شد.

## تلاش‌های زنده S04 روی testAmir5-3

### Attempt 1 — missing `file`

زمان: حدود `2026-08-27 23:44 UTC`

Stage قبل از mutation اصلی با این خطا fail شد:

```text
line 179: file: command not found
ERROR: pvnaive binary architecture mismatch
S04_RESULT=FAILED
ROLLBACK=COMPLETED
```

علت: package کوچک `file` روی Ubuntu نصب نبود ولی Stage برای verify معماری ELF به آن نیاز داشت.

بعداً `file 1:5.46-5build2` نصب شد و هر دو binary درست تشخیص داده شدند:

```text
ELF 64-bit LSB executable, x86-64, statically linked, stripped
```

پس این blocker بسته شد.

### Attempt 2 — backup timestamp collision

زمان: `2026-08-27 23:46 UTC`

Stage:

1. backup schema v1 موفق ساخت:
   `/var/backups/pvnaive/database/20260827T234607Z/pvnaive.dump.age`
2. migration `0002` را موفق apply کرد.
3. schema به v2 رسید.
4. backup دوم در همان ثانیه ساخته شد و به همان path برخورد کرد.

خطا:

```text
ERROR: backup destination already exists
S04_RESULT=FAILED
FAILED_LINE=276
ROLLBACK=INCOMPLETE
ROLLBACK_FAILED_STEP=rollback-migration-0002-no-schema2-backup
```

ریشه‌ی قطعی: `scripts/db/backup.sh` از timestamp با دقت فقط ثانیه استفاده می‌کرد (`%Y%m%dT%H%M%SZ`) و S04 دو backup پشت سر هم در همان second می‌ساخت.

نتیجه مهم: migration 0002 روی DB ماند چون Stage عمداً بدون schema-v2 backup rollback destructive انجام نداد. artifactهای S04 پاک شدند و Caddy/SSH/Firewall تغییری نکرد.

### Recovery preflight بعد از Attempt 2

زمان: `2026-08-27 23:49:59 UTC`

وضعیت به‌طور read-only verify شد:

```text
SCHEMA_VERSION=2
STORED_0002=0002_auth_foundation.up.sql|84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886
MANIFEST_0002=84bb735877d531c08ff4e7819c421c3746c00f1473ce185fd82ae4659815b886
SCHEMA_RECOVERY_STATE=PASSED
S04_ARTIFACT_CLEANUP=PASSED
INFRASTRUCTURE_INVARIANTS=PASSED
RECOVERY_PREFLIGHT=PASSED
```

Caddy checksum همان مقدار baseline بود. Portهای 22/80/443 و PostgreSQL loopback 5432 سالم بودند.

### Attempt 3 — recovery mode + API service startup blocker

Stage recovery خودش schema v2 را پذیرفت:

```text
RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER
```

schema-v2 backup موفق ساخته شد:

`/var/backups/pvnaive/database/20260827T234959Z/pvnaive.dump.age`

سپس systemd unit نصب/enable شد ولی service active نشد:

```text
Created symlink '/etc/systemd/system/multi-user.target.wants/pvnaive-api.service' → '/etc/systemd/system/pvnaive-api.service'.
ERROR: pvnaive-api.service is not active
S04_RESULT=FAILED
FAILED_LINE=298
ROLLBACK=STARTED
ROLLBACK=COMPLETED
CADDY_ACTION=none
SSH_ACTION=none
FIREWALL_ACTION=none
```

این آخرین وضعیت زنده است.

### برداشت از rollback Attempt 3

در recovery mode، migration قبلاً وجود داشت و `migration_owned=0` بود؛ بنابراین rollback فقط artifactهای S04 را پاک کرد و DB schema v2 را نگه داشت. با توجه به code و `ROLLBACK=COMPLETED` انتظار می‌رود اکنون:

- schema = 2
- `/opt/pvnaive/S04_AUTH.json` وجود نداشته باشد
- `/etc/pvnaive/auth.key` حذف شده باشد
- `/opt/pvnaive/bin/pvnaive` حذف شده باشد
- `/opt/pvnaive/bin/pvnaive-password` حذف شده باشد
- `/etc/systemd/system/pvnaive-api.service` حذف شده باشد
- `pvnaive-api.service` active نباشد
- port 8080 listener نداشته باشد
- Caddy/SSH/Firewall unchanged

اما Chat جدید باید این‌ها را قبل از retry دوباره verify کند؛ فقط از روی انتظار code فرض نکند.

## blocker فعلی و قدم بعدی دقیق

Blocker فعلی فقط این است:

`pvnaive-api.service is not active`

هیچ fix حدسی نزن. اول evidence بگیر.

### اولین دستور تشخیصی بعدی روی سرور

یک command فقط‌خواندنی بده که حداقل این موارد را جمع کند:

```bash
systemctl status pvnaive-api.service --no-pager -l || true
journalctl -u pvnaive-api.service -b --no-pager -n 200 || true
systemctl cat pvnaive-api.service || true
systemctl show pvnaive-api.service -p Result -p ExecMainStatus -p ExecMainCode -p ActiveState -p SubState || true
namei -l /etc/pvnaive/db.env /etc/pvnaive/auth.key /opt/pvnaive/bin/pvnaive 2>/dev/null || true
ss -lntp | grep -E ':(22|80|443|5432|8080)([[:space:]]|$)' || true
```

چون rollback unit/key/binary را حذف کرده، اگر journal از service سابق هنوز موجود باشد علت startup را نشان می‌دهد. اگر journal insufficient بود، Stage را مستقیم retry نکن؛ یک diagnostic install بدون enable/start یا manual `runuser -u pvnaive` probe از bundle بساز تا علت دقیق معلوم شود.

موارد محتمل مثل permission روی `db.env`, `PGPASSFILE`, `auth.key`, systemd sandbox یا runtime config فقط hypothesis هستند و تا journal evidence نیامده نباید به‌عنوان علت اعلام شوند.

## backup collision fix در branch

بعد از failure زنده، regression test `tests/db/backup_collision_test.sh` اضافه شد و CI wiring انجام شد. branch head در زمان این handoff:

`ab025aebe96af2920f7e2a405b63b1bd9c965ad9`

اما CI run `33127608010` روی این head هنوز `database=failure` دارد؛ Go و Web PASS هستند، rehearsal/bundle به‌دلیل database gate skip شده‌اند. بنابراین backup collision fix هنوز VERIFIED نهایی نیست و نباید bundle جدید از این head production تلقی شود تا database test root cause اصلاح و کل CI دوباره سبز شود.

آخرین production bundle قابل استناد همچنان bundle commit `b4803e27...` است؛ با این تفاوت که روی سرور به‌دلیل recovery mode فقط یک backup می‌گیرد و collision قبلی در همان مسیر تکرار نمی‌شود.

## systemd unit S04

unit intended:

- User/Group: `pvnaive`
- EnvironmentFile: `/etc/pvnaive/db.env`
- `PVNAIVE_AUTH_KEY_FILE=/etc/pvnaive/auth.key`
- `PVNAIVE_LISTEN=127.0.0.1:8080`
- ExecStart: `/opt/pvnaive/bin/pvnaive`
- hardening: NoNewPrivileges, ProtectSystem=strict, ProtectHome, PrivateTmp, PrivateDevices, RestrictAddressFamilies `AF_UNIX AF_INET`, empty capability set

Stage بعد از `systemctl enable --now` تا 20 بار readiness را probe می‌کند و سپس active/listener/live/ready/schema/Caddy invariants را می‌سنجد.

## چیزهایی که هنوز نباید انجام شوند

- S04 را PASSED اعلام نکن.
- PR #2 را merge نکن تا live localhost stage + independent postflight PASS شود و CI head نهایی هم سبز باشد.
- Caddyfile را هنوز برای panel/API تغییر نده.
- Owner واقعی را هنوز bootstrap نکن تا API localhost سالم و marker S04 ساخته شود.
- S05 را شروع نکن.
- schema v2 را دستی drop/rollback نکن مگر rollback procedure و backup verified صریح لازم شود.
- Naive/Caddy forward proxy configuration را تغییر نده.

## بعد از رفع service startup blocker

ترتیب صحیح:

1. root cause service را از journal مشخص کن.
2. regression test برای همان root cause در branch اضافه کن.
3. CI کامل را سبز کن.
4. اگر fix نیازمند bundle جدید است، artifact جدید بساز و SHA-gated deploy کن؛ اگر فقط محیطی و bundle فعلی صحیح است، recovery امن را تکرار کن.
5. S04 stage باید خروجی زیر بدهد:

```text
S04_RESULT=PASSED
S04_MODE=LOCALHOST_READY
S04_SCHEMA_VERSION=2
S04_API_LISTENER=127.0.0.1:8080
CADDY_ACTION=none
SSH_ACTION=none
FIREWALL_ACTION=none
```

6. postflight مستقل S04: marker، schema v2، service active، 127.0.0.1:8080 only، `/live`, `/ready`, auth key permissions، binary/service checks، DB roles/RLS، encrypted backup checksum، Caddy SHA و ports 22/80/443/5432.
7. فقط بعد از postflight، S04 = PASSED.
8. سپس owner واقعی با bootstrap local/root/TTY ساخته شود؛ password هرگز در command line یا repo ثبت نشود.
9. login واقعی روی localhost test شود.
10. بعد از آن Caddy integration برای `/panel/` و `/api/` با backup/validate/reload/rollback انجام شود.
11. browser login نهایی test شود.
12. سپس S05-USERS = NEXT.

## قواعد امنیتی ثابت پروژه

- account status و online/presence یکی نیستند.
- depleted/expired باید جدا نمایش داده شوند.
- unsupported capability در UI پنهان بماند.
- Domain Activity پیش‌فرض خاموش و بدون TLS MITM است.
- log نباید secret/token/password/path/query حساس داشته باشد.
- config apply باید validate → stage → atomic → rollback باشد.
- هیچ default password/token.
- endpoint بدون Access contract ممنوع.
- سایت عمومی و data plane از پنل جدا هستند.
- Controller در MVP وجود ندارد.
- Naive اولین Adapter است؛ طراحی باید extensible بماند.

## چیزی که Chat جدید باید به کاربر بگوید

نیازی نیست کاربر تاریخچه را دوباره توضیح دهد. بگو handoff را از Repository خواندی و آخرین blocker `pvnaive-api.service is not active` بعد از recovery S04 است. سپس فقط diagnostic read-only برای journal/status بده و از همان خروجی ادامه بده.
