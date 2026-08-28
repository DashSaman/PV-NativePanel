import { FormEvent, useEffect, useState } from "react";
import { AuthError, login, logout, me, Principal, readCookie } from "./auth";
import { RuntimeNaive } from "./RuntimeNaive";
import { assertRouteManifest } from "./routes";

assertRouteManifest();

type Theme = "system" | "dark" | "light";
type AuthState = "loading" | "anonymous" | "authenticated";
type View = "dashboard" | "runtime-naive";
const metrics = [["وضعیت Runtime","قابل مدیریت"],["اکانت‌های Naive","از Runtime"],["مصرف امروز","—"],["اعتبار TLS","—"]];

function currentView(): View {
  return window.location.hash === "#/runtime/naive" ? "runtime-naive" : "dashboard";
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

function Sidebar({ principal, signOut }: { principal: Principal; signOut: () => Promise<void> }) {
  return <aside className="sidebar">
    <div className="brand"><img src="/pvnaive-mark.svg" alt="" width="44" height="44"/><div><strong>PVNaive</strong><span>PVNETWORK</span></div></div>
    <nav aria-label="ناوبری اصلی">
      <a href="/panel/">داشبورد</a>
      {principal.role === "owner" && <a href="/panel/#/runtime/naive">Naive Runtime</a>}
    </nav>
    <ThemeSwitch />
    <button className="logout-button" onClick={signOut}>خروج امن</button>
    <div className="stage">S04R · Naive credential manager</div>
  </aside>;
}

export function App() {
  const [authState, setAuthState] = useState<AuthState>("loading");
  const [principal, setPrincipal] = useState<Principal | null>(null);
  const [authMessage, setAuthMessage] = useState("");
  const [view, setView] = useState<View>(currentView);

  useEffect(() => {
    const onHash = () => setView(currentView());
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, []);

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

  if (view === "runtime-naive" && principal.role === "owner") {
    return <div className="shell">
      <Sidebar principal={principal} signOut={signOut} />
      <RuntimeNaive />
      <nav className="mobile-nav" aria-label="ناوبری موبایل"><a href="/panel/">داشبورد</a><button onClick={signOut}>خروج</button></nav>
    </div>;
  }

  return (
    <div className="shell">
      <Sidebar principal={principal} signOut={signOut} />
      <main>
        <header><div><p className="eyebrow">Standalone control plane</p><h1>داشبورد PVNaive</h1></div><span className="badge">ورود امن فعال</span></header>
        <section className="notice" role="status">{principal.display_name} · {principal.email} · نقش: {principal.role}</section>
        <section className="metrics" aria-label="شاخص‌ها">{metrics.map(([label,value]) => <article key={label}><span>{label}</span><strong>{value}</strong></article>)}</section>
        <section className="panel"><div><p className="eyebrow">Naive Runtime</p><h2>مدیریت اکانت‌های Naive آماده است</h2><p>{principal.role === "owner" ? "ابتدا اکانت فعلی سرور به‌صورت امن Import می‌شود؛ سپس ساخت، تغییر رمز، فعال/غیرفعال و لغو اکانت‌ها در دسترس است." : "مدیریت Runtime فقط برای Owner در دسترس است."}</p></div>{principal.role === "owner" ? <a className="button-secondary" href="/panel/#/runtime/naive">مدیریت Naive</a> : <button disabled>فقط Owner</button>}</section>
      </main>
      <nav className="mobile-nav" aria-label="ناوبری موبایل">
        <a href="/panel/">داشبورد</a>
        {principal.role === "owner" ? <a href="/panel/#/runtime/naive">Runtime</a> : <button onClick={signOut}>خروج</button>}
      </nav>
    </div>
  );
}
