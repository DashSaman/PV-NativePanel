#!/usr/bin/env bash
set -Eeuo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
# shellcheck source=scripts/stages/lib.sh
source "${repo_root}/scripts/stages/lib.sh"

listeners="$(
  printf '%s\n' \
    'LISTEN 0 4096 0.0.0.0:22 0.0.0.0:*' \
    'LISTEN 0 4096 127.0.0.1:80 0.0.0.0:*' \
    'LISTEN 0 4096 [::]:443 [::]:*' \
    'LISTEN 0 4096 127.0.0.1:122 0.0.0.0:*'
)"

for expected_port in 22 80 443; do
  printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening "${expected_port}"
done

if printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening 5432; then
  echo 'ERROR: absent listener port was accepted' >&2
  exit 1
fi

if printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening invalid; then
  echo 'ERROR: invalid listener port was accepted' >&2
  exit 1
else
  invalid_status="$?"
fi
[[ "${invalid_status}" == "2" ]] || {
  echo "ERROR: invalid listener port returned ${invalid_status}, expected 2" >&2
  exit 1
}

set +e
failure_output="$(
  bash -c 'set -Eeuo pipefail; source "$1"; trap '\''echo ERR_TRAP=CALLED'\'' ERR; false || die "regression failure"' \
    _ "${repo_root}/scripts/stages/lib.sh" 2>&1
)"
failure_status="$?"
set -e

[[ "${failure_status}" == "1" ]] || {
  echo "ERROR: die returned ${failure_status}, expected 1" >&2
  exit 1
}
grep -qx 'ERROR: regression failure' <<< "${failure_output}"
grep -qx 'ERR_TRAP=CALLED' <<< "${failure_output}"

echo 'S03_PREFLIGHT_TEST=PASSED'
