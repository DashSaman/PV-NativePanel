import type { SystemStatus } from "./systemStatus";

export const LIVE_HISTORY_LIMIT = 60;

export type LivePoint = {
  cpu: number;
  memory: number;
  rx: number | null;
  tx: number | null;
};

export function livePointFromStatus(status: SystemStatus): LivePoint {
  const sample = status.sample;
  return {
    cpu: sample.cpu_percent,
    memory: sample.memory_used_percent,
    rx: sample.rate_available ? sample.rx_bytes_per_second : null,
    tx: sample.rate_available ? sample.tx_bytes_per_second : null,
  };
}

export function appendLivePoint(history: LivePoint[], point: LivePoint): LivePoint[] {
  return [...history, point].slice(-LIVE_HISTORY_LIMIT);
}
export type LiveStreamState = "connecting" | "live" | "disconnected";

export function liveStreamLabel(state: LiveStreamState): string {
  switch (state) {
    case "live": return "جریان SSE زنده";
    case "disconnected": return "SSE قطع‌شده";
    default: return "SSE در حال اتصال";
  }
}
