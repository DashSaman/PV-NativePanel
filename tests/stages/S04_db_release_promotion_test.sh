#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
tmp_root="$(mktemp -d)"
cleanup() { rm -rf -- "${tmp_root}"; }
trap cleanup EXIT HUP INT TERM

mkdir -p "${tmp_root}/db/releases/0001-old/scripts/db"
printf '%s\n' '#!/usr/bin/env bash' 'echo OLD_S03_HEALTH' > "${tmp_root}/db/releases/0001-old/scripts/db/health.sh"
chmod 0750 "${tmp_root}/db/releases/0001-old/scripts/db/health.sh"
ln -s "${tmp_root}/db/releases/0001-old" "${tmp_root}/db/current"

out="$(
  PVNAIVE_DB_RELEASE_SOURCE_ROOT="${repo_root}" \
  PVNAIVE_DB_RELEASE_ROOT="${tmp_root}/db/releases" \
  PVNAIVE_DB_CURRENT_LINK="${tmp_root}/db/current" \
  PVNAIVE_DB_RELEASE_SCHEMA_VERSION=2 \
  PVNAIVE_DB_RELEASE_MIGRATION_FILE=0002_auth_foundation.up.sql \
  PVNAIVE_DB_RELEASE_OWNER_USER="$(id -un)" \
  PVNAIVE_DB_RELEASE_OWNER_GROUP="$(id -gn)" \
    bash "${repo_root}/scripts/db/promote-release.sh"
)"

printf '%s\n' "${out}"
new_target="$(readlink -f "${tmp_root}/db/current")"
expected_sha="$(sha256sum "${repo_root}/db/migrations/0002_auth_foundation.up.sql" | awk '{print $1}')"
expected_target="${tmp_root}/db/releases/0002-${expected_sha:0:12}"

[[ "${new_target}" == "${expected_target}" ]] || {
  echo "ERROR: current link target mismatch: ${new_target}" >&2
  exit 1
}
[[ -d "${tmp_root}/db/releases/0001-old" ]] || {
  echo 'ERROR: old immutable release was removed' >&2
  exit 1
}
cmp -s "${repo_root}/scripts/db/health.sh" "${new_target}/scripts/db/health.sh" || {
  echo 'ERROR: promoted health.sh differs from S04 source' >&2
  exit 1
}
grep -Fq 'actor_totp_factors' "${new_target}/scripts/db/health.sh" || {
  echo 'ERROR: promoted health release is not S04-aware' >&2
  exit 1
}
(
  cd "${new_target}/db/migrations"
  sha256sum --check --strict SHA256SUMS >/dev/null
)

# Promotion must also be idempotent when the immutable release already exists
# and still exactly matches the pinned source.
PVNAIVE_DB_RELEASE_SOURCE_ROOT="${repo_root}" \
PVNAIVE_DB_RELEASE_ROOT="${tmp_root}/db/releases" \
PVNAIVE_DB_CURRENT_LINK="${tmp_root}/db/current" \
PVNAIVE_DB_RELEASE_SCHEMA_VERSION=2 \
PVNAIVE_DB_RELEASE_MIGRATION_FILE=0002_auth_foundation.up.sql \
PVNAIVE_DB_RELEASE_OWNER_USER="$(id -un)" \
PVNAIVE_DB_RELEASE_OWNER_GROUP="$(id -gn)" \
  bash "${repo_root}/scripts/db/promote-release.sh" >/dev/null

[[ "$(readlink -f "${tmp_root}/db/current")" == "${expected_target}" ]] || {
  echo 'ERROR: idempotent promotion changed the release target' >&2
  exit 1
}

stage="${repo_root}/scripts/stages/S04-auth.sh"
grep -Fq 'scripts/db/promote-release.sh' "${stage}" || {
  echo 'ERROR: S04 stage does not require the DB release promotion helper' >&2
  exit 1
}
grep -Fq 'promote_db_tooling_release' "${stage}" || {
  echo 'ERROR: S04 stage does not wire DB release promotion' >&2
  exit 1
}
call_count="$(grep -Fc 'promote_db_tooling_release' "${stage}")"
((call_count >= 3)) || {
  echo "ERROR: S04 stage must define promotion and invoke it for existing-marker plus fresh/recovery paths; count=${call_count}" >&2
  exit 1
}

echo 'S04_DB_RELEASE_PROMOTION_TEST=PASSED'
