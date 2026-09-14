// R5 Pool Manager model (GATE STEER-006, UI slice R5-UI-001).
//
// Pure functions only: every rule here mirrors the schema-33 backend truth
// (internal/fleet/store.go + db/migrations/0033_pool_registry.up.sql) so the
// owner UI can never present a state the trusted boundary would refuse.
// The drain state machine is enforced server-side (active -> draining ->
// disabled, re-activation allowed from draining/disabled); the UI only
// offers transitions that the database accepts.

export type PoolHealth = "unknown" | "healthy" | "degraded" | "offline";
export type PoolMaintenance = "active" | "draining" | "disabled";
export type PoolDrift = "in_sync" | "pending" | "ahead";

export type PoolNodeRow = {
  id: string;
  display_name: string;
  region: string;
  capacity_weight: number;
  health: PoolHealth;
  maintenance: PoolMaintenance;
  desired_revision: number;
  applied_revision: number;
  drift: PoolDrift;
  last_seen_at: string | null;
};

const HEALTH_STATES: readonly PoolHealth[] = ["unknown", "healthy", "degraded", "offline"];
const MAINTENANCE_STATES: readonly PoolMaintenance[] = ["active", "draining", "disabled"];
const DRIFT_STATES: readonly PoolDrift[] = ["in_sync", "pending", "ahead"];

function oneOf<T extends string>(value: unknown, allowed: readonly T[], fallback: T): T {
  return typeof value === "string" && (allowed as readonly string[]).includes(value) ? (value as T) : fallback;
}

export type NormalizedPoolNodes = { nodes: PoolNodeRow[]; count: number };

/** Normalizes GET /api/v1/pool/nodes payloads; unknown states collapse to
 *  their safe defaults instead of rendering fabricated health. */
export function normalizePoolNodes(payload: unknown): NormalizedPoolNodes {
  const source = (payload ?? {}) as { nodes?: unknown; count?: unknown };
  const rawNodes = Array.isArray(source.nodes) ? source.nodes : [];
  const nodes: PoolNodeRow[] = [];
  for (const raw of rawNodes) {
    if (typeof raw !== "object" || raw === null) continue;
    const row = raw as Record<string, unknown>;
    const id = typeof row.id === "string" ? row.id : "";
    if (!id) continue;
    const desired = revisionNumber(row.desired_revision);
    const applied = revisionNumber(row.applied_revision);
    nodes.push({
      id,
      display_name: typeof row.display_name === "string" && row.display_name ? row.display_name : "گره بی‌نام",
      region: typeof row.region === "string" ? row.region : "",
      capacity_weight: typeof row.capacity_weight === "number" && Number.isFinite(row.capacity_weight) ? row.capacity_weight : 1,
      health: oneOf(row.health, HEALTH_STATES, "unknown"),
      maintenance: oneOf(row.maintenance, MAINTENANCE_STATES, "active"),
      desired_revision: desired,
      applied_revision: applied,
      drift: oneOf(row.drift, DRIFT_STATES, deriveDrift(desired, applied)),
      last_seen_at: typeof row.last_seen_at === "string" && row.last_seen_at ? row.last_seen_at : null,
    });
  }
  const declared = typeof source.count === "number" && Number.isFinite(source.count) ? source.count : nodes.length;
  return { nodes, count: typeof source.count === "number" ? declared : nodes.length };
}

function revisionNumber(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value) && value >= 0) return value;
  if (typeof value === "string" && value !== "" && Number.isFinite(Number(value))) return Number(value);
  return 0;
}

function deriveDrift(desired: number, applied: number): PoolDrift {
  if (applied === desired) return "in_sync";
  if (applied < desired) return "pending";
  return "ahead";
}

const DRIFT_LABELS: Record<PoolDrift, string> = {
  in_sync: "هم‌گام",
  pending: "در انتظار اعمال",
  ahead: "جلوتر از وضعیت مطلوب",
};

export function driftLabel(drift: PoolDrift): string {
  return DRIFT_LABELS[drift];
}

