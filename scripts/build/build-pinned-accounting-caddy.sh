#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
forwardproxy_commit="$(tr -d '[:space:]' < "${repo_root}/third_party/forwardproxy/UPSTREAM_COMMIT")"
caddy_version="$(tr -d '[:space:]' < "${repo_root}/third_party/forwardproxy/CADDY_VERSION")"
xcaddy_version="$(tr -d '[:space:]' < "${repo_root}/third_party/forwardproxy/XCADDY_VERSION")"
patch_file="${repo_root}/third_party/forwardproxy/patches/0001-pvnaive-exact-accounting.patch"
overlay_dir="${repo_root}/third_party/forwardproxy/overlay"
out_dir="${PVNAIVE_ACCOUNTING_BUILD_OUT:-${repo_root}/dist/ws1-accounting}"

for cmd in git go sha256sum cp mkdir mktemp; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "ERROR: ${cmd} is required" >&2; exit 1; }
done
[[ "${forwardproxy_commit}" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid pinned forwardproxy commit' >&2; exit 1; }
[[ "${caddy_version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'ERROR: invalid pinned Caddy version' >&2; exit 1; }
[[ "${xcaddy_version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'ERROR: invalid pinned xcaddy version' >&2; exit 1; }
[[ -f "${patch_file}" ]] || { echo 'ERROR: forwardproxy patch is missing' >&2; exit 1; }
[[ -f "${overlay_dir}/pvnaive_accounting.go" ]] || { echo 'ERROR: forwardproxy overlay is missing' >&2; exit 1; }

tmp="$(mktemp -d)"
cleanup() { rm -rf -- "${tmp}"; }
trap cleanup EXIT
src="${tmp}/forwardproxy"
git init -q "${src}"
git -C "${src}" remote add origin https://github.com/klzgrad/forwardproxy.git
git -C "${src}" fetch --quiet --depth=1 origin "${forwardproxy_commit}"
git -C "${src}" checkout --quiet --detach FETCH_HEAD
[[ "$(git -C "${src}" rev-parse HEAD)" == "${forwardproxy_commit}" ]] || { echo 'ERROR: forwardproxy pin mismatch' >&2; exit 1; }

git -C "${src}" apply --check "${patch_file}"
git -C "${src}" apply "${patch_file}"
cp "${overlay_dir}/"*.go "${src}/"

gofmt -w "${src}/pvnaive_accounting.go" "${src}/forwardproxy.go" "${src}/caddyfile.go"
(
  cd "${src}"
  go test ./...
  go vet ./...
)

mkdir -p "${tmp}/bin" "${out_dir}"
GOBIN="${tmp}/bin" go install "github.com/caddyserver/xcaddy/cmd/xcaddy@${xcaddy_version}"
"${tmp}/bin/xcaddy" build "${caddy_version}" \
  --output "${out_dir}/caddy-pvnaive-accounting" \
  --with "github.com/caddyserver/forwardproxy=${src}"

"${out_dir}/caddy-pvnaive-accounting" version > "${out_dir}/caddy.version.txt"
"${out_dir}/caddy-pvnaive-accounting" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null
sha256sum "${out_dir}/caddy-pvnaive-accounting" > "${out_dir}/caddy-pvnaive-accounting.sha256"
sha256sum "${patch_file}" > "${out_dir}/forwardproxy-patch.sha256"
sha256sum "${overlay_dir}/pvnaive_accounting.go" > "${out_dir}/forwardproxy-overlay.sha256"
cat > "${out_dir}/PROVENANCE.txt" <<EOF
product=PVNaive
caddy_version=${caddy_version}
forwardproxy_repo=https://github.com/klzgrad/forwardproxy.git
forwardproxy_commit=${forwardproxy_commit}
xcaddy_version=${xcaddy_version}
patch_sha256=$(sha256sum "${patch_file}" | awk '{print $1}')
overlay_sha256=$(sha256sum "${overlay_dir}/pvnaive_accounting.go" | awk '{print $1}')
binary_sha256=$(sha256sum "${out_dir}/caddy-pvnaive-accounting" | awk '{print $1}')
EOF

echo "PVNAIVE_ACCOUNTING_CADDY_BUILD=PASSED"
cat "${out_dir}/PROVENANCE.txt"
