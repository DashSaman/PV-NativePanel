# PVNaive — راهنمای نصب Pilot و تحویل اکانت به مشتری

آخرین بروزرسانی: 2026-08-28

این Runbook برای **سرور فعلی `testAmir5-3`** نوشته شده است؛ یعنی همان سروری که S03 و S04 Auth روی آن نصب شده، پنل روی `https://namir.softarg.ir/panel/` در دسترس است و Caddy/NaiveProxy از قبل فعال است.

این فایل **Fresh Installer عمومی** نیست. هدف این مرحله این است که همین نصب موجود را به S04R ارتقا بدهیم تا Owner بتواند اکانت‌های واقعی NaiveProxy را از داخل پنل بسازد، رمز را rotate کند، اکانت را فعال/غیرفعال/لغو کند و لینک آماده‌ی `naive+https://...` را به مشتری بدهد. قابلیت‌هایی مثل quota، مصرف دقیق، expiry تجاری، reseller، subscription page و customer portal هنوز جزو این Pilot نیستند.

## قبل از شروع

فقط artifactای را استفاده کن که تمام jobهای CI آن سبز باشند:

- Go
- Web
- Database / PostgreSQL 18
- Full S04R rehearsal
- Production bundle contract + checksum

نام artifact به شکل زیر است:

```text
PVNaive-S04-<40-char-commit>
```

داخل ZIP دو فایل وجود دارد:

```text
PVNaive-S04R-<12-char-commit>.tar.gz
PVNaive-S04R-<12-char-commit>.tar.gz.sha256
```

## قانون مهم Pilot

- اطلاعات ورود Owner را به مشتری نده.
- مشتری فقط لینک Naive خودش را می‌گیرد.
- Runtime فعلی ابتدا **بدون تغییر Caddy** Import می‌شود.
- تا Import امن و equivalence check پاس نشده، هیچ credential جدیدی نساز.
- Installer خود S04R روی Caddy هیچ write/reload/restart انجام نمی‌دهد.
- تغییرات credential بعداً توسط Runtime Agent با چرخه `expected SHA → backup → validate → write → reload-only → postflight → rollback on failure` انجام می‌شوند.

## مرحله 1 — انتقال و بررسی artifact روی سرور

فایل‌های `.tar.gz` و `.tar.gz.sha256` را روی سرور مثلاً داخل `/root/pvnaive-s04r/` قرار بده.

سپس:

```bash
set -Eeuo pipefail
umask 077
cd /root/pvnaive-s04r

sha256sum --check --strict PVNaive-S04R-*.tar.gz.sha256
rm -rf ./bundle
mkdir -p ./bundle
tar -xzf PVNaive-S04R-*.tar.gz -C ./bundle
cd ./bundle/PVNaive-S04R-*
sha256sum --check --strict SHA256SUMS
```

هر دو checksum باید `OK` باشند. در غیر این صورت نصب را ادامه نده.

## مرحله 2 — Read-only live preflight

این مرحله برای خواندن وضعیت زنده است و نباید چیزی را تغییر دهد:

```bash
set -Eeuo pipefail
umask 077
cd /root/pvnaive-s04r/bundle/PVNaive-S04R-*

bash scripts/stages/S04R-preflight.sh | tee /root/PVNaive-S04R-preflight.log
```

ادامه فقط وقتی مجاز است که انتهای خروجی این باشد:

```text
PREFLIGHT_RESULT=PASS
```

همچنین مقدار زیر را دقیقاً از خروجی نگه دار:

```text
CADDYFILE_SHA256=<64-hex>
```

اگر Preflight شکست خورد، upgrade را اجرا نکن و خروجی کامل را بررسی کن.

## مرحله 3 — Guarded S04R upgrade

مقدار SHA مرحله قبل را بدون تغییر جایگزین کن:

```bash
set -Eeuo pipefail
umask 077
cd /root/pvnaive-s04r/bundle/PVNaive-S04R-*

export PVNAIVE_EXPECTED_CADDY_SHA256='SHA_FROM_PREFLIGHT'
bash scripts/stages/S04R-upgrade.sh | tee /root/PVNaive-S04R-upgrade.log
unset PVNAIVE_EXPECTED_CADDY_SHA256
```

Upgrade قبل از migration یک backup رمز‌شده DB می‌گیرد و Caddy SHA/PID/NRestarts را قفل می‌کند.

موفقیت فقط با این خروجی‌ها پذیرفته می‌شود:

```text
S04R_RESULT=PASSED
SCHEMA_VERSION=3
CADDY_ACTION=none
```

و `CADDY_SHA256_AFTER` باید با SHA قبل برابر باشد.

## مرحله 4 — Postflight فوری روی سرور

