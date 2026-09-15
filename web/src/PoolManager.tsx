import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import type { Principal } from "./auth";
import { readCookie } from "./auth";
import { Icon } from "./ui";
import { fmtNum } from "./charts";
import {
  buildRevisionPayload,
  driftLabel,
  formatLastSeen,
  healthLabel,
  maintenanceActions,
  maintenanceLabel,
  manifestSummary,
  normalizeEnrollToken,
  normalizeManifest,
  normalizePoolNodes,
  validateEnrollInput,
  validateRevisionInput,
  type EndpointInput,
  type EnrollTokenResult,
  type PoolNodeRow,
} from "./poolManager";

// R5 Pool Manager — owner-only command surface for the pull-model node
// registry (GATE STEER-006, UI slice R5-UI-001). Every mutation rides the
// backend trusted boundary; this view adds no authority of its own and
// renders state exactly as the registry reports it (no fabricated health,
// no client-side "optimistic" revisions).

type Props = { principal: Principal };

type Toast = { kind: "error" | "success"; text: string } | null;

function mutationHeaders(): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf");
  if (!csrf) throw new Error("توکن CSRF در دسترس نیست.");
  return { "Content-Type": "application/json", "X-CSRF-Token": csrf };
}

async function callAPI(path: string, method: string, body: unknown): Promise<Record<string, unknown>> {
  const response = await fetch(path, { method, credentials: "same-origin", headers: mutationHeaders(), body: JSON.stringify(body) });
  const contentType = response.headers.get("Content-Type") || "";
  const payload = contentType.includes("application/json") ? (await response.json()) as Record<string, unknown> : {};
  if (!response.ok) {
    const message = typeof payload.message === "string" ? payload.message : "درخواست انجام نشد.";
    throw new Error(message);
  }
  return payload;
}

async function getJSON(path: string): Promise<Record<string, unknown>> {
  const response = await fetch(path, { credentials: "same-origin" });
  const contentType = response.headers.get("Content-Type") || "";
  const payload = contentType.includes("application/json") ? (await response.json()) as Record<string, unknown> : {};
  if (!response.ok) {
    const message = typeof payload.message === "string" ? payload.message : "خواندن اطلاعات ناموفق بود.";
    throw new Error(message);
  }
  return payload;
}

const healthPillClass: Record<string, string> = { healthy: "pill-green", degraded: "pill-orange", offline: "pill-red", unknown: "pill-muted" };
const driftPillClass: Record<string, string> = { in_sync: "pill-green", pending: "pill-orange", ahead: "pill-blue" };
const maintenancePillClass: Record<string, string> = { active: "pill-green", draining: "pill-orange", disabled: "pill-red" };

const actionLabels: Record<string, string> = {
  draining: "شروع تخلیه",
  disabled: "تکمیل خروج از استخر",
  active: "بازگشت به سرویس",
};

const TTL_OPTIONS = [
  { minutes: 30, label: "۳۰ دقیقه" },
  { minutes: 120, label: "۲ ساعت (پیش‌فرض)" },
  { minutes: 480, label: "۸ ساعت" },
  { minutes: 1440, label: "۲۴ ساعت" },
];

