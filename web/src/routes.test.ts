import { describe, expect, it } from "vitest";
import { appRoutes, assertRouteManifest } from "./routes";

describe("route manifest", () => {
  it("is valid", () => expect(() => assertRouteManifest()).not.toThrow());
  it("contains no Iran, node or fleet page in standalone MVP", () => {
    expect(appRoutes.map((route) => route.path).join(" ")).not.toMatch(/iran|node|fleet/i);
  });
  it("keeps sensitive pages restricted", () => {
    expect(appRoutes.find((r) => r.path === "/settings/security")?.permission).toBe("owner");
    expect(appRoutes.find((r) => r.path === "/diagnostics/domain-activity")?.permission).toBe("owner");
    expect(appRoutes.find((r) => r.path === "/runtime/naive")?.permission).toBe("owner");
  });
  it("exposes the usable Naive runtime manager in authenticated navigation", () => {
    const route = appRoutes.find((r) => r.path === "/runtime/naive");
    expect(route?.navigation).toBe(true);
    expect(route?.label).toBe("Naive Runtime");
  });
  it("keeps subscription page public but outside navigation", () => {
    const route = appRoutes.find((r) => r.path === "/s/:token");
    expect(route?.permission).toBe("public");
    expect(route?.navigation).toBe(false);
  });
});
