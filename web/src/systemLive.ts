import { normalizeSystemStatus, type SystemStatus } from "./systemStatus";

export const LIVE_HISTORY_LIMIT = 90;

export type LivePoint = {
  t: number;
  cpu: number;
  memory: number;
  rx: number | null;
  tx: number | null;
};

export function normalizeLiveStatus(payload: unknown): SystemStatus {
  return normalizeSystemStatus(payload);
}

export function historyPointFromStatus(status: SystemStatus, now = Date.now()): LivePoint {
  const sample = status.sample;
  return {
    t: now,
    cpu: sample.cpu_percent,
    memory: sample.memory_used_percent,
    rx: sample.rate_available ? sample.rx_bytes_per_second : null,
    tx: sample.rate_available ? sample.tx_bytes_per_second : null,
  };
}

export function historyPointFromPayload(payload: unknown, now = Date.now()): LivePoint {
  return historyPointFromStatus(normalizeLiveStatus(payload), now);
}

export function appendLivePoint(current: LivePoint[], point: LivePoint): LivePoint[] {
  return [...current, point].slice(-LIVE_HISTORY_LIMIT);
}
