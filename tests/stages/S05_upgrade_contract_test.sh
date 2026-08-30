#!/usr/bin/env bash
set -Eeuo pipefail

preflight="scripts/stages/S05-preflight.sh"
upgrade="scripts/stages/S05-upgrade.sh"
expected_schema_setter="scripts/db/set-expected-schema-version.sh"

[[ -f "${preflight}" ]] || { echo "ERROR: missing ${preflight}" >&2; exit 1; }
[[ -f "${upgrade}" ]] || { echo "ERROR: missing ${upgrade}" >&2; exit 1; }
[[ -f "${expected_schema_setter}" ]] || { echo "ERROR: missing ${expected_schema_setter}" >&2; exit 1; }
bash -n "${preflight}" "${upgrade}" "${expected_schema_setter}"

for required in \
  'PVNAIVE_NAIVE_PUBLIC_HOST' \
  'PREFLIGHT_RESULT=PASS' \
  'CADDYFILE_SHA256=' \
  'DB_SCHEMA_VERSION='; do
  grep -Fq -- "${required}" "${preflight}" || { echo "ERROR: S05 preflight missing contract token: ${required}" >&2; exit 1; }
done

for required in \
  'PVNAIVE_EXPECTED_CADDY_SHA256' \
  'PVNAIVE_NAIVE_PUBLIC_HOST' \
  '/etc/pvnaive/api.env' \
  'caddy_bin="/usr/local/bin/caddy"' \
  'validate --config' \
  'systemctl show caddy-naive.service' \
  'MainPID' \
  'NRestarts' \
  'scripts/db/backup.sh' \
  'scripts/db/migrate.sh' \
  '/etc/pvnaive/runtime.key' \
  'pvnaive-runtime-agent.service' \
  'pvnaive-api.service' \
  'runtime-agent.sock' \
  '0006_direct_subscription_tokens.up.sql' \
  'PVNAIVE_DB_RELEASE_SCHEMA_VERSION=6' \
  'SCHEMA_VERSION=6' \
  'ROLLBACK_DATABASE=PASS' \
  'ROLLBACK_DATABASE=FAIL' \
  'ROLLBACK_API_ENV=PASS' \
  'ROLLBACK_MIGRATION_CHAIN' \
  'PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA' \
  'S05_RESULT=PASSED'; do
  grep -Fq -- "${required}" "${upgrade}" || { echo "ERROR: S05 upgrade missing contract token: ${required}" >&2; exit 1; }
done

if grep -Eq 'systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy-naive\.service' "${upgrade}"; then
  echo 'ERROR: S05 upgrade must never restart or reload Caddy' >&2
  exit 1
fi
if grep -Eq '(cat|xxd|base64|od)[[:space:]].*/etc/pvnaive/runtime\.key' "${upgrade}"; then
  echo 'ERROR: S05 upgrade may not print runtime key material' >&2
  exit 1
fi

backup_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/backup.sh"' "${upgrade}" | head -n1 | cut -d: -f1)"
migrate_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/migrate.sh"' "${upgrade}" | tail -n1 | cut -d: -f1)"
[[ -n "${backup_line}" && -n "${migrate_line}" && "${backup_line}" -lt "${migrate_line}" ]] || {
  echo 'ERROR: encrypted DB backup must precede migration' >&2
  exit 1
}

env_write_line="$(grep -n -F 'PVNAIVE_NAIVE_PUBLIC_HOST=${naive_public_host}' "${upgrade}" | head -n1 | cut -d: -f1)"
api_restart_line="$(grep -n -F 'systemctl restart pvnaive-api.service' "${upgrade}" | head -n1 | cut -d: -f1)"
[[ -n "${env_write_line}" && -n "${api_restart_line}" && "${env_write_line}" -lt "${api_restart_line}" ]] || {
  echo 'ERROR: public Naive host must be persisted before API restart' >&2
  exit 1
}

grep -Fq 'systemctl restart pvnaive-runtime-agent.service' "${upgrade}" || {
  echo 'ERROR: S05 upgrade must restart runtime agent after binary install' >&2
  exit 1
}

rollback_loop_line="$(grep -n -F 'while [[ "${current_schema}"' "${upgrade}" | head -n1 | cut -d: -f1)"
rollback_call_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/rollback.sh"' "${upgrade}" | head -n1 | cut -d: -f1)"
[[ -n "${rollback_loop_line}" && -n "${rollback_call_line}" && "${rollback_loop_line}" -lt "${rollback_call_line}" ]] || {
  echo 'ERROR: S05 rollback must loop across every migration newer than the pre-upgrade schema' >&2
  exit 1
}

# Regression proof for the production failure observed during the guarded S05
# upgrade: schema 6 must be a valid expected schema, while unknown future
# schema versions remain rejected. The shared setter may know newer released
# schemas; S05 therefore tests the first unsupported version beyond the current released schema.
setter_tmp="$(mktemp -d)"
cleanup_setter_tmp() {
  rm -rf -- "${setter_tmp}"
}
trap cleanup_setter_tmp EXIT HUP INT TERM
setter_env="${setter_tmp}/db.env"
printf '%s\n' \
  'PVNAIVE_EXPECTED_SCHEMA_VERSION=3' \
  'PVNAIVE_DB_NAME=pvnaive' >"${setter_env}"
chmod 0640 "${setter_env}"
PVNAIVE_DB_ENV_FILE="${setter_env}" bash "${expected_schema_setter}" 6 >/dev/null
grep -Fxq 'PVNAIVE_EXPECTED_SCHEMA_VERSION=6' "${setter_env}" || {
  echo 'ERROR: expected schema setter did not persist schema version 6' >&2
  exit 1
}
latest_schema="$(find db/migrations -maxdepth 1 -type f -name '????_*.up.sql' -printf '%f\n' | cut -c1-4 | sort -n | tail -n1)"
[[ "${latest_schema}" =~ ^[0-9]{4}$ ]] || { echo 'ERROR: could not resolve latest schema for setter contract' >&2; exit 1; }
unsupported_schema=$((10#${latest_schema} + 1))
if PVNAIVE_DB_ENV_FILE="${setter_env}" bash "${expected_schema_setter}" "${unsupported_schema}" >/dev/null 2>&1; then
  echo "ERROR: expected schema setter accepted unsupported schema version ${unsupported_schema}" >&2
  exit 1
fi
trap - EXIT HUP INT TERM
cleanup_setter_tmp

# Real PostgreSQL 18 integration proof for the frozen S05 chain safety gate.
bash tests/db/rollback_chain_test.sh

echo 'S05_UPGRADE_CONTRACT=PASSED'
