import { describe, expect, it } from "vitest";
import { buildUnifiedCustomerRows } from "./customerRows";
import type { CustomerView } from "./customers";
import type { RuntimeCredential } from "./runtime";

const customer: CustomerView = {
  id: "user-1",
  username: "managed",
  status: "active",
  service_term_id: "term-1",
  service_state: "active",
  quota_bytes: 50 * 1024 * 1024 * 1024,
  duration_seconds: 30 * 86400,
  start_policy: "on_creation",
  runtime_credential_id: "runtime-1",
  subscription_available: true,
  usage_capability: { available: false },
};

const runtime: RuntimeCredential[] = [
  { id: "runtime-1", username: "managed", status: "active", origin: "imported", revision: 1 },
  { id: "runtime-2", username: "old-account", status: "active", origin: "imported", revision: 1 },
];

describe("unified customer rows", () => {
  it("shows managed and pre-panel runtime accounts in the same row collection without duplicates", () => {
    const rows = buildUnifiedCustomerRows([customer], runtime);

    expect(rows).toHaveLength(2);
    expect(rows.map((row) => row.username).sort()).toEqual(["managed", "old-account"]);
    expect(rows.find((row) => row.username === "managed")?.kind).toBe("customer");
    expect(rows.find((row) => row.username === "old-account")?.kind).toBe("runtime");
  });

  it("keeps revoked runtime-only credentials out of the customer management list", () => {
    const rows = buildUnifiedCustomerRows([], [
      { id: "runtime-revoked", username: "gone", status: "revoked", origin: "imported", revision: 1 },
    ]);
    expect(rows).toEqual([]);
  });
});
