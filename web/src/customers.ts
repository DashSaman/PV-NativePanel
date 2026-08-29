import { readCookie } from "./auth";

export type CustomerValidityMode = "on_creation" | "on_first_successful_connection" | "fixed_expiry";

export type CustomerServiceSettingsRequest = {
  quota_gb: number | null;
  validity: {
    mode: CustomerValidityMode;
    duration_days?: number;
    expires_at?: string;
  };
};

export type CreateCustomerRequest = CustomerServiceSettingsRequest & {
  username: string;
  password: string;
  generate_password: boolean;
};

export type UsageCapability = {
  available: boolean;
  reason?: string;
};

export type CustomerCreateResult = {
  user: {
    id: string;
    username: string;
    display_name?: string;
    status: string;
    revision: number;
  };
  service_term: {
    id: string;
    state: string;
    quota_bytes?: number | null;
    duration_seconds: number;
    start_policy: string;
    starts_at?: string;
    expires_at?: string;
    revision?: number;
  };
  runtime_credential: {
    id: string;
    username: string;
    status: string;
  };
  generated_password?: string;
  subscription_path: string;
  delivery_notice?: string;
  usage_capability: UsageCapability;
};

export type CustomerView = {
  id: string;
  username: string;
  status: string;
  service_term_id: string;
  service_state: string;
  quota_bytes: number | null;
  duration_seconds: number;
  start_policy: string;
  starts_at?: string;
  first_connected_at?: string;
  expires_at?: string;
  runtime_credential_id: string;
  subscription_available: boolean;
  subscription_retrievable?: boolean;
  usage_capability: UsageCapability;
  upload_bytes: number;
  download_bytes: number;
  used_bytes: number;
  remaining_bytes?: number | null;
  accounting_complete: boolean;
  online: boolean;
  online_sessions: number;
  last_online?: string;
};

export type SubscriptionDelivery = {
  subscription_path: string;
  direct_uri?: string;
  delivery_notice?: string;
};

export type CustomerServiceUpdateResult = {
  service_term: CustomerCreateResult["service_term"];
  runtime_mutated: boolean;
  message?: string;
};

export type CustomerLifecycleResult = {
  status: string;
  runtime_credential?: CustomerCreateResult["runtime_credential"];
  message?: string;
};

export type PasswordRotationRequest = {
  password: string;
  generate_password: boolean;
};

export type PasswordRotationResult = {
  runtime_credential: CustomerCreateResult["runtime_credential"];
  generated_password?: string;
  delivery_notice?: string;
};

export type CustomerAPIError = Error & { code?: string; status?: number };
type Fetcher = typeof fetch;

function browserCookieSource(): string {
  return typeof document === "undefined" ? "" : document.cookie;
}

async function parseJSON(response: Response): Promise<Record<string, unknown>> {
  const contentType = response.headers.get("Content-Type") || "";
  if (!contentType.includes("application/json")) return {};
  return (await response.json()) as Record<string, unknown>;
}

function mutationHeaders(): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf", browserCookieSource());
  if (!csrf) throw new Error("CSRF token is unavailable.");
  return {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrf,
    "Idempotency-Key": `customer-${crypto.randomUUID()}`,
  };
}

function apiError(body: Record<string, unknown>, status: number): CustomerAPIError {
  const error = new Error(
    typeof body.message === "string" ? body.message : "Customer request failed.",
  ) as CustomerAPIError;
  error.code = typeof body.code === "string" ? body.code : undefined;
  error.status = status;
  return error;
}

