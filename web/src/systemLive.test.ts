import { describe, expect, it } from "vitest";
import { appendLivePoint, livePointFromStatus, LIVE_HISTORY_LIMIT } from "./systemLive";
import type { LivePoint } from "./systemLive";
import type { SystemStatus } from "./systemStatus";

function status(rateAvailable = true): SystemStatus {
  return {
    sample: {
      sampled_at: "2026-09-14T19:00:00Z",
      cpu_percent: 10,
      memory_total_bytes: 1000,
      memory_available_bytes: 400,
      memory_used_percent: 60,
      disk_total_bytes: 2000,
      disk_available_bytes: 1000,
      disk_used_percent: 50,
      load_1: 0.1, load_5: 0.2, load_15: 0.3,
      uptime_seconds: 123,
      rx_bytes: 100, tx_bytes: 200,
      network_interface: "eth0",
      rate_available: rateAvailable,
      sample_window_seconds: 1,
      rx_bytes_per_second: 25,
      tx_bytes_per_second: 40,
    },
    dependencies: {},
    traffic_semantics: "server_counter_delta",
  };
}

describe("system live history", () => {
  it("keeps unavailable network rates as unknown instead of inventing zero", () => {
    expect(livePointFromStatus(status(false))).toEqual({ cpu: 10, memory: 60, rx: null, tx: null });
  });

  it("drops the oldest samples when the bounded history is full", () => {
    let history: LivePoint[] = Array.from({ length: LIVE_HISTORY_LIMIT }, (_, index) => ({ cpu: index, memory: index, rx: index, tx: index }));
    history = appendLivePoint(history, { cpu: 999, memory: 999, rx: 999, tx: 999 });
    expect(history).toHaveLength(LIVE_HISTORY_LIMIT);
    expect(history[0]?.cpu).toBe(1);
    expect(history.at(-1)?.cpu).toBe(999);
  });
});