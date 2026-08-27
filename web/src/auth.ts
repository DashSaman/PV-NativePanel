export type Principal = {
  actor_id: string;
  tenant_id?: string;
  role: "owner" | "admin" | "operator" | "auditor" | "reseller";
  email: string;
  display_name: string;
};

export type LoginInput = {
  email: string;
  password: string;
  totpCode?: string;
};

export type AuthError = Error & { code?: string; status?: number };

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
    const error = new Error(typeof body.message === "string" ? body.message : "Request failed.") as AuthError;
    error.code = typeof body.code === "string" ? body.code : undefined;
    error.status = response.status;
    throw error;
  }
  return body;
}

export function readCookie(name: string, cookieHeader = document.cookie): string | null {
  for (const part of cookieHeader.split(";")) {
    const trimmed = part.trim();
    const separator = trimmed.indexOf("=");
    if (separator < 0) continue;
    if (trimmed.slice(0, separator) !== name) continue;
    try {
      return decodeURIComponent(trimmed.slice(separator + 1));
    } catch {
      return null;
    }
  }
  return null;
}

export async function login(input: LoginInput, fetcher: Fetcher = fetch): Promise<Record<string, unknown>> {
  return request(
    "/api/v1/auth/login",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: input.email,
        password: input.password,
        ...(input.totpCode ? { totp_code: input.totpCode } : {}),
      }),
    },
    fetcher,
  );
}

export async function me(fetcher: Fetcher = fetch): Promise<Principal> {
  const body = await request("/api/v1/me", { method: "GET" }, fetcher);
  const principal = body.principal;
  if (!principal || typeof principal !== "object") throw new Error("Invalid authentication response.");
  return principal as Principal;
}

export async function logout(csrfToken: string, fetcher: Fetcher = fetch): Promise<void> {
  await request(
    "/api/v1/auth/logout",
    {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": csrfToken },
      body: "{}",
    },
    fetcher,
  );
}
