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


# Encrypted archive validation must materialize a private temporary file instead
# of piping age into pg_restore --list. pg_restore may stop reading early after
# the TOC, which gives age SIGPIPE (141) under pipefail on production hosts.
archive_test_root="$(mktemp -d)"
cleanup_archive_test() { rm -rf -- "${archive_test_root}"; }
trap cleanup_archive_test EXIT
mkdir -p "${archive_test_root}/bin"
printf 'archive-bytes\n' > "${archive_test_root}/archive.age"
printf 'AGE-SECRET-KEY-test\n' > "${archive_test_root}/identity"
cat > "${archive_test_root}/bin/age" <<'EOF'
#!/usr/bin/env bash
set -Eeuo pipefail
output=''
input=''
while (($#)); do
  case "$1" in
    --decrypt) shift ;;
    --identity) shift 2 ;;
    --output) output="$2"; shift 2 ;;
    *) input="$1"; shift ;;
  esac
done
[[ -n "${output}" && -n "${input}" ]] || { echo 'fake age requires --output and input file' >&2; exit 91; }
cp -- "${input}" "${output}"
EOF
cat > "${archive_test_root}/bin/pg_restore" <<'EOF'
#!/usr/bin/env bash
set -Eeuo pipefail
[[ "$1" == '--list' && $# -eq 2 && -f "$2" ]] || { echo 'fake pg_restore requires --list FILE' >&2; exit 92; }
grep -Fqx 'archive-bytes' "$2"
EOF
chmod 0755 "${archive_test_root}/bin/age" "${archive_test_root}/bin/pg_restore"
TMPDIR="${archive_test_root}" PATH="${archive_test_root}/bin:${PATH}" \
  pvnaive_validate_encrypted_archive "${archive_test_root}/archive.age" "${archive_test_root}/identity"
[[ -z "$(find "${archive_test_root}" -maxdepth 1 -type f -name 'pvnaive-archive-validate.*' -print -quit)" ]] || {
  echo 'ERROR: archive validation left plaintext temp material behind' >&2
  exit 1
}
cleanup_archive_test
trap - EXIT

echo 'PVNAIVE_DB_LIB_TEST=PASSED'
