// @ts-expect-error Test-only Node builtin; browser bundle does not ship Node typings.
import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const source = readFileSync(new URL("./ProductCustomers.tsx", import.meta.url), "utf8");

describe("manual usage reset UI contract", () => {
  it("requires a visible confirmation and uses the dedicated reset API", () => {
    expect(source).toContain("resetProductUsage");
    expect(source).toContain("مصرف فعلی");
    expect(source).toContain("صفر شود");
    expect(source).toContain("Password و Subscription تغییر نمی‌کنند");
  });
});
