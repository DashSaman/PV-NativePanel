import { describe, expect, it } from "vitest";
import { buildCustomerDashboard } from "./customerDashboard";
import { buildUnifiedCustomerRows } from "./customerRows";
import type { CustomerView } from "./customers";
import type { RuntimeCredential } from "./runtime";

function customer(overrides: Partial<CustomerView>): CustomerView {
  return {
    id: "user-1",
    username: "user",
    status: "active",
    service_term_id: "term-1",
    service_state: "active",
    quota_bytes: 50 * 1024 * 1024 * 1024,
    duration_seconds: 30 * 86400,
    start_policy: "on_creation",
    runtime_credential_id: "runtime-1",
    subscription_available: true,
    usage_capability: { available: false },
    ...overrides,
  };
}

const runtime: RuntimeCredential[] = [
  { id: "runtime-1", username: "active", status: "active", origin: "imported", revision: 1 },
  { id: "runtime-2", username: "needs-setup", status: "active", origin: "imported", revision: 1 },
  { id: "runtime-3", username: "suspended", status: "active", origin: "imported", revision: 1 },
];

describe("customer dashboard metrics", () => {
  it("builds honest account, quota and expiry metrics without inventing usage", () => {
    const now = new Date("2026-08-29T14:00:00Z");
    const customers = [
      customer({ id: "user-1", username: "active", runtime_credential_id: "runtime-1", expires_at: "2026-09-03T14:00:00Z" }),
      customer({ id: "user-2", username: "suspended", runtime_credential_id: "runtime-3", status: "suspended", quota_bytes: 20 * 1024 * 1024 * 1024, expires_at: "2026-09-18T14:00:00Z" }),
      customer({ id: "user-3", username: "unlimited", runtime_credential_id: "runtime-4", quota_bytes: null, expires_at: undefined }),
    ];
    const rows = buildUnifiedCustomerRows(customers, runtime);
    const metrics = buildCustomerDashboard(rows, now);

    expect(metrics.totalAccounts).toBe(4);
    expect(metrics.managedAccounts).toBe(3);
    expect(metrics.needsSetup).toBe(1);
    expect(metrics.suspended).toBe(1);
    expect(metrics.configuredQuotaGB).toBe(70);
    expect(metrics.unlimitedAccounts).toBe(1);
    expect(metrics.expiry.within7Days).toBe(1);
    expect(metrics.expiry.within30Days).toBe(1);
    expect(metrics.expiry.noExpiry).toBe(2);
    expect(metrics.usageProven).toBe(false);
  });
});
