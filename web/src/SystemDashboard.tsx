import { useEffect, useMemo, useRef, useState } from "react";
import { connectSystemStream } from "./systemStream";
import {
  dependencyEntries,
  fetchSystemStatus,
  formatBytes,
  formatRate,
  formatUptime,
  normalizeSystemStatus,
  SystemStatus,
} from "./systemStatus";
import { LiveAreaChart, RadialGauge } from "./charts";
import { appendLivePoint, historyPointFromStatus, type LivePoint } from "./systemLive";

/* Live server telemetry console. Frames arrive over the R8 SSE stream
   (1s server ticks); a polling fallback keeps data honest if the stream
   drops. Every rendered number originates from a real server sample —
   nothing here is synthesized in the browser. */

type HistoryPoint = LivePoint;
type StreamMode = "connecting" | "live" | "polling";

const POLL_FALLBACK_MS = 5000;
const STREAM_RETRY_MS = 10000;

function percentText(value: number): string {
  return Number.isFinite(value) ? `${value.toLocaleString("fa-IR", { maximumFractionDigits: 1 })}٪` : "—";
}

export function SystemDashboard() {
  const [status, setStatus] = useState<SystemStatus | null>(null);
  const [history, setHistory] = useState<HistoryPoint[]>([]);
  const [error, setError] = useState("");
  const [mode, setMode] = useState<StreamMode>("connecting");
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null);
  const modeRef = useRef<StreamMode>("connecting");

  useEffect(() => {
    let active = true;
    let pollTimer = 0;
    let retryTimer = 0;
    let closeStream: (() => void) | null = null;

    const applySample = (next: SystemStatus) => {
      if (!active) return;
      setStatus(next);
      setUpdatedAt(new Date());
      setError("");
      setHistory((current) => appendLivePoint(current, historyPointFromStatus(next)));
    };

    const setModeSafe = (value: StreamMode) => {
      modeRef.current = value;
      if (active) setMode(value);
    };

    const startPolling = () => {
      if (modeRef.current === "polling") return;
      setModeSafe("polling");
      const poll = async () => {
        try {
          applySample(await fetchSystemStatus());
        } catch {
          if (active) setError("خواندن وضعیت سرور ناموفق بود؛ آخرین نمونه معتبر نگه داشته شده است.");
        } finally {
          if (active) pollTimer = window.setTimeout(poll, POLL_FALLBACK_MS);
        }
      };
      void poll();
    };

    const stopPolling = () => { window.clearTimeout(pollTimer); };

    const connect = () => {
      if (!active) return;
      setModeSafe("connecting");
      closeStream = connectSystemStream(
        {
          onStatus: (payload) => {
            try {
              applySample(normalizeSystemStatus(payload));
              stopPolling();
              setModeSafe("live");
            } catch { /* malformed frame: skip */ }
          },
          onError: () => {
            if (!active) return;
            closeStream?.();
            closeStream = null;
            startPolling();
            retryTimer = window.setTimeout(() => { if (active) connect(); }, STREAM_RETRY_MS);
          },
        },
        { url: "/api/v1/system/stream?interval=1s" },
      );
    };

    connect();
    return () => {
      active = false;
      closeStream?.();
      stopPolling();
      window.clearTimeout(retryTimer);
    };
  }, []);

  const series = useMemo(() => [
    { name: "دریافت", color: "var(--blue)", values: history.map((p) => p.rx) },
    { name: "ارسال", color: "var(--gold)", values: history.map((p) => p.tx) },
  ], [history]);

  if (!status) {
    return <section className="dashboard-card system-monitor system-fallback" aria-live="polite">
      <div className="monitor-card-head">
        <div><p className="eyebrow">Server telemetry</p><h2>مانیتورینگ زنده سرور</h2></div>
        <span className="live-pill" data-mode="connecting"><i />در حال اتصال…</span>
      </div>
      <p className="muted">{error || "در حال دریافت نمونه واقعی از سرور…"}</p>
    </section>;
  }

  const sample = status.sample;
  const memoryUsedBytes = sample.memory_total_bytes - sample.memory_available_bytes;
  const diskUsedBytes = sample.disk_total_bytes - sample.disk_available_bytes;
  const latest = history[history.length - 1];

  return <section className="dashboard-card system-monitor" aria-label="مانیتورینگ واقعی سرور">
    <div className="monitor-card-head">
      <div><p className="eyebrow">Server telemetry · R8 stream</p><h2>مانیتورینگ زنده سرور</h2></div>
      <div className="monitor-head-side">
        <span className="live-pill" data-mode={mode}><i />{mode === "live" ? "زنده · هر ۱ ثانیه" : mode === "polling" ? "نمونه‌برداری دوره‌ای" : "در حال اتصال…"}</span>
        <div className="dependency-row">
          {dependencyEntries(status.dependencies).map((item) => <span key={item.label} className={`dependency ${item.status === "ok" ? "ok" : "bad"}`}><i />{item.label}: {item.status === "ok" ? "OK" : "Unavailable"}</span>)}
        </div>
      </div>
    </div>
    {error && <div className="system-warning" role="alert">{error} آخرین نمونه معتبر نگه داشته شده است.</div>}

    <div className="monitor-gauges">
      <RadialGauge percent={sample.cpu_percent} label="پردازنده" valueText={percentText(sample.cpu_percent)} caption={`load ${sample.load_1.toLocaleString("fa-IR", { maximumFractionDigits: 2 })}`} />
      <RadialGauge percent={sample.memory_used_percent} label="حافظه" valueText={percentText(sample.memory_used_percent)} caption={`${formatBytes(memoryUsedBytes)} / ${formatBytes(sample.memory_total_bytes)}`} />
      <RadialGauge percent={sample.disk_used_percent} label="دیسک" valueText={percentText(sample.disk_used_percent)} caption={`${formatBytes(diskUsedBytes)} / ${formatBytes(sample.disk_total_bytes)}`} />
      <div className="monitor-uptime">
        <span className="monitor-uptime-label">آپ‌تایم سرور</span>
        <strong>{formatUptime(sample.uptime_seconds)}</strong>
        <div className="monitor-uptime-meta">
          <div><span>Load 1/5/15</span><b>{sample.load_1.toLocaleString("fa-IR", { maximumFractionDigits: 2 })} · {sample.load_5.toLocaleString("fa-IR", { maximumFractionDigits: 2 })} · {sample.load_15.toLocaleString("fa-IR", { maximumFractionDigits: 2 })}</b></div>
          <div><span>اینترفیس</span><b>{sample.network_interface || "—"}</b></div>
          <div><span>معناشناسی ترافیک</span><b className="system-semantics">{status.traffic_semantics}</b></div>
        </div>
      </div>
    </div>

    <div className="monitor-net">
      <div className="monitor-net-head">
        <h3>ترافیک شبکه</h3>
        <div className="monitor-net-now">
          <div className="net-stat rx"><i />دریافت<b>{formatRate(latest?.rx ?? sample.rx_bytes_per_second, sample.rate_available)}</b></div>
          <div className="net-stat tx"><i />ارسال<b>{formatRate(latest?.tx ?? sample.tx_bytes_per_second, sample.rate_available)}</b></div>
        </div>
      </div>
      <LiveAreaChart series={series} height={190} formatValue={(v) => formatRate(v, true)} ariaLabel="نمودار زنده ترافیک شبکه" />
      <p className="monitor-net-note">نرخ از اختلاف counter و timestamp سمت سرور محاسبه می‌شود؛ مرورگر عددی حدس نمی‌زند. {sample.rate_available ? `پنجره نمونه: ${sample.sample_window_seconds.toLocaleString("fa-IR", { maximumFractionDigits: 1 })} ثانیه` : ""}</p>
    </div>

    <p className="sample-meta">آخرین نمونه معتبر: {updatedAt?.toLocaleTimeString("fa-IR") || "—"} · server sample: {new Date(sample.sampled_at).toLocaleTimeString("fa-IR")}</p>
  </section>;
}
