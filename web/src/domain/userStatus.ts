export type AccountStatus =
  | "draft" | "on_hold" | "active" | "suspended" | "expired" | "depleted" | "revoked" | "error";
export type Presence = "online" | "idle" | "offline" | "unknown";
export type QuotaState = "unlimited" | "healthy" | "warning" | "critical" | "depleted";

export const accountStatusPresentation: Record<AccountStatus, { label: string; tone: string }> = {
  draft: { label: "پیش‌نویس", tone: "neutral" },
  on_hold: { label: "در انتظار اولین اتصال", tone: "purple" },
  active: { label: "فعال", tone: "green" },
  suspended: { label: "غیرفعال دستی", tone: "orange" },
  expired: { label: "منقضی", tone: "red" },
  depleted: { label: "حجم تمام", tone: "red-strong" },
  revoked: { label: "لغوشده", tone: "gray" },
  error: { label: "خطای اعمال", tone: "magenta" }
};

export function quotaState(used: number, limit: number | null): QuotaState {
  if (limit === null || limit === 0) return "unlimited";
  if (used >= limit) return "depleted";
  const ratio = used / limit;
  if (ratio >= 0.95) return "critical";
  if (ratio >= 0.8) return "warning";
  return "healthy";
}

export function concurrencyLabel(limit: number | null): string {
  if (limit === null) return "چندکاربره نامحدود";
  if (limit === 1) return "تک‌کاربره";
  return `چندکاربره تا ${limit} اتصال`;
}

export function deriveEffectiveStatus(input: {
  configured: Exclude<AccountStatus, "expired" | "depleted">;
  expiresAt: number | null;
  now: number;
  used: number;
  limit: number | null;
}): AccountStatus {
  if (["draft", "on_hold", "suspended", "revoked", "error"].includes(input.configured)) return input.configured;
  if (input.expiresAt !== null && input.expiresAt <= input.now) return "expired";
  if (quotaState(input.used, input.limit) === "depleted") return "depleted";
  return "active";
}
