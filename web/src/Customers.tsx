import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import {
  adoptRuntimeCustomer,
  createCustomer,
  expiryLabel,
  listCustomers,
  quotaLabel,
  rotateSubscription,
  subscriptionURL,
  updateCustomerService,
  type CustomerCreateResult,
  type CustomerServiceSettingsRequest,
  type CustomerValidityMode,
  type CustomerView,
  type CreateCustomerRequest,
} from "./customers";
import { buildNaiveURI, listRuntimeCredentials, type RuntimeCredential } from "./runtime";
import { encodeQR } from "./qr";
import "./customers.css";

type QuotaMode = "limited" | "unlimited";
type CreateDelivery = { result: CustomerCreateResult; password: string };
type LinkDelivery = { username: string; subscriptionPath: string; notice?: string };
type ServiceEditorTarget =
  | { kind: "adopt"; credential: RuntimeCredential }
  | { kind: "edit"; customer: CustomerView };

function copyText(value: string) {
  return navigator.clipboard.writeText(value);
}

function LocalQRCode({ value }: { value: string }) {
  const matrix = useMemo(() => {
    try { return encodeQR(value); } catch { return null; }
  }, [value]);
  if (!matrix) return <div className="qr-unavailable">لینک برای QR محلی طولانی است؛ از دکمه کپی استفاده کنید.</div>;
  const quiet = 4;
  const size = matrix.length + quiet * 2;
  return <svg className="customer-qr" viewBox={`0 0 ${size} ${size}`} role="img" aria-label="QR لینک اشتراک" shapeRendering="crispEdges">
    <rect width={size} height={size} fill="white" />
    {matrix.flatMap((row, y) => row.map((dark, x) => dark ? <rect key={`${x}-${y}`} x={x + quiet} y={y + quiet} width="1" height="1" fill="black" /> : null))}
  </svg>;
}

function DeliveryModal({ delivery, onClose }: { delivery: CreateDelivery; onClose: () => void }) {
  const { result, password } = delivery;
  const link = subscriptionURL(result.subscription_path);
  const naiveURI = password && typeof window !== "undefined" ? buildNaiveURI(result.runtime_credential.username, password, window.location.hostname) : "";
  const [copied, setCopied] = useState("");
  async function copy(label: string, value: string) {
    await copyText(value); setCopied(label); window.setTimeout(() => setCopied(""), 1500);
  }
  return <div className="delivery-backdrop" role="presentation"><section className="delivery-modal" role="dialog" aria-modal="true" aria-labelledby="delivery-title">
    <div className="delivery-heading"><div><p className="eyebrow">One-time delivery</p><h2 id="delivery-title">اکانت ساخته شد</h2></div><button className="icon-button" onClick={onClose} aria-label="بستن">×</button></div>
    <div className="delivery-grid"><div className="delivery-details">
      <div className="delivery-success">✓ کاربر <strong>{result.user.username}</strong> با موفقیت ایجاد شد.</div>
      <label className="delivery-field"><span>لینک اشتراک</span><div><input readOnly value={link}/><button onClick={() => copy("link", link)}>{copied === "link" ? "کپی شد" : "کپی"}</button></div></label>
      {naiveURI && <label className="delivery-field"><span>لینک مستقیم Naive</span><div><input readOnly value={naiveURI}/><button onClick={() => copy("naive", naiveURI)}>{copied === "naive" ? "کپی شد" : "کپی"}</button></div></label>}
      {password && <label className="delivery-field secret-field"><span>رمز عبور — فقط همین بار نمایش داده می‌شود</span><div><input readOnly value={password}/><button onClick={() => copy("password", password)}>{copied === "password" ? "کپی شد" : "کپی"}</button></div></label>}
      <div className="delivery-meta"><span>حجم</span><strong>{quotaLabel(result.service_term.quota_bytes ?? null)}</strong></div>
      <div className="delivery-meta"><span>وضعیت سرویس</span><strong>{result.service_term.state === "pending" ? "منتظر اولین اتصال موفق" : "فعال"}</strong></div>
      <div className="delivery-meta"><span>مصرف</span><strong>{result.usage_capability.available ? "فعال" : "در انتظار Accounting دقیق"}</strong></div>
      {!result.usage_capability.available && <p className="capability-warning">تا زمان اثبات Accounting دقیق Naive، پنل عدد مصرف ساختگی نمایش نمی‌دهد و قطع حجمی را فعال نمی‌کند.</p>}
    </div><div className="delivery-qr-wrap"><LocalQRCode value={link}/><small>QR در خود مرورگر ساخته می‌شود و لینک به سرویس ثالث ارسال نمی‌شود.</small></div></div>
    <div className="delivery-actions"><button className="button-secondary" onClick={() => copy("link", link)}>کپی لینک اشتراک</button><button className="primary-action" onClick={onClose}>تحویل انجام شد</button></div>
  </section></div>;
}

