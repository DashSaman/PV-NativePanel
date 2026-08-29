#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

artifacts_dir="${1:-dist/artifacts}"
archive="$(find "${artifacts_dir}" -maxdepth 1 -type f -name 'PVNaive-S05-*.tar.gz' -print | sort | tail -n1)"
[[ -n "${archive}" ]] || { echo 'ERROR: S05 archive missing' >&2; exit 1; }
[[ -f "${archive}.sha256" ]] || { echo 'ERROR: S05 archive checksum file missing' >&2; exit 1; }
(
  cd "${artifacts_dir}"
  sha256sum --check --strict "$(basename "${archive}").sha256" >/dev/null
)

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "${tmpdir}"' EXIT HUP INT TERM
tar -xzf "${archive}" -C "${tmpdir}"
root="$(find "${tmpdir}" -mindepth 1 -maxdepth 1 -type d -print -quit)"
[[ -n "${root}" ]] || { echo 'ERROR: S05 archive root missing' >&2; exit 1; }

required=(
  RELEASE.json
  bin/pvnaive
  bin/pvnaive-password
  bin/pvnaive-runtime-agent
  web/index.html
  public-site/index.html
  db/migrations/0003_naive_runtime_credentials.up.sql
  db/migrations/0003_naive_runtime_credentials.down.sql
  db/migrations/0004_customer_lifecycle_foundation.up.sql
  db/migrations/0004_customer_lifecycle_foundation.down.sql
  db/migrations/0005_customer_mutation_idempotency.up.sql
  db/migrations/0005_customer_mutation_idempotency.down.sql
  db/migrations/0006_direct_subscription_tokens.up.sql
  db/migrations/0006_direct_subscription_tokens.down.sql
  scripts/stages/S05-preflight.sh
  scripts/stages/S05-upgrade.sh
  systemd/pvnaive-api.service
  systemd/pvnaive-runtime-agent.service
  SHA256SUMS
)
for path in "${required[@]}"; do
  [[ -f "${root}/${path}" ]] || { echo "ERROR: S05 bundle missing ${path}" >&2; exit 1; }
done

for path in bin/pvnaive bin/pvnaive-password bin/pvnaive-runtime-agent; do
  [[ -x "${root}/${path}" ]] || { echo "ERROR: ${path} is not executable" >&2; exit 1; }
done
(
  cd "${root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
)

grep -Fq '"stage": "S05-CUSTOMER-SERVICE"' "${root}/RELEASE.json" || { echo 'ERROR: release stage is not S05' >&2; exit 1; }
grep -Fq '"schema_version": 6' "${root}/RELEASE.json" || { echo 'ERROR: release schema is not 6' >&2; exit 1; }
grep -Fq 'EnvironmentFile=-/etc/pvnaive/api.env' "${root}/systemd/pvnaive-api.service" || { echo 'ERROR: API unit does not load S05 api.env' >&2; exit 1; }

if grep -Rqs -- 'systemctl restart caddy-naive.service' "${root}/scripts"; then
  echo 'ERROR: S05 bundle contains forbidden Caddy restart' >&2
  exit 1
fi
if grep -Rqs -- 'systemctl reload caddy-naive.service' "${root}/scripts/stages/S05-upgrade.sh"; then
  echo 'ERROR: S05 installer contains forbidden Caddy reload' >&2
  exit 1
fi

forbidden='PVNaive|PVNETWORK|NaiveProxy|VPN|Proxy|Tunnel|Credential|Subscription Management'
if find "${root}/public-site" -type f -printf '%P\n' | grep -Eiq "${forbidden}"; then
  echo 'ERROR: bundled public newsroom contains a product-revealing file name' >&2
  exit 1
fi
if grep -RIiEq --binary-files=without-match "${forbidden}" "${root}/public-site"; then
  echo 'ERROR: bundled public newsroom leaks management/data-plane vocabulary' >&2
  exit 1
fi

echo 'S05_BUNDLE_CONTRACT=PASSED'
