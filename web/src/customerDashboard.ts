import type { UnifiedCustomerRow } from "./customerRows";

const GIB = 1024 * 1024 * 1024;
const DAY = 24 * 60 * 60 * 1000;

export type CustomerDashboard = {
  totalAccounts: number;
  managedAccounts: number;
  needsSetup: number;
  active: number;
  pending: number;
  suspended: number;
  ended: number;
  configuredQuotaGB: number;
  unlimitedAccounts: number;
  usageProven: boolean;
  expiry: {
    within7Days: number;
    within30Days: number;
    later: number;
    noExpiry: number;
  };
};

export function buildCustomerDashboard(rows: UnifiedCustomerRow[], now = new Date()): CustomerDashboard {
  const managed = rows.flatMap((row) => row.kind === "customer" ? [row.customer] : []);
  const needsSetup = rows.filter((row) => row.kind === "runtime").length;
  let active = 0;
  let pending = 0;
  let suspended = 0;
  let ended = 0;
  let configuredQuotaBytes = 0;
  let unlimitedAccounts = 0;
  const expiry = { within7Days: 0, within30Days: 0, later: 0, noExpiry: needsSetup };

  for (const customer of managed) {
    if (customer.status === "suspended") suspended += 1;
    else if (customer.status === "revoked" || customer.service_state === "expired" || customer.service_state === "quota_depleted" || customer.service_state === "ended" || customer.service_state === "revoked") ended += 1;
    else if (customer.service_state === "pending") pending += 1;
    else if (customer.status === "active" && customer.service_state === "active") active += 1;

    if (customer.quota_bytes === null) unlimitedAccounts += 1;
    else if (customer.quota_bytes > 0) configuredQuotaBytes += customer.quota_bytes;

    if (!customer.expires_at) {
      expiry.noExpiry += 1;
      continue;
    }
    const expires = new Date(customer.expires_at).getTime();
    const delta = expires - now.getTime();
    if (!Number.isFinite(expires) || delta <= 0) continue;
    if (delta <= 7 * DAY) expiry.within7Days += 1;
    else if (delta <= 30 * DAY) expiry.within30Days += 1;
    else expiry.later += 1;
  }

  return {
    totalAccounts: rows.length,
    managedAccounts: managed.length,
    needsSetup,
    active,
    pending,
    suspended,
    ended,
    configuredQuotaGB: Math.round(configuredQuotaBytes / GIB),
    unlimitedAccounts,
    usageProven: managed.length > 0 && managed.every((customer) => customer.usage_capability.available),
    expiry,
  };
}
