export type AppRoute = {
  path: string;
  label: string;
  permission: "public" | "authenticated" | "admin" | "operator" | "auditor" | "owner";
  navigation: boolean;
};

export const appRoutes: AppRoute[] = [
  { path: "/login", label: "ورود", permission: "public", navigation: false },
  { path: "/", label: "داشبورد", permission: "authenticated", navigation: true },
  { path: "/users", label: "کاربران", permission: "admin", navigation: true },
  { path: "/users/:id", label: "جزئیات کاربر", permission: "admin", navigation: false },
  { path: "/subscriptions", label: "اشتراک‌ها", permission: "admin", navigation: true },
  { path: "/runtime", label: "Naive Runtime", permission: "admin", navigation: true },
  { path: "/usage", label: "حجم و مصرف", permission: "auditor", navigation: true },
  { path: "/system", label: "وضعیت سیستم", permission: "operator", navigation: true },
  { path: "/audit", label: "گزارش امنیتی", permission: "auditor", navigation: true },
  { path: "/settings/security", label: "امنیت", permission: "owner", navigation: true },
  { path: "/settings/backup", label: "پشتیبان‌گیری", permission: "owner", navigation: true },
  { path: "/setup", label: "راه‌اندازی اولیه", permission: "public", navigation: false }
];

export function assertRouteManifest(): void {
  const paths = new Set<string>();
  for (const route of appRoutes) {
    if (paths.has(route.path)) throw new Error(`Duplicate route: ${route.path}`);
    paths.add(route.path);
    if (route.navigation && route.permission === "public") {
      throw new Error(`Public route cannot be in authenticated navigation: ${route.path}`);
    }
  }
}
