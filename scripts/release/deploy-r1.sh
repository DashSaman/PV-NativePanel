#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

bundle="${1:-}"
[[ ${EUID} -eq 0 ]] || { echo 'ERROR: deploy must run as root' >&2; exit 1; }
[[ -d "$bundle" && -f "$bundle/RELEASE.json" && -f "$bundle/SHA256SUMS" ]] || { echo 'ERROR: unpacked R1 bundle required' >&2; exit 1; }
for cmd in sha256sum systemctl systemd-tmpfiles curl readlink install cp ln mkdir runuser psql awk sed tar; do
  command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }
done
( cd "$bundle" && sha256sum --check --strict SHA256SUMS >/dev/null ) || { echo 'ERROR: release checksums failed' >&2; exit 1; }
grep -q '"product"[[:space:]]*:[[:space:]]*"PVNaive"' "$bundle/RELEASE.json" || { echo 'ERROR: product mismatch' >&2; exit 1; }

commit="$(sed -nE 's/.*"source_commit"[[:space:]]*:[[:space:]]*"([0-9a-f]{40})".*/\1/p' "$bundle/RELEASE.json")"
release_schema="$(sed -nE 's/.*"schema_version"[[:space:]]*:[[:space:]]*([0-9]+).*/\1/p' "$bundle/RELEASE.json")"
[[ "$commit" =~ ^[0-9a-f]{40}$ ]] || { echo 'ERROR: invalid source commit' >&2; exit 1; }
[[ "$release_schema" =~ ^[1-9][0-9]*$ ]] || { echo 'ERROR: invalid release schema' >&2; exit 1; }

[[ -r /etc/pvnaive/db.env ]] || { echo 'ERROR: /etc/pvnaive/db.env missing' >&2; exit 1; }
set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
: "${PVNAIVE_DB_PORT:=5432}"
current_schema="$(runuser -u postgres -- psql --no-psqlrc -At --host /var/run/postgresql --port "$PVNAIVE_DB_PORT" --username postgres --dbname pvnaive --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "$current_schema" == "$release_schema" ]] || { echo "ERROR: same-schema deploy required: release=$release_schema current=$current_schema" >&2; exit 1; }

for unit in postgresql.service caddy-naive.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service pvnaive-api.service; do
  systemctl is-active --quiet "$unit" || { echo "ERROR: $unit inactive before deploy" >&2; exit 1; }
done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || { echo 'ERROR: API not ready before deploy' >&2; exit 1; }
curl -fsS --unix-socket /run/pvnaive/runtime-agent.sock http://unix/v1/health | grep -q '"status":"ok"' || { echo 'ERROR: runtime agent unhealthy before deploy' >&2; exit 1; }
curl -fsS --unix-socket /run/pvnaive/accounting.sock http://unix/v1/accounting/health | grep -q '"status":"ok"' || { echo 'ERROR: telemetry accounting.sock unhealthy before deploy' >&2; exit 1; }

caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
caddy_pid="$(systemctl show caddy-naive.service -p MainPID --value)"
caddy_restarts="$(systemctl show caddy-naive.service -p NRestarts --value)"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup="/var/backups/pvnaive/releases/${stamp}-${commit:0:12}"
install -d -m 0700 "$backup"

# Mandatory encrypted PostgreSQL + configuration snapshot before any replacement.
backup_output="$(PVNAIVE_DB_SCRIPT_ROOT="$bundle/scripts/db" bash "$bundle/scripts/ops/backup-run.sh")"
grep -q '^PVNAIVE_SCHEDULED_BACKUP_RESULT=PASSED$' <<<"$backup_output" || { echo 'ERROR: pre-deploy encrypted snapshot failed' >&2; exit 1; }
printf '%s\n' "$backup_output" >"$backup/pre-deploy-backup.txt"

for name in pvnaive pvnaive-password pvnaive-runtime-agent pvnaive-telemetry-agent; do
  [[ -f "/opt/pvnaive/bin/$name" ]] || { echo "ERROR: current binary missing: $name" >&2; exit 1; }
  cp -a "/opt/pvnaive/bin/$name" "$backup/$name.before"
done
cp -a /etc/caddy/Caddyfile "$backup/Caddyfile.before"

for unit in pvnaive-api.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service pvnaive-backup.service pvnaive-backup.timer pvnaive-restore-drill.service pvnaive-restore-drill.timer; do
  if [[ -f "/etc/systemd/system/$unit" ]]; then
    cp -a "/etc/systemd/system/$unit" "$backup/$unit.before"
  else
    : >"$backup/$unit.missing"
  fi
