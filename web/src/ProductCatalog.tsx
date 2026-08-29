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

function resetLabel(plan: ProductPlan): string {
  if (plan.reset_strategy === "none") return "بدون ریست دوره‌ای";
  if (plan.reset_strategy === "daily") return "روزانه";
  if (plan.reset_strategy === "weekly") return "هفتگی";
  if (plan.reset_strategy === "monthly") return "ماهانه";
  if (plan.reset_strategy === "yearly") return "سالانه";
  return `هر ${plan.reset_custom_days || "؟"} روز`;
}

export function ProductCatalog({ role }: Props) {
  const [plans, setPlans] = useState<ProductPlan[]>([]);
  const [groups, setGroups] = useState<ProductGroup[]>([]);
  const [tags, setTags] = useState<ProductTag[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");
  const [planName, setPlanName] = useState("");
  const [planQuota, setPlanQuota] = useState("50");
  const [planUnlimited, setPlanUnlimited] = useState(false);
  const [planDays, setPlanDays] = useState("30");
  const [planNoExpiry, setPlanNoExpiry] = useState(false);
  const [planStart, setPlanStart] = useState<"on_creation" | "on_first_successful_connection">("on_creation");
  const [resetStrategy, setResetStrategy] = useState<ProductPlan["reset_strategy"]>("none");
  const [resetCustomDays, setResetCustomDays] = useState("30");
  const [defaultGroup, setDefaultGroup] = useState("");
  const [planTags, setPlanTags] = useState<Set<string>>(new Set());
  const [groupName, setGroupName] = useState("");
  const [tagName, setTagName] = useState("");
  const [busy, setBusy] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    setMessage("");
    try {
      const [nextPlans, nextGroups, nextTags] = await Promise.all([
        listProductPlans(), listProductGroups(), listProductTags(),
      ]);
      setPlans(nextPlans);
      setGroups(nextGroups);
      setTags(nextTags);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "کاتالوگ سرویس بارگذاری نشد.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void refresh(); }, [refresh]);

  async function submitPlan(event: FormEvent) {
    event.preventDefault();
    if (!canManagePlans(role)) return;
    setBusy(true); setMessage("");
    try {
      await createProductPlan({
        name: planName.trim(),
        ...(planUnlimited ? { unlimited_quota: true } : { quota_gb: Number(planQuota) }),
        ...(planNoExpiry ? { no_expiry: true } : { validity_days: Number(planDays) }),
        start_policy: planNoExpiry ? "on_creation" : planStart,
        reset_strategy: resetStrategy,
        ...(resetStrategy === "custom" ? { reset_custom_days: Number(resetCustomDays) } : {}),
        ...(defaultGroup ? { default_group_id: defaultGroup } : {}),
        tag_ids: Array.from(planTags),
        enabled: true,
      });
      setPlanName("");
      setMessage("پلن ساخته شد.");
      await refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "ساخت پلن انجام نشد.");
    } finally { setBusy(false); }
  }

  async function submitGroup(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      await createProductGroup(groupName.trim());
      setGroupName(""); setMessage("گروه ساخته شد."); await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت گروه انجام نشد."); }
    finally { setBusy(false); }
  }

  async function submitTag(event: FormEvent) {
    event.preventDefault(); setBusy(true); setMessage("");
    try {
      await createProductTag(tagName.trim());
      setTagName(""); setMessage("تگ ساخته شد."); await refresh();
    } catch (error) { setMessage(error instanceof Error ? error.message : "ساخت تگ انجام نشد."); }
    finally { setBusy(false); }
  }

  return <main className="product-page">
    <header className="product-hero">
      <div><p className="eyebrow">PVNaive · Product catalog</p><h1>پلن‌ها، گروه‌ها و تگ‌ها</h1><p>کاتالوگ واقعی WS2؛ هر چیزی که اینجا می‌بینی مستقیماً از API آماده محصول خوانده می‌شود.</p></div>
      <button className="button-secondary" onClick={() => void refresh()} disabled={loading}>↻ بروزرسانی</button>
    </header>
    {message && <div className="product-message" role="status">{message}</div>}

    <section className="product-grid two">
      <article className="product-card">
        <div className="product-card-head"><div><p className="eyebrow">Plans</p><h2>پلن‌های آماده</h2></div><span className="product-count">{plans.length}</span></div>
        <div className="catalog-list">
          {loading && <p className="muted">در حال بارگذاری…</p>}
          {!loading && plans.length === 0 && <p className="muted">هنوز پلنی ساخته نشده است.</p>}
          {plans.map((plan) => <div className="catalog-item" key={plan.id}>
            <div><strong>{plan.name}</strong><small>{plan.quota_bytes === null ? "حجم نامحدود" : formatBytes(plan.quota_bytes)} · {plan.no_expiry ? "بدون انقضا" : `${Math.round((plan.validity_seconds || 0) / 86400).toLocaleString("fa-IR")} روز`}</small></div>
            <div className="catalog-badges"><span>{plan.start_policy === "on_first_successful_connection" ? "از اولین اتصال" : "از زمان ساخت"}</span><span>{resetLabel(plan)}</span></div>
          </div>)}
        </div>
        {plans.some((plan) => plan.reset_strategy !== "none" && !plan.reset_enforcement_available) && <div className="product-warning"><strong>تعریف reset strategy فعال است، اما enforcement دوره‌ای هنوز آماده نیست.</strong><span>پنل آن را به‌عنوان محدودیت اجرایی قطعی نمایش نمی‌دهد.</span></div>}
      </article>

      {canManagePlans(role) ? <article className="product-card">
        <p className="eyebrow">Create preset</p><h2>ساخت پلن</h2>
        <form className="product-form" onSubmit={submitPlan}>
          <label>نام پلن<input value={planName} onChange={(e) => setPlanName(e.target.value)} required maxLength={120} /></label>
          <div className="form-row"><label className="check-line"><input type="checkbox" checked={planUnlimited} onChange={(e) => setPlanUnlimited(e.target.checked)} /><span>حجم نامحدود</span></label>{!planUnlimited && <label>حجم (GB)<input type="number" min="1" value={planQuota} onChange={(e) => setPlanQuota(e.target.value)} required /></label>}</div>
          <div className="form-row"><label className="check-line"><input type="checkbox" checked={planNoExpiry} onChange={(e) => setPlanNoExpiry(e.target.checked)} /><span>بدون انقضا</span></label>{!planNoExpiry && <label>اعتبار (روز)<input type="number" min="1" value={planDays} onChange={(e) => setPlanDays(e.target.value)} required /></label>}</div>
          {!planNoExpiry && <label>شروع اعتبار<select value={planStart} onChange={(e) => setPlanStart(e.target.value as typeof planStart)}><option value="on_creation">از زمان ساخت</option><option value="on_first_successful_connection">از اولین اتصال موفق</option></select></label>}
          <label>استراتژی ریست حجم<select value={resetStrategy} onChange={(e) => setResetStrategy(e.target.value as ProductPlan["reset_strategy"])}><option value="none">بدون ریست</option><option value="daily">روزانه</option><option value="weekly">هفتگی</option><option value="monthly">ماهانه</option><option value="yearly">سالانه</option><option value="custom">سفارشی</option></select></label>
          {resetStrategy === "custom" && <label>هر چند روز<input type="number" min="1" max="3660" value={resetCustomDays} onChange={(e) => setResetCustomDays(e.target.value)} /></label>}
          <label>گروه پیش‌فرض<select value={defaultGroup} onChange={(e) => setDefaultGroup(e.target.value)}><option value="">بدون گروه پیش‌فرض</option>{groups.filter((g) => g.enabled).map((g) => <option key={g.id} value={g.id}>{g.name}</option>)}</select></label>
          <fieldset><legend>تگ‌های پیش‌فرض</legend><div className="chip-checks">{tags.filter((tag) => tag.enabled).map((tag) => <label key={tag.id}><input type="checkbox" checked={planTags.has(tag.id)} onChange={() => setPlanTags((current) => { const next = new Set(current); next.has(tag.id) ? next.delete(tag.id) : next.add(tag.id); return next; })} /><span>{tag.name}</span></label>)}</div></fieldset>
          <button className="primary-action" disabled={busy}>{busy ? "در حال ذخیره…" : "ساخت پلن"}</button>
        </form>
      </article> : <article className="product-card"><p className="eyebrow">RBAC</p><h2>دسترسی نماینده</h2><p className="muted">نماینده می‌تواند پلن‌های فعال را ببیند و برای مشتری اعمال کند؛ ایجاد پلن جدید فقط برای Owner/Admin مجاز است.</p></article>}
    </section>

    <section className="product-grid two">
      <article className="product-card"><div className="product-card-head"><div><p className="eyebrow">Groups</p><h2>گروه‌ها</h2></div><span className="product-count">{groups.length}</span></div><div className="chip-cloud">{groups.map((group) => <span key={group.id} className={!group.enabled ? "disabled" : ""}>{group.name}</span>)}</div><form className="inline-create" onSubmit={submitGroup}><input placeholder="نام گروه جدید" value={groupName} onChange={(e) => setGroupName(e.target.value)} required /><button disabled={busy}>＋ گروه</button></form></article>
      <article className="product-card"><div className="product-card-head"><div><p className="eyebrow">Tags</p><h2>تگ‌ها</h2></div><span className="product-count">{tags.length}</span></div><div className="chip-cloud">{tags.map((tag) => <span key={tag.id} className={!tag.enabled ? "disabled" : ""}>{tag.name}</span>)}</div><form className="inline-create" onSubmit={submitTag}><input placeholder="نام تگ جدید" value={tagName} onChange={(e) => setTagName(e.target.value)} required /><button disabled={busy}>＋ تگ</button></form></article>
    </section>
  </main>;
}
