import { describe, expect, it } from "vitest";
import {
  buildRevisionPayload,
  driftLabel,
  formatLastSeen,
  healthLabel,
  maintenanceActions,
  maintenanceLabel,
  manifestSummary,
  normalizeEnrollToken,
  normalizeManifest,
  normalizePoolNodes,
  REVISION_LIMITS,
  validateEnrollInput,
  validateRevisionInput,
  type PoolNodeRow,
} from "./poolManager";

const validRow: PoolNodeRow = {
  id: "0d3c6a1e-1111-4222-8333-444455556666",
  display_name: "frankfurt-edge",
  region: "eu-central",
  capacity_weight: 4,
  health: "healthy",
  maintenance: "active",
  desired_revision: 7,
  applied_revision: 7,
  drift: "in_sync",
  last_seen_at: "2026-09-14T18:00:00Z",
};

describe("normalizePoolNodes", () => {
  it("keeps well-formed rows and the declared count", () => {
    const result = normalizePoolNodes({ nodes: [validRow], count: 1 });
    expect(result.count).toBe(1);
    expect(result.nodes).toHaveLength(1);
    expect(result.nodes[0].display_name).toBe("frankfurt-edge");
    expect(result.nodes[0].health).toBe("healthy");
  });

  it("drops rows without an id instead of fabricating identity", () => {
    const result = normalizePoolNodes({ nodes: [{ ...validRow, id: "" }, { display_name: "x" }], count: 2 });
    expect(result.nodes).toHaveLength(0);
  });

  it("collapses unknown states to safe defaults (fail-closed)", () => {
    const result = normalizePoolNodes({
      nodes: [{ ...validRow, health: "blazing", maintenance: "warp", drift: "sideways" }],
      count: 1,
    });
    expect(result.nodes[0].health).toBe("unknown");
    expect(result.nodes[0].maintenance).toBe("active");
    expect(result.nodes[0].drift).toBe("in_sync");
  });

  it("derives drift from revisions when the flag is missing", () => {
    const result = normalizePoolNodes({ nodes: [{ ...validRow, drift: undefined, desired_revision: 9, applied_revision: 7 }] });
    expect(result.nodes[0].drift).toBe("pending");
  });

  it("accepts string-encoded revisions (json numeric edge)", () => {
    const result = normalizePoolNodes({ nodes: [{ ...validRow, desired_revision: "12", applied_revision: "12" }] });
    expect(result.nodes[0].desired_revision).toBe(12);
  });

  it("returns an empty list for malformed payloads", () => {
    expect(normalizePoolNodes(null).nodes).toHaveLength(0);
    expect(normalizePoolNodes({ nodes: "nope" }).nodes).toHaveLength(0);
  });

  it("keeps null last_seen_at as null (never-seen is truthful)", () => {
    const result = normalizePoolNodes({ nodes: [{ ...validRow, last_seen_at: undefined }] });
    expect(result.nodes[0].last_seen_at).toBeNull();
  });
});

describe("labels", () => {
  it("labels every drift, health and maintenance state in Persian", () => {
    expect(driftLabel("in_sync")).toBe("هم‌گام");
    expect(driftLabel("pending")).toBe("در انتظار اعمال");
    expect(driftLabel("ahead")).toBe("جلوتر از وضعیت مطلوب");
    expect(healthLabel("healthy")).toBe("سالم");
    expect(healthLabel("offline")).toBe("خارج از دسترس");
    expect(maintenanceLabel("draining")).toBe("در حال تخلیه");
  });
});

describe("maintenanceActions — mirrors the 0033 state machine", () => {
  it("active may only start draining (direct disable is refused by db)", () => {
    expect(maintenanceActions("active")).toEqual(["draining"]);
  });
  it("draining may finish the exit or return to service", () => {
    expect(maintenanceActions("draining")).toEqual(["disabled", "active"]);
  });
  it("disabled may only be re-activated", () => {
    expect(maintenanceActions("disabled")).toEqual(["active"]);
  });
});

describe("formatLastSeen", () => {
  const now = new Date("2026-09-14T20:00:00Z");
  it("renders a never-seen node without inventing a time", () => {
    expect(formatLastSeen(null, now)).toBe("هرگز دیده نشده");
  });
  it("buckets sub-minute ages", () => {
    expect(formatLastSeen("2026-09-14T19:59:30Z", now)).toBe("چند لحظه پیش");
  });
  it("renders minutes and hours in Persian digits", () => {
    expect(formatLastSeen("2026-09-14T19:30:00Z", now)).toBe("۳۰ دقیقه پیش");
    expect(formatLastSeen("2026-09-14T16:00:00Z", now)).toBe("۴ ساعت پیش");
  });
  it("renders days beyond 48h", () => {
    expect(formatLastSeen("2026-09-11T18:00:00Z", now)).toBe("۳ روز پیش");
  });
  it("admits malformed input as unknown instead of crashing", () => {
    expect(formatLastSeen("not-a-date", now)).toBe("نامشخص");
  });
});

