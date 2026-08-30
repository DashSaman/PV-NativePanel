export type DependencyStatus = { status: string };
export type SystemDependencies = {
  api?: DependencyStatus;
  database?: DependencyStatus;
  runtime?: DependencyStatus;
  telemetry?: DependencyStatus;
};

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
  dependencies: SystemDependencies;
  traffic_semantics: string;
};

type Fetcher = typeof fetch;

function object(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("Invalid system status response.");
  return value as Record<string, unknown>;
}

function numberField(value: unknown, field: string): number {
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) throw new Error(`Invalid system metric: ${field}`);
  return value;
}

function dependency(value: unknown): DependencyStatus | undefined {
  if (!value || typeof value !== "object" || Array.isArray(value)) return undefined;
  const status = (value as Record<string, unknown>).status;
  return typeof status === "string" ? { status } : undefined;
}

export function normalizeSystemStatus(value: unknown): SystemStatus {
  const root = object(value);
  const metrics = object(root.metrics);
  const rawSample = object(metrics.sample);
  const rawDependencies = object(metrics.dependencies);
  const sampledAt = rawSample.sampled_at;
  if (typeof sampledAt !== "string" || !sampledAt) throw new Error("Invalid system metric: sampled_at");
  const networkInterface = rawSample.network_interface;
  if (typeof networkInterface !== "string") throw new Error("Invalid system metric: network_interface");
  const rateAvailable = rawSample.rate_available;
  if (typeof rateAvailable !== "boolean") throw new Error("Invalid system metric: rate_available");

  const sample: SystemSample = {
    sampled_at: sampledAt,
    cpu_percent: numberField(rawSample.cpu_percent, "cpu_percent"),
    memory_total_bytes: numberField(rawSample.memory_total_bytes, "memory_total_bytes"),
    memory_available_bytes: numberField(rawSample.memory_available_bytes, "memory_available_bytes"),
    memory_used_percent: numberField(rawSample.memory_used_percent, "memory_used_percent"),
    disk_total_bytes: numberField(rawSample.disk_total_bytes, "disk_total_bytes"),
    disk_available_bytes: numberField(rawSample.disk_available_bytes, "disk_available_bytes"),
    disk_used_percent: numberField(rawSample.disk_used_percent, "disk_used_percent"),
    load_1: numberField(rawSample.load_1, "load_1"),
    load_5: numberField(rawSample.load_5, "load_5"),
    load_15: numberField(rawSample.load_15, "load_15"),
    uptime_seconds: numberField(rawSample.uptime_seconds, "uptime_seconds"),
    rx_bytes: numberField(rawSample.rx_bytes, "rx_bytes"),
    tx_bytes: numberField(rawSample.tx_bytes, "tx_bytes"),
    network_interface: networkInterface,
    rate_available: rateAvailable,
    sample_window_seconds: numberField(rawSample.sample_window_seconds, "sample_window_seconds"),
    rx_bytes_per_second: numberField(rawSample.rx_bytes_per_second, "rx_bytes_per_second"),
    tx_bytes_per_second: numberField(rawSample.tx_bytes_per_second, "tx_bytes_per_second"),
  };

  const semantics = root.traffic_semantics;
  if (typeof semantics !== "string" || !semantics) throw new Error("Invalid traffic semantics.");

  return {
    sample,
    dependencies: {
      api: dependency(rawDependencies.api),
      database: dependency(rawDependencies.database),
      runtime: dependency(rawDependencies.runtime),
      telemetry: dependency(rawDependencies.telemetry),
    },
    traffic_semantics: semantics,
  };
}

export function dependencyEntries(dependencies: SystemDependencies): Array<{ label: string; status: string }> {
  return [
    { label: "API", status: dependencies.api?.status ?? "unavailable" },
    { label: "DB", status: dependencies.database?.status ?? "unavailable" },
    { label: "Runtime", status: dependencies.runtime?.status ?? "unavailable" },
    { label: "Telemetry", status: dependencies.telemetry?.status ?? "unavailable" },
  ];
}

export function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) return "—";
  if (value >= 1_073_741_824) return `${(value / 1_073_741_824).toLocaleString("fa-IR", { maximumFractionDigits: 2 })} GB`;
  if (value >= 1_048_576) return `${(value / 1_048_576).toLocaleString("fa-IR", { maximumFractionDigits: 1 })} MB`;
  if (value >= 1024) return `${(value / 1024).toLocaleString("fa-IR", { maximumFractionDigits: 1 })} KB`;
  return `${value.toLocaleString("fa-IR")} B`;
}

export function formatRate(value: number, available: boolean): string {
  if (!available || !Number.isFinite(value) || value < 0) return "—";
  return `${formatBytes(value)}/s`;
}

export function formatUptime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "—";
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  if (days > 0) return `${days.toLocaleString("fa-IR")} روز ${hours.toLocaleString("fa-IR")} ساعت`;
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${hours.toLocaleString("fa-IR")} ساعت ${minutes.toLocaleString("fa-IR")} دقیقه`;
}

export async function fetchSystemStatus(fetcher: Fetcher = fetch): Promise<SystemStatus> {
  const response = await fetcher("/api/v1/system/status", { method: "GET", credentials: "same-origin" });
  const contentType = response.headers.get("Content-Type") || "";
  const body: unknown = contentType.includes("application/json") ? await response.json() : {};
  if (!response.ok) throw new Error("وضعیت زنده سرور در دسترس نیست.");
  return normalizeSystemStatus(body);
}
