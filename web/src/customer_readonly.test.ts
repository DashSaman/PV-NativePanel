import { describe, expect, it, vi } from "vitest";
import { getCurrentSubscription, subscriptionURL } from "./customers";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("customer read-only delivery", () => {
  it("fetches the current subscription with GET and no mutation headers", async () => {
    const fetcher = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) =>
      jsonResponse({ subscription_path: "/api/v1/subscriptions/current-token", delivery_notice: "read only" }),
    );

    const result = await getCurrentSubscription("user-1", fetcher as typeof fetch);

    expect(fetcher).toHaveBeenCalledTimes(1);
    const [path, init] = fetcher.mock.calls[0];
    expect(path).toBe("/api/v1/customers/user-1/subscription");
    expect(init?.method).toBe("GET");
    expect(init?.headers).toBeUndefined();
    expect(init?.body).toBeUndefined();
    expect(subscriptionURL(result.subscription_path, "https://panel.example/panel/")).toBe(
      "https://panel.example/api/v1/subscriptions/current-token",
    );
  });
});
