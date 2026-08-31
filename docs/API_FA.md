# PVNaive — قرارداد API v1

آخرین reconciliation: 2026-08-30

Prefix اصلی مدیریتی: `/api/v1`

## قانون مهم درباره Routeها

`internal/httpapi/routes.go` علاوه بر endpointهای واقعی، contractهای آینده را هم نگه می‌دارد. وجود Route در registry به معنی implemented بودن نیست.

برای current truth باید هر Route با این موارد تطبیق داده شود:

- handler واقعی در `NewServer` یا `customerExtraHandler`؛
- service/store واقعی؛
- authorization؛
- DB/schema؛
- tests؛
- UI در صورت operator-facing بودن؛
- Production behavior در صورت Production-facing بودن.

Routeهای بدون implementation ممکن است به `501 not_implemented` برسند. بنابراین این سند فقط capabilityهای فعلی را به‌عنوان current معرفی می‌کند و future contracts را جدا نگه می‌دارد.

## Public / health / auth

| Method | Route | وضعیت فعلی | توضیح |
|---|---|---|---|
| GET | `/api/v1/health/live` | implemented | process liveness |
| GET | `/api/v1/health/ready` | PARTIAL | endpoint واقعی است، ولی هنوز bounded DB/schema readiness probe ندارد |
| POST | `/api/v1/auth/login` | implemented | Owner/admin-style session login بر مبنای auth service |
| POST | `/api/v1/auth/logout` | implemented | session revoke؛ generic commit-before-success bug هنوز در middleware باز است |
| POST | `/api/v1/auth/refresh` | implemented with known bug | refresh rotation وجود دارد؛ revoked-token reuse-family path هنوز نیاز به fix دارد |
| GET | `/sub/{token}` | implemented | machine Subscription برای client/Karing-compatible import |
| GET | `/s/{token}` | implemented | human Account Page |
| GET | `/api/v1/subscriptions/{token}` | implemented compatibility path | machine-style Subscription compatibility endpoint |

`/sub` و `/s` read-only هستند و به‌تنهایی password/token را rotate نمی‌کنند و first-use activation ایجاد نمی‌کنند.

## Authenticated account

Handlerهای فعلی شامل:

- `GET /api/v1/me`
- `GET /api/v1/me/sessions`
- `DELETE /api/v1/me/sessions/{id}`
- `POST /api/v1/me/mfa/totp/enroll`
- `POST /api/v1/me/mfa/totp/confirm`
- `DELETE /api/v1/me/mfa/totp`

Recovery codes برای MFA-management وجود دارند، اما recovery-code **login** هنوز product decision/implementation کامل ندارد.

## Owner direct-customer API

این surface مدل تجاری Customer/ServiceTerm را از Runtime credential جدا نگه می‌دارد.

Handlerهای فعلی:

| Method | Route | Access | توضیح |
|---|---|---|---|
| GET | `/api/v1/customers` | owner | customer list projection؛ accounting از read-model current استفاده می‌کند |
| POST | `/api/v1/customers` | owner | ایجاد customer + service state + Runtime binding/delivery flow |
| POST | `/api/v1/customers/adopt-runtime` | owner | adopt کردن Runtime credential موجود بدون ساخت password جدید |
| PATCH | `/api/v1/customers/{id}/service` | owner | quota/expiry/service update؛ نباید token/password را rotate کند |
| GET | `/api/v1/customers/{id}/subscription` | owner | read-only current Subscription metadata/delivery |
| POST | `/api/v1/customers/{id}/subscription/rotate` | owner | Subscription reissue؛ مستقل از password rotation |
| POST | `/api/v1/customers/{id}/suspend` | owner | suspend |
| POST | `/api/v1/customers/{id}/resume` | owner | resume |
| DELETE | `/api/v1/customers/{id}` | owner | safe revoke/delete semantics؛ hard destructive delete نیست |
| POST | `/api/v1/customers/{id}/rotate-password` | owner | Runtime password rotation؛ مستقل از Subscription reissue |
| POST | `/api/v1/customers/{id}/volume/add` | owner | add quota volume |
| POST | `/api/v1/customers/{id}/validity/extend` | owner | extend validity |
| POST | `/api/v1/customers/{id}/reset-usage` | owner | reset دقیق usage با idempotency/audit؛ credential/token ثابت می‌ماند |

Manual Reset Usage روی هر دو surface آماده است: `/api/v1/customers/{id}/reset-usage` و `/api/v1/users/{id}/reset-usage`. Bulk `reset_usage` نیز فقط برای Owner از Preview → Execute استفاده می‌کند.

## Customer-product / reseller-scoped API

Current `customerExtraHandler` پیاده‌سازی واقعی برای این contractها دارد:

- `GET /api/v1/users`
- `POST /api/v1/users`
- `PATCH /api/v1/users/{id}`
- `POST /api/v1/users/{id}/suspend`
- `POST /api/v1/users/{id}/resume`
- `POST /api/v1/users/{id}/revoke`
- `POST /api/v1/users/{id}/renew`
- `GET /api/v1/users/{id}/subscription`
- `POST /api/v1/users/{id}/subscription/rotate`
- `POST /api/v1/users/{id}/rotate-password`
- `PATCH /api/v1/users/{id}/service`
- `POST /api/v1/users/{id}/volume/add`
- `POST /api/v1/users/{id}/validity/extend`
- `POST /api/v1/users/{id}/reset-usage` (Owner-only)
- `GET/POST /api/v1/plans`
- `GET/POST /api/v1/customer-groups`
- `GET/POST /api/v1/customer-tags`
- `POST /api/v1/users/bulk/preview`
- `POST /api/v1/users/bulk/execute`

