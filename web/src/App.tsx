import { useEffect, useState } from "react";
import { appRoutes, assertRouteManifest } from "./routes";

assertRouteManifest();

type Theme = "system" | "dark" | "light";
const metrics = [["وضعیت Runtime","در انتظار اتصال"],["کاربران فعال","—"],["مصرف امروز","—"],["اعتبار TLS","—"]];
const mobileRoutes = appRoutes.filter((r) => r.navigation).slice(0, 3);

function ThemeSwitch() {
  const [theme, setTheme] = useState<Theme>(() => (localStorage.getItem("pvnaive.theme") as Theme) || "system");
  useEffect(() => {
    if (theme === "system") delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = theme;
    localStorage.setItem("pvnaive.theme", theme);
  }, [theme]);
  const next = theme === "system" ? "dark" : theme === "dark" ? "light" : "system";
  return <button className="theme-switch" onClick={() => setTheme(next)} aria-label="تغییر پوسته">پوسته: {theme}</button>;
}

export function App() {
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand"><img src="/pvnaive-mark.svg" alt="" width="44" height="44"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div>
        <nav aria-label="ناوبری اصلی">{appRoutes.filter((r) => r.navigation).map((r) => <a href={r.path} key={r.path}>{r.label}</a>)}</nav>
        <ThemeSwitch />
        <div className="stage">Engineering preview · 0.0.2</div>
      </aside>
      <main>
        <header><div><p className="eyebrow">Standalone control plane</p><h1>داشبورد PVNaive</h1></div><span className="badge">در حال پیاده‌سازی</span></header>
        <section className="notice" role="status">Backend تجاری هنوز فعال نشده است؛ داده ساختگی نمایش داده نمی‌شود.</section>
        <section className="metrics" aria-label="شاخص‌ها">{metrics.map(([label,value]) => <article key={label}><span>{label}</span><strong>{value}</strong></article>)}</section>
        <section className="panel"><div><p className="eyebrow">اولویت مهندسی</p><h2>Accounting قابل اتکا قبل از فروش</h2><p>محاسبه حجم هر credential، restart و H2 multiplex باید در PoC اثبات شود.</p></div><button disabled>ایجاد کاربر</button></section>
      </main>
      <nav className="mobile-nav" aria-label="ناوبری موبایل">
        {mobileRoutes.map((r) => <a href={r.path} key={r.path}>{r.label}</a>)}
        <a href="/system">بیشتر</a>
      </nav>
    </div>
  );
}
