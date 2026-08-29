# قرارداد API نسخه v1

Prefix: `/api/v1`

## عمومی

| Method | Route | توضیح |
|---|---|---|
| GET | `/health/live` | زنده‌بودن process |
| GET | `/health/ready` | آمادگی dependencyها |
| POST | `/auth/login` | ایجاد session امن |
| POST | `/auth/logout` | revoke session |
| POST | `/auth/refresh` | rotation refresh token |
| GET | `/subscriptions/{token}` | خروجی Subscription با یک URI معتبر `naive+https://...`؛ بدون نیاز به session |
| GET | `/subscriptions/{token}/info` | رزرو برای اطلاعات حداقلی صفحه اشتراک؛ هنوز capability آماده ندارد |

### قرارداد امنیتی Subscription

- token خام ۲۵۶ بیتی فقط هنگام ایجاد/صدور مجدد به Owner تحویل می‌شود؛ دیتابیس فقط SHA-256 و prefix غیرحساس را نگه می‌دارد.
- endpoint عمومی token را با state زنده‌ی کاربر، service term، binding و Runtime credential بررسی می‌کند؛ token لغوشده، credential غیرفعال، user تعلیق/لغوشده یا term منقضی resolve نمی‌شود.
- رمز Runtime در پاسخ‌های مدیریتی list برنمی‌گردد. renderer داخلی فقط برای ساخت URI، ciphertext را با کلید Runtime باز می‌کند و plaintext را در log نمی‌نویسد.
- host داخل URI از `PVNAIVE_NAIVE_PUBLIC_HOST` می‌آید و هنگام startup validate می‌شود؛ Host header درخواست مبنای ساخت کانفیگ نیست.
- دریافت Subscription به‌تنهایی «اولین اتصال» محسوب نمی‌شود.

## حساب

`GET /me`، `GET /me/sessions`، حذف session و enrollment/confirm/delete برای TOTP.

## مدیریت مستقیم مشتری Naive

این API برای جریان مستقیم Owner است و عمداً مدل business service را از Runtime credential جدا نگه می‌دارد.

| Method | Route | Access | توضیح |
|---|---|---|---|
| GET | `/customers` | owner | فهرست safe projection مشتری، term، quota، expiry و capability؛ بدون password/token خام |
| POST | `/customers` | owner | ایجاد User + service term snapshot + Runtime credential + token اشتراک در یک saga |
| POST | `/customers/{id}/subscription/rotate` | owner | لغو token فعال قبلی و صدور token جدید یک‌بارنمایش؛ idempotent |

نمونه payload ساخت:

```json
{
  "username": "test1",
  "generate_password": true,
  "password": "",
  "quota_gb": 50,
  "validity": {
    "mode": "on_first_successful_connection",
    "duration_days": 30
  }
}
```

`quota_gb: null` یعنی نامحدود. سه mode اعتبار پشتیبانی می‌شود:

- `on_first_successful_connection`؛ پیشنهاد پیش‌فرض، تا سیگنال CONNECT احراز‌شده term در `pending` می‌ماند.
- `on_creation`؛ از زمان ساخت شروع می‌شود.
- `fixed_expiry`؛ با `expires_at` دقیق.

پاسخ create شامل password تولیدشده و `subscription_path` فقط برای همان تحویل است. UI باید آن‌ها را one-time delivery تلقی کند.

### وضعیت Usage/Quota

وجود `quota_bytes` به معنی اثبات Accounting نیست. تا زمانی که PVN-045..049 پاس نشده‌اند:

```json
{"available":false,"reason":"exact_accounting_not_proven"}
```

باید نمایش داده شود. UI حق ندارد مصرف را `0`، remaining را برابر quota یا hard-enforcement را فعال نشان دهد.

## کاربران و Session

CRUD کاربر، suspend/resume/revoke/reset usage، credential create/rotate/revoke و:

- `GET /users/{id}/sessions`
- `DELETE /users/{id}/sessions/{sessionId}`

تغییر `concurrency_limit` از PATCH کاربر انجام و Audit می‌شود.

## Runtime و Usage

Status، revisions، validate/apply/rollback، summary، user usage و reconciliation.

Raw Runtime credential manager در `/runtime/naive` سطح پیشرفته است؛ ساخت مشتری تجاری باید از `/customers` انجام شود تا quota/term/subscription به credential خام نچسبد.

## Logs و Diagnostics

| Method | Route | Access |
|---|---|---|
| GET | `/logs/application` | operator |
| GET | `/logs/runtime` | operator |
| GET | `/logs/security` | auditor |
| GET | `/diagnostics/requests/{requestId}` | operator |
| GET | `/diagnostics/domain-activity` | owner |
| POST | `/diagnostics/domain-activity/enable` | owner |
| POST | `/diagnostics/domain-activity/disable` | owner |
| DELETE | `/diagnostics/domain-activity` | owner |
| POST | `/diagnostics/bundles` | owner |

Domain Activity پیش‌فرض خاموش است و path/query ذخیره نمی‌کند.

## System و Backup

System status، Audit events، create/list/verify backup.

## قواعد

- JSON به‌جز بدنه‌ی Subscription که `text/plain` است.
- خطا: `code`، `message`، `request_id` و optional `fields`.
- cursor pagination برای لیست‌های بزرگ.
- `Idempotency-Key` برای mutation حساس؛ subscription reissue با replay دوباره token را rotate نمی‌کند.
- revision/ETag برای optimistic concurrency.
- rate limit جداگانه login/subscription/admin.
- token/query/Authorization در log ممنوع.
- endpoint بدون Access در Route Registry نباید build شود.