function LinkDeliveryModal({ delivery, onClose }: { delivery: LinkDelivery; onClose: () => void }) {
  const link = subscriptionURL(delivery.subscriptionPath);
  const [copied, setCopied] = useState(false);
  return <div className="delivery-backdrop" role="presentation"><section className="delivery-modal compact-delivery" role="dialog" aria-modal="true" aria-labelledby="subscription-delivery-title">
    <div className="delivery-heading"><div><p className="eyebrow">Subscription</p><h2 id="subscription-delivery-title">لینک {delivery.username}</h2></div><button className="icon-button" onClick={onClose} aria-label="بستن">×</button></div>
    <div className="delivery-grid"><div className="delivery-details"><p className="capability-warning">{delivery.notice || "این لینک را همین حالا تحویل دهید."}</p><label className="delivery-field"><span>لینک اشتراک</span><div><input readOnly value={link}/><button onClick={async () => { await copyText(link); setCopied(true); }}>{copied ? "کپی شد" : "کپی"}</button></div></label></div><div className="delivery-qr-wrap"><LocalQRCode value={link}/><small>QR محلی</small></div></div>
    <div className="delivery-actions"><button className="primary-action" onClick={onClose}>بستن</button></div>
  </section></div>;
}

function initialQuota(customer?: CustomerView): { mode: QuotaMode; gb: string } {
  if (!customer || customer.quota_bytes === null) return { mode: "unlimited", gb: "50" };
  return { mode: "limited", gb: String(Math.max(1, Math.round(customer.quota_bytes / 1073741824))) };
}

