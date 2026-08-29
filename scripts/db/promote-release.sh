#!/usr/bin/env bash
# Install and atomically select an immutable PVNaive database tooling release.
set -Eeuo pipefail
umask 077

pvnaive_release_die() {
  echo "ERROR: $*" >&2
  exit 1
}

for command_name in awk sha256sum realpath install find diff readlink ln mv id getent; do
  command -v "${command_name}" >/dev/null 2>&1 || pvnaive_release_die "required command not found: ${command_name}"
done

source_root="${PVNAIVE_DB_RELEASE_SOURCE_ROOT:-}"
release_root="${PVNAIVE_DB_RELEASE_ROOT:-/opt/pvnaive/db/releases}"
current_link="${PVNAIVE_DB_CURRENT_LINK:-/opt/pvnaive/db/current}"
schema_version="${PVNAIVE_DB_RELEASE_SCHEMA_VERSION:-}"
migration_file="${PVNAIVE_DB_RELEASE_MIGRATION_FILE:-}"
owner_user="${PVNAIVE_DB_RELEASE_OWNER_USER:-root}"
owner_group="${PVNAIVE_DB_RELEASE_OWNER_GROUP:-pvnaive}"

[[ -n "${source_root}" ]] || pvnaive_release_die "PVNAIVE_DB_RELEASE_SOURCE_ROOT is required"
[[ "${source_root}" == /* ]] || pvnaive_release_die "release source root must be absolute"
[[ "${release_root}" == /* ]] || pvnaive_release_die "release root must be absolute"
[[ "${current_link}" == /* ]] || pvnaive_release_die "current link must be absolute"
[[ "${schema_version}" =~ ^[1-9][0-9]*$ ]] || pvnaive_release_die "invalid schema version"
[[ "${migration_file}" =~ ^([0-9]{4})_[a-z0-9_]+\.up\.sql$ ]] || pvnaive_release_die "invalid migration filename"
version_prefix="${BASH_REMATCH[1]}"
((10#${version_prefix} == 10#${schema_version})) || pvnaive_release_die "migration version does not match requested schema version"

canonical_source="$(realpath --canonicalize-existing -- "${source_root}")"
canonical_release_root="$(realpath --canonicalize-missing -- "${release_root}")"
canonical_current_parent="$(realpath --canonicalize-missing -- "$(dirname -- "${current_link}")")"
[[ "${canonical_release_root}" == "${canonical_current_parent}/releases" ]] || \
  pvnaive_release_die "release root/current link layout mismatch"

id "${owner_user}" >/dev/null 2>&1 || pvnaive_release_die "unknown release owner user: ${owner_user}"
getent group "${owner_group}" >/dev/null 2>&1 || pvnaive_release_die "unknown release owner group: ${owner_group}"

migration_dir="${canonical_source}/db/migrations"
scripts_dir="${canonical_source}/scripts/db"
manifest="${migration_dir}/SHA256SUMS"
[[ -d "${migration_dir}" && -d "${scripts_dir}" ]] || pvnaive_release_die "source DB release directories are missing"
[[ -f "${manifest}" ]] || pvnaive_release_die "migration SHA256SUMS is missing"
[[ -f "${migration_dir}/${migration_file}" ]] || pvnaive_release_die "requested migration is missing"
[[ -f "${scripts_dir}/health.sh" ]] || pvnaive_release_die "health.sh is missing"

if [[ -n "$(find "${migration_dir}" "${scripts_dir}" -type l -print -quit)" ]]; then
  pvnaive_release_die "release source contains symlinks"
fi

(
  cd "${migration_dir}"
  sha256sum --check --strict SHA256SUMS >/dev/null
) || pvnaive_release_die "migration checksum manifest verification failed"

manifest_sha="$(awk -v file="${migration_file}" '$2 == file {print $1}' "${manifest}")"
[[ "${manifest_sha}" =~ ^[0-9a-f]{64}$ ]] || pvnaive_release_die "migration checksum missing or invalid in manifest"
actual_sha="$(sha256sum "${migration_dir}/${migration_file}" | awk '{print $1}')"
[[ "${actual_sha}" == "${manifest_sha}" ]] || pvnaive_release_die "migration checksum differs from manifest"

release_id="${version_prefix}-${actual_sha:0:12}"
release_dir="${canonical_release_root}/${release_id}"
install -d -o "${owner_user}" -g "${owner_group}" -m 0750 "${canonical_release_root}"

temp_dir=""
cleanup() {
  local code="$1"
  trap - EXIT HUP INT TERM
  if [[ -n "${temp_dir}" && -d "${temp_dir}" ]]; then
    rm -rf -- "${temp_dir}" || true
  fi
  rm -f -- "${current_link}.new.${BASHPID}" 2>/dev/null || true
  exit "${code}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

if [[ -e "${release_dir}" ]]; then
  [[ -d "${release_dir}" ]] || pvnaive_release_die "immutable release path is not a directory"
  diff -qr "${migration_dir}" "${release_dir}/db/migrations" >/dev/null || \
    pvnaive_release_die "existing immutable release migration content mismatch"
  diff -qr "${scripts_dir}" "${release_dir}/scripts/db" >/dev/null || \
    pvnaive_release_die "existing immutable release script content mismatch"
else
  temp_dir="${canonical_release_root}/.${release_id}.tmp.${BASHPID}"
  [[ ! -e "${temp_dir}" ]] || pvnaive_release_die "temporary release path already exists"
  install -d -o "${owner_user}" -g "${owner_group}" -m 0750 \
    "${temp_dir}" "${temp_dir}/db" "${temp_dir}/db/migrations" \
    "${temp_dir}/scripts" "${temp_dir}/scripts/db"

  while IFS= read -r -d '' file_path; do
    install -o "${owner_user}" -g "${owner_group}" -m 0640 \
      "${file_path}" "${temp_dir}/db/migrations/$(basename -- "${file_path}")"
  done < <(find "${migration_dir}" -maxdepth 1 -type f -print0 | LC_ALL=C sort -z)

  while IFS= read -r -d '' file_path; do
    install -o "${owner_user}" -g "${owner_group}" -m 0750 \
      "${file_path}" "${temp_dir}/scripts/db/$(basename -- "${file_path}")"
  done < <(find "${scripts_dir}" -maxdepth 1 -type f -name '*.sh' -print0 | LC_ALL=C sort -z)

  diff -qr "${migration_dir}" "${temp_dir}/db/migrations" >/dev/null || \
    pvnaive_release_die "staged immutable release migration content mismatch"
  diff -qr "${scripts_dir}" "${temp_dir}/scripts/db" >/dev/null || \
    pvnaive_release_die "staged immutable release script content mismatch"

  mv -- "${temp_dir}" "${release_dir}"
  temp_dir=""
fi

old_target=""
if [[ -e "${current_link}" || -L "${current_link}" ]]; then
  [[ -L "${current_link}" ]] || pvnaive_release_die "current DB release path is not a symlink"
  old_target="$(readlink -f -- "${current_link}")"
fi

new_link="${current_link}.new.${BASHPID}"
ln -s "${release_dir}" "${new_link}"
mv -Tf -- "${new_link}" "${current_link}"
[[ "$(readlink -f -- "${current_link}")" == "${release_dir}" ]] || pvnaive_release_die "atomic current-link promotion failed"

trap - EXIT HUP INT TERM
echo "PVNAIVE_DB_RELEASE_PROMOTION=PASSED"
echo "PVNAIVE_DB_RELEASE_ID=${release_id}"
echo "PVNAIVE_DB_RELEASE_OLD=${old_target:-none}"
echo "PVNAIVE_DB_RELEASE_NEW=${release_dir}"
