# معماری قابل توسعه Protocol و Runtime

## هدف

Naive پروتکل اول است، نه فرض ثابت کل سیستم. افزودن Hysteria2، TUIC، AnyTLS، WireGuard یا Xray-family نباید User/Quota/UI/Subscription را بازنویسی کند.

## جداسازی Domain

- User و Plan مستقل از protocol
- Credential یک هویت دسترسی است
- ProtocolBinding credential را به protocol وصل می‌کند
- ProfileVariant یک گزینه نمایش در Subscription است
- RuntimeInstance موتور اجراست
- Listener منبع مشترک IP/port/TLS است
- UsageLedger مستقل از runtime است
- SubscriptionRenderer خروجی هر client/format را می‌سازد

## Adapter contract

هر Adapter باید اعلام کند:

- ID/version
- runtime family
- transport requirements
- credential schema
- config schema و UI schema
- listener compatibility
- subscription formats
- client compatibility
- accounting: exact/estimated/none
- session visibility
- speed enforcement
- concurrency/device/IP capabilities
- validate/render/apply/rollback/health
- migration compatibility

اگر accounting برابر exact نباشد، quota enforcement با warning یا ممنوعیت Production روبه‌رو می‌شود.

## Registry

نسخه اول adapterها compile-time و signed هستند. Go `.so` plugin یا اجرای کد third-party داخل API ممنوع است. در آینده extension خارجی فقط به‌صورت sidecar sandboxed با API versioned، mTLS، timeout، resource limit و capability allowlist اضافه می‌شود.

## Schema-driven UI

Protocol form از schema versioned ساخته می‌شود، اما fieldهای حساس widget امن دارند و raw HTML/JS از Adapter پذیرفته نمی‌شود. Advanced JSON فقط Owner، با schema validation، diff، redaction و rollback.

## Listener conflict

قبل از apply یک Global Listener Registry ترکیب IP/port/network/TLS/ALPN را بررسی می‌کند. هیچ Adapter مستقلاً port bind نمی‌کند بدون reservation. conflict باید پیش از restart کشف شود.

## Migration

- schema version برای هر binding
- plan: inspect → transform → validate → stage → canary → commit
- rollback adapter قبلی تا پایان observation window
- subscription dual-publish اختیاری در migration
- credential rotation مستقل از migration
- عدم حذف protocol قدیمی تا صفرشدن assignment

## Compatibility matrix

هر release فایل machine-readable دارد:

`protocol × server-version × client × OS × subscription-format × tested-result`

وضعیت‌ها: supported، experimental، deprecated، blocked.

## Security

Adapter secret را log نمی‌کند؛ command shell آزاد ندارد؛ path و URL validation؛ SSRF deny-by-default؛ egress allowlist؛ signed artifact؛ SBOM؛ resource budget و circuit breaker.
