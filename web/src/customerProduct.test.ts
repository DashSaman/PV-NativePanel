import { describe, expect, it } from "vitest";
import {
  DEFAULT_CUSTOMER_COLUMNS,
  customerFiltersFromSearch,
  customerFiltersToSearch,
  quotaPresentation,
  type CustomerProductStatus,
} from "./customerProduct";

describe("customer product filters", () => {
  it("round-trips stable filters through URL search params", () => {
    const source = "?q=amir&status=suspended&plan=p1&group=g1&tag=vip&page=3&page_size=50&sort=expiry&dir=asc";
    const filters = customerFiltersFromSearch(source);
    expect(filters.q).toBe("amir");
    expect(filters.status).toBe("suspended");
    expect(filters.page).toBe(3);
    expect(filters.pageSize).toBe(50);
    const roundTrip = customerFiltersFromSearch(`?${customerFiltersToSearch(filters).toString()}`);
    expect(roundTrip).toEqual(filters);
  });

  it("uses safe defaults for malformed pagination and sorting", () => {
    const filters = customerFiltersFromSearch("?page=-2&page_size=999&sort=hacker&dir=drop");
    expect(filters.page).toBe(1);
    expect(filters.pageSize).toBe(50);
    expect(filters.sort).toBe("updated");
    expect(filters.direction).toBe("desc");
  });
});

describe("truthful status presentation", () => {
  it("does not convert unavailable finite quota into healthy usage", () => {
    const status: CustomerProductStatus = {
      lifecycle: "active",
      commercial: "active",
      presence: "unknown",
      quota: "unavailable",
      runtime: "unknown",
    };
    expect(quotaPresentation(status.quota).label).toContain("نامشخص");
  });

  it("keeps the default table compact", () => {
    expect(DEFAULT_CUSTOMER_COLUMNS).toEqual(["username", "status", "plan", "quota", "expiry", "actions"]);
  });
});
