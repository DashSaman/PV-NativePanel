#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S05-PREFLIGHT"
panel_host="${PVNAIVE_PUBLIC_HOST:-namir.softarg.ir}"
naive_public_host="${PVNAIVE_NAIVE_PUBLIC_HOST:-}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
db_env="/etc/pvnaive/db.env"
api_env="/etc/pvnaive/api.env"
failures=()

record_failure() { failures+=("$1"); printf 'CHECK_%s=FAIL\n' "$1"; }
record_pass() { printf 'CHECK_%s=PASS\n' "$1"; }

validate_public_host() {
  local value="$1" host port=""
  [[ -n "${value}" ]] || return 1
  [[ "${value}" != *://* && "${value}" != */* && "${value}" != *' '* ]] || return 1
  host="${value}"
  if [[ "${value}" == *:* ]]; then
    host="${value%:*}"
    port="${value##*:}"
    [[ "${port}" =~ ^[1-9][0-9]{0,4}$ ]] || return 1
    ((10#${port} <= 65535)) || return 1
  fi
  [[ "${host}" =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]]
}

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: run read-only S05 preflight as root' >&2; exit 1; }

echo "=== ${stage_id} ==="
echo "UTC=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "HOSTNAME=$(hostname)"
echo "KERNEL=$(uname -srmo)"
echo "PANEL_PUBLIC_HOST=${panel_host}"
echo "NAIVE_PUBLIC_HOST=${naive_public_host:-unset}"

if [[ -z "${naive_public_host}" ]]; then
  record_failure NAIVE_PUBLIC_HOST
elif validate_public_host "${naive_public_host}"; then
  record_pass NAIVE_PUBLIC_HOST
else
  record_failure NAIVE_PUBLIC_HOST
fi

if [[ -x "${caddy_bin}" && -f "${caddy_file}" ]]; then
  caddy_sha="$(sha256sum "${caddy_file}" | awk '{print $1}')"
  echo "CADDYFILE_SHA256=${caddy_sha}"
  echo "CADDY_BINARY_SHA256=$(sha256sum "${caddy_bin}" | awk '{print $1}')"
  echo "CADDY_VERSION=$(${caddy_bin} version 2>&1 | head -n1)"
  if "${caddy_bin}" validate --config "${caddy_file}" --adapter caddyfile >/dev/null; then record_pass CADDY_VALIDATE; else record_failure CADDY_VALIDATE; fi
  if "${caddy_bin}" list-modules 2>/dev/null | grep -Fx 'http.handlers.forward_proxy' >/dev/null; then record_pass FORWARD_PROXY_MODULE; else record_failure FORWARD_PROXY_MODULE; fi
else
  echo 'CADDYFILE_SHA256=unavailable'
  record_failure CADDY_BASELINE
fi

caddy_active="$(systemctl show caddy-naive.service --property=ActiveState --value 2>/dev/null || true)"
caddy_pid="$(systemctl show caddy-naive.service --property=MainPID --value 2>/dev/null || true)"
caddy_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value 2>/dev/null || true)"
echo "CADDY_ACTIVE_STATE=${caddy_active:-unknown}"
echo "CADDY_MAINPID=${caddy_pid:-unknown}"
echo "CADDY_NRESTARTS=${caddy_restarts:-unknown}"
[[ "${caddy_active}" == active && "${caddy_pid}" =~ ^[1-9][0-9]*$ && "${caddy_restarts}" =~ ^[0-9]+$ ]] && record_pass CADDY_SERVICE || record_failure CADDY_SERVICE

api_active="$(systemctl show pvnaive-api.service --property=ActiveState --value 2>/dev/null || true)"
echo "API_ACTIVE_STATE=${api_active:-unknown}"
[[ "${api_active}" == active ]] && record_pass API_SERVICE || record_failure API_SERVICE
if curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/api/v1/health/live >/dev/null 2>&1; then record_pass API_LIVE; else record_failure API_LIVE; fi
if curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then record_pass API_READY; else record_failure API_READY; fi

if [[ -S /run/pvnaive/runtime-agent.sock ]]; then echo 'RUNTIME_AGENT_SOCKET=present'; else echo 'RUNTIME_AGENT_SOCKET=absent'; fi
if [[ -e /etc/pvnaive/runtime.key ]]; then
  echo 'RUNTIME_KEY=present'
  stat -c 'RUNTIME_KEY_STAT=mode=%a owner=%U group=%G size=%s' /etc/pvnaive/runtime.key 2>/dev/null || true
else
  echo 'RUNTIME_KEY=absent'
fi

listener_snapshot="$(ss -H -lnt 2>/dev/null || true)"
for port in 22 80 443; do
  if awk -v p=":${port}" '$4 ~ p"$" {found=1} END {exit !found}' <<<"${listener_snapshot}"; then record_pass "TCP_${port}_LISTENER"; else record_failure "TCP_${port}_LISTENER"; fi
done
if awk '$4 == "127.0.0.1:8080" {found=1} END {exit !found}' <<<"${listener_snapshot}"; then record_pass API_LOOPBACK_LISTENER; else record_failure API_LOOPBACK_LISTENER; fi

ssh_active="$(systemctl show ssh.service --property=ActiveState --value 2>/dev/null || true)"
echo "SSH_ACTIVE_STATE=${ssh_active:-unknown}"
[[ "${ssh_active}" == active ]] && record_pass SSH_SERVICE || record_failure SSH_SERVICE

panel_code="$(curl --silent --show-error --max-time 8 --resolve "${panel_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${panel_host}/panel/" 2>/dev/null || true)"
camouflage_code="$(curl --silent --show-error --max-time 8 --resolve "${panel_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${panel_host}/" 2>/dev/null || true)"
echo "PANEL_HTTP=${panel_code:-unavailable}"
echo "CAMOUFLAGE_ROOT_HTTP=${camouflage_code:-unavailable}"
[[ "${panel_code}" == 200 ]] && record_pass PANEL_HTTPS || record_failure PANEL_HTTPS
[[ "${camouflage_code}" == 200 ]] && record_pass CAMOUFLAGE_HTTPS || record_failure CAMOUFLAGE_HTTPS

if [[ -r "${db_env}" && -x "$(command -v psql 2>/dev/null || true)" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${db_env}"
  set +a
  expected_schema="${PVNAIVE_EXPECTED_SCHEMA_VERSION:-}"
  db_schema="$(runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql --port "${PVNAIVE_DB_PORT:-5432}" --username postgres --dbname pvnaive \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations' 2>/dev/null || true)"
  echo "DB_EXPECTED_SCHEMA_VERSION=${expected_schema:-unknown}"
  echo "DB_SCHEMA_VERSION=${db_schema:-unavailable}"
  if [[ "${db_schema}" =~ ^[2-6]$ ]]; then record_pass DB_SCHEMA_SUPPORTED; else record_failure DB_SCHEMA_SUPPORTED; fi
  if [[ -n "${expected_schema}" && "${expected_schema}" == "${db_schema}" ]]; then record_pass DB_SCHEMA_EXPECTATION; else record_failure DB_SCHEMA_EXPECTATION; fi
else
  echo 'DB_SCHEMA_VERSION=unavailable'
  record_failure DB_BASELINE
fi

if [[ -r "${api_env}" ]]; then
  configured_host="$(awk -F= '$1=="PVNAIVE_NAIVE_PUBLIC_HOST" {print $2}' "${api_env}" | tail -n1)"
  echo "API_ENV_NAIVE_PUBLIC_HOST=${configured_host:-unset}"
else
  echo 'API_ENV_NAIVE_PUBLIC_HOST=absent'
fi

if ((${#failures[@]} > 0)); then
  printf 'PREFLIGHT_FAILURE=%s\n' "${failures[*]}"
  echo 'PREFLIGHT_RESULT=FAIL'
  exit 1
fi

echo 'PREFLIGHT_RESULT=PASS'
