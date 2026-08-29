#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

artifacts_dir="${1:-dist/artifacts}"
archive="$(find "${artifacts_dir}" -maxdepth 1 -type f -name 'PVNaive-S04*.tar.gz' -print | sort | tail -n1)"
[[ -n "${archive}" ]] || { echo 'ERROR: S04/S04R archive missing' >&2; exit 1; }
[[ -f "${archive}.sha256" ]] || { echo 'ERROR: archive checksum file missing' >&2; exit 1; }
(
  cd "${artifacts_dir}"
  sha256sum --check --strict "$(basename "${archive}").sha256" >/dev/null
)

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "${tmpdir}"' EXIT HUP INT TERM
tar -xzf "${archive}" -C "${tmpdir}"
root="$(find "${tmpdir}" -mindepth 1 -maxdepth 1 -type d -print -quit)"
[[ -n "${root}" ]] || { echo 'ERROR: archive root missing' >&2; exit 1; }

required=(
  bin/pvnaive
  bin/pvnaive-password
  bin/pvnaive-runtime-agent
  web/index.html
  public-site/index.html
  public-site/assets/site.css
  public-site/assets/site.js
  public-site/assets/news-mark.svg
  public-site/assets/news-fallback.svg
  public-site/data/articles.json
  db/migrations/0003_naive_runtime_credentials.up.sql
  db/migrations/0003_naive_runtime_credentials.down.sql
  db/migrations/0004_customer_lifecycle_foundation.up.sql
  db/migrations/0004_customer_lifecycle_foundation.down.sql
  systemd/pvnaive-api.service
  systemd/pvnaive-runtime-agent.service
  scripts/stages/S04R-preflight.sh
  scripts/stages/S04R-upgrade.sh
  SHA256SUMS
)
for path in "${required[@]}"; do
  [[ -f "${root}/${path}" ]] || { echo "ERROR: bundle missing ${path}" >&2; exit 1; }
done

[[ -x "${root}/bin/pvnaive" ]] || { echo 'ERROR: pvnaive is not executable' >&2; exit 1; }
[[ -x "${root}/bin/pvnaive-password" ]] || { echo 'ERROR: pvnaive-password is not executable' >&2; exit 1; }
[[ -x "${root}/bin/pvnaive-runtime-agent" ]] || { echo 'ERROR: pvnaive-runtime-agent is not executable' >&2; exit 1; }

(
  cd "${root}"
  sha256sum --check --strict SHA256SUMS >/dev/null
)

if grep -Rqs -- 'systemctl restart caddy-naive.service' "${root}/scripts"; then
  echo 'ERROR: bundle contains forbidden Caddy restart' >&2
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

grep -Fq '/data/articles.json' "${root}/public-site/assets/site.js" || {
  echo 'ERROR: bundled public newsroom lost its local-cache contract' >&2
  exit 1
}

echo 'S04R_BUNDLE_CONTRACT=PASSED'
