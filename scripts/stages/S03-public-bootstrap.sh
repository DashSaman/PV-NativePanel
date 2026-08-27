#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S03-PUBLIC-BOOTSTRAP"
repo_url="https://github.com/DashSaman/PV-NativePanel.git"
source_commit="27899b63fa676029ec4f24a30f16933515f5fe21"
expected_host="testAmir5-3"
expected_caddy_sha256="101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
work="$(mktemp -d /root/pvnaive-s03-src.${stamp}.XXXXXX)"
log="/root/pvnaive-s03-${stamp}.log"

cleanup() {
  local rc="$1"
  trap - EXIT HUP INT TERM
  if [[ "${rc}" == "0" ]]; then
    rm -rf -- "${work}"
  else
    echo "SOURCE_WORKDIR_PRESERVED=${work}"
  fi
  exit "${rc}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

fail() {
  echo "S03_PUBLIC_BOOTSTRAP=FAILED"
  echo "ERROR=$*"
  return 1
}

[[ ${EUID} -eq 0 ]] || fail "run as root"
[[ "$(hostname)" == "${expected_host}" ]] || fail "unexpected host"
command -v git >/dev/null 2>&1 || fail "git is missing"
command -v apt-get >/dev/null 2>&1 || fail "apt-get is missing"
command -v apt-cache >/dev/null 2>&1 || fail "apt-cache is missing"
[[ -f /opt/pvnaive/FOUNDATION.json ]] || fail "S02 foundation marker is missing"
[[ ! -f /var/run/reboot-required ]] || fail "reboot is required before S03"
[[ -f /etc/caddy/Caddyfile ]] || fail "Caddyfile is missing"
current_caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${current_caddy_sha}" == "${expected_caddy_sha256}" ]] || fail "Caddyfile checksum mismatch"
systemctl is-active --quiet caddy-naive.service || fail "caddy-naive.service is not active"

if [[ -f /opt/pvnaive/S03_DATABASE.json ]]; then
  fail "S03 marker already exists; use the canonical verifier instead of a fresh bootstrap"
fi

cat <<EOF
=== ${stage_id} ===
UTC=${stamp}
HOST=$(hostname)
REPO=${repo_url}
PINNED_SOURCE_COMMIT=${source_commit}
LOG=${log}
EOF

# Fetch only the pinned, audited commit. No GitHub credentials are needed while
# the repository is public.
git -C "${work}" init --quiet
git -C "${work}" remote add origin "${repo_url}"
GIT_TERMINAL_PROMPT=0 git -C "${work}" fetch --quiet --depth=1 origin "${source_commit}"
fetched_commit="$(git -C "${work}" rev-parse FETCH_HEAD)"
[[ "${fetched_commit}" == "${source_commit}" ]] || fail "fetched commit mismatch: ${fetched_commit}"
git -C "${work}" checkout --quiet --detach "${source_commit}"
[[ "$(git -C "${work}" rev-parse HEAD)" == "${source_commit}" ]] || fail "checked out commit mismatch"
git -C "${work}" fsck --full --no-dangling >/dev/null
[[ "$(git -C "${work}" remote get-url origin)" == "${repo_url}" ]] || fail "origin URL mismatch"

echo "GIT_SOURCE_VERIFY=PASSED"

# Byte-level contract for the 13 production S03 files at the pinned commit.
declare -A expected_blobs=(
  ["db/migrations/0001_initial.down.sql"]="4ff634e1986720f9f283b57a93540fc2f643f6ea"
  ["db/migrations/0001_initial.up.sql"]="99bcccb65b180855860e971a937756792684c776"
  ["db/migrations/SHA256SUMS"]="e7cff7d3257ed95cd3207f50f8faea2779507164"
  ["scripts/db/backup.sh"]="1dc43cda6ac7011b80fc8f0d0f1215d7479fdcba"
  ["scripts/db/health.sh"]="fd6533101e4457878168591d354186957931bae9"
  ["scripts/db/lib.sh"]="17c3a19092e85da873192ab8ea880757a085b805"
  ["scripts/db/migrate.sh"]="dbfed23b748474e65eeb296dbb77a0d9bb656eb0"
  ["scripts/db/restore.sh"]="d370fda0d56bf267f0e50a20b42d80de23bcdf97"
  ["scripts/db/rollback.sh"]="5603aa46fb30510fa66f7ef52fba962050f3a11c"
  ["scripts/stages/S03-database.sh"]="95c8cd0e5923b8071d9c9749eec1ff46ccf3004d"
  ["scripts/stages/lib.sh"]="615d1ed4bc8fbaaa6ef9db2e3478e2d0ed6ff0fc"
  ["ops/systemd/pvnaive-db-health.service"]="78acd260310c75c19588ba4b8c9374e139ecb403"
  ["ops/systemd/pvnaive-db-health.timer"]="4c3067faf887b48cf84cde27cd69bb74a4444185"
)

for path in "${!expected_blobs[@]}"; do
  [[ -f "${work}/${path}" ]] || fail "missing pinned source file: ${path}"
  actual_blob="$(git -C "${work}" hash-object -- "${path}")"
  [[ "${actual_blob}" == "${expected_blobs[${path}]}" ]] || \
    fail "Git blob mismatch for ${path}: ${actual_blob}"
done
echo "S03_SOURCE_BLOBS=PASSED"
echo "S03_SOURCE_FILES=13"

# Static and regression gates before any package install.
bash -n "${work}"/scripts/db/*.sh "${work}"/scripts/stages/*.sh "${work}"/tests/stages/*.sh
(
  cd "${work}/db/migrations"
  sha256sum --check --strict SHA256SUMS
)
bash "${work}/tests/stages/S03_preflight_test.sh"
bash "${work}/tests/stages/S03_apt_candidate_pipefail_test.sh"
bash "${work}/tests/stages/S03_ubuntu2604_contract_test.sh"
echo "S03_LOCAL_REGRESSION_GATES=PASSED"

# Target OS and APT dependency-resolution gate. apt-get update only refreshes
# package metadata; package installation is still blocked until simulation passes.
. /etc/os-release
[[ "${ID:-}" == "ubuntu" && "${VERSION_ID:-}" == "26.04" ]] || \
  fail "unsupported OS ${ID:-unknown} ${VERSION_ID:-unknown}"

export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=l
apt-get update
packages=(postgresql postgresql-client age)
install_specs=()
for package_name in "${packages[@]}"; do
  candidate="$(apt-cache policy "${package_name}" | awk '/Candidate:/ {candidate=$2} END {if (candidate != "") print candidate}')"
  [[ -n "${candidate}" && "${candidate}" != "(none)" ]] || fail "no APT candidate for ${package_name}"
  echo "APT_CANDIDATE_${package_name//-/_}=${candidate}"
  install_specs+=("${package_name}=${candidate}")
done
[[ "${install_specs[0]}" == postgresql=18+* ]] || fail "PostgreSQL meta package is not major 18"
pg18_candidate="$(apt-cache policy postgresql-18 | awk '/Candidate:/ {candidate=$2} END {if (candidate != "") print candidate}')"
[[ -n "${pg18_candidate}" && "${pg18_candidate}" != "(none)" ]] || fail "postgresql-18 has no APT candidate"
apt-cache show postgresql-18 | awk '/^Provides:/ && $0 ~ /postgresql-contrib-18/ {found=1} END {exit !found}' || \
  fail "postgresql-18 does not provide postgresql-contrib-18"
apt-get --simulate install --no-install-recommends "${install_specs[@]}" >/dev/null
[[ ! -f /var/run/reboot-required ]] || fail "reboot became required during bootstrap"
echo "APT_SIMULATION=PASSED"
echo "POSTGRESQL18_CANDIDATE=${pg18_candidate}"

# Record invariant immediately before the real stage.
caddy_before="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${caddy_before}" == "${expected_caddy_sha256}" ]] || fail "Caddy changed before S03"

cat <<'EOF'
==================================================
ALL BOOTSTRAP GATES PASSED — STARTING REAL S03
==================================================
EOF

set +e
bash "${work}/scripts/stages/S03-database.sh" 2>&1 | tee "${log}"
stage_rc=${PIPESTATUS[0]}
set -e

echo "S03_STAGE_EXIT=${stage_rc}"
echo "S03_LOG=${log}"

caddy_after="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${caddy_after}" == "${caddy_before}" ]] || fail "Caddyfile changed during S03"
systemctl is-active --quiet caddy-naive.service || fail "Caddy inactive after S03"

if [[ "${stage_rc}" != "0" ]]; then
  echo "S03_PUBLIC_BOOTSTRAP=FAILED"
  exit "${stage_rc}"
fi

[[ -f /opt/pvnaive/S03_DATABASE.json ]] || fail "S03 exited zero without success marker"
grep -Fqx '  "stage": "S03-DATABASE",' /opt/pvnaive/S03_DATABASE.json || fail "invalid S03 success marker"

systemctl is-active --quiet pvnaive-db-health.timer || fail "database health timer inactive"
systemctl start pvnaive-db-health.service
[[ "$(systemctl show --property=Result --value pvnaive-db-health.service)" == "success" ]] || \
  fail "database health service did not finish successfully"

printf '%s\n' "S03_PUBLIC_BOOTSTRAP=PASSED" "PINNED_SOURCE_COMMIT=${source_commit}" "S03_LOG=${log}"
