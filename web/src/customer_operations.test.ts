import { afterEach, describe, expect, it, vi } from "vitest";
import {
  deleteCustomer,
  resumeCustomer,
  rotateCustomerPassword,
  suspendCustomer,
} from "./customers";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function installCSRF() {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "__Host-pvnaive_csrf=csrf-test-value" },
  });
}

afterEach(() => {
  Reflect.deleteProperty(globalThis, "document");
});

describe("customer lifecycle API", () => {
  it.each([
    ["suspend", suspendCustomer, "/api/v1/customers/user-1/suspend"],
    ["resume", resumeCustomer, "/api/v1/customers/user-1/resume"],
  ] as const)("%s is an explicit protected mutation", async (_name, action, path) => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ status: "ok" }),
    );

    await action("user-1", fetcher as typeof fetch);

    const [actualPath, init] = fetcher.mock.calls[0];
    expect(actualPath).toBe(path);
    expect(init?.method).toBe("POST");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-test-value");
    expect((init?.headers as Record<string, string>)["Idempotency-Key"]).toMatch(/^customer-/);
  });

  it("delete uses safe DELETE revoke API", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ status: "revoked" }),
    );

    await deleteCustomer("user-1", fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-1");
    expect(init?.method).toBe("DELETE");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-test-value");
  });

  it("password rotation is independent and can request a generated password", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({
        runtime_credential: { id: "runtime-1", username: "alice", status: "active" },
        generated_password: "generated-once",
        delivery_notice: "Subscription unchanged",
      }),
    );

    const result = await rotateCustomerPassword(
      "user-1",
      { password: "", generate_password: true },
      fetcher as typeof fetch,
    );

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-1/rotate-password");
    expect(init?.method).toBe("POST");
    expect(JSON.parse(String(init?.body))).toEqual({ password: "", generate_password: true });
    expect(result.generated_password).toBe("generated-once");
  });
});