export function PoolManager({ principal: _principal }: Props) {
  const [nodes, setNodes] = useState<PoolNodeRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [toast, setToast] = useState<Toast>(null);

  // Step 1 — enrollment token wizard state.
  const [tokenName, setTokenName] = useState("");
  const [tokenTTL, setTokenTTL] = useState(120);
  const [issuedToken, setIssuedToken] = useState<EnrollTokenResult | null>(null);

  // Step 2 — enrollment state.
  const [enrollToken, setEnrollToken] = useState("");
  const [enrollName, setEnrollName] = useState("");
  const [enrollRegion, setEnrollRegion] = useState("");
  const [enrollWeight, setEnrollWeight] = useState(1);

  // Step 3 — revision publish state.
  const [revisionNode, setRevisionNode] = useState("");
  const [poolId, setPoolId] = useState("pvnaive-primary");
  const [weight, setWeight] = useState("1");
  const [validityMinutes, setValidityMinutes] = useState("2880");
  const [endpoints, setEndpoints] = useState<EndpointInput[]>([{ host: "", port: 443, sni: "" }]);

  // Manifest viewer state.
  const [manifestNode, setManifestNode] = useState<PoolNodeRow | null>(null);
  const [manifestLines, setManifestLines] = useState<string[]>([]);
  const [manifestRevision, setManifestRevision] = useState<number | null>(null);
  const [manifestSignature, setManifestSignature] = useState("");

  const usable = useMemo(() => nodes.filter((node) => node.maintenance !== "disabled"), [nodes]);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      const payload = await getJSON("/api/v1/pool/nodes");
      setNodes(normalizePoolNodes(payload).nodes);
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "خواندن فهرست گره‌ها ناموفق بود." });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  useEffect(() => {
    if (!toast) return;
    const timer = window.setTimeout(() => setToast(null), 6000);
    return () => window.clearTimeout(timer);
  }, [toast]);

  async function issueToken(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setToast(null);
    const name = tokenName.trim();
    if (name.length < 1 || name.length > 120) {
      setToast({ kind: "error", text: "نام گره برای توکن ثبت‌نام باید ۱ تا ۱۲۰ نویسه باشد." });
      return;
    }
    setBusy(true);
    try {
      const payload = await callAPI("/api/v1/pool/enrollment-tokens", "POST", { node_name: name, ttl_minutes: tokenTTL });
      const token = normalizeEnrollToken(payload);
      if (!token) {
        setToast({ kind: "error", text: "پاسخ سرور فاقد توکن معتبر بود." });
        return;
      }
      setIssuedToken(token);
      setEnrollToken(token.token);
      setEnrollName(name);
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "صدور توکن انجام نشد." });
    } finally {
      setBusy(false);
    }
  }

  async function submitEnroll(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setToast(null);
    const problem = validateEnrollInput({ token: enrollToken, displayName: enrollName, region: enrollRegion, capacityWeight: enrollWeight });
    if (problem) {
      setToast({ kind: "error", text: problem });
      return;
    }
    setBusy(true);
    try {
      await callAPI("/api/v1/pool/nodes", "POST", {
        token: enrollToken.trim(),
        display_name: enrollName.trim(),
        region: enrollRegion.trim(),
        capacity_weight: enrollWeight,
      });
      setToast({ kind: "success", text: `گره «${enrollName.trim()}» ثبت‌نام شد.` });
      setEnrollToken("");
      setEnrollName("");
      setEnrollRegion("");
      setEnrollWeight(1);
      await reload();
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "ثبت‌نام گره انجام نشد." });
    } finally {
      setBusy(false);
    }
  }

  async function publishRevision(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setToast(null);
    if (!revisionNode) {
      setToast({ kind: "error", text: "گره مقصد را برای انتشار وضعیت مطلوب انتخاب کنید." });
      return;
    }
    const parsedWeight = Number(weight);
    const parsedValidity = Number(validityMinutes);
    const input = {
      poolId,
      endpoints,
      weight: parsedWeight,
      validityMinutes: parsedValidity,
    };
    const problem = validateRevisionInput(input);
    if (problem) {
      setToast({ kind: "error", text: problem });
      return;
    }
    setBusy(true);
    try {
      const payload = await callAPI(`/api/v1/pool/nodes/${encodeURIComponent(revisionNode)}/revisions`, "POST", buildRevisionPayload(input));
      const revision = typeof payload.revision === "number" ? payload.revision : null;
      setToast({ kind: "success", text: `وضعیت مطلوب منتشر شد${revision !== null ? ` (نسخه ${fmtNum(revision)})` : ""}. عامل گره آن را در چرخه بعدی دریافت می‌کند.` });
      await reload();
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "انتشار وضعیت مطلوب انجام نشد." });
    } finally {
      setBusy(false);
    }
  }

  async function setMaintenance(node: PoolNodeRow, state: string) {
    setToast(null);
    setBusy(true);
    try {
      await callAPI(`/api/v1/pool/nodes/${encodeURIComponent(node.id)}/maintenance`, "POST", { state });
      const label = actionLabels[state] ?? state;
      setToast({ kind: "success", text: `وضعیت «${node.display_name}» به «${label}» تغییر کرد.` });
      await reload();
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "تغییر وضعیت گره انجام نشد." });
    } finally {
      setBusy(false);
    }
  }

  async function openManifest(node: PoolNodeRow) {
    setToast(null);
    setManifestNode(node);
    setManifestLines([]);
    setManifestRevision(null);
    setManifestSignature("");
    try {
      const payload = await getJSON(`/api/v1/pool/nodes/${encodeURIComponent(node.id)}/manifest`);
      const manifest = normalizeManifest(payload);
      if (!manifest) {
        setToast({ kind: "error", text: "مانیفست ذخیره‌شده قابل خواندن نبود." });
        return;
      }
      setManifestRevision(manifest.revision);
      setManifestSignature(manifest.signature);
      setManifestLines(manifestSummary(manifest.manifest));
    } catch (cause) {
      setToast({ kind: "error", text: cause instanceof Error ? cause.message : "خواندن مانیفست ناموفق بود." });
    }
  }

  function updateEndpoint(index: number, patch: Partial<EndpointInput>) {
    setEndpoints((current) => current.map((endpoint, i) => (i === index ? { ...endpoint, ...patch } : endpoint)));
  }

  const stats = useMemo(() => ({
    total: nodes.length,
    inSync: nodes.filter((node) => node.drift === "in_sync" && node.maintenance === "active").length,
    pending: nodes.filter((node) => node.drift === "pending").length,
    drainingOrDisabled: nodes.filter((node) => node.maintenance !== "active").length,
  }), [nodes]);

  return (
    <div className="runtime-page pool-page" dir="rtl">
      <div className="runtime-heading">
        <div>
          <p className="eyebrow">استخر / Pool Manager</p>
          <h1>مدیریت استخر گره‌ها</h1>
          <p className="runtime-muted">ثبت‌نام گره‌های برادر، انتشار وضعیت مطلوب امضاشده و گردش تخلیه — مدل pull؛ گره‌ها خودشان وضعیت را دریافت می‌کنند.</p>
        </div>
        <div className="runtime-heading-actions">
          <button className="button-secondary" onClick={() => void reload()} disabled={busy || loading}>↻ بروزرسانی</button>
        </div>
      </div>

      {toast && <div className={toast.kind === "error" ? "runtime-alert error" : "runtime-alert success"} role={toast.kind === "error" ? "alert" : "status"}>{toast.text}</div>}

      <section className="runtime-stats" aria-label="خلاصه استخر">
        <article><span>کل گره‌ها</span><strong>{fmtNum(stats.total)}</strong></article>
        <article><span>فعال و هم‌گام</span><strong>{fmtNum(stats.inSync)}</strong></article>
        <article><span>در انتظار اعمال نسخه</span><strong>{fmtNum(stats.pending)}</strong></article>
        <article><span>در تخلیه/غیرفعال</span><strong>{fmtNum(stats.drainingOrDisabled)}</strong></article>
      </section>

      <section className="runtime-card">
        <div className="runtime-section-title">
          <div><p className="eyebrow">مرحله ۱</p><h2>صدور توکن ثبت‌نام</h2></div>
        </div>
        <p className="runtime-muted pool-hint">توکن فقط یک‌بار نمایش داده می‌شود و فقط هش آن ذخیره می‌گردد. عامل روی گره برادر با همین توکن خود را ثبت‌نام می‌کند.</p>
        <form className="runtime-form pool-form" onSubmit={issueToken}>
          <label>نام گره<input value={tokenName} onChange={(event) => setTokenName(event.target.value)} placeholder="مثلاً frankfurt-edge" maxLength={120} /></label>
          <label>اعتبار توکن
            <select value={tokenTTL} onChange={(event) => setTokenTTL(Number(event.target.value))}>
              {TTL_OPTIONS.map((option) => <option key={option.minutes} value={option.minutes}>{option.label}</option>)}
            </select>
          </label>
          <button className="primary-action" disabled={busy}>صدور توکن</button>
        </form>
      </section>

      <section className="runtime-card">
        <div className="runtime-section-title">
          <div><p className="eyebrow">مرحله ۲</p><h2>ثبت‌نام گره</h2></div>
        </div>
        <form className="runtime-form pool-form pool-form-wide" onSubmit={submitEnroll}>
          <label>توکن ثبت‌نام<input className="mono" value={enrollToken} onChange={(event) => setEnrollToken(event.target.value)} placeholder="توکن صادرشده مرحله ۱" autoComplete="off" /></label>
          <label>نام نمایشی<input value={enrollName} onChange={(event) => setEnrollName(event.target.value)} maxLength={120} placeholder="نام خوانا برای گره" /></label>
          <label>ناحیه (اختیاری)<input value={enrollRegion} onChange={(event) => setEnrollRegion(event.target.value)} maxLength={60} placeholder="مثلاً eu-central" /></label>
          <label>وزن ظرفیت<input type="number" min={1} max={10000} value={enrollWeight} onChange={(event) => setEnrollWeight(Number(event.target.value))} /></label>
          <button className="primary-action" disabled={busy}>ثبت‌نام گره</button>
        </form>
      </section>

      <section className="runtime-card">
        <div className="runtime-section-title">
          <div><p className="eyebrow">مرحله ۳</p><h2>انتشار وضعیت مطلوب (مانیفست امضاشده)</h2></div>
        </div>
        <p className="runtime-muted pool-hint">هر انتشار یک نسخه یک‌نواخت جدید می‌سازد؛ عامل گره نسخه را با اعتبارسنجی امضا دریافت می‌کند. حداکثر ۳۲ نقطه اتصال.</p>
        <form className="pool-revision-form" onSubmit={publishRevision}>
          <div className="runtime-form pool-form">
            <label>گره مقصد
              <select value={revisionNode} onChange={(event) => setRevisionNode(event.target.value)}>
                <option value="">— انتخاب گره —</option>
                {usable.map((node) => <option key={node.id} value={node.id}>{node.display_name}</option>)}
              </select>
            </label>
            <label>شناسه استخر<input value={poolId} onChange={(event) => setPoolId(event.target.value)} maxLength={60} /></label>
            <label>وزن (۰٫۱ تا ۱۰۰۰)<input type="number" step="0.1" min={0.1} max={1000} value={weight} onChange={(event) => setWeight(event.target.value)} /></label>
            <label>اعتبار (دقیقه؛ ۶۰ تا ۲۰۱۶۰)<input type="number" min={60} max={20160} value={validityMinutes} onChange={(event) => setValidityMinutes(event.target.value)} /></label>
          </div>
          <fieldset className="pool-endpoints">
            <legend>نقاط اتصال</legend>
            {endpoints.map((endpoint, index) => (
              <div className="pool-endpoint-row" key={index}>
                <input value={endpoint.host} onChange={(event) => updateEndpoint(index, { host: event.target.value })} placeholder="میزبان (مثلاً edge.example.net)" maxLength={255} aria-label={`میزبان نقطه ${index + 1}`} />
                <input type="number" min={1} max={65535} value={endpoint.port} onChange={(event) => updateEndpoint(index, { port: Number(event.target.value) })} aria-label={`پورت نقطه ${index + 1}`} />
                <input value={endpoint.sni} onChange={(event) => updateEndpoint(index, { sni: event.target.value })} placeholder="SNI (اختیاری)" maxLength={255} aria-label={`SNI نقطه ${index + 1}`} />
                <button type="button" className="button-secondary pool-endpoint-remove" onClick={() => setEndpoints((current) => (current.length > 1 ? current.filter((_, i) => i !== index) : current))} aria-label={`حذف نقطه ${index + 1}`}><Icon name="close" size={14} /></button>
              </div>
            ))}
            <div className="pool-endpoint-actions">
              <button type="button" className="button-secondary" onClick={() => setEndpoints((current) => (current.length < 32 ? [...current, { host: "", port: 443, sni: "" }] : current))} disabled={endpoints.length >= 32}><Icon name="plus" size={14} /> افزودن نقطه</button>
              <button className="primary-action" disabled={busy}>انتشار نسخه جدید</button>
            </div>
          </fieldset>
        </form>
      </section>

      <section className="runtime-card">
        <div className="runtime-section-title">
          <div><p className="eyebrow">فهرست گره‌ها</p><h2>وضعیت زنده استخر</h2></div>
          <span className="badge">{fmtNum(nodes.length)} گره</span>
        </div>
        {loading ? (
          <p className="runtime-muted">در حال خواندن فهرست گره‌ها…</p>
        ) : nodes.length === 0 ? (
          <p className="runtime-muted">هنوز گره‌ای ثبت‌نام نکرده است. از مرحله ۱ شروع کنید.</p>
        ) : (
          <div className="runtime-table-wrap">
            <table className="runtime-table pool-table">
              <thead>
                <tr>
                  <th>گره</th>
                  <th>ناحیه</th>
                  <th>وزن</th>
                  <th>سلامت</th>
                  <th>هم‌گامی</th>
                  <th>نگهداری</th>
                  <th>نسخه‌ها</th>
                  <th>آخرین دیدار</th>
                  <th>عملیات</th>
                </tr>
              </thead>
              <tbody>
                {nodes.map((node) => (
                  <tr key={node.id}>
                    <td><strong>{node.display_name}</strong><span className="runtime-revision mono">{node.id.slice(0, 8)}</span></td>
                    <td>{node.region || "—"}</td>
                    <td>{fmtNum(node.capacity_weight)}</td>
                    <td><span className={`status-pill ${healthPillClass[node.health] ?? "pill-muted"}`}>{healthLabel(node.health)}</span></td>
                    <td><span className={`status-pill ${driftPillClass[node.drift] ?? "pill-muted"}`}>{driftLabel(node.drift)}</span></td>
                    <td><span className={`status-pill ${maintenancePillClass[node.maintenance] ?? "pill-muted"}`}>{maintenanceLabel(node.maintenance)}</span></td>
                    <td><span className="mono">مطلوب {node.desired_revision} / اعمال‌شده {node.applied_revision}</span></td>
                    <td>{formatLastSeen(node.last_seen_at)}</td>
                    <td>
                      <div className="pool-row-actions">
                        {maintenanceActions(node.maintenance).map((target) => (
                          <button key={target} className={target === "disabled" ? "button-secondary danger-action" : "button-secondary"} disabled={busy} onClick={() => void setMaintenance(node, target)}>{actionLabels[target]}</button>
                        ))}
                        <button className="button-secondary" disabled={busy} onClick={() => void openManifest(node)}>مانیفست</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <p className="runtime-muted pool-hint">قانون ایمن استخر: گره فعال مستقیماً غیرفعال نمی‌شود؛ ابتدا «شروع تخلیه» و سپس «تکمیل خروج». بازگشت به سرویس از حالت تخلیه یا غیرفعال مجاز است.</p>
      </section>

      {manifestNode && (
        <div className="secret-backdrop" role="dialog" aria-modal="true" aria-label={`مانیفست گره ${manifestNode.display_name}`} onClick={() => setManifestNode(null)}>
          <div className="secret-dialog" onClick={(event) => event.stopPropagation()}>
            <div className="runtime-section-title">
              <div><p className="eyebrow">مانیفست امضاشده</p><h2>گره «{manifestNode.display_name}»</h2></div>
              <button className="button-secondary" onClick={() => setManifestNode(null)} aria-label="بستن"><Icon name="close" size={15} /></button>
            </div>
            {manifestRevision !== null ? (
              <>
                <p className="runtime-muted">نسخه {fmtNum(manifestRevision)} — سرور امضا را پیش از ارائه تأیید کرده است.</p>
                <ul className="pool-manifest-list">
                  {manifestLines.map((line, index) => <li key={index}>{line}</li>)}
                </ul>
                <details className="pool-signature">
                  <summary>امضا (base64)</summary>
                  <code className="mono">{manifestSignature}</code>
                </details>
              </>
            ) : (
              <p className="runtime-muted">در حال خواندن مانیفست…</p>
            )}
          </div>
        </div>
      )}

      {issuedToken && (
        <div className="secret-backdrop" role="dialog" aria-modal="true" aria-label="توکن یک‌بارمصرف ثبت‌نام">
          <div className="secret-dialog" onClick={(event) => event.stopPropagation()}>
            <div className="runtime-section-title">
              <div><p className="eyebrow">فقط همین یک‌بار</p><h2>توکن ثبت‌نام «{tokenName.trim()}»</h2></div>
            </div>
            <p className="runtime-muted">این توکن دوباره نمایش داده نمی‌شود؛ الان کپی کنید. انقضا: {new Date(issuedToken.expires_at).toLocaleString("fa-IR")}</p>
            <code className="mono">{issuedToken.token}</code>
            <div className="secret-actions">
              <button className="primary-action" onClick={() => { void navigator.clipboard?.writeText(issuedToken.token); setToast({ kind: "success", text: "توکن در حافظه کپی شد." }); }}>کپی توکن</button>
              <button className="button-secondary" onClick={() => setIssuedToken(null)}>ذخیره کردم؛ بستن</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
