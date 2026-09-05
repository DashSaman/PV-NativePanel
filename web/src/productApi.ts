import { readCookie } from "./auth";

export type ProductStatusDimensions = {
  lifecycle: "active" | "disabled" | "suspended" | "revoked";
  commercial: "pending_first_use" | "active" | "expired" | "depleted" | "on_hold";
  presence: "online" | "offline" | "unknown";
  quota: "unlimited" | "healthy" | "warning" | "depleted" | "unavailable";
  runtime: "healthy" | "degraded" | "down" | "unknown";
};

export type ProductGroup = { id: string; name: string; enabled: boolean; sort_order: number };
export type ProductTag = { id: string; name: string; enabled: boolean; sort_order: number };

export type ProductUsage = {
  available: boolean;
  accounting_complete: boolean;
  upload_bytes: number;
  download_bytes: number;
  used_bytes: number;
  remaining_bytes: number | null;
  last_online?: string;
  online: boolean;
  session_count: number;
};

export type ProductCustomer = {
  id: string;
  username: string;
  display_name?: string;
  status: string;
  status_dimensions: ProductStatusDimensions;
  service_term_id: string;
  service_state: string;
  plan_id?: string;
  plan_name?: string;
  quota_bytes: number | null;
  duration_seconds: number;
  no_expiry: boolean;
  start_policy: string;
  starts_at?: string;
  first_connected_at?: string;
  expires_at?: string;
  runtime_credential_id: string;
  subscription_available: boolean;
  subscription_retrievable: boolean;
  usage_capability: { available: boolean; reason?: string };
  usage?: ProductUsage;
  note?: string;
  group?: ProductGroup;
  tags?: ProductTag[];
  assigned_actor_id?: string;
  created_by_actor_id?: string;
  reseller_id?: string;
  created_at?: string;
  updated_at?: string;
  last_renewal_at?: string;
  on_hold: boolean;
  next_plan_id?: string;
  next_plan_name?: string;
};

export type ProductPlan = {
  id: string;
  name: string;
  quota_bytes: number | null;
  validity_seconds?: number;
  no_expiry: boolean;
  start_policy: "on_creation" | "on_first_successful_connection";
  reset_strategy: "none" | "daily" | "weekly" | "monthly" | "yearly" | "custom";
  reset_custom_days?: number;
  concurrency_limit: number | null;
  unique_ip_limit: number | null;
  default_group_id?: string;
  tag_ids?: string[];
  enabled: boolean;
  sort_order: number;
  reset_enforcement_available?: boolean;
};

export type ProductFilters = {
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
  sort: "username" | "created" | "updated" | "expiry" | "last_renewal";
  direction: "asc" | "desc";
};

export const DEFAULT_PRODUCT_FILTERS: ProductFilters = {
  q: "", status: "", plan: "", group: "", tag: "", reseller: "",
  expiryFrom: "", expiryTo: "", unlimitedVolume: null, unlimitedExpiry: null,
  page: 1, pageSize: 50, sort: "updated", direction: "desc",
};

export type ProductCustomerPage = {
  customers: ProductCustomer[];
  page: number;
  page_size: number;
  total: number;
};

export type ProductMetadataUpdate = {
  note?: string;
  group_id?: string;
  on_hold?: boolean;
  add_tag_ids?: string[];
  remove_tag_ids?: string[];
  next_plan_id?: string;
};

export type ProductCreateInput = {
  username: string;
  password?: string;
  generate_password: boolean;
  plan_id?: string;
  quota_gb?: number;
  unlimited_quota?: boolean;
  no_expiry?: boolean;
  validity?: { mode: "on_creation" | "on_first_successful_connection" | "fixed_expiry"; duration_days?: number; expires_at?: string };
  group_id?: string;
  tag_ids?: string[];
  on_hold?: boolean;
};

export type ProductCreateResult = {
  user: { id: string; username: string; status: string };
  service_term: { id: string; state: string };
  runtime_credential: { id: string; username: string; status: string };
  generated_password?: string;
  subscription_path: string;
  account_page_path?: string;
  delivery_notice?: string;
};

export type ProductSubscriptionDelivery = {
  subscription_path: string;
  account_page_path?: string;
  direct_uri?: string;
  delivery_notice?: string;
};


