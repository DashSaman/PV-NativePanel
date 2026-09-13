import { afterEach, describe, expect, it, vi } from "vitest";
import {
  executeProductBulk,
  getProductSubscription,
  listProductCustomers,
  listProductPlans,
  killProductCustomerSession,
  killProductCustomerSessionAndReload,
  previewProductBulk,
  reissueProductSubscription,
  renewProductCustomer,
  resetProductUsage,
  revokeProductCustomer,
  rotateProductPassword,
  suspendProductCustomer,
  updateProductCustomer,
} from "./productApi";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function installCSRF() {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "__Host-pvnaive_csrf=csrf-product-test" },
  });
}

function callsOf(fetcher: ReturnType<typeof vi.fn>): Array<[RequestInfo | URL, RequestInit?]> {
  return fetcher.mock.calls as unknown as Array<[RequestInfo | URL, RequestInit?]>;
}

afterEach(() => {
  Reflect.deleteProperty(globalThis, "document");
});

describe("WS2 product API client", () => {
  it("sends customer directory filters to the server-side /users endpoint", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ customers: [], page: 2, page_size: 25, total: 0 }));

    await listProductCustomers({
      q: "ali",
      status: "on_hold",
      plan: "plan-1",
      group: "group-1",
      tag: "tag-1",
      reseller: "reseller-1",
      expiryFrom: "2026-08-01T00:00:00.000Z",
      expiryTo: "2026-09-01T00:00:00.000Z",
      unlimitedVolume: true,
      unlimitedExpiry: false,
      page: 2,
      pageSize: 25,
      sort: "expiry",
      direction: "asc",
    }, fetcher as typeof fetch);

    const [raw] = callsOf(fetcher)[0];
    const url = new URL(String(raw), "https://panel.example");
    expect(url.pathname).toBe("/api/v1/users");
    expect(url.searchParams.get("q")).toBe("ali");
    expect(url.searchParams.get("status")).toBe("on_hold");
    expect(url.searchParams.get("plan")).toBe("plan-1");
    expect(url.searchParams.get("group")).toBe("group-1");
    expect(url.searchParams.get("tag")).toBe("tag-1");
    expect(url.searchParams.get("reseller")).toBe("reseller-1");
    expect(url.searchParams.get("unlimited_volume")).toBe("true");
    expect(url.searchParams.get("unlimited_expiry")).toBe("false");
    expect(url.searchParams.get("page")).toBe("2");
    expect(url.searchParams.get("page_size")).toBe("25");
    expect(url.searchParams.get("sort")).toBe("expiry");
    expect(url.searchParams.get("dir")).toBe("asc");
  });

  it("resets one customer usage only with explicit confirmation and idempotency", async () => {
    installCSRF();
    const fetcher = vi.fn(async () => jsonResponse({ reset_event: { id: "r1", previous_used_bytes: 300 }, idempotent_replay: false }));
    await resetProductUsage("user 1", fetcher as typeof fetch);
    const [path, init] = callsOf(fetcher)[0];
    expect(path).toBe("/api/v1/users/user%201/reset-usage");
    expect(init?.method).toBe("POST");
    expect(JSON.parse(String(init?.body))).toEqual({ confirm: true });
    const headers = init?.headers as Record<string, string>;
    expect(headers["X-CSRF-Token"]).toBe("csrf-product-test");
    expect(headers["Idempotency-Key"]).toMatch(/^product-reset-usage-/);
  });

  it("loads plans from the ready product catalog endpoint", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ plans: [{ id: "p1", name: "50G", enabled: true }] }));
    const result = await listProductPlans(fetcher as typeof fetch);
    expect(callsOf(fetcher)[0][0]).toBe("/api/v1/plans");
    expect(result[0].name).toBe("50G");
  });

  it("updates note/group/on-hold/tags/next-plan through PATCH /users/:id", async () => {
    installCSRF();
    const fetcher = vi.fn(async () => jsonResponse({ metadata: { on_hold: true } }));

    await updateProductCustomer("user 1", {
      note: "VIP",
      group_id: "g1",
      on_hold: true,
      add_tag_ids: ["t1"],
      remove_tag_ids: ["t2"],
      next_plan_id: "p2",
    }, fetcher as typeof fetch);

    const [path, init] = callsOf(fetcher)[0];
    expect(path).toBe("/api/v1/users/user%201");
    expect(init?.method).toBe("PATCH");
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-product-test");
  });

  it("uses reseller-ready renewal, subscription, lifecycle and password endpoints", async () => {
    installCSRF();
    const fetcher = vi.fn(async (input: RequestInfo | URL) => {
      const path = String(input);
      if (path.endsWith("/subscription")) return jsonResponse({ subscription_path: "/sub/token", direct_uri: "naive+https://example" });
      if (path.endsWith("/rotate-password")) return jsonResponse({ generated_password: "once-only" });
      return jsonResponse({ status: "ok" });
    });

    await getProductSubscription("user 1", fetcher as typeof fetch);
    await reissueProductSubscription("user 1", fetcher as typeof fetch);
    await renewProductCustomer("user 1", { mode: "renew_current" }, fetcher as typeof fetch);
    await rotateProductPassword("user 1", { password: "", generate_password: true }, fetcher as typeof fetch);
    await suspendProductCustomer("user 1", fetcher as typeof fetch);
    await revokeProductCustomer("user 1", fetcher as typeof fetch);

    const calls = callsOf(fetcher);
    expect(calls[0][0]).toBe("/api/v1/users/user%201/subscription");
    expect(calls[0][1]?.method).toBe("GET");
    expect(calls[1][0]).toBe("/api/v1/users/user%201/subscription/rotate");
    expect(calls[1][1]?.method).toBe("POST");
    expect(calls[2][0]).toBe("/api/v1/users/user%201/renew");
    expect(calls[3][0]).toBe("/api/v1/users/user%201/rotate-password");
    expect(calls[4][0]).toBe("/api/v1/users/user%201/suspend");
    expect(calls[5][0]).toBe("/api/v1/users/user%201/revoke");
    expect((calls[1][1]?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-product-test");
  });

  it("kills exactly one selected active session with CSRF and no tuple fields", async () => {
    installCSRF();
    const fetcher = vi.fn(async () => jsonResponse({ status: "completed", found: true, killed: true, session_id: "session 1", credential_mutated: false }));

    const result = await killProductCustomerSession("user 1", "session 1", fetcher as typeof fetch);

    const [path, init] = callsOf(fetcher)[0];
    expect(path).toBe("/api/v1/users/user%201/sessions/session%201");
    expect(init?.method).toBe("DELETE");
    expect(init?.body).toBeUndefined();
    expect((init?.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-product-test");
    expect(result).toEqual({ status: "completed", found: true, killed: true, session_id: "session 1", credential_mutated: false });
  });

  it("reloads active sessions only after the selected session kill completes", async () => {
    installCSRF();
    const fetcher = vi
      .fn()
      .mockImplementationOnce(async () => jsonResponse({ status: "completed", found: true, killed: true, session_id: "s1", credential_mutated: false }))
      .mockImplementationOnce(async () => jsonResponse({ sessions: [{ session_id: "s2" }], observed_at: "2026-09-01T19:00:00Z" }));

    const result = await killProductCustomerSessionAndReload("u1", "s1", fetcher as typeof fetch);

    const calls = callsOf(fetcher);
    expect(calls.map(([path]) => path)).toEqual(["/api/v1/users/u1/sessions/s1", "/api/v1/users/u1/sessions"]);
    expect(result.sessions).toEqual([{ session_id: "s2" }]);
  });

  it("reuses exactly the preview idempotency key for bulk execute", async () => {
    installCSRF();
    const request = { action: "add_volume" as const, customer_ids: ["u1", "u2"], volume_gb: 10 };
    const fetcher = vi
      .fn()
      .mockImplementationOnce(async () => jsonResponse({ bulk: { id: "b1", status: "previewed", preview: { requested: 2, affected: 2, changes: ["add_volume"], conflicts: [], skipped: [], invalid: [] }, request } }))
      .mockImplementationOnce(async () => jsonResponse({ bulk: { id: "b1", status: "executed", preview: { requested: 2, affected: 2, changes: ["add_volume"], conflicts: [], skipped: [], invalid: [] }, result: { succeeded: 2, failed: 0, skipped: 0, items: [] }, request } }));

    const preview = await previewProductBulk(request, fetcher as typeof fetch);
    await executeProductBulk(request, preview.idempotencyKey, fetcher as typeof fetch);

    const calls = callsOf(fetcher);
    expect(calls[0][0]).toBe("/api/v1/users/bulk/preview");
    expect(calls[1][0]).toBe("/api/v1/users/bulk/execute");
    expect((calls[0][1]?.headers as Record<string, string>)["Idempotency-Key"])
      .toBe((calls[1][1]?.headers as Record<string, string>)["Idempotency-Key"]);
  });

  it("supports reset_usage as a previewed bulk action with the same execute key", async () => {
    installCSRF();
    const request = { action: "reset_usage" as const, customer_ids: ["u1", "u2"] };
    const fetcher = vi
      .fn()
      .mockImplementationOnce(async () => jsonResponse({ bulk: { id: "b-reset", status: "previewed", preview: { requested: 2, affected: 2, changes: ["reset_usage"], conflicts: [], skipped: [], invalid: [] }, request } }))
      .mockImplementationOnce(async () => jsonResponse({ bulk: { id: "b-reset", status: "executed", preview: { requested: 2, affected: 2, changes: ["reset_usage"], conflicts: [], skipped: [], invalid: [] }, result: { succeeded: 2, failed: 0, skipped: 0, items: [] }, request } }));

    const preview = await previewProductBulk(request, fetcher as typeof fetch);
    await executeProductBulk(request, preview.idempotencyKey, fetcher as typeof fetch);
    const calls = callsOf(fetcher);
    expect((calls[0][1]?.headers as Record<string, string>)["Idempotency-Key"])
      .toBe((calls[1][1]?.headers as Record<string, string>)["Idempotency-Key"]);
  });

});
