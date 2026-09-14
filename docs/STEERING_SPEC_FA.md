# مشخصات فنی هدایت هوشمند و فلیت‌لایت (STEERING_SPEC_FA)

نسخه ۱ — 2026-09-14. سند مرجع اجرای R1→R4 و گیت‌های STEER-001..005/006.
سند مادر: `docs/AGENT_TASKS.md` (بستهٔ ارتقای مالک). قواعد: migrations فقط-جلو و
checksummed، کلید همان boundary معتمد Task12/15، «Unknown» هرگز ساختگی نمی‌شود.

## ۱. R1 — تله‌متری شبکه سمت سرور

### ۱.۱ DDL پیش‌نویس `session_network_samples` (partition ماهانه)

```sql
CREATE TABLE pvnaive.session_network_samples (
  id               BIGINT GENERATED ALWAYS AS IDENTITY,
  user_id          UUID        NOT NULL,
  node_id          TEXT        NOT NULL,
  boot_id          UUID        NOT NULL,
  session_seq      BIGINT      NOT NULL,
  sampled_at       TIMESTAMPTZ NOT NULL,
  rtt_micros       INTEGER     NOT NULL CHECK (rtt_micros  >= 0),
  rtt_var_micros   INTEGER     NOT NULL CHECK (rtt_var_micros >= 0),
  segs_out         BIGINT      NOT NULL CHECK (segs_out >= 0),
  segs_retrans     BIGINT      NOT NULL CHECK (segs_retrans >= 0),
  bytes_in         BIGINT      NOT NULL,
  bytes_out        BIGINT      NOT NULL,
  duration_micros  BIGINT      NOT NULL,
  PRIMARY KEY (id, sampled_at)
) PARTITION BY RANGE (sampled_at);

CREATE TABLE pvnaive.session_network_samples_2026_09
  PARTITION OF pvnaive.session_network_samples
  FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
```

- نمونه‌برداری داخل pinned forwardproxy روی سوکت‌های زنده، پیش‌فرض هر ۷ ثانیه
  (`PVNAIVE_NET_SAMPLE_INTERVAL_SECS`، بازه مجاز 5..10).
- کلید هویت: همان credential identity حسابداری دقیق؛ RemoteAddr معتبر فقط از Caddy.
- Ingest از مسیر Telemetry Agent موجود: batched + فشرده + idempotent
  (uniqueness روی `boot_id, session_seq, sampled_at`) و restart-safe
  (semantic همان accounting: cumulative/sequence). خطای ingest ⇒ fail-closed،
  هرگز drop خاموش.

### ۱.۲ aggregate ها (EWMA per user,node)

```sql
CREATE TABLE pvnaive.user_node_network_agg (
  user_id            UUID   NOT NULL,
  node_id            TEXT   NOT NULL,
  rtt_ewma_micros    BIGINT NOT NULL,
  jitter_ewma_micros BIGINT NOT NULL,
  retrans_ratio_ewma REAL   NOT NULL,
  throughput_bps     BIGINT NOT NULL,
  success_rate_ewma  REAL   NOT NULL,
  sample_count       BIGINT NOT NULL,
  last_sampled_at    TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (user_id, node_id)
);
```

فرمول به‌روزرسانی (سرِ نگهدارنده، نه دیتابیس):
`ewma_new = alpha * sample + (1 - alpha) * ewma_old` با `alpha` پیش‌فرض **0.2**.
`sample_count` شرط حداقل اعتبار: تا رسیدن به `min_samples` (پیش‌فرض 8) رکورد
«Unknown» است و در scoring وارد نمی‌شود.

## ۲. R2 — جدول پارامترهای موتور امتیاز و چرخش

| پارامتر | پیش‌فرض | محدوده مجاز | توضیح |
|---|---|---|---|
| rotation_window | 4h | 30m..24h | مرز چرخش؛ فقط در مرز پنجره سوییچ |
| grace_minutes | 15 | 5..60 | اعتبار هم‌زمان نود قبلی و جدید |
| ewma_alpha | 0.2 | 0.05..0.5 | نرمی؛ جلوگیری از واکنش به نمونه تکی |
| candidate_threshold | 0.85 | 0.7..0.95 | کاندیدها: score ≥ 85٪ بهترین |
| hysteresis_margin | 0.25 | 0.1..0.5 | سوییچ فقط اگر ≥۲۵٪ بهتر |
| hysteresis_windows | 2 | 1..4 | تعداد پنجره متوالی لازم |
| phase_offset | hash(userID) mod window | — | کاربرها همزمان نمی‌پرند |
| kill_switch | health failure → immediate | — | بدون hysteresis |
| top_k_mobile | 10 | 3..50 | زیرمجموعه per-user در sub موبایل |

