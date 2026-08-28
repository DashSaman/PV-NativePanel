import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  buildNaiveURI,
  createRuntimeCredential,
  getRuntimeStatus,
  importCurrentRuntime,
  listRuntimeCredentials,
  revokeRuntimeCredential,
  rotateRuntimeCredential,
  RuntimeCredential,
  RuntimeError,
  RuntimeStatus,
  updateRuntimeCredential,
} from "./runtime";
import "./runtime.css";

type SecretView = { username: string; password: string } | null;

function friendlyError(cause: unknown): string {
  const error = cause as RuntimeError;
  switch (error.code) {
    case "last_active_credential":
      return "آخرین اکانت فعال را نمی‌توان غیرفعال یا حذف کرد.";
    case "revision_conflict":
      return "اطلاعات اکانت تغییر کرده؛ صفحه را تازه‌سازی و دوباره امتحان کنید.";
    case "username_conflict":
      return "این نام کاربری قبلاً وجود دارد.";
    case "runtime_import_equivalence_failed":
      return "Import متوقف شد چون بازسازی Caddy دقیقاً با فایل فعلی یکسان نبود. هیچ تغییری روی Runtime اعمال نشد.";
    case "runtime_already_owned":
      return "اکانت فعلی قبلاً وارد پنل شده است.";
    case "csrf_failed":
      return "نشست امنیتی معتبر نیست؛ یک بار از پنل خارج و دوباره وارد شوید.";
    default:
      return error.message || "عملیات Runtime انجام نشد.";
  }
}

function shortSHA(value?: string): string {
  if (!value) return "—";
  return `${value.slice(0, 12)}…${value.slice(-8)}`;
}

function formatTime(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("fa-IR", { dateStyle: "short", timeStyle: "short" }).format(date);
}

