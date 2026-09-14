import { useEffect, useMemo, useState } from "react";
import type { Principal } from "./auth";
import { DEFAULT_PRODUCT_FILTERS, listProductCustomers, listProductPlans } from "./productApi";
import { getRuntimeStatus } from "./runtime";
import { canUseCustomerProduct, canUseRawRuntime } from "./productPanelModel";
import { SystemDashboard } from "./SystemDashboard";
import { DonutChart } from "./charts";
import { Icon } from "./ui";

type Props = { role: Principal["role"] };
type Snapshot = {
  total: number;
  active: number;
  pending: number;
  suspended: number;
  ended: number;
  expiring7: number;
  expiring30: number;
  plans: number;
  runtimeReady: boolean | null;
};

const empty: Snapshot = { total: 0, active: 0, pending: 0, suspended: 0, ended: 0, expiring7: 0, expiring30: 0, plans: 0, runtimeReady: null };

export function Dashboard({ role }: Props) {
  const [snapshot, setSnapshot] = useState<Snapshot>(empty);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");

  async function refresh() {
    if (!canUseCustomerProduct(role)) { setLoading(false); return; }
    setLoading(true); setMessage("");
    const now = new Date();
    const in7 = new Date(now.getTime() + 7 * 86400000).toISOString();
    const in30 = new Date(now.getTime() + 30 * 86400000).toISOString();
    try {
      const base = { ...DEFAULT_PRODUCT_FILTERS, page: 1, pageSize: 10 as const };
      const [all, active, pending, suspended, expired, depleted, expiring7, expiring30, plans, runtime] = await Promise.all([
        listProductCustomers(base),
        listProductCustomers({ ...base, status: "active" }),
        listProductCustomers({ ...base, status: "pending" }),
        listProductCustomers({ ...base, status: "suspended" }),
        listProductCustomers({ ...base, status: "expired" }),
        listProductCustomers({ ...base, status: "depleted" }),
        listProductCustomers({ ...base, expiryFrom: now.toISOString(), expiryTo: in7, unlimitedExpiry: false }),
        listProductCustomers({ ...base, expiryFrom: now.toISOString(), expiryTo: in30, unlimitedExpiry: false }),
        listProductPlans(),
        canUseRawRuntime(role) ? getRuntimeStatus().catch(() => null) : Promise.resolve(null),
      ]);
      setSnapshot({
        total: all.total,
        active: active.total,
        pending: pending.total,
        suspended: suspended.total,
        ended: expired.total + depleted.total,
        expiring7: expiring7.total,
        expiring30: expiring30.total,
        plans: plans.length,
        runtimeReady: runtime ? runtime.runtime_available : null,
      });
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "داده‌های داشبورد بارگذاری نشد.");
    } finally { setLoading(false); }
  }

  useEffect(() => { void refresh(); }, [role]);

  /* Every ratio below derives from live queries — no synthetic history. */
  const ratios = useMemo(() => {
    const total = Math.max(snapshot.total, 1);
    return {
      active: snapshot.active / total,
      pending: snapshot.pending / total,
      suspended: snapshot.suspended / total,
      ended: Math.min(1, Math.max(0, 1 - (snapshot.active + snapshot.pending + snapshot.suspended) / total)),
      expiring7: snapshot.total ? snapshot.expiring7 / snapshot.total : 0,
      expiring30: snapshot.total ? snapshot.expiring30 / snapshot.total : 0,
      attention: snapshot.total ? (snapshot.suspended + snapshot.ended) / snapshot.total : 0,
    };
  }, [snapshot]);

  const donutValues = [snapshot.active, snapshot.pending, snapshot.suspended, snapshot.ended];

  if (!canUseCustomerProduct(role)) return <main className="dashboard-page"><section className="dashboard-card"><h1>داشبورد</h1><p className="muted">این نقش دسترسی عملیاتی به مدیریت کاربران ندارد.</p></section></main>;

  return <main className="dashboard-page">
    <header className="dashboard-hero">
      <div><p className="eyebrow">PVNaive Control Panel</p><h1>داشبورد</h1><p>وضعیت کلی سرویس‌ها و کاربران در یک نگاه.</p></div>
      <button className="button-secondary" onClick={() => void refresh()} disabled={loading}><Icon name="refresh" size={15}/> بروزرسانی</button>
    </header>
    {message && <div className="product-message danger">{message}</div>}

    <section className="dashboard-kpis">
      <article>
        <span className="kpi-symbol"><Icon name="users" size={20}/></span>
        <div><small>کل کاربران</small><strong>{loading ? "…" : snapshot.total.toLocaleString("fa-IR")}</strong><em>حساب‌های ثبت‌شده</em></div>
      </article>
      <article>
        <span className="kpi-symbol success"><Icon name="check" size={20}/></span>
        <div><small>کاربران فعال</small><strong>{loading ? "…" : snapshot.active.toLocaleString("fa-IR")}</strong><em>سرویس قابل استفاده</em></div>
        <i className="kpi-share success" aria-hidden="true"><b style={{ width: `${Math.round(ratios.active * 100)}%` }}/></i>
      </article>
      <article>
        <span className="kpi-symbol gold"><Icon name="plans" size={20}/></span>
        <div><small>پلن‌ها</small><strong>{loading ? "…" : snapshot.plans.toLocaleString("fa-IR")}</strong><em>پلن‌های تعریف‌شده</em></div>
      </article>
      <article>
        <span className="kpi-symbol warning"><Icon name="alert" size={20}/></span>
        <div><small>نیازمند توجه</small><strong>{loading ? "…" : (snapshot.suspended + snapshot.ended).toLocaleString("fa-IR")}</strong><em>تعلیق یا پایان سرویس</em></div>
        <i className="kpi-share warning" aria-hidden="true"><b style={{ width: `${Math.round(ratios.attention * 100)}%` }}/></i>
      </article>
    </section>

    <section className="dashboard-grid">
      <article className="dashboard-card">
        <div className="dashboard-card-head"><div><p className="eyebrow">وضعیت کاربران</p><h2>توزیع سرویس‌ها</h2></div><a href="/panel/#/customers">مشاهده کاربران ←</a></div>
        <div className="status-overview">
          <DonutChart
            values={donutValues}
            colors={["var(--green)", "var(--gold)", "var(--orange)", "var(--red)"]}
            center={snapshot.total.toLocaleString("fa-IR")}
            caption="کاربر"
            ariaLabel="توزیع وضعیت کاربران"
          />
          <div className="status-legend">
            <div><i className="dot success"/><span>فعال</span><strong>{snapshot.active.toLocaleString("fa-IR")}</strong><em className="legend-share">{Math.round(ratios.active * 100).toLocaleString("fa-IR")}٪</em></div>
            <div><i className="dot gold"/><span>منتظر اتصال</span><strong>{snapshot.pending.toLocaleString("fa-IR")}</strong><em className="legend-share">{Math.round(ratios.pending * 100).toLocaleString("fa-IR")}٪</em></div>
            <div><i className="dot warning"/><span>تعلیق</span><strong>{snapshot.suspended.toLocaleString("fa-IR")}</strong><em className="legend-share">{Math.round(ratios.suspended * 100).toLocaleString("fa-IR")}٪</em></div>
            <div><i className="dot danger"/><span>پایان‌یافته</span><strong>{snapshot.ended.toLocaleString("fa-IR")}</strong><em className="legend-share">{Math.round(ratios.ended * 100).toLocaleString("fa-IR")}٪</em></div>
          </div>
        </div>
      </article>

      <article className="dashboard-card">
        <div className="dashboard-card-head"><div><p className="eyebrow">انقضای نزدیک</p><h2>سرویس‌های در آستانه پایان</h2></div></div>
        <div className="expiry-overview">
          <div><span>۷ روز آینده</span><strong>{snapshot.expiring7.toLocaleString("fa-IR")}</strong><i><b className="bar-gold" style={{ width: `${Math.min(100, ratios.expiring7 * 100)}%` }}/></i></div>
          <div><span>۳۰ روز آینده</span><strong>{snapshot.expiring30.toLocaleString("fa-IR")}</strong><i><b style={{ width: `${Math.min(100, ratios.expiring30 * 100)}%` }}/></i></div>
          <div className="runtime-summary"><span>Runtime</span><strong>{snapshot.runtimeReady === null ? "طبق نقش" : snapshot.runtimeReady ? "آماده" : "نیازمند بررسی"}</strong></div>
        </div>
      </article>
    </section>

    {canUseRawRuntime(role) && <SystemDashboard/>}

    <section className="quick-actions dashboard-card">
      <div><p className="eyebrow">دسترسی سریع</p><h2>مدیریت روزمره</h2></div>
      <div><a className="quick-link primary" href="/panel/#/customers"><Icon name="plus" size={15}/> ساخت و مدیریت کاربر</a><a className="quick-link" href="/panel/#/catalog"><Icon name="plans" size={15}/> پلن‌ها و دسته‌بندی‌ها</a>{canUseRawRuntime(role) && <a className="quick-link" href="/panel/#/runtime/naive"><Icon name="system" size={15}/> سیستم / Runtime</a>}</div>
    </section>
  </main>;
}
