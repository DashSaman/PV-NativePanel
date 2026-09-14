import { useEffect, useState } from "react";
import { AuthError, logout, me, Principal, readCookie } from "./auth";
import { Dashboard } from "./Dashboard";
import { PoolManager } from "./PoolManager";
import { ErrorBoundary } from "./ErrorBoundary";
import { ProductCatalog } from "./ProductCatalog";
import { ProductCustomers } from "./ProductCustomers";
import { RuntimeAdoption } from "./RuntimeAdoption";
import { RuntimeNaive } from "./RuntimeNaive";
import { SettingsSecurity } from "./SettingsSecurity";
import { StealthLogin } from "./StealthLogin";
import { canUseCustomerProduct, canUseRawRuntime } from "./productPanelModel";
import { assertRouteManifest } from "./routes";
import { Icon } from "./ui";

assertRouteManifest();

type Theme = "system" | "dark" | "light";
type AuthState = "loading" | "anonymous" | "authenticated";
type View = "dashboard" | "customers" | "catalog" | "runtime-adoption" | "runtime-naive" | "settings-security" | "pool";

export function currentView(hash = window.location.hash): View {
  if (hash === "#/customers") return "customers";
  if (hash === "#/catalog" || hash === "#/plans") return "catalog";
  if (hash === "#/customers/runtime-adoption") return "runtime-adoption";
  if (hash === "#/runtime/naive") return "runtime-naive";
  if (hash === "#/pool") return "pool";
  if (hash === "#/settings/security") return "settings-security";
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
  const glyph = theme === "dark" ? "moon" : theme === "light" ? "sun" : "monitor";
  return <button className="theme-switch" onClick={() => setTheme(next)} aria-label="تغییر پوسته"><Icon name={glyph} size={16}/><span>{theme === "dark" ? "تیره" : theme === "light" ? "روشن" : "سیستم"}</span></button>;
}

function LoginScreen({ onAuthenticated }: { onAuthenticated: (principal: Principal) => void }) {
  // R8 stealth login: animated, hover/focus-revealed fields. The previous
  // static form logic moved into StealthLogin (same auth contract).
  return <StealthLogin onAuthenticated={onAuthenticated}/>;
}

function Sidebar({ principal, view, signOut }: { principal: Principal; view: View; signOut: () => Promise<void> }) {
  const product = canUseCustomerProduct(principal.role); const runtime = canUseRawRuntime(principal.role);
  const linkClass = (active: boolean) => active ? "nav-link active" : "nav-link";
  return <aside className="sidebar"><div className="brand"><img className="brand-mark" src={`${import.meta.env.BASE_URL}private-network.webp`} alt="Private Network" width="44" height="44"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div><nav aria-label="ناوبری اصلی"><a className={linkClass(view === "dashboard")} href="/panel/"><b><Icon name="dashboard" size={17}/></b><span>داشبورد</span></a>{product && <><a className={linkClass(view === "customers")} href="/panel/#/customers"><b><Icon name="users" size={17}/></b><span>کاربران</span></a><a className={linkClass(view === "catalog")} href="/panel/#/catalog"><b><Icon name="plans" size={17}/></b><span>پلن‌ها و دسته‌بندی</span></a></>}{runtime && <a className={linkClass(view === "runtime-naive" || view === "runtime-adoption")} href="/panel/#/runtime/naive"><b><Icon name="system" size={17}/></b><span>سیستم / Runtime</span></a>}{principal.role === "owner" && <a className={linkClass(view === "pool")} href="/panel/#/pool"><b><Icon name="gauge" size={17}/></b><span>استخر گره‌ها</span></a>}{principal.role === "owner" && <a className={linkClass(view === "settings-security")} href="/panel/#/settings/security"><b><Icon name="shield" size={17}/></b><span>امنیت و حساب</span></a>}</nav><div className="sidebar-footer"><ThemeSwitch/><button className="logout-button" onClick={signOut}><Icon name="logout" size={16}/><span>خروج امن</span></button><small>{principal.display_name || principal.email}</small></div></aside>;
}

function MobileNav({ principal, view, signOut }: { principal: Principal; view: View; signOut: () => Promise<void> }) {
  const product = canUseCustomerProduct(principal.role); const runtime = canUseRawRuntime(principal.role);
  return <nav className="mobile-nav"><a className={view === "dashboard" ? "active" : ""} href="/panel/">داشبورد</a>{product && <a className={view === "customers" ? "active" : ""} href="/panel/#/customers">کاربران</a>}{product && <a className={view === "catalog" ? "active" : ""} href="/panel/#/catalog">پلن‌ها</a>}{runtime && <a className={view.startsWith("runtime") ? "active" : ""} href="/panel/#/runtime/naive">سیستم</a>}{principal.role === "owner" && <a className={view === "pool" ? "active" : ""} href="/panel/#/pool">استخر</a>}<button onClick={signOut}>خروج</button></nav>;
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
  else if (view === "pool" && principal.role === "owner") content = <PoolManager principal={principal}/>;
  else if (view === "settings-security") content = <SettingsSecurity principal={principal}/>;
  return <ErrorBoundary><Shell principal={principal} view={view} signOut={signOut}>{content}</Shell></ErrorBoundary>;
}
