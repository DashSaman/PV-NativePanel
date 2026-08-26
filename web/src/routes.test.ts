import { describe, expect, it } from "vitest";
import { appRoutes, assertRouteManifest } from "./routes";

describe("route manifest", () => {
  it("is valid", () => expect(() => assertRouteManifest()).not.toThrow());
  it("contains no Iran, node or fleet page in standalone MVP", () => {
    const paths = appRoutes.map((route) => route.path).join(" ");
    expect(paths).not.toMatch(/iran|node|fleet/i);
  });
  it("keeps security settings owner-only", () => {
    expect(appRoutes.find((route) => route.path === "/settings/security")?.permission).toBe("owner");
  });
});
