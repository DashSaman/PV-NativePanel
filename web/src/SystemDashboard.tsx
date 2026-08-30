import { useEffect, useMemo, useState } from "react";
import { dependencyEntries, fetchSystemStatus, formatBytes, formatRate, formatUptime, SystemStatus } from "./systemStatus";

type HistoryPoint = { cpu: number; memory: number; rx: number; tx: number };

function percent(value: number): string {
  return Number.isFinite(value) ? `${value.toLocaleString("fa-IR", { maximumFractionDigits: 1 })}%` : "—";
}

function SparkBars({ values, max }: { values: number[]; max: number }) {
  const safeMax = Math.max(max, ...values, 1);
  return <div className="system-spark" aria-hidden="true">{values.map((value, index) => <i key={index} style={{ height: `${Math.max(4, Math.min(100, value / safeMax * 100))}%` }} />)}</div>;
}

export function SystemDashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [history, setHistory] = useState<HistoryPoint[]>([]);
  const [error, setError] = useState("");
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);

  useEffect(() => {
    let active = true;
    let timer = 0;
    const poll = async () => {
      try {
        const next = await fetchSystemStatus();
        if (!active) return;
        setStatus(next);
        setUpdatedAt(new Date());
        setError("");
        setHistory((current) => [...current, {
          cpu: next.sample.cpu_percent,
          memory: next.sample.memory_used_percent,
          rx: next.sample.rate_available ? next.sample.rx_bytes_per_second : 0,
          tx: next.sample.rate_available ? next.sample.tx_bytes_per_second : 0,
        }].slice(-24));
      } catch {
        if (active) setError("خواندن وضعیت زنده سرور ناموفق بود؛ داده ساختگی نمایش داده نمی‌شود.");
      } finally {
        if (active) timer = window.setTimeout(poll, 5000);
      }
    };
    void poll();
    return () => { active = false; window.clearTimeout(timer); };
  }, []);

  const networkMax = useMemo(() => Math.max(1, ...history.flatMap((item) => [item.rx, item.tx])), [history]);

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
    <p className="sample-meta">آخرین نمونه معتبر: {updatedAt?.toLocaleTimeString("fa-IR") || "—"} · server sample: {new Date(sample.sampled_at).toLocaleTimeString("fa-IR")}</p>
  </section>;
}
