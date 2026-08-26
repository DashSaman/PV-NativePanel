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
| GET | `/subscriptions/{token}` | config اشتراک |
| GET | `/subscriptions/{token}/info` | اطلاعات حداقلی صفحه اشتراک |

## حساب

`GET /me`، `GET /me/sessions`، حذف session و enrollment/confirm/delete برای TOTP.

## کاربران و Session

CRUD کاربر، suspend/resume/revoke/reset usage، credential create/rotate/revoke و:

- `GET /users/{id}/sessions`
- `DELETE /users/{id}/sessions/{sessionId}`

تغییر `concurrency_limit` از PATCH کاربر انجام و Audit می‌شود.

## Runtime و Usage

Status، revisions، validate/apply/rollback، summary، user usage و reconciliation.

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

- JSON به‌جز Subscription
- خطا: `code`، `message`، `request_id` و optional `fields`
- cursor pagination
- idempotency key برای mutation حساس
- revision/ETag برای optimistic concurrency
- rate limit جداگانه login/subscription/admin
- token/query/Authorization در log ممنوع
- endpoint بدون Access در Route Registry نباید build شود.
