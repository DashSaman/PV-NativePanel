import { describe, expect, it } from "vitest";
import { concurrencyLabel, deriveEffectiveStatus, quotaState } from "./userStatus";

describe("user status model", () => {
  it("makes depleted and expired explicit", () => {
    expect(deriveEffectiveStatus({ configured: "active", expiresAt: null, now: 10, used: 100, limit: 100 })).toBe("depleted");
    expect(deriveEffectiveStatus({ configured: "active", expiresAt: 9, now: 10, used: 0, limit: 100 })).toBe("expired");
  });
  it("does not confuse manual suspension with quota", () => {
    expect(deriveEffectiveStatus({ configured: "suspended", expiresAt: 1, now: 10, used: 100, limit: 100 })).toBe("suspended");
  });
  it("uses warning thresholds consistently", () => {
    expect(quotaState(79, 100)).toBe("healthy");
    expect(quotaState(80, 100)).toBe("warning");
    expect(quotaState(95, 100)).toBe("critical");
  });
  it("labels single and multi-user plans", () => {
    expect(concurrencyLabel(1)).toBe("تک‌کاربره");
    expect(concurrencyLabel(4)).toContain("4");
    expect(concurrencyLabel(null)).toContain("نامحدود");
  });
});
