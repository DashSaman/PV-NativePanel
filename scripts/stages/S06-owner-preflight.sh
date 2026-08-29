#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S06-OWNER-PREFLIGHT"
panel_host="${PVNAIVE_PUBLIC_HOST:-namir.softarg.ir}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
db_env="/etc/pvnaive/db.env"
api_env="/etc/pvnaive/api.env"
runtime_key="/etc/pvnaive/runtime.key"
runtime_socket="/run/pvnaive/runtime-agent.sock"
failures=()

pass() { printf 'CHECK_%s=PASS\n' "$1"; }
failcheck() { failures+=("$1"); printf 'CHECK_%s=FAIL\n' "$1"; }

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: run S06 preflight as root' >&2; exit 1; }

echo "=== ${stage_id} ==="
echo "UTC=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "HOSTNAME=$(hostname)"

if [[ -x "${caddy_bin}" && -f "${caddy_file}" ]]; then
  caddy_sha="$(sha256sum "${caddy_file}" | awk '{print $1}')"
  echo "CADDYFILE_SHA256=${caddy_sha}"
  echo "CADDY_VERSION=$(${caddy_bin} version 2>&1 | head -n1)"
  "${caddy_bin}" validate --config "${caddy_file}" --adapter caddyfile >/dev/null && pass CADDY_VALIDATE || failcheck CADDY_VALIDATE
  "${caddy_bin}" list-modules 2>/dev/null | grep -Fx 'http.handlers.forward_proxy' >/dev/null && pass FORWARD_PROXY_MODULE || failcheck FORWARD_PROXY_MODULE
else
  echo 'CADDYFILE_SHA256=unavailable'
  failcheck CADDY_BASELINE
fi

caddy_state="$(systemctl show caddy-naive.service --property=ActiveState --value 2>/dev/null || true)"
caddy_pid="$(systemctl show caddy-naive.service --property=MainPID --value 2>/dev/null || true)"
caddy_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value 2>/dev/null || true)"
echo "CADDY_ACTIVE_STATE=${caddy_state:-unknown}"
echo "CADDY_MAINPID=${caddy_pid:-unknown}"
echo "CADDY_NRESTARTS=${caddy_restarts:-unknown}"
[[ "${caddy_state}" == active && "${caddy_pid}" =~ ^[1-9][0-9]*$ && "${caddy_restarts}" =~ ^[0-9]+$ ]] && pass CADDY_SERVICE || failcheck CADDY_SERVICE

systemctl is-active --quiet pvnaive-api.service && pass API_SERVICE || failcheck API_SERVICE
curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true' && pass API_READY || failcheck API_READY
systemctl is-active --quiet pvnaive-runtime-agent.service && pass RUNTIME_AGENT_SERVICE || failcheck RUNTIME_AGENT_SERVICE
if [[ -S "${runtime_socket}" ]] && curl --fail --silent --show-error --max-time 5 --unix-socket "${runtime_socket}" http://unix/v1/health 2>/dev/null | grep -q '"status":"ok"'; then
  pass RUNTIME_AGENT_HEALTH
else
  failcheck RUNTIME_AGENT_HEALTH
fi
[[ -f "${runtime_key}" && "$(stat -c '%s' "${runtime_key}" 2>/dev/null || true)" == 32 ]] && pass RUNTIME_KEY || failcheck RUNTIME_KEY

if [[ -r "${db_env}" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${db_env}"
  set +a
  expected="${PVNAIVE_EXPECTED_SCHEMA_VERSION:-}"
  actual="$(runuser -u postgres -- psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 --host /var/run/postgresql --port "${PVNAIVE_DB_PORT:-5432}" --username postgres --dbname pvnaive --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations' 2>/dev/null || true)"
  echo "DB_EXPECTED_SCHEMA_VERSION=${expected:-unknown}"
  echo "DB_SCHEMA_VERSION=${actual:-unavailable}"
  [[ "${actual}" == 7 ]] && pass DB_SCHEMA_BASELINE || failcheck DB_SCHEMA_BASELINE
  [[ "${expected}" == 7 ]] && pass DB_SCHEMA_EXPECTATION || failcheck DB_SCHEMA_EXPECTATION
else
  failcheck DB_ENV
fi

if [[ -r "${api_env}" ]]; then
  naive_host="$(awk -F= '$1=="PVNAIVE_NAIVE_PUBLIC_HOST" {print $2}' "${api_env}" | tail -n1)"
  echo "NAIVE_PUBLIC_HOST=${naive_host:-unset}"
  [[ -n "${naive_host}" ]] && pass NAIVE_PUBLIC_HOST || failcheck NAIVE_PUBLIC_HOST
else
  failcheck API_ENV
fi

for port in 22 80 443; do
  ss -H -lnt | awk -v p=":${port}" '$4 ~ p"$" {found=1} END {exit !found}' && pass "TCP_${port}_LISTENER" || failcheck "TCP_${port}_LISTENER"
done
systemctl is-active --quiet ssh.service && pass SSH_SERVICE || failcheck SSH_SERVICE

panel_code="$(curl --silent --show-error --max-time 8 --resolve "${panel_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${panel_host}/panel/" 2>/dev/null || true)"
root_code="$(curl --silent --show-error --max-time 8 --resolve "${panel_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${panel_host}/" 2>/dev/null || true)"
echo "PANEL_HTTP=${panel_code:-unavailable}"
echo "ROOT_HTTP=${root_code:-unavailable}"
[[ "${panel_code}" == 200 ]] && pass PANEL_HTTPS || failcheck PANEL_HTTPS
[[ "${root_code}" == 200 ]] && pass ROOT_HTTPS || failcheck ROOT_HTTPS

if ((${#failures[@]})); then
  printf 'PREFLIGHT_FAILURE=%s\n' "${failures[*]}"
  echo 'PREFLIGHT_RESULT=FAIL'
  exit 1
fi

echo 'PREFLIGHT_RESULT=PASS'
echo 'S06_BASE_SCHEMA=7'
echo 'S06_TARGET_SCHEMA=8'
