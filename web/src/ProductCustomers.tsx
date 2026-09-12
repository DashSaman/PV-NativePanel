import { FormEvent, ReactNode, useCallback, useEffect, useMemo, useState } from "react";
import type { Principal } from "./auth";
import {
  createProductCustomer,
  executeProductBulk,
  getProductSubscription,
  listProductCustomers,
  listProductCustomerSessions,
  killProductCustomerSessionAndReload,
  listProductGroups,
  listProductPlans,
  listProductTags,
  previewProductBulk,
  reissueProductSubscription,
  renewProductCustomer,
  resetProductUsage,
  resumeProductCustomer,
  revokeProductCustomer,
  rotateProductPassword,
  suspendProductCustomer,
  updateProductCustomer,
  DEFAULT_PRODUCT_FILTERS,
  type ProductActiveSession,
  type ProductBulkAction,
  type ProductBulkOperation,
  type ProductBulkRequest,
  type ProductCustomer,
  type ProductFilters,
  type ProductGroup,
  type ProductPlan,
  type ProductRenewalInput,
  type ProductSubscriptionDelivery,
  type ProductTag,
} from "./productApi";
import { encodeQR } from "./qr";
import { formatBytes, formatPanelDate, usagePresentation } from "./productPanelModel";
import "./product-panel.css";

type Props = { role: Principal["role"] };

type DialogState =
  | { type: "create" }
  | { type: "metadata"; customer: ProductCustomer }
  | { type: "renew"; customer: ProductCustomer }
  | { type: "password"; customer: ProductCustomer }
  | { type: "subscription"; customer: ProductCustomer; delivery: ProductSubscriptionDelivery; rotated: boolean }
  | { type: "secret"; username: string; password?: string; subscriptionPath?: string; accountPagePath?: string; notice: string }
  | { type: "sessions"; customer: ProductCustomer; sessions: ProductActiveSession[]; observedAt: string; loading: boolean; error: string }
  | { type: "bulk" };

type BulkPreviewState = { request: ProductBulkRequest; operation: ProductBulkOperation; idempotencyKey: string };

function Modal({ title, eyebrow, onClose, children, wide = false }: { title: string; eyebrow: string; onClose: () => void; children: ReactNode; wide?: boolean }) {
  return <div className="product-modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
    <section className={`product-modal ${wide ? "wide" : ""}`} role="dialog" aria-modal="true">
      <div className="product-modal-head"><div><p className="eyebrow">{eyebrow}</p><h2>{title}</h2></div><button className="icon-button" onClick={onClose} aria-label="بستن">×</button></div>
      {children}
    </section>
  </div>;
}

function QR({ value }: { value: string }) {
  const matrix = useMemo(() => {
    try { return encodeQR(value); } catch { return null; }
  }, [value]);
  if (!matrix) return <div className="qr-unavailable">QR برای این مقدار ساخته نشد؛ لینک را کپی کن.</div>;
  const quiet = 4;
  const size = matrix.length + quiet * 2;
  return <svg className="product-qr" viewBox={`0 0 ${size} ${size}`} role="img" aria-label="QR اشتراک" shapeRendering="crispEdges">
    <rect width={size} height={size} fill="white" />
    {matrix.flatMap((row, y) => row.map((dark, x) => dark ? <rect key={`${x}-${y}`} x={x + quiet} y={y + quiet} width="1" height="1" fill="black" /> : null))}
  </svg>;
}

function absoluteSubscription(path: string): string {
  const base = typeof window === "undefined" ? "https://localhost/" : window.location.origin;
  return new URL(path, base).toString();
}

function humanAccountPath(subscriptionPath: string, explicitPath?: string): string {
  if (explicitPath) return explicitPath;
  const prefixes = ["/sub/", "/api/v1/subscriptions/"];
  for (const prefix of prefixes) {
    if (subscriptionPath.startsWith(prefix)) {
      const token = subscriptionPath.slice(prefix.length).split("/", 1)[0];
      if (token) return `/s/${token}`;
    }
  }
  return subscriptionPath;
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (m < 60) return s > 0 ? `${m}m ${s}s` : `${m}m`;
  const h = Math.floor(m / 60);
  const rm = m % 60;
  return rm > 0 ? `${h}h ${rm}m` : `${h}h`;
}

function commercialLabel(customer: ProductCustomer): string {
  const value = customer.status_dimensions?.commercial;
  if (value === "pending_first_use") return "منتظر اولین اتصال";
  if (value === "expired") return "منقضی";
  if (value === "depleted") return "حجم تمام";
  if (value === "on_hold") return "On hold";
  if (customer.status_dimensions?.lifecycle === "suspended") return "تعلیق";
  if (customer.status_dimensions?.lifecycle === "revoked") return "لغوشده";
  return "فعال";
}

function statusTone(customer: ProductCustomer): string {
  const label = commercialLabel(customer);
  if (label === "فعال") return "success";
  if (label === "منتظر اولین اتصال" || label === "On hold") return "warning";
  return "danger";
}

function CustomServiceFields({ quotaUnlimited, setQuotaUnlimited, quotaGB, setQuotaGB, noExpiry, setNoExpiry, validityMode, setValidityMode, days, setDays, fixedExpiry, setFixedExpiry }: {
  quotaUnlimited: boolean; setQuotaUnlimited: (value: boolean) => void; quotaGB: string; setQuotaGB: (value: string) => void;
  noExpiry: boolean; setNoExpiry: (value: boolean) => void;
  validityMode: "on_creation" | "on_first_successful_connection" | "fixed_expiry"; setValidityMode: (value: "on_creation" | "on_first_successful_connection" | "fixed_expiry") => void;
  days: string; setDays: (value: string) => void; fixedExpiry: string; setFixedExpiry: (value: string) => void;
}) {
  return <>
    <fieldset><legend>حجم</legend><label className="check-line"><input type="checkbox" checked={quotaUnlimited} onChange={(e) => setQuotaUnlimited(e.target.checked)} /><span>حجم نامحدود</span></label>{!quotaUnlimited && <label>حجم کل (GB)<input type="number" min="1" step="1" value={quotaGB} onChange={(e) => setQuotaGB(e.target.value)} required /></label>}</fieldset>
    <fieldset><legend>اعتبار</legend><label className="check-line"><input type="checkbox" checked={noExpiry} onChange={(e) => setNoExpiry(e.target.checked)} /><span>بدون انقضا</span></label>{!noExpiry && <><label>نوع شروع<select value={validityMode} onChange={(e) => setValidityMode(e.target.value as typeof validityMode)}><option value="on_creation">از همین حالا</option><option value="on_first_successful_connection">از اولین اتصال موفق</option><option value="fixed_expiry">تاریخ پایان دستی</option></select></label>{validityMode === "fixed_expiry" ? <label>تاریخ پایان<input type="datetime-local" value={fixedExpiry} onChange={(e) => setFixedExpiry(e.target.value)} required /></label> : <label>تعداد روز<input type="number" min="1" step="1" value={days} onChange={(e) => setDays(e.target.value)} required /></label>}</>}</fieldset>
  </>;
}

