#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
builder="${repo_root}/scripts/build/build-pinned-accounting-caddy.sh"

[[ -f "${builder}" ]] || { echo 'ERROR: accounting Caddy builder missing' >&2; exit 1; }

grep -Fq 'go mod vendor' "${builder}" || { echo 'ERROR: builder must vendor the pinned module graph' >&2; exit 1; }
grep -Fq -- '-mod=vendor' "${builder}" || { echo 'ERROR: final Caddy build must use vendored sources' >&2; exit 1; }
grep -Fq -- '-buildid=' "${builder}" || { echo 'ERROR: deterministic empty Go build id is required' >&2; exit 1; }
grep -Fq 'PVNAIVE_REPRODUCIBILITY_PROOF=PASSED' "${builder}" || { echo 'ERROR: builder must compare two independent output hashes' >&2; exit 1; }
grep -Fq 'reproducibility_verified=true' "${builder}" || { echo 'ERROR: provenance must record reproducibility verification' >&2; exit 1; }
grep -Fq 'build_driver=go-vendor' "${builder}" || { echo 'ERROR: provenance must identify the vendorized build driver' >&2; exit 1; }

if grep -Fq -- '--with "github.com/caddyserver/forwardproxy=${src}"' "${builder}"; then
  echo 'ERROR: final binary must not embed a random local module replacement path' >&2
  exit 1
fi

echo 'WS1_REPRODUCIBLE_CADDY_BUILD_CONTRACT=PASSED'
