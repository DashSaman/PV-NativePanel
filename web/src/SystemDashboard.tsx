import { useEffect, useMemo, useState } from "react";
import { fetchSystemStatus, formatBytes, formatRate, formatUptime, SystemStatus } from "./systemStatus";
import "./system.css";

type HistoryPoint = { at: number; cpu: number; memory: number; rx: number; tx: number };

function percent(value: number): string {
  return Number.isFinite(value) ? `${value.toFixed(1)}%` : "—";
}

function Dependency({ label, status }: { label: string; status?: string }) {
  const ok = status === "ok";
  return <span className={`dependency ${ok ? "ok" : "bad"}`}><i />{label}: {ok ? "OK" : "Unavailable"}</span>;
}

function SparkBars({ values, max }: { values: number[]; max: number }) {
  const safeMax = Math.max(max, ...values, 1);
  return <div className="spark" aria-hidden="true">{values.map((value, index) => <i key={index} style={{ height: `${Math.max(4, Math.min(100, value / safeMax * 100))}%` }} />)}</div>;
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
          at: Date.now(), cpu: next.sample.cpu_percent, memory: next.sample.memory_used_percent,
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

  if (!status) return <section className="system-card system-fallback" aria-live="polite"><div><p className="eyebrow">System monitoring</p><h2>وضعیت زنده سرور</h2></div><p>{error || "در حال دریافت نمونه واقعی از سرور…"}</p></section>;

  const sample = status.sample;
  return <section className="system-monitor" aria-label="مانیتورینگ واقعی سرور">
    <div className="system-heading"><div><p className="eyebrow">Server telemetry</p><h2>وضعیت زنده سرور</h2><p>نرخ شبکه از اختلاف counter و timestamp سمت سرور محاسبه می‌شود، نه فاصله polling مرورگر.</p></div><div className="dependency-row"><Dependency label="API" status={status.dependencies.api?.status}/><Dependency label="DB" status={status.dependencies.database?.status}/><Dependency label="Runtime" status={status.dependencies.runtime?.status}/></div></div>
    {error && <div className="system-warning" role="alert">{error} آخرین نمونه معتبر نگه داشته شده است.</div>}
    <div className="system-kpis">
      <article><span>CPU</span><strong>{percent(sample.cpu_percent)}</strong><SparkBars values={history.map((item) => item.cpu)} max={100}/></article>
      <article><span>RAM</span><strong>{percent(sample.memory_used_percent)}</strong><small>{formatBytes(sample.memory_total_bytes - sample.memory_available_bytes)} / {formatBytes(sample.memory_total_bytes)}</small><SparkBars values={history.map((item) => item.memory)} max={100}/></article>
      <article><span>Disk /</span><strong>{percent(sample.disk_used_percent)}</strong><small>{formatBytes(sample.disk_total_bytes - sample.disk_available_bytes)} / {formatBytes(sample.disk_total_bytes)}</small></article>
      <article><span>Load 1/5/15</span><strong>{sample.load_1.toFixed(2)}</strong><small>{sample.load_5.toFixed(2)} · {sample.load_15.toFixed(2)}</small></article>
      <article><span>RX Rate</span><strong>{formatRate(sample.rx_bytes_per_second, sample.rate_available)}</strong><small>{sample.network_interface || "interface unavailable"}</small><SparkBars values={history.map((item) => item.rx)} max={networkMax}/></article>
      <article><span>TX Rate</span><strong>{formatRate(sample.tx_bytes_per_second, sample.rate_available)}</strong><small>window {sample.rate_available ? `${sample.sample_window_seconds.toFixed(1)}s` : "—"}</small><SparkBars values={history.map((item) => item.tx)} max={networkMax}/></article>
      <article><span>Uptime</span><strong>{formatUptime(sample.uptime_seconds)}</strong></article>
      <article><span>Accounting/Online</span><strong>در دسترس نیست</strong><small>تا اتصال capability دقیق WS1 هیچ عددی ساخته نمی‌شود.</small></article>
    </div>
    <p className="sample-meta">آخرین نمونه: {updatedAt?.toLocaleTimeString("fa-IR") || "—"} · semantics: {status.traffic_semantics}</p>
  </section>;
}
