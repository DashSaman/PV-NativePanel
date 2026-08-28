import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  createRuntimeCredential,
  importCurrentRuntime,
  listRuntimeCredentials,
  revokeRuntimeCredential,
  rotateRuntimeCredential,
  updateRuntimeCredential,
  type RuntimeCredential,
} from "./runtime";

const credential: RuntimeCredential = {
  id: "cred/one",
  username: "customer.one",
  status: "active",
  origin: "panel",
  revision: 7,
};

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

beforeEach(() => {
  Object.defineProperty(globalThis, "document", {
    configurable: true,
    value: { cookie: "__Host-pvnaive_csrf=csrf-runtime-test" },
  });
});

describe("runtime API client", () => {
  it("lists credentials with same-origin credentials and no secret fields", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ credentials: [credential] }));
    const result = await listRuntimeCredentials(fetcher as typeof fetch);

    expect(result).toEqual([credential]);
    expect(fetcher).toHaveBeenCalledWith(
      "/api/v1/runtime/naive/credentials",
      expect.objectContaining({ method: "GET", credentials: "same-origin" }),
    );
    expect(JSON.stringify(result)).not.toMatch(/password|secret_ciphertext|secret_nonce|secret_hash/i);
  });

  it("imports current runtime with CSRF and an idempotency key", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ credentials: [credential] }));
    await importCurrentRuntime(fetcher as typeof fetch);

    const [path, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(path).toBe("/api/v1/runtime/naive/import");
    expect(init.credentials).toBe("same-origin");
    expect(init.headers).toMatchObject({
      "Content-Type": "application/json",
      "X-CSRF-Token": "csrf-runtime-test",
    });
    expect((init.headers as Record<string, string>)["Idempotency-Key"]).toMatch(/^runtime-/);
  });

  it("creates a generated-password credential through the guarded mutation contract", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ credential, generated_password: "one-time-secret" }, 201));
    const result = await createRuntimeCredential("customer.one", "", true, fetcher as typeof fetch);

    const [, init] = fetcher.mock.calls[0] as unknown as [string, RequestInit];
    expect(JSON.parse(String(init.body))).toEqual({
      username: "customer.one",
      password: "",
      generate_password: true,
    });
    expect((init.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-runtime-test");
    expect(result.generated_password).toBe("one-time-secret");
  });

  it("sends If-Match on rename/status, rotate, and revoke", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ credential }));

    await updateRuntimeCredential(credential, "customer.renamed", "disabled", fetcher as typeof fetch);
    await rotateRuntimeCredential(credential, "", true, fetcher as typeof fetch);
    await revokeRuntimeCredential(credential, fetcher as typeof fetch);

    expect(fetcher).toHaveBeenCalledTimes(3);
    for (const call of fetcher.mock.calls) {
      const [, init] = call as unknown as [string, RequestInit];
      expect((init.headers as Record<string, string>)["If-Match"]).toBe("7");
      expect((init.headers as Record<string, string>)["X-CSRF-Token"]).toBe("csrf-runtime-test");
      expect((init.headers as Record<string, string>)["Idempotency-Key"]).toMatch(/^runtime-/);
    }
  });

  it("preserves stale-revision API code and status for the UI", async () => {
    const fetcher = vi.fn(async () => jsonResponse({ code: "revision_conflict", message: "Refresh and retry." }, 409));

    await expect(updateRuntimeCredential(credential, "customer.two", "active", fetcher as typeof fetch)).rejects.toMatchObject({
      message: "Refresh and retry.",
      code: "revision_conflict",
      status: 409,
    });
  });
});
