import { FormEvent, useEffect, useState } from "react";
import { AuthError, login, logout, me, Principal, readCookie } from "./auth";
import { ProductCatalog } from "./ProductCatalog";
import { ProductCustomers } from "./ProductCustomers";
import { RuntimeAdoption } from "./RuntimeAdoption";
import { RuntimeNaive } from "./RuntimeNaive";
import { canUseCustomerProduct, canUseRawRuntime } from "./productPanelModel";
import { assertRouteManifest } from "./routes";

assertRouteManifest();

type Theme = "system" | "dark" | "light";
type AuthState = "loading" | "anonymous" | "authenticated";
type View = "dashboard" | "customers" | "catalog" | "runtime-adoption" | "runtime-naive";

export function currentView(hash = window.location.hash): View {
  if (hash === "#/customers") return "customers";
  if (hash === "#/catalog" || hash === "#/plans") return "catalog";
  if (hash === "#/customers/runtime-adoption") return "runtime-adoption";
  if (hash === "#/runtime/naive") return "runtime-naive";
  return "dashboard";
}

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

function LoginScreen({ onAuthenticated }: { onAuthenticated: (principal: Principal) => void }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [requiresMFA, setRequiresMFA] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setSubmitting(true); setMessage("");
    try {
      await login({ email, password, totpCode: requiresMFA ? totpCode : undefined });
      setPassword(""); setTotpCode("");
      onAuthenticated(await me());
    } catch (cause) {
      const error = cause as AuthError;
      if (error.code === "mfa_required") { setRequiresMFA(true); setMessage("کد شش‌رقمی برنامه احراز هویت را وارد کنید."); }
      else setMessage("ورود انجام نشد. اطلاعات ورود را بررسی کنید.");
    } finally { setSubmitting(false); }
  }

  return <main className="auth-page"><section className="auth-card" aria-labelledby="login-title">
    <div className="brand auth-brand"><img src="/pvnaive-mark.svg" alt="" width="48" height="48"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div>
    <p className="eyebrow">Secure control plane</p><h1 id="login-title">ورود به پنل</h1><p className="auth-copy">ورود مدیریت با نشست امن، CSRF و احراز هویت دومرحله‌ای.</p>
    <form className="auth-form" onSubmit={submit}><label>ایمیل<input type="email" autoComplete="username" value={email} onChange={(event) => setEmail(event.target.value)} required disabled={submitting}/></label><label>رمز عبور<input type="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} required disabled={submitting}/></label>{requiresMFA && <label>کد TOTP<input inputMode="numeric" pattern="[0-9]{6}" maxLength={6} autoComplete="one-time-code" value={totpCode} onChange={(event) => setTotpCode(event.target.value.replace(/\D/g, ""))} required disabled={submitting}/></label>}{message && <p className="auth-message" role="status">{message}</p>}<button type="submit" disabled={submitting || (requiresMFA && totpCode.length !== 6)}>{submitting ? "در حال بررسی…" : requiresMFA ? "تأیید و ورود" : "ورود"}</button></form>
  </section></main>;
}

function Sidebar({ principal, signOut }: { principal: Principal; signOut: () => Promise<void> }) {
  const customerProduct = canUseCustomerProduct(principal.role);
  const rawRuntime = canUseRawRuntime(principal.role);
  return <aside className="sidebar"><div className="brand"><img src="/pvnaive-mark.svg" alt="" width="44" height="44"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div><nav aria-label="ناوبری اصلی"><a href="/panel/">داشبورد</a>{customerProduct && <><a href="/panel/#/customers">کاربران</a><a href="/panel/#/catalog">پلن‌ها / گروه‌ها / تگ‌ها</a></>}{rawRuntime && <><a href="/panel/#/customers/runtime-adoption">اکانت‌های قدیمی Runtime</a><a href="/panel/#/runtime/naive">Runtime پیشرفته</a></>}</nav><ThemeSwitch/><button className="logout-button" onClick={signOut}>خروج امن</button><div className="stage">PVN-044 · WS1 + WS2 product UI</div></aside>;
}

function MobileNav({ principal, signOut }: { principal: Principal; signOut: () => Promise<void> }) {
  const customerProduct = canUseCustomerProduct(principal.role);
  const rawRuntime = canUseRawRuntime(principal.role);
  return <nav className="mobile-nav" aria-label="ناوبری موبایل"><a href="/panel/">داشبورد</a>{customerProduct && <a href="/panel/#/customers">کاربران</a>}{customerProduct && <a href="/panel/#/catalog">پلن‌ها</a>}{rawRuntime && <a href="/panel/#/runtime/naive">Runtime</a>}<button onClick={signOut}>خروج</button></nav>;
}

