import { useCallback, useEffect, useMemo, useState } from "react";
import { customerUsagePresentation, expiryLabel, listCustomers, trafficLabel, type CustomerView } from "./customers";
import "./accounting-live.css";

export function AccountingLive() {
  const [customers, setCustomers] = useState<CustomerView[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);

  const refresh = useCallback(async () => {
    try {
      const next = await listCustomers();
      setCustomers(next); setUpdatedAt(new Date()); setMessage("");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Accounting قابل دریافت نیست.");
    } finally { setLoading(false); }
  }, []);

  useEffect(() => {
    void refresh();
    const timer = window.setInterval(() => void refresh(), 10_000);
    return () => window.clearInterval(timer);
  }, [refresh]);

  const summary = useMemo(() => {
    const exact = customers.filter((customer) => customer.usage_capability.available && customer.accounting_complete === true);
    return {
      total: customers.length,
      exact: exact.length,
      incomplete: customers.length - exact.length,
      online: exact.filter((customer) => customer.online === true).length,
      sessions: exact.reduce((sum, customer) => sum + (customer.online_sessions ?? 0), 0),
      used: exact.reduce((sum, customer) => sum + (customer.used_bytes ?? 0), 0),
    };
  }, [customers]);

  return <main className="accounting-live-page">
    <header className="owner-hero"><div className="owner-hero-copy"><p className="eyebrow">PVNaive · Exact Accounting</p><h1>مصرف زنده و Online</h1><p>Source of truth: بایت‌های موفق CONNECT در forwardproxy؛ بدون تخمین از access log.</p></div><div className="header-actions"><button className="button-secondary" onClick={() => void refresh()}>↻ بروزرسانی</button></div></header>
    <section className="accounting-kpis" aria-label="خلاصه accounting">
      <article><span>Accounting دقیق</span><strong>{summary.exact.toLocaleString("fa-IR")}/{summary.total.toLocaleString("fa-IR")}</strong><small>{summary.incomplete ? `${summary.incomplete.toLocaleString("fa-IR")} اکانت نیازمند بررسی` : "همه کامل"}</small></article>
      <article><span>Online</span><strong>{summary.online.toLocaleString("fa-IR")}</strong><small>{summary.sessions.toLocaleString("fa-IR")} session فعال</small></article>
      <article><span>مصرف قطعی</span><strong>{trafficLabel(summary.used)}</strong><small>جمع اکانت‌های complete</small></article>
      <article><span>آخرین بروزرسانی</span><strong className="accounting-time">{updatedAt ? updatedAt.toLocaleTimeString("fa-IR") : "—"}</strong><small>refresh خودکار ۱۰ ثانیه</small></article>
    </section>
    {message && <p className="customer-message" role="alert">{message}</p>}
    <section className="customer-table-card accounting-table-card"><div className="directory-heading"><div><p className="eyebrow">Trusted read model</p><h2>مصرف هر اکانت</h2></div></div><div className="customer-table-wrap"><table className="customer-table owner-table"><thead><tr><th>اکانت</th><th>وضعیت</th><th>مصرف / باقی‌مانده</th><th>Upload</th><th>Download</th><th>آخرین فعالیت</th><th>Accounting</th></tr></thead><tbody>
      {loading && <tr><td colSpan={7}><div className="table-loading">در حال خواندن ledger…</div></td></tr>}
      {!loading && customers.map((customer) => { const usage = customerUsagePresentation(customer); return <tr key={customer.id}>
        <td><div className="account-cell"><span className="account-avatar">{customer.username.slice(0, 1).toUpperCase()}</span><div><strong>{customer.username}</strong><small>{customer.runtime_credential_id.slice(0, 8)}</small></div></div></td>
        <td>{usage.exact ? <span className={`dimension-badge ${customer.online === true ? "success" : "neutral"}`}><i />{customer.online === true ? `آنلاین · ${customer.online_sessions ?? 0}` : "آفلاین"}</span> : <span className="dimension-badge warning"><i />نامشخص</span>}</td>
        <td><strong>{usage.primary}</strong><small className="accounting-secondary">{usage.secondary}</small></td><td>{trafficLabel(customer.upload_bytes)}</td><td>{trafficLabel(customer.download_bytes)}</td><td>{customer.last_online ? expiryLabel(customer.last_online) : "—"}</td><td>{usage.exact ? <span className="usage-ready">کامل</span> : <span className="usage-locked">ناقص / fail-closed</span>}</td>
      </tr>; })}
    </tbody></table></div></section>
  </main>;
}