const HEALTH_LABELS: Record<PoolHealth, string> = {
  unknown: "نامشخص",
  healthy: "سالم",
  degraded: "کاهش‌یافته",
  offline: "خارج از دسترس",
};

export function healthLabel(health: PoolHealth): string {
  return HEALTH_LABELS[health];
}

const MAINTENANCE_LABELS: Record<PoolMaintenance, string> = {
  active: "فعال",
  draining: "در حال تخلیه",
  disabled: "غیرفعال",
};

export function maintenanceLabel(state: PoolMaintenance): string {
  return MAINTENANCE_LABELS[state];
}

/** Transition set offered per row — exactly what 0033 accepts:
 *  active -> draining; draining -> disabled | active; disabled -> active. */
export function maintenanceActions(state: PoolMaintenance): PoolMaintenance[] {
  switch (state) {
    case "active":
      return ["draining"];
    case "draining":
      return ["disabled", "active"];
    case "disabled":
      return ["active"];
  }
}

/** Formats an ISO timestamp as a coarse Persian relative age; returns an
 *  explicit never-seen label for null instead of inventing a time. */
export function formatLastSeen(iso: string | null, now: Date = new Date()): string {
  if (!iso) return "هرگز دیده نشده";
  const parsed = Date.parse(iso);
  if (!Number.isFinite(parsed)) return "نامشخص";
  const seconds = Math.max(0, Math.floor((now.getTime() - parsed) / 1000));
  if (seconds < 60) return "چند لحظه پیش";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes.toLocaleString("fa-IR")} دقیقه پیش`;
  const hours = Math.floor(minutes / 60);
  if (hours < 48) return `${hours.toLocaleString("fa-IR")} ساعت پیش`;
  const days = Math.floor(hours / 24);
  return `${days.toLocaleString("fa-IR")} روز پیش`;
}

export type EnrollTokenResult = { token: string; expires_at: string };

/** Extracts the one-time enrollment token response; anything else is a
 *  failure so the UI never shows a placeholder as a real token. */
export function normalizeEnrollToken(payload: unknown): EnrollTokenResult | null {
  const source = (payload ?? {}) as Record<string, unknown>;
  if (typeof source.token !== "string" || source.token.length < 16) return null;
  if (typeof source.expires_at !== "string" || !Number.isFinite(Date.parse(source.expires_at))) return null;
  return { token: source.token, expires_at: source.expires_at };
}

export type EnrollInput = { token: string; displayName: string; region: string; capacityWeight: number };

/** Client-side mirror of fleet.Validate* — the server stays authoritative;
 *  these checks only prevent obvious round-trips. */
export function validateEnrollInput(input: EnrollInput): string | null {
  const token = input.token.trim();
  if (token.length < 16) return "توکن ثبت‌نام نامعتبر است (حداقل ۱۶ نویسه).";
  const name = input.displayName.trim();
  if (name.length < 1 || name.length > 120) return "نام نمایشی گره باید ۱ تا ۱۲۰ نویسه باشد.";
  if (input.region.trim().length > 60) return "ناحیه حداکثر ۶۰ نویسه است.";
  if (!Number.isInteger(input.capacityWeight) || input.capacityWeight < 1 || input.capacityWeight > 10000) {
    return "وزن ظرفیت باید عدد صحیح ۱ تا ۱۰۰۰۰ باشد.";
  }
  return null;
}

export type EndpointInput = { host: string; port: number; sni: string };
export type RevisionInput = { poolId: string; endpoints: EndpointInput[]; weight: number; validityMinutes: number };

export const REVISION_LIMITS = {
  poolIdMax: 60,
  endpointsMin: 1,
  endpointsMax: 32,
  weightMin: 0.1,
  weightMax: 1000,
  validityMin: 60,
  validityMax: 20160,
} as const;

/** Mirrors httpapi.buildManifest validation (fail-closed input contract). */
export function validateRevisionInput(input: RevisionInput): string | null {
  const poolId = input.poolId.trim();
  if (poolId.length < 1 || poolId.length > REVISION_LIMITS.poolIdMax) {
    return `شناسه استخر باید ۱ تا ${REVISION_LIMITS.poolIdMax} نویسه باشد.`;
  }
  if (input.endpoints.length < REVISION_LIMITS.endpointsMin || input.endpoints.length > REVISION_LIMITS.endpointsMax) {
    return `حداقل یک نقطه اتصال و حداکثر ${REVISION_LIMITS.endpointsMax} نقطه مجاز است.`;
  }
  for (const endpoint of input.endpoints) {
    const host = endpoint.host.trim();
    if (host.length < 1 || host.length > 255) return "میزبان نقطه اتصال باید ۱ تا ۲۵۵ نویسه باشد.";
    if (!Number.isInteger(endpoint.port) || endpoint.port < 1 || endpoint.port > 65535) {
      return "پورت نقطه اتصال باید عدد صحیح ۱ تا ۶۵۵۳۵ باشد.";
    }
    if (endpoint.sni.trim().length > 255) return "SNI حداکثر ۲۵۵ نویسه است.";
  }
  if (!Number.isFinite(input.weight) || input.weight <= 0 || input.weight > REVISION_LIMITS.weightMax) {
    return `وزن باید بین ${REVISION_LIMITS.weightMin} تا ${REVISION_LIMITS.weightMax} باشد.`;
  }
  if (!Number.isInteger(input.validityMinutes) || input.validityMinutes < REVISION_LIMITS.validityMin || input.validityMinutes > REVISION_LIMITS.validityMax) {
    return `اعتبار مانیفست باید ${REVISION_LIMITS.validityMin} تا ${REVISION_LIMITS.validityMax} دقیقه باشد.`;
  }
  return null;
}

/** Serializes the validated form into the exact wire contract of
 *  POST /api/v1/pool/nodes/{id}/revisions. */
export function buildRevisionPayload(input: RevisionInput): Record<string, unknown> {
  return {
    pool_id: input.poolId.trim(),
    endpoints: input.endpoints.map((endpoint) => ({
      host: endpoint.host.trim(),
      port: endpoint.port,
      sni: endpoint.sni.trim(),
    })),
    weight: input.weight,
    validity_minutes: input.validityMinutes,
  };
}

export type NormalizedManifest = { revision: number; manifest: Record<string, unknown>; signature: string };

/** Normalizes GET /api/v1/pool/nodes/{id}/manifest; the signature is
 *  displayed read-only, never re-verified client-side (server verifies). */
export function normalizeManifest(payload: unknown): NormalizedManifest | null {
  const source = (payload ?? {}) as Record<string, unknown>;
  const revision = revisionNumber(source.revision);
  if (typeof source.manifest !== "object" || source.manifest === null) return null;
  if (typeof source.signature !== "string" || !source.signature) return null;
  return { revision, manifest: source.manifest as Record<string, unknown>, signature: source.signature };
}

/** Summarizes the stored manifest for the owner list (truthful labels for
 *  the fields the operator actually steers on). */
export function manifestSummary(manifest: Record<string, unknown>): string[] {
  const lines: string[] = [];
  const poolId = typeof manifest.pool_id === "string" ? manifest.pool_id : "";
  if (poolId) lines.push(`استخر: ${poolId}`);
  const weight = typeof manifest.weight === "number" ? manifest.weight : null;
  if (weight !== null) lines.push(`وزن: ${weight.toLocaleString("fa-IR")}`);
  if (Array.isArray(manifest.endpoints)) {
    for (const endpoint of manifest.endpoints) {
      if (typeof endpoint !== "object" || endpoint === null) continue;
      const item = endpoint as Record<string, unknown>;
      const host = typeof item.host === "string" ? item.host : "?";
      const port = typeof item.port === "number" ? item.port : "?";
      const sni = typeof item.sni === "string" && item.sni ? ` (SNI: ${item.sni})` : "";
      lines.push(`نقطه اتصال: ${host}:${port}${sni}`);
    }
  }
  const validUntil = typeof manifest.valid_until === "string" ? manifest.valid_until : "";
  if (validUntil && Number.isFinite(Date.parse(validUntil))) {
    lines.push(`اعتبار تا: ${new Date(validUntil).toLocaleString("fa-IR")}`);
  }
  return lines;
}
