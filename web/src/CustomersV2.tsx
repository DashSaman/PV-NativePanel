import { FormEvent, ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import {
  adoptRuntimeCustomer,
  createCustomer,
  deleteCustomer,
  expiryLabel,
  getCurrentSubscription,
  listCustomers,
  quotaLabel,
  resumeCustomer,
  rotateCustomerPassword,
  rotateSubscription,
  subscriptionURL,
  suspendCustomer,
  updateCustomerService,
  type CustomerCreateResult,
  type CustomerServiceSettingsRequest,
  type CustomerValidityMode,
  type CustomerView,
} from "./customers";
import { addCustomerVolume, extendCustomerTime } from "./customerAdjustments";
import { listRuntimeCredentials, type RuntimeCredential } from "./runtime";
import { encodeQR } from "./qr";
import { buildUnifiedCustomerRows } from "./customerRows";
import { buildCustomerDashboard, type CustomerDashboard } from "./customerDashboard";
import "./customers.css";
import "./customers-v2.css";

type DialogKind =
  | { type: "create" }
  | { type: "adopt"; credential: RuntimeCredential }
  | { type: "edit"; customer: CustomerView }
  | { type: "details"; customer: CustomerView }
  | { type: "password"; customer: CustomerView }
  | { type: "add-volume"; customer: CustomerView }
  | { type: "extend"; customer: CustomerView }
  | { type: "subscription"; customer: CustomerView; path: string; directURI?: string; notice?: string; readOnly: boolean }
  | { type: "secret"; username: string; password: string; notice: string };

type StatusFilter = "all" | "active" | "suspended" | "revoked" | "pending" | "expired" | "depleted" | "unmanaged";
type SortKey = "username" | "created" | "expiry" | "quota";
type BulkAction = "suspend" | "resume" | "delete" | "add-volume" | "extend";

function Dialog({ title, eyebrow, children, onClose, wide = false }: { title: string; eyebrow: string; children: ReactNode; onClose: () => void; wide?: boolean }) {
  return (
    <div className="delivery-backdrop" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <section className={`delivery-modal ${wide ? "dialog-wide" : "compact-delivery"}`} role="dialog" aria-modal="true">
        <div className="delivery-heading">
          <div><p className="eyebrow">{eyebrow}</p><h2>{title}</h2></div>
          <button className="icon-button" onClick={onClose} aria-label="بستن">×</button>
        </div>
        {children}
      </section>
    </div>
  );
}

function QR({ value }: { value: string }) {
  const matrix = useMemo(() => {
    try { return encodeQR(value); } catch { return null; }
  }, [value]);
  if (!matrix) return <div className="qr-unavailable">QR برای این مقدار قابل ساخت نیست؛ لینک را کپی کنید.</div>;
  const quiet = 4;
  const size = matrix.length + quiet * 2;
  return (
    <svg className="customer-qr" viewBox={`0 0 ${size} ${size}`} role="img" aria-label="QR اشتراک" shapeRendering="crispEdges">
      <rect width={size} height={size} fill="white" />
      {matrix.flatMap((row, y) => row.map((dark, x) => dark ? <rect key={`${x}-${y}`} x={x + quiet} y={y + quiet} width="1" height="1" fill="black" /> : null))}
    </svg>
  );
}

function statusDimension(customer: CustomerView): { label: string; tone: string } {
  if (customer.status === "revoked") return { label: "لغوشده", tone: "danger" };
  if (customer.status === "suspended") return { label: "تعلیق", tone: "warning" };
  if (customer.service_state === "pending") return { label: "منتظر اتصال", tone: "info" };
  if (customer.service_state === "expired") return { label: "منقضی", tone: "danger" };
  if (customer.service_state === "quota_depleted") return { label: "حجم تمام", tone: "danger" };
  return { label: "فعال", tone: "success" };
}

function startPolicyLabel(value: string) {
  if (value === "on_first_successful_connection") return "اولین اتصال موفق";
  if (value === "fixed_timestamp") return "تاریخ دستی";
  return "از زمان ثبت";
}

function toLocalDateTime(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function ServiceForm({ mode, customer, credential, onSaved, onClose }: {
  mode: "create" | "edit" | "adopt";
  customer?: CustomerView;
  credential?: RuntimeCredential;
  onSaved: (result?: CustomerCreateResult) => Promise<void>;
  onClose: () => void;
}) {
  const [username, setUsername] = useState(customer?.username || credential?.username || "");
  const [quotaMode, setQuotaMode] = useState<"limited" | "unlimited">(customer?.quota_bytes === null ? "unlimited" : "limited");
  const [quotaGB, setQuotaGB] = useState(() => customer?.quota_bytes ? String(Math.max(1, Math.round(customer.quota_bytes / 1073741824))) : "50");
  const [validityMode, setValidityMode] = useState<CustomerValidityMode>(() => {
    if (customer?.start_policy === "on_first_successful_connection") return "on_first_successful_connection";
    if (customer?.start_policy === "fixed_timestamp") return "fixed_expiry";
    return "on_creation";
  });
  const [days, setDays] = useState(() => String(Math.max(1, Math.ceil((customer?.duration_seconds || 30 * 86400) / 86400))));
  const [fixedExpiry, setFixedExpiry] = useState(() => toLocalDateTime(customer?.expires_at));
  const [generatePassword, setGeneratePassword] = useState(true);
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  function settings(): CustomerServiceSettingsRequest {
    const validity: CustomerServiceSettingsRequest["validity"] = { mode: validityMode };
    if (validityMode === "fixed_expiry") {
      if (!fixedExpiry) throw new Error("تاریخ پایان را مشخص کنید.");
      validity.expires_at = new Date(fixedExpiry).toISOString();
    } else {
      validity.duration_days = Number(days);
    }
    return { quota_gb: quotaMode === "unlimited" ? null : Number(quotaGB), validity };
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    setBusy(true); setMessage("");
    try {
      let result: CustomerCreateResult | undefined;
      if (mode === "create") {
        result = await createCustomer({
          username: username.trim(), password: generatePassword ? "" : password,
          generate_password: generatePassword, ...settings(),
        });
      } else if (mode === "adopt" && credential) {
        result = await adoptRuntimeCustomer(credential.id, settings());
      } else if (mode === "edit" && customer) {
        const updated = await updateCustomerService(customer.id, settings());
        if (updated.runtime_mutated) throw new Error("ویرایش سرویس نباید Runtime را تغییر دهد.");
      }
      await onSaved(result);
      onClose();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "ذخیره انجام نشد.");
    } finally { setBusy(false); }
  }

  return (
    <form className="owner-form" onSubmit={submit}>
      {(mode === "create") && <label>نام کاربری<input value={username} onChange={(e) => setUsername(e.target.value)} required autoComplete="off" /></label>}
      {(mode === "adopt") && <div className="readonly-banner">Username و Password فعلی <strong>{credential?.username}</strong> حفظ می‌شود.</div>}
      <fieldset><legend>حجم</legend><div className="segmented"><button type="button" className={quotaMode === "limited" ? "active" : ""} onClick={() => setQuotaMode("limited")}>حجمی</button><button type="button" className={quotaMode === "unlimited" ? "active" : ""} onClick={() => setQuotaMode("unlimited")}>نامحدود</button></div>{quotaMode === "limited" && <label>حجم کل (GB)<input type="number" min="1" step="1" value={quotaGB} onChange={(e) => setQuotaGB(e.target.value)} required /></label>}</fieldset>
      <fieldset><legend>اعتبار</legend><div className="validity-options"><label><input type="radio" checked={validityMode === "on_creation"} onChange={() => setValidityMode("on_creation")} /><span><strong>از همین حالا</strong><small>مدت از زمان ثبت این term</small></span></label><label><input type="radio" checked={validityMode === "on_first_successful_connection"} onChange={() => setValidityMode("on_first_successful_connection")} /><span><strong>از اولین اتصال موفق</strong><small>فقط producer معتبر Runtime حق فعال‌سازی دارد</small></span></label><label><input type="radio" checked={validityMode === "fixed_expiry"} onChange={() => setValidityMode("fixed_expiry")} /><span><strong>تاریخ دستی</strong><small>پایان دقیق</small></span></label></div>{validityMode === "fixed_expiry" ? <label>تاریخ پایان<input type="datetime-local" value={fixedExpiry} onChange={(e) => setFixedExpiry(e.target.value)} required /></label> : <label>تعداد روز<input type="number" min="1" step="1" value={days} onChange={(e) => setDays(e.target.value)} required /></label>}</fieldset>
      {mode === "create" && <fieldset><legend>رمز</legend><label className="check-row"><input type="checkbox" checked={generatePassword} onChange={(e) => setGeneratePassword(e.target.checked)} /><span><strong>رمز امن خودکار</strong><small>فقط یک بار بعد از ساخت نمایش داده می‌شود</small></span></label>{!generatePassword && <label>رمز سفارشی<input type="password" minLength={12} value={password} onChange={(e) => setPassword(e.target.value)} required /></label>}</fieldset>}
      {message && <p className="customer-message">{message}</p>}
      <div className="delivery-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال ذخیره…" : "ذخیره"}</button></div>
    </form>
  );
}

