import { describe, expect, it, vi } from "vitest";
import { createCustomer, subscriptionURL, type CreateCustomerRequest } from "./customers";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("customer API", () => {
  it("posts the owner customer form with CSRF and idempotency protection", async () => {
    Object.defineProperty(document, "cookie", {
      configurable: true,
      value: "__Host-pvnaive_csrf=csrf-test-value",
    });
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({
        user: { id: "user-1", username: "customer1", status: "active", revision: 1 },
        service_term: { id: "term-1", state: "pending", duration_seconds: 2592000, start_policy: "on_first_successful_connection" },
        runtime_credential: { id: "runtime-1", username: "customer1", status: "active" },
        generated_password: "one-time-secret",
        subscription_path: "/api/v1/subscriptions/opaque-token",
        usage_capability: { available: false, reason: "exact_accounting_not_proven" },
      }, 201),
    );
    const request: CreateCustomerRequest = {
      username: "customer1",
      generate_password: true,
      password: "",
      quota_gb: 50,
      validity: { mode: "on_first_successful_connection", duration_days: 30 },
    };

    const result = await createCustomer(request, fetcher as typeof fetch);

    expect(fetcher).toHaveBeenCalledTimes(1);
    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers");
    expect(init?.method).toBe("POST");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-test-value");
    expect((init?.headers as Record<string, string>)["Idempotency-Key"]).toMatch(/^customer-/);
    expect(JSON.parse(String(init?.body))).toEqual(request);
    expect(result.generated_password).toBe("one-time-secret");
    expect(result.subscription_path).toBe("/api/v1/subscriptions/opaque-token");
  });

  it("builds a same-origin subscription URL without trusting API host input", () => {
    expect(subscriptionURL("/api/v1/subscriptions/opaque-token", "https://panel.example/panel/")).toBe(
      "https://panel.example/api/v1/subscriptions/opaque-token",
    );
  });
});
