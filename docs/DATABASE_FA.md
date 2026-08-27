# طراحی دیتابیس PostgreSQL — S03

این سند رفتار واقعی پیاده‌سازی S03 در PVNaive را ثبت می‌کند. S03 تا زمانی که اسکریپت Stage روی `testAmir5-3` با خروجی `S03_RESULT=PASSED` تمام نشود، همچنان `NEXT` است و پروژه Production-ready نیست.

## مدل داده

Schema برنامه `pvnaive` است و Migration اولیهٔ نسخه‌دار در `db/migrations/0001_initial.up.sql` قرار دارد.

| حوزه | جدول‌ها و رفتار اصلی |
|---|---|
| هویت و دسترسی | `tenants`، `actors` برای owner/admin/operator/auditor/reseller، و `auth_sessions` با token hash و refresh family |
| نمایندگی | `resellers`، `reseller_plan_terms` و `reseller_credit_ledger`؛ ledger مالی append-only، سریال‌شده و دارای idempotency key است |
| فروش | `users`، `plans`، `purchases`، `subscriptions` و `quota_policies`؛ plan خرید باید در terms همان نماینده مجاز باشد |
| دسترسی کاربر | `subscription_tokens` فقط با hash، و `credentials` با hash به‌علاوه ciphertext/nonce/key-id برای envelope encryption آینده |
| مصرف | `sessions`، `usage_ledger`، `usage_reset_events` و `renewal_events`؛ deltaها با boot-id/sequence idempotent و ledgerها append-only هستند |
| اعلان | `notification_rules`، `notification_outbox` و `notification_deliveries` با deduplication key، idempotency key، attempt limit و retry state |
| Runtime و عملیات | `runtime_revisions`، `runtime_health`، `audit_events`، `log_metadata` و `backups` |
| Schema | `schema_migrations` با version، filename و SHA-256 تغییرناپذیر |

Viewهای `subscription_usage` و `subscription_summary` حجم upload/download/total، حجم باقی‌مانده، درصد مصرف، reset/expiry، first connection، limit اتصال/دستگاه و زمان آخرین به‌روزرسانی را برای Backend صفحه Subscription فراهم می‌کنند. روز و ساعت باقی‌مانده و deep-link در Backend/renderer ساخته می‌شوند، نه در SQL.

## Tenant isolation

ایزوله‌سازی فقط UI نیست و سه لایه دارد:

1. Foreign keyهای composite جلوی اتصال User/Subscription/Credential/Session/Usage/Purchase از دو tenant متفاوت را می‌گیرند.
2. روی ۲۵ جدول tenant-aware، RLS فعال است؛ application role دارای `NOBYPASSRLS` است و بیشتر جدول‌ها `FORCE ROW LEVEL SECURITY` دارند.
3. Backend باید هر درخواست را در یک `sql.Tx` اجرا و با `database.BindRequestContext` فقط SHA-256 session token را bind کند. PostgreSQL tenant و role را از session فعال استخراج و context را با کلید ۲۵۶بیتی داخلی امضا می‌کند. تغییر دستی custom GUC بدون signature معتبر fail-closed است.

`auth_sessions` تطابق tenant و actor را با trigger بررسی می‌کند تا تغییر `actor_id` نتواند reseller را به owner تبدیل کند. Global plan برای نماینده قابل مشاهده است، اما global runtime، audit، log، backup و notification state برای reseller قابل مشاهده نیست.

RLS جای RBAC endpoint را نمی‌گیرد. Route و business layer همچنان باید role/capability را deny-by-default بررسی کنند؛ helper فعلی فقط boundary امن query را آماده کرده است.

## Migration و Rollback

`scripts/db/migrate.sh` پیش از هر SQL این gateها را اعمال می‌کند:

- تمام فایل‌ها و manifest با SHA-256 معتبر باشند؛ فایل unlisted پذیرفته نمی‌شود.
- شماره‌ها از `0001` پیوسته باشند و هر up یک down checksummed داشته باشد.
- Migration فقط transactional و `destructive false` باشد.
- الگوهای DROP/TRUNCATE/DELETE/ALTER-DROP/COPY-PROGRAM و include خارجی رد شوند.
- Migration اعمال‌شده با filename/checksum ثبت‌شده دقیقاً یکسان باشد.
- advisory transaction lock از اجرای همزمان جلوگیری کند.

`scripts/db/rollback.sh` عمداً destructive است. روی دیتابیس واقعی فقط با تأیید صریح، backup رمزنگاری‌شده، manifest سالم، metadata سالم، schema version یکسان، کلید age و parse موفق archive اجرا می‌شود. اجرای down روی دیتابیس disposable تست از gate جداگانه استفاده می‌کند.

## Secret، Backup و Restore

- password برنامه با ۲۵۶ بیت randomness ساخته می‌شود و فقط در `/etc/pvnaive/db.pgpass` با mode `0600` و owner `pvnaive` قرار می‌گیرد.
- `/etc/pvnaive/db.env` secret ندارد و فقط host/port/name/user/schema version و مسیر pgpass را نگه می‌دارد.
- PostgreSQL فقط روی `127.0.0.1` و `::1` گوش می‌دهد؛ HBA فقط SCRAM-SHA-256 را برای `pvnaive_app` می‌پذیرد.
- backup با custom-format `pg_dump` ساخته، با age رمزنگاری و سپس نسخهٔ plaintext حذف می‌شود.
- `metadata.json` و archive هر دو داخل `SHA256SUMS` هستند.
- restore فقط به نام جدید `pvnaive_restore_test_*` مجاز است؛ دیتابیس موجود overwrite نمی‌شود و schema/data بعد از restore بررسی می‌شوند.

## Health و systemd

`scripts/db/health.sh` اتصال، schema version، ۲۶ جدول ضروری، ۲۵ جدول RLS، نبود Migration مخرب، عدم دسترسی app role به کلید امضای context، فعال‌بودن row security و loopback بودن اتصال را بررسی می‌کند.

Timer یک‌دقیقه‌ای `pvnaive-db-health.timer` فقط health check را با user بدون shell `pvnaive` اجرا می‌کند. Unit دارای filesystem/process/network restriction، empty capability set، `NoNewPrivileges` و system-call allowlist است. Unit توزیع PostgreSQL بازنویسی نمی‌شود.

## Rollback خود Stage

پیش از تغییر، HBA، `postgresql.auto.conf`، نسخه packageها و cluster state در `/var/backups/pvnaive/<timestamp>-S03-pre` با checksum ذخیره می‌شوند. در failure، Stage دیتابیس/role/secret/unit/release جدید همان اجرا را حذف، config قبلی PostgreSQL را restore و فقط PostgreSQL را restart می‌کند. Packageهای pin‌شده برای بررسی باقی می‌مانند. Caddy، NaiveProxy، SSH و Firewall هیچ‌وقت توسط S03 تغییر نمی‌کنند.
