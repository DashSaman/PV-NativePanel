# PVNaive — راهنمای نصب S05 روی سرور موجود

آخرین بروزرسانی: 2026-08-29

این Runbook فقط برای سرور موجود `testAmir5-3` نوشته شده است؛ همان نصب production که پنل روی `https://namir.softarg.ir/panel/` فعال است، Caddy/NaiveProxy از قبل کار می‌کند و API مدیریت روی loopback قرار دارد.

> **هشدار:** برای این مرحله از artifact یا اسکریپت قدیمی `S04R` استفاده نکن. S05 تا schema `6` می‌رود و فقط باید با `PVNaive-S05-*`, `S05-preflight.sh` و `S05-upgrade.sh` نصب شود.

این مسیر Fresh Installer عمومی نیست. هدف آن ارتقای کنترل‌شده‌ی نصب موجود به S05 Customer Service است: مدیریت Runtime واقعی Naive، ساخت مشتری، quota تجاری، تاریخ اعتبار، Subscription revocable و QR محلی. مصرف دقیق بایت و producer قابل‌اعتماد برای first-success هنوز اثبات نشده‌اند.

## شرط Release

فقط artifactی مجاز است که تمام jobهای CI همان HEAD سبز باشند:

- Go formatting/vet/tests؛
- Web tests/build؛
- PostgreSQL 18 + migrations/backup/restore/rollback؛
- `S05_UPGRADE_CONTRACT=PASSED`؛
- S04/S04R regression rehearsal؛
- S05 bundle contract + archive checksum.

Artifact workflow:

```text
PVNaive-S05-<40-char-commit>
```

داخل ZIP:

```text
PVNaive-S05-<12-char-commit>.tar.gz
PVNaive-S05-<12-char-commit>.tar.gz.sha256
```

## متغیر production این سرور

برای نصب فعلی، endpoint عمومی Naive همان دامنه production روی 443 است:

```bash
export PVNAIVE_NAIVE_PUBLIC_HOST='namir.softarg.ir:443'
```

Scheme یا path نگذار. مقدارهایی مانند `https://namir.softarg.ir/` معتبر نیستند.

## مرحله 1 — انتقال و verify کردن artifact

دو فایل `.tar.gz` و `.tar.gz.sha256` را در `/root/pvnaive-s05/` قرار بده، سپس:

```bash
set -Eeuo pipefail
umask 077
install -d -m 0700 /root/pvnaive-s05
cd /root/pvnaive-s05

sha256sum --check --strict PVNaive-S05-*.tar.gz.sha256
rm -rf ./bundle
mkdir -p ./bundle
tar -xzf PVNaive-S05-*.tar.gz -C ./bundle
cd ./bundle/PVNaive-S05-*
sha256sum --check --strict SHA256SUMS
cat RELEASE.json
```

ادامه فقط وقتی مجاز است که هر دو checksum `OK` باشند و `RELEASE.json` شامل این موارد باشد:

```text
"stage": "S05-CUSTOMER-SERVICE"
"schema_version": 6
"caddy_installer_mutation": false
```

## مرحله 2 — Read-only S05 preflight

این مرحله نباید production را تغییر دهد:

```bash
set -Eeuo pipefail
umask 077
cd /root/pvnaive-s05/bundle/PVNaive-S05-*

export PVNAIVE_NAIVE_PUBLIC_HOST='namir.softarg.ir:443'
bash scripts/stages/S05-preflight.sh | tee /root/PVNaive-S05-preflight.log
```

فقط با این marker ادامه بده:

```text
PREFLIGHT_RESULT=PASS
```

مقدار زیر را از خروجی دقیقاً ذخیره کن:

```text
CADDYFILE_SHA256=<64-hex>
```

Preflight همچنین باید schema فعلی DB، Caddy PID/NRestarts، API readiness، SSH، listenerهای 22/80/443 و loopback `127.0.0.1:8080` را نمایش دهد.

اگر `PREFLIGHT_RESULT=FAIL` دیدی، **upgrade را اجرا نکن**.

## مرحله 3 — Guarded S05 upgrade

SHA مرحله قبل را مستقیم از log استخراج کن تا اشتباه تایپی نشود:

```bash
set -Eeuo pipefail
umask 077
cd /root/pvnaive-s05/bundle/PVNaive-S05-*

export PVNAIVE_NAIVE_PUBLIC_HOST='namir.softarg.ir:443'
export PVNAIVE_EXPECTED_CADDY_SHA256="$(awk -F= '$1=="CADDYFILE_SHA256" {print $2}' /root/PVNaive-S05-preflight.log | tail -n1)"

[[ "${PVNAIVE_EXPECTED_CADDY_SHA256}" =~ ^[0-9a-f]{64}$ ]]

bash scripts/stages/S05-upgrade.sh | tee /root/PVNaive-S05-upgrade.log
```

Upgrade قبل از migration یک backup رمز‌شده DB می‌گیرد. سپس در صورت نیاز schema را از یکی از نسخه‌های پشتیبانی‌شده `2..5` به `6` می‌برد، Runtime key موجود را حفظ می‌کند، `/etc/pvnaive/api.env` را با public Naive host می‌سازد/به‌روز می‌کند، Runtime Agent و API را ارتقا می‌دهد و web release را atomically publish می‌کند.

Installer نباید Caddy را reload/restart کند.

موفقیت فقط با markerهای زیر پذیرفته می‌شود:

```text
S05_RESULT=PASSED
SCHEMA_VERSION=6
NAIVE_PUBLIC_HOST=namir.softarg.ir:443
CADDY_ACTION=none
```

