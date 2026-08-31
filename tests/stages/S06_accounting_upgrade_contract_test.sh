#!/usr/bin/env bash
set -Eeuo pipefail

preflight="scripts/stages/S06-accounting-preflight.sh"
upgrade="scripts/stages/S06-accounting-upgrade.sh"

[[ -f "${preflight}" ]] || { echo 'ERROR: S06 accounting preflight missing' >&2; exit 1; }
[[ -f "${upgrade}" ]] || { echo 'ERROR: S06 accounting upgrade missing' >&2; exit 1; }
bash -n "${preflight}" "${upgrade}"

if grep -Eq 'S06_ACCOUNTING_(BASE|TARGET)_SCHEMA=(7|8)|base_schema_version[^0-9]*(7|8)|schema_version[^0-9]*(7|8)|0007_exact_accounting' "${preflight}" "${upgrade}"; then
  echo 'ERROR: stale schema7/8 accounting release constants detected' >&2
  exit 1
fi

for token in \
  'S06_ACCOUNTING_BASE_SCHEMA=16' \
  'S06_ACCOUNTING_TARGET_SCHEMA=17' \
  'CADDYFILE_SHA256=' \
  'ACCOUNTING_CADDY_SHA256=' \
  'ACCOUNTING_CADDY_MODULE=' \
  'PREFLIGHT_RESULT=PASS' \
  'RUNTIME_AGENT_HEALTH'; do
  grep -Fq -- "${token}" "${preflight}" || { echo "ERROR: S06 accounting preflight missing ${token}" >&2; exit 1; }
done

for token in \
  'PVNAIVE_ACCOUNTING_CADDY_SHA256' \
  '"stage": "S06-ACCOUNTING-RELEASE"' \
  '"base_schema_version": 16' \
  '"schema_version": 17' \
  '0017_operator_session_peers.up.sql' \
  'scripts/db/backup.sh' \
  'scripts/db/migrate.sh' \
  'scripts/db/rollback.sh' \
  'PVNAIVE_ROLLBACK_CHAIN_TARGET_SCHEMA=16' \
  'PVNAIVE_DB_RELEASE_SCHEMA_VERSION=17' \
  'PVNAIVE_DB_RELEASE_MIGRATION_FILE=0017_operator_session_peers.up.sql' \
  'PVNAIVE_EXPECTED_SCHEMA_VERSION=17' \
  'caddy-naive.service' \
  'ROLLBACK_CADDY_INVARIANTS=PASS' \
  'S06_ACCOUNTING_RESULT=PASSED' \
  'ONE_CONTROLLED_CADDY_RESTART=CONFIRMED' \
  'POSTFLIGHT_CADDY_RESTART_COUNT=1'; do
  grep -Fq -- "${token}" "${upgrade}" || { echo "ERROR: S06 accounting upgrade missing ${token}" >&2; exit 1; }
done

backup_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/backup.sh"' "${upgrade}" | head -n1 | cut -d: -f1)"
migrate_line="$(grep -n -F 'bash "${bundle_root}/scripts/db/migrate.sh"' "${upgrade}" | tail -n1 | cut -d: -f1)"
[[ -n "${backup_line}" && -n "${migrate_line}" && "${backup_line}" -lt "${migrate_line}" ]] || {
  echo 'ERROR: S06 accounting DB backup must precede migration' >&2
  exit 1
}

caddy_backup_line="$(grep -n -F 'caddy-before' "${upgrade}" | head -n1 | cut -d: -f1)"
[[ -n "${caddy_backup_line}" && -n "${backup_line}" && "${caddy_backup_line}" -lt "${migrate_line}" ]] || {
  echo 'ERROR: S06 accounting Caddy binary backup must precede migration' >&2
  exit 1
}

restart_lines="$(grep -c -E 'systemctl[[:space:]]+(restart)[[:space:]]+caddy-naive\.service' "${upgrade}" || true)"
[[ "${restart_lines}" -eq 1 ]] || {
  echo "ERROR: S06 accounting upgrade must perform exactly one Caddy restart, found ${restart_lines}" >&2
  exit 1
}

if grep -Eq '(^|[[:space:]])(cp|mv|install|sed|perl)[^\n]*/etc/caddy/Caddyfile' "${upgrade}"; then
  echo 'ERROR: S06 accounting upgrade must not rewrite Caddyfile' >&2
  exit 1
fi
if grep -Fq '/var/www/naive' "${upgrade}"; then
  echo 'ERROR: S06 accounting upgrade must not mutate public root' >&2
  exit 1
fi

if grep -Eq '(cat|xxd|base64|od)[[:space:]].*/etc/pvnaive/runtime\.key' "${upgrade}"; then
  echo 'ERROR: S06 accounting upgrade may not print runtime key material' >&2
  exit 1
fi

# PVNaive accounting is patched into http.handlers.forward_proxy; it is not a
# separately registered Caddy module. Release safety must prove the exact live
# Caddyfile accounting directive validates with the candidate/installed binary.
if grep -Fq "list-modules | grep -Fq 'pvnaive_accounting'" "${upgrade}"; then
  echo 'ERROR: S06 accounting upgrade must not require a nonexistent standalone pvnaive_accounting module' >&2
  exit 1
fi
grep -Fq "grep -Fq 'pvnaive_accounting_socket'" "${upgrade}" || { echo 'ERROR: S06 upgrade must require the live accounting directive' >&2; exit 1; }
grep -Fq "failed exact live accounting config validation" "${upgrade}" || { echo 'ERROR: S06 upgrade must functionally validate the exact live accounting Caddyfile' >&2; exit 1; }

echo 'S06_ACCOUNTING_UPGRADE_CONTRACT=PASSED'
# Rollback must restore every runtime artifact/symlink changed before postflight.
grep -Fq 'pvnaive-telemetry-agent.before' "${upgrade}" || { echo 'ERROR: telemetry agent backup/rollback missing' >&2; exit 1; }
grep -Fq 'web-current.before' "${upgrade}" || { echo 'ERROR: web current rollback marker missing' >&2; exit 1; }
grep -Fq 'db-current.before' "${upgrade}" || { echo 'ERROR: DB current rollback marker missing' >&2; exit 1; }
# After new Caddy has accepted traffic, schema17 peer evidence is immutable. Recovery
# must restore the old Caddy first and preserve schema17 rather than blindly down-migrate.
grep -Fq 'caddy_activated=0' "${upgrade}" || { echo 'ERROR: Caddy activation phase marker missing' >&2; exit 1; }
grep -Fq 'caddy_activated=1' "${upgrade}" || { echo 'ERROR: Caddy activation success marker missing' >&2; exit 1; }
grep -Fq 'trusted_peer_evidence' "${upgrade}" || { echo 'ERROR: post-activation trusted peer evidence guard missing' >&2; exit 1; }
grep -Fq 'ROLLBACK_DATABASE=PRESERVED_TRUSTED_PEER_EVIDENCE' "${upgrade}" || { echo 'ERROR: evidence-preserving schema17 recovery marker missing' >&2; exit 1; }
