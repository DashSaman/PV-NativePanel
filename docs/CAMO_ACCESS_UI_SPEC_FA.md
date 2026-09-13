# مشخصات فنی سایت پوششی، مدیریت دسترسی و UI مرکز فرمان (CAMO_ACCESS_UI_SPEC_FA)

نسخه ۱ — 2026-09-14. مرجع اجرای R6/R7/R8 و گیت‌های CAMO-001..004، ACCESS-001..004، UI-001..004.
سند مادر: `docs/AGENT_TASKS.md`. خط قرمز: طرح «سبک رسمیِ» اورجینال — کلون هویت سازمان
واقعی ممنوع؛ ویدیو فقط embed/لینک؛ هیچ اشاره‌ای به پنل در سایت پوششی.

## ۱. R6 — معماری `coverd`

### ۱.۱ مسیریابی Caddy (ترتیب سخت‌گیرانه، per node)

```
① مسیر پنل (base_path، پیش‌فرض /panel)  → reverse_proxy 127.0.0.1:<panel_port>
② CONNECT / دیتاپلین NaiveProxy          → pinned forwardproxy (دست‌نخورده)
③ همهٔ مسیرهای دیگر                      → reverse_proxy 127.0.0.1:<coverd_port>
```

`coverd` فقط روی loopback گوش می‌دهد و فقط از طریق Caddy دیده می‌شود. هیچ پورت
عمومی، هیچ مسیر مدیریتی خارجی.

### ۱.۲ اسکیمای جدول `cover_content` (partition ماهانه)