function Shell({ principal, signOut, children }: { principal: Principal; signOut: () => Promise<void>; children: React.ReactNode }) {
  return <div className="shell"><Sidebar principal={principal} signOut={signOut}/>{children}<MobileNav principal={principal} signOut={signOut}/></div>;
}

export function App() {
  const [authState, setAuthState] = useState<AuthState>("loading");
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [authMessage, setAuthMessage] = useState("");
  const [view, setView] = useState<View>(currentView);
  useEffect(() => { const onHash = () => setView(currentView()); window.addEventListener("hashchange", onHash); return () => window.removeEventListener("hashchange", onHash); }, []);
  useEffect(() => { let active = true; me().then((current) => { if (!active) return; setPrincipal(current); setAuthState("authenticated"); }).catch((cause: AuthError) => { if (!active) return; setPrincipal(null); setAuthState("anonymous"); if (cause.status && cause.status !== 401) setAuthMessage("سرویس احراز هویت در دسترس نیست."); }); return () => { active = false; }; }, []);

  async function signOut() {
    const csrf = readCookie("__Host-pvnaive_csrf");
    if (!csrf) { setPrincipal(null); setAuthState("anonymous"); return; }
    try { await logout(csrf); } finally { setPrincipal(null); setAuthState("anonymous"); }
  }

  if (authState === "loading") return <main className="auth-page"><section className="auth-card"><p>در حال بررسی نشست امن…</p></section></main>;
  if (authState === "anonymous" || !principal) return <><LoginScreen onAuthenticated={(current) => { setPrincipal(current); setAuthState("authenticated"); setAuthMessage(""); }}/>{authMessage && <div className="global-auth-message" role="alert">{authMessage}</div>}</>;

  const customerProduct = canUseCustomerProduct(principal.role);
  const rawRuntime = canUseRawRuntime(principal.role);

  if (view === "customers" && customerProduct) return <Shell principal={principal} signOut={signOut}><ProductCustomers role={principal.role}/></Shell>;
  if (view === "catalog" && customerProduct) return <Shell principal={principal} signOut={signOut}><ProductCatalog role={principal.role}/></Shell>;
  if (view === "runtime-adoption" && rawRuntime) return <Shell principal={principal} signOut={signOut}><RuntimeAdoption/></Shell>;
  if (view === "runtime-naive" && rawRuntime) return <Shell principal={principal} signOut={signOut}><RuntimeNaive/></Shell>;

  const metrics = [
    ["Customer Product", customerProduct ? "فعال" : "طبق RBAC"],
    ["Subscription / QR", customerProduct ? "فعال" : "طبق RBAC"],
    ["Accounting دقیق", "فعال · capability-gated"],
    ["Runtime خام", rawRuntime ? "Owner-only" : "محدود"],
  ];

  return <Shell principal={principal} signOut={signOut}><main><header><div><p className="eyebrow">Standalone control plane</p><h1>داشبورد PVNaive</h1></div><span className="badge">ورود امن فعال</span></header><section className="notice" role="status">{principal.display_name} · {principal.email} · نقش: {principal.role}</section><section className="metrics" aria-label="شاخص‌ها">{metrics.map(([label,value]) => <article key={label}><span>{label}</span><strong>{value}</strong></article>)}</section>{customerProduct ? <><section className="panel"><div><p className="eyebrow">Customer product</p><h2>مدیریت کاربران، پلن و تمدید</h2><p>Plan، Renewal، Group/Tag، Next Plan، On hold، Bulk Preview/Execute، Subscription و Accounting دقیق حالا از مسیر Product API در پنل قابل استفاده‌اند.</p></div><a className="button-secondary" href="/panel/#/customers">مدیریت کاربران</a></section><section className="panel"><div><p className="eyebrow">Catalog</p><h2>پلن‌ها، گروه‌ها و تگ‌ها</h2><p>{principal.role === "reseller" ? "پلن‌های فعال را ببین و برای مشتری استفاده کن؛ ساخت Plan طبق RBAC برای Owner/Admin است." : "پلن‌های آماده و دسته‌بندی مشتریان را از خود پنل مدیریت کن."}</p></div><a className="button-secondary" href="/panel/#/catalog">بازکردن کاتالوگ</a></section></> : <section className="panel"><div><p className="eyebrow">RBAC</p><h2>دسترسی عملیاتی محدود است</h2><p>این نقش برای عملیات Customer Product مجاز نیست. صفحه‌هایی که Backend آن‌ها هنوز Ready نیست نیز عمداً در منو نمایش داده نمی‌شوند.</p></div><button disabled>طبق نقش فعلی</button></section>}</main></Shell>;
}