export type ProductUsageResetResult = {
  reset_event: {
    id: string;
    user_id: string;
    service_term_id: string;
    reason: "manual" | "bulk" | "scheduled";
    reset_at: string;
    previous_upload_bytes: number;
    previous_download_bytes: number;
    previous_used_bytes: number;
  };
  idempotent_replay: boolean;
  runtime_mutated: false;
  password_rotated: false;
  subscription_reissued: false;
  message?: string;
};

export type ProductPasswordRotationInput = {
  password: string;
  generate_password: boolean;
};

export type ProductPasswordRotationResult = {
  runtime_credential?: { id: string; username: string; status: string };
  generated_password?: string;
  delivery_notice?: string;
};

export type ProductRenewalInput =
  | { mode: "renew_current" }
  | { mode: "renew_plan"; plan_id: string }
  | { mode: "next_plan" }
  | {
      mode: "custom";
      quota_gb?: number;
      unlimited_quota?: boolean;
      no_expiry?: boolean;
      validity?: { mode: "on_creation" | "on_first_successful_connection" | "fixed_expiry"; duration_days?: number; expires_at?: string };
    };

export type ProductBulkAction =
  | "enable" | "suspend" | "revoke" | "safe_delete" | "extend_days" | "add_volume"
  | "set_volume" | "apply_plan" | "assign_group" | "add_tag" | "remove_tag" | "reissue_subscription" | "reset_usage";

export type ProductBulkRequest = {
  action: ProductBulkAction;
  customer_ids: string[];
  days?: number;
  volume_gb?: number;
  total_volume_gb?: number;
  unlimited_volume?: boolean;
  plan_id?: string;
  group_id?: string;
  tag_id?: string;
};

export type ProductBulkOperation = {
  id: string;
  action: ProductBulkAction;
  status: string;
  preview: {
    requested: number;
    affected: number;
    changes: string[];
    conflicts: Array<{ id: string; reason: string }>;
    skipped: Array<{ id: string; reason: string }>;
    invalid: Array<{ id: string; reason: string }>;
  };
  result?: {
    succeeded: number;
    failed: number;
    skipped: number;
    items: Array<{ id: string; status: string; reason?: string }>;
  };
  request: ProductBulkRequest;
};

type Fetcher = typeof fetch;
type APIError = Error & { code?: string; status?: number };

function cookieSource(): string {
  return typeof document === "undefined" ? "" : document.cookie;
}

async function parseJSON(response: Response): Promise<Record<string, unknown>> {
  const type = response.headers.get("Content-Type") || "";
  if (!type.includes("application/json")) return {};
  return (await response.json()) as Record<string, unknown>;
}

function errorFrom(body: Record<string, unknown>, status: number): APIError {
  const error = new Error(typeof body.message === "string" ? body.message : "Product request failed.") as APIError;
  error.code = typeof body.code === "string" ? body.code : undefined;
  error.status = status;
  return error;
}

function csrfHeaders(idempotencyKey?: string): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf", cookieSource());
  if (!csrf) throw new Error("CSRF token is unavailable.");
  const headers: Record<string, string> = { "Content-Type": "application/json", "X-CSRF-Token": csrf };
  if (idempotencyKey) headers["Idempotency-Key"] = idempotencyKey;
  return headers;
}

function newProductKey(prefix: string): string {
  return `${prefix}-${crypto.randomUUID()}`;
}

async function requestJSON(path: string, init: RequestInit, fetcher: Fetcher): Promise<Record<string, unknown>> {
  const response = await fetcher(path, { credentials: "same-origin", ...init });
  const body = await parseJSON(response);
  if (!response.ok) throw errorFrom(body, response.status);
  return body;
}

export function productFiltersToSearch(filters: ProductFilters): URLSearchParams {
  const params = new URLSearchParams();
  const set = (key: string, value: string) => { if (value) params.set(key, value); };
  set("q", filters.q.trim()); set("status", filters.status); set("plan", filters.plan); set("group", filters.group);
  set("tag", filters.tag); set("reseller", filters.reseller); set("expiry_from", filters.expiryFrom); set("expiry_to", filters.expiryTo);
  if (filters.unlimitedVolume !== null) params.set("unlimited_volume", String(filters.unlimitedVolume));
  if (filters.unlimitedExpiry !== null) params.set("unlimited_expiry", String(filters.unlimitedExpiry));
  params.set("page", String(filters.page));
  params.set("page_size", String(filters.pageSize));
  params.set("sort", filters.sort);
  params.set("dir", filters.direction);
  return params;
}