```sql
CREATE TABLE pvnaive.cover_content (
  id           BIGINT GENERATED ALWAYS AS IDENTITY,
  node_id      TEXT        NOT NULL,
  source       TEXT        NOT NULL,   -- 'khamenei.ir' | 'irib' | 'irna' | ...
  kind         TEXT        NOT NULL,   -- 'news' | 'video' | 'message'
  title        TEXT        NOT NULL,
  summary      TEXT,
  url          TEXT        NOT NULL,
  thumb_url    TEXT,
  published_at TIMESTAMPTZ,
  fetched_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  stale        BOOLEAN     NOT NULL DEFAULT false,
  PRIMARY KEY (id, fetched_at)
) PARTITION BY RANGE (fetched_at);

CREATE TABLE pvnaive.cover_nodes (
  node_id    TEXT PRIMARY KEY,
  persona_id TEXT NOT NULL,             -- 6+ پک؛ پیش‌فرض hash(node_id) mod N
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

RLS بر اساس node scope؛ سیاست پاکسازی: نگهداری ۳۵ روز.

### ۱.۳ خط تولید محتوا

- زمان‌بندی: 6h ± jitter (تصادفی 0..30m)، per-source conditional GET
  (ETag/Last-Modified)، rate-limit محترمانه (≥2s فاصله)، backoff نمایی 2^n تا 1h.
- منابع: RSS/فیدهای عمومی منابع رسمی داخلی. فقط title/summary/thumbnail/link کش
  می‌شود؛ ویدیو/پیام‌ها فقط embed از پخش‌کنندهٔ رسمی یا لینک.
- خطا: منبع مرده ⇒ آخرین snapshot با برچسب stale، سایت هرگز خالی/خراب نمی‌شود؛
  ورودی خراب فید ⇒ صفر crash (فاز fuzz الزامی).
- سلامت در پنل: کارت «Cover health» — per-source: آخرین fetch، سن محتوا، stale.

### ۱.۴ پرسونا پک‌ها (≥۶ عدد، اورجینال)

| persona_id | قالب | ویژگی بصری |
|---|---|---|
| news-portal | پرتال خبری | اسلایدر + ستون‌بندی ۳تایی، پالت سرمه‌ای |
| cultural | بنیاد فرهنگی | تایپوگرافی سنتی، کرم/طلایی |
| city-services | خدمات شهری | کارت‌های خدمات، آبی/سبز |
| news-agency | خبرگزاری منطقه‌ای | فید زنده، قرمز تیره |
| research | مؤسسه پژوهشی | گزارش‌محور، خاکستری/فیروزه‌ای |
| charity | خیریه | کمپین‌محور، سبز |

هر پک: ساختار HTML + واژگان کلاس + CSS + فونت + لوگو + favicon + footer + sitemap
+ خروجی RSS خودش. تخصیص پیش‌فرض `stable_hash(node_id) mod N`؛ تغییر از UI با
پیش‌نمایش زنده. الزام: DOM-hash دو نود بالای آستانهٔ فاصلهٔ ساختاری؛ سن محتوا < 24h.

### ۱.۵ بهداشت پاسخ (CAMO)

بدون cookie، بدون CORS، بدون X-Powered-By، بدون Server معرف، بدون CSP با اشارهٔ
پنل، 404/500 انسانی، robots.txt + sitemap.xml، تاریخ جلالی + نماز + بخش پیام‌ها.
CAMO-002: sweep مسیرهای تصادفی/حساس (/admin، /.env، /wp-login.php، …) باید از یک
سایت فارسی عادی غیرقابل‌تفکیک باشد (آدیت خودکار grep + هدر در CI).

## ۲. R7 — مدیریت دسترسی پنل (ACCESS)

### ۲.۱ اسکیمای `panel_settings` (تک‌ردیف)

```sql
CREATE TABLE pvnaive.panel_settings (
  id             INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  admin_username TEXT    NOT NULL,
  password_hash  TEXT    NOT NULL,        -- argon2id
  base_path      TEXT    NOT NULL DEFAULT '/panel'
                 CHECK (base_path ~ '^/[a-z0-9][a-z0-9\-_]{2,63}$'),
  listen_port    INTEGER NOT NULL DEFAULT 8080 CHECK (listen_port BETWEEN 1 AND 65535),
  session_ttl    INTERVAL NOT NULL DEFAULT '12h',
  grace_minutes  INTEGER  NOT NULL DEFAULT 10 CHECK (grace_minutes BETWEEN 1 AND 60),
  exposure_mode  TEXT     NOT NULL DEFAULT 'reverse_proxy'
                 CHECK (exposure_mode IN ('reverse_proxy','direct')),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_by     UUID
);
```

ممنوع برای base_path: `/api`، `/assets`، `/health`، مسیرهای cover (لیست در
config). audit فقط-الحاقی: actor، old→new، secretها redacted.

### ۲.۲ چهار فلو (همه با step-up auth + CSRF + rate limit 5/min)

1. **یوزرنیم** — یکتایی، apply ساده.
2. **رمز** — ≥12 کاراکتر + zxcvbn ≥ 3؛ argon2id؛ چرخش session secret؛ ابطال همهٔ
   sessionهای دیگر، حفظ session فعلی (همان قرارداد `/api/v1/me/password` موجود —
   این فلو آن را به panel_settings تعمیم می‌دهد).
3. **base_path** — اعتبارسنجی regex؛ در grace (پیش‌فرض ۱۰ دقیقه) هر دو مسیر
   پذیرفته، بعد مسیر قدیمی hard-deny؛ همهٔ لینک‌های UI prefix-relative
   (هیچ absolute path در UI نیست).
4. **پورت** — حالت توصیه‌شده: پنل loopback پشت Caddy؛ تغییر = rebind اتمیک +
   آپدیت upstream از admin API با health-check + rollback خودکار. حالت `direct`
   با هشدار صریح + تأیید تایپی.

### ۲.۳ ضد قفل‌شدگی (قفل‌شدگی = باگ)

- apply تراکنشی: write → rebind/verify → commit؛ هر شکست ⇒ rollback خودکار به
  آخرین وضعیت سالم؛ نیمه‌اعمال‌شده ممنوع.
- CLI نجات روی خود سرور: `pvnaive admin reset-access --username … --password …
  --base-path … --port …` (بدون UI)، در راهنمای پنل مستند؛ drill بازیابی < ۵ دقیقه.
- ACCESS-002: chaos — کشتن listener حین تغییر ⇒ rollback خودکار، صفر قفل.

## ۳. R8 — UI مرکز فرمان

### ۳.۱ دیزاین توکن‌ها (قبل از هر کد صفحه)

- متدولوژی ui-ux-pro-max-skill: مقیاس تایپوگرافی، گرید 8pt، نقش‌های معنایی رنگ
  (primary/accent/success/warn/danger)، کنتراست WCAG AA، قواعد موشن (durations
  120/180/240ms، easing استاندارد). خروجی: `web/src/design-tokens.ts` + مستند.
- فارسی RTL لوکال اول: فونت وزیرمتن‌کلاس (self-hosted، الان نصب)، تاریخ جلالی.
- الهام از 21st.dev فقط پس از ممیزی لایسنس (MIT/ISC/Apache-2.0)؛ ساخت
  کتابخانهٔ کامپوننت داخلی کوچک — هیچ snippet چسبانده‌ای.

### ۳.۲ لاگین استلث (UI-001)

- پس‌زمینهٔ تمام‌صفحهٔ انیمیشنی (canvas: gradient mesh/ذرات، GPU-cheap، محترم به
  `prefers-reduced-motion`).
- فیلدهای یوزرنیم/رمز در حالت عادی نامرئی (بدون باکس/placeholder/لیبل)؛
  materialize با hover-reveal (fade/slide ظریف 150–250ms) یا focus کیبورد.
  Tab و پسوردمنیجر باید بی‌نقص کار کنند (fallback دسترسی‌پذیری). caret و
  glass-morphism ظریف تنها نشانه‌ها؛ هیچ outline هوشمندي موقعیت را لو نمی‌دهد.
- لاگین ناموفق layout را نمی‌پراند (بدون layout-shift tell).
- استلث لایهٔ UX است نه کنترل امنیتی — کنترل واقعی: rate limit/fail2ban (ACCESS-003)
  و RBAC؛ این جمله در کد و مستند صریح است.

### ۳.۳ مانیتورینگ زنده (UI-002/003/004)

- ترنسپورت: WebSocket (docs: reconnect با backoff، backpressure، drop-oldest
  per-client، ring buffer).
- نمودارها (uPlot/ECharts — ADR): توتال فلیت in/out؛ کارت per-node (sparkline +
  bps + session + badge سلامت، بازشونده)؛ per-user (bps جاری + sparkline 60s).
- RBAC روی endpoint استریم: owner همه، tenant فقط خودش، deny-by-default، 403 بسته.
- ~1s رفرش؛ فقط رندر smooth می‌شود نه داده؛ last-updated stamp؛ gap صادقانه؛
  هیچ مقدار ساختگی. Unknown همان Unknown.
- Reconcile: جمع bps نودها == توتال (در tolerance)؛ جمع شبانه == ledger دقیق
  (job شبانه)؛ در غیر این‌صورت bug جدید در KNOWN_ISSUES.
- پرفورمنس (UI-003): 100 نود + 10k session ⇒ 60fps، تب < 250MB، WS payload/s
  مستند، CPU اضافهٔ سرور < 2%.

## ۴. نقشهٔ گیت‌ها (این سند)

| گیت | معیار |
|---|---|
| CAMO-001 | همهٔ مسیرهای غیر-پنل → cover؛ صفر artifact پنل (آدیت CI) |
| CAMO-002 | پروب‌سوئیپ غیرقابل‌تفکیک از سایت عادی |
| CAMO-003 | تنوع DOM-hash بالای آستانه + محتوا < 24h |
| CAMO-004 | منابع مرده ⇒ snapshot؛ fuzz فید ⇒ صفر crash |
| ACCESS-001 | چهار تغییر از UI + grace + session + audit |
| ACCESS-002 | chaos ⇒ rollback خودکار، صفر قفل |
| ACCESS-003 | ضد brute-force + سیاست رمز |
| ACCESS-004 | بازیابی CLI < ۵ دقیقه |
| UI-001 | لاگین استلث + دسترسی‌پذیری + reduced-motion |
| UI-002 | reconcile زنده و شبانه با ledger |
| UI-003 | پرفورمنس 100 نود با اعداد مستند |
| UI-004 | کیفیت RTL + جداسازی RBAC استریم |
