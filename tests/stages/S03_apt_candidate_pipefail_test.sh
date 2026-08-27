#!/usr/bin/env bash
set -Eeuo pipefail

producer() {
  printf '%s\n' \
    'postgresql:' \
    '  Installed: (none)' \
    '  Candidate: 18+999'
  local i
  for ((i = 0; i < 20000; i++)); do
    printf '  %05d dummy-policy-line-to-force-the-producer-to-keep-writing-after-candidate\n' "$i"
  done
}

candidate="$(producer | awk '/Candidate:/ {candidate=$2} END {if (candidate != "") print candidate}')"
[[ "${candidate}" == '18+999' ]]

stage_file="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)/scripts/stages/S03-database.sh"
if [[ -f "${stage_file}" ]]; then
  ! grep -Fq "awk '/Candidate:/ {print \$2; exit}'" "${stage_file}"
fi

printf '%s\n' 'S03_APT_CANDIDATE_PIPEFAIL_TEST=PASSED'
