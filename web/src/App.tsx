import { FormEvent, useEffect, useState } from "react";
import { AuthError, login, logout, me, Principal, readCookie } from "./auth";
import { appRoutes, assertRouteManifest } from "./routes";

assertRouteManifest();

type Theme = "system" | "dark" | "light";
type AuthState = "loading" | "anonymous" | "authenticated";
const metrics = [["وضعیت Runtime","در انتظار اتصال"],["کاربران فعال","—"],["مصرف امروز","—"],["اعتبار TLS","—"]];
const enabledRoutes = appRoutes.filter((route) => route.navigation && route.path === "/");

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
    event.preventDefault();
    setSubmitting(true);
    setMessage("");
    try {
      await login({ email, password, totpCode: requiresMFA ? totpCode : undefined });
      setPassword("");
      setTotpCode("");
      const principal = await me();
      onAuthenticated(principal);
    } catch (cause) {
      const error = cause as AuthError;
      if (error.code === "mfa_required") {
        setRequiresMFA(true);
        setMessage("کد شش‌رقمی برنامه احراز هویت را وارد کنید.");
      } else {
        setMessage("ورود انجام نشد. اطلاعات ورود را بررسی کنید.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="auth-page">
      <section className="auth-card" aria-labelledby="login-title">
        <div className="brand auth-brand"><img src="/pvnaive-mark.svg" alt="" width="48" height="48"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div>
        <p className="eyebrow">Secure control plane</p>
        <h1 id="login-title">ورود به پنل</h1>
        <p className="auth-copy">ورود مدیریت با نشست امن، CSRF و احراز هویت دومرحله‌ای.</p>
        <form className="auth-form" onSubmit={submit}>
          <label>ایمیل<input type="email" autoComplete="username" value={email} onChange={(event) => setEmail(event.target.value)} required disabled={submitting}/></label>
          <label>رمز عبور<input type="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} required disabled={submitting}/></label>
          {requiresMFA && <label>کد TOTP<input inputMode="numeric" pattern="[0-9]{6}" maxLength={6} autoComplete="one-time-code" value={totpCode} onChange={(event) => setTotpCode(event.target.value.replace(/\D/g, ""))} required disabled={submitting}/></label>}
          {message && <p className="auth-message" role="status">{message}</p>}
          <button type="submit" disabled={submitting || (requiresMFA && totpCode.length !== 6)}>{submitting ? "در حال بررسی…" : requiresMFA ? "تأیید و ورود" : "ورود"}</button>
        </form>
      </section>
    </main>
  );
}

export function App() {
  const [authState, setAuthState] = useState<AuthState>("loading");
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [authMessage, setAuthMessage] = useState("");

  useEffect(() => {
    let active = true;
    me().then((current) => {
      if (!active) return;
      setPrincipal(current);
      setAuthState("authenticated");
    }).catch((cause: AuthError) => {
      if (!active) return;
      setPrincipal(null);
      setAuthState("anonymous");
      if (cause.status && cause.status !== 401) setAuthMessage("سرویس احراز هویت در دسترس نیست.");
    });
    return () => { active = false; };
  }, []);

  async function signOut() {
    const csrf = readCookie("__Host-pvnaive_csrf");
    if (!csrf) {
      setPrincipal(null);
      setAuthState("anonymous");
      return;
    }
    try {
      await logout(csrf);
    } finally {
      setPrincipal(null);
      setAuthState("anonymous");
    }
  }

  if (authState === "loading") {
    return <main className="auth-page"><section className="auth-card"><p>در حال بررسی نشست امن…</p></section></main>;
  }
  if (authState === "anonymous" || !principal) {
    return <><LoginScreen onAuthenticated={(current) => { setPrincipal(current); setAuthState("authenticated"); setAuthMessage(""); }}/>{authMessage && <div className="global-auth-message" role="alert">{authMessage}</div>}</>;
  }

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand"><img src="/pvnaive-mark.svg" alt="" width="44" height="44"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div>
        <nav aria-label="ناوبری اصلی">{enabledRoutes.map((route) => <a href={route.path} key={route.path}>{route.label}</a>)}</nav>
        <ThemeSwitch />
        <button className="logout-button" onClick={signOut}>خروج امن</button>
        <div className="stage">S04 Auth · protected preview</div>
      </aside>
      <main>
        <header><div><p className="eyebrow">Standalone control plane</p><h1>داشبورد PVNaive</h1></div><span className="badge">ورود امن فعال</span></header>
        <section className="notice" role="status">{principal.display_name} · {principal.email} · نقش: {principal.role}</section>
        <section className="metrics" aria-label="شاخص‌ها">{metrics.map(([label,value]) => <article key={label}><span>{label}</span><strong>{value}</strong></article>)}</section>
        <section className="panel"><div><p className="eyebrow">S04 Authentication</p><h2>هویت و نشست مدیریت فعال است</h2><p>قابلیت‌های کاربران، فروش و Runtime تا عبور مرحله‌های بعدی عمداً غیرفعال می‌مانند.</p></div><button disabled>ایجاد کاربر</button></section>
      </main>
      <nav className="mobile-nav" aria-label="ناوبری موبایل">
        <a href="/">داشبورد</a>
        <button onClick={signOut}>خروج</button>
      </nav>
    </div>
  );
}
