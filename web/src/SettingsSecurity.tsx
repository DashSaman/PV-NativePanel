import { FormEvent, useState } from "react";
import type { Principal } from "./auth";
import { readCookie } from "./auth";

type Props = { principal: Principal };

function mutationHeaders(): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf");
  if (!csrf) throw new Error("توکن CSRF در دسترس نیست.");
  return { "Content-Type": "application/json", "X-CSRF-Token": csrf };
}

async function callAPI(path: string, method: string, body: unknown): Promise<Record<string, unknown>> {
  const response = await fetch(path, {
    method,
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: JSON.stringify(body),
  });
  const contentType = response.headers.get("Content-Type") || "";
  const payload = contentType.includes("application/json") ? await response.json() as Record<string, unknown> : {};
  if (!response.ok) {
    const message = typeof payload.message === "string" ? payload.message : "درخواست انجام نشد.";
    throw new Error(message);
  }
  return payload;
}

export function SettingsSecurity({ principal }: Props) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordBusy, setPasswordBusy] = useState(false);
  const [passwordMessage, setPasswordMessage] = useState("");

  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [profileBusy, setProfileBusy] = useState(false);
  const [profileMessage, setProfileMessage] = useState("");

  async function submitPassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPasswordMessage("");
    if (newPassword !== confirmPassword) {
      setPasswordMessage("تکرار رمز جدید با رمز جدید یکسان نیست.");
      return;
    }
    if (newPassword.length < 14) {
      setPasswordMessage("رمز جدید باید حداقل ۱۴ نویسه باشد.");
      return;
    }
    setPasswordBusy(true);
    try {
      await callAPI("/api/v1/me/password", "POST", { current_password: currentPassword, new_password: newPassword });
      setPasswordMessage("رمز عبور تغییر کرد. سایر نشست‌ها بسته شد.");
      setCurrentPassword(""); setNewPassword(""); setConfirmPassword("");
    } catch (cause) {
      setPasswordMessage(cause instanceof Error ? cause.message : "تغییر رمز انجام نشد.");
    } finally {
      setPasswordBusy(false);
    }
  }

  async function submitProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setProfileMessage("");
    if (!email && !displayName) {
      setProfileMessage("برای به‌روزرسانی، ایمیل یا نام نمایشی را وارد کنید.");
      return;
    }
    setProfileBusy(true);
    try {
      const payload = await callAPI("/api/v1/me/profile", "PATCH", { email, display_name: displayName });
      setProfileMessage(`به‌روزرسانی شد: ${typeof payload.email === "string" ? payload.email : principal.email}`);
      setEmail(""); setDisplayName("");
    } catch (cause) {
      setProfileMessage(cause instanceof Error ? cause.message : "به‌روزرسانی انجام نشد.");
    } finally {
      setProfileBusy(false);
    }
  }

  return <main className="dashboard-page">
    <header className="dashboard-hero"><div><p className="eyebrow">Account</p><h1>امنیت و حساب</h1><p>مدیریت رمز عبور و مشخصات ورود حساب شما.</p></div></header>
    <div className="dashboard-grid">
      <section className="dashboard-card" aria-labelledby="pw-title">
        <div className="dashboard-card-head"><h2 id="pw-title">تغییر رمز عبور</h2></div>
        <form className="auth-form" onSubmit={submitPassword}>
          <label>رمز عبور فعلی<input type="password" autoComplete="current-password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} required disabled={passwordBusy}/></label>
          <label>رمز عبور جدید (حداقل ۱۴ نویسه)<input type="password" autoComplete="new-password" minLength={14} maxLength={1024} value={newPassword} onChange={(e) => setNewPassword(e.target.value)} required disabled={passwordBusy}/></label>
          <label>تکرار رمز عبور جدید<input type="password" autoComplete="new-password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} required disabled={passwordBusy}/></label>
          {passwordMessage && <p className="auth-message">{passwordMessage}</p>}
          <button type="submit" disabled={passwordBusy}>{passwordBusy ? "در حال ثبت…" : "تغییر رمز عبور"}</button>
          <small className="auth-copy">با تغییر رمز، همه نشست‌های دیگر روی همه دستگاه‌ها بسته می‌شوند و فقط همین نشست باز می‌ماند.</small>
        </form>
      </section>
      <section className="dashboard-card" aria-labelledby="id-title">
        <div className="dashboard-card-head"><h2 id="id-title">مشخصات ورود</h2></div>
        <p className="auth-copy">حساب فعلی: <strong>{principal.email}</strong>{principal.display_name ? ` — ${principal.display_name}` : ""}</p>
        <form className="auth-form" onSubmit={submitProfile}>
          <label>ایمیل جدید (خالی = بدون تغییر)<input type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} disabled={profileBusy}/></label>
          <label>نام نمایشی جدید (خالی = بدون تغییر)<input type="text" maxLength={160} value={displayName} onChange={(e) => setDisplayName(e.target.value)} disabled={profileBusy}/></label>
          {profileMessage && <p className="auth-message">{profileMessage}</p>}
          <button type="submit" disabled={profileBusy}>{profileBusy ? "در حال ثبت…" : "به‌روزرسانی مشخصات"}</button>
          <small className="auth-copy">از این پس با ایمیل/مشخصات جدید وارد پنل می‌شوید.</small>
        </form>
      </section>
    </div>
  </main>;
}
