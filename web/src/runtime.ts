import { readCookie } from "./auth";

export type RuntimeCredentialStatus = "active" | "disabled" | "revoked";
export type RuntimeCredential = {
  id: string;
  username: string;
  status: RuntimeCredentialStatus;
  origin: "imported" | "panel";
  revision: number;
  created_at?: string;
  updated_at?: string;
  rotated_at?: string;
  revoked_at?: string;
};

export type RuntimeStatus = {
  status: string;
  runtime_available: boolean;
  caddy_sha256?: string;
};

export type RuntimeMutationResult = {
  credential: RuntimeCredential;
  runtime_revision_id?: string;
  generated_password?: string;
  generated_password_notice?: string;
};

export type RuntimeError = Error & { code?: string; status?: number };
type Fetcher = typeof fetch;

async function parseJSON(response: Response): Promise<Record<string, unknown>> {
  const contentType = response.headers.get("Content-Type") || "";
  if (!contentType.includes("application/json")) return {};
  return (await response.json()) as Record<string, unknown>;
}

async function request(path: string, init: RequestInit, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  const response = await fetcher(path, { credentials: "same-origin", ...init });
  const body = await parseJSON(response);
  if (!response.ok) {
    const error = new Error(typeof body.message === "string" ? body.message : "Runtime request failed.") as RuntimeError;
    error.code = typeof body.code === "string" ? body.code : undefined;
    error.status = response.status;
    throw error;
  }
  return body;
}

function mutationHeaders(expectedRevision?: number): Record<string, string> {
  const csrf = readCookie("__Host-pvnaive_csrf");
  if (!csrf) throw new Error("CSRF token is unavailable.");
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-CSRF-Token": csrf,
    "Idempotency-Key": `runtime-${crypto.randomUUID()}`,
  };
  if (expectedRevision !== undefined) headers["If-Match"] = String(expectedRevision);
  return headers;
}

export async function getRuntimeStatus(fetcher: Fetcher = fetch): Promise<RuntimeStatus> {
  return (await request("/api/v1/runtime/naive", { method: "GET" }, fetcher)) as unknown as RuntimeStatus;
}

export async function listRuntimeCredentials(fetcher: Fetcher = fetch): Promise<RuntimeCredential[]> {
  const body = await request("/api/v1/runtime/naive/credentials", { method: "GET" }, fetcher);
  return Array.isArray(body.credentials) ? (body.credentials as RuntimeCredential[]) : [];
}

export async function importCurrentRuntime(fetcher: Fetcher = fetch): Promise<RuntimeCredential[]> {
  const body = await request(
    "/api/v1/runtime/naive/import",
    { method: "POST", headers: mutationHeaders(), body: "{}" },
    fetcher,
  );
  return Array.isArray(body.credentials) ? (body.credentials as RuntimeCredential[]) : [];
}

export async function createRuntimeCredential(
  username: string,
  password: string,
  generatePassword: boolean,
  fetcher: Fetcher = fetch,
): Promise<RuntimeMutationResult> {
  return (await request(
    "/api/v1/runtime/naive/credentials",
    {
      method: "POST",
      headers: mutationHeaders(),
      body: JSON.stringify({ username, password, generate_password: generatePassword }),
    },
    fetcher,
  )) as unknown as RuntimeMutationResult;
}

export async function updateRuntimeCredential(
  credential: RuntimeCredential,
  username: string,
  status: RuntimeCredentialStatus,
  fetcher: Fetcher = fetch,
): Promise<RuntimeMutationResult> {
  return (await request(
    `/api/v1/runtime/naive/credentials/${encodeURIComponent(credential.id)}`,
    {
      method: "PATCH",
      headers: mutationHeaders(credential.revision),
      body: JSON.stringify({ username, status }),
    },
    fetcher,
  )) as unknown as RuntimeMutationResult;
}

export async function rotateRuntimeCredential(
  credential: RuntimeCredential,
  password: string,
  generatePassword: boolean,
  fetcher: Fetcher = fetch,
): Promise<RuntimeMutationResult> {
  return (await request(
    `/api/v1/runtime/naive/credentials/${encodeURIComponent(credential.id)}/rotate-password`,
    {
      method: "POST",
      headers: mutationHeaders(credential.revision),
      body: JSON.stringify({ password, generate_password: generatePassword }),
    },
    fetcher,
  )) as unknown as RuntimeMutationResult;
}

export async function revokeRuntimeCredential(
  credential: RuntimeCredential,
  fetcher: Fetcher = fetch,
): Promise<RuntimeMutationResult> {
  return (await request(
    `/api/v1/runtime/naive/credentials/${encodeURIComponent(credential.id)}`,
    { method: "DELETE", headers: mutationHeaders(credential.revision), body: "{}" },
    fetcher,
  )) as unknown as RuntimeMutationResult;
}
