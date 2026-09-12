#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
forwardproxy_commit="$(tr -d '[:space:]' < "${repo_root}/third_party/forwardproxy/UPSTREAM_COMMIT")"
patch1="${repo_root}/third_party/forwardproxy/patches/0001-pvnaive-exact-accounting.patch"
patch2="${repo_root}/third_party/forwardproxy/patches/0002-pvnaive-session-control.patch"
patch3="${repo_root}/third_party/forwardproxy/patches/0003-pvnaive-session-control-lifecycle.patch"
overlay="${repo_root}/third_party/forwardproxy/overlay"
work="$(mktemp -d)"
trap 'rm -rf -- "${work}"' EXIT

src="${work}/forwardproxy"
git init -q "${src}"
git -C "${src}" remote add origin https://github.com/klzgrad/forwardproxy.git
git -C "${src}" fetch --quiet --depth=1 origin "${forwardproxy_commit}"
git -C "${src}" checkout --quiet --detach FETCH_HEAD
[[ "$(git -C "${src}" rev-parse HEAD)" == "${forwardproxy_commit}" ]]

git -C "${src}" apply --check "${patch1}"
git -C "${src}" apply "${patch1}"
cp "${overlay}/pvnaive_accounting.go.src" "${src}/pvnaive_accounting.go"
cp "${overlay}/pvnaive_accounting_test.go.src" "${src}/pvnaive_accounting_test.go"
git -C "${src}" apply --check "${patch2}"
git -C "${src}" apply "${patch2}"
git -C "${src}" apply --check "${patch3}"
git -C "${src}" apply "${patch3}"
gofmt -w "${src}"/*.go
(
  cd "${src}"
  go test -race ./... -count=1
)

echo 'TASK13_FORWARDPROXY_SESSION_CONTROL=PASSED'
