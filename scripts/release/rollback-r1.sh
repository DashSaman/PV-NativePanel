#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

backup="${1:-}"
[[ ${EUID} -eq 0 ]] || { echo 'ERROR: rollback must run as root' >&2; exit 1; }
[[ "$backup" == /var/backups/pvnaive/releases/* && -d "$backup" ]] || { echo 'ERROR: valid PVNaive release backup directory required' >&2; exit 1; }
for cmd in sha256sum systemctl systemd-tmpfiles curl install ln readlink; do
  command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }
done
for name in pvnaive pvnaive-password pvnaive-runtime-agent pvnaive-telemetry-agent; do
  [[ -f "$backup/$name.before" ]] || { echo "ERROR: rollback asset missing: $name.before" >&2; exit 1; }
done
[[ -f "$backup/web.before" && -f "$backup/db.before" ]] || { echo 'ERROR: rollback target metadata missing' >&2; exit 1; }

web_before="$(cat "$backup/web.before")"
db_before="$(cat "$backup/db.before")"
if [[ -n "$web_before" ]]; then [[ "$web_before" == /opt/pvnaive/web/releases/* && -d "$web_before" ]] || { echo 'ERROR: unsafe prior web target' >&2; exit 1; }; fi
if [[ -n "$db_before" ]]; then [[ "$db_before" == /opt/pvnaive/db/releases/* && -d "$db_before" ]] || { echo 'ERROR: unsafe prior DB-script target' >&2; exit 1; }; fi

caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid="$(systemctl show caddy-naive.service -p MainPID --value)"
caddy_restarts="$(systemctl show caddy-naive.service -p NRestarts --value)"

systemctl stop pvnaive-api.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service || true
for name in pvnaive pvnaive-password pvnaive-runtime-agent pvnaive-telemetry-agent; do
  install -o root -g pvnaive -m 0750 "$backup/$name.before" "/opt/pvnaive/bin/$name"
done

for unit in pvnaive-api.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service pvnaive-backup.service pvnaive-backup.timer pvnaive-restore-drill.service pvnaive-restore-drill.timer; do
  if [[ -f "$backup/$unit.before" ]]; then
    install -o root -g root -m 0644 "$backup/$unit.before" "/etc/systemd/system/$unit"
  elif [[ -f "$backup/$unit.missing" ]]; then
    systemctl disable --now "$unit" >/dev/null 2>&1 || true
    rm -f -- "/etc/systemd/system/$unit"
  else
    echo "ERROR: rollback unit state missing: $unit" >&2
    exit 1
  fi
done

if [[ -f "$backup/pvnaive.conf.before" ]]; then
  install -o root -g root -m 0644 "$backup/pvnaive.conf.before" /etc/tmpfiles.d/pvnaive.conf
  systemd-tmpfiles --create /etc/tmpfiles.d/pvnaive.conf
elif [[ -f "$backup/pvnaive.conf.missing" ]]; then
  rm -f -- /etc/tmpfiles.d/pvnaive.conf
fi

if [[ -n "$web_before" ]]; then ln -sfn "$web_before" /opt/pvnaive/web/current; fi
if [[ -n "$db_before" ]]; then ln -sfn "$db_before" /opt/pvnaive/db/current; fi
if [[ -f "$backup/CURRENT.before" ]]; then
  install -o root -g root -m 0644 "$backup/CURRENT.before" /opt/pvnaive/release/CURRENT
elif [[ -f "$backup/CURRENT.missing" ]]; then
  rm -f -- /opt/pvnaive/release/CURRENT
fi
if [[ -f "$backup/RELEASE.json.before" ]]; then
  install -o root -g root -m 0644 "$backup/RELEASE.json.before" /opt/pvnaive/release/RELEASE.json
elif [[ -f "$backup/RELEASE.json.missing" ]]; then
  rm -f -- /opt/pvnaive/release/RELEASE.json
fi

systemctl daemon-reload
systemctl restart pvnaive-telemetry-agent.service
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service
for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then break; fi
  sleep 1
done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true'
curl -fsS --unix-socket /run/pvnaive/runtime-agent.sock http://unix/v1/health | grep -q '"status":"ok"'
curl -fsS --unix-socket /run/pvnaive/accounting.sock http://unix/v1/accounting/health | grep -q '"status":"ok"'
[[ "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')" == "$caddy_sha" ]]
[[ "$(systemctl show caddy-naive.service -p MainPID --value)" == "$caddy_pid" ]]
[[ "$(systemctl show caddy-naive.service -p NRestarts --value)" == "$caddy_restarts" ]]

echo 'PVNAIVE_R1_ROLLBACK_RESULT=PASSED'
echo 'PVNAIVE_R1_CADDY_ACTION=NONE'