Current bulk action set شامل lifecycle/product/subscription/volume/plan/group/tag actions و `reset_usage` است. `reset_usage` فقط برای Owner مجاز است، Preview و Execute باید همان `Idempotency-Key` را استفاده کنند، و اجرای هر کاربر transaction مستقل با نتیجه‌ی per-item دارد؛ Password و Subscription token در این عملیات rotate نمی‌شوند.

## Runtime Naive API

Owner-only Runtime credential controls فعلی:

- `GET /api/v1/runtime/naive`
- `GET /api/v1/runtime/naive/credentials`
- `POST /api/v1/runtime/naive/import`
- `POST /api/v1/runtime/naive/credentials`
- `PATCH /api/v1/runtime/naive/credentials/{id}`
- `POST /api/v1/runtime/naive/credentials/{id}/rotate-password`
- `DELETE /api/v1/runtime/naive/credentials/{id}`

Runtime mutation boundary:

`validate → exact backup → expected-SHA apply → Caddy reload → postflight → rollback on failure`

API process نباید arbitrary root shell/filesystem/service control داشته باشد.

## Accounting state

Exact direct-Naive accounting دیگر یک future-only concept نیست. Current main/Production دارای:

- stable Runtime credential UUID identity؛
- trusted successful CONNECT producer؛
- dedicated telemetry Unix socket؛
- append-only/idempotent event ingest؛
- boot/session/sequence/cumulative semantics؛
- ServiceTerm-isolated accounting؛
- session/presence projection؛
- finite-quota reservation/settlement core.

اما این به معنی آماده‌بودن همه future accounting API route declarations نیست.

Periodic reset execution نیز current capability است: policy `none/daily/weekly/monthly/yearly/custom` هنگام ساخت ServiceTerm فریز می‌شود؛ scheduler با cursor پایدار، UTC policy، `FOR UPDATE SKIP LOCKED`، idempotent reset event، audit/history و retry/defer روی accounting unsafe state اجرا می‌شود. تغییر بعدی Plan قرارداد ServiceTerm فعال را بازنویسی نمی‌کند. `renew_current` policy فریز‌شده را حفظ می‌کند.

Current missing product actions:

- operator session kill/limits؛
- کامل‌شدن accounting/presence projection در تمام UI/APIهای customer؛
- controlled Production acceptance proof برای hard quota و first-CONNECT races/restarts.

## Future/declaration-only surfaces

تا زمان پیاده‌سازی و test، این دسته‌ها را current API capability حساب نکن:

- operator customer sessions/kill؛
- full reseller CRUD/credit/ledger handlers؛
- notification preferences/rules/delivery surface؛
- system metrics/logs/diagnostics/support bundle؛
- backup UI/API workflow؛
- OpenAPI/Swagger current-main integration؛
- fleet/multi-node؛
- webhooks.

بخش‌هایی از ops/OpenAPI/notifications/fleet در PR قدیمی #16 وجود دارد ولی تا استخراج روی latest main و CI مجدد current محسوب نمی‌شود.

## Periodic reset scheduler operational contract

- `PVNAIVE_PERIODIC_RESET_INTERVAL_SECONDS`: default `30`، بازه مجاز `5..3600` ثانیه.
- `PVNAIVE_PERIODIC_RESET_BATCH_LIMIT`: default `50`، بازه مجاز `1..100` ServiceTerm در هر batch.
- API فقط تابع محدود `pvnaive.execute_due_scheduled_usage_resets(limit)` را صدا می‌زند؛ caller نمی‌تواند user یا timestamp دلخواه برای reset بدهد.
- هر batch از `clock_timestamp()` سرور DB استفاده می‌کند و timezone policy فریز‌شده `UTC` است.
- success، deferred و skipped در `scheduled_usage_reset_attempts` append-only ثبت می‌شوند؛ success به `direct_naive_accounting_reset_events` با reason=`scheduled` لینک می‌شود.
- اگر Manual/Bulk reset جدیدتری از boundary زمان‌بندی‌شده وجود داشته باشد، scheduler double-reset نمی‌کند و cadence را از epoch اثبات‌شده جدید ادامه می‌دهد.

## Security invariants

- cookie mutations با CSRF binding محافظت می‌شوند؛
- session cookie از `__Host-` + Secure + HttpOnly + SameSite Strict استفاده می‌کند؛
- raw Runtime password یا Subscription token نباید در list/log/audit برگردد؛
- Host header منبع canonical Naive destination نیست؛
- request body محدود و JSON حساس strict decode می‌شود؛
- route access role به‌تنهایی جای tenant/IDOR test را نمی‌گیرد؛
- generic commit-before-success و refresh reuse-family bug هنوز باید fix شوند؛
- readiness هنوز باید DB/schema-backed شود.

## Current source-of-truth links داخل Repository

1. `internal/httpapi/routes.go`
2. `internal/httpapi/server.go`
3. `internal/httpapi/customer_extra_routes.go`
4. `PROJECT_STATUS.md`
5. `FEATURE_MATRIX.md`
6. `docs/PANEL_PARITY_MASTER_2026-08-30.md`
7. `KNOWN_ISSUES.md`

این سند contract summary است؛ در اختلاف، implementation + tests + current Production evidence بر متن aspirational قدیمی اولویت دارد.
