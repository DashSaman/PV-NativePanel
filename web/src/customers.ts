import { readCookie } from "./auth";

export type CustomerValidityMode = "on_creation" | "on_first_successful_connection" | "fixed_expiry";

export type CreateCustomerRequest = {
  username: string;
  password: string;
  generate_password: boolean;
  quota_gb: number | null;
  validity: {
    mode: CustomerValidityMode;
    duration_days?: number;
    expires_at?: string;
  };
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
  };
  runtime_credential: {
    id: string;
    username: string;
    status: string;
  };
  generated_password?: string;
  subscription_path: string;
  delivery_notice?: string;
  usage_capability: {
    available: boolean;
    reason?: string;
  };
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

export async function createCustomer(
  input: CreateCustomerRequest,
  fetcher: Fetcher = fetch,
): Promise<CustomerCreateResult> {
  const csrf = readCookie("__Host-pvnaive_csrf", browserCookieSource());
  if (!csrf) throw new Error("CSRF token is unavailable.");

  const response = await fetcher("/api/v1/customers", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrf,
      "Idempotency-Key": `customer-${crypto.randomUUID()}`,
    },
    body: JSON.stringify(input),
  });
  const body = await parseJSON(response);
  if (!response.ok) {
    const error = new Error(
      typeof body.message === "string" ? body.message : "Customer request failed.",
    ) as CustomerAPIError;
    error.code = typeof body.code === "string" ? body.code : undefined;
    error.status = response.status;
    throw error;
  }
  return body as unknown as CustomerCreateResult;
}

export function subscriptionURL(path: string, base?: string): string {
  const source = base || (typeof window === "undefined" ? "https://localhost/" : window.location.href);
  return new URL(path, source).toString();
}
