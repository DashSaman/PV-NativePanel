#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: diagnostic bundle must run as root' >&2; exit 1; }
for cmd in tar journalctl systemctl sed sha256sum; do command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }; done
root="${PVNAIVE_DIAGNOSTIC_ROOT:-/var/lib/pvnaive/diagnostics}"
[[ "$root" == /var/lib/pvnaive/* ]] || { echo 'ERROR: unsafe diagnostics root' >&2; exit 1; }
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
tmp="$(mktemp -d "${root}.tmp.${stamp}.XXXXXX")"
out="${root}/pvnaive-diagnostics-${stamp}.tar.gz"
cleanup(){ rm -rf -- "$tmp"; }
trap cleanup EXIT HUP INT TERM
install -d -m 0700 "$root"

redact(){
  sed -E \
    -e 's#(Authorization:[[:space:]]*Bearer[[:space:]]+)[^[:space:],;]+#\1[REDACTED]#Ig' \
    -e 's#((password|token|secret|private[_-]?key|subscription[_-]?token)[[:space:]]*[=:][[:space:]]*)[^[:space:],;]+#\1[REDACTED]#Ig' \
    -e 's#(postgres(ql)?://[^:/[:space:]@]+:)[^@[:space:]]+@#\1[REDACTED]@#Ig' \
    -e 's#(naive\+https://[^:/[:space:]@]+:)[^@[:space:]]+@#\1[REDACTED]@#Ig' \
    -e 's#(/(sub|s)/)[A-Za-z0-9_-]{32,}#\1[REDACTED]#g' \
    -e 's#(/api/v1/subscriptions/)[A-Za-z0-9_-]{32,}#\1[REDACTED]#g'
}

{
  echo "product=PVNaive"
  echo "created_at_utc=$stamp"
  uname -srmo
} >"$tmp/system.txt"

if /opt/pvnaive/bin/pvnaive doctor --json >"$tmp/doctor.json" 2>"$tmp/doctor.stderr"; then :; else true; fi
for unit in pvnaive-api.service pvnaive-runtime-agent.service caddy-naive.service postgresql.service pvnaive-backup.timer pvnaive-restore-drill.timer; do
  systemctl show "$unit" --property=Id,LoadState,ActiveState,SubState,MainPID,NRestarts --no-pager 2>/dev/null || true
done | redact >"$tmp/services.txt"

# Environment values are deliberately excluded. Only variable names are useful for support.
for env_file in /etc/pvnaive/api.env /etc/pvnaive/db.env; do
  [[ -r "$env_file" ]] || continue
  echo "[$env_file]"
  sed -nE 's/^([A-Za-z_][A-Za-z0-9_]*)=.*/\1=[PRESENT]/p' "$env_file"
done >"$tmp/environment-keys.txt"

for unit in pvnaive-api.service pvnaive-runtime-agent.service caddy-naive.service; do
  echo "===== $unit ====="
  journalctl -u "$unit" --since '-30 minutes' --no-pager -n 300 2>/dev/null || true
done | redact >"$tmp/recent-journal.txt"

{
  echo '=== listeners ==='
  ss -ltn 2>/dev/null | sed -n '1,100p' || true
  echo '=== disk ==='
  df -h / /var /opt 2>/dev/null || true
  echo '=== memory ==='
  free -h 2>/dev/null || true
} | redact >"$tmp/resources.txt"

if [[ -f /opt/pvnaive/release/CURRENT ]]; then cat /opt/pvnaive/release/CURRENT | redact >"$tmp/release.txt"; fi
(
  cd "$tmp"
  sha256sum ./* > SHA256SUMS
)
tar -C "$tmp" -czf "$out" .
chmod 0600 "$out"
trap - EXIT HUP INT TERM
rm -rf -- "$tmp"
echo 'PVNAIVE_DIAGNOSTIC_BUNDLE_RESULT=PASSED'
echo "PVNAIVE_DIAGNOSTIC_BUNDLE=$out"
