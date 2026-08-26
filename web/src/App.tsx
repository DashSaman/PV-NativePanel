import { appRoutes, assertRouteManifest } from "./routes";

assertRouteManifest();

const metrics = [
  ["وضعیت Runtime", "در انتظار اتصال"],
  ["کاربران فعال", "—"],
  ["مصرف امروز", "—"],
  ["اعتبار TLS", "—"]
];

export function App() {
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <img src="/pvnative-mark.svg" alt="" width="44" height="44" />
          <div><strong>PVNative</strong><span>PVNETWORK</span></div>
        </div>
        <nav aria-label="ناوبری اصلی">
          {appRoutes.filter((route) => route.navigation).map((route) => (
            <a href={route.path} key={route.path}>{route.label}</a>
          ))}
        </nav>
        <div className="stage">Research scaffold · 0.0.1</div>
      </aside>

      <main>
        <header>
          <div>
            <p className="eyebrow">Standalone control plane</p>
            <h1>داشبورد PVNative</h1>
          </div>
          <span className="badge">پیاده‌سازی نشده</span>
        </header>

        <section className="notice" role="status">
          این رابط فعلاً اسکلت طراحی است. هیچ داده یا عملیات Production شبیه‌سازی نشده است.
        </section>

        <section className="metrics" aria-label="شاخص‌ها">
          {metrics.map(([label, value]) => (
            <article key={label}><span>{label}</span><strong>{value}</strong></article>
          ))}
        </section>

        <section className="panel">
          <div>
            <p className="eyebrow">اولویت مهندسی</p>
            <h2>Accounting قابل اتکا قبل از فروش</h2>
            <p>محاسبه حجم هر credential، restart و H2 multiplex باید در PoC اثبات شود.</p>
          </div>
          <button disabled>ایجاد کاربر</button>
        </section>
      </main>
    </div>
  );
}