```bash
set -Eeuo pipefail

systemctl is-active pvnaive-runtime-agent.service
systemctl is-active pvnaive-api.service
systemctl is-active caddy-naive.service

curl --fail --silent --show-error \
  --unix-socket /run/pvnaive/runtime-agent.sock \
  http://unix/v1/health

curl --fail --silent --show-error \
  http://127.0.0.1:8080/api/v1/health/ready

ss -H -lnt | grep -E '(:22|:80|:443|127\.0\.0\.1:8080)'
```

Runtime Agent باید فقط Unix socket داشته باشد؛ API باید روی `127.0.0.1:8080` باقی بماند.

## مرحله 5 — تحویل Runtime فعلی به پنل

در Browser:

1. وارد `https://namir.softarg.ir/panel/` شو.
2. با حساب **Owner** وارد شو.
3. از منوی `Naive Runtime` وارد صفحه مدیریت شو.
4. اگر لیست خالی است، روی **«Import امن اکانت فعلی»** بزن.
5. Import باید بدون نمایش password فعلی و بدون reload Caddy انجام شود.
6. بعد از Import، credential فعلی باید در لیست با منشأ `واردشده` دیده شود.

اگر Import پیام equivalence failure داد، هیچ mutation بعدی انجام نده.

## مرحله 6 — ساخت اکانت مشتری

1. در بخش «ساخت اکانت جدید» username مشتری را وارد کن.
2. گزینه «تولید رمز امن توسط سرور» روشن باشد.
3. روی «ساخت و اعمال اکانت» بزن.
4. بعد از apply موفق، رمز فقط **یک بار** نمایش داده می‌شود.
5. در همان پنجره دکمه **«کپی لینک Karing/Naive»** وجود دارد.
6. لینک به فرم زیر است:

```text
naive+https://USERNAME:PASSWORD@namir.softarg.ir:443
```

Username/password در لینک percent-encode می‌شوند تا کاراکترهای رزرو شده URL باعث خراب شدن لینک نشوند.

این لینک را به مشتری بده؛ **Owner login را هرگز به مشتری نده.**

## مرحله 7 — تست با Karing قبل از تحویل نهایی

روی یک کلاینت تست:

1. لینک ساخته‌شده را Copy کن.
2. در Karing از Import from Clipboard استفاده کن.
3. اتصال را فعال کن.
4. یک سایت HTTPS معمولی را باز کن.
5. بررسی کن credential قدیمی سرور نیز همچنان کار می‌کند.

فقط بعد از اینکه اکانت جدید واقعاً وصل شد، می‌توانی rotate/disable/revoke را برای تست عملی انجام بدهی.

## رفتارهای آماده در Pilot

- Secure Owner login
- Runtime status
- Import امن credential فعلی
- ساخت credential جدید
- generated password یک‌بار مصرف
- لینک آماده Karing/Naive
- rename
- password rotate تصادفی یا دلخواه
- enable / disable
- soft revoke
- last-active guard
- optimistic revision / stale-write rejection
- idempotency protection
- encrypted secret envelope در DB
- Runtime Agent فقط روی Unix socket
- Caddy validate/backup/reload-only/rollback برای mutationها

## چیزهایی که هنوز نباید به مشتری وعده داده شوند

- حجم/traffic quota دقیق
- مصرف لحظه‌ای یا billable usage
- expiry خودکار تجاری
- محدودیت device / HWID / concurrent session
- speed limit
- subscription token/page
- customer self-service portal
- reseller/credit
- notification/Telegram

این موارد بعد از Pilot تکمیل می‌شوند و تا زمانی که capability مربوطه اثبات نشده، UI نباید مقدار ساختگی نشان بدهد.

## اگر Upgrade شکست خورد

`S04R-upgrade.sh` rollback best-effort دارد. بعد از failure دوباره آن را کورکورانه اجرا نکن.

این خروجی‌ها را جمع کن:

```bash
systemctl status pvnaive-runtime-agent.service pvnaive-api.service caddy-naive.service --no-pager -l || true
journalctl -u pvnaive-runtime-agent.service -n 150 --no-pager || true
journalctl -u pvnaive-api.service -n 150 --no-pager || true
journalctl -u caddy-naive.service -n 100 --no-pager || true
sha256sum /etc/caddy/Caddyfile
ss -H -lntp
```

و فایل‌های زیر را نگه دار:

```text
/root/PVNaive-S04R-preflight.log
/root/PVNaive-S04R-upgrade.log
/var/backups/pvnaive/s04r/<timestamp>/
```

هیچ secret، runtime key، auth key یا backup age identity را داخل issue/chat عمومی paste نکن.

## وضعیت Release

این مسیر یک **Pilot روی نصب موجود** است، نه Release Candidate عمومی. بعد از اینکه Pilot واقعی با حداقل یک credential جدید و یک client Karing پاس شد، evidence آن باید در Repo ثبت شود و سپس توسعه‌ی security hardening، user lifecycle، accounting، subscription و installer عمومی ادامه پیدا کند.
