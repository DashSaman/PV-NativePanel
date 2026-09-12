#!/usr/bin/env bash
set -euo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
builder="$root/scripts/release/build-r1-bundle.sh"
deploy="$root/scripts/release/deploy-r1.sh"
rollback="$root/scripts/release/rollback-r1.sh"
for f in "$builder" "$deploy" "$rollback"; do [[ -f "$f" ]] || { echo "ERROR: missing $f" >&2; exit 1; }; done

grep -Fq 'dist/ws1-accounting/caddy-pvnaive-accounting' "$builder" || { echo 'ERROR: R1 builder does not require the pinned Task13 Caddy candidate' >&2; exit 1; }
grep -Fq 'caddy/caddy-pvnaive-accounting' "$builder" || { echo 'ERROR: R1 builder does not package the Task13 Caddy candidate' >&2; exit 1; }
grep -Fq 'caddy-naive.service.d/20-pvnaive-accounting.conf' "$builder" || { echo 'ERROR: R1 builder does not package the Caddy Task13 drop-in' >&2; exit 1; }

backup_line="$(grep -n -F 'caddy-before' "$deploy" | head -1 | cut -d: -f1)"
install_line="$(grep -n -F 'install -o root -g root -m 0755 "$bundle/caddy/caddy-pvnaive-accounting" /usr/local/bin/caddy' "$deploy" | head -1 | cut -d: -f1)"
[[ -n "$backup_line" && -n "$install_line" && "$backup_line" -lt "$install_line" ]] || { echo 'ERROR: Caddy backup must precede Task13 binary install' >&2; exit 1; }
grep -Fq 'systemctl restart caddy-naive.service' "$deploy" || { echo 'ERROR: Task13 R1 deploy must perform one controlled Caddy restart' >&2; exit 1; }
[[ "$(grep -c -F 'systemctl restart caddy-naive.service' "$deploy")" -eq 1 ]] || { echo 'ERROR: Task13 R1 deploy must contain exactly one controlled Caddy restart' >&2; exit 1; }

grep -Fq 'caddy-before' "$rollback" || { echo 'ERROR: rollback does not restore prior Caddy binary' >&2; exit 1; }
grep -Fq '20-pvnaive-accounting.conf.before' "$rollback" || { echo 'ERROR: rollback does not restore prior Caddy drop-in state' >&2; exit 1; }
grep -Fq 'systemctl restart caddy-naive.service' "$rollback" || { echo 'ERROR: rollback does not reactivate restored Caddy binary' >&2; exit 1; }

echo 'TASK13_R1_RELEASE_CONTRACT=PASSED'
