#!/usr/bin/env bash
set -Eeuo pipefail

builder="scripts/release/build-s06-accounting-bundle.sh"
[[ -f "${builder}" ]] || { echo 'ERROR: S06 accounting bundle builder missing' >&2; exit 1; }
bash -n "${builder}"

if grep -Eq 'base_schema_version[^0-9]*(7|8)|schema_version[^0-9]*(7|8)|0007_exact_accounting' "${builder}"; then
  echo 'ERROR: stale schema7/8 accounting bundle constants detected' >&2
  exit 1
fi

for token in \
  'caddy-pvnaive-accounting' \
  'caddy-pvnaive-accounting.sha256' \
  'PVNAIVE_ACCOUNTING_CADDY_SHA256=' \
  'RELEASE.json' \
  'SHA256SUMS' \
  'S06-ACCOUNTING-RELEASE' \
  '"base_schema_version": 16' \
  '"schema_version": 17' \
  '"usage_accounting_proven": false' \
  'build-pinned-accounting-caddy.sh' \
  'PROVENANCE.txt'; do
  grep -Fq -- "${token}" "${builder}" || { echo "ERROR: S06 accounting bundle builder missing ${token}" >&2; exit 1; }
done

bundle_script_dir="$(dirname "${builder}")"
preflight="scripts/stages/S06-accounting-preflight.sh"
upgrade="scripts/stages/S06-accounting-upgrade.sh"
[[ -f "${preflight}" ]] || { echo 'ERROR: S06 accounting preflight missing' >&2; exit 1; }
[[ -f "${upgrade}" ]] || { echo 'ERROR: S06 accounting upgrade missing' >&2; exit 1; }
bash -n "${preflight}" "${upgrade}"

echo 'S06_ACCOUNTING_BUNDLE_CONTRACT=PASSED'
# Task12 trusted peer registration is served by the telemetry agent; release must ship it.
grep -Fq 'pvnaive-telemetry-agent' "${builder}" || { echo 'ERROR: S06 accounting bundle must build/package telemetry agent' >&2; exit 1; }
# Release provenance must be the exact clean checkout, never an arbitrary label.
grep -Fq 'PVNAIVE source tree is dirty' "${builder}" || { echo 'ERROR: S06 builder must reject dirty source trees' >&2; exit 1; }
grep -Fq 'source commit does not match checkout HEAD' "${builder}" || { echo 'ERROR: S06 builder must bind source_commit to checkout HEAD' >&2; exit 1; }
# The Caddy checksum shipped inside the bundle must use the bundle-local basename,
# not an absolute build-workspace path that the upgrade gate cannot accept.
grep -Fq 'bundle-local basename' "${builder}" || { echo 'ERROR: S06 builder must normalize the Caddy manifest to a bundle-local basename' >&2; exit 1; }
grep -Fq "'caddy-pvnaive-accounting'" "${builder}" || { echo 'ERROR: S06 builder must emit the Caddy basename in its checksum manifest' >&2; exit 1; }
