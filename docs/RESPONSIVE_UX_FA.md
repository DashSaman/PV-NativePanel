# Responsive UX و Accessibility

## Breakpointهای هدف

- Mobile: 360–767px
- Tablet: 768–1023px
- Desktop: 1024–1439px
- Wide: 1440px+

طراحی از 360px شروع می‌شود، نه shrink نسخه دسکتاپ.

## Navigation

- Desktop: sidebar ثابت و collapsible
- Tablet: sidebar فشرده
- Mobile: bottom navigation برای Dashboard/Users/Usage/More
- action اصلی با FAB یا sticky action bar
- Back behavior و deep-link قابل پیش‌بینی

## Users

Desktop table؛ Mobile card با status، presence، used/limit، expiry و action menu. Bulk selection روی موبایل وارد selection mode می‌شود و نوار sticky نشان می‌دهد. ستون‌های ثانویه به Details منتقل می‌شوند، نه horizontal scroll اجباری.

## Form

یک ستون روی موبایل، دو ستون روی desktop؛ sectionهای Identity/Plan/Access/Subscription/Advanced. تغییرات unsaved warning؛ validation کنار field و summary بالا. Secret با reveal/copy محدود و countdown.

## Charts

Canvas/SVG responsive، tooltip قابل لمس، جدول جایگزین برای accessibility، عدم reliance فقط به hover.

## Accessibility

- WCAG AA
- keyboard navigation و focus visible
- aria-label برای icon-only
- status علاوه بر رنگ با text/icon
- target لمسی حداقل 44×44
- reduced motion
- RTL/LTR واقعی
- تاریخ/عدد locale-aware
- screen-reader announcement برای mutation و error

## Performance

- server-side pagination
- virtualize فقط در صورت نیاز
- WebSocket update تجمیعی؛ هر counter باعث rerender کل جدول نشود
- pause realtime در tab مخفی
- skeleton بدون layout shift
- code split صفحات سنگین
- budget اولیه: JS gzip کمتر از 250KB برای shell

## Test matrix

Chromium/Firefox/Safari، Windows/macOS، Android Chrome و iOS Safari؛ عرض‌های 360/390/768/1024/1440، Dark/Light، RTL/LTR، keyboard-only و slow 3G برای Subscription page.
