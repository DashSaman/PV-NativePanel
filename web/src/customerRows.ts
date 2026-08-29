import type { CustomerView } from "./customers";
import type { RuntimeCredential } from "./runtime";

export type UnifiedCustomerRow =
  | {
      kind: "customer";
      id: string;
      username: string;
      runtimeCredentialID: string;
      customer: CustomerView;
    }
  | {
      kind: "runtime";
      id: string;
      username: string;
      runtimeCredentialID: string;
      credential: RuntimeCredential;
    };

export function buildUnifiedCustomerRows(
  customers: CustomerView[],
  runtime: RuntimeCredential[],
): UnifiedCustomerRow[] {
  const managedRuntime = new Set(customers.map((customer) => customer.runtime_credential_id));
  const rows: UnifiedCustomerRow[] = customers.map((customer) => ({
    kind: "customer",
    id: customer.id,
    username: customer.username,
    runtimeCredentialID: customer.runtime_credential_id,
    customer,
  }));
  for (const credential of runtime) {
    if (credential.status === "revoked" || managedRuntime.has(credential.id)) continue;
    rows.push({
      kind: "runtime",
      id: `runtime:${credential.id}`,
      username: credential.username,
      runtimeCredentialID: credential.id,
      credential,
    });
  }
  return rows;
}
