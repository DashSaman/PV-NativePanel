#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
api_unit="${root}/ops/systemd/pvnaive-api.service"
caddy_dropin="${root}/ops/systemd/caddy-naive.service.d/20-pvnaive-accounting.conf"
foundation="${root}/scripts/stages/S02-foundation.sh"
deploy="${root}/scripts/release/deploy-r1.sh"
patch3="${root}/third_party/forwardproxy/patches/0003-pvnaive-session-control-lifecycle.patch"

grep -Eq '^SupplementaryGroups=.*pvnaive-session-control([[:space:]]|$)' "${api_unit}" || {
  echo 'ERROR: API service lacks dedicated session-control supplementary group' >&2; exit 1;
}
grep -Eq '^SupplementaryGroups=.*pvnaive-telemetry.*pvnaive-session-control|^SupplementaryGroups=.*pvnaive-session-control.*pvnaive-telemetry' "${caddy_dropin}" || {
  echo 'ERROR: Caddy must retain telemetry group and add dedicated session-control group' >&2; exit 1;
}
grep -Fq 'getent group pvnaive-session-control' "${foundation}" && grep -Fq 'groupadd --system pvnaive-session-control' "${foundation}" || {
  echo 'ERROR: fresh foundation must provision dedicated session-control group' >&2; exit 1;
}
grep -Fq 'getent group pvnaive-session-control' "${deploy}" && grep -Fq 'groupadd --system pvnaive-session-control' "${deploy}" || {
  echo 'ERROR: same-schema release deploy must provision dedicated session-control group before service activation' >&2; exit 1;
}
grep -Fq 'pvnaive-session-control' "${patch3}" || {
  echo 'ERROR: forwardproxy session-control socket must assign the dedicated group' >&2; exit 1;
}
echo 'TASK13_SESSION_CONTROL_PERMISSIONS=PASSED'