همچنین `CADDY_SHA256_AFTER`, `CADDY_MainPID_AFTER` و `CADDY_NRestarts_AFTER` باید با pre-upgrade state برابر باشند.

بعد از موفقیت:

```bash
unset PVNAIVE_EXPECTED_CADDY_SHA256
```

## مرحله 4 — Postflight فوری

```bash
set -Eeuo pipefail

systemctl is-active pvnaive-runtime-agent.service
systemctl is-active pvnaive-api.service
systemctl is-active caddy-naive.service
systemctl is-active ssh.service

curl --fail --silent --show-error \
  --unix-socket /run/pvnaive/runtime-agent.sock \
  http://unix/v1/health

echo
curl --fail --silent --show-error \
  http://127.0.0.1:8080/api/v1/health/ready

echo
printf 'SCHEMA='
runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align \
  --host /var/run/postgresql --dbname pvnaive \
  --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations'

printf 'CADDY_SHA='
sha256sum /etc/caddy/Caddyfile
systemctl show caddy-naive.service --property=MainPID,NRestarts --no-pager
ss -H -lnt | grep -E '(:22|:80|:443|127\.0\.0\.1:8080)'
```

انتظار:

- Runtime Agent health = `status: ok`؛
- API `ready=true`؛
- schema = `6`؛
- API فقط روی `127.0.0.1:8080`؛
- Caddy/SSH فعال؛
- Caddy SHA/PID/NRestarts بدون تغییر نسبت به preflight.

## مرحله 5 — Import امن Runtime فعلی

قبل از ساخت هر مشتری جدید:

1. وارد `https://namir.softarg.ir/panel/` شو.
2. با Owner login کن.
3. وارد `Naive Runtime` شو.
4. اگر Runtime credential هنوز وارد DB نشده، `Import امن اکانت فعلی` را اجرا کن.
5. Import باید existing credential را بدون تغییر data-plane و بدون نمایش plaintext قبلی ثبت کند.
6. equivalence باید PASS شود.

اگر import/equivalence fail شد، هیچ Create/Rotate/Disable/Revoke انجام نده.

## مرحله 6 — ساخت اولین مشتری آزمایشی

بعد از Import موفق وارد `/panel/#/customers` شو.

برای تست اول پیشنهاد production-safe:

- Username آزمایشی جدید؛
- Generated password؛
- مثلاً `1 GB` یا Unlimited؛
- validity = **از زمان ساخت (`on_creation`)** یا fixed expiry؛
- فعلاً برای Pilot از `on_first_successful_connection` استفاده نکن، چون trusted CONNECT producer هنوز end-to-end روی production اثبات نشده است.

بعد از Create:

- password فقط یک بار نمایش داده می‌شود؛
- Subscription URL فقط یک بار تحویل داده می‌شود؛
- QR داخل Browser ساخته می‌شود؛
- direct Naive URI را می‌توان برای Karing کپی کرد.

## مرحله 7 — تست Client

با Karing یا client استاندارد Naive:

1. direct Naive URI مشتری جدید را import کن؛
2. اتصال HTTPS را تست کن؛
3. Subscription URL را جداگانه تست کن؛
4. existing credential قدیمی را هم دوباره تست کن تا regression نداشته باشیم؛
5. سپس در صورت نیاز rotate/disable/revoke را روی **اکانت آزمایشی** تست کن، نه credential اصلی production.

## چیزهایی که در S05 آماده‌اند

- Owner authentication؛
- Runtime import/create/rename/rotate/enable/disable/revoke؛
- last-active guard؛
- Runtime AES-GCM secret envelope؛
- expected-SHA + Caddy validate + backup + reload-only برای mutationهای credential؛
- customer creation؛
- finite GB quota یا Unlimited به‌عنوان service-state تجاری؛
- `on_creation`, `fixed_expiry`, و domain support برای first-success؛
- revocable opaque Subscription token؛
- local QR؛
- subscription token rotation/revocation؛
- idempotency و optimistic concurrency؛
- secret-safe customer list.

## مرزهایی که هنوز باید صادقانه حفظ شوند

- exact byte accounting هنوز proven نیست؛
- used/remaining traffic نباید عدد ساختگی داشته باشد؛
- hard byte quota enforcement تا PVN-045..049 غیرفعال است؛
- trusted first-successful-CONNECT producer هنوز production proof ندارد؛
- device/HWID limit، concurrent session limit، speed limit، reseller/credit و customer self-service هنوز در این Pilot نیستند.

## اگر Upgrade شکست خورد

`S05-upgrade.sh` rollback best-effort دارد و برای DB می‌تواند چند migration را تا schema قبل از upgrade برگرداند. بعد از failure، command را کورکورانه دوباره اجرا نکن.

این موارد را جمع کن:

```bash
systemctl status pvnaive-runtime-agent.service pvnaive-api.service caddy-naive.service ssh.service --no-pager -l || true
journalctl -u pvnaive-runtime-agent.service -n 150 --no-pager || true
journalctl -u pvnaive-api.service -n 150 --no-pager || true
journalctl -u caddy-naive.service -n 100 --no-pager || true
sha256sum /etc/caddy/Caddyfile
ss -H -lntp
```

و این evidenceها را نگه دار:

```text
/root/PVNaive-S05-preflight.log
/root/PVNaive-S05-upgrade.log
/var/backups/pvnaive/s05/<timestamp>/
```

هیچ runtime key، auth key، plaintext password، raw Subscription token یا backup age identity را در issue/chat عمومی قرار نده.
