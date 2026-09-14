import { useEffect, useMemo, useState } from "react";
import { dependencyEntries, fetchSystemStatus, formatBytes, formatRate, formatUptime, normalizeSystemStatus, SystemStatus } from "./systemStatus";
import { connectSystemStream } from "./systemStream";
import { appendLivePoint, livePointFromStatus, LivePoint } from "./systemLive";

type HistoryPoint = LivePoint;

function percent(value: number): string {
  return Number.isFinite(value) ? `${value.toLocaleString("fa-IR", { maximumFractionDigits: 1 })}%` : "—";
}

function SparkBars({ values, max }: { values: Array<number | null>; max: number }) {
  const finite = values.filter((value): value is number => value !== null);
  const safeMax = Math.max(max, ...finite, 1);
  return <div className="system-spark" aria-hidden="true">{values.map((value, index) => value === null ? <i key={index} className="gap" /> : <i key={index} style={{ height: `${Math.max(4, Math.min(100, value / safeMax * 100))}%` }} />)}</div>;
}

export function SystemDashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [history, setHistory] = useState<HistoryPoint[]>([]);
  const [error, setError] = useState("");
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);

  useEffect(() => {
    let active = true;

    const accept = (next: SystemStatus) => {
      if (!active) return;
      setStatus(next);
      setUpdatedAt(new Date());
      setError("");
      setHistory((current) => appendLivePoint(current, livePointFromStatus(next)));
    };

    // Bootstrap once from the ordinary endpoint so the dashboard still paints
    // when EventSource startup is delayed. Steady-state updates come from SSE.
    void fetchSystemStatus().then(accept).catch(() => {
      if (active) setError("اتصال زنده در حال برقراری است؛ داده ساختگی نمایش داده نمی‌شود.");
    });

    const close = connectSystemStream({
      onStatus: (payload) => {
        try { accept(normalizeSystemStatus(payload)); } catch {
          if (active) setError("نمونه نامعتبر از جریان زنده رد شد؛ آخرین داده معتبر حفظ شده است.");
        }
      },
      onError: () => {
        if (active) setError("جریان زنده قطع شده است؛ مرورگر به‌صورت خودکار دوباره متصل می‌شود.");
      },
    });

    return () => { active = false; close(); };
  }, []);

  const networkValues = useMemo(() => history.flatMap((item) => [item.rx, item.tx]).filter((value): value is number => value !== null), [history]);
  const networkMax = useMemo(() => Math.max(1, ...networkValues), [networkValues]);

  if (!status) {
    return <section className="dashboard-card system-monitor system-fallback" aria-live="polite">
      <div><p className="eyebrow">Server telemetry</p><h2>وضعیت زنده سرور</h2></div>
      <p className="muted">{error || "در حال دریافت نمونه واقعی از سرور…"}</p>
    </section>;
  }

  const sample = status.sample;
  return <section className="dashboard-card system-monitor" aria-label="مانیتورینگ واقعی سرور">
    <div className="system-heading">
      <div><p className="eyebrow">Server telemetry</p><h2>وضعیت زنده سرور</h2><p className="muted">نرخ شبکه از اختلاف counter و timestamp سمت سرور محاسبه می‌شود؛ مرورگر عددی حدس نمی‌زند.</p></div>
      <div className="dependency-row">{dependencyEntries(status.dependencies).map((item) => <span key={item.label} className={`dependency ${item.status === "ok" ? "ok" : "bad"}`}><i />{item.label}: {item.status === "ok" ? "OK" : "Unavailable"}</span>)}</div>
    </div>
    {error && <div className="system-warning" role="alert">{error} آخرین نمونه معتبر نگه داشته شده است.</div>}
    <div className="system-kpis">
      <article><span>CPU</span><strong>{percent(sample.cpu_percent)}</strong><SparkBars values={history.map((item) => item.cpu)} max={100}/></article>
      <article><span>RAM</span><strong>{percent(sample.memory_used_percent)}</strong><small>{formatBytes(sample.memory_total_bytes - sample.memory_available_bytes)} / {formatBytes(sample.memory_total_bytes)}</small><SparkBars values={history.map((item) => item.memory)} max={100}/></article>
      <article><span>Disk /</span><strong>{percent(sample.disk_used_percent)}</strong><small>{formatBytes(sample.disk_total_bytes - sample.disk_available_bytes)} / {formatBytes(sample.disk_total_bytes)}</small></article>
      <article><span>Load 1/5/15</span><strong>{sample.load_1.toLocaleString("fa-IR", { maximumFractionDigits: 2 })}</strong><small>{sample.load_5.toLocaleString("fa-IR", { maximumFractionDigits: 2 })} · {sample.load_15.toLocaleString("fa-IR", { maximumFractionDigits: 2 })}</small></article>
      <article><span>RX Rate</span><strong>{formatRate(sample.rx_bytes_per_second, sample.rate_available)}</strong><small>{sample.network_interface || "interface unavailable"}</small><SparkBars values={history.map((item) => item.rx)} max={networkMax}/></article>
      <article><span>TX Rate</span><strong>{formatRate(sample.tx_bytes_per_second, sample.rate_available)}</strong><small>{sample.rate_available ? `window ${sample.sample_window_seconds.toLocaleString("fa-IR", { maximumFractionDigits: 1 })}s` : "window —"}</small><SparkBars values={history.map((item) => item.tx)} max={networkMax}/></article>
      <article><span>Uptime</span><strong>{formatUptime(sample.uptime_seconds)}</strong></article>
      <article><span>Traffic semantics</span><strong className="system-semantics">{status.traffic_semantics}</strong><small>Accounting/Online در این کارت ساخته یا تخمین زده نمی‌شود.</small></article>
    </div>
    <p className="sample-meta">جریان SSE · آخرین نمونه معتبر: {updatedAt?.toLocaleTimeString("fa-IR") || "—"} · server sample: {new Date(sample.sampled_at).toLocaleTimeString("fa-IR")}</p>
  </section>;
}
