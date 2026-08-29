import { afterEach, describe, expect, it, vi } from "vitest";
import {
  adoptRuntimeCustomer,
  createCustomer,
  subscriptionURL,
  updateCustomerService,
  type CreateCustomerRequest,
  type CustomerServiceSettingsRequest,
} from "./customers";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

afterEach(() => {
  Reflect.deleteProperty(globalThis, "document");
});

function installCSRF() {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "__Host-pvnaive_csrf=csrf-test-value" },
  });
}

describe("customer API", () => {
  it("posts the owner customer form with CSRF and idempotency protection", async () => {
    installCSRF();
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

  it("adopts an existing Runtime credential without sending a password", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({
        user: { id: "user-old", username: "amirreza", status: "active", revision: 1 },
        service_term: { id: "term-old", state: "active", duration_seconds: 2592000, start_policy: "on_creation" },
        runtime_credential: { id: "runtime-old", username: "amirreza", status: "active" },
        subscription_path: "/api/v1/subscriptions/adopted-token",
        usage_capability: { available: false, reason: "exact_accounting_not_proven" },
      }, 201),
    );
    const settings: CustomerServiceSettingsRequest = {
      quota_gb: 80,
      validity: { mode: "on_creation", duration_days: 30 },
    };

    const result = await adoptRuntimeCustomer("runtime-old", settings, fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/adopt-runtime");
    expect(init?.method).toBe("POST");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-test-value");
    expect(JSON.parse(String(init?.body))).toEqual({ runtime_credential_id: "runtime-old", ...settings });
    expect(String(init?.body)).not.toContain("password");
    expect(result.runtime_credential.id).toBe("runtime-old");
  });

  it("updates quota and validity on a managed customer without runtime fields", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ service_term: { id: "term-old", quota_bytes: 128849018880, state: "active", revision: 2 }, runtime_mutated: false }),
    );
    const settings: CustomerServiceSettingsRequest = {
      quota_gb: 120,
      validity: { mode: "fixed_expiry", expires_at: "2026-09-30T20:30:00.000Z" },
    };

    const result = await updateCustomerService("user-old", settings, fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-old/service");
    expect(init?.method).toBe("PATCH");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-test-value");
    expect(JSON.parse(String(init?.body))).toEqual(settings);
    expect(String(init?.body)).not.toContain("runtime_credential");
    expect(result.runtime_mutated).toBe(false);
  });

  it("builds a same-origin subscription URL without trusting API host input", () => {
    expect(subscriptionURL("/api/v1/subscriptions/opaque-token", "https://panel.example/panel/")).toBe(
      "https://panel.example/api/v1/subscriptions/opaque-token",
    );
  });
});