function PasswordForm({ customer, onDone, onClose }: { customer: CustomerView; onDone: (password?: string) => Promise<void>; onClose: () => void }) {
  const [generate, setGenerate] = useState(true);
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      const result = await rotateCustomerPassword(customer.id, { password: generate ? "" : password, generate_password: generate });
      await onDone(result.generated_password || (generate ? undefined : password));
      onClose();
    } catch (error) { setMessage(error instanceof Error ? error.message : "تغییر رمز انجام نشد."); }
    finally { setBusy(false); }
  }
  return <form className="owner-form" onSubmit={submit}><p className="readonly-banner">این عملیات فقط Password همان Runtime credential را تغییر می‌دهد و Subscription token را reissue نمی‌کند.</p><label className="check-row"><input type="checkbox" checked={generate} onChange={(e) => setGenerate(e.target.checked)} /><span><strong>تولید رمز امن</strong><small>نمایش یک‌باره</small></span></label>{!generate && <label>رمز جدید<input type="password" minLength={12} value={password} onChange={(e) => setPassword(e.target.value)} required /></label>}{message && <p className="customer-message">{message}</p>}<div className="delivery-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال تغییر…" : "تغییر رمز"}</button></div></form>;
}

function NumberActionForm({ label, suffix, initial, onSubmit, onClose }: { label: string; suffix: string; initial: number; onSubmit: (value: number) => Promise<void>; onClose: () => void }) {
  const [value, setValue] = useState(String(initial));
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try { await onSubmit(Number(value)); onClose(); }
    catch (error) { setMessage(error instanceof Error ? error.message : "عملیات انجام نشد."); }
    finally { setBusy(false); }
  }
  return <form className="owner-form" onSubmit={submit}><label>{label}<div className="number-with-suffix"><input type="number" min="1" step="1" value={value} onChange={(e) => setValue(e.target.value)} required /><span>{suffix}</span></div></label>{message && <p className="customer-message">{message}</p>}<div className="delivery-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>اعمال</button></div></form>;
}

