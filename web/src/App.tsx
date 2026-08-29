import { FormEvent, useEffect, useState } from "react";
import { AuthError, login, logout, me, Principal, readCookie } from "./auth";
import { Dashboard } from "./Dashboard";
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
  return <button className="theme-switch" onClick={() => setTheme(next)} aria-label="تغییر پوسته">◐ <span>{theme === "dark" ? "تیره" : theme === "light" ? "روشن" : "سیستم"}</span></button>;
}

function LoginScreen({ onAuthenticated }: { onAuthenticated: (principal: Principal) => void }) {
  const [email, setEmail] = useState(""); const [password, setPassword] = useState(""); const [totpCode, setTotpCode] = useState("");
  const [requiresMFA, setRequiresMFA] = useState(false); const [submitting, setSubmitting] = useState(false); const [message, setMessage] = useState("");
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setSubmitting(true); setMessage("");
    try { await login({ email, password, totpCode: requiresMFA ? totpCode : undefined }); setPassword(""); setTotpCode(""); onAuthenticated(await me()); }
    catch (cause) { const error = cause as AuthError; if (error.code === "mfa_required") { setRequiresMFA(true); setMessage("کد شش‌رقمی برنامه احراز هویت را وارد کنید."); } else setMessage("ورود انجام نشد. اطلاعات ورود را بررسی کنید."); }
    finally { setSubmitting(false); }
  }
  return <main className="auth-page"><section className="auth-card" aria-labelledby="login-title"><div className="brand auth-brand"><img src="/pvnaive-mark.svg" alt="" width="48" height="48"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div><p className="eyebrow">Secure Control Panel</p><h1 id="login-title">ورود به پنل</h1><p className="auth-copy">مدیریت امن سرویس و کاربران</p><form className="auth-form" onSubmit={submit}><label>ایمیل<input type="email" autoComplete="username" value={email} onChange={(e) => setEmail(e.target.value)} required disabled={submitting}/></label><label>رمز عبور<input type="password" autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} required disabled={submitting}/></label>{requiresMFA && <label>کد TOTP<input inputMode="numeric" pattern="[0-9]{6}" maxLength={6} autoComplete="one-time-code" value={totpCode} onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ""))} required disabled={submitting}/></label>}{message && <p className="auth-message">{message}</p>}<button type="submit" disabled={submitting || (requiresMFA && totpCode.length !== 6)}>{submitting ? "در حال بررسی…" : "ورود"}</button></form></section></main>;
}

function Sidebar({ principal, view, signOut }: { principal: Principal; view: View; signOut: () => Promise<void> }) {
  const product = canUseCustomerProduct(principal.role); const runtime = canUseRawRuntime(principal.role);
  const linkClass = (active: boolean) => active ? "nav-link active" : "nav-link";
  return <aside className="sidebar"><div className="brand"><img src="/pvnaive-mark.svg" alt="" width="42" height="42"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div><nav aria-label="ناوبری اصلی"><a className={linkClass(view === "dashboard")} href="/panel/"><b>⌂</b><span>داشبورد</span></a>{product && <><a className={linkClass(view === "customers")} href="/panel/#/customers"><b>♙</b><span>کاربران</span></a><a className={linkClass(view === "catalog")} href="/panel/#/catalog"><b>▦</b><span>پلن‌ها و دسته‌بندی</span></a></>}{runtime && <a className={linkClass(view === "runtime-naive" || view === "runtime-adoption")} href="/panel/#/runtime/naive"><b>⚙</b><span>سیستم / Runtime</span></a>}</nav><div className="sidebar-footer"><ThemeSwitch/><button className="logout-button" onClick={signOut}>⇥ <span>خروج امن</span></button><small>{principal.display_name || principal.email}</small></div></aside>;
}

function MobileNav({ principal, view, signOut }: { principal: Principal; view: View; signOut: () => Promise<void> }) {
  const product = canUseCustomerProduct(principal.role); const runtime = canUseRawRuntime(principal.role);
  return <nav className="mobile-nav"><a className={view === "dashboard" ? "active" : ""} href="/panel/">داشبورد</a>{product && <a className={view === "customers" ? "active" : ""} href="/panel/#/customers">کاربران</a>}{product && <a className={view === "catalog" ? "active" : ""} href="/panel/#/catalog">پلن‌ها</a>}{runtime && <a className={view.startsWith("runtime") ? "active" : ""} href="/panel/#/runtime/naive">سیستم</a>}<button onClick={signOut}>خروج</button></nav>;
}

function Shell({ principal, view, signOut, children }: { principal: Principal; view: View; signOut: () => Promise<void>; children: React.ReactNode }) {
  return <div className="shell"><Sidebar principal={principal} view={view} signOut={signOut}/><div className="content-shell">{children}</div><MobileNav principal={principal} view={view} signOut={signOut}/></div>;
}

export function App() {
  const [authState, setAuthState] = useState<AuthState>("loading"); const [principal, setPrincipal] = useState<Principal | null>(null); const [authMessage, setAuthMessage] = useState(""); const [view, setView] = useState<View>(currentView);
  useEffect(() => { const onHash = () => setView(currentView()); window.addEventListener("hashchange", onHash); return () => window.removeEventListener("hashchange", onHash); }, []);
  useEffect(() => { let active = true; me().then((current) => { if (active) { setPrincipal(current); setAuthState("authenticated"); } }).catch((cause: AuthError) => { if (active) { setPrincipal(null); setAuthState("anonymous"); if (cause.status && cause.status !== 401) setAuthMessage("سرویس ورود در دسترس نیست."); } }); return () => { active = false; }; }, []);
  async function signOut() { const csrf = readCookie("__Host-pvnaive_csrf"); if (!csrf) { setPrincipal(null); setAuthState("anonymous"); return; } try { await logout(csrf); } finally { setPrincipal(null); setAuthState("anonymous"); } }
  if (authState === "loading") return <main className="auth-page"><section className="auth-card"><p>در حال بررسی نشست…</p></section></main>;
  if (authState === "anonymous" || !principal) return <><LoginScreen onAuthenticated={(current) => { setPrincipal(current); setAuthState("authenticated"); setAuthMessage(""); }}/>{authMessage && <div className="global-auth-message">{authMessage}</div>}</>;
  const product = canUseCustomerProduct(principal.role); const runtime = canUseRawRuntime(principal.role);
  let content: React.ReactNode = <Dashboard role={principal.role}/>;
  if (view === "customers" && product) content = <ProductCustomers role={principal.role}/>;
  else if (view === "catalog" && product) content = <ProductCatalog role={principal.role}/>;
  else if (view === "runtime-adoption" && runtime) content = <RuntimeAdoption/>;
  else if (view === "runtime-naive" && runtime) content = <RuntimeNaive/>;
  return <Shell principal={principal} view={view} signOut={signOut}>{content}</Shell>;
}