function createValidity(noExpiry: boolean, validityMode: "on_creation" | "on_first_successful_connection" | "fixed_expiry", days: string, fixedExpiry: string) {
  if (noExpiry) return undefined;
  if (validityMode === "fixed_expiry") return { mode: validityMode, expires_at: new Date(fixedExpiry).toISOString() } as const;
  return { mode: validityMode, duration_days: Number(days) } as const;
}

function CreateForm({ plans, groups, tags, onDone, onClose }: { plans: ProductPlan[]; groups: ProductGroup[]; tags: ProductTag[]; onDone: (result: { username: string; password?: string; subscriptionPath?: string; accountPagePath?: string }) => Promise<void>; onClose: () => void }) {
  const [username, setUsername] = useState("");
  const [preset, setPreset] = useState("");
  const [generate, setGenerate] = useState(true);
  const [password, setPassword] = useState("");
  const [quotaUnlimited, setQuotaUnlimited] = useState(false);
  const [quotaGB, setQuotaGB] = useState("50");
  const [noExpiry, setNoExpiry] = useState(false);
  const [validityMode, setValidityMode] = useState<"on_creation" | "on_first_successful_connection" | "fixed_expiry">("on_creation");
  const [days, setDays] = useState("30");
  const [fixedExpiry, setFixedExpiry] = useState("");
  const [groupID, setGroupID] = useState("");
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set());
  const [onHold, setOnHold] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      const result = await createProductCustomer({
        username: username.trim(), generate_password: generate, ...(generate ? {} : { password }),
        ...(preset ? { plan_id: preset } : {
          ...(quotaUnlimited ? { unlimited_quota: true } : { quota_gb: Number(quotaGB) }),
          ...(noExpiry ? { no_expiry: true } : {}),
          validity: createValidity(noExpiry, validityMode, days, fixedExpiry),
        }),
        ...(groupID ? { group_id: groupID } : {}), tag_ids: Array.from(selectedTags), on_hold: onHold,
      });
      await onDone({ username: result.user.username, password: result.generated_password, subscriptionPath: result.subscription_path, accountPagePath: result.account_page_path });
    } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت اکانت انجام نشد."); }
    finally { setBusy(false); }
  }

  return <form className="product-form" onSubmit={submit}>
    <label>نام کاربری<input value={username} onChange={(e) => setUsername(e.target.value)} autoComplete="off" required /></label>
    <label>پلن<select value={preset} onChange={(e) => setPreset(e.target.value)}><option value="">Custom / دستی</option>{plans.filter((plan) => plan.enabled).map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}</select></label>
    {!preset && <CustomServiceFields {...{ quotaUnlimited, setQuotaUnlimited, quotaGB, setQuotaGB, noExpiry, setNoExpiry, validityMode, setValidityMode, days, setDays, fixedExpiry, setFixedExpiry }} />}
    <fieldset><legend>رمز عبور</legend><label className="check-line"><input type="checkbox" checked={generate} onChange={(e) => setGenerate(e.target.checked)} /><span>تولید رمز امن خودکار</span></label>{!generate && <label>رمز سفارشی<input type="password" minLength={12} value={password} onChange={(e) => setPassword(e.target.value)} required /></label>}</fieldset>
    <label>گروه<select value={groupID} onChange={(e) => setGroupID(e.target.value)}><option value="">بدون گروه</option>{groups.filter((group) => group.enabled).map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}</select></label>
    <fieldset><legend>تگ‌ها</legend><div className="chip-checks">{tags.filter((tag) => tag.enabled).map((tag) => <label key={tag.id}><input type="checkbox" checked={selectedTags.has(tag.id)} onChange={() => setSelectedTags((current) => { const next = new Set(current); next.has(tag.id) ? next.delete(tag.id) : next.add(tag.id); return next; })} /><span>{tag.name}</span></label>)}</div></fieldset>
    <label className="check-line"><input type="checkbox" checked={onHold} onChange={(e) => setOnHold(e.target.checked)} /><span>ساخت در وضعیت On hold</span></label>
    {message && <div className="product-message danger">{message}</div>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال ساخت…" : "ساخت اکانت"}</button></div>
  </form>;
}

