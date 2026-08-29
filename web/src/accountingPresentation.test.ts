import { describe, expect, it } from "vitest";
import { customerUsagePresentation, type CustomerView } from "./customers";

function customer(overrides: Partial<CustomerView>): CustomerView {
  return {
    id: "user-1",
    username: "demo",
    status: "active",
    service_term_id: "term-1",
    service_state: "active",
    quota_bytes: 50 * 1024 ** 3,
    duration_seconds: 30 * 86400,
    start_policy: "on_creation",
    runtime_credential_id: "runtime-1",
    subscription_available: true,
    usage_capability: { available: true },
    accounting_complete: true,
    upload_bytes: 5 * 1024 ** 3,
    download_bytes: 10 * 1024 ** 3,
    used_bytes: 15 * 1024 ** 3,
    remaining_bytes: 35 * 1024 ** 3,
    online: true,
    online_sessions: 2,
    ...overrides,
  };
}

describe("customer exact accounting presentation", () => {
  it("shows exact used, remaining and online sessions when accounting is complete", () => {
    expect(customerUsagePresentation(customer({}))).toEqual({
      primary: "15 GB مصرف",
      secondary: "35 GB مانده · آنلاین (2)",
      exact: true,
    });
  });

  it("never presents remaining bytes as exact when accounting is incomplete", () => {
    expect(customerUsagePresentation(customer({
      usage_capability: { available: false, reason: "accounting_incomplete" },
      accounting_complete: false,
      used_bytes: 20 * 1024 ** 2,
      remaining_bytes: 49 * 1024 ** 3,
      online: false,
      online_sessions: 0,
    }))).toEqual({
      primary: "حداقل 20 MB ثبت‌شده",
      secondary: "Accounting ناقص · باقی‌مانده نامشخص",
      exact: false,
    });
  });
});
