#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
temp_root="$(mktemp -d)"
trap 'rm -rf -- "${temp_root}"' EXIT

env_file="${temp_root}/db.env"
cat >"${env_file}" <<'EOF'
PVNAIVE_DB_HOST=127.0.0.1
PVNAIVE_DB_PORT=5432
PVNAIVE_DB_NAME=pvnaive
PVNAIVE_DB_USER=pvnaive_app
PVNAIVE_DB_CONNECT_TIMEOUT=5
PVNAIVE_EXPECTED_SCHEMA_VERSION=1
PGPASSFILE=/etc/pvnaive/db.pgpass
EOF
chmod 0640 "${env_file}"

PVNAIVE_DB_ENV_FILE="${env_file}" bash "${repo_root}/scripts/db/set-expected-schema-version.sh" 2 >/dev/null
grep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=2' "${env_file}"
[[ "$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' "${env_file}")" == "1" ]]

PVNAIVE_DB_ENV_FILE="${env_file}" bash "${repo_root}/scripts/db/set-expected-schema-version.sh" 3 >/dev/null
grep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=3' "${env_file}"
[[ "$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' "${env_file}")" == "1" ]]

PVNAIVE_DB_ENV_FILE="${env_file}" bash "${repo_root}/scripts/db/set-expected-schema-version.sh" 7 >/dev/null
grep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=7' "${env_file}"
[[ "$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' "${env_file}")" == "1" ]]

if PVNAIVE_DB_ENV_FILE="${env_file}" bash "${repo_root}/scripts/db/set-expected-schema-version.sh" 8 >/dev/null 2>&1; then
  echo 'ERROR: unsupported schema version was accepted' >&2
  exit 1
fi

grep -Fqx 'PVNAIVE_EXPECTED_SCHEMA_VERSION=7' "${env_file}"
echo 'S04_DB_ENV_VERSION_TEST=PASSED'
