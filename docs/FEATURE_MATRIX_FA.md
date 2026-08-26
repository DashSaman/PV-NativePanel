# ماتریس قابلیت‌ها

| قابلیت | 3x-ui | PasarGuard | Marzban | PV/OV patterns | NativePanel هدف |
|---|---:|---:|---:|---:|---:|
| تک‌سرور | بله | بله | بله | بله | MVP |
| چندنود | بله، panel-to-panel | بله | بله | controller/agent | MVP |
| mTLS نود | بله | بررسی جزئیات لازم | بررسی جزئیات لازم | هدف | MVP |
| desired-state reconcile | محدود/مدیریتی | sync | sync | بله | MVP |
| rollback last-good | نامشخص | نامشخص | نامشخص | الگو | MVP |
| Managed Host | بله | Host دارد | Host/subscription | الگو | MVP |
| quota/expiry/reset | بله | بله | بله | بله | MVP |
| آمار دقیق per-user | Xray-based | Xray-based | Xray-based | runtime-specific | **PoC blocker** |
| RBAC/reseller scope | قابل بررسی | بله | محدود/WIP در README | بله | MVP |
| routing/balancer | Xray JSON + test | Xray config | Xray config | node policy | Phase 2 |
| drain/canary/maintenance | بخشی | قابل بررسی | قابل بررسی | بله | MVP |
| Subscription چندفرمتی | بله | بله | بله | بله | MVP |
| Naive استاندارد | خارج از هسته | خیر | خیر | خیر | هسته |
| نصب یک‌فرمانی | بله | بله | بله | بله | پس از MVP پایدار |

## قابلیت‌هایی که عمداً وارد MVP نمی‌شوند

- ساخت پروتکل شبکه اختصاصی
- packet routing پیچیده شبیه Xray
- CDN orange-cloud برای data plane
- اپ اختصاصی همه پلتفرم‌ها
- billing مالی کامل
- ادعای device limit قطعی بدون client cooperation

ابتدا باید account lifecycle، multi-node، accounting، subscription و عملیات failover درست شوند.
