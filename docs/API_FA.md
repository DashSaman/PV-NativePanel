# قرارداد API نسخه v1

Prefix: `/api/v1`

## عمومی

| Method | Route | توضیح |
|---|---|---|
| GET | `/health/live` | زنده‌بودن process؛ بدون جزئیات حساس |
| GET | `/health/ready` | آمادگی dependencyها؛ جزئیات فقط authenticated |
| POST | `/auth/login` | ایجاد session امن |
| POST | `/auth/logout` | revoke session |
| POST | `/auth/refresh` | rotation refresh token |
| GET | `/subscriptions/{token}` | دریافت Subscription |

## حساب و امنیت

| Method | Route |
|---|---|
| GET | `/me` |
| GET | `/me/sessions` |
| DELETE | `/me/sessions/{id}` |
| POST | `/me/mfa/totp/enroll` |
| POST | `/me/mfa/totp/confirm` |
| DELETE | `/me/mfa/totp` |

## کاربران

| Method | Route |
|---|---|
| GET/POST | `/users` |
| GET/PATCH | `/users/{id}` |
| POST | `/users/{id}/suspend` |
| POST | `/users/{id}/resume` |
| POST | `/users/{id}/revoke` |
| POST | `/users/{id}/reset-usage` |
| GET/POST | `/users/{id}/credentials` |
| POST | `/users/{id}/credentials/{credentialId}/rotate` |
| DELETE | `/users/{id}/credentials/{credentialId}` |

## Runtime و Usage

| Method | Route |
|---|---|
| GET | `/runtime/status` |
| GET | `/runtime/revisions` |
| POST | `/runtime/revisions/validate` |
| POST | `/runtime/revisions/{id}/apply` |
| POST | `/runtime/revisions/{id}/rollback` |
| GET | `/usage/summary` |
| GET | `/usage/users/{id}` |
| GET | `/usage/reconciliation` |

## عملیات

| Method | Route |
|---|---|
| GET | `/system/status` |
| GET | `/audit-events` |
| POST | `/backups` |
| GET | `/backups` |
| POST | `/backups/{id}/verify` |

## قواعد

- JSON فقط، به‌جز Subscription
- خطا با `code`، `message`، `request_id` و optional `fields`
- pagination مبتنی بر cursor
- mutationهای حساس با idempotency key
- optimistic concurrency با revision/ETag
- rate limit جدا برای login، subscription و admin API
- endpoint بدون تعریف مجوز در route registry نباید build شود.
