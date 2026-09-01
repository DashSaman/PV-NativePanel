import { FormEvent, useCallback, useEffect, useState } from "react";
import type { Principal } from "./auth";
import {
  createProductGroup,
  createProductPlan,
  createProductTag,
  listProductGroups,
  listProductPlans,
  listProductTags,
  type ProductGroup,
  type ProductPlan,
  type ProductTag,
} from "./productApi";
import { canManagePlans, formatBytes } from "./productPanelModel";
import "./product-panel.css";

type Props = { role: Principal["role"] };
type CatalogTab = "plans" | "groups" | "tags";

function resetLabel(plan: ProductPlan): string {
  if (plan.reset_strategy === "none") return "بدون ریست";
  if (plan.reset_strategy === "daily") return "روزانه";
  if (plan.reset_strategy === "weekly") return "هفتگی";
  if (plan.reset_strategy === "monthly") return "ماهانه";
  if (plan.reset_strategy === "yearly") return "سالانه";
  return `هر ${plan.reset_custom_days || "؟"} روز`;
}

export function ProductCatalog({ role }: Props) {
  const [tab, setTab] = useState<CatalogTab>("plans");
  const [showPlanForm, setShowPlanForm] = useState(false);
  const [plans, setPlans] = useState<ProductPlan[]>([]); const [groups, setGroups] = useState<ProductGroup[]>([]); const [tags, setTags] = useState<ProductTag[]>([]);
  const [loading, setLoading] = useState(true); const [message, setMessage] = useState(""); const [busy, setBusy] = useState(false);
  const [planName, setPlanName] = useState(""); const [planQuota, setPlanQuota] = useState("50"); const [planUnlimited, setPlanUnlimited] = useState(false);
  const [planDays, setPlanDays] = useState("30"); const [planNoExpiry, setPlanNoExpiry] = useState(false); const [planStart, setPlanStart] = useState<"on_creation" | "on_first_successful_connection">("on_creation");
  const [resetStrategy, setResetStrategy] = useState<ProductPlan["reset_strategy"]>("none"); const [resetCustomDays, setResetCustomDays] = useState("30");
  const [planConcurrency, setPlanConcurrency] = useState("1"); const [planConcurrencyUnlimited, setPlanConcurrencyUnlimited] = useState(false);
  const [planUniqueIP, setPlanUniqueIP] = useState("1"); const [planUniqueIPUnlimited, setPlanUniqueIPUnlimited] = useState(true);
  const [defaultGroup, setDefaultGroup] = useState(""); const [planTags, setPlanTags] = useState<Set<string>>(new Set());
  const [groupName, setGroupName] = useState(""); const [tagName, setTagName] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true); setMessage("");
    try { const [p, g, t] = await Promise.all([listProductPlans(), listProductGroups(), listProductTags()]); setPlans(p); setGroups(g); setTags(t); }
    catch (error) { setMessage(error instanceof Error ? error.message : "اطلاعات بارگذاری نشد."); }
    finally { setLoading(false); }
  }, []);
  useEffect(() => { void refresh(); }, [refresh]);

  async function submitPlan(event: FormEvent) {
    event.preventDefault(); if (!canManagePlans(role)) return; setBusy(true); setMessage("");
    try {
      await createProductPlan({ name: planName.trim(), ...(planUnlimited ? { unlimited_quota: true } : { quota_gb: Number(planQuota) }), ...(planNoExpiry ? { no_expiry: true } : { validity_days: Number(planDays) }), start_policy: planNoExpiry ? "on_creation" : planStart, reset_strategy: resetStrategy, ...(resetStrategy === "custom" ? { reset_custom_days: Number(resetCustomDays) } : {}), concurrency_limit: planConcurrencyUnlimited ? null : Number(planConcurrency), unique_ip_limit: planUniqueIPUnlimited ? null : Number(planUniqueIP), ...(defaultGroup ? { default_group_id: defaultGroup } : {}), tag_ids: Array.from(planTags), enabled: true });
      setPlanName(""); setShowPlanForm(false); setMessage("پلن ساخته شد."); await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت پلن انجام نشد."); }
    finally { setBusy(false); }
  }
  async function submitGroup(event: FormEvent) { event.preventDefault(); setBusy(true); setMessage(""); try { await createProductGroup(groupName.trim()); setGroupName(""); await refresh(); } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت گروه انجام نشد."); } finally { setBusy(false); } }
  async function submitTag(event: FormEvent) { event.preventDefault(); setBusy(true); setMessage(""); try { await createProductTag(tagName.trim()); setTagName(""); await refresh(); } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت تگ انجام نشد."); } finally { setBusy(false); } }

  return <main className="product-page catalog-page">
    <header className="product-hero clean-hero"><div><p className="eyebrow">سرویس‌ها و دسته‌بندی</p><h1>پلن‌ها و دسته‌بندی‌ها</h1><p>پلن‌ها، گروه‌ها و تگ‌ها را مرتب و یکجا مدیریت کن.</p></div><button className="button-secondary" onClick={() => void refresh()} disabled={loading}>↻ بروزرسانی</button></header>
    {message && <div className="product-message" role="status">{message}</div>}

    <nav className="catalog-tabs" aria-label="کاتالوگ"><button className={tab === "plans" ? "active" : ""} onClick={() => setTab("plans")}>پلن‌ها <span>{plans.length}</span></button><button className={tab === "groups" ? "active" : ""} onClick={() => setTab("groups")}>گروه‌ها <span>{groups.length}</span></button><button className={tab === "tags" ? "active" : ""} onClick={() => setTab("tags")}>تگ‌ها <span>{tags.length}</span></button></nav>

    {tab === "plans" && <section className="product-card catalog-main-card">
      <div className="catalog-section-head"><div><h2>پلن‌های سرویس</h2><p>پلن‌های آماده برای ساخت و تمدید کاربران.</p></div>{canManagePlans(role) && <button className="primary-action" onClick={() => setShowPlanForm((value) => !value)}>{showPlanForm ? "بستن" : "＋ پلن جدید"}</button>}</div>
      {showPlanForm && canManagePlans(role) && <form className="product-form plan-form-card" onSubmit={submitPlan}>
        <div className="form-row"><label>نام پلن<input value={planName} onChange={(e) => setPlanName(e.target.value)} required maxLength={120}/></label><label>شروع اعتبار<select value={planStart} onChange={(e) => setPlanStart(e.target.value as typeof planStart)} disabled={planNoExpiry}><option value="on_creation">از زمان ساخت</option><option value="on_first_successful_connection">از اولین اتصال موفق</option></select></label></div>
        <div className="form-row"><fieldset><legend>حجم</legend><label className="check-line"><input type="checkbox" checked={planUnlimited} onChange={(e) => setPlanUnlimited(e.target.checked)}/><span>نامحدود</span></label>{!planUnlimited && <label>حجم (GB)<input type="number" min="1" value={planQuota} onChange={(e) => setPlanQuota(e.target.value)} required/></label>}</fieldset><fieldset><legend>اعتبار</legend><label className="check-line"><input type="checkbox" checked={planNoExpiry} onChange={(e) => setPlanNoExpiry(e.target.checked)}/><span>بدون انقضا</span></label>{!planNoExpiry && <label>روز<input type="number" min="1" value={planDays} onChange={(e) => setPlanDays(e.target.value)} required/></label>}</fieldset></div>
        <div className="form-row"><label>گروه پیش‌فرض<select value={defaultGroup} onChange={(e) => setDefaultGroup(e.target.value)}><option value="">بدون گروه</option>{groups.filter((g) => g.enabled).map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}</select></label><label>ریست حجم<select value={resetStrategy} onChange={(e) => setResetStrategy(e.target.value as ProductPlan["reset_strategy"])}><option value="none">بدون ریست</option><option value="daily">روزانه</option><option value="weekly">هفتگی</option><option value="monthly">ماهانه</option><option value="yearly">سالانه</option><option value="custom">سفارشی</option></select></label></div>
        {resetStrategy === "custom" && <label>هر چند روز<input type="number" min="1" max="3660" value={resetCustomDays} onChange={(e) => setResetCustomDays(e.target.value)}/></label>}
        <fieldset><legend>اتصال همزمان</legend><label className="check-line"><input type="checkbox" checked={planConcurrencyUnlimited} onChange={(e) => setPlanConcurrencyUnlimited(e.target.checked)}/><span>نامحدود</span></label>{!planConcurrencyUnlimited && <label>تعداد اتصال همزمان<input type="number" min="1" max="10000" value={planConcurrency} onChange={(e) => setPlanConcurrency(e.target.value)} required/></label>}</fieldset>
        <fieldset><legend>آی‌پی یکتا</legend><label className="check-line"><input type="checkbox" checked={planUniqueIPUnlimited} onChange={(e) => setPlanUniqueIPUnlimited(e.target.checked)}/><span>نامحدود</span></label>{!planUniqueIPUnlimited && <label>حداکثر آی‌پی یکتا<input type="number" min="1" max="10000" value={planUniqueIP} onChange={(e) => setPlanUniqueIP(e.target.value)} required/></label>}</fieldset>
        <fieldset><legend>تگ‌های پیش‌فرض</legend><div className="chip-checks">{tags.filter((tag) => tag.enabled).map((tag) => <label key={tag.id}><input type="checkbox" checked={planTags.has(tag.id)} onChange={() => setPlanTags((current) => { const next = new Set(current); next.has(tag.id) ? next.delete(tag.id) : next.add(tag.id); return next; })}/><span>{tag.name}</span></label>)}</div></fieldset>
        <div className="modal-actions"><button type="button" className="button-secondary" onClick={() => setShowPlanForm(false)}>انصراف</button><button className="primary-action" disabled={busy}>{busy ? "در حال ذخیره…" : "ذخیره پلن"}</button></div>
      </form>}
      <div className="plan-grid">{loading && <p className="muted">در حال بارگذاری…</p>}{!loading && plans.length === 0 && <div className="empty-state"><strong>هنوز پلنی ساخته نشده</strong><span>برای شروع یک پلن جدید بساز.</span></div>}{plans.map((plan) => <article className="plan-card" key={plan.id}><div className="plan-card-head"><strong>{plan.name}</strong><span className={plan.enabled ? "enabled" : "disabled"}>{plan.enabled ? "فعال" : "غیرفعال"}</span></div><div className="plan-price"><b>{plan.quota_bytes === null ? "∞" : formatBytes(plan.quota_bytes)}</b><small>{plan.no_expiry ? "بدون انقضا" : `${Math.round((plan.validity_seconds || 0) / 86400).toLocaleString("fa-IR")} روز`}</small></div><div className="plan-meta"><span>{plan.start_policy === "on_first_successful_connection" ? "شروع از اولین اتصال" : "شروع از زمان ساخت"}</span><span>{resetLabel(plan)}</span><span>اتصال همزمان: {plan.concurrency_limit === null ? "نامحدود" : plan.concurrency_limit.toLocaleString("fa-IR")}</span><span>آی‌پی یکتا: {plan.unique_ip_limit === null ? "نامحدود" : plan.unique_ip_limit.toLocaleString("fa-IR")}</span></div></article>)}</div>
      {plans.some((plan) => plan.reset_strategy !== "none" && !plan.reset_enforcement_available) && <div className="subtle-warning">ریست دوره‌ای برای این پلن‌ها ذخیره می‌شود، اما اجرای خودکار آن هنوز فعال نشده است.</div>}
    </section>}

    {tab === "groups" && <section className="product-card catalog-main-card"><div className="catalog-section-head"><div><h2>گروه‌ها</h2><p>کاربران را برای مدیریت سریع دسته‌بندی کن.</p></div></div><form className="inline-create catalog-create" onSubmit={submitGroup}><input placeholder="نام گروه جدید" value={groupName} onChange={(e) => setGroupName(e.target.value)} required/><button disabled={busy}>＋ افزودن گروه</button></form><div className="catalog-chip-list">{groups.map((group) => <div key={group.id}><span className="chip-icon">▦</span><strong>{group.name}</strong><small>{group.enabled ? "فعال" : "غیرفعال"}</small></div>)}</div></section>}

    {tab === "tags" && <section className="product-card catalog-main-card"><div className="catalog-section-head"><div><h2>تگ‌ها</h2><p>برای جستجو و فیلتر بهتر، برچسب بساز.</p></div></div><form className="inline-create catalog-create" onSubmit={submitTag}><input placeholder="نام تگ جدید" value={tagName} onChange={(e) => setTagName(e.target.value)} required/><button disabled={busy}>＋ افزودن تگ</button></form><div className="catalog-chip-list">{tags.map((tag) => <div key={tag.id}><span className="chip-icon">#</span><strong>{tag.name}</strong><small>{tag.enabled ? "فعال" : "غیرفعال"}</small></div>)}</div></section>}
  </main>;
}
