import { describe, expect, it } from "vitest";
import { fetchSystemStatus, formatBytes, formatRate, formatUptime } from "./systemStatus";

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

describe("system status API", () => {
  it("reads server-side sample/rate semantics", async () => {
    const fetcher = (async (path: RequestInfo | URL) => {
      expect(path).toBe("/api/v1/system/status");
      return jsonResponse({
        metrics: {
          sample: {
            sampled_at: "2026-08-29T18:00:00Z", cpu_percent: 12.5,
            memory_total_bytes: 100, memory_available_bytes: 40, memory_used_percent: 60,
            disk_total_bytes: 200, disk_available_bytes: 100, disk_used_percent: 50,
            load_1: 0.5, load_5: 0.4, load_15: 0.3, uptime_seconds: 90000,
            rx_bytes: 1000, tx_bytes: 2000, network_interface: "eth0", rate_available: true,
            sample_window_seconds: 5, rx_bytes_per_second: 100, tx_bytes_per_second: 200,
          },
          dependencies: { api: { status: "ok" }, database: { status: "ok" }, runtime: { status: "ok" } },
        },
        traffic_semantics: "server_counter_delta",
      });
    }) as typeof fetch;
    const status = await fetchSystemStatus(fetcher);
    expect(status.sample.rx_bytes_per_second).toBe(100);
    expect(status.traffic_semantics).toBe("server_counter_delta");
    expect(status.dependencies.runtime.status).toBe("ok");
  });

  it("does not fabricate a rate before the server has a delta window", () => {
    expect(formatRate(0, false)).toContain("نمونه");
    expect(formatBytes(1024 * 1024 * 2)).toBe("2.00 MB");
    expect(formatUptime(90000)).toContain("1 روز");
  });
});
