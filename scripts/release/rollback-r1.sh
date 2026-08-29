#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

backup="${1:-}"
[[ ${EUID} -eq 0 ]] || { echo 'ERROR: run as root' >&2; exit 1; }
[[ "$backup" == /var/backups/pvnaive/releases/* && -d "$backup" ]] || { echo 'ERROR: valid PVNaive release backup directory required' >&2; exit 1; }
for f in pvnaive.before pvnaive-password.before pvnaive-runtime-agent.before pvnaive-api.service.before pvnaive-runtime-agent.service.before web.before preview.before; do [[ -f "$backup/$f" ]] || { echo "ERROR: rollback asset missing: $f" >&2; exit 1; }; done
web_before="$(cat "$backup/web.before")"
preview_before="$(cat "$backup/preview.before")"
[[ "$web_before" == /opt/pvnaive/web/releases/* && -d "$web_before" ]] || { echo 'ERROR: unsafe prior web target' >&2; exit 1; }
[[ "$preview_before" == /var/www/pvnaive-preview/releases/* && -d "$preview_before" ]] || { echo 'ERROR: unsafe prior preview target' >&2; exit 1; }

caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid="$(systemctl show caddy-naive.service -p MainPID --value)"
caddy_restarts="$(systemctl show caddy-naive.service -p NRestarts --value)"
systemctl stop pvnaive-api.service
systemctl stop pvnaive-runtime-agent.service
install -o root -g pvnaive -m 0750 "$backup/pvnaive.before" /opt/pvnaive/bin/pvnaive
install -o root -g pvnaive -m 0750 "$backup/pvnaive-password.before" /opt/pvnaive/bin/pvnaive-password
install -o root -g pvnaive -m 0750 "$backup/pvnaive-runtime-agent.before" /opt/pvnaive/bin/pvnaive-runtime-agent
install -o root -g root -m 0644 "$backup/pvnaive-api.service.before" /etc/systemd/system/pvnaive-api.service
install -o root -g root -m 0644 "$backup/pvnaive-runtime-agent.service.before" /etc/systemd/system/pvnaive-runtime-agent.service
ln -sfn "$web_before" /opt/pvnaive/web/current
ln -sfn "$preview_before" /var/www/pvnaive-preview/current
systemctl daemon-reload
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service
for _ in $(seq 1 30); do curl -fsS http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true' && break; sleep 1; done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true'
[[ "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')" == "$caddy_sha" ]]
[[ "$(systemctl show caddy-naive.service -p MainPID --value)" == "$caddy_pid" ]]
[[ "$(systemctl show caddy-naive.service -p NRestarts --value)" == "$caddy_restarts" ]]
echo 'PVNAIVE_R1_ROLLBACK_RESULT=PASSED'
echo 'PVNAIVE_R1_CADDY_ACTION=NONE'
