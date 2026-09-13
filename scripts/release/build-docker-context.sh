#!/usr/bin/env bash
# Build the self-contained all-in-one Docker build context (dist/docker).
#
# The Dockerfile expects prebuilt amd64 binaries, the built web panel and the
# pinned accounting proxy binary; this script assembles exactly that layout so
#   docker build -t pvnaive:local -f docker/Dockerfile .
# works straight from a repository clone.
#
# Usage:
#   bash scripts/release/build-docker-context.sh [--build-caddy]
#
# The pinned accounting proxy binary is reused from
# PVNAIVE_PINNED_CADDY_BIN (or dist/docker/caddy-pvnaive-accounting) unless
# --build-caddy runs the full xcaddy build from third_party pins.
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
out="${repo_root}/dist/docker"
build_caddy=false
for arg in "$@"; do
  case "${arg}" in
    --build-caddy) build_caddy=true ;;
    *) echo "ERROR: unknown argument: ${arg}" >&2; exit 1 ;;
  esac
done

for cmd in go npm sha256sum cp rm mkdir; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "ERROR: ${cmd} is required" >&2; exit 1; }
done

rm -rf "${out}"
mkdir -p "${out}/web"

echo "BUILD_GO_BINARIES"
build_go() {
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false \
    -ldflags='-s -w -buildid=' -o "${out}/$1" "$2"
}
build_go pvnaive-linux-amd64 ./cmd/pvnaive
build_go pvnaive-password-linux-amd64 ./cmd/pvnaive-password
build_go pvnaive-runtime-agent-linux-amd64 ./cmd/pvnaive-runtime-agent
build_go pvnaive-telemetry-agent-linux-amd64 ./cmd/pvnaive-telemetry-agent

echo "BUILD_WEB_PANEL"
( cd "${repo_root}/web" && npm ci --no-audit --no-fund >/dev/null && npm run build )
web_dist="$(ls -d "${repo_root}/web"/dist 2>/dev/null || true)"
[[ -n "${web_dist}" ]] || { echo "ERROR: web build output not found" >&2; exit 1; }
cp -a "${web_dist}/." "${out}/web/"

if [[ "${build_caddy}" == true ]]; then
  echo "BUILD_PINNED_ACCOUNTING_PROXY"
  bash "${repo_root}/scripts/build/build-pinned-accounting-caddy.sh"
  cp "${repo_root}/dist/ws1-accounting/caddy-pvnaive-accounting" "${out}/caddy-pvnaive-accounting"
else
  pinned="${PVNAIVE_PINNED_CADDY_BIN:-${out}/caddy-pvnaive-accounting}"
  if [[ ! -f "${pinned}" ]]; then
    echo "ERROR: pinned accounting proxy binary not found at ${pinned}" >&2
    echo "       set PVNAIVE_PINNED_CADDY_BIN or pass --build-caddy" >&2
    exit 1
  fi
  cp "${pinned}" "${out}/caddy-pvnaive-accounting"
fi
chmod 0755 "${out}"/*.linux-amd64 "${out}/caddy-pvnaive-accounting"

echo "VERIFY_CONTEXT"
required=(
  pvnaive-linux-amd64
  pvnaive-password-linux-amd64
  pvnaive-runtime-agent-linux-amd64
  pvnaive-telemetry-agent-linux-amd64
  caddy-pvnaive-accounting
  web/index.html
)
for item in "${required[@]}"; do
  [[ -s "${out}/${item}" ]] || { echo "ERROR: missing build context artifact: ${item}" >&2; exit 1; }
done
( cd "${out}" && sha256sum pvnaive-linux-amd64 pvnaive-password-linux-amd64 pvnaive-runtime-agent-linux-amd64 pvnaive-telemetry-agent-linux-amd64 caddy-pvnaive-accounting > SHA256SUMS )
echo "PVNAIVE_DOCKER_CONTEXT=${out}"
echo "PVNAIVE_DOCKER_CONTEXT_BUILD=PASSED"
