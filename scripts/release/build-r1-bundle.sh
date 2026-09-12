#!/usr/bin/env bash
set -Eeuo pipefail
umask 022

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
out="${1:-${root}/dist}"
for cmd in go npm git sha256sum tar awk sort; do
  command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }
done

commit="$(git -C "$root" rev-parse HEAD)"
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid source commit' >&2; exit 1; }

latest_schema="$({
  find "$root/db/migrations" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]_*.up.sql' -printf '%f\n'
} | sed -nE 's/^([0-9]{4})_.*/\1/p' | sort -n | tail -n1 | sed -E 's/^0+//')"
[[ "$latest_schema" =~ ^[1-9][0-9]*$ ]] || { echo 'ERROR: could not derive latest schema' >&2; exit 1; }

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
release="PVNaive-R1-${stamp}-${commit:0:12}"
tmp="$(mktemp -d)"
trap 'rm -rf -- "$tmp"' EXIT
bundle="$tmp/$release"
mkdir -p "$bundle"/{bin,web,db/migrations,scripts/db,scripts/ops,scripts/release,systemd/caddy-naive.service.d,tmpfiles,sbom,caddy}

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
CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o "$bundle/bin/pvnaive-telemetry-agent" ./cmd/pvnaive-telemetry-agent

cp -a web/dist/. "$bundle/web/"
cp -a db/migrations/. "$bundle/db/migrations/"
cp -a scripts/db/. "$bundle/scripts/db/"
cp -a scripts/ops/. "$bundle/scripts/ops/"
cp -a scripts/release/*.sh scripts/release/*.py "$bundle/scripts/release/" 2>/dev/null || true
cp -a ops/systemd/pvnaive-*.service ops/systemd/pvnaive-*.timer "$bundle/systemd/" 2>/dev/null || true
cp -a ops/systemd/caddy-naive.service.d/20-pvnaive-accounting.conf "$bundle/systemd/caddy-naive.service.d/20-pvnaive-accounting.conf"

# Task13 changes the Caddy data plane, so an R1 release must carry the exact
# reproducible candidate that passed the pinned forwardproxy gate.
caddy_src="$root/dist/ws1-accounting/caddy-pvnaive-accounting"
[[ -x "$caddy_src" ]] || { echo 'ERROR: pinned Task13 Caddy candidate missing; run scripts/build/build-pinned-accounting-caddy.sh first' >&2; exit 1; }
for proof in caddy-pvnaive-accounting.sha256 caddy.version.txt caddy.buildinfo.txt PROVENANCE.txt; do
  [[ -f "$root/dist/ws1-accounting/$proof" ]] || { echo "ERROR: pinned Task13 Caddy proof missing: $proof" >&2; exit 1; }
done
cp -a "$caddy_src" "$bundle/caddy/caddy-pvnaive-accounting"
cp -a "$root/dist/ws1-accounting/caddy-pvnaive-accounting.sha256" "$bundle/caddy/"
cp -a "$root/dist/ws1-accounting/caddy.version.txt" "$bundle/caddy/"
cp -a "$root/dist/ws1-accounting/caddy.buildinfo.txt" "$bundle/caddy/"
cp -a "$root/dist/ws1-accounting/PROVENANCE.txt" "$bundle/caddy/"
chmod 0755 "$bundle/caddy/caddy-pvnaive-accounting"
caddy_sha="$(sha256sum "$bundle/caddy/caddy-pvnaive-accounting" | awk '{print $1}')"
grep -Fq "binary_sha256=$caddy_sha" "$bundle/caddy/PROVENANCE.txt" || { echo 'ERROR: Task13 Caddy provenance mismatch' >&2; exit 1; }
grep -Fq 'reproducibility_verified=true' "$bundle/caddy/PROVENANCE.txt" || { echo 'ERROR: Task13 Caddy reproducibility proof missing' >&2; exit 1; }

cp -a ops/tmpfiles/. "$bundle/tmpfiles/"

if git grep -n -E -- '-----BEGIN (RSA |EC |OPENSSH |)?PRIVATE KEY-----' -- ':!**/*_test.go' ':!docs/**' ':!**/*.md'; then
  echo 'ERROR: private key material detected in tracked release input' >&2
  exit 1
fi

cat >"$bundle/RELEASE.json" <<JSON
{
  "product": "PVNaive",
  "channel": "r1",
  "source_commit": "$commit",
  "built_at_utc": "$stamp",
  "schema_version": $latest_schema,
  "standalone_required": true,
  "fleet_controller_required": false,
  "caddy_mutation_required": true,
  "caddy_runtime_action": "one-controlled-binary-swap-restart",
  "task13_caddy_sha256": "$caddy_sha",
  "artifact_signed": false,
  "signature_note": "No signing key supplied; checksums and source provenance are authoritative."
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

echo 'PVNAIVE_R1_BUILD_RESULT=PASSED'
echo "PVNAIVE_R1_SOURCE_COMMIT=$commit"
echo "PVNAIVE_R1_SCHEMA_VERSION=$latest_schema"
echo "PVNAIVE_R1_ARTIFACT=$out/${release}.tar.gz"