export function RuntimeNaive() {
  const [status, setStatus] = useState<RuntimeStatus | null>(null);
  const [credentials, setCredentials] = useState<RuntimeCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [error, setError] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [generatePassword, setGeneratePassword] = useState(true);
  const [secret, setSecret] = useState<SecretView>(null);

  const activeCount = useMemo(() => credentials.filter((item) => item.status === "active").length, [credentials]);
  const owned = credentials.length > 0;
  const customerURI = secret ? buildNaiveURI(secret.username, secret.password, window.location.hostname) : "";

  const reload = useCallback(async () => {
    const [nextStatus, nextCredentials] = await Promise.all([getRuntimeStatus(), listRuntimeCredentials()]);
    setStatus(nextStatus);
    setCredentials(nextCredentials);
  }, []);

  useEffect(() => {
    reload()
      .catch((cause) => setError(friendlyError(cause)))
      .finally(() => setLoading(false));
  }, [reload]);

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await action();
    } catch (cause) {
      setError(friendlyError(cause));
    } finally {
      setBusy(false);
    }
  }

  async function importLive() {
    await run(async () => {
      const imported = await importCurrentRuntime();
      setCredentials(imported);
      await reload();
      setNotice("اکانت فعلی بدون تغییر Caddy و بدون نمایش رمز در مرورگر، امن وارد پنل شد.");
    });
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await run(async () => {
      const result = await createRuntimeCredential(username.trim(), generatePassword ? "" : password, generatePassword);
      setUsername("");
      setPassword("");
      await reload();
      if (result.generated_password) setSecret({ username: result.credential.username, password: result.generated_password });
      setNotice("اکانت جدید با موفقیت روی Runtime اعمال شد.");
    });
  }

  async function rename(credential: RuntimeCredential) {
    const next = window.prompt("نام کاربری جدید را وارد کنید:", credential.username);
    if (!next || next === credential.username) return;
    await run(async () => {
      await updateRuntimeCredential(credential, next.trim(), credential.status);
      await reload();
      setNotice("نام کاربری با موفقیت تغییر کرد.");
    });
  }

  async function toggle(credential: RuntimeCredential) {
    if (credential.status === "active" && activeCount <= 1) {
      setError("آخرین اکانت فعال را نمی‌توان غیرفعال کرد.");
      return;
    }
    const nextStatus = credential.status === "active" ? "disabled" : "active";
    const verb = nextStatus === "active" ? "فعال" : "غیرفعال";
    if (!window.confirm(`اکانت ${credential.username} ${verb} شود؟`)) return;
    await run(async () => {
      await updateRuntimeCredential(credential, credential.username, nextStatus);
      await reload();
      setNotice(`اکانت ${verb} شد.`);
    });
  }

  async function rotateGenerated(credential: RuntimeCredential) {
    if (!window.confirm(`برای ${credential.username} رمز جدید امن تولید و روی Runtime اعمال شود؟`)) return;
    await run(async () => {
      const result = await rotateRuntimeCredential(credential, "", true);
      await reload();
      if (result.generated_password) setSecret({ username: credential.username, password: result.generated_password });
      setNotice("رمز جدید اعمال شد. رمز قبلی دیگر معتبر نیست.");
    });
  }

  async function rotateCustom(credential: RuntimeCredential) {
    const next = window.prompt("رمز جدید را وارد کنید (حداقل ۱۴ کاراکتر ASCII قابل‌نمایش):", "");
    if (!next) return;
    await run(async () => {
      await rotateRuntimeCredential(credential, next, false);
      await reload();
      setNotice("رمز دلخواه با موفقیت اعمال شد.");
    });
  }

  async function revoke(credential: RuntimeCredential) {
    if (credential.status === "active" && activeCount <= 1) {
      setError("آخرین اکانت فعال را نمی‌توان حذف کرد.");
      return;
    }
    if (!window.confirm(`اکانت ${credential.username} لغو شود؟ این کار در پنل به‌صورت soft revoke انجام می‌شود.`)) return;
    await run(async () => {
      await revokeRuntimeCredential(credential);
      await reload();
      setNotice("اکانت لغو شد.");
    });
  }

  if (loading) return <section className="runtime-card"><p>در حال خواندن وضعیت Naive Runtime…</p></section>;

  return (
    <div className="runtime-page" dir="rtl">
      <div className="runtime-heading">
        <div>
          <p className="eyebrow">NaiveProxy Runtime</p>
          <h1>مدیریت اکانت‌های Naive</h1>
          <p className="runtime-muted">تغییرات با validate، backup، reload-only و rollback محافظت می‌شوند.</p>
        </div>
        <a className="button-secondary" href="/panel/">بازگشت به داشبورد</a>
      </div>

      {error && <div className="runtime-alert error" role="alert">{error}</div>}
      {notice && <div className="runtime-alert success" role="status">{notice}</div>}

      <section className="runtime-stats" aria-label="وضعیت Runtime">
        <article><span>Runtime</span><strong>{status?.runtime_available ? "آماده" : "در دسترس نیست"}</strong></article>
        <article><span>اکانت فعال</span><strong>{activeCount}</strong></article>
        <article><span>کل اکانت‌ها</span><strong>{credentials.length}</strong></article>
        <article title={status?.caddy_sha256 || ""}><span>Caddy SHA</span><strong className="mono">{shortSHA(status?.caddy_sha256)}</strong></article>
      </section>

      {!owned ? (
        <section className="runtime-card ownership-card">
          <div>
            <p className="eyebrow">مرحله ایمن اول</p>
            <h2>اکانت فعلی سرور را به پنل تحویل بده</h2>
            <p>رمز فعلی فقط داخل Unix socket خوانده و رمزنگاری می‌شود؛ به مرورگر، لاگ یا پاسخ API برنمی‌گردد. Caddy در این مرحله تغییر یا reload نمی‌شود.</p>
          </div>
          <button onClick={importLive} disabled={busy || !status?.runtime_available}>{busy ? "در حال بررسی…" : "Import امن اکانت فعلی"}</button>
        </section>
      ) : (
        <>
          <section className="runtime-card">
            <div className="runtime-section-title">
              <div><p className="eyebrow">Add credential</p><h2>ساخت اکانت جدید</h2></div>
            </div>
            <form className="runtime-form" onSubmit={create}>
              <label>نام کاربری
                <input value={username} onChange={(event) => setUsername(event.target.value)} maxLength={64} pattern="[A-Za-z0-9._@+\-]+" required disabled={busy} placeholder="customer.name" />
              </label>
              <label className="runtime-check"><input type="checkbox" checked={generatePassword} onChange={(event) => setGeneratePassword(event.target.checked)} disabled={busy} /> تولید رمز امن توسط سرور</label>
              {!generatePassword && <label>رمز عبور
                <input type="password" value={password} onChange={(event) => setPassword(event.target.value)} minLength={14} maxLength={128} required disabled={busy} autoComplete="new-password" />
              </label>}
              <button type="submit" disabled={busy}>{busy ? "در حال اعمال…" : "ساخت و اعمال اکانت"}</button>
            </form>
          </section>

          <section className="runtime-card">
            <div className="runtime-section-title"><div><p className="eyebrow">Credentials</p><h2>اکانت‌های Runtime</h2></div><button className="button-secondary" onClick={() => run(reload)} disabled={busy}>تازه‌سازی</button></div>
            <div className="runtime-table-wrap">
              <table className="runtime-table">
                <thead><tr><th>نام کاربری</th><th>وضعیت</th><th>منشأ</th><th>Revision</th><th>آخرین تغییر</th><th>عملیات</th></tr></thead>
                <tbody>
                  {credentials.map((credential) => {
                    const lastActive = credential.status === "active" && activeCount <= 1;
                    return <tr key={credential.id}>
                      <td className="mono">{credential.username}</td>
                      <td><span className={`status-pill ${credential.status}`}>{credential.status === "active" ? "فعال" : credential.status === "disabled" ? "غیرفعال" : "لغوشده"}</span></td>
                      <td>{credential.origin === "imported" ? "واردشده" : "پنل"}</td>
                      <td>{credential.revision}</td>
                      <td>{formatTime(credential.updated_at)}</td>
                      <td>
                        {credential.status !== "revoked" && <div className="runtime-actions">
                          <button className="button-link" onClick={() => rename(credential)} disabled={busy}>تغییر نام</button>
                          <button className="button-link" onClick={() => rotateGenerated(credential)} disabled={busy}>رمز تصادفی</button>
                          <button className="button-link" onClick={() => rotateCustom(credential)} disabled={busy}>رمز دلخواه</button>
                          <button className="button-link" onClick={() => toggle(credential)} disabled={busy || lastActive}>{credential.status === "active" ? "غیرفعال" : "فعال"}</button>
                          <button className="button-danger" onClick={() => revoke(credential)} disabled={busy || lastActive}>لغو</button>
                        </div>}
                      </td>
                    </tr>;
                  })}
                </tbody>
              </table>
            </div>
          </section>
        </>
      )}

      {secret && <div className="secret-backdrop" role="presentation">
        <section className="secret-dialog" role="dialog" aria-modal="true" aria-labelledby="secret-title">
          <p className="eyebrow">فقط یک‌بار نمایش</p>
          <h2 id="secret-title">رمز اکانت {secret.username}</h2>
          <p className="runtime-muted">همین الان ذخیره‌اش کن. بعد از بستن این پنجره از پنل قابل مشاهده نیست.</p>
          <code>{secret.password}</code>
          <p className="runtime-muted">لینک آماده برای Karing/کلاینت‌های سازگار با Naive:</p>
          <code className="secret-uri">{customerURI}</code>
          <div className="secret-actions">
            <button onClick={() => navigator.clipboard.writeText(secret.password)}>کپی رمز</button>
            <button onClick={() => navigator.clipboard.writeText(customerURI)}>کپی لینک Karing/Naive</button>
            <button className="button-secondary" onClick={() => setSecret(null)}>ذخیره کردم؛ بستن</button>
          </div>
        </section>
      </div>}
    </div>
  );
}
