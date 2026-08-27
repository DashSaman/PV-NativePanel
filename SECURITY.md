# Security Policy

PVNaive هنوز در مرحلهٔ scaffold است و برای Production آماده نیست.

## اصول امنیتی

- secure-by-default و least privilege
- session cookie از نوع HttpOnly، Secure و SameSite
- CSRF protection برای mutationهای مبتنی بر cookie
- Argon2id برای password hash با پارامتر benchmark‌شده
- MFA نوع TOTP و recovery code hash‌شده
- refresh token rotation و تشخیص reuse
- RBAC deny-by-default
- rate limit و progressive delay برای login
- security header، CSP بدون unsafe-inline در build نهایی
- secrets خارج از repository و encrypted-at-rest
- audit append-only برای login، secret، quota، runtime و backup
- config apply اتمیک با validate و rollback
- dependency pin، SBOM، secret scan، SAST و image scan
- backup رمزنگاری‌شده و restore drill

## Threatهای اولویت‌دار

1. credential stuffing و brute force پنل
2. سرقت subscription token
3. IDOR بین کاربران یا resellerها
4. command/config injection در Runtime
5. SSRF در URLها و health checks
6. path traversal در backup/restore
7. double counting یا bypass quota
8. secret leakage در log، metrics و error
9. supply-chain compromise در installer/update
10. قطع دیتاپلین بر اثر خرابی پنل

## گزارش آسیب‌پذیری

Issue عمومی حاوی secret یا جزئیات exploit باز نکنید. تا زمان تعریف کانال رسمی، یافته‌ها مستقیماً و خصوصی به مالک repository گزارش شوند.

## ممنوعیت Production

تا تکمیل threat model، تست authorization، تست restore، PoC accounting و release signing، استقرار عمومی ممنوع است.