function MetadataForm({ customer, groups, tags, plans, onDone, onClose }: { customer: ProductCustomer; groups: ProductGroup[]; tags: ProductTag[]; plans: ProductPlan[]; onDone: () => Promise<void>; onClose: () => void }) {
  const [note, setNote] = useState(customer.note || "");
  const [groupID, setGroupID] = useState(customer.group?.id || "");
  const [nextPlan, setNextPlan] = useState(customer.next_plan_id || "");
  const [onHold, setOnHold] = useState(customer.on_hold);
  const [selectedTags, setSelectedTags] = useState<Set<string>>(new Set((customer.tags || []).map((tag) => tag.id)));
  const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");
  const initialTags = useMemo(() => new Set((customer.tags || []).map((tag) => tag.id)), [customer.tags]);

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    const add = Array.from(selectedTags).filter((id) => !initialTags.has(id));
    const remove = Array.from(initialTags).filter((id) => !selectedTags.has(id));
    try {
      await updateProductCustomer(customer.id, { note, group_id: groupID, on_hold: onHold, add_tag_ids: add, remove_tag_ids: remove, next_plan_id: nextPlan });
      await onDone(); onClose();
    } catch (error) { setMessage(error instanceof Error ? error.message : "ویرایش انجام نشد."); }
    finally { setBusy(false); }
  }
  return <form className="product-form" onSubmit={submit}>
    <label>یادداشت<textarea value={note} onChange={(e) => setNote(e.target.value)} rows={3} /></label>
    <label>گروه<select value={groupID} onChange={(e) => setGroupID(e.target.value)}><option value="">بدون گروه</option>{groups.filter((group) => group.enabled).map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}</select></label>
    <label>Next Plan<select value={nextPlan} onChange={(e) => setNextPlan(e.target.value)}><option value="">بدون Next Plan</option>{plans.filter((plan) => plan.enabled).map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}</select></label>
    <label className="check-line"><input type="checkbox" checked={onHold} onChange={(e) => setOnHold(e.target.checked)} /><span>On hold</span></label>
    <fieldset><legend>تگ‌ها</legend><div className="chip-checks">{tags.filter((tag) => tag.enabled).map((tag) => <label key={tag.id}><input type="checkbox" checked={selectedTags.has(tag.id)} onChange={() => setSelectedTags((current) => { const next = new Set(current); next.has(tag.id) ? next.delete(tag.id) : next.add(tag.id); return next; })} /><span>{tag.name}</span></label>)}</div></fieldset>
    {message && <div className="product-message danger">{message}</div>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال ذخیره…" : "ذخیره مشخصات"}</button></div>
  </form>;
}

function RenewalForm({ customer, plans, onDone, onClose }: { customer: ProductCustomer; plans: ProductPlan[]; onDone: () => Promise<void>; onClose: () => void }) {
  const [mode, setMode] = useState<"renew_current" | "renew_plan" | "next_plan" | "custom">(customer.next_plan_id ? "next_plan" : customer.plan_id ? "renew_current" : "custom");
  const [planID, setPlanID] = useState(customer.plan_id || "");
  const [quotaUnlimited, setQuotaUnlimited] = useState(customer.quota_bytes === null);
  const [quotaGB, setQuotaGB] = useState(customer.quota_bytes ? String(Math.max(1, Math.round(customer.quota_bytes / 1073741824))) : "50");
  const [noExpiry, setNoExpiry] = useState(customer.no_expiry);
  const [validityMode, setValidityMode] = useState<"on_creation" | "on_first_successful_connection" | "fixed_expiry">(customer.start_policy === "on_first_successful_connection" ? "on_first_successful_connection" : "on_creation");
  const [days, setDays] = useState(String(Math.max(1, Math.round((customer.duration_seconds || 30 * 86400) / 86400))));
  const [fixedExpiry, setFixedExpiry] = useState("");
  const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    let input: ProductRenewalInput;
    if (mode === "renew_current") input = { mode };
    else if (mode === "next_plan") input = { mode };
    else if (mode === "renew_plan") input = { mode, plan_id: planID };
    else input = { mode, ...(quotaUnlimited ? { unlimited_quota: true } : { quota_gb: Number(quotaGB) }), ...(noExpiry ? { no_expiry: true } : {}), validity: createValidity(noExpiry, validityMode, days, fixedExpiry) };
    try { await renewProductCustomer(customer.id, input); await onDone(); onClose(); }
    catch (error) { setMessage(error instanceof Error ? error.message : "تمدید انجام نشد."); }
    finally { setBusy(false); }
  }

  return <form className="product-form" onSubmit={submit}>
    <fieldset><legend>روش تمدید</legend><div className="segmented wrap"><button type="button" className={mode === "renew_current" ? "active" : ""} disabled={!customer.plan_id} onClick={() => setMode("renew_current")}>پلن فعلی</button><button type="button" className={mode === "renew_plan" ? "active" : ""} onClick={() => setMode("renew_plan")}>پلن دیگر</button><button type="button" className={mode === "next_plan" ? "active" : ""} disabled={!customer.next_plan_id} onClick={() => setMode("next_plan")}>Next Plan</button><button type="button" className={mode === "custom" ? "active" : ""} onClick={() => setMode("custom")}>Custom</button></div></fieldset>
    {mode === "renew_plan" && <label>پلن<select value={planID} onChange={(e) => setPlanID(e.target.value)} required><option value="">انتخاب پلن</option>{plans.filter((plan) => plan.enabled).map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}</select></label>}
    {mode === "next_plan" && <div className="readonly-banner">Next Plan: <strong>{customer.next_plan_name || customer.next_plan_id}</strong></div>}
    {mode === "custom" && <CustomServiceFields {...{ quotaUnlimited, setQuotaUnlimited, quotaGB, setQuotaGB, noExpiry, setNoExpiry, validityMode, setValidityMode, days, setDays, fixedExpiry, setFixedExpiry }} />}
    {message && <div className="product-message danger">{message}</div>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال تمدید…" : "تمدید سرویس"}</button></div>
  </form>;
}

function PasswordForm({ customer, onDone, onClose }: { customer: ProductCustomer; onDone: (password?: string) => Promise<void>; onClose: () => void }) {
  const [generate, setGenerate] = useState(true); const [password, setPassword] = useState(""); const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");
  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try { const result = await rotateProductPassword(customer.id, { password: generate ? "" : password, generate_password: generate }); await onDone(result.generated_password || (generate ? undefined : password)); }
    catch (error) { setMessage(error instanceof Error ? error.message : "تغییر رمز انجام نشد."); }
    finally { setBusy(false); }
  }
  return <form className="product-form" onSubmit={submit}><div className="readonly-banner">تغییر رمز از Subscription reissue جدا است.</div><label className="check-line"><input type="checkbox" checked={generate} onChange={(e) => setGenerate(e.target.checked)} /><span>رمز امن خودکار</span></label>{!generate && <label>رمز جدید<input type="password" minLength={12} value={password} onChange={(e) => setPassword(e.target.value)} required /></label>}{message && <div className="product-message danger">{message}</div>}<div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>تغییر رمز</button></div></form>;
}