function ServiceEditorModal({ target, onClose, onSaved }: {
  target: ServiceEditorTarget;
  onClose: () => void;
  onSaved: (delivery?: LinkDelivery) => Promise<void>;
}) {
  const customer = target.kind === "edit" ? target.customer : undefined;
  const initial = initialQuota(customer);
  const [quotaMode, setQuotaMode] = useState<QuotaMode>(initial.mode);
  const [quotaGB, setQuotaGB] = useState(initial.gb);
  const [validityMode, setValidityMode] = useState<"on_creation" | "fixed_expiry">(customer?.expires_at ? "fixed_expiry" : "on_creation");
  const [durationDays, setDurationDays] = useState("30");
  const [fixedExpiry, setFixedExpiry] = useState(() => {
    if (!customer?.expires_at) return "";
    const d = new Date(customer.expires_at);
    const local = new Date(d.getTime() - d.getTimezoneOffset() * 60000);
    return local.toISOString().slice(0, 16);
  });
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const username = target.kind === "adopt" ? target.credential.username : target.customer.username;

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      const validity: CustomerServiceSettingsRequest["validity"] = { mode: validityMode };
      if (validityMode === "fixed_expiry") {
        if (!fixedExpiry) throw new Error("تاریخ پایان را وارد کنید.");
        validity.expires_at = new Date(fixedExpiry).toISOString();
      } else {
        validity.duration_days = Number(durationDays);
      }
      const settings: CustomerServiceSettingsRequest = {
        quota_gb: quotaMode === "unlimited" ? null : Number(quotaGB),
        validity,
      };
      if (target.kind === "adopt") {
        const result = await adoptRuntimeCustomer(target.credential.id, settings);
        await onSaved({
          username,
          subscriptionPath: result.subscription_path,
          notice: "رمز، نام کاربری و کانفیگ قبلی بدون تغییر ماند. فقط حجم، اعتبار و Subscription به مدیریت مشتری اضافه شد.",
        });
      } else {
        const result = await updateCustomerService(target.customer.id, settings);
        if (result.runtime_mutated) throw new Error("ویرایش سرویس به‌طور غیرمنتظره Runtime را تغییر داد.");
        await onSaved();
      }
    } catch (cause) {
      setMessage(cause instanceof Error ? cause.message : "ذخیره تنظیمات انجام نشد.");
      setBusy(false);
      return;
    }
    setBusy(false);
    onClose();
  }

  return <div className="delivery-backdrop" role="presentation"><section className="delivery-modal compact-delivery" role="dialog" aria-modal="true">
    <div className="delivery-heading"><div><p className="eyebrow">Service management</p><h2>{target.kind === "adopt" ? `مدیریت اکانت قدیمی ${username}` : `ویرایش ${username}`}</h2></div><button className="icon-button" onClick={onClose} aria-label="بستن">×</button></div>
    {target.kind === "adopt" && <p className="capability-warning">این عملیات رمز، Username و Runtime UUID فعلی را تغییر نمی‌دهد و Caddy credential جدید نمی‌سازد.</p>}
    <form onSubmit={save} className="form-grid">
      <fieldset className="field-group field-wide"><legend>حجم سرویس</legend><div className="segmented"><button type="button" className={quotaMode === "limited" ? "active" : ""} onClick={() => setQuotaMode("limited")}>حجمی</button><button type="button" className={quotaMode === "unlimited" ? "active" : ""} onClick={() => setQuotaMode("unlimited")}>نامحدود</button></div>{quotaMode === "limited" && <label className="inline-input"><input type="number" min="1" step="1" value={quotaGB} onChange={(e) => setQuotaGB(e.target.value)} required/><span>GB</span></label>}</fieldset>
      <fieldset className="field-group field-wide"><legend>اعتبار</legend><div className="validity-options"><label><input type="radio" checked={validityMode === "on_creation"} onChange={() => setValidityMode("on_creation")}/><span><strong>از همین حالا</strong><small>تعداد روز را از زمان ذخیره محاسبه می‌کند.</small></span></label><label><input type="radio" checked={validityMode === "fixed_expiry"} onChange={() => setValidityMode("fixed_expiry")}/><span><strong>تاریخ پایان دستی</strong><small>زمان پایان دقیق تعیین می‌شود.</small></span></label></div>{validityMode === "fixed_expiry" ? <label className="field compact-field">تاریخ پایان<input type="datetime-local" value={fixedExpiry} onChange={(e) => setFixedExpiry(e.target.value)} required/></label> : <label className="inline-input"><input type="number" min="1" step="1" value={durationDays} onChange={(e) => setDurationDays(e.target.value)} required/><span>روز</span></label>}</fieldset>
      {message && <p className="customer-message field-wide" role="status">{message}</p>}
      <div className="delivery-actions field-wide"><button type="button" className="button-secondary" onClick={onClose} disabled={busy}>انصراف</button><button className="primary-action" type="submit" disabled={busy}>{busy ? "در حال ذخیره…" : target.kind === "adopt" ? "افزودن به مدیریت مشتری" : "ذخیره حجم و تاریخ"}</button></div>
    </form>
  </section></div>;
}

function statusLabel(customer: CustomerView): string {
  if (customer.status === "suspended") return "تعلیق";
  if (customer.status === "revoked") return "لغوشده";
  if (customer.service_state === "pending") return "منتظر اتصال";
  if (customer.service_state === "expired") return "منقضی";
  if (customer.service_state === "quota_depleted") return "حجم تمام";
  return "فعال";
}

function startModeLabel(mode: string): string {
  if (mode === "on_first_successful_connection") return "اتصال اول";
  if (mode === "on_creation") return "از ساخت/ویرایش";
  return "تاریخ دستی";
}