function SubscriptionView({ customer, path, directURI, notice, readOnly }: { customer: CustomerView; path: string; directURI?: string; notice?: string; readOnly: boolean }) {
  const link = subscriptionURL(path);
  const [copied, setCopied] = useState("");
  async function copy(name: string, value: string) { await navigator.clipboard.writeText(value); setCopied(name); window.setTimeout(() => setCopied(""), 1200); }
  return <div className="subscription-view"><p className="readonly-banner">{notice || (readOnly ? "نمایش فقط‌خواندنی؛ هیچ Stateی تغییر نکرد." : "لینک جدید صادر شد.")}</p><div className="subscription-layout"><div><label>Subscription URL<div className="copy-row"><input readOnly value={link} /><button onClick={() => void copy("sub", link)}>{copied === "sub" ? "کپی شد" : "کپی"}</button></div></label>{directURI && <label>Direct Naive<div className="copy-row"><input readOnly value={directURI} /><button onClick={() => void copy("direct", directURI)}>{copied === "direct" ? "کپی شد" : "کپی"}</button></div></label>}</div><div className="delivery-qr-wrap"><QR value={link} /><small>QR محلی؛ هیچ URL حساسی به سرویس ثالث ارسال نمی‌شود.</small></div></div><p className="muted-line">{customer.username}</p></div>;
}

function StatusDonut({ dashboard }: { dashboard: CustomerDashboard }) {
  const segments = [
    { key: "active", label: "فعال", value: dashboard.active },
    { key: "pending", label: "منتظر اتصال", value: dashboard.pending },
    { key: "suspended", label: "تعلیق", value: dashboard.suspended },
    { key: "ended", label: "پایان‌یافته", value: dashboard.ended },
    { key: "setup", label: "نیازمند تنظیم", value: dashboard.needsSetup },
  ];
  const total = Math.max(1, segments.reduce((sum, item) => sum + item.value, 0));
  let offset = 0;
  return <div className="status-chart"><div className="donut-wrap"><svg className="status-donut" viewBox="0 0 120 120" role="img" aria-label="نمودار وضعیت اکانت‌ها"><circle className="donut-track" cx="60" cy="60" r="46" pathLength="100" />{segments.map((item) => { const percent = item.value / total * 100; const dashOffset = -offset; offset += percent; return <circle key={item.key} className={`donut-segment ${item.key}`} cx="60" cy="60" r="46" pathLength="100" strokeDasharray={`${percent} ${100 - percent}`} strokeDashoffset={dashOffset} />; })}</svg><div className="donut-center"><strong>{dashboard.totalAccounts}</strong><span>اکانت</span></div></div><div className="chart-legend">{segments.map((item) => <div key={item.key}><i className={`legend-dot ${item.key}`} /><span>{item.label}</span><strong>{item.value}</strong></div>)}</div></div>;
}

