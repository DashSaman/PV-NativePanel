import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

describe("RuntimeNaive Karing handoff", () => {
  it("keeps raw Naive URI copy separate from an explicit Karing sing-box profile copy action", () => {
    const source = readFileSync(new URL("./RuntimeNaive.tsx", import.meta.url), "utf8");

    expect(source).toContain("buildKaringSingBoxProfile");
    expect(source).toContain("const karingProfile");
    expect(source).toContain("navigator.clipboard.writeText(karingProfile)");
    expect(source).toContain("کپی کانفیگ Karing");
    expect(source).toContain("navigator.clipboard.writeText(customerURI)");
    expect(source).toContain("کپی لینک Naive");
  });
});
