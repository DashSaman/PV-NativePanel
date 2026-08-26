# قرارداد توسعه‌پذیری و چندپروتکلی PVNative

## اصل مرکزی

User، Plan، Quota، Credential و Subscription مستقل از پروتکل هستند. Naive فقط اولین Runtime Adapter است، نه هویت دیتامدل.

## مرزها

```text
Domain Core
  ├── Runtime Adapter
  ├── Usage Collector
  ├── Subscription Renderer
  ├── Health Probe
  └── Config Validator/Applier
```

هر Adapter یک manifest نسخه‌دار منتشر می‌کند:

- protocol_id و adapter_version
- config_schema_version
- capability flags
- supported clients/renderers
- health semantics
- counter semantics: monotonic/resettable/unknown
- migration compatibility
- secret fields
- minimum runtime version

## Capabilityها

`per_credential_usage`، `live_sessions`، `speed_limit`، `device_limit`، `atomic_reload`، `zero_downtime_reload`، `padding_control`، `destination_metadata` و `multi_endpoint`.

UI و API فقط capability واقعی را نشان می‌دهند. Unsupported هرگز با مقدار ساختگی نمایش داده نمی‌شود.

## قواعد امنیت

- Adapter داخلی با interface کامپایل‌شده؛ plugin دلخواه داخل process ممنوع.
- توسعه شخص ثالث آینده فقط sidecar محدود با mTLS، allowlist و API versioned.
- secret از manifest/log/export حذف می‌شود.
- config قبل از apply: render، schema validation، semantic validation، dry-run و atomic swap.
- شکست reload باید config قبلی را حفظ کند.
- billing مستقیماً به counter خام Adapter وابسته نمی‌شود؛ Usage Ledger نرمال‌شده لازم است.

## افزودن پروتکل جدید

1. ADR و threat model
2. manifest و capability matrix
3. config schema و migration
4. runtime adapter
5. usage normalization
6. subscription renderer
7. health/failure tests
8. compatibility matrix کلاینت
9. canary و rollback
10. فعال‌سازی feature flag

پروتکل دوم باید این قرارداد را اثبات کند؛ پیشنهاد برای آزمون معماری، یک Adapter آزمایشی fake در تست‌هاست، نه اتصال فوری Xray به MVP.
