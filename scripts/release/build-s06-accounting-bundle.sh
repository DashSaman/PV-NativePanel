#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
for cmd in go npm tar sha256sum git cp awk; do
  command -v "${cmd}" >/dev/null 2>&1 || { echo "ERROR: missing ${cmd}" >&2; exit 1; }
done

checkout_head="$(git -C "${repo_root}" rev-parse HEAD)"
source_commit="${PVNAIVE_SOURCE_COMMIT:-${checkout_head}}"
[[ "${source_commit}" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid source commit' >&2; exit 1; }
[[ "${source_commit}" == "${checkout_head}" ]] || { echo 'ERROR: source commit does not match checkout HEAD' >&2; exit 1; }
if ! git -C "${repo_root}" diff --quiet -- || ! git -C "${repo_root}" diff --cached --quiet -- || [[ -n "$(git -C "${repo_root}" ls-files --others --exclude-standard)" ]]; then
  echo 'ERROR: PVNAIVE source tree is dirty; build from an exact clean checkout' >&2
  exit 1
fi
short_commit="${source_commit:0:12}"
output_dir="${PVNAIVE_OUTPUT_DIR:-${repo_root}/dist/artifacts}"
work_root="$(mktemp -d)"
trap 'rm -rf -- "${work_root}"' EXIT HUP INT TERM
bundle_name="PVNaive-S06-Accounting-${short_commit}"
bundle_root="${work_root}/${bundle_name}"
mkdir -p "${bundle_root}"/{bin,web,db/migrations,scripts/db,scripts/stages,systemd,caddy}

cd "${repo_root}"
go mod verify
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive" ./cmd/pvnaive
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive-password" ./cmd/pvnaive-password
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive-runtime-agent" ./cmd/pvnaive-runtime-agent
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive-telemetry-agent" ./cmd/pvnaive-telemetry-agent
(
  cd web
  npm ci --ignore-scripts --no-audit
  npm test
  npm run build
)
cp -a web/dist/. "${bundle_root}/web/"
cp -a db/migrations/*.sql db/migrations/SHA256SUMS "${bundle_root}/db/migrations/"
cp -a scripts/db/*.sh "${bundle_root}/scripts/db/"
cp -a scripts/stages/S06-accounting-preflight.sh scripts/stages/S06-accounting-upgrade.sh "${bundle_root}/scripts/stages/"
cp -a ops/systemd/pvnaive-api.service ops/systemd/pvnaive-runtime-agent.service ops/systemd/pvnaive-telemetry-agent.service "${bundle_root}/systemd/"

if [[ ! -f "${repo_root}/dist/ws1-accounting/caddy-pvnaive-accounting" ]]; then
  echo 'ERROR: pinned accounting Caddy binary is missing; run scripts/build/build-pinned-accounting-caddy.sh first' >&2
  exit 1
fi
cp -a "${repo_root}/dist/ws1-accounting/caddy-pvnaive-accounting" "${bundle_root}/caddy/caddy-pvnaive-accounting"
cp -a "${repo_root}/dist/ws1-accounting/caddy-pvnaive-accounting.sha256" "${bundle_root}/caddy/caddy-pvnaive-accounting.sha256"
cp -a "${repo_root}/dist/ws1-accounting/PROVENANCE.txt" "${bundle_root}/PROVENANCE.txt"
cp -a "${repo_root}/dist/ws1-accounting/caddy.version.txt" "${bundle_root}/caddy/caddy.version.txt"
cp -a "${repo_root}/dist/ws1-accounting/caddy.buildinfo.txt" "${bundle_root}/caddy/caddy.buildinfo.txt"

accounting_caddy_sha="$(sha256sum "${bundle_root}/caddy/caddy-pvnaive-accounting" | awk '{print $1}')"
candidate_manifest_sha="$(awk '{print $1}' "${bundle_root}/caddy/caddy-pvnaive-accounting.sha256")"
[[ "${accounting_caddy_sha}" == "${candidate_manifest_sha}" ]] || {
  echo "ERROR: accounting Caddy SHA ${accounting_caddy_sha} does not match manifest ${candidate_manifest_sha}" >&2
  exit 1
}

chmod 0750 "${bundle_root}"/bin/* "${bundle_root}"/scripts/db/*.sh "${bundle_root}"/scripts/stages/*.sh
chmod 0755 "${bundle_root}"/caddy/caddy-pvnaive-accounting
chmod 0644 "${bundle_root}"/systemd/*.service

cat >"${bundle_root}/RELEASE.json" <<JSON
{
  "product": "PVNaive",
  "stage": "S06-ACCOUNTING-RELEASE",
  "mode": "guarded-incremental-existing-server",
  "source_commit": "${source_commit}",
  "target_host": "testAmir5-3",
  "panel_host": "namir.softarg.ir",
  "api_listener": "127.0.0.1:8080",
  "runtime_socket": "/run/pvnaive/runtime-agent.sock",
  "base_schema_version": 16,
  "schema_version": 17,
  "migration": "0017_operator_session_peers.up.sql",
  "panel_base": "/panel/",
  "caddy_installer_mutation": false,
  "caddy_runtime_action": "one-controlled-binary-swap-restart",
  "public_root_mutation": false,
  "accounting_caddy_sha256": "${accounting_caddy_sha}",
  "usage_accounting_proven": false,
  "hard_quota_enforcement_enabled": false,
  "first_success_connect_producer_proven": false
}
JSON

(
  cd "${bundle_root}"
  find RELEASE.json PROVENANCE.txt bin web db scripts systemd caddy -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)

mkdir -p "${output_dir}"
archive="${output_dir}/${bundle_name}.tar.gz"
rm -f -- "${archive}" "${archive}.sha256"
tar --sort=name --mtime='UTC 2026-01-01' --owner=0 --group=0 --numeric-owner -C "${work_root}" -czf "${archive}" "${bundle_name}"
archive_sha="$(sha256sum "${archive}" | awk '{print $1}')"
printf '%s  %s\n' "${archive_sha}" "$(basename -- "${archive}")" >"${archive}.sha256"

echo 'S06_ACCOUNTING_BUNDLE_BUILD=PASSED'
echo "S06_SOURCE_COMMIT=${source_commit}"
echo "PVNAIVE_ACCOUNTING_CADDY_SHA256=${accounting_caddy_sha}"
echo "S06_BUNDLE_PATH=${archive}"
echo "S06_BUNDLE_SHA256=${archive_sha}"
