// @ts-expect-error Test-only Node builtin; browser bundle does not ship Node typings.
import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const source = readFileSync(new URL("./ProductCustomers.tsx", import.meta.url), "utf8");

describe("usage reset UI contract", () => {
  it("requires explicit confirmation and states identity invariants for manual reset", () => {
    expect(source).toContain("Reset مصرف");
    expect(source).toContain("مصرف فعلی");
    expect(source).toContain("صفر شود");
    expect(source).toContain("Password و Subscription تغییر نمی‌کنند");
  });

  it("offers reset_usage in the bulk preview flow", () => {
    expect(source).toContain('<option value="reset_usage">');
    expect(source).toContain("Password و Subscription");
    expect(source).toContain("per-item");
  });
});
