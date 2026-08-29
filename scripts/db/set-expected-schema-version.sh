#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

version="${1:-}"
env_file="${PVNAIVE_DB_ENV_FILE:-/etc/pvnaive/db.env}"

case "${version}" in
  1|2|3|4|5|6) ;;
  *)
    echo "ERROR: expected schema version must be between 1 and 6" >&2
    exit 1
    ;;
esac

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
