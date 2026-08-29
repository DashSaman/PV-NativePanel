import { afterEach, describe, expect, it, vi } from "vitest";
import { addCustomerVolume, extendCustomerTime } from "./customerAdjustments";

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

afterEach(() => Reflect.deleteProperty(globalThis, "document"));

describe("customer service adjustments", () => {
  it("adds volume as a delta instead of replacing the total", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ service_term: { id: "term-1" }, runtime_mutated: false }),
    );

    await addCustomerVolume("user-1", 20, fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-1/volume/add");
    expect(init?.method).toBe("POST");
    expect(JSON.parse(String(init?.body))).toEqual({ delta_gb: 20 });
  });

  it("extends validity by whole days without runtime fields", async () => {
    installCSRF();
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ service_term: { id: "term-1" }, runtime_mutated: false }),
    );

    await extendCustomerTime("user-1", 30, fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-1/validity/extend");
    expect(init?.method).toBe("POST");
    expect(JSON.parse(String(init?.body))).toEqual({ days: 30 });
  });
});