function ExpiryBars({ dashboard }: { dashboard: CustomerDashboard }) {
  const items = [
    { key: "soon", label: "تا ۷ روز", value: dashboard.expiry.within7Days },
    { key: "month", label: "۸ تا ۳۰ روز", value: dashboard.expiry.within30Days },
    { key: "later", label: "بیش از ۳۰ روز", value: dashboard.expiry.later },
    { key: "none", label: "بدون انقضا / تنظیم", value: dashboard.expiry.noExpiry },
  ];
  const max = Math.max(1, ...items.map((item) => item.value));
  return <div className="expiry-chart">{items.map((item) => <div className="expiry-row" key={item.key}><div className="expiry-label"><span>{item.label}</span><strong>{item.value}</strong></div><div className="expiry-track"><span className={`expiry-fill ${item.key}`} style={{ width: `${item.value / max * 100}%` }} /></div></div>)}</div>;
}

export function CustomersV2() {
  const [customers, setCustomers] = useState<CustomerView[]>([]);
  const [runtime, setRuntime] = useState<RuntimeCredential[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [dialog, setDialog] = useState<DialogKind | null>(null);
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState<StatusFilter>("all");
  const [sort, setSort] = useState<SortKey>("username");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [busyID, setBusyID] = useState("");
  const [bulkBusy, setBulkBusy] = useState(false);

  const refresh = useCallback(async () => {
    setMessage("");
    try {
      const [nextCustomers, nextRuntime] = await Promise.all([listCustomers(), listRuntimeCredentials()]);
      setCustomers(nextCustomers); setRuntime(nextRuntime);
      setSelected((current) => new Set([...current].filter((id) => nextCustomers.some((item) => item.id === id))));
    } catch (error) { setMessage(error instanceof Error ? error.message : "بارگذاری انجام نشد."); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void refresh(); }, [refresh]);

  const unifiedRows = useMemo(() => buildUnifiedCustomerRows(customers, runtime), [customers, runtime]);
  const dashboard = useMemo(() => buildCustomerDashboard(unifiedRows), [unifiedRows]);

  const visible = useMemo(() => {
    const needle = search.trim().toLowerCase();
    const matches = unifiedRows.filter((row) => {
      if (needle && !`${row.username} ${row.id} ${row.runtimeCredentialID}`.toLowerCase().includes(needle)) return false;
      if (row.kind === "runtime") {
        if (filter === "all" || filter === "unmanaged") return true;
        if (filter === "active") return row.credential.status === "active";
        return false;
      }
      const customer = row.customer;
      if (filter === "unmanaged") return false;
      if (filter === "all") return true;
      if (filter === "suspended" || filter === "revoked" || filter === "active") return customer.status === filter;
      if (filter === "pending") return customer.service_state === "pending";
      if (filter === "expired") return customer.service_state === "expired";
      if (filter === "depleted") return customer.service_state === "quota_depleted";
      return true;
    });
    matches.sort((a, b) => {
      if (sort === "expiry") {
        const av = a.kind === "customer" ? (a.customer.expires_at || "9999") : "9999";
        const bv = b.kind === "customer" ? (b.customer.expires_at || "9999") : "9999";
        return av.localeCompare(bv);
      }
      if (sort === "quota") {
        const av = a.kind === "customer" ? (a.customer.quota_bytes ?? Number.MAX_SAFE_INTEGER) : Number.MAX_SAFE_INTEGER;
        const bv = b.kind === "customer" ? (b.customer.quota_bytes ?? Number.MAX_SAFE_INTEGER) : Number.MAX_SAFE_INTEGER;
        return av - bv;
      }
      if (sort === "created") return a.id.localeCompare(b.id);
      return a.username.localeCompare(b.username, "fa");
    });
    return matches;
  }, [unifiedRows, filter, search, sort]);

  const pageCount = Math.max(1, Math.ceil(visible.length / pageSize));
  useEffect(() => { if (page > pageCount) setPage(pageCount); }, [page, pageCount]);
  const pageRows = visible.slice((page - 1) * pageSize, page * pageSize);
  const selectablePageIDs = pageRows.flatMap((row) => row.kind === "customer" ? [row.customer.id] : []);

  const selectedCustomers = customers.filter((customer) => selected.has(customer.id));

  function toggle(id: string) {
    setSelected((current) => { const next = new Set(current); if (next.has(id)) next.delete(id); else next.add(id); return next; });
  }

  async function saved(result?: CustomerCreateResult) {
    if (result) {
      const password = result.generated_password;
      if (password) setDialog({ type: "secret", username: result.user.username, password, notice: "رمز فقط همین بار نمایش داده می‌شود." });
      else setMessage(`اکانت ${result.user.username} ذخیره شد.`);
    }
    await refresh();
  }

  async function readSubscription(customer: CustomerView) {
    setBusyID(customer.id); setMessage("");
    try {
      const result = await getCurrentSubscription(customer.id);
      setDialog({ type: "subscription", customer, path: result.subscription_path, directURI: result.direct_uri, notice: result.delivery_notice, readOnly: true });
    } catch (error) { setMessage(error instanceof Error ? error.message : "خواندن Subscription انجام نشد."); }
    finally { setBusyID(""); }
  }

  async function reissue(customer: CustomerView) {
    if (!window.confirm(`لینک قبلی ${customer.username} لغو و لینک جدید صادر شود؟ Password تغییر نمی‌کند.`)) return;
    setBusyID(customer.id); setMessage("");
    try {
      const result = await rotateSubscription(customer.id);
      setDialog({ type: "subscription", customer, path: result.subscription_path, directURI: result.direct_uri, notice: result.delivery_notice, readOnly: false });
      await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "صدور Subscription انجام نشد."); }
    finally { setBusyID(""); }
  }

  async function lifecycle(customer: CustomerView, action: "suspend" | "resume" | "delete") {
    const dangerous = action === "delete";
    if (dangerous && !window.confirm(`اکانت ${customer.username} به‌صورت امن revoke شود؟ تاریخچه پاک نمی‌شود.`)) return;
    setBusyID(customer.id); setMessage("");
    try {
      if (action === "suspend") await suspendCustomer(customer.id);
      else if (action === "resume") await resumeCustomer(customer.id);
      else await deleteCustomer(customer.id);
      await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "عملیات مشتری انجام نشد."); }
    finally { setBusyID(""); }
  }

  async function bulk(action: BulkAction) {
    if (!selectedCustomers.length) return;
    const label = action === "suspend" ? "تعلیق" : action === "resume" ? "فعال‌سازی" : action === "delete" ? "لغو" : action === "add-volume" ? "افزایش حجم" : "تمدید";
    if (!window.confirm(`${label} برای ${selectedCustomers.length} مشتری اجرا شود؟ عملیات Runtime به‌ترتیب و بدون اجرای موازی انجام می‌شود.`)) return;
    let value = 0;
    if (action === "add-volume" || action === "extend") {
      const raw = window.prompt(action === "add-volume" ? "چند GB اضافه شود؟" : "چند روز تمدید شود؟", action === "add-volume" ? "10" : "30");
      if (raw === null) return;
      value = Number(raw);
      if (!Number.isFinite(value) || value <= 0) { setMessage("مقدار باید عدد مثبت باشد."); return; }
    }
    setBulkBusy(true); setMessage("");
    const failures: string[] = [];
    for (const customer of selectedCustomers) {
      try {
        if (action === "suspend") await suspendCustomer(customer.id);
        else if (action === "resume") await resumeCustomer(customer.id);
        else if (action === "delete") await deleteCustomer(customer.id);
        else if (action === "add-volume") await addCustomerVolume(customer.id, value);
        else await extendCustomerTime(customer.id, value);
      } catch (error) { failures.push(`${customer.username}: ${error instanceof Error ? error.message : "خطا"}`); }
    }
    setBulkBusy(false);
    setSelected(new Set());
    setMessage(failures.length ? `${failures.length} عملیات ناموفق: ${failures.slice(0, 3).join(" | ")}` : `${label} برای همه مشتریان انتخاب‌شده انجام شد.`);
    await refresh();
  }

  return (
    <main className="customers-page owner-customers-v2">
      <header className="owner-hero">
        <div className="owner-hero-copy"><p className="eyebrow">PVNaive · Owner Control</p><h1>مدیریت اکانت‌ها</h1><p>همه اکانت‌های قدیمی و جدید در یک نمای واحد؛ مدیریت حجم، اعتبار، QR و وضعیت بدون دستکاری ناخواسته Runtime.</p></div>
        <div className="header-actions"><button className="button-secondary refresh-action" onClick={() => void refresh()}>↻ بروزرسانی</button><button className="primary-action create-action" onClick={() => setDialog({ type: "create" })}>＋ ساخت اکانت</button></div>
      </header>

      <section className="owner-kpis" aria-label="خلاصه اکانت‌ها">
        <article className="kpi-card"><div className="kpi-icon">◎</div><div><span>کل اکانت‌ها</span><strong>{dashboard.totalAccounts}</strong><small>{dashboard.managedAccounts} اکانت مدیریت‌شده</small></div></article>
        <article className="kpi-card success"><div className="kpi-icon">✓</div><div><span>فعال</span><strong>{dashboard.active}</strong><small>{dashboard.pending} منتظر اولین اتصال</small></div></article>
        <article className="kpi-card warning"><div className="kpi-icon">⚙</div><div><span>نیازمند تنظیم</span><strong>{dashboard.needsSetup}</strong><small>همان اکانت‌های موجود Runtime</small></div></article>
        <article className="kpi-card quota"><div className="kpi-icon">◫</div><div><span>حجم تعریف‌شده</span><strong>{dashboard.configuredQuotaGB.toLocaleString("fa-IR")} <em>GB</em></strong><small>{dashboard.unlimitedAccounts} اکانت نامحدود</small></div></article>
      </section>

      <section className="owner-analytics">
        <article className="analytics-card"><div className="analytics-heading"><div><p className="eyebrow">Account health</p><h2>وضعیت اکانت‌ها</h2></div><span className="analytics-badge">Live config</span></div><StatusDonut dashboard={dashboard} /></article>
        <article className="analytics-card"><div className="analytics-heading"><div><p className="eyebrow">Expiry horizon</p><h2>افق انقضا</h2></div><span className="analytics-badge">{dashboard.managedAccounts} سرویس</span></div><ExpiryBars dashboard={dashboard} /></article>
      </section>

      {!dashboard.usageProven && <div className="accounting-note"><span className="accounting-icon">i</span><div><strong>Accounting دقیق هنوز Proof-gated است</strong><p>حجم کل و تاریخ واقعی‌اند؛ تا تکمیل accounting در Naive/Caddy، مصرف‌شده و باقی‌مانده جعلی نمایش داده نمی‌شود.</p></div></div>}

      <section className="customer-table-card owner-directory">
        <div className="directory-heading"><div><p className="eyebrow">Accounts directory</p><h2>لیست اکانت‌ها</h2><p>{visible.length} نتیجه از {dashboard.totalAccounts} اکانت</p></div></div>
        <div className="owner-toolbar">
          <label className="search-box"><span>⌕</span><input value={search} onChange={(e) => { setSearch(e.target.value); setPage(1); }} placeholder="جستجو با نام کاربری، ID یا Runtime ID" /></label>
          <label className="toolbar-field"><span>وضعیت</span><select value={filter} onChange={(e) => { setFilter(e.target.value as StatusFilter); setPage(1); }}><option value="all">همه</option><option value="active">فعال</option><option value="pending">منتظر اتصال</option><option value="suspended">تعلیق</option><option value="expired">منقضی</option><option value="depleted">حجم تمام</option><option value="revoked">لغوشده</option><option value="unmanaged">نیازمند تنظیم</option></select></label>
          <label className="toolbar-field"><span>مرتب‌سازی</span><select value={sort} onChange={(e) => setSort(e.target.value as SortKey)}><option value="username">نام کاربری</option><option value="expiry">نزدیک‌ترین انقضا</option><option value="quota">حجم</option><option value="created">شناسه</option></select></label>
          <label className="toolbar-field compact"><span>نمایش</span><select value={pageSize} onChange={(e) => { setPageSize(Number(e.target.value)); setPage(1); }}><option>25</option><option>50</option><option>100</option></select></label>
        </div>
        {selected.size > 0 && <div className="bulk-bar"><strong>{selected.size} اکانت انتخاب شده</strong><div><button disabled={bulkBusy} onClick={() => void bulk("suspend")}>تعلیق</button><button disabled={bulkBusy} onClick={() => void bulk("resume")}>فعال‌سازی</button><button disabled={bulkBusy} onClick={() => void bulk("add-volume")}>＋ حجم</button><button disabled={bulkBusy} onClick={() => void bulk("extend")}>＋ زمان</button><button className="danger-action" disabled={bulkBusy} onClick={() => void bulk("delete")}>لغو / Revoke</button></div></div>}
        {message && <p className="customer-message" role="status">{message}</p>}
        <div className="customer-table-wrap"><table className="customer-table owner-table"><thead><tr><th><input className="customer-checkbox" type="checkbox" aria-label="انتخاب صفحه" checked={selectablePageIDs.length > 0 && selectablePageIDs.every((id) => selected.has(id))} onChange={() => setSelected((current) => { const next = new Set(current); const all = selectablePageIDs.every((id) => next.has(id)); selectablePageIDs.forEach((id) => all ? next.delete(id) : next.add(id)); return next; })} /></th><th>اکانت</th><th>وضعیت</th><th>حجم</th><th>انقضا</th><th>شروع اعتبار</th><th>مصرف</th><th>عملیات</th></tr></thead><tbody>
          {loading && <tr><td colSpan={8}><div className="table-loading">در حال بارگذاری اکانت‌ها…</div></td></tr>}
          {!loading && pageRows.length === 0 && <tr><td colSpan={8}><div className="empty-state"><strong>اکانتی پیدا نشد</strong><span>فیلتر یا عبارت جستجو را تغییر بده.</span></div></td></tr>}
          {pageRows.map((row) => {
            if (row.kind === "runtime") {
              const credential = row.credential;
              return <tr key={row.id} className="setup-row"><td></td><td><div className="account-cell"><span className="account-avatar setup">{row.username.slice(0, 1).toUpperCase()}</span><div><strong>{row.username}</strong><small>Runtime · {credential.id.slice(0, 8)}</small></div></div></td><td><span className="dimension-badge warning"><i />نیازمند تنظیم</span><small className="state-detail">اکانت موجود؛ هنوز سرویس تجاری ندارد</small></td><td><span className="cell-muted">تنظیم نشده</span></td><td><span className="cell-muted">—</span></td><td><span className="cell-muted">—</span></td><td><span className="usage-locked">در انتظار Accounting</span></td><td><div className="row-actions single"><button className="primary-action" disabled={credential.status !== "active"} onClick={() => setDialog({ type: "adopt", credential })}>تنظیم سرویس</button></div></td></tr>;
            }
            const customer = row.customer; const status = statusDimension(customer); const busy = busyID === customer.id;
            return <tr key={customer.id} className={customer.status === "revoked" ? "row-muted" : ""}><td><input className="customer-checkbox" type="checkbox" checked={selected.has(customer.id)} onChange={() => toggle(customer.id)} /></td><td><div className="account-cell"><span className="account-avatar">{customer.username.slice(0, 1).toUpperCase()}</span><div><strong>{customer.username}</strong><small>{customer.id.slice(0, 8)} · {customer.runtime_credential_id.slice(0, 8)}</small></div></div></td><td><span className={`dimension-badge ${status.tone}`}><i />{status.label}</span><small className="state-detail">{customer.service_state === "pending" ? "در انتظار اولین اتصال موفق" : customer.service_state}</small></td><td><strong className="quota-cell">{quotaLabel(customer.quota_bytes)}</strong></td><td><strong className="expiry-cell">{expiryLabel(customer.expires_at)}</strong></td><td><span className="policy-cell">{startPolicyLabel(customer.start_policy)}</span></td><td>{customer.usage_capability.available ? <span className="usage-ready">آماده</span> : <span className="usage-locked">در انتظار Accounting</span>}</td><td><div className="row-actions"><button className="action-primary" disabled={busy || !customer.subscription_available} onClick={() => void readSubscription(customer)}>QR / لینک</button><button disabled={busy || customer.status === "revoked"} onClick={() => setDialog({ type: "edit", customer })}>ویرایش</button><button disabled={busy} onClick={() => setDialog({ type: "details", customer })}>جزئیات</button><details className="row-more"><summary aria-label="عملیات بیشتر">•••</summary><div className="more-menu">{customer.status === "suspended" ? <button disabled={busy} onClick={() => void lifecycle(customer, "resume")}>فعال‌سازی مجدد</button> : customer.status !== "revoked" && <button disabled={busy} onClick={() => void lifecycle(customer, "suspend")}>تعلیق اکانت</button>}{customer.status !== "revoked" && <><button onClick={() => setDialog({ type: "password", customer })}>تغییر رمز</button><button onClick={() => setDialog({ type: "add-volume", customer })}>افزودن حجم</button><button onClick={() => setDialog({ type: "extend", customer })}>تمدید زمان</button><button onClick={() => void reissue(customer)}>صدور لینک جدید</button><button className="danger-action" onClick={() => void lifecycle(customer, "delete")}>حذف امن / Revoke</button></>}</div></details></div></td></tr>;
          })}
        </tbody></table></div>
        <div className="pagination"><span>صفحه {page.toLocaleString("fa-IR")} از {pageCount.toLocaleString("fa-IR")}</span><div><button disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>‹ قبلی</button><strong>{visible.length.toLocaleString("fa-IR")} نتیجه</strong><button disabled={page >= pageCount} onClick={() => setPage((p) => p + 1)}>بعدی ›</button></div></div>
      </section>

      {dialog?.type === "create" && <Dialog eyebrow="Create customer" title="ساخت اکانت" onClose={() => setDialog(null)}><ServiceForm mode="create" onSaved={saved} onClose={() => setDialog(null)} /></Dialog>}
      {dialog?.type === "adopt" && <Dialog eyebrow="Configure account" title={`مدیریت ${dialog.credential.username}`} onClose={() => setDialog(null)}><ServiceForm mode="adopt" credential={dialog.credential} onSaved={saved} onClose={() => setDialog(null)} /></Dialog>}
      {dialog?.type === "edit" && <Dialog eyebrow="Edit service" title={`ویرایش ${dialog.customer.username}`} onClose={() => setDialog(null)}><ServiceForm mode="edit" customer={dialog.customer} onSaved={saved} onClose={() => setDialog(null)} /></Dialog>}
      {dialog?.type === "details" && <Dialog eyebrow="Customer details" title={dialog.customer.username} onClose={() => setDialog(null)}><div className="detail-list"><div><span>Customer ID</span><strong>{dialog.customer.id}</strong></div><div><span>Runtime ID</span><strong>{dialog.customer.runtime_credential_id}</strong></div><div><span>Account</span><strong>{dialog.customer.status}</strong></div><div><span>Service</span><strong>{dialog.customer.service_state}</strong></div><div><span>Volume</span><strong>{quotaLabel(dialog.customer.quota_bytes)}</strong></div><div><span>Expiry</span><strong>{expiryLabel(dialog.customer.expires_at)}</strong></div><div><span>First connected</span><strong>{dialog.customer.first_connected_at ? expiryLabel(dialog.customer.first_connected_at) : "ثبت نشده"}</strong></div><div><span>Accounting</span><strong>{dialog.customer.usage_capability.available ? "available" : "exact_accounting_not_proven"}</strong></div></div></Dialog>}
      {dialog?.type === "password" && <Dialog eyebrow="Explicit security action" title={`تغییر رمز ${dialog.customer.username}`} onClose={() => setDialog(null)}><PasswordForm customer={dialog.customer} onClose={() => setDialog(null)} onDone={async (password) => { await refresh(); if (password) setDialog({ type: "secret", username: dialog.customer.username, password, notice: "Subscription تغییر نکرد. رمز جدید فقط همین بار نمایش داده می‌شود." }); else setMessage("رمز تغییر کرد؛ Subscription همان لینک قبلی است."); }} /></Dialog>}
      {dialog?.type === "add-volume" && <Dialog eyebrow="Quota adjustment" title={`افزودن حجم ${dialog.customer.username}`} onClose={() => setDialog(null)}><NumberActionForm label="حجم افزایشی" suffix="GB" initial={10} onClose={() => setDialog(null)} onSubmit={async (value) => { const result = await addCustomerVolume(dialog.customer.id, value); if (result.runtime_mutated) throw new Error("Volume adjustment unexpectedly mutated Runtime"); await refresh(); setMessage(`${value}GB به حجم ${dialog.customer.username} اضافه شد.`); }} /></Dialog>}
      {dialog?.type === "extend" && <Dialog eyebrow="Validity adjustment" title={`تمدید ${dialog.customer.username}`} onClose={() => setDialog(null)}><NumberActionForm label="مدت افزایشی" suffix="روز" initial={30} onClose={() => setDialog(null)} onSubmit={async (value) => { const result = await extendCustomerTime(dialog.customer.id, value); if (result.runtime_mutated) throw new Error("Validity adjustment unexpectedly mutated Runtime"); await refresh(); setMessage(`${value} روز به اعتبار ${dialog.customer.username} اضافه شد.`); }} /></Dialog>}
      {dialog?.type === "subscription" && <Dialog eyebrow={dialog.readOnly ? "Read-only delivery" : "Security reissue"} title={`Subscription ${dialog.customer.username}`} onClose={() => setDialog(null)} wide><SubscriptionView customer={dialog.customer} path={dialog.path} directURI={dialog.directURI} notice={dialog.notice} readOnly={dialog.readOnly} /></Dialog>}
      {dialog?.type === "secret" && <Dialog eyebrow="One-time secret" title={`رمز ${dialog.username}`} onClose={() => setDialog(null)}><p className="capability-warning">{dialog.notice}</p><div className="copy-row"><input readOnly value={dialog.password} /><button onClick={() => void navigator.clipboard.writeText(dialog.password)}>کپی</button></div></Dialog>}
    </main>
  );
}
