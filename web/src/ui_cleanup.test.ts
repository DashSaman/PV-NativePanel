// @ts-expect-error Test-only Node builtin; browser bundle does not ship Node typings.
import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";

const app = readFileSync(new URL("./App.tsx", import.meta.url), "utf8");
const customers = readFileSync(new URL("./ProductCustomers.tsx", import.meta.url), "utf8");
const catalog = readFileSync(new URL("./ProductCatalog.tsx", import.meta.url), "utf8");

describe("operator UI cleanup contract", () => {
  it("collapses runtime navigation into one system entry and supports active navigation state", () => {
    expect(app).toContain("سیستم / Runtime");
    expect(app).toContain("nav-link active");
    expect(app).not.toContain('>اکانت‌های قدیمی Runtime</a>');
    expect(app).not.toContain('>Runtime پیشرفته</a>');
  });

  it("keeps the customer table compact with advanced filters and one operations menu", () => {
    expect(customers).toContain('className="more-filters"');
    expect(customers).toContain('className="row-more"');
    expect(customers).toContain('className="toggle-switch"');
    expect(customers).not.toContain("Proof-gated");
    expect(customers).not.toContain("Accounting ناقص");
    expect(customers).not.toContain("Customer product");
  });

  it("renders catalog as tabs and hides plan creation until requested", () => {
    expect(catalog).toContain('className="catalog-tabs"');
    expect(catalog).toContain("پلن جدید");
    expect(catalog).toContain("showPlanForm");
  });
});