export async function createCustomer(
  input: CreateCustomerRequest,
  fetcher: Fetcher = fetch,
): Promise<CustomerCreateResult> {
  const response = await fetcher("/api/v1/customers", {
    method: "POST",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: JSON.stringify(input),
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as CustomerCreateResult;
}

export async function adoptRuntimeCustomer(
  runtimeCredentialID: string,
  settings: CustomerServiceSettingsRequest,
  fetcher: Fetcher = fetch,
): Promise<CustomerCreateResult> {
  const response = await fetcher("/api/v1/customers/adopt-runtime", {
    method: "POST",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: JSON.stringify({ runtime_credential_id: runtimeCredentialID, ...settings }),
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as CustomerCreateResult;
}

export async function updateCustomerService(
  customerID: string,
  settings: CustomerServiceSettingsRequest,
  fetcher: Fetcher = fetch,
): Promise<CustomerServiceUpdateResult> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}/service`, {
    method: "PATCH",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: JSON.stringify(settings),
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as CustomerServiceUpdateResult;
}

export async function listCustomers(fetcher: Fetcher = fetch): Promise<CustomerView[]> {
  const response = await fetcher("/api/v1/customers", {
    method: "GET",
    credentials: "same-origin",
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return Array.isArray(body.customers) ? (body.customers as CustomerView[]) : [];
}

export async function getCurrentSubscription(
  customerID: string,
  fetcher: Fetcher = fetch,
): Promise<SubscriptionDelivery> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}/subscription`, {
    method: "GET",
    credentials: "same-origin",
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as SubscriptionDelivery;
}

export async function rotateSubscription(
  customerID: string,
  fetcher: Fetcher = fetch,
): Promise<SubscriptionDelivery> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}/subscription/rotate`, {
    method: "POST",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: "{}",
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as SubscriptionDelivery;
}

async function customerLifecycleMutation(
  customerID: string,
  action: "suspend" | "resume",
  fetcher: Fetcher,
): Promise<CustomerLifecycleResult> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}/${action}`, {
    method: "POST",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: "{}",
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as CustomerLifecycleResult;
}

export async function suspendCustomer(
  customerID: string,
  fetcher: Fetcher = fetch,
): Promise<CustomerLifecycleResult> {
  return customerLifecycleMutation(customerID, "suspend", fetcher);
}

export async function resumeCustomer(
  customerID: string,
  fetcher: Fetcher = fetch,
): Promise<CustomerLifecycleResult> {
  return customerLifecycleMutation(customerID, "resume", fetcher);
}

export async function deleteCustomer(
  customerID: string,
  fetcher: Fetcher = fetch,
): Promise<CustomerLifecycleResult> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}`, {
    method: "DELETE",
    credentials: "same-origin",
    headers: mutationHeaders(),
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as CustomerLifecycleResult;
}

export async function rotateCustomerPassword(
  customerID: string,
  input: PasswordRotationRequest,
  fetcher: Fetcher = fetch,
): Promise<PasswordRotationResult> {
  const response = await fetcher(`/api/v1/customers/${encodeURIComponent(customerID)}/rotate-password`, {
    method: "POST",
    credentials: "same-origin",
    headers: mutationHeaders(),
    body: JSON.stringify(input),
  });
  const body = await parseJSON(response);
  if (!response.ok) throw apiError(body, response.status);
  return body as unknown as PasswordRotationResult;
}

export function subscriptionURL(path: string, base?: string): string {
  const source = base || (typeof window === "undefined" ? "https://localhost/" : window.location.href);
  return new URL(path, source).toString();
}

export function quotaLabel(bytes: number | null): string {
  if (bytes === null || bytes === undefined) return "نامحدود";
  return `${Math.round(bytes / 1073741824)} GB`;
}

export function trafficLabel(bytes: number): string {
  const safe = Number.isFinite(bytes) && bytes > 0 ? bytes : 0;
  const gib = 1024 ** 3;
  const mib = 1024 ** 2;
  if (safe >= gib) return `${Math.round((safe / gib) * 10) / 10} GB`;
  if (safe >= mib) return `${Math.round(safe / mib)} MB`;
  if (safe >= 1024) return `${Math.round(safe / 1024)} KB`;
  return `${Math.round(safe)} B`;
}

export function customerUsagePresentation(customer: CustomerView): { primary: string; secondary: string; exact: boolean } {
  if (!customer.usage_capability.available || !customer.accounting_complete) {
    return {
      primary: `حداقل ${trafficLabel(customer.used_bytes)} ثبت‌شده`,
      secondary: "Accounting ناقص · باقی‌مانده نامشخص",
      exact: false,
    };
  }
  const primary = `${trafficLabel(customer.used_bytes)} مصرف`;
  const remaining = customer.quota_bytes === null
    ? "حجم نامحدود"
    : `${trafficLabel(customer.remaining_bytes ?? 0)} مانده`;
  const presence = customer.online
    ? `آنلاین (${Math.max(1, customer.online_sessions)})`
    : "آفلاین";
  return { primary, secondary: `${remaining} · ${presence}`, exact: true };
}

export function expiryLabel(value?: string): string {
  if (!value) return "پس از اولین اتصال";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : date.toLocaleString("fa-IR");
}