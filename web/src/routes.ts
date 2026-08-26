export type Permission = "public" | "authenticated" | "reseller" | "admin" | "operator" | "auditor" | "owner";
export type AppRoute = { path: string; label: string; permission: Permission; navigation: boolean };

export const appRoutes: AppRoute[] = [
  { path: "/login", label: "ورود", permission: "public", navigation: false },
  { path: "/setup", label: "راه‌اندازی اولیه", permission: "public", navigation: false },
  { path: "/s/:token", label: "صفحه اشتراک", permission: "public", navigation: false },
  { path: "/", label: "داشبورد", permission: "authenticated", navigation: true },
  { path: "/users", label: "کاربران", permission: "reseller", navigation: true },
  { path: "/users/:id", label: "جزئیات کاربر", permission: "reseller", navigation: false },
  { path: "/resellers", label: "نمایندگان", permission: "admin", navigation: true },
  { path: "/resellers/:id", label: "جزئیات نماینده", permission: "admin", navigation: false },
  { path: "/plans", label: "پلن‌ها", permission: "reseller", navigation: true },
  { path: "/subscriptions", label: "اشتراک‌ها", permission: "reseller", navigation: true },
  { path: "/notifications", label: "اعلان‌ها", permission: "admin", navigation: true },
  { path: "/runtime", label: "Naive Runtime", permission: "admin", navigation: true },
  { path: "/usage", label: "حجم و مصرف", permission: "auditor", navigation: true },
  { path: "/system", label: "وضعیت سیستم", permission: "operator", navigation: true },
  { path: "/logs/application", label: "لاگ برنامه", permission: "operator", navigation: true },
  { path: "/logs/runtime", label: "لاگ Runtime", permission: "operator", navigation: false },
  { path: "/logs/security", label: "لاگ امنیت", permission: "auditor", navigation: false },
  { path: "/diagnostics/domain-activity", label: "فعالیت دامنه‌ها", permission: "owner", navigation: false },
  { path: "/diagnostics/requests/:requestId", label: "ردیابی درخواست", permission: "operator", navigation: false },
  { path: "/audit", label: "گزارش امنیتی", permission: "auditor", navigation: true },
  { path: "/settings/security", label: "امنیت", permission: "owner", navigation: true },
  { path: "/settings/appearance", label: "ظاهر", permission: "authenticated", navigation: false },
  { path: "/settings/backup", label: "پشتیبان‌گیری", permission: "owner", navigation: true }
];

export function assertRouteManifest(): void {
  const paths = new Set<string>();
  for (const route of appRoutes) {
    if (paths.has(route.path)) throw new Error(`Duplicate route: ${route.path}`);
    paths.add(route.path);
    if (route.navigation && route.permission === "public") throw new Error(`Public route cannot be in authenticated navigation: ${route.path}`);
    if (route.path.includes("domain-activity") && route.permission !== "owner") throw new Error("Domain activity must remain owner-only.");
  }
}
