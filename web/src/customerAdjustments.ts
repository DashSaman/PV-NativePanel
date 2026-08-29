import { readCookie } from "./auth";
import type { CustomerCreateResult, CustomerAPIError } from "./customers";

type Fetcher = typeof fetch;

export type CustomerAdjustmentResult = {
  service_term: CustomerCreateResult["service_term"];
  runtime_mutated: boolean;
  message?: string;
};

function cookieSource(): string {
  return typeof document === "undefined" ? "" : document.cookie;
}

function headers(): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf", cookieSource());
  if (!csrf) throw new Error("CSRF token is unavailable.");
  return {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrf,
    "Idempotency-Key": `customer-${crypto.randomUUID()}`,
  };
}

async function parse(response: Response): Promise<Record<string, unknown>> {
  const contentType = response.headers.get("Content-Type") || "";
  return contentType.includes("application/json") ? ((await response.json()) as Record<string, unknown>) : {};
}

function responseError(body: Record<string, unknown>, status: number): CustomerAPIError {
  const error = new Error(typeof body.message === "string" ? body.message : "Customer adjustment failed.") as CustomerAPIError;
  error.code = typeof body.code === "string" ? body.code : undefined;
  error.status = status;
  return error;
}

async function postAdjustment(
  path: string,
  body: Record<string, number>,
  fetcher: Fetcher,
): Promise<CustomerAdjustmentResult> {
  const response = await fetcher(path, {
    method: "POST",
    credentials: "same-origin",
    headers: headers(),
    body: JSON.stringify(body),
  });
  const payload = await parse(response);
  if (!response.ok) throw responseError(payload, response.status);
  return payload as unknown as CustomerAdjustmentResult;
}

export function addCustomerVolume(customerID: string, deltaGB: number, fetcher: Fetcher = fetch) {
  return postAdjustment(
    `/api/v1/customers/${encodeURIComponent(customerID)}/volume/add`,
    { delta_gb: deltaGB },
    fetcher,
  );
}

export function extendCustomerTime(customerID: string, days: number, fetcher: Fetcher = fetch) {
  return postAdjustment(
    `/api/v1/customers/${encodeURIComponent(customerID)}/validity/extend`,
    { days },
    fetcher,
  );
}
