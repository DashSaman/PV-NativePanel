#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

max_version="${1:-}"
test_script="${2:-}"
[[ "${max_version}" =~ ^[1-9][0-9]*$ ]] || { echo 'ERROR: fixture max version must be numeric' >&2; exit 1; }
[[ -n "${test_script}" && -f "${test_script}" ]] || { echo 'ERROR: fixture test script is missing' >&2; exit 1; }

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
tmp="$(mktemp -d)"
trap 'rm -rf -- "${tmp}"' EXIT HUP INT TERM
migrations="${tmp}/migrations"
mkdir -p "${migrations}"

for ((version=1; version<=max_version; version++)); do
  prefix="$(printf '%04d' "${version}")"
  mapfile -t matches < <(find "${repo_root}/db/migrations" -maxdepth 1 -type f -name "${prefix}_*.sql" -print | LC_ALL=C sort)
  ((${#matches[@]} == 2)) || { echo "ERROR: expected up/down pair for fixture migration ${prefix}" >&2; exit 1; }
  cp -- "${matches[@]}" "${migrations}/"
done
(
  cd "${migrations}"
  sha256sum *.sql > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)

export PVNAIVE_MIGRATIONS_DIR="${migrations}"
echo "SCHEMA_FIXTURE_MAX_VERSION=${max_version}"
exec bash "${test_script}"
