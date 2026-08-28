#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
preflight="${repo_root}/scripts/stages/S04R-preflight.sh"
upgrade="${repo_root}/scripts/stages/S04R-upgrade.sh"

for target in "${preflight}" "${upgrade}"; do
  [[ -f "${target}" ]] || { echo "ERROR: missing ${target}" >&2; exit 1; }
  if grep -nE 'list-modules.*\|.*grep[[:space:]]+-[^[:space:]]*q' "${target}" >/dev/null; then
    echo "ERROR: ${target} uses grep -q on caddy list-modules under pipefail; this can false-negative when Caddy receives SIGPIPE" >&2
    exit 1
  fi
done

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "${tmpdir}"' EXIT HUP INT TERM
producer="${tmpdir}/module-producer"
cat >"${producer}" <<'PRODUCER'
#!/usr/bin/env bash
set -Eeuo pipefail
printf '%s\n' 'http.handlers.forward_proxy'
for ((i=0; i<200000; i++)); do
  printf 'http.handlers.synthetic_%06d\n' "${i}"
done
PRODUCER
chmod 0755 "${producer}"

set +e
set -o pipefail
"${producer}" | grep -Fxq 'http.handlers.forward_proxy'
unsafe_rc=$?
set +o pipefail
set -e
[[ "${unsafe_rc}" -ne 0 ]] || {
  echo 'ERROR: regression harness did not reproduce the grep -q/pipefail SIGPIPE hazard' >&2
  exit 1
}

set +e
set -o pipefail
"${producer}" | grep -Fx 'http.handlers.forward_proxy' >/dev/null
safe_rc=$?
set +o pipefail
set -e
[[ "${safe_rc}" -eq 0 ]] || {
  echo "ERROR: full-consumption exact-match module check failed with rc=${safe_rc}" >&2
  exit 1
}

echo 'S04R_MODULE_PIPEFAIL_REGRESSION=PASSED'
