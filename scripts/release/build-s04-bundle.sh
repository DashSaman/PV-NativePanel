#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
for command_name in go npm tar sha256sum git; do
  command -v "${command_name}" >/dev/null 2>&1 || { echo "ERROR: missing ${command_name}" >&2; exit 1; }
done

source_commit="${PVNAIVE_SOURCE_COMMIT:-$(git -C "${repo_root}" rev-parse HEAD)}"
[[ "${source_commit}" =~ ^[0-9a-f]{40}$ ]] || { echo "ERROR: invalid source commit" >&2; exit 1; }
short_commit="${source_commit:0:12}"
output_dir="${PVNAIVE_OUTPUT_DIR:-${repo_root}/dist/artifacts}"
work_root="$(mktemp -d)"
cleanup() { rm -rf -- "${work_root}"; }
trap cleanup EXIT HUP INT TERM
bundle_name="PVNaive-S04-${short_commit}"
bundle_root="${work_root}/${bundle_name}"
mkdir -p "${bundle_root}"/{db/migrations,scripts/db,scripts/auth,scripts/stages,ops/systemd,dist/s04/linux-amd64,dist/s04/web}

cd "${repo_root}"
go mod verify
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/dist/s04/linux-amd64/pvnaive" ./cmd/pvnaive
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags='-s -w -buildid=' -o "${bundle_root}/dist/s04/linux-amd64/pvnaive-password" ./cmd/pvnaive-password

(
  cd web
  npm ci --ignore-scripts --no-audit
  npm test
  npm run build
)
cp -a web/dist/. "${bundle_root}/dist/s04/web/"

cp -a db/migrations/*.sql db/migrations/SHA256SUMS "${bundle_root}/db/migrations/"
cp -a scripts/db/*.sh "${bundle_root}/scripts/db/"
cp -a scripts/auth/bootstrap-owner.sh "${bundle_root}/scripts/auth/"
cp -a scripts/stages/lib.sh scripts/stages/S04-auth.sh "${bundle_root}/scripts/stages/"
cp -a ops/systemd/pvnaive-api.service "${bundle_root}/ops/systemd/"
chmod 0750 "${bundle_root}/dist/s04/linux-amd64/pvnaive" "${bundle_root}/dist/s04/linux-amd64/pvnaive-password"
chmod 0750 "${bundle_root}"/scripts/db/*.sh "${bundle_root}"/scripts/auth/*.sh "${bundle_root}"/scripts/stages/*.sh

cat >"${bundle_root}/RELEASE.json" <<JSON
{
  "product": "PVNaive",
  "stage": "S04-AUTH",
  "mode": "localhost-first",
  "source_commit": "${source_commit}",
  "target_host": "testAmir5-3",
  "api_listener": "127.0.0.1:8080",
  "caddy_mutation": false,
  "ssh_mutation": false,
  "firewall_mutation": false
}
JSON

(
  cd "${bundle_root}"
  find RELEASE.json db scripts ops dist -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum > S04_SHA256SUMS
  sha256sum --check --strict S04_SHA256SUMS >/dev/null
)

mkdir -p "${output_dir}"
archive="${output_dir}/${bundle_name}.tar.gz"
rm -f -- "${archive}" "${archive}.sha256"
tar --sort=name --mtime='UTC 2026-01-01' --owner=0 --group=0 --numeric-owner \
  -C "${work_root}" -czf "${archive}" "${bundle_name}"
archive_sha="$(sha256sum "${archive}" | awk '{print $1}')"
printf '%s  %s\n' "${archive_sha}" "$(basename -- "${archive}")" > "${archive}.sha256"

echo "S04_BUNDLE_BUILD=PASSED"
echo "S04_SOURCE_COMMIT=${source_commit}"
echo "S04_BUNDLE_PATH=${archive}"
echo "S04_BUNDLE_SHA256=${archive_sha}"
