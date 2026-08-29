#!/usr/bin/env bash
set -Eeuo pipefail

preflight="scripts/stages/S06-owner-preflight.sh"
upgrade="scripts/stages/S06-owner-upgrade.sh"
[[ -f "${preflight}" && -f "${upgrade}" ]] || { echo 'ERROR: S06 stage scripts missing' >&2; exit 1; }
bash -n "${preflight}" "${upgrade}"

for token in 'S06_BASE_SCHEMA=7' 'S06_TARGET_SCHEMA=8' 'CADDYFILE_SHA256=' 'PREFLIGHT_RESULT=PASS' 'DB_SCHEMA_BASELINE' 'RUNTIME_AGENT_HEALTH'; do
  grep -Fq -- "${token}" "${preflight}" || { echo "ERROR: S06 preflight missing ${token}" >&2; exit 1; }
done
for token in \
  'PVNAIVE_EXPECTED_CADDY_SHA256' \
  '"stage": "S06-OWNER-CUSTOMER-OPS"' \
  '"base_schema_version": 7' \
  '"schema_version": 8' \
  '0008_subscription_profile_projection.up.sql' \
  'scripts/db/backup.sh' \
  'scripts/db/migrate.sh' \
  'scripts/db/rollback.sh' \
  'PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=7' \
  'PVNAIVE_DB_RELEASE_SCHEMA_VERSION=8' \
  'PVNAIVE_DB_RELEASE_MIGRATION_FILE=0008_subscription_profile_projection.up.sql' \
  'PVNAIVE_EXPECTED_SCHEMA_VERSION=8' \
  'systemctl restart pvnaive-runtime-agent.service' \
  'systemctl restart pvnaive-api.service' \
  'ROLLBACK_CADDY_INVARIANTS=PASS' \
  'S06_RESULT=PASSED' \
  'USAGE_ACCOUNTING=NOT_PROVEN'; do
  grep -Fq -- "${token}" "${upgrade}" || { echo "ERROR: S06 upgrade missing ${token}" >&2; exit 1; }
done

backup_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/backup.sh"' "${upgrade}" | head -n1 | cut -d: -f1)"
migrate_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/migrate.sh"' "${upgrade}" | tail -n1 | cut -d: -f1)"
[[ -n "${backup_line}" && -n "${migrate_line}" && "${backup_line}" -lt "${migrate_line}" ]] || { echo 'ERROR: S06 DB backup must precede migration' >&2; exit 1; }

if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${upgrade}"; then
  echo 'ERROR: S06 owner release must not restart/reload Caddy' >&2
  exit 1
fi
if grep -Eq '(^|[[:space:]])(cp|mv|install|sed|perl)[^\n]*/etc/caddy/Caddyfile' "${upgrade}"; then
  echo 'ERROR: S06 owner release must not rewrite Caddyfile' >&2
  exit 1
fi
if grep -Fq '/var/www/naive' "${upgrade}"; then
  echo 'ERROR: S06 owner release must not mutate public root' >&2
  exit 1
fi

echo 'S06_OWNER_UPGRADE_CONTRACT=PASSED'
