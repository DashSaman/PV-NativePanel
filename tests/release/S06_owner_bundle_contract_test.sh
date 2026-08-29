#!/usr/bin/env bash
set -Eeuo pipefail

artifact_dir="${1:-dist/artifacts}"
archive="$(find "${artifact_dir}" -maxdepth 1 -type f -name 'PVNaive-S06-Owner-*.tar.gz' -print -quit)"
[[ -n "${archive}" && -f "${archive}" ]] || { echo 'ERROR: S06 owner archive missing' >&2; exit 1; }
[[ -f "${archive}.sha256" ]] || { echo 'ERROR: S06 owner archive checksum missing' >&2; exit 1; }
(
  cd "$(dirname -- "${archive}")"
  sha256sum --check --strict "$(basename -- "${archive}").sha256" >/dev/null
)

tmp="$(mktemp -d)"
trap 'rm -rf -- "${tmp}"' EXIT HUP INT TERM
tar -xzf "${archive}" -C "${tmp}"
root="$(find "${tmp}" -mindepth 1 -maxdepth 1 -type d -name 'PVNaive-S06-Owner-*' -print -quit)"
[[ -n "${root}" ]] || { echo 'ERROR: S06 bundle root missing' >&2; exit 1; }
(
  cd "${root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
)

for required in \
  RELEASE.json SHA256SUMS bin/pvnaive bin/pvnaive-password bin/pvnaive-runtime-agent web/index.html \
  db/migrations/0007_subscription_token_recovery.up.sql db/migrations/0007_subscription_token_recovery.down.sql \
  db/migrations/SHA256SUMS scripts/db/backup.sh scripts/db/migrate.sh scripts/db/rollback.sh \
  scripts/db/promote-release.sh scripts/db/set-expected-schema-version.sh \
  scripts/stages/S06-owner-preflight.sh scripts/stages/S06-owner-upgrade.sh \
  systemd/pvnaive-api.service systemd/pvnaive-runtime-agent.service; do
  [[ -f "${root}/${required}" ]] || { echo "ERROR: S06 bundle missing ${required}" >&2; exit 1; }
done

for token in \
  '"product": "PVNaive"' \
  '"stage": "S06-OWNER-CUSTOMER-OPS"' \
  '"base_schema_version": 6' \
  '"schema_version": 7' \
  '"caddy_installer_mutation": false' \
  '"public_root_mutation": false' \
  '"usage_accounting_proven": false' \
  '"hard_quota_enforcement_enabled": false' \
  '"first_success_connect_producer_proven": false'; do
  grep -Fq -- "${token}" "${root}/RELEASE.json" || { echo "ERROR: release metadata missing ${token}" >&2; exit 1; }
done

if find "${root}" -maxdepth 1 -type d -name public-site -print -quit | grep -q .; then
  echo 'ERROR: S06 owner bundle must not publish/mutate public root'
  exit 1
fi
bash -n "${root}/scripts/stages/S06-owner-preflight.sh" "${root}/scripts/stages/S06-owner-upgrade.sh"

echo 'S06_OWNER_BUNDLE_CONTRACT=PASSED'
