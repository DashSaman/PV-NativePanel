/// <reference types="vite/client" />
import { describe, expect, it } from "vitest";
import source from "./ProductCustomers.tsx?raw";

describe("Task13 exact-session kill UI wiring", () => {
  it("offers a per-session kill action that refreshes the trusted session list", () => {
    expect(source).toContain("killProductCustomerSessionAndReload");
    expect(source).toContain("session.session_id");
    expect(source).toContain("قطع نشست");
    expect(source).toContain("رمز و لینک اشتراک تغییر نمی‌کنند");
    expect(source).not.toContain("runtime_credential_id, node_id, boot_id");
  });
});