امتیاز: `score = rank(rtt_ewma) − w_j·jitter − w_r·retrans_ratio − w_l·node_load`
(وزن‌ها config؛ پیش‌فرض w_j=0.2, w_r=0.3, w_l=0.2). تصمیم‌ها audit فقط-الحاقی
(append-only) با redaction. شبیه‌ساز قطعی (داده سینتتیک) گیت STEER-002 است.

## ۳. R3 — قرارداد رندر per خانوادهٔ کلاینت

مبنا: `GET /sub/<token>` با نگاه به User-Agent (صفحهٔ انسانی `/s/<token>` تغییر
رفتار نمی‌کند). هدرهای الزامی همهٔ فرمت‌ها:

```
profile-update-interval: 4
subscription-userinfo: upload=…; download=…; total=…; expire=…
cache-control: no-store
```

- **Mihomo/Clash**: YAML با `proxy-provider` (نه inline) ⇒ آپدیت گرم:
  ```yaml
  proxy-providers:
    pvnaive:
      type: http
      url: https://HOST/sub/TOKEN?family=mihomo
      interval: 4h
      health-check: {enable: true, url: http://www.gstatic.com/generate_204, interval: 300}
  proxy-groups:
    - {name: PV-AUTO, type: url-test, use: [pvnaive], url: http://www.gstatic.com/generate_204,
       interval: 300, tolerance: 50}
    - {name: PV-RR,   type: load-balance, use: [pvnaive], strategy: round-robin}
  ```
  هرگز `interrupt-exist-connections` ست نمی‌شود.
- **sing-box/Karing**: JSON با outbound گروه `urltest` (url همان 204، interval 300s،
  tolerance 50ms) + outbound های نود به‌صورت server با `server` = IP و `server_name` = SNI.
- **Hiddify**: فرمت بومی + پروفایل update-interval.
- **v2rayNG/base64 عمومی**: لیست node ها مرتب‌شده بر اساس تصمیم جاری (اولی = primary).
- **naive خام**: `naive://USER:PASS@IP:443#NAME` سرور جاری + alternates.

همهٔ رندرها: IP + SNI hardcode (بدون DNS)، atomic-apply (validate→apply→rollback)،
fail-closed، و اعمال steering فقط در مرز پنجره + grace ۱۵ دقیقه؛ هرگز ابطال credential
نودی که session زنده دارد (kill فقط با اکشن صریح اپراتور).

## ۴. R4 — Node Manifest (Ed25519)

```json
{
  "schema": "pvnaive.node.manifest.v1",
  "node_id": "node-…",
  "pool_id": "pool-…",
  "endpoints": [{"host": "203.0.113.10", "port": 443, "sni": "placeholder.example"}],
  "weight": 1.0,
  "valid_from": "2026-09-14T00:00:00Z",
  "valid_until": "2026-09-15T00:00:00Z"
}
```
امضا: Ed25519 روی canonical JSON (کلیدهای امضا هرگز در repo نیستند). توزیع out-of-band؛
قبل از اعتماد verify می‌شود؛ منقضی = نادیده. Replication کرفکرنسل از Runtime Agent
موجود روی mTLS و outbound-only. Accounting تجمیعی = جمع idempotent نودها
(جمع نودها == ledger، بدون double-count) — گیت STEER-004 با rehearsal کنترل‌شده.

## ۵. STEER-005 — قرارداد «یک اینترنت پایدار»

شش مکانیزم + سه تست پذیرش (sustained-flow: صفر fail و stall < 1s؛ hard-death:
failover در یک interval؛ anti-flap: دو نود <15ms ⇒ صفر نوسان در 24h شبیه‌سازی) —
شرح کامل در `docs/AGENT_TASKS.md` بخش STEER-005. TLS session ticket key مشترک
pool از مسیر secret/Runtime Agent، rotated و encrypted-at-rest.

## ۶. نقشهٔ گیت‌ها

| گیت | معیار پذیرش |
|---|---|
| STEER-001 | aggregate per (user,node,window) قابل کوئری؛ restart/reload بدون double-count؛ WS1 سبز |
| STEER-002 | شبیه‌سازی قطعی: بدون flapping؛ kill-switch درست؛ Unknown صادقانه |
| STEER-003 | ماتریس کلاینت واقعی (Hiddify/Karing/v2rayNG/Mihomo/naive) |
| STEER-004 | قطع نود وسط پنجره: سوییچ نامرئی + بایت دقیق |
| STEER-005 | سه تست بخش ۵ |
| STEER-006 | مقیاس ۱۰۰ نود (manifest ≤256KB، propagation ≤90s، رندر p99 <150ms، drain×5، outage) |