export async function listProductCustomers(filters: ProductFilters, fetcher: Fetcher = fetch): Promise<ProductCustomerPage> {
  const query = productFiltersToSearch(filters);
  const body = await requestJSON(`/api/v1/users?${query.toString()}`, { method: "GET" }, fetcher);
  return {
    customers: Array.isArray(body.customers) ? body.customers as ProductCustomer[] : [],
    page: typeof body.page === "number" ? body.page : filters.page,
    page_size: typeof body.page_size === "number" ? body.page_size : filters.pageSize,
    total: typeof body.total === "number" ? body.total : 0,
  };
}

export async function createProductCustomer(input: ProductCreateInput, fetcher: Fetcher = fetch): Promise<ProductCreateResult> {
  const key = newProductKey("product-create");
  const body = await requestJSON("/api/v1/users", {
    method: "POST", headers: csrfHeaders(key), body: JSON.stringify(input),
  }, fetcher);
  return body as unknown as ProductCreateResult;
}

export async function updateProductCustomer(id: string, input: ProductMetadataUpdate, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return requestJSON(`/api/v1/users/${encodeURIComponent(id)}`, {
    method: "PATCH", headers: csrfHeaders(), body: JSON.stringify(input),
  }, fetcher);
}

export async function renewProductCustomer(id: string, input: ProductRenewalInput, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return requestJSON(`/api/v1/users/${encodeURIComponent(id)}/renew`, {
    method: "POST", headers: csrfHeaders(), body: JSON.stringify(input),
  }, fetcher);
}

export async function getProductSubscription(id: string, fetcher: Fetcher = fetch): Promise<ProductSubscriptionDelivery> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/subscription`, { method: "GET" }, fetcher);
  return body as unknown as ProductSubscriptionDelivery;
}

export async function reissueProductSubscription(id: string, fetcher: Fetcher = fetch): Promise<ProductSubscriptionDelivery> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/subscription/rotate`, {
    method: "POST", headers: csrfHeaders(newProductKey("product-subscription")), body: "{}",
  }, fetcher);
  return body as unknown as ProductSubscriptionDelivery;
}


export async function resetProductUsage(id: string, fetcher: Fetcher = fetch): Promise<ProductUsageResetResult> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/reset-usage`, {
    method: "POST", headers: csrfHeaders(newProductKey("product-reset-usage")), body: JSON.stringify({ confirm: true }),
  }, fetcher);
  return body as unknown as ProductUsageResetResult;
}

export async function rotateProductPassword(id: string, input: ProductPasswordRotationInput, fetcher: Fetcher = fetch): Promise<ProductPasswordRotationResult> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/rotate-password`, {
    method: "POST", headers: csrfHeaders(newProductKey("product-password")), body: JSON.stringify(input),
  }, fetcher);
  return body as unknown as ProductPasswordRotationResult;
}

async function productLifecycle(id: string, action: "suspend" | "resume" | "revoke", fetcher: Fetcher): Promise<Record<string, unknown>> {
  return requestJSON(`/api/v1/users/${encodeURIComponent(id)}/${action}`, {
    method: "POST", headers: csrfHeaders(newProductKey(`product-${action}`)), body: "{}",
  }, fetcher);
}

export async function suspendProductCustomer(id: string, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return productLifecycle(id, "suspend", fetcher);
}

export async function resumeProductCustomer(id: string, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return productLifecycle(id, "resume", fetcher);
}

export async function revokeProductCustomer(id: string, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return productLifecycle(id, "revoke", fetcher);
}

export async function listProductPlans(fetcher: Fetcher = fetch): Promise<ProductPlan[]> {
  const body = await requestJSON("/api/v1/plans", { method: "GET" }, fetcher);
  return Array.isArray(body.plans) ? body.plans as ProductPlan[] : [];
}

