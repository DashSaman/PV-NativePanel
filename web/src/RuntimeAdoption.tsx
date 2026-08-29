import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { adoptRuntimeCustomer, type CustomerValidityMode } from "./customers";
import { DEFAULT_PRODUCT_FILTERS, listProductCustomers } from "./productApi";
import { listRuntimeCredentials, type RuntimeCredential } from "./runtime";
import "./product-panel.css";

type AdoptionDialog = { credential: RuntimeCredential } | null;

async function allManagedRuntimeIDs(): Promise<Set<string>> {
  const ids = new Set<string>();
  let page = 1;
  while (true) {
    const result = await listProductCustomers({ ...DEFAULT_PRODUCT_FILTERS, page, pageSize: 100 });
    result.customers.forEach((customer) => { if (customer.runtime_credential_id) ids.add(customer.runtime_credential_id); });
    if (page * result.page_size >= result.total) break;
    page += 1;
  }
  return ids;
}

function AdoptionForm({ credential, onDone, onClose }: { credential: RuntimeCredential; onDone: () => Promise<void>; onClose: () => void }) {
  const [quotaUnlimited, setQuotaUnlimited] = useState(false);
  const [quotaGB, setQuotaGB] = useState("50");
  const [validityMode, setValidityMode] = useState<CustomerValidityMode>("on_creation");
  const [days, setDays] = useState("30");
  const [fixedExpiry, setFixedExpiry] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      const validity = validityMode === "fixed_expiry"
        ? { mode: validityMode, expires_at: new Date(fixedExpiry).toISOString() }
        : { mode: validityMode, duration_days: Number(days) };
      await adoptRuntimeCustomer(credential.id, { quota_gb: quotaUnlimited ? null : Number(quotaGB), validity });
      await onDone(); onClose();
    } catch (error) { setMessage(error instanceof Error ? error.message : "تنظیم سرویس انجام نشد."); }
    finally { setBusy(false); }
  }

  return <form className="product-form" onSubmit={submit}>
    <div className="readonly-banner">Username و Password موجود Runtime برای <strong>{credential.username}</strong> حفظ می‌شود؛ فقط لایه تجاری و Subscription به آن متصل می‌شود.</div>
    <fieldset><legend>حجم</legend><label className="check-line"><input type="checkbox" checked={quotaUnlimited} onChange={(e) => setQuotaUnlimited(e.target.checked)} /><span>حجم نامحدود</span></label>{!quotaUnlimited && <label>حجم کل (GB)<input type="number" min="1" value={quotaGB} onChange={(e) => setQuotaGB(e.target.value)} required /></label>}</fieldset>
    <fieldset><legend>اعتبار</legend><label>نوع شروع<select value={validityMode} onChange={(e) => setValidityMode(e.target.value as CustomerValidityMode)}><option value="on_creation">از همین حالا</option><option value="on_first_successful_connection">از اولین اتصال موفق</option><option value="fixed_expiry">تاریخ پایان دستی</option></select></label>{validityMode === "fixed_expiry" ? <label>تاریخ پایان<input type="datetime-local" value={fixedExpiry} onChange={(e) => setFixedExpiry(e.target.value)} required /></label> : <label>تعداد روز<input type="number" min="1" value={days} onChange={(e) => setDays(e.target.value)} required /></label>}</fieldset>
    {message && <div className="product-message danger">{message}</div>}
    <div className="modal-actions"><button type="button" className="button-secondary" onClick={onClose}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال اتصال…" : "اتصال به مدیریت کاربران"}</button></div>
  </form>;
}

export function RuntimeAdoption() {
  const [credentials, setCredentials] = useState<RuntimeCredential[]>([]);
  const [managedIDs, setManagedIDs] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [dialog, setDialog] = useState<AdoptionDialog>(null);

  const refresh = useCallback(async () => {
    setLoading(true); setMessage("");
    try {
      const [runtime, managed] = await Promise.all([listRuntimeCredentials(), allManagedRuntimeIDs()]);
      setCredentials(runtime); setManagedIDs(managed);
    } catch (error) { setMessage(error instanceof Error ? error.message : "Runtime credentials بارگذاری نشد."); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void refresh(); }, [refresh]);
  const unmanaged = useMemo(() => credentials.filter((credential) => credential.status === "active" && !managedIDs.has(credential.id)), [credentials, managedIDs]);

  return <main className="product-page">
    <header className="product-hero"><div><p className="eyebrow">اتصال اکانت‌های قبلی</p><h1>اکانت‌های قدیمی</h1><p>اکانت‌هایی که از قبل روی Runtime بوده‌اند را بدون تغییر نام کاربری و رمز، به مدیریت کاربران متصل کن.</p></div><div className="hero-actions"><a className="button-secondary" href="/panel/#/runtime/naive">Runtime</a><button className="button-secondary" onClick={() => void refresh()}>↻ بروزرسانی</button></div></header>
    <section className="product-stats"><article><span>Runtime کل</span><strong>{credentials.length.toLocaleString("fa-IR")}</strong></article><article><span>نیازمند اتصال</span><strong>{unmanaged.length.toLocaleString("fa-IR")}</strong></article><article><span>متصل‌شده</span><strong>{managedIDs.size.toLocaleString("fa-IR")}</strong></article><article><span>آماده مدیریت</span><strong>{managedIDs.size.toLocaleString("fa-IR")}</strong></article></section>
    {message && <div className="product-message">{message}</div>}
    <section className="product-card"><div className="product-card-head"><div><p className="eyebrow">نیازمند اتصال</p><h2>اکانت‌های آماده تنظیم</h2></div></div><div className="catalog-list">{loading && <p className="muted">در حال بارگذاری…</p>}{!loading && unmanaged.length === 0 && <p className="muted">همه اکانت‌های فعال Runtime به مدیریت کاربران متصل هستند.</p>}{unmanaged.map((credential) => <div className="catalog-item" key={credential.id}><div><strong>{credential.username}</strong><small>{credential.origin} · {credential.id.slice(0, 12)} · revision {credential.revision}</small></div><button className="primary-action" onClick={() => setDialog({ credential })}>تنظیم سرویس</button></div>)}</div></section>
    {dialog && <div className="product-modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && setDialog(null)}><section className="product-modal" role="dialog" aria-modal="true"><div className="product-modal-head"><div><p className="eyebrow">Adopt existing Runtime</p><h2>{dialog.credential.username}</h2></div><button className="icon-button" onClick={() => setDialog(null)}>×</button></div><AdoptionForm credential={dialog.credential} onClose={() => setDialog(null)} onDone={refresh} /></section></div>}
  </main>;
}
