#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
# shellcheck source=scripts/db/lib.sh
source "${repo_root}/scripts/db/lib.sh"

PVNAIVE_DB_HOST=127.0.0.1
PVNAIVE_DB_NAME=pvnaive
PVNAIVE_DB_USER=pvnaive_app

for valid_port in 1 5432 65535; do
  PVNAIVE_DB_PORT="${valid_port}"
  pvnaive_db_defaults
done

for invalid_port in 0 65536 invalid; do
  set +e
  invalid_output="$(
    PVNAIVE_DB_PORT="${invalid_port}" bash -c '
      set -Eeuo pipefail
      source "$1"
      PVNAIVE_DB_HOST=127.0.0.1
      PVNAIVE_DB_NAME=pvnaive
      PVNAIVE_DB_USER=pvnaive_app
      pvnaive_db_defaults
    ' _ "${repo_root}/scripts/db/lib.sh" 2>&1
  )"
  invalid_status="$?"
  set -e
  [[ "${invalid_status}" == "1" ]] || {
    echo "ERROR: invalid database port ${invalid_port} returned ${invalid_status}" >&2
    exit 1
  }
  grep -qx 'ERROR: invalid database port' <<< "${invalid_output}"
done

pvnaive_validate_storage_root /var/backups/pvnaive
for invalid_root in / /etc relative/path /var/backups/../backups/pvnaive; do
  set +e
  storage_output="$(
    bash -c 'set -Eeuo pipefail; source "$1"; pvnaive_validate_storage_root "$2"' \
      _ "${repo_root}/scripts/db/lib.sh" "${invalid_root}" 2>&1
  )"
  storage_status="$?"
  set -e
  [[ "${storage_status}" == "1" ]] || {
    echo "ERROR: unsafe storage root ${invalid_root} returned ${storage_status}" >&2
    exit 1
  }
  grep -Eq '^ERROR: storage root (is too broad|must be an absolute canonical path without symlink traversal)$' <<< "${storage_output}"
done

echo 'PVNAIVE_DB_LIB_TEST=PASSED'