export async function createProductPlan(input: {
  name: string; quota_gb?: number; unlimited_quota?: boolean; validity_days?: number; no_expiry?: boolean;
  start_policy: "on_creation" | "on_first_successful_connection"; reset_strategy: ProductPlan["reset_strategy"];
  reset_custom_days?: number; concurrency_limit?: number | null; unique_ip_limit?: number | null;
  default_group_id?: string; tag_ids?: string[]; enabled?: boolean; sort_order?: number;
}, fetcher: Fetcher = fetch): Promise<ProductPlan> {
  const body = await requestJSON("/api/v1/plans", {
    method: "POST", headers: csrfHeaders(), body: JSON.stringify(input),
  }, fetcher);
  return body.plan as ProductPlan;
}

export async function listProductGroups(fetcher: Fetcher = fetch): Promise<ProductGroup[]> {
  const body = await requestJSON("/api/v1/customer-groups", { method: "GET" }, fetcher);
  return Array.isArray(body.groups) ? body.groups as ProductGroup[] : [];
}

export async function createProductGroup(name: string, sortOrder = 0, fetcher: Fetcher = fetch): Promise<ProductGroup> {
  const body = await requestJSON("/api/v1/customer-groups", {
    method: "POST", headers: csrfHeaders(), body: JSON.stringify({ name, sort_order: sortOrder }),
  }, fetcher);
  return body.group as ProductGroup;
}

export async function listProductTags(fetcher: Fetcher = fetch): Promise<ProductTag[]> {
  const body = await requestJSON("/api/v1/customer-tags", { method: "GET" }, fetcher);
  return Array.isArray(body.tags) ? body.tags as ProductTag[] : [];
}

export async function createProductTag(name: string, sortOrder = 0, fetcher: Fetcher = fetch): Promise<ProductTag> {
  const body = await requestJSON("/api/v1/customer-tags", {
    method: "POST", headers: csrfHeaders(), body: JSON.stringify({ name, sort_order: sortOrder }),
  }, fetcher);
  return body.tag as ProductTag;
}

export async function previewProductBulk(request: ProductBulkRequest, fetcher: Fetcher = fetch): Promise<{ bulk: ProductBulkOperation; idempotencyKey: string }> {
  const idempotencyKey = newProductKey("product-bulk");
  const body = await requestJSON("/api/v1/users/bulk/preview", {
    method: "POST", headers: csrfHeaders(idempotencyKey), body: JSON.stringify(request),
  }, fetcher);
  return { bulk: body.bulk as ProductBulkOperation, idempotencyKey };
}

export async function executeProductBulk(request: ProductBulkRequest, idempotencyKey: string, fetcher: Fetcher = fetch): Promise<ProductBulkOperation> {
  const body = await requestJSON("/api/v1/users/bulk/execute", {
    method: "POST", headers: csrfHeaders(idempotencyKey), body: JSON.stringify(request),
  }, fetcher);
  return body.bulk as ProductBulkOperation;
}

export type ProductActiveSession = {
  runtime_credential_id: string;
  node_id: string;
  boot_id: string;
  session_id: string;
  service_term_id: string;
  client_ip: string;
  connected_at: string;
  last_activity_at: string;
  duration_seconds: number;
  upload_bytes: number;
  download_bytes: number;
};

export type ProductActiveSessionResponse = {
  sessions: ProductActiveSession[];
  observed_at: string;
};

export async function listProductCustomerSessions(id: string, fetcher: Fetcher = fetch): Promise<ProductActiveSessionResponse> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/sessions`, { method: "GET" }, fetcher);
  return {
    sessions: Array.isArray(body.sessions) ? body.sessions as ProductActiveSession[] : [],
    observed_at: typeof body.observed_at === "string" ? body.observed_at : new Date().toISOString(),
  };
}

export type ProductSessionKillResponse = {
  status: string;
  found: boolean;
  killed: boolean;
  session_id: string;
  credential_mutated: false;
};

export async function killProductCustomerSession(id: string, sessionID: string, fetcher: Fetcher = fetch): Promise<ProductSessionKillResponse> {
  const body = await requestJSON(`/api/v1/users/${encodeURIComponent(id)}/sessions/${encodeURIComponent(sessionID)}`, {
    method: "DELETE", headers: csrfHeaders(),
  }, fetcher);
  return body as ProductSessionKillResponse;
}

export async function killProductCustomerSessionAndReload(id: string, sessionID: string, fetcher: Fetcher = fetch): Promise<ProductActiveSessionResponse> {
  await killProductCustomerSession(id, sessionID, fetcher);
  return listProductCustomerSessions(id, fetcher);
}
