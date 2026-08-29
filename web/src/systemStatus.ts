export type DependencyStatus = { status: "ok" | "unavailable" | string };

export type SystemSample = {
  sampled_at: string;
  cpu_percent: number;
  memory_total_bytes: number;
  memory_available_bytes: number;
  memory_used_percent: number;
  disk_total_bytes: number;
  disk_available_bytes: number;
  disk_used_percent: number;
  load_1: number;
  load_5: number;
  load_15: number;
  uptime_seconds: number;
  rx_bytes: number;
  tx_bytes: number;
  network_interface: string;
  rate_available: boolean;
  sample_window_seconds: number;
  rx_bytes_per_second: number;
  tx_bytes_per_second: number;
};

export type SystemStatus = {
  sample: SystemSample;
  dependencies: {
    api: DependencyStatus;
    database: DependencyStatus;
    runtime: DependencyStatus;
  };
  traffic_semantics: string;
};

type Fetcher = typeof fetch;

export async function fetchSystemStatus(fetcher: Fetcher = fetch): Promise<SystemStatus> {
  const response = await fetcher("/api/v1/system/status", {
    method: "GET",
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) throw new Error(`system status HTTP ${response.status}`);
  const body = await response.json() as {
    metrics?: { sample?: SystemSample; dependencies?: SystemStatus["dependencies"] };
    traffic_semantics?: string;
  };
  if (!body.metrics?.sample || !body.metrics.dependencies) throw new Error("invalid system status response");
  return {
    sample: body.metrics.sample,
    dependencies: body.metrics.dependencies,
    traffic_semantics: body.traffic_semantics || "unknown",
  };
}

export function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "—";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let scaled = value;
  let index = 0;
  while (scaled >= 1024 && index < units.length - 1) {
    scaled /= 1024;
    index++;
  }
  const digits = index < 2 ? 0 : scaled >= 100 ? 0 : scaled >= 10 ? 1 : 2;
  return `${scaled.toFixed(digits)} ${units[index]}`;
}

export function formatRate(value: number, available: boolean): string {
  return available ? `${formatBytes(value)}/s` : "در حال نمونه‌گیری…";
}

export function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "—";
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  return days > 0 ? `${days} روز ${hours} ساعت` : `${hours} ساعت`;
}
