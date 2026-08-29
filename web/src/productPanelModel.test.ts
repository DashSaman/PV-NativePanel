import { describe, expect, it } from "vitest";
import { canManagePlans, canUseCustomerProduct, canUseRawRuntime, usagePresentation } from "./productPanelModel";
import type { ProductCustomer } from "./productApi";

describe("customer product panel policy", () => {
  it("exposes customer product operations to owner/admin/reseller but raw runtime only to owner", () => {
    expect(canUseCustomerProduct("owner")).toBe(true);
    expect(canUseCustomerProduct("admin")).toBe(true);
    expect(canUseCustomerProduct("reseller")).toBe(true);
    expect(canUseCustomerProduct("operator")).toBe(false);
    expect(canUseCustomerProduct("auditor")).toBe(false);
    expect(canManagePlans("owner")).toBe(true);
    expect(canManagePlans("admin")).toBe(true);
    expect(canManagePlans("reseller")).toBe(false);
    expect(canUseRawRuntime("owner")).toBe(true);
    expect(canUseRawRuntime("admin")).toBe(false);
  });

  it("never presents incomplete accounting as exact consumption", () => {
    const customer = {
      quota_bytes: 1000,
      usage_capability: { available: false, reason: "accounting_incomplete" },
      usage: {
        available: false,
        accounting_complete: false,
        upload_bytes: 100,
        download_bytes: 200,
        used_bytes: 300,
        remaining_bytes: 700,
        online: true,
        session_count: 1,
      },
    } as ProductCustomer;

    expect(usagePresentation(customer)).toEqual({
      exact: false,
      used: "نامشخص",
      remaining: "نامشخص",
      presence: "نامشخص",
      sessions: "—",
      lastOnline: "—",
    });
  });
});
