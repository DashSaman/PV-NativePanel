import { describe, expect, it } from "vitest";
import { RuntimeNaive } from "./RuntimeNaive";

describe("RuntimeNaive Karing handoff", () => {
  it("keeps raw Naive URI copy separate from an explicit Karing sing-box profile copy action", () => {
    const source = RuntimeNaive.toString();

    expect(source).toContain("buildKaringSingBoxProfile");
    expect(source).toContain("karingProfile");
    expect(source).toContain("clipboard.writeText(karingProfile)");
    expect(source).toContain("کپی کانفیگ Karing");
    expect(source).toContain("clipboard.writeText(customerURI)");
    expect(source).toContain("کپی لینک Naive");
  });
});