function SubscriptionContent({ delivery }: { delivery: ProductSubscriptionDelivery }) {
  const subscription = absoluteSubscription(delivery.subscription_path);
  const accountPage = absoluteSubscription(humanAccountPath(delivery.subscription_path, delivery.account_page_path));
  const [copied, setCopied] = useState("");
  async function copy(key: string, value: string) { await navigator.clipboard.writeText(value); setCopied(key); window.setTimeout(() => setCopied(""), 1000); }
  return <div className="subscription-delivery">
    <section className="subscription-primary-card">
      <div className="subscription-primary-copy">
        <span className="subscription-kicker">Karing / Subscription</span>
        <h3>لینک ساب کلاینت</h3>
        <p>برای اضافه‌کردن سرویس در Karing همین QR را اسکن کن یا لینک را کپی کن.</p>
        <div className="copy-row subscription-link-row"><input readOnly value={subscription} /><button className="primary-action" onClick={() => void copy("sub", subscription)}>{copied === "sub" ? "کپی شد ✓" : "کپی لینک"}</button></div>
      </div>
      <div className="subscription-primary-qr"><QR value={subscription} /><small>QR اشتراک Karing</small></div>
    </section>

    <section className="account-page-box">
      <div><strong>صفحه وضعیت کاربر</strong><span>نمایش حجم، انقضا و راهنمای اتصال در مرورگر.</span></div>
      <div className="account-page-actions"><a className="open-account-page" href={accountPage} target="_blank" rel="noreferrer">باز کردن صفحه</a><button className="button-secondary" onClick={() => void copy("page", accountPage)}>{copied === "page" ? "کپی شد ✓" : "کپی لینک صفحه"}</button></div>
    </section>

    {delivery.direct_uri && <details className="direct-details"><summary>Direct Naive · تنظیمات پیشرفته</summary><div className="direct-details-body"><div className="subscription-fields"><p className="field-hint">برای ورود دستی یا تست مستقیم استفاده می‌شود؛ برای Karing معمولاً Subscription پیشنهاد می‌شود.</p><div className="copy-row"><input readOnly value={delivery.direct_uri} /><button onClick={() => void copy("direct", delivery.direct_uri!)}>{copied === "direct" ? "کپی شد ✓" : "کپی"}</button></div></div><div className="direct-qr"><QR value={delivery.direct_uri} /><small>QR مستقیم Naive</small></div></div></details>}
    {delivery.delivery_notice && <p className="subscription-notice">{delivery.delivery_notice}</p>}
  </div>;
}

