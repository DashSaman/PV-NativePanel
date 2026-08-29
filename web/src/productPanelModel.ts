import type { Principal } from "./auth";
import type { ProductCustomer } from "./productApi";

export type ProductRole = Principal["role"];

export function canUseCustomerProduct(role: ProductRole): boolean {
  return role === "owner" || role === "admin" || role === "reseller";
}

export function canManagePlans(role: ProductRole): boolean {
  return role === "owner" || role === "admin";
}

export function canUseRawRuntime(role: ProductRole): boolean {
  return role === "owner";
}

export function formatBytes(value: number | null | undefined): string {
  if (value === null || value === undefined || !Number.isFinite(value) || value < 0) return "—";
  const gib = value / 1073741824;
  if (gib >= 1 || value === 0) return `${gib.toLocaleString("fa-IR", { maximumFractionDigits: 2 })} GB`;
  const mib = value / 1048576;
  return `${mib.toLocaleString("fa-IR", { maximumFractionDigits: 1 })} MB`;
}

export function formatPanelDate(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("fa-IR", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

export type UsagePresentation = {
  exact: boolean;
  used: string;
  remaining: string;
  presence: string;
  sessions: string;
  lastOnline: string;
};

export function usagePresentation(customer: ProductCustomer): UsagePresentation {
  const usage = customer.usage;
  const exact = Boolean(customer.usage_capability?.available && usage?.available && usage.accounting_complete);
  if (!exact || !usage) {
    return {
      exact: false,
      used: "نامشخص",
      remaining: "نامشخص",
      presence: "نامشخص",
      sessions: "—",
      lastOnline: "—",
    };
  }
  return {
    exact: true,
    used: formatBytes(usage.used_bytes),
    remaining: customer.quota_bytes === null ? "نامحدود" : formatBytes(usage.remaining_bytes),
    presence: usage.online ? "آنلاین" : "آفلاین",
    sessions: String(usage.session_count),
    lastOnline: formatPanelDate(usage.last_online),
  };
}