export function Customers() {
  const [username, setUsername] = useState("");
  const [quotaMode, setQuotaMode] = useState<QuotaMode>("limited");
  const [quotaGB, setQuotaGB] = useState("50");
  const [validityMode, setValidityMode] = useState<CustomerValidityMode>("on_creation");
  const [durationDays, setDurationDays] = useState("30");
  const [fixedExpiry, setFixedExpiry] = useState("");
  const [generatePassword, setGeneratePassword] = useState(true);
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState("");
  const [delivery, setDelivery] = useState<CreateDelivery | null>(null);
  const [linkDelivery, setLinkDelivery] = useState<LinkDelivery | null>(null);
  const [customers, setCustomers] = useState<CustomerView[]>([]);
  const [runtimeCredentials, setRuntimeCredentials] = useState<RuntimeCredential[]>([]);
  const [editor, setEditor] = useState<ServiceEditorTarget | null>(null);
  const [loading, setLoading] = useState(true);
  const [listMessage, setListMessage] = useState("");
  const [rotatingID, setRotatingID] = useState("");

  const unmanagedRuntime = useMemo(() => {
    const managed = new Set(customers.map((item) => item.runtime_credential_id));
    return runtimeCredentials.filter((credential) => credential.status === "active" && !managed.has(credential.id));
  }, [customers, runtimeCredentials]);

  const refresh = useCallback(async () => {
    setListMessage("");
    try {
      const [nextCustomers, nextRuntime] = await Promise.all([listCustomers(), listRuntimeCredentials()]);
      setCustomers(nextCustomers); setRuntimeCredentials(nextRuntime);
    } catch (cause) {
      setListMessage(cause instanceof Error ? cause.message : "فهرست مشتریان بارگذاری نشد.");
    } finally { setLoading(false); }
  }, []);
  useEffect(() => { void refresh(); }, [refresh]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setMessage(""); setSubmitting(true);
    try {
      const validity: CreateCustomerRequest["validity"] = { mode: validityMode };
      if (validityMode === "fixed_expiry") {
        if (!fixedExpiry) throw new Error("تاریخ پایان را وارد کنید.");
        validity.expires_at = new Date(fixedExpiry).toISOString();
      } else validity.duration_days = Number(durationDays);
      const request: CreateCustomerRequest = { username: username.trim(), password: generatePassword ? "" : password, generate_password: generatePassword, quota_gb: quotaMode === "unlimited" ? null : Number(quotaGB), validity };
      const localPassword = generatePassword ? "" : password;
      const result = await createCustomer(request);
      setDelivery({ result, password: result.generated_password || localPassword });
      setPassword(""); setUsername(""); setMessage("اکانت آماده تحویل است."); await refresh();
    } catch (cause) { setMessage(cause instanceof Error ? cause.message : "ساخت اکانت انجام نشد."); }
    finally { setSubmitting(false); }
  }

  async function reissue(customer: CustomerView) {
    if (!window.confirm(`لینک اشتراک قبلی ${customer.username} لغو و لینک جدید ساخته شود؟`)) return;
    setRotatingID(customer.id); setListMessage("");
    try {
      const result = await rotateSubscription(customer.id);
      setLinkDelivery({ username: customer.username, subscriptionPath: result.subscription_path, notice: "لینک قبلی لغو شده است. این لینک جدید را تحویل دهید." });
      await refresh();
    } catch (cause) { setListMessage(cause instanceof Error ? cause.message : "صدور لینک جدید انجام نشد."); }
    finally { setRotatingID(""); }
  }

  async function editorSaved(nextDelivery?: LinkDelivery) {
    if (nextDelivery) setLinkDelivery(nextDelivery);
    setListMessage("تغییرات ذخیره شد؛ رمز و Runtime credential بدون تغییر باقی ماند.");
    await refresh();
  }

  return <main className="customers-page">
    <header className="customers-header"><div><p className="eyebrow">Customer service</p><h1>مدیریت اکانت‌های Naive</h1><p>اکانت‌های جدید و قدیمی را با حجم، اعتبار، Subscription و QR مدیریت کنید؛ رمز قدیمی هنگام Adopt تغییر نمی‌کند.</p></div><span className="badge">Direct · NaiveProxy</span></header>
    <section className="customer-summary" aria-label="وضعیت قابلیت‌ها"><article><span>شروع پیش‌فرض</span><strong>از همین حالا</strong></article><article><span>اکانت قدیمی</span><strong>Adopt بدون تغییر رمز</strong></article><article><span>تحویل</span><strong>Subscription + QR</strong></article><article><span>Accounting دقیق</span><strong className="muted-value">هنوز قفل است</strong></article></section>

    {unmanagedRuntime.length > 0 && <section className="customer-table-card"><div className="section-title"><div><p className="eyebrow">Legacy Runtime</p><h2>اکانت‌های قبلی Runtime</h2><p>این اکانت‌ها فعال‌اند ولی هنوز حجم و تاریخ Customer ندارند. افزودن به مدیریت، رمز و کانفیگ فعلی را تغییر نمی‌دهد.</p></div></div><div className="customer-table-wrap"><table className="customer-table"><thead><tr><th>نام کاربری</th><th>منشأ</th><th>Runtime ID</th><th>عملیات</th></tr></thead><tbody>{unmanagedRuntime.map((credential) => <tr key={credential.id}><td><strong>{credential.username}</strong></td><td>{credential.origin === "imported" ? "واردشده" : "پنل قدیمی"}</td><td><small>{credential.id.slice(0, 12)}</small></td><td><button className="primary-action compact-button" onClick={() => setEditor({ kind: "adopt", credential })}>افزودن حجم و تاریخ</button></td></tr>)}</tbody></table></div></section>}

    <section className="customer-layout">
      <form className="customer-create-card" onSubmit={submit}><div className="section-title"><div><p className="eyebrow">Create user</p><h2>ساخت اکانت جدید</h2></div><span className="step-pill">۱ مرحله</span></div><div className="form-grid">
        <label className="field field-wide">نام کاربری<input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="مثلاً saman01" autoComplete="off" required disabled={submitting}/></label>
        <fieldset className="field-group field-wide"><legend>حجم سرویس</legend><div className="segmented"><button type="button" className={quotaMode === "limited" ? "active" : ""} onClick={() => setQuotaMode("limited")}>حجمی</button><button type="button" className={quotaMode === "unlimited" ? "active" : ""} onClick={() => setQuotaMode("unlimited")}>نامحدود</button></div>{quotaMode === "limited" && <label className="inline-input"><input type="number" min="1" step="1" value={quotaGB} onChange={(e) => setQuotaGB(e.target.value)} required/><span>GB</span></label>}</fieldset>
        <fieldset className="field-group field-wide"><legend>شروع و پایان اعتبار</legend><div className="validity-options"><label><input type="radio" name="validity" checked={validityMode === "on_creation"} onChange={() => setValidityMode("on_creation")}/><span><strong>از همین حالا</strong><small>اعتبار بلافاصله شروع می‌شود.</small></span></label><label><input type="radio" name="validity" checked={validityMode === "fixed_expiry"} onChange={() => setValidityMode("fixed_expiry")}/><span><strong>تاریخ پایان دستی</strong><small>یک زمان پایان دقیق تعیین کنید.</small></span></label><label><input type="radio" name="validity" checked={validityMode === "on_first_successful_connection"} onChange={() => setValidityMode("on_first_successful_connection")}/><span><strong>از اولین اتصال موفق</strong><small>فقط پس از اثبات producer معتبر Runtime استفاده شود.</small></span></label></div>{validityMode === "fixed_expiry" ? <label className="field compact-field">تاریخ پایان<input type="datetime-local" value={fixedExpiry} onChange={(e) => setFixedExpiry(e.target.value)} required/></label> : <label className="inline-input"><input type="number" min="1" step="1" value={durationDays} onChange={(e) => setDurationDays(e.target.value)} required/><span>روز</span></label>}</fieldset>
        <fieldset className="field-group field-wide"><legend>رمز عبور</legend><label className="check-row"><input type="checkbox" checked={generatePassword} onChange={(e) => setGeneratePassword(e.target.checked)}/><span><strong>تولید رمز امن خودکار</strong><small>رمز فقط بعد از ساخت و یک بار نمایش داده می‌شود.</small></span></label>{!generatePassword && <label className="field compact-field">رمز سفارشی<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} minLength={12} required autoComplete="new-password"/></label>}</fieldset>
      </div>{message && <p className="customer-message" role="status">{message}</p>}<div className="create-footer"><p>Accounting دقیق هنوز فعال نیست؛ حجم به‌عنوان قرارداد سرویس ذخیره می‌شود و قطع حجمی خودکار فعلاً انجام نمی‌شود.</p><button className="primary-action" type="submit" disabled={submitting}>{submitting ? "در حال ساخت…" : "ساخت و تحویل اکانت"}</button></div></form>
      <aside className="customer-safety-card"><p className="eyebrow">Safety gates</p><h3>اکانت قبلی دست‌نخورده می‌ماند</h3><p>Adopt فقط Customer metadata و Subscription می‌سازد. Password، Username و Runtime UUID قبلی را تغییر نمی‌دهد.</p><ul><li>Token خام در دیتابیس ذخیره نمی‌شود.</li><li>رمز Runtime رمزنگاری‌شده است.</li><li>QR کاملاً محلی ساخته می‌شود.</li><li>Accounting دقیق تا اثبات capability قفل است.</li></ul><a className="button-secondary" href="/panel/#/runtime/naive">Runtime پیشرفته</a></aside>
    </section>

    <section className="customer-table-card"><div className="section-title"><div><p className="eyebrow">Customers</p><h2>اکانت‌های مدیریت‌شده</h2></div><button className="button-secondary" onClick={() => void refresh()} disabled={loading}>بروزرسانی</button></div>{listMessage && <p className="customer-message" role="status">{listMessage}</p>}<div className="customer-table-wrap"><table className="customer-table"><thead><tr><th>نام کاربری</th><th>وضعیت</th><th>حجم</th><th>مصرف / باقی‌مانده</th><th>انقضا</th><th>شروع</th><th>عملیات</th></tr></thead><tbody>
      {loading && <tr><td colSpan={7}>در حال بارگذاری…</td></tr>}{!loading && customers.length === 0 && <tr><td colSpan={7}>هنوز اکانتی وارد مدیریت مشتری نشده است.</td></tr>}{customers.map((customer) => <tr key={customer.id}><td><strong>{customer.username}</strong><small>{customer.runtime_credential_id.slice(0, 8)}</small></td><td><span className={`state-pill state-${customer.service_state}`}>{statusLabel(customer)}</span></td><td>{quotaLabel(customer.quota_bytes)}</td><td>{customer.usage_capability.available ? "فعال" : <span className="usage-locked">در دسترس نیست — حسابداری دقیق هنوز تأیید نشده</span>}</td><td>{expiryLabel(customer.expires_at)}</td><td>{startModeLabel(customer.start_policy)}</td><td><div className="runtime-actions"><button className="button-secondary compact-button" onClick={() => setEditor({ kind: "edit", customer })}>ویرایش حجم/تاریخ</button><button className="button-secondary compact-button" disabled={rotatingID === customer.id} onClick={() => void reissue(customer)}>{rotatingID === customer.id ? "در حال صدور…" : customer.subscription_available ? "لینک جدید + QR" : "صدور لینک + QR"}</button></div></td></tr>)}
    </tbody></table></div></section>
    {delivery && <DeliveryModal delivery={delivery} onClose={() => setDelivery(null)}/>} {linkDelivery && <LinkDeliveryModal delivery={linkDelivery} onClose={() => setLinkDelivery(null)}/>} {editor && <ServiceEditorModal target={editor} onClose={() => setEditor(null)} onSaved={editorSaved}/>} 
  </main>;
}