function BulkForm({ selected, plans, groups, tags, onDone, onClose }: { selected: string[]; plans: ProductPlan[]; groups: ProductGroup[]; tags: ProductTag[]; onDone: () => Promise<void>; onClose: () => void }) {
  const [action, setAction] = useState<ProductBulkAction>("suspend");
  const [value, setValue] = useState("10");
  const [planID, setPlanID] = useState(""); const [groupID, setGroupID] = useState(""); const [tagID, setTagID] = useState("");
  const [preview, setPreview] = useState<BulkPreviewState | null>(null);
  const [busy, setBusy] = useState(false); const [message, setMessage] = useState("");

  function request(): ProductBulkRequest {
    const base: ProductBulkRequest = { action, customer_ids: selected };
    if (action === "add_volume") base.volume_gb = Number(value);
    if (action === "extend_days") base.days = Number(value);
    if (action === "set_volume") base.total_volume_gb = Number(value);
    if (action === "apply_plan") base.plan_id = planID;
    if (action === "assign_group") base.group_id = groupID;
    if (action === "add_tag" || action === "remove_tag") base.tag_id = tagID;
    return base;
  }
  async function makePreview(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try { const req = request(); const result = await previewProductBulk(req); setPreview({ request: req, operation: result.bulk, idempotencyKey: result.idempotencyKey }); }
    catch (error) { setMessage(error instanceof Error ? error.message : "Preview انجام نشد."); }
    finally { setBusy(false); }
  }
  async function execute() {
    if (!preview) return; setBusy(true); setMessage("");
    try { const result = await executeProductBulk(preview.request, preview.idempotencyKey); setPreview({ ...preview, operation: result }); if (result.status === "executed") await onDone(); }
    catch (error) { setMessage(error instanceof Error ? error.message : "Bulk execute انجام نشد."); }
    finally { setBusy(false); }
  }

  const needsNumber = action === "add_volume" || action === "extend_days" || action === "set_volume";
  return <form className="product-form" onSubmit={makePreview}>
    <div className="readonly-banner"><strong>{selected.length.toLocaleString("fa-IR")}</strong> اکانت انتخاب شده است. هیچ تغییری قبل از Preview و تأیید نهایی اجرا نمی‌شود.</div>
    <label>عملیات<select value={action} onChange={(e) => { setAction(e.target.value as ProductBulkAction); setPreview(null); }}><option value="suspend">تعلیق</option><option value="enable">فعال‌سازی</option><option value="revoke">Revoke</option><option value="add_volume">افزایش حجم</option><option value="set_volume">Set Volume</option><option value="extend_days">تمدید روز</option><option value="apply_plan">اعمال پلن</option><option value="assign_group">تغییر گروه</option><option value="add_tag">افزودن تگ</option><option value="remove_tag">حذف تگ</option><option value="reissue_subscription">Reissue Subscription</option><option value="reset_usage">Reset مصرف</option></select></label>
    {needsNumber && <label>{action === "extend_days" ? "روز" : "GB"}<input type="number" min="1" value={value} onChange={(e) => setValue(e.target.value)} required /></label>}
    {action === "apply_plan" && <label>پلن<select value={planID} onChange={(e) => setPlanID(e.target.value)} required><option value="">انتخاب پلن</option>{plans.filter((plan) => plan.enabled).map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}</select></label>}
    {action === "assign_group" && <label>گروه<select value={groupID} onChange={(e) => setGroupID(e.target.value)} required><option value="">انتخاب گروه</option>{groups.filter((group) => group.enabled).map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}</select></label>}
    {(action === "add_tag" || action === "remove_tag") && <label>تگ<select value={tagID} onChange={(e) => setTagID(e.target.value)} required><option value="">انتخاب تگ</option>{tags.filter((tag) => tag.enabled).map((tag) => <option key={tag.id} value={tag.id}>{tag.name}</option>)}</select></label>}
    {action === "reset_usage" && <div className="product-warning"><strong>Reset مصرف هر کاربر جداگانه و فقط پس از Preview اجرا می‌شود.</strong><span>Password و Subscription تغییر نمی‌کنند؛ حساب‌های ناامن/غیرقابل Reset به‌صورت per-item ناموفق می‌شوند و سایر موارد ادامه پیدا می‌کنند.</span></div>}
    {preview && <div className="bulk-preview"><div className="bulk-numbers"><span>درخواست <strong>{preview.operation.preview.requested}</strong></span><span>قابل اجرا <strong>{preview.operation.preview.affected}</strong></span><span>Conflict <strong>{preview.operation.preview.conflicts.length}</strong></span><span>Skipped <strong>{preview.operation.preview.skipped.length}</strong></span></div>{preview.operation.preview.changes.length > 0 && <ul>{preview.operation.preview.changes.map((change) => <li key={change}>{change}</li>)}</ul>}{preview.operation.preview.conflicts.length > 0 && <div className="product-warning"><strong>Conflictها</strong><span>{preview.operation.preview.conflicts.slice(0, 5).map((item) => `${item.id}: ${item.reason}`).join(" | ")}</span></div>}{preview.operation.result && <div className="product-message">نتیجه: {preview.operation.result.succeeded} موفق · {preview.operation.result.failed} ناموفق · {preview.operation.result.skipped} ردشده</div>}</div>}
    {message && <div className="product-message danger">{message}</div>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>بستن</button><button disabled={busy} type="submit">Preview</button>{preview && preview.operation.status !== "executed" && <button type="button" className="primary-action" disabled={busy || preview.operation.preview.affected === 0} onClick={() => void execute()}>تأیید و اجرا</button>}</div>
  </form>;
}

function SessionsModal({ dialog, onClose, onKill, busySessionID }: { dialog: Extract<DialogState, { type: "sessions" }>; onClose: () => void; onKill: (session: ProductActiveSession) => void; busySessionID: string }) {
  const { customer, sessions, observedAt, loading, error } = dialog;
  return <Modal title={`نشست‌های فعال · ${customer.username}`} eyebrow="مدیریت نشست" onClose={onClose}>
    {loading && <div className="table-loading">در حال بارگذاری نشست‌ها…</div>}
    {!loading && error && <div className="product-message danger">{error}</div>}
    {!loading && !error && sessions.length === 0 && <div className="empty-state"><strong>نشست فعالی وجود ندارد</strong><span>کاربر در حال حاضر اتصال فعال Naive ندارد. نشست‌های منقضی یا ناقص نمایش داده نمی‌شوند.</span></div>}
    {!loading && !error && sessions.length > 0 && <>
      <div className="readonly-banner">{sessions.length} نشست فعال · زمان بروزرسانی: {new Date(observedAt).toLocaleString("fa-IR")}</div>
      <div className="product-table-wrap"><table className="product-table compact-table"><thead><tr>
        <th>آدرس IP</th><th>گره</th><th>متصل شده</th><th>آخرین فعالیت</th><th>مدت</th><th>آپلود</th><th>دانلود</th><th>عملیات</th>
      </tr></thead><tbody>
        {sessions.map((session) => <tr key={session.session_id}>
          <td><code>{session.client_ip}</code></td>
          <td>{session.node_id}</td>
          <td>{new Date(session.connected_at).toLocaleString("fa-IR")}</td>
          <td>{new Date(session.last_activity_at).toLocaleString("fa-IR")}</td>
          <td>{formatDuration(session.duration_seconds)}</td>
          <td>{formatBytes(session.upload_bytes)}</td>
          <td>{formatBytes(session.download_bytes)}</td>
          <td><button type="button" className="danger-action" disabled={busySessionID === session.session_id} onClick={() => onKill(session)}>{busySessionID === session.session_id ? "در حال قطع…" : "قطع نشست"}</button></td>
        </tr>)}
      </tbody></table></div>
    </>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>بستن</button></div>
  </Modal>;
}

export function ProductCustomers({ role: _role }: Props) {
  const [filters, setFilters] = useState<ProductFilters>({ ...DEFAULT_PRODUCT_FILTERS });
  const [searchDraft, setSearchDraft] = useState("");
  const [customers, setCustomers] = useState<ProductCustomer[]>([]);
  const [total, setTotal] = useState(0);
  const [plans, setPlans] = useState<ProductPlan[]>([]); const [groups, setGroups] = useState<ProductGroup[]>([]); const [tags, setTags] = useState<ProductTag[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [dialog, setDialog] = useState<DialogState | null>(null);
  const [loading, setLoading] = useState(true); const [busyID, setBusyID] = useState(""); const [sessionBusyID, setSessionBusyID] = useState(""); const [message, setMessage] = useState("");

  const refreshCatalog = useCallback(async () => {
    const [nextPlans, nextGroups, nextTags] = await Promise.all([listProductPlans(), listProductGroups(), listProductTags()]);
    setPlans(nextPlans); setGroups(nextGroups); setTags(nextTags);
  }, []);

  const refresh = useCallback(async () => {
    setLoading(true); setMessage("");
    try { const page = await listProductCustomers(filters); setCustomers(page.customers); setTotal(page.total); }
    catch (error) { setMessage(error instanceof Error ? error.message : "لیست کاربران بارگذاری نشد."); }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { void refreshCatalog().catch((error) => setMessage(error instanceof Error ? error.message : "اطلاعات پلن‌ها بارگذاری نشد.")); }, [refreshCatalog]);
  useEffect(() => { void refresh(); }, [refresh]);
  useEffect(() => { setSelected(new Set()); }, [filters.page, filters.pageSize]);

  const pageCount = Math.max(1, Math.ceil(total / filters.pageSize));
  const allPageSelected = customers.length > 0 && customers.every((customer) => selected.has(customer.id));
  const onlineCount = customers.filter((customer) => customer.usage?.available && customer.usage.online).length;
  const activeCount = customers.filter((customer) => commercialLabel(customer) === "فعال").length;

  function patchFilters(patch: Partial<ProductFilters>) { setFilters((current) => ({ ...current, ...patch, page: patch.page ?? 1 })); }
  function resetFilters() { setSearchDraft(""); setFilters({ ...DEFAULT_PRODUCT_FILTERS }); }
  function toggle(id: string) { setSelected((current) => { const next = new Set(current); next.has(id) ? next.delete(id) : next.add(id); return next; }); }
  async function copiedSecret(username: string, password?: string, subscriptionPath?: string, accountPagePath?: string, notice = "اطلاعات حساس فقط همین بار نمایش داده می‌شود.") { setDialog({ type: "secret", username, password, subscriptionPath, accountPagePath, notice }); await refresh(); }

  async function openSubscription(customer: ProductCustomer, rotate = false) {
    setBusyID(customer.id); setMessage("");
    try {
      if (rotate && !window.confirm(`لینک اشتراک قبلی ${customer.username} باطل و لینک جدید صادر شود؟`)) return;
      const delivery = rotate ? await reissueProductSubscription(customer.id) : await getProductSubscription(customer.id);
      setDialog({ type: "subscription", customer, delivery, rotated: rotate });
      if (rotate) await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "لینک اشتراک در دسترس نیست."); }
    finally { setBusyID(""); }
  }

  async function resetUsage(customer: ProductCustomer) {
    if (!window.confirm(`مصرف فعلی ${customer.username} صفر شود؟ Password و Subscription تغییر نمی‌کنند.`)) return;
    setBusyID(customer.id); setMessage("");
    try {
      const result = await resetProductUsage(customer.id);
      await refresh();
      setMessage(result.idempotent_replay ? `Reset مصرف ${customer.username} قبلاً با همین درخواست ثبت شده بود.` : `مصرف ${customer.username} صفر شد؛ Password و Subscription تغییر نکردند.`);
    } catch (error) { setMessage(error instanceof Error ? error.message : "Reset مصرف انجام نشد."); }
    finally { setBusyID(""); }
  }

  async function lifecycle(customer: ProductCustomer, action: "suspend" | "resume" | "revoke") {
    if (action === "revoke" && !window.confirm(`اکانت ${customer.username} لغو شود؟`)) return;
    setBusyID(customer.id); setMessage("");
    try {
      if (action === "suspend") await suspendProductCustomer(customer.id);
      else if (action === "resume") await resumeProductCustomer(customer.id);
      else await revokeProductCustomer(customer.id);
      await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "عملیات انجام نشد."); }
    finally { setBusyID(""); }
  }

  async function openSessions(customer: ProductCustomer) {
    setBusyID(customer.id); setMessage("");
    setDialog({ type: "sessions", customer, sessions: [], observedAt: "", loading: true, error: "" });
    try {
      const result = await listProductCustomerSessions(customer.id);
      setDialog({ type: "sessions", customer, sessions: result.sessions, observedAt: result.observed_at, loading: false, error: "" });
    } catch (error) { setDialog({ type: "sessions", customer, sessions: [], observedAt: "", loading: false, error: error instanceof Error ? error.message : "نشست‌ها بارگذاری نشد." }); }
    finally { setBusyID(""); }
  }

  async function killSession(customer: ProductCustomer, session: ProductActiveSession) {
    if (!window.confirm(`فقط نشست انتخاب‌شده برای ${customer.username} قطع شود؟ رمز و لینک اشتراک تغییر نمی‌کنند.`)) return;
    setSessionBusyID(session.session_id);
    try {
      const result = await killProductCustomerSessionAndReload(customer.id, session.session_id);
      setDialog({ type: "sessions", customer, sessions: result.sessions, observedAt: result.observed_at, loading: false, error: "" });
    } catch (error) {
      setDialog((current) => current?.type === "sessions" && current.customer.id === customer.id
        ? { ...current, loading: false, error: error instanceof Error ? error.message : "قطع نشست انجام نشد." }
        : current);
    } finally { setSessionBusyID(""); }
  }

  return <main className="product-page customers-product-page">
    <header className="product-hero clean-hero"><div><p className="eyebrow">مدیریت سرویس</p><h1>کاربران</h1><p>ساخت، تمدید و مدیریت حساب‌ها از یک صفحه ساده.</p></div><div className="hero-actions"><button className="button-secondary" onClick={() => void refresh()}>↻ بروزرسانی</button><button className="primary-action" onClick={() => setDialog({ type: "create" })}>＋ کاربر جدید</button></div></header>

    <section className="product-stats clean-stats">
      <article><span className="stat-icon">◎</span><div><small>کل کاربران</small><strong>{total.toLocaleString("fa-IR")}</strong></div></article>
      <article><span className="stat-icon success">✓</span><div><small>فعال در این صفحه</small><strong>{activeCount.toLocaleString("fa-IR")}</strong></div></article>
      <article><span className="stat-icon online">●</span><div><small>آنلاین در این صفحه</small><strong>{onlineCount.toLocaleString("fa-IR")}</strong></div></article>
      <article><span className="stat-icon gold">▣</span><div><small>پلن‌ها</small><strong>{plans.length.toLocaleString("fa-IR")}</strong></div></article>
    </section>

    <section className="product-card product-directory clean-directory">
      <div className="directory-title"><div><h2>لیست کاربران</h2><p>جستجو و فیلتر سریع</p></div><span>{total.toLocaleString("fa-IR")} نتیجه</span></div>
      <form className="product-toolbar compact-toolbar" onSubmit={(event) => { event.preventDefault(); patchFilters({ q: searchDraft }); }}>
        <label className="search-box"><span>⌕</span><input value={searchDraft} onChange={(e) => setSearchDraft(e.target.value)} placeholder="جستجوی نام کاربری یا یادداشت…" /></label>
        <label><span>وضعیت</span><select value={filters.status} onChange={(e) => patchFilters({ status: e.target.value })}><option value="">همه وضعیت‌ها</option><option value="active">فعال</option><option value="pending">منتظر اتصال</option><option value="on_hold">On hold</option><option value="suspended">تعلیق</option><option value="expired">منقضی</option><option value="depleted">حجم تمام</option><option value="revoked">لغوشده</option></select></label>
        <label><span>پلن</span><select value={filters.plan} onChange={(e) => patchFilters({ plan: e.target.value })}><option value="">همه پلن‌ها</option>{plans.map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}</select></label>
        <button type="submit">جستجو</button>
        <details className="more-filters"><summary>فیلترهای بیشتر</summary><div className="advanced-filter-grid">
          <label><span>گروه</span><select value={filters.group} onChange={(e) => patchFilters({ group: e.target.value })}><option value="">همه</option>{groups.map((group) => <option key={group.id} value={group.id}>{group.name}</option>)}</select></label>
          <label><span>تگ</span><select value={filters.tag} onChange={(e) => patchFilters({ tag: e.target.value })}><option value="">همه</option>{tags.map((tag) => <option key={tag.id} value={tag.id}>{tag.name}</option>)}</select></label>
          <label><span>حجم</span><select value={filters.unlimitedVolume === null ? "" : String(filters.unlimitedVolume)} onChange={(e) => patchFilters({ unlimitedVolume: e.target.value === "" ? null : e.target.value === "true" })}><option value="">همه</option><option value="false">حجمی</option><option value="true">نامحدود</option></select></label>
          <label><span>انقضا</span><select value={filters.unlimitedExpiry === null ? "" : String(filters.unlimitedExpiry)} onChange={(e) => patchFilters({ unlimitedExpiry: e.target.value === "" ? null : e.target.value === "true" })}><option value="">همه</option><option value="false">دارای انقضا</option><option value="true">بدون انقضا</option></select></label>
          <label><span>از تاریخ</span><input type="date" value={filters.expiryFrom ? filters.expiryFrom.slice(0, 10) : ""} onChange={(e) => patchFilters({ expiryFrom: e.target.value ? `${e.target.value}T00:00:00.000Z` : "" })} /></label>
          <label><span>تا تاریخ</span><input type="date" value={filters.expiryTo ? filters.expiryTo.slice(0, 10) : ""} onChange={(e) => patchFilters({ expiryTo: e.target.value ? `${e.target.value}T23:59:59.999Z` : "" })} /></label>
          <label><span>مرتب‌سازی</span><select value={filters.sort} onChange={(e) => patchFilters({ sort: e.target.value as ProductFilters["sort"] })}><option value="updated">آخرین تغییر</option><option value="username">نام کاربری</option><option value="expiry">انقضا</option><option value="created">ساخت</option><option value="last_renewal">آخرین تمدید</option></select></label>
          <label><span>تعداد در صفحه</span><select value={filters.pageSize} onChange={(e) => patchFilters({ pageSize: Number(e.target.value) as ProductFilters["pageSize"] })}><option value="10">10</option><option value="20">20</option><option value="25">25</option><option value="50">50</option><option value="100">100</option></select></label>
          <button type="button" className="button-secondary" onClick={resetFilters}>پاک کردن فیلترها</button>
        </div></details>
      </form>

      {selected.size > 0 && <div className="bulk-strip"><strong>{selected.size.toLocaleString("fa-IR")} کاربر انتخاب شده</strong><div><button className="primary-action" onClick={() => setDialog({ type: "bulk" })}>عملیات گروهی</button><button className="button-secondary" onClick={() => setSelected(new Set())}>لغو انتخاب</button></div></div>}
      {message && <div className="product-message" role="status">{message}</div>}

      <div className="product-table-wrap"><table className="product-table compact-table"><thead><tr><th><input type="checkbox" aria-label="انتخاب صفحه" checked={allPageSelected} onChange={() => setSelected((current) => { const next = new Set(current); const all = customers.every((c) => next.has(c.id)); customers.forEach((c) => all ? next.delete(c.id) : next.add(c.id)); return next; })} /></th><th>کاربر</th><th>پلن</th><th>حجم</th><th>مصرف</th><th>وضعیت اتصال</th><th>انقضا</th><th>فعال</th><th></th></tr></thead><tbody>
        {loading && <tr><td colSpan={9}><div className="table-loading">در حال بارگذاری…</div></td></tr>}
        {!loading && customers.length === 0 && <tr><td colSpan={9}><div className="empty-state"><strong>کاربری پیدا نشد</strong><span>فیلترها را تغییر بده.</span></div></td></tr>}
        {customers.map((customer) => {
          const usage = usagePresentation(customer); const busy = busyID === customer.id; const suspended = customer.status_dimensions.lifecycle === "suspended"; const revoked = customer.status_dimensions.lifecycle === "revoked";
          return <tr key={customer.id} className={revoked ? "muted-row" : ""}>
            <td><input type="checkbox" checked={selected.has(customer.id)} onChange={() => toggle(customer.id)} /></td>
            <td><div className="account-cell"><span className="account-avatar">{customer.username.slice(0, 1).toUpperCase()}</span><div><strong>{customer.username}</strong><span className={`status-pill ${statusTone(customer)}`}>{commercialLabel(customer)}</span>{customer.note && <small title={customer.note}>{customer.note}</small>}</div></div></td>
            <td><div className="stack-cell"><strong>{customer.plan_name || "Custom"}</strong><small>{customer.group?.name || "بدون گروه"}</small>{customer.tags?.length ? <div className="tag-row">{customer.tags.slice(0, 2).map((tag) => <span key={tag.id}>{tag.name}</span>)}</div> : null}</div></td>
            <td><strong>{customer.quota_bytes === null ? "نامحدود" : formatBytes(customer.quota_bytes)}</strong></td>
            <td><div className="stack-cell"><strong>{usage.exact ? usage.used : "—"}</strong><small>{usage.exact ? `باقی: ${usage.remaining}` : "در انتظار داده"}</small></div></td>
            <td><div className="presence-cell"><i className={usage.presence === "آنلاین" ? "online" : "offline"}/><div><strong>{usage.exact ? usage.presence : "—"}</strong><small>{usage.exact ? `${usage.sessions} نشست` : ""}</small></div></div></td>
            <td><div className="stack-cell"><strong>{customer.no_expiry ? "بدون انقضا" : formatPanelDate(customer.expires_at)}</strong><small>{customer.start_policy === "on_first_successful_connection" ? "از اولین اتصال" : "از زمان ثبت"}</small></div></td>
            <td><label className="toggle-switch" data-disabled={revoked ? "true" : "false"} title={suspended ? "فعال‌سازی" : "تعلیق"}><input type="checkbox" checked={!suspended && !revoked} disabled={busy || revoked} onChange={() => void lifecycle(customer, suspended ? "resume" : "suspend")} /><span/></label></td>
            <td><div className="row-actions"><details className="row-more"><summary aria-label={`عملیات ${customer.username}`}>•••</summary><div className="more-menu"><button disabled={busy} onClick={() => setDialog({ type: "metadata", customer })}>ویرایش مشخصات</button><button disabled={busy} onClick={() => setDialog({ type: "renew", customer })}>تمدید سرویس</button><button disabled={busy || !customer.subscription_retrievable} onClick={() => void openSubscription(customer)}>اشتراک و QR</button><button disabled={busy} onClick={() => setDialog({ type: "password", customer })}>تغییر رمز</button><button disabled={busy} onClick={() => void openSubscription(customer, true)}>صدور لینک جدید</button><button disabled={busy || revoked} onClick={() => void resetUsage(customer)}>Reset مصرف</button><button disabled={busy} onClick={() => void openSessions(customer)}>نشست‌های فعال</button><button className="danger-action" disabled={busy || revoked} onClick={() => void lifecycle(customer, "revoke")}>لغو حساب</button></div></details></div></td>
          </tr>;
        })}
      </tbody></table></div>
      <div className="pagination"><div><button disabled={filters.page <= 1} onClick={() => patchFilters({ page: filters.page - 1 })}>قبلی</button><button disabled={filters.page >= pageCount} onClick={() => patchFilters({ page: filters.page + 1 })}>بعدی</button></div><span>صفحه {filters.page.toLocaleString("fa-IR")} از {pageCount.toLocaleString("fa-IR")}</span></div>
    </section>

    {dialog?.type === "create" && <Modal title="کاربر جدید" eyebrow="ساخت حساب" wide onClose={() => setDialog(null)}><CreateForm plans={plans} groups={groups} tags={tags} onClose={() => setDialog(null)} onDone={async (result) => copiedSecret(result.username, result.password, result.subscriptionPath, result.accountPagePath, "اکانت ساخته شد. رمز تولیدشده فقط همین بار قابل مشاهده است.")} /></Modal>}
    {dialog?.type === "metadata" && <Modal title={`ویرایش ${dialog.customer.username}`} eyebrow="مشخصات کاربر" wide onClose={() => setDialog(null)}><MetadataForm customer={dialog.customer} plans={plans} groups={groups} tags={tags} onClose={() => setDialog(null)} onDone={refresh} /></Modal>}
    {dialog?.type === "renew" && <Modal title={`تمدید ${dialog.customer.username}`} eyebrow="تمدید سرویس" onClose={() => setDialog(null)}><RenewalForm customer={dialog.customer} plans={plans} onClose={() => setDialog(null)} onDone={refresh} /></Modal>}
    {dialog?.type === "password" && <Modal title={`تغییر رمز ${dialog.customer.username}`} eyebrow="امنیت" onClose={() => setDialog(null)}><PasswordForm customer={dialog.customer} onClose={() => setDialog(null)} onDone={async (password) => copiedSecret(dialog.customer.username, password, undefined, undefined, "رمز تغییر کرد؛ لینک اشتراک بدون تغییر باقی ماند.")} /></Modal>}
    {dialog?.type === "subscription" && <Modal title={`اشتراک · ${dialog.customer.username}`} eyebrow={dialog.rotated ? "لینک جدید" : "نمایش اشتراک"} wide onClose={() => setDialog(null)}><div className="readonly-banner">{dialog.rotated ? "لینک قبلی باطل و لینک جدید صادر شد؛ رمز کاربر تغییر نکرد." : "لینک اشتراک و QR آماده استفاده است."}</div><SubscriptionContent delivery={dialog.delivery} /></Modal>}
    {dialog?.type === "secret" && <Modal title={`تحویل امن · ${dialog.username}`} eyebrow="فقط یک‌بار" onClose={() => setDialog(null)}><div className="product-warning"><strong>{dialog.notice}</strong></div>{dialog.password && <label>رمز<div className="copy-row"><input readOnly value={dialog.password} /><button onClick={() => void navigator.clipboard.writeText(dialog.password!)}>کپی</button></div></label>}{dialog.subscriptionPath && <><label>صفحه اشتراک کاربر<div className="copy-row"><input readOnly value={absoluteSubscription(humanAccountPath(dialog.subscriptionPath, dialog.accountPagePath))} /><a className="open-account-page" href={absoluteSubscription(humanAccountPath(dialog.subscriptionPath, dialog.accountPagePath))} target="_blank" rel="noreferrer">باز کردن صفحه</a></div></label><label>لینک ساب کلاینت<div className="copy-row"><input readOnly value={absoluteSubscription(dialog.subscriptionPath)} /><button onClick={() => void navigator.clipboard.writeText(absoluteSubscription(dialog.subscriptionPath!))}>کپی</button></div></label></>}</Modal>}
    {dialog?.type === "bulk" && <Modal title="عملیات گروهی" eyebrow="بازبینی قبل از اجرا" wide onClose={() => setDialog(null)}><BulkForm selected={Array.from(selected)} plans={plans} groups={groups} tags={tags} onClose={() => setDialog(null)} onDone={async () => { setSelected(new Set()); await refresh(); }} /></Modal>}
    {dialog?.type === "sessions" && <SessionsModal dialog={dialog} busySessionID={sessionBusyID} onKill={(session) => void killSession(dialog.customer, session)} onClose={() => setDialog(null)} />}
  </main>;
}
