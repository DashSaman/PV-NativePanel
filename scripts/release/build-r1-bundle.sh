#!/usr/bin/env bash
set -Eeuo pipefail
umask 022

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
out="${1:-${root}/dist}"
for cmd in go npm git sha256sum tar; do command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }; done
commit="$(git -C "$root" rev-parse HEAD)"
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || exit 1
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
release="PVNaive-R1-${stamp}-${commit:0:12}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
bundle="$tmp/$release"
mkdir -p "$bundle"/{bin,web,db/migrations,scripts/db,scripts/ops,scripts/release,systemd,sbom}

cd "$root"
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
(
  cd web
  npm ci --ignore-scripts --no-audit
  npm test
  npm run build
  npm ls --all --json >"$bundle/sbom/npm.json" || true
)
go list -m all >"$bundle/sbom/go-modules.txt"
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$bundle/bin/pvnaive" ./cmd/pvnaive
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$bundle/bin/pvnaive-password" ./cmd/pvnaive-password
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$bundle/bin/pvnaive-runtime-agent" ./cmd/pvnaive-runtime-agent
cp -a web/dist/. "$bundle/web/"
cp -a db/migrations/. "$bundle/db/migrations/"
cp -a scripts/db/. "$bundle/scripts/db/"
cp -a scripts/ops/. "$bundle/scripts/ops/"
cp -a scripts/release/*.sh scripts/release/*.py "$bundle/scripts/release/" 2>/dev/null || true
cp -a ops/systemd/pvnaive-*.service ops/systemd/pvnaive-*.timer "$bundle/systemd/" 2>/dev/null || true

# Hard fail if a tracked real private-key PEM appears in the release input.
if git grep -n -E -- '-----BEGIN (RSA |EC |OPENSSH |)?PRIVATE KEY-----' -- ':!**/*_test.go' ':!docs/**' ':!**/*.md'; then
  echo 'ERROR: private key material detected in tracked release input' >&2
  exit 1
fi
cat >"$bundle/RELEASE.json" <<JSON
{
  "product":"PVNaive",
  "channel":"r1",
  "source_commit":"$commit",
  "built_at_utc":"$stamp",
  "schema_version":8,
  "standalone_required":true,
  "fleet_controller_required":false,
  "accounting_exact_required_for_quota_alerts":true,
  "artifact_signed":false,
  "signature_note":"No signing key was supplied; checksums and source provenance are authoritative for this build."
}
JSON
(
  cd "$bundle"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)
mkdir -p "$out"
tar -C "$tmp" -czf "$out/${release}.tar.gz" "$release"
sha256sum "$out/${release}.tar.gz" >"$out/${release}.tar.gz.sha256"
echo "PVNAIVE_R1_BUILD_RESULT=PASSED"
echo "PVNAIVE_R1_ARTIFACT=$out/${release}.tar.gz"
