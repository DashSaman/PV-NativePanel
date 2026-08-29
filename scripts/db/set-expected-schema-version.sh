#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

version="${1:-}"
env_file="${PVNAIVE_DB_ENV_FILE:-/etc/pvnaive/db.env}"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
migrations_dir="${PVNAIVE_MIGRATIONS_DIR:-${script_dir}/../../db/migrations}"

[[ "${version}" =~ ^[1-9][0-9]*$ ]] || {
  echo "ERROR: expected schema version must be a positive integer" >&2
  exit 1
}
[[ -d "${migrations_dir}" ]] || {
  echo "ERROR: migrations directory is missing: ${migrations_dir}" >&2
  exit 1
}
latest_file="$(find "${migrations_dir}" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]_*.up.sql' -printf '%f\n' | LC_ALL=C sort | tail -n1)"
[[ "${latest_file}" =~ ^([0-9]{4})_ ]] || {
  echo "ERROR: no valid migrations found in ${migrations_dir}" >&2
  exit 1
}
latest_version=$((10#${BASH_REMATCH[1]}))
(( version >= 1 && version <= latest_version )) || {
  echo "ERROR: expected schema version must be between 1 and ${latest_version}" >&2
  exit 1
}

[[ -f "${env_file}" ]] || {
  echo "ERROR: database environment file is missing: ${env_file}" >&2
  exit 1
}

count="$(grep -c '^PVNAIVE_EXPECTED_SCHEMA_VERSION=' "${env_file}" || true)"
[[ "${count}" == "1" ]] || {
  echo "ERROR: database environment must contain exactly one schema expectation" >&2
  exit 1
}

tmp="$(mktemp "${env_file}.tmp.XXXXXX")"
cleanup() {
  rm -f -- "${tmp}" >/dev/null 2>&1 || true
}
trap cleanup EXIT HUP INT TERM

awk -v version="${version}" '
  BEGIN { changed=0 }
  /^PVNAIVE_EXPECTED_SCHEMA_VERSION=/ {
    print "PVNAIVE_EXPECTED_SCHEMA_VERSION=" version
    changed++
    next
  }
  { print }
  END { if (changed != 1) exit 1 }
' "${env_file}" > "${tmp}"

chown --reference="${env_file}" "${tmp}"
chmod --reference="${env_file}" "${tmp}"
mv -f -- "${tmp}" "${env_file}"
trap - EXIT HUP INT TERM

echo "PVNAIVE_DB_EXPECTED_SCHEMA_VERSION=${version}"
