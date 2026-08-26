# مشخصات دقیق UI/UX

## Theme

سه حالت:

- Light
- Dark
- System

انتخاب در local preference ذخیره می‌شود؛ preference حساب نیز در DB ذخیره و پس از login sync می‌شود. قبل از hydrate، theme سیستم اعمال می‌شود تا flash رخ ندهد. تمام رنگ‌ها در هر دو theme باید WCAG AA و علاوه بر رنگ دارای متن/icon باشند.

## Account status

| Status | فارسی | رنگ | رفتار |
|---|---|---|---|
| `draft` | پیش‌نویس | آبی خاکستری | هنوز صادر نشده |
| `on_hold` | در انتظار اولین اتصال | بنفش | expiry از اولین اتصال |
| `active` | فعال | سبز | قابل اتصال |
| `suspended` | غیرفعال دستی | نارنجی | اتصال جدید رد می‌شود |
| `expired` | منقضی | قرمز | ردیف قرمز کم‌رنگ |
| `depleted` | حجم تمام | قرمز تیره | progress کامل قرمز |
| `revoked` | لغوشده | خاکستری تیره | credentialها revoke |
| `error` | خطای اعمال | ارغوانی/قرمز | نیاز به reconcile |

Expired و Depleted هر دو قرمزند ولی icon و متن متفاوت دارند. رنگ تنها نشانه نیست.

## Presence

- Online: نقطه سبز + تعداد Session + سرعت
- Idle: زرد؛ session موجود ولی بدون ترافیک در threshold
- Offline: خاکستری + Last online
- Unknown: خط‌چین/علامت سؤال؛ collector مشکل دارد

Online باید از session/traffic window با hysteresis محاسبه شود، نه یک درخواست لحظه‌ای. Refresh UI باعث Offline شدن نمی‌شود.

## تک‌کاربره و چندکاربره

- `concurrency_limit=1`: تک‌کاربره
- `concurrency_limit=N`: چندکاربره محدود
- `concurrency_limit=null`: نامحدود، فقط با مجوز Owner
- یک User می‌تواند چند Credential داشته باشد.
- هر Credential نام/دستگاه اختیاری، created/last-seen/revoked و session count دارد.
- عبور از limit: session تازه رد می‌شود؛ session قدیمی بدون policy صریح قطع نمی‌شود.
- شمارش device با Credential جدا قابل اتکاتر از IP است.

## Traffic Cell

نمایش ثابت:

- Used / Limit
- Remaining
- Upload و Download در popover/detail
- Lifetime usage
- دوره Reset و زمان Reset بعدی
- درصد با progress
- Unlimited با علامت ∞
- Warning در 80%، Critical در 95%، Depleted در 100%
- اختلاف Accounting با badge `Needs reconciliation`

واحد SI/IEC باید در Settings مشخص و در همه صفحات یکسان باشد.

## جدول Users

ستون‌های قابل انتخاب:

Checkbox، Enabled، Presence، User، Group، Credentials، Sessions، Used/Limit، Remaining، Speed، Expiry، Last online، Reset، Created by، Comment و Actions.

امکانات:

- server-side pagination و cursor
- search username/ID/subscription/credential/comment
- filter status/presence/quota/group/expiry/usage/concurrency
- sort created/updated/online/usage/remaining/expiry/speed
- filterها در URL و saved view
- density compact/comfortable
- desktop table و mobile card
- export فقط با permission و audit
- bulk enable/disable/reset/extend/group/revoke/rotate

## User detail tabs

Overview، Credentials، Sessions، Usage، Subscription، Domain Activity، Audit و Notes.

## Dashboard

کارت‌ها:

Total، Active، Online، Depleted، Expired، Suspended، Runtime status، Traffic rate، Today usage، TLS expiry، Disk و Accounting lag.

نمودارها:

Traffic rate، daily usage، status distribution، errors و top users براساس حجم؛ Domainها فقط وقتی diagnostic فعال است.

## Error/empty/loading states

هر صفحه Skeleton، Empty state، Permission denied، Dependency down، Partial data و Retry دارد. خطا شامل request_id و زمان است؛ stack trace یا secret در UI نمایش داده نمی‌شود.
