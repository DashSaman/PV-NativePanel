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
    'LISTEN 0 4096 127.0.0.1:5432 0.0.0.0:*' \
    'LISTEN 0 4096 [::1]:5432 [::]:*' \
    'LISTEN 0 4096 127.0.0.1:122 0.0.0.0:*'
)"

for expected_port in 22 80 443; do
  printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening "${expected_port}"
done

if printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening 8443; then
  echo 'ERROR: absent listener port was accepted' >&2
  exit 1
fi

only_port_122='LISTEN 0 4096 127.0.0.1:122 0.0.0.0:*'
if printf '%s\n' "${only_port_122}" | pvnaive_tcp_port_is_listening 22; then
  echo 'ERROR: listener port 122 was accepted as port 22' >&2
  exit 1
fi

for invalid_port in invalid 0 65536; do
  if printf '%s\n' "${listeners}" | pvnaive_tcp_port_is_listening "${invalid_port}"; then
    echo "ERROR: invalid listener port was accepted: ${invalid_port}" >&2
    exit 1
  else
    invalid_status="$?"
  fi
  [[ "${invalid_status}" == "2" ]] || {
    echo "ERROR: invalid listener port ${invalid_port} returned ${invalid_status}, expected 2" >&2
    exit 1
  }
done

if printf '%s\n' "${listeners}" | pvnaive_tcp_has_non_loopback_postgres_listener; then
  echo 'ERROR: loopback-only PostgreSQL listeners were rejected' >&2
  exit 1
fi

external_postgres_listener='LISTEN 0 4096 0.0.0.0:5432 0.0.0.0:*'
printf '%s\n' "${external_postgres_listener}" | pvnaive_tcp_has_non_loopback_postgres_listener

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

set +e
snapshot_failure_output="$(
  bash -c 'set -Eeuo pipefail; source "$1"; ss(){ return 42; }; trap '\''echo SNAPSHOT_ERR_TRAP=CALLED'\'' ERR; snapshot="$(pvnaive_tcp_listener_snapshot)"' \
    _ "${repo_root}/scripts/stages/lib.sh" 2>&1
)"
snapshot_failure_status="$?"
set -e
[[ "${snapshot_failure_status}" == "42" ]] || {
  echo "ERROR: listener snapshot failure returned ${snapshot_failure_status}, expected 42" >&2
  exit 1
}
grep -qx 'SNAPSHOT_ERR_TRAP=CALLED' <<< "${snapshot_failure_output}"

rollback_counter="$(mktemp)"
trap 'rm -f -- "${rollback_counter}"' EXIT
set +e
bash -c '
  set -Eeuo pipefail
  source "$1"
  counter_file="$2"
  root_bashpid="${BASHPID}"
  printf "ROOT=%s\n" "${root_bashpid}" > "${counter_file}"
  rollback_once() {
    local code="$1"
    if ! pvnaive_is_root_bash_process "${root_bashpid}"; then
      trap - ERR
      exit "${code}"
    fi
    trap - ERR
    printf "ROLLBACK=%s\n" "${BASHPID}" >> "${counter_file}"
    exit "${code}"
  }
  trap '\''rollback_once "$?"'\'' ERR
  captured="$(false)"
' _ "${repo_root}/scripts/stages/lib.sh" "${rollback_counter}" >/dev/null 2>&1
rollback_status="$?"
set -e
[[ "${rollback_status}" == "1" ]] || {
  echo "ERROR: guarded rollback returned ${rollback_status}, expected 1" >&2
  exit 1
}
root_pid="$(sed -n 's/^ROOT=//p' "${rollback_counter}")"
mapfile -t rollback_pids < <(sed -n 's/^ROLLBACK=//p' "${rollback_counter}")
[[ "${#rollback_pids[@]}" == "1" && "${rollback_pids[0]}" == "${root_pid}" ]] || {
  echo "ERROR: rollback was not executed exactly once in the root Bash process" >&2
  cat "${rollback_counter}" >&2
  exit 1
}
rm -f -- "${rollback_counter}"
trap - EXIT

echo 'S03_PREFLIGHT_TEST=PASSED'
