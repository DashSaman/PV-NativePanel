#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
for command_name in go npm tar sha256sum git; do
  command -v "${command_name}" >/dev/null 2>&1 || { echo "ERROR: missing ${command_name}" >&2; exit 1; }
done

source_commit="${PVNAIVE_SOURCE_COMMIT:-$(git -C "${repo_root}" rev-parse HEAD)}"
[[ "${source_commit}" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid source commit' >&2; exit 1; }
short_commit="${source_commit:0:12}"
output_dir="${PVNAIVE_OUTPUT_DIR:-${repo_root}/dist/artifacts}"
work_root="$(mktemp -d)"
cleanup() { rm -rf -- "${work_root}"; }
trap cleanup EXIT HUP INT TERM
bundle_name="PVNaive-S05-${short_commit}"
bundle_root="${work_root}/${bundle_name}"
mkdir -p "${bundle_root}"/{bin,web,public-site,db/migrations,scripts/db,scripts/auth,scripts/stages,systemd}

cd "${repo_root}"
go mod verify
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive" ./cmd/pvnaive
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive-password" ./cmd/pvnaive-password
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/bin/pvnaive-runtime-agent" ./cmd/pvnaive-runtime-agent

(
  cd web
  npm ci --ignore-scripts --no-audit
  npm test
  npm run build
)
cp -a web/dist/. "${bundle_root}/web/"
cp -a site/. "${bundle_root}/public-site/"

cp -a db/migrations/*.sql db/migrations/SHA256SUMS "${bundle_root}/db/migrations/"
cp -a scripts/db/*.sh "${bundle_root}/scripts/db/"
cp -a scripts/auth/bootstrap-owner.sh "${bundle_root}/scripts/auth/"
cp -a scripts/stages/lib.sh scripts/stages/S05-preflight.sh scripts/stages/S05-upgrade.sh "${bundle_root}/scripts/stages/"
cp -a ops/systemd/pvnaive-api.service ops/systemd/pvnaive-runtime-agent.service "${bundle_root}/systemd/"
chmod 0750 "${bundle_root}"/bin/*
chmod 0750 "${bundle_root}"/scripts/db/*.sh "${bundle_root}"/scripts/auth/*.sh "${bundle_root}"/scripts/stages/*.sh
chmod 0644 "${bundle_root}"/systemd/*.service

cat >"${bundle_root}/RELEASE.json" <<JSON
{
  "product": "PVNaive",
  "stage": "S05-CUSTOMER-SERVICE",
  "mode": "guarded-existing-server",
  "source_commit": "${source_commit}",
  "target_host": "testAmir5-3",
  "panel_host": "namir.softarg.ir",
  "api_listener": "127.0.0.1:8080",
  "api_environment": "/etc/pvnaive/api.env",
  "runtime_socket": "/run/pvnaive/runtime-agent.sock",
  "runtime_key": "/etc/pvnaive/runtime.key",
  "schema_version": 6,
  "panel_base": "/panel/",
  "public_site_path": "public-site/",
  "caddy_installer_mutation": false,
  "caddy_runtime_action": "validate-then-reload-only-for-credential-mutations",
  "ssh_mutation": false,
  "firewall_mutation": false,
  "usage_accounting_proven": false,
  "first_success_connect_producer_proven": false
}
JSON

(
  cd "${bundle_root}"
  find RELEASE.json bin web public-site db scripts systemd -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)

mkdir -p "${output_dir}"
archive="${output_dir}/${bundle_name}.tar.gz"
rm -f -- "${archive}" "${archive}.sha256"
tar --sort=name --mtime='UTC 2026-01-01' --owner=0 --group=0 --numeric-owner \
  -C "${work_root}" -czf "${archive}" "${bundle_name}"
archive_sha="$(sha256sum "${archive}" | awk '{print $1}')"
printf '%s  %s\n' "${archive_sha}" "$(basename -- "${archive}")" > "${archive}.sha256"

echo 'S05_BUNDLE_BUILD=PASSED'
echo "S05_SOURCE_COMMIT=${source_commit}"
echo "S05_BUNDLE_PATH=${archive}"
echo "S05_BUNDLE_SHA256=${archive_sha}"