describe("normalizeEnrollToken", () => {
  it("accepts a real one-time token response", () => {
    const result = normalizeEnrollToken({ token: "a".repeat(64), expires_at: "2026-09-14T20:00:00Z" });
    expect(result?.token).toBe("a".repeat(64));
  });
  it("rejects placeholder tokens so the UI never fakes success", () => {
    expect(normalizeEnrollToken({ token: "short", expires_at: "2026-09-14T20:00:00Z" })).toBeNull();
    expect(normalizeEnrollToken({ token: "a".repeat(64) })).toBeNull();
    expect(normalizeEnrollToken(null)).toBeNull();
  });
});

describe("validateEnrollInput — mirrors fleet.Validate*", () => {
  const base = { token: "a".repeat(64), displayName: "vienna-2", region: "", capacityWeight: 1 };
  it("accepts a valid enrollment", () => {
    expect(validateEnrollInput(base)).toBeNull();
  });
  it("enforces token length, name bounds and weight bounds", () => {
    expect(validateEnrollInput({ ...base, token: "short" })).toContain("توکن");
    expect(validateEnrollInput({ ...base, displayName: "" })).toContain("نام نمایشی");
    expect(validateEnrollInput({ ...base, displayName: "x".repeat(121) })).toContain("نام نمایشی");
    expect(validateEnrollInput({ ...base, region: "x".repeat(61) })).toContain("ناحیه");
    expect(validateEnrollInput({ ...base, capacityWeight: 0 })).toContain("وزن");
    expect(validateEnrollInput({ ...base, capacityWeight: 10001 })).toContain("وزن");
  });
});

describe("validateRevisionInput — mirrors httpapi.buildManifest", () => {
  const base = { poolId: "eu-primary", endpoints: [{ host: "edge.example.net", port: 443, sni: "edge.example.net" }], weight: 2, validityMinutes: 2880 };
  it("accepts a valid revision", () => {
    expect(validateRevisionInput(base)).toBeNull();
  });
  it("enforces pool id, endpoint count, host/port/sni rules", () => {
    expect(validateRevisionInput({ ...base, poolId: "" })).toContain("شناسه استخر");
    expect(validateRevisionInput({ ...base, poolId: "x".repeat(REVISION_LIMITS.poolIdMax + 1) })).toContain("شناسه استخر");
    expect(validateRevisionInput({ ...base, endpoints: [] })).toContain("نقطه اتصال");
    expect(validateRevisionInput({ ...base, endpoints: Array.from({ length: REVISION_LIMITS.endpointsMax + 1 }, () => base.endpoints[0]) })).toContain("نقطه اتصال");
    expect(validateRevisionInput({ ...base, endpoints: [{ host: "", port: 443, sni: "" }] })).toContain("میزبان");
    expect(validateRevisionInput({ ...base, endpoints: [{ host: "h", port: 70000, sni: "" }] })).toContain("پورت");
    expect(validateRevisionInput({ ...base, endpoints: [{ host: "h", port: 443, sni: "s".repeat(256) }] })).toContain("SNI");
  });
  it("enforces weight and validity windows", () => {
    expect(validateRevisionInput({ ...base, weight: 0 })).toContain("وزن");
    expect(validateRevisionInput({ ...base, weight: 1001 })).toContain("وزن");
    expect(validateRevisionInput({ ...base, validityMinutes: 59 })).toContain("اعتبار");
    expect(validateRevisionInput({ ...base, validityMinutes: 20161 })).toContain("اعتبار");
  });
});

describe("buildRevisionPayload", () => {
  it("serializes the exact wire contract with trimmed values", () => {
    const payload = buildRevisionPayload({
      poolId: " eu-primary ",
      endpoints: [{ host: " edge.example.net ", port: 443, sni: " edge.example.net " }],
      weight: 2,
      validityMinutes: 2880,
    });
    expect(payload).toEqual({
      pool_id: "eu-primary",
      endpoints: [{ host: "edge.example.net", port: 443, sni: "edge.example.net" }],
      weight: 2,
      validity_minutes: 2880,
    });
  });
});

describe("normalizeManifest + manifestSummary", () => {
  const manifest = {
    schema: "pvnaive.pool.manifest/v1",
    pool_id: "eu-primary",
    weight: 2,
    endpoints: [{ host: "edge.example.net", port: 443, sni: "edge.example.net" }],
    valid_until: "2026-09-16T12:00:00Z",
  };
  it("accepts a signed envelope", () => {
    const result = normalizeManifest({ revision: 7, manifest, signature: "base64sig" });
    expect(result?.revision).toBe(7);
    expect(result?.signature).toBe("base64sig");
  });
  it("rejects envelopes without manifest or signature", () => {
    expect(normalizeManifest({ revision: 7, signature: "sig" })).toBeNull();
    expect(normalizeManifest({ revision: 7, manifest })).toBeNull();
  });
  it("summarizes the operator-relevant manifest fields", () => {
    const lines = manifestSummary(manifest);
    expect(lines.some((line) => line.includes("eu-primary"))).toBe(true);
    expect(lines.some((line) => line.includes("edge.example.net:443"))).toBe(true);
    expect(lines.some((line) => line.includes("SNI"))).toBe(true);
    expect(lines.some((line) => line.includes("اعتبار"))).toBe(true);
  });
});