done
if [[ -f /etc/tmpfiles.d/pvnaive.conf ]]; then
  cp -a /etc/tmpfiles.d/pvnaive.conf "$backup/pvnaive.conf.before"
else
  : >"$backup/pvnaive.conf.missing"
fi

web_before="$(readlink -f /opt/pvnaive/web/current 2>/dev/null || true)"
db_before="$(readlink -f /opt/pvnaive/db/current 2>/dev/null || true)"
printf '%s\n' "$web_before" >"$backup/web.before"
printf '%s\n' "$db_before" >"$backup/db.before"
if [[ -f /opt/pvnaive/release/CURRENT ]]; then cp -a /opt/pvnaive/release/CURRENT "$backup/CURRENT.before"; else : >"$backup/CURRENT.missing"; fi
if [[ -f /opt/pvnaive/release/RELEASE.json ]]; then cp -a /opt/pvnaive/release/RELEASE.json "$backup/RELEASE.json.before"; else : >"$backup/RELEASE.json.missing"; fi

web_release="/opt/pvnaive/web/releases/${stamp}-${commit:0:12}"
db_release="/opt/pvnaive/db/releases/${stamp}-${commit:0:12}"
install -d -o root -g pvnaive -m 0750 "$web_release"
cp -a "$bundle/web/." "$web_release/"
chown -R root:pvnaive "$web_release"
install -d -o root -g pvnaive -m 0750 "$db_release/scripts/db"
cp -a "$bundle/scripts/db/." "$db_release/scripts/db/"
chown -R root:pvnaive "$db_release"

rollback_on_error() {
  code=$?
  trap - ERR INT TERM HUP
  set +e
  bash "$bundle/scripts/release/rollback-r1.sh" "$backup"
  exit "$code"
}
trap rollback_on_error ERR INT TERM HUP

systemctl stop pvnaive-api.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service
for name in pvnaive pvnaive-password pvnaive-runtime-agent pvnaive-telemetry-agent; do
  install -o root -g pvnaive -m 0750 "$bundle/bin/$name" "/opt/pvnaive/bin/$name"
done

install -d -o root -g root -m 0750 /opt/pvnaive/ops /opt/pvnaive/release /opt/pvnaive/db/releases /opt/pvnaive/web/releases
install -o root -g root -m 0750 "$bundle/scripts/ops/"*.sh /opt/pvnaive/ops/
for file in "$bundle/systemd/"*; do
  [[ -f "$file" ]] && install -o root -g root -m 0644 "$file" "/etc/systemd/system/$(basename "$file")"
done
if [[ -f "$bundle/tmpfiles/pvnaive.conf" ]]; then
  install -o root -g root -m 0644 "$bundle/tmpfiles/pvnaive.conf" /etc/tmpfiles.d/pvnaive.conf
  systemd-tmpfiles --create /etc/tmpfiles.d/pvnaive.conf
fi
ln -sfn "$web_release" /opt/pvnaive/web/current
ln -sfn "$db_release" /opt/pvnaive/db/current
cp -a "$bundle/RELEASE.json" /opt/pvnaive/release/RELEASE.json
printf '%s\n' "$commit" >/opt/pvnaive/release/CURRENT

systemctl daemon-reload
systemctl restart pvnaive-telemetry-agent.service
systemctl restart pvnaive-runtime-agent.service
systemctl restart pvnaive-api.service
systemctl enable --now pvnaive-backup.timer pvnaive-restore-drill.timer

for _ in $(seq 1 30); do
  if curl -fsS http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then break; fi
  sleep 1
done
curl -fsS http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true'
curl -fsS --unix-socket /run/pvnaive/runtime-agent.sock http://unix/v1/health | grep -q '"status":"ok"'
curl -fsS --unix-socket /run/pvnaive/accounting.sock http://unix/v1/accounting/health | grep -q '"status":"ok"'
for unit in pvnaive-api.service pvnaive-runtime-agent.service pvnaive-telemetry-agent.service pvnaive-backup.timer pvnaive-restore-drill.timer; do systemctl is-active --quiet "$unit"; done
[[ "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')" == "$caddy_sha" ]]
[[ "$(systemctl show caddy-naive.service -p MainPID --value)" == "$caddy_pid" ]]
[[ "$(systemctl show caddy-naive.service -p NRestarts --value)" == "$caddy_restarts" ]]
trap - ERR INT TERM HUP

echo 'PVNAIVE_R1_DEPLOY_RESULT=PASSED'
echo "PVNAIVE_R1_SOURCE_COMMIT=$commit"
echo "PVNAIVE_R1_SCHEMA_VERSION=$release_schema"
echo "PVNAIVE_R1_BACKUP_DIR=$backup"
echo 'PVNAIVE_R1_CADDY_ACTION=NONE'
