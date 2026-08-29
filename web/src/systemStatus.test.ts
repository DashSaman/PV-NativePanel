import { describe, expect, it } from "vitest";
import { dependencyEntries, formatBytes, formatRate, formatUptime, normalizeSystemStatus } from "./systemStatus";

const response = {
  metrics: {
    sample: {
      sampled_at: "2026-08-30T00:00:00Z",
      cpu_percent: 31.5,
      memory_total_bytes: 8_589_934_592,
      memory_available_bytes: 3_221_225_472,
      memory_used_percent: 62.5,
      disk_total_bytes: 53_687_091_200,
      disk_available_bytes: 10_737_418_240,
      disk_used_percent: 80,
      load_1: 0.4,
      load_5: 0.3,
      load_15: 0.2,
      uptime_seconds: 172_800,
      rx_bytes: 1000,
      tx_bytes: 2000,
      network_interface: "eth0",
      rate_available: true,
      sample_window_seconds: 5,
      rx_bytes_per_second: 4096,
      tx_bytes_per_second: 2048,
    },
    dependencies: {
      api: { status: "ok" },
      database: { status: "ok" },
      runtime: { status: "ok" },
      telemetry: { status: "ok" },
    },
  },
  traffic_semantics: "server_counter_delta",
};

describe("system status", () => {
  it("normalizes the real backend envelope and preserves telemetry dependency", () => {
    const status = normalizeSystemStatus(response);
    expect(status.sample.cpu_percent).toBe(31.5);
    expect(status.dependencies.telemetry?.status).toBe("ok");
    expect(status.traffic_semantics).toBe("server_counter_delta");
  });

  it("includes Telemetry in dependency presentation", () => {
    const status = normalizeSystemStatus(response);
    expect(dependencyEntries(status.dependencies).map((item) => item.label)).toEqual(["API", "DB", "Runtime", "Telemetry"]);
  });

  it("never presents a numeric network rate when the server rate is unavailable", () => {
    expect(formatRate(4096, false)).toBe("—");
    expect(formatRate(4096, true)).not.toBe("—");
  });

  it("formats byte and uptime values without fabricating data", () => {
    expect(formatBytes(1_073_741_824)).toContain("GB");
    expect(formatUptime(172_800)).toContain("2");
  });
});
