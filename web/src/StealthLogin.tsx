import { FormEvent, useEffect, useReducer, useRef, useState } from "react";
import { AuthError, login, me, Principal } from "./auth";
import { Icon } from "./ui";
import {
  MESSAGE_SLOT_MIN_PX,
  StealthField,
  initialStealthState,
  isRevealed,
  reduceStealth,
} from "./stealthModel";
import "./stealth.css";

// StealthLogin implements the R8 animated stealth login (gate UI-001).
//
// IMPORTANT: stealth is a UX layer, NOT a security control. The real controls
// are the login rate limit / fail2ban (ACCESS-003) and RBAC. Field positions
// are merely not highlighted to a casual onlooker; keyboard users and password
// managers keep full access (labels are announced to assistive technology).

type Props = { onAuthenticated: (principal: Principal) => void };

function prefersReducedMotion(): boolean {
  return typeof window !== "undefined" &&
    typeof window.matchMedia === "function" &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

// AuroraMesh paints a slow, GPU-cheap gradient field. With reduced motion it
// renders exactly one static frame and never schedules another animation.
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
        [0.3 + 0.1 * Math.sin(t / 9000), 0.35 + 0.08 * Math.cos(t / 11000), 0.55, "rgba(245,185,66,0.10)"],
        [0.7 + 0.08 * Math.cos(t / 8000), 0.3 + 0.1 * Math.sin(t / 12000), 0.5, "rgba(96,165,250,0.08)"],
        [0.5 + 0.12 * Math.sin(t / 13000), 0.75 + 0.06 * Math.cos(t / 9000), 0.6, "rgba(74,222,128,0.05)"],
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

type FieldProps = {
  field: StealthField;
  label: string;
  revealed: boolean;
  hovered: boolean;
  children: React.ReactNode;
  onHover: (field: StealthField, enter: boolean) => void;
  onFocus: (field: StealthField, focused: boolean) => void;
};

function StealthFieldWrap({ field, label, revealed, hovered, children, onHover, onFocus }: FieldProps) {
  const cls = "stealth-field" + (revealed ? " revealed" : "") + (hovered && !revealed ? " hinted" : "");
  return (
    <div
      className={cls}
      data-field={field}
      onMouseEnter={() => onHover(field, true)}
      onMouseLeave={() => onHover(field, false)}
      onFocusCapture={() => onFocus(field, true)}
      onBlurCapture={() => onFocus(field, false)}
    >
      <span className="stealth-label" aria-hidden="true">{label}</span>
      <span className="stealth-sr-label">{label}</span>
      {children}
    </div>
  );
}

export function StealthLogin({ onAuthenticated }: Props) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [requiresMFA, setRequiresMFA] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");
  const [state, dispatch] = useReducer(reduceStealth, initialStealthState);

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

  const hover = (field: StealthField, enter: boolean) =>
    dispatch({ type: enter ? "hover-enter" : "hover-leave", field });
  const focus = (field: StealthField, focused: boolean) =>
    dispatch({ type: focused ? "focus" : "blur", field });

  return (
    <main className="auth-page stealth-page">
      <AuroraMesh/>
      <section className="auth-card stealth-card" aria-labelledby="login-title">
        <div className="brand auth-brand">
          <img src={`${import.meta.env.BASE_URL}pvnaive-mark.svg`} alt="" width="48" height="48"/>
          <div><strong>PVNaive</strong><span>PVNETWORK</span></div>
        </div>
        <p className="eyebrow"><Icon name="activity" size={12}/> Secure Control Panel</p>
        <h1 id="login-title">ورود به پنل</h1>
        <p className="auth-copy">مدیریت امن سرویس و کاربران</p>
        <form className="auth-form stealth-form" onSubmit={submit} autoComplete="on">
          <StealthFieldWrap field="email" label="ایمیل" revealed={isRevealed(state, "email")}
            hovered={state.hovered === "email"} onHover={hover} onFocus={focus}>
            <input type="email" aria-label="ایمیل" autoComplete="username" value={email}
              onChange={(e) => setEmail(e.target.value)} required disabled={submitting}/>
          </StealthFieldWrap>
          <StealthFieldWrap field="password" label="رمز عبور" revealed={isRevealed(state, "password")}
            hovered={state.hovered === "password"} onHover={hover} onFocus={focus}>
            <input type="password" aria-label="رمز عبور" autoComplete="current-password" value={password}
              onChange={(e) => setPassword(e.target.value)} required disabled={submitting}/>
          </StealthFieldWrap>
          {requiresMFA && (
            <StealthFieldWrap field="totp" label="کد TOTP" revealed={isRevealed(state, "totp")}
              hovered={state.hovered === "totp"} onHover={hover} onFocus={focus}>
              <input inputMode="numeric" pattern="[0-9]{6}" maxLength={6} aria-label="کد TOTP"
                autoComplete="one-time-code" value={totpCode}
                onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ""))} required disabled={submitting}/>
            </StealthFieldWrap>
          )}
          <div className="stealth-message-slot" style={{ minHeight: MESSAGE_SLOT_MIN_PX }} role="status">
            {message && <p className="auth-message">{message}</p>}
          </div>
          <button type="submit" disabled={submitting || (requiresMFA && totpCode.length !== 6)}>
            <Icon name="lock" size={16}/>{submitting ? "در حال بررسی…" : "ورود امن"}
          </button>
        </form>
        <p className="stealth-note">حرکت نشانگر یا Tab، فیلدها را نشان می‌دهد.</p>
      </section>
    </main>
  );
}
