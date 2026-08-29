#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

bundle="${1:-}"
[[ ${EUID} -eq 0 ]] || { echo 'ERROR: run as root' >&2; exit 1; }
[[ -d "$bundle" && -f "$bundle/RELEASE.json" && -f "$bundle/SHA256SUMS" ]] || { echo 'ERROR: unpacked R1 bundle required' >&2; exit 1; }
for cmd in sha256sum systemctl curl readlink install cp ln mv runuser psql awk grep stat; do command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }; done
( cd "$bundle" && sha256sum --check --strict SHA256SUMS >/dev/null ) || { echo 'ERROR: release checksums failed' >&2; exit 1; }
grep -q '"product":"PVNaive"' "$bundle/RELEASE.json" || { echo 'ERROR: product mismatch' >&2; exit 1; }
grep -q '"schema_version":8' "$bundle/RELEASE.json" || { echo 'ERROR: schema mismatch' >&2; exit 1; }
commit="$(sed -nE 's/.*"source_commit":"([0-9a-f]{40})".*/\1/p' "$bundle/RELEASE.json")"
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid source commit' >&2; exit 1; }

set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
schema="$(runuser -u postgres -- psql --no-psqlrc -At --host /var/run/postgresql --port "$PVNAIVE_DB_PORT" --username postgres --dbname pvnaive --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "$schema" == 8 ]] || { echo "ERROR: R1 deploy requires schema 8, got $schema" >&2; exit 1; }
for unit in postgresql.service caddy-naive.service pvnaive-runtime-agent.service pvnaive-api.service; do systemctl is-active --quiet "$unit" || { echo "ERROR: $unit inactive" >&2; exit 1; }; done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || { echo 'ERROR: API not ready before deploy' >&2; exit 1; }
caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid="$(systemctl show caddy-naive.service -p MainPID --value)"
caddy_restarts="$(systemctl show caddy-naive.service -p NRestarts --value)"

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup="/var/backups/pvnaive/releases/${stamp}-${commit:0:12}"
install -d -m 0700 "$backup"
cp -a /opt/pvnaive/bin/pvnaive "$backup/pvnaive.before"
cp -a /opt/pvnaive/bin/pvnaive-password "$backup/pvnaive-password.before"
cp -a /opt/pvnaive/bin/pvnaive-runtime-agent "$backup/pvnaive-runtime-agent.before"
cp -a /etc/systemd/system/pvnaive-api.service "$backup/pvnaive-api.service.before"
cp -a /etc/systemd/system/pvnaive-runtime-agent.service "$backup/pvnaive-runtime-agent.service.before"
web_before="$(readlink -f /opt/pvnaive/web/current)"
preview_before="$(readlink -f /var/www/pvnaive-preview/current)"
printf '%s\n' "$web_before" >"$backup/web.before"
printf '%s\n' "$preview_before" >"$backup/preview.before"

# Mandatory encrypted DB backup before service replacement.
db_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql PVNAIVE_DB_PORT="$PVNAIVE_DB_PORT" PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres PVNAIVE_RUN_AS_OS_USER=postgres \
  bash "$bundle/scripts/db/backup.sh"
)"
db_backup="$(awk -F= '$1=="PVNAIVE_BACKUP_PATH"{print $2}' <<<"$db_output")"
[[ -f "$db_backup" ]] || { echo 'ERROR: pre-deploy DB backup failed' >&2; exit 1; }
printf '%s\n' "$db_backup" >"$backup/database-backup.path"

web_release="/opt/pvnaive/web/releases/${stamp}-${commit:0:12}"
preview_release="/var/www/pvnaive-preview/releases/${stamp}-${commit:0:12}"
rollback(){
  code=$?
  trap - ERR INT TERM HUP
  set +e
  systemctl stop pvnaive-api.service
  systemctl stop pvnaive-runtime-agent.service
  cp -a "$backup/pvnaive.before" /opt/pvnaive/bin/pvnaive
  cp -a "$backup/pvnaive-password.before" /opt/pvnaive/bin/pvnaive-password
  cp -a "$backup/pvnaive-runtime-agent.before" /opt/pvnaive/bin/pvnaive-runtime-agent
  cp -a "$backup/pvnaive-api.service.before" /etc/systemd/system/pvnaive-api.service
  cp -a "$backup/pvnaive-runtime-agent.service.before" /etc/systemd/system/pvnaive-runtime-agent.service
  [[ -d "$web_before" ]] && ln -sfn "$web_before" /opt/pvnaive/web/current
  [[ -d "$preview_before" ]] && ln -sfn "$preview_before" /var/www/pvnaive-preview/current
  systemctl daemon-reload
  systemctl restart pvnaive-runtime-agent.service
  systemctl restart pvnaive-api.service
  echo 'PVNAIVE_R1_ROLLBACK=COMPLETED_BEST_EFFORT'
  exit "$code"
}
trap rollback ERR INT TERM HUP

install -d -o root -g pvnaive -m 0750 "$web_release"
cp -a "$bundle/web/." "$web_release/"
chown -R root:pvnaive "$web_release"
ln -sfn "$web_release" /opt/pvnaive/web/current
install -d -o root -g caddy -m 0750 "$preview_release"
cp -a "$bundle/web/." "$preview_release/"
chown -R root:caddy "$preview_release"
find "$preview_release" -type d -exec chmod 0750 {} +
find "$preview_release" -type f -exec chmod 0640 {} +
ln -sfn "$preview_release" /var/www/pvnaive-preview/current

install -o root -g pvnaive -m 0750 "$bundle/bin/pvnaive" /opt/pvnaive/bin/pvnaive
install -o root -g pvnaive -m 0750 "$bundle/bin/pvnaive-password" /opt/pvnaive/bin/pvnaive-password
install -o root -g pvnaive -m 0750 "$bundle/bin/pvnaive-runtime-agent" /opt/pvnaive/bin/pvnaive-runtime-agent
install -d -o root -g root -m 0750 /opt/pvnaive/ops
install -o root -g root -m 0750 "$bundle/scripts/ops/"*.sh /opt/pvnaive/ops/
for f in "$bundle/systemd/"*; do [[ -f "$f" ]] && install -o root -g root -m 0644 "$f" "/etc/systemd/system/$(basename "$f")"; done
install -d -o root -g root -m 0750 /opt/pvnaive/release
cp -a "$bundle/RELEASE.json" /opt/pvnaive/release/RELEASE.json
printf '%s\n' "$commit" >/opt/pvnaive/release/CURRENT

systemctl daemon-reload
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service
systemctl enable --now pvnaive-backup.timer pvnaive-restore-drill.timer
for _ in $(seq 1 30); do curl -fsS http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true' && break; sleep 1; done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true'
curl -fsS --unix-socket /run/pvnaive/runtime-agent.sock http://unix/v1/health | grep -q '"status":"ok"'
[[ "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')" == "$caddy_sha" ]]
[[ "$(systemctl show caddy-naive.service -p MainPID --value)" == "$caddy_pid" ]]
[[ "$(systemctl show caddy-naive.service -p NRestarts --value)" == "$caddy_restarts" ]]
trap - ERR INT TERM HUP

echo 'PVNAIVE_R1_DEPLOY_RESULT=PASSED'
echo "PVNAIVE_R1_SOURCE_COMMIT=$commit"
echo "PVNAIVE_R1_BACKUP_DIR=$backup"
echo "PVNAIVE_R1_DB_BACKUP=$db_backup"
echo 'PVNAIVE_R1_CADDY_ACTION=NONE'
