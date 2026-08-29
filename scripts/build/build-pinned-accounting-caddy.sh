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

for cmd in git go sha256sum cp mkdir mktemp find rm grep awk cmp; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "ERROR: ${cmd} is required" >&2; exit 1; }
done
[[ "${forwardproxy_commit}" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid pinned forwardproxy commit' >&2; exit 1; }
[[ "${caddy_version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'ERROR: invalid pinned Caddy version' >&2; exit 1; }
[[ "${xcaddy_version}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'ERROR: invalid pinned xcaddy version' >&2; exit 1; }
[[ -f "${patch_file}" ]] || { echo 'ERROR: forwardproxy patch is missing' >&2; exit 1; }
[[ -f "${overlay_dir}/pvnaive_accounting.go.src" ]] || { echo 'ERROR: forwardproxy overlay is missing' >&2; exit 1; }
[[ -f "${overlay_dir}/pvnaive_accounting_test.go.src" ]] || { echo 'ERROR: forwardproxy accounting tests are missing' >&2; exit 1; }

work_a="$(mktemp -d)"
work_b="$(mktemp -d)"
cleanup() {
  rm -rf -- "${work_a}" "${work_b}"
}
trap cleanup EXIT

prepare_forwardproxy() {
  local src="$1"
  git init -q "${src}"
  git -C "${src}" remote add origin https://github.com/klzgrad/forwardproxy.git
  git -C "${src}" fetch --quiet --depth=1 origin "${forwardproxy_commit}"
  git -C "${src}" checkout --quiet --detach FETCH_HEAD
  [[ "$(git -C "${src}" rev-parse HEAD)" == "${forwardproxy_commit}" ]] || { echo 'ERROR: forwardproxy pin mismatch' >&2; exit 1; }

  git -C "${src}" apply --check "${patch_file}"
  git -C "${src}" apply "${patch_file}"
  cp "${overlay_dir}/pvnaive_accounting.go.src" "${src}/pvnaive_accounting.go"
  cp "${overlay_dir}/pvnaive_accounting_test.go.src" "${src}/pvnaive_accounting_test.go"
  gofmt -w "${src}/pvnaive_accounting.go" "${src}/pvnaive_accounting_test.go" "${src}/forwardproxy.go" "${src}/caddyfile.go"

  (
    cd "${src}"
    go test ./...
    find . -type f -name '*_test.go' -delete
    go vet ./...
  )
}

build_vendor_candidate() {
  local work_root="$1"
  local patched_src="$2"
  local output="$3"
  local build_root="${work_root}/caddy-build"
  local local_fp="${build_root}/forwardproxy"

  mkdir -p "${build_root}"
  cp -a "${patched_src}" "${local_fp}"
  rm -rf -- "${local_fp}/.git"
  find "${local_fp}" -type f -name '*_test.go' -delete

  cat > "${build_root}/main.go" <<'GOEOF'
package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/caddyserver/forwardproxy"
)

func main() {
	caddycmd.Main()
}
GOEOF

  (
    cd "${build_root}"
    go mod init caddy >/dev/null
    go get "github.com/caddyserver/caddy/v2@${caddy_version}" >/dev/null
    # The upstream accounting source is already independently fetched and
    # verified by exact Git SHA above. A synthetic module version plus a fixed
    # relative replacement lets `go mod vendor` consume those audited bytes
    # without asking the Go proxy to resolve an unadvertised Git commit.
    go mod edit -require=github.com/caddyserver/forwardproxy@v0.0.0
    go mod edit -replace=github.com/caddyserver/forwardproxy=./forwardproxy
    go mod tidy >/dev/null
    go mod vendor

    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
      go build -mod=vendor -trimpath -buildvcs=false \
      -ldflags='-w -s -buildid=' \
      -o "${output}" .
  )
}

src_a="${work_a}/forwardproxy-src"
prepare_forwardproxy "${src_a}"
src_b="${work_b}/forwardproxy-src"
mkdir -p "${src_b}"
cp -a "${src_a}/." "${src_b}/"

mkdir -p "${out_dir}"
primary="${out_dir}/caddy-pvnaive-accounting"
secondary="${out_dir}/caddy-pvnaive-accounting.repro"
build_vendor_candidate "${work_a}" "${src_a}" "${primary}"
build_vendor_candidate "${work_b}" "${src_b}" "${secondary}"

sha_primary="$(sha256sum "${primary}" | awk '{print $1}')"
sha_secondary="$(sha256sum "${secondary}" | awk '{print $1}')"
[[ "${sha_primary}" == "${sha_secondary}" ]] || {
  echo "ERROR: accounting Caddy is not bit reproducible: ${sha_primary} != ${sha_secondary}" >&2
  exit 1
}
cmp -s "${primary}" "${secondary}" || { echo 'ERROR: reproducible SHA matched but binary bytes differ' >&2; exit 1; }

"${primary}" version > "${out_dir}/caddy.version.txt"
"${primary}" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null
go version -m "${primary}" > "${out_dir}/caddy.buildinfo.txt"
# A deterministic relative replacement is allowed; an absolute workspace path
# is not. This line must be identical in both independent builds.
grep -Fq $'\t=>\t./forwardproxy\t(devel)' "${out_dir}/caddy.buildinfo.txt"
if grep -Eq $'\t=>\t/' "${out_dir}/caddy.buildinfo.txt"; then
  echo 'ERROR: final binary contains an absolute local Go module replacement path' >&2
  exit 1
fi
grep -Fq $'github.com/caddyserver/caddy/v2\t'"${caddy_version}" "${out_dir}/caddy.buildinfo.txt"
grep -Fq $'github.com/caddyserver/forwardproxy\tv0.0.0' "${out_dir}/caddy.buildinfo.txt"

sha256sum "${primary}" > "${out_dir}/caddy-pvnaive-accounting.sha256"
printf '%s  caddy-pvnaive-accounting\n%s  caddy-pvnaive-accounting.repro\n' \
  "${sha_primary}" "${sha_secondary}" > "${out_dir}/caddy-pvnaive-accounting.repro.sha256"
sha256sum "${patch_file}" > "${out_dir}/forwardproxy-patch.sha256"
sha256sum "${overlay_dir}/pvnaive_accounting.go.src" > "${out_dir}/forwardproxy-overlay.sha256"
sha256sum "${overlay_dir}/pvnaive_accounting_test.go.src" > "${out_dir}/forwardproxy-overlay-test.sha256"
cat > "${out_dir}/PROVENANCE.txt" <<EOF
product=PVNaive
build_driver=go-vendor
caddy_version=${caddy_version}
forwardproxy_repo=https://github.com/klzgrad/forwardproxy.git
forwardproxy_commit=${forwardproxy_commit}
forwardproxy_module_version=v0.0.0
forwardproxy_replace=./forwardproxy
xcaddy_version_pin=${xcaddy_version}
xcaddy_used=false
patch_sha256=$(sha256sum "${patch_file}" | awk '{print $1}')
overlay_sha256=$(sha256sum "${overlay_dir}/pvnaive_accounting.go.src" | awk '{print $1}')
overlay_test_sha256=$(sha256sum "${overlay_dir}/pvnaive_accounting_test.go.src" | awk '{print $1}')
binary_sha256=${sha_primary}
reproducibility_verified=true
reproducible_second_sha256=${sha_secondary}
EOF

rm -f -- "${secondary}"
echo "PVNAIVE_REPRODUCIBILITY_PROOF=PASSED"
echo "PVNAIVE_ACCOUNTING_CADDY_BUILD=PASSED"
cat "${out_dir}/PROVENANCE.txt"
