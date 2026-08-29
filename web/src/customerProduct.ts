export type CustomerLifecycle = "active" | "disabled" | "suspended" | "revoked";
export type CustomerCommercial = "pending_first_use" | "active" | "expired" | "depleted" | "on_hold";
export type CustomerPresence = "online" | "offline" | "unknown";
export type CustomerQuota = "unlimited" | "healthy" | "warning" | "depleted" | "unavailable";
export type CustomerRuntime = "healthy" | "degraded" | "down" | "unknown";

export type CustomerProductStatus = {
  lifecycle: CustomerLifecycle;
  commercial: CustomerCommercial;
  presence: CustomerPresence;
  quota: CustomerQuota;
  runtime: CustomerRuntime;
};

export type CustomerSort = "username" | "created" | "updated" | "expiry" | "last_renewal";
export type SortDirection = "asc" | "desc";
export type CustomerColumn = "username" | "status" | "plan" | "group" | "tags" | "quota" | "expiry" | "created" | "updated" | "last_renewal" | "reseller" | "actions";

export const DEFAULT_CUSTOMER_COLUMNS: CustomerColumn[] = ["username", "status", "plan", "quota", "expiry", "actions"];

export type CustomerFilters = {
  q: string;
  status: string;
  plan: string;
  group: string;
  tag: string;
  reseller: string;
  expiryFrom: string;
  expiryTo: string;
  unlimitedVolume: boolean | null;
  unlimitedExpiry: boolean | null;
  page: number;
  pageSize: 10 | 20 | 25 | 50 | 100;
  sort: CustomerSort;
  direction: SortDirection;
};

export const DEFAULT_CUSTOMER_FILTERS: CustomerFilters = {
  q: "", status: "", plan: "", group: "", tag: "", reseller: "", expiryFrom: "", expiryTo: "",
  unlimitedVolume: null, unlimitedExpiry: null, page: 1, pageSize: 50, sort: "updated", direction: "desc",
};

const allowedStatuses = new Set(["active", "disabled", "suspended", "revoked", "expired", "depleted", "pending", "on_hold"]);
const allowedPageSizes = new Set([10, 20, 25, 50, 100]);
const allowedSort = new Set<CustomerSort>(["username", "created", "updated", "expiry", "last_renewal"]);

function positiveInt(raw: string | null, fallback: number): number {
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function triState(raw: string | null): boolean | null {
  if (raw === "true") return true;
  if (raw === "false") return false;
  return null;
}

export function customerFiltersFromSearch(search: string): CustomerFilters {
  const params = new URLSearchParams(search.startsWith("?") ? search.slice(1) : search);
  const size = positiveInt(params.get("page_size"), DEFAULT_CUSTOMER_FILTERS.pageSize);
  const sort = params.get("sort") as CustomerSort | null;
  const direction = params.get("dir");
  const status = params.get("status") || "";
  return {
    q: (params.get("q") || "").trim(),
    status: allowedStatuses.has(status) ? status : "",
    plan: params.get("plan") || "",
    group: params.get("group") || "",
    tag: params.get("tag") || "",
    reseller: params.get("reseller") || "",
    expiryFrom: params.get("expiry_from") || "",
    expiryTo: params.get("expiry_to") || "",
    unlimitedVolume: triState(params.get("unlimited_volume")),
    unlimitedExpiry: triState(params.get("unlimited_expiry")),
    page: positiveInt(params.get("page"), 1),
    pageSize: (allowedPageSizes.has(size) ? size : 50) as CustomerFilters["pageSize"],
    sort: sort && allowedSort.has(sort) ? sort : "updated",
    direction: direction === "asc" || direction === "desc" ? direction : "desc",
  };
}

export function customerFiltersToSearch(filters: CustomerFilters): URLSearchParams {
  const params = new URLSearchParams();
  const set = (key: string, value: string) => { if (value) params.set(key, value); };
  set("q", filters.q.trim()); set("status", filters.status); set("plan", filters.plan); set("group", filters.group);
  set("tag", filters.tag); set("reseller", filters.reseller); set("expiry_from", filters.expiryFrom); set("expiry_to", filters.expiryTo);
  if (filters.unlimitedVolume !== null) params.set("unlimited_volume", String(filters.unlimitedVolume));
  if (filters.unlimitedExpiry !== null) params.set("unlimited_expiry", String(filters.unlimitedExpiry));
  if (filters.page !== 1) params.set("page", String(filters.page));
  if (filters.pageSize !== 50) params.set("page_size", String(filters.pageSize));
  if (filters.sort !== "updated") params.set("sort", filters.sort);
  if (filters.direction !== "desc") params.set("dir", filters.direction);
  return params;
}

const quotaMap: Record<CustomerQuota, { label: string; tone: string }> = {
  unlimited: { label: "نامحدود", tone: "neutral" },
  healthy: { label: "سالم", tone: "green" },
  warning: { label: "نزدیک سقف", tone: "orange" },
  depleted: { label: "تمام‌شده", tone: "red" },
  unavailable: { label: "مصرف نامشخص", tone: "gray" },
};

export function quotaPresentation(value: CustomerQuota) { return quotaMap[value]; }
