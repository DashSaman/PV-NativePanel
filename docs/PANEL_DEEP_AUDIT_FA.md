# ممیزی عمیق پنل‌ها برای PVNaive

تاریخ Snapshot: 2026-08-26

## روش

قابلیت صرفاً براساس README ثبت نشده؛ مدل User، جدول کاربر، فیلتر، action، traffic cell و API پنل‌های مرجع بررسی شده است. SHAهای مشاهده‌شده:

- 3x-ui: `f727d04f6522bb94a8fb52e8352fdcafb51c11e1`
- PasarGuard Panel: `e81877c0df64e5f5235f4355b0490b6bb38e3adc`
- Marzban: `7f396db3e703d71a28060bc9ce4a532ec64cb1f4`
- Remnawave Panel: `545e9a484bad9bc8d538aa79a364a651c1ae4b5f`
- PVNetwork قابل دسترس: `DashSaman/OV-PvNetwork`

سورس دقیق پنل Production تصویر 3X هنوز در دسترس نیست؛ قابلیت‌های آن از تصویر و مستندات خود پروژه استخراج شده، نه ممیزی ادعایی سورس.

## 3x-ui

### User/Client

- چند Client داخل یک inbound و اتصال Client به چند inbound
- enable/disable مستقل
- حجم upload، download، used، total و remaining
- unlimited traffic
- expiry و delayed-start
- subscription استاندارد/JSON/Clash، QR و download
- IP log و HWID management
- Telegram ID، comment، group، external link
- realtime speed و online set با WebSocket
- import/export و bulk add

### دسته‌بندی و فیلتر

Bucketهای `active`، `deactive`، `depleted` و `expiring`. فیلتر پروتکل، inbound، node، group، بازه expiry، بازه usage، auto-renew، Telegram ID و comment. Sort براساس creation/update/last-online/email/traffic/remaining/expiry.

### عملیات

Edit، delete، enable/disable، reset traffic، attach/detach inbound، group، QR، info، subscription و bulk action. Traffic Cell، upload/download و remaining را جدا نشان می‌دهد.

### درس

فیلتر و Bulk برای هزاران کاربر الزامی است؛ ولی status حساب نباید فقط از enabled boolean نتیجه شود.

## PasarGuard

### Status و User

Statusهای active، disabled، limited، expired و on_hold. داده‌ها شامل used traffic، lifetime traffic، data limit، expire، online_at، edit_at، owner/admin، group، note، HWID limit، next plan و reset strategy است.

### جستجو و جدول

- Search username، ID، protocol ID و حتی subscription URL
- filter status، owner، group، online
- min/max حجم
- بازه expiry و online
- no limit/no expiry
- sort username/created/edit/expire/usage/last-online
- auto refresh خاموش، 5، 15، 30 و 60 ثانیه
- تنظیم ستون‌ها، page size و checkbox
- URL-persisted filters

### Bulk

Enable، disable، reset usage، revoke subscription، delete، apply template و set owner.

### درس

On-hold و Next Plan برای فروش واقعی مفید است. Permission باید action-level و scope-aware باشد.

## Marzban

Statusهای active، disabled، limited، expired و on_hold. Reset strategy: no_reset/day/week/month/year. مدل همچنین used traffic، lifetime traffic، subscription last user-agent، online_at، auto delete، next plan و چند protocol/inbound دارد.

درس: status محاسباتی limited/expired باید از status دستوری active/disabled جدا باشد؛ کاربر نباید limited را مستقیماً به active تبدیل کند بدون رفع علت.

## Remnawave

الگوی قوی در users/nodes/hosts/subscription/routing و عملیات چندنودی دارد. در MVP مستقل، فقط الگوهای Host abstraction، Subscription page، audit و deployment hygiene استفاده می‌شوند؛ Fleet به فاز بعد می‌رود.

## PVNetwork / OV-PvNetwork

- desired-state و reconciliation
- online/offline و داشبورد realtime
- user/node assignment
- node recommendation
- bandwidth control
- domain activity/history اختیاری
- monitoring/Telegram
- drain/maintenance/canary
- update/backup/rollback/doctor
- محاسبه سرعت server-side از cumulative counter و elapsed time واقعی
- Node TX = دانلود کاربر و Node RX = آپلود کاربر

درس: collector اختیاری نباید bootstrap یا data plane را متوقف کند.

## نتیجه برای PVNaive

PVNaive باید این چهار محور را جدا نگه دارد:

1. **Account lifecycle:** وضعیت مجازبودن حساب
2. **Presence:** Online/Idle/Offline/Unknown
3. **Quota:** Unlimited/Healthy/Warning/Depleted
4. **Runtime health:** Healthy/Degraded/Down/Unknown

ترکیب این محورها فقط برای badge/summary است؛ یکی جای دیگری را تغییر نمی‌دهد.

