import { describe, expect, it } from "vitest";
import { login, readCookie, logout } from "./auth";

describe("auth client", () => {
  it("reads the CSRF cookie without decoding another cookie", () => {
    expect(readCookie("__Host-pvnaive_csrf", "a=1; __Host-pvnaive_csrf=abc%2F123; z=2")).toBe("abc/123");
    expect(readCookie("missing", "a=1")).toBeNull();
  });

  it("submits login credentials without adding CSRF before authentication", async () => {
    const calls: Array<[RequestInfo | URL, RequestInit | undefined]> = [];
    const fetcher: typeof fetch = async (input, init) => {
      calls.push([input, init]);
      return new Response(JSON.stringify({ status: "authenticated", role: "owner" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    };
    const result = await login({ email: "owner@example.invalid", password: "secret", totpCode: "" }, fetcher);
    expect(result.status).toBe("authenticated");
    const [url, init] = calls[0]!;
    expect(url).toBe("/api/v1/auth/login");
    expect(init?.credentials).toBe("same-origin");
    expect(new Headers(init?.headers).has("X-CSRF-Token")).toBe(false);
  });

  it("binds logout to the browser CSRF token", async () => {
    const calls: Array<[RequestInfo | URL, RequestInit | undefined]> = [];
    const fetcher: typeof fetch = async (input, init) => {
      calls.push([input, init]);
      return new Response(JSON.stringify({ status: "logged_out" }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    };
    await logout("csrf-value", fetcher);
    const [, init] = calls[0]!;
    expect(new Headers(init?.headers).get("X-CSRF-Token")).toBe("csrf-value");
    expect(init?.credentials).toBe("same-origin");
  });
});
