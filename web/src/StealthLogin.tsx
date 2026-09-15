import { FormEvent, useEffect, useRef, useState } from "react";
import { AuthError, login, me, Principal } from "./auth";
import { Icon } from "./ui";
import "./stealth.css";

// Login view — "Gold Reception" redesign (UI-002).
//
// Product decision (owner feedback): the previous stealth behaviour hid the
// email/password fields until hover and users could not find them. The login
// is now a fully visible, graphically rich split card: brand hero on one
// side, always-visible labeled inputs on the other. Stealth remains a UX
// option elsewhere, NOT a security control — the real controls are the login
// rate limit / fail2ban (ACCESS-003) and RBAC.

type Props = { onAuthenticated: (principal: Principal) => void };

function prefersReducedMotion(): boolean {
  return typeof window !== "undefined" &&
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

// AuroraMesh paints a slow, GPU-cheap gradient field in the brand palette
// (gold/amber). With reduced motion it renders exactly one static frame.
function AuroraMesh() {
  const ref = useRef<HTMLCanvasElement | null>(null);
  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    let raf = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const resize = () => {
      canvas.width = canvas.clientWidth * dpr;
      canvas.height = canvas.clientHeight * dpr;
    };
    resize();
    const draw = (t: number) => {
      const w = canvas.width, h = canvas.height;
      ctx.clearRect(0, 0, w, h);
      const blobs: Array<[number, number, number, string]> = [
        [0.28 + 0.1 * Math.sin(t / 9000), 0.32 + 0.08 * Math.cos(t / 11000), 0.55, "rgba(245,182,46,0.10)"],
        [0.72 + 0.08 * Math.cos(t / 8000), 0.28 + 0.1 * Math.sin(t / 12000), 0.5, "rgba(232,153,12,0.09)"],
        [0.5 + 0.12 * Math.sin(t / 13000), 0.78 + 0.06 * Math.cos(t / 9000), 0.6, "rgba(255,215,106,0.05)"],
      ];
      for (const [cx, cy, rr, color] of blobs) {
        const g = ctx.createRadialGradient(cx * w, cy * h, 0, cx * w, cy * h, rr * Math.max(w, h));
        g.addColorStop(0, color);
        g.addColorStop(1, "rgba(0,0,0,0)");
        ctx.fillStyle = g;
        ctx.fillRect(0, 0, w, h);
      }
    };
    if (prefersReducedMotion()) {
      draw(0);
      return () => { cancelAnimationFrame(raf); };
    }
    const loop = (t: number) => { draw(t); raf = requestAnimationFrame(loop); };
    raf = requestAnimationFrame(loop);
    window.addEventListener("resize", resize);
    return () => { cancelAnimationFrame(raf); window.removeEventListener("resize", resize); };
  }, []);
  return <canvas ref={ref} className="stealth-canvas" aria-hidden="true"/>;
}

type Props2 = { label: string; icon: string; children: React.ReactNode; trailing?: React.ReactNode };

function LoginField({ label, icon, children, trailing }: Props2) {
  return (
    <label className="login-field">
      <span className="login-field-label">{label}</span>
      <span className="login-input-wrap">
        <span className="login-input-icon" aria-hidden="true"><Icon name={icon} size={16}/></span>
        {children}
        {trailing}
      </span>
    </label>
  );
}

export function StealthLogin({ onAuthenticated }: Props) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
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
      onAuthenticated(await me());
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
    <main className="auth-page stealth-page">
      <AuroraMesh/>
      <section className="login-card" aria-labelledby="login-title">
        <aside className="login-hero" aria-hidden="true">
          <div className="login-hero-glow"/>
          <img className="login-hero-mark" src={`${import.meta.env.BASE_URL}private-network.webp`} alt="" width="84" height="84"/>
          <strong className="login-hero-title">PVNaive</strong>
          <span className="login-hero-sub">PVNETWORK · Private Network</span>
          <ul className="login-hero-features">
            <li><span className="login-hero-dot"><Icon name="users" size={13}/></span>مدیریت کاربران، حجم و اعتبار</li>
            <li><span className="login-hero-dot"><Icon name="gauge" size={13}/></span>مانیتورینگ زنده سرور</li>
            <li><span className="login-hero-dot"><Icon name="qr" size={13}/></span>اشتراک و QR دوگانه</li>
            <li><span className="login-hero-dot"><Icon name="shield" size={13}/></span>نشست امن با رمزنگاری کامل</li>
          </ul>
          <span className="login-hero-foot">NaiveProxy Control Panel</span>
        </aside>

        <div className="login-body">
          <div className="brand auth-brand">
            <img className="brand-mark" src={`${import.meta.env.BASE_URL}private-network.webp`} alt="Private Network" width="44" height="44"/>
            <div><strong>PVNaive</strong><span>PVNETWORK</span></div>
          </div>
          <h1 id="login-title">ورود به پنل</h1>
          <p className="auth-copy">مدیریت امن سرویس و کاربران</p>
          <form className="login-form" onSubmit={submit} autoComplete="on">
            <LoginField label="ایمیل" icon="mail">
              <input type="email" aria-label="ایمیل" placeholder="admin@example.com" autoComplete="username" value={email}
                onChange={(e) => setEmail(e.target.value)} required disabled={submitting}/>
            </LoginField>
            <LoginField label="رمز عبور" icon="lock"
              trailing={
                <button type="button" className="login-eye" aria-label={showPassword ? "پنهان کردن رمز" : "نمایش رمز"}
                  aria-pressed={showPassword} onClick={() => setShowPassword((v) => !v)} tabIndex={-1}>
                  <Icon name={showPassword ? "eye-off" : "eye"} size={15}/>
                </button>
              }>
              <input type={showPassword ? "text" : "password"} aria-label="رمز عبور" placeholder="••••••••••••"
                autoComplete="current-password" value={password}
                onChange={(e) => setPassword(e.target.value)} required disabled={submitting}/>
            </LoginField>
            {requiresMFA && (
              <LoginField label="کد TOTP" icon="key">
                <input inputMode="numeric" pattern="[0-9]{6}" maxLength={6} aria-label="کد TOTP" placeholder="000000"
                  autoComplete="one-time-code" value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ""))} required disabled={submitting}/>
              </LoginField>
            )}
            <div className="login-message-slot" style={{ minHeight: 30 }} role="status">
              {message && <p className="auth-message">{message}</p>}
            </div>
            <button type="submit" className="login-submit" disabled={submitting || (requiresMFA && totpCode.length !== 6)}>
              <Icon name="lock" size={16}/>{submitting ? "در حال بررسی…" : "ورود امن"}
            </button>
          </form>
          <p className="login-footnote"><Icon name="shield" size={12}/> اتصال رمزنگاری‌شده · نشست با کوکی امن</p>
        </div>
      </section>
    </main>
  );
}
