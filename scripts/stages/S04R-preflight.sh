#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S04R-PREFLIGHT"
public_host="${PVNAIVE_PUBLIC_HOST:-namir.softarg.ir}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
runtime_key="/etc/pvnaive/runtime.key"
db_env="/etc/pvnaive/db.env"
failures=()

record_failure() {
  failures+=("$1")
  printf 'CHECK_%s=FAIL\n' "$1"
}

record_pass() {
  printf 'CHECK_%s=PASS\n' "$1"
}

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: run read-only preflight as root' >&2; exit 1; }

echo "=== ${stage_id} ==="
echo "UTC=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "HOSTNAME=$(hostname)"
echo "KERNEL=$(uname -srmo)"
if [[ -r /etc/os-release ]]; then
  os_pretty="$(awk -F= '$1=="PRETTY_NAME" {gsub(/^"|"$/, "", $2); print $2}' /etc/os-release)"
  echo "OS=${os_pretty:-unknown}"
fi

if [[ -x "${caddy_bin}" && -f "${caddy_file}" ]]; then
  caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
  caddy_binary_sha="$(sha256sum "${caddy_bin}" | awk '{print $1}')"
  caddy_version="$(${caddy_bin} version 2>&1 | head -n1)"
  echo "CADDYFILE_SHA256=${caddy_sha}"
  echo "CADDY_BINARY_SHA256=${caddy_binary_sha}"
  echo "CADDY_VERSION=${caddy_version}"
  if /usr/local/bin/caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/dev/null; then
    record_pass CADDY_VALIDATE
  else
    record_failure CADDY_VALIDATE
  fi
  if "${caddy_bin}" list-modules 2>/dev/null | grep -Fx 'http.handlers.forward_proxy' >/dev/null; then
    record_pass FORWARD_PROXY_MODULE
  else
    record_failure FORWARD_PROXY_MODULE
  fi
else
  record_failure CADDY_BASELINE
fi

echo '--- caddy-naive.service ---'
systemctl show caddy-naive.service --no-pager \
  --property=LoadState,ActiveState,SubState,MainPID,NRestarts,ExecMainStartTimestamp,FragmentPath 2>&1 || true
caddy_active="$(systemctl show caddy-naive.service --property=ActiveState --value 2>/dev/null || true)"
caddy_pid="$(systemctl show caddy-naive.service --property=MainPID --value 2>/dev/null || true)"
caddy_restarts="$(systemctl show caddy-naive.service --property=NRestarts --value 2>/dev/null || true)"
echo "CADDY_ACTIVE_STATE=${caddy_active:-unknown}"
echo "CADDY_MAINPID=${caddy_pid:-unknown}"
echo "CADDY_NRESTARTS=${caddy_restarts:-unknown}"
[[ "${caddy_active}" == "active" && "${caddy_pid}" =~ ^[1-9][0-9]*$ ]] && record_pass CADDY_SERVICE || record_failure CADDY_SERVICE

echo '--- pvnaive-api.service ---'
systemctl show pvnaive-api.service --no-pager \
  --property=LoadState,ActiveState,SubState,MainPID,NRestarts,ExecMainStartTimestamp,FragmentPath 2>&1 || true
api_active="$(systemctl show pvnaive-api.service --property=ActiveState --value 2>/dev/null || true)"
[[ "${api_active}" == "active" ]] && record_pass API_SERVICE || record_failure API_SERVICE

if curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/api/v1/health/live >/dev/null 2>&1; then
  record_pass API_LIVE
else
  record_failure API_LIVE
fi
if curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/api/v1/health/ready 2>/dev/null | grep -q '"ready":true'; then
  record_pass API_READY
else
  record_failure API_READY
fi

echo '--- pvnaive-runtime-agent.service ---'
systemctl show pvnaive-runtime-agent.service --no-pager \
  --property=LoadState,ActiveState,SubState,MainPID,NRestarts,ExecMainStartTimestamp,FragmentPath 2>&1 || true
if [[ -S /run/pvnaive/runtime-agent.sock ]]; then
  echo 'RUNTIME_AGENT_SOCKET=present'
else
  echo 'RUNTIME_AGENT_SOCKET=absent'
fi
if [[ -e "${runtime_key}" ]]; then
  echo 'RUNTIME_KEY=present'
  stat -c 'RUNTIME_KEY_STAT=mode=%a owner=%U group=%G size=%s' "${runtime_key}" 2>/dev/null || true
else
  echo 'RUNTIME_KEY=absent'
fi

echo '--- listeners ---'
ss -H -lntp 2>&1 || true
listener_snapshot="$(ss -H -lnt 2>/dev/null || true)"
for port in 22 80 443; do
  if awk -v p=":${port}" '$4 ~ p"$" {found=1} END {exit !found}' <<<"${listener_snapshot}"; then
    record_pass "TCP_${port}_LISTENER"
  else
    record_failure "TCP_${port}_LISTENER"
  fi
done
if awk '$4 == "127.0.0.1:8080" {found=1} END {exit !found}' <<<"${listener_snapshot}"; then
  record_pass API_LOOPBACK_LISTENER
else
  record_failure API_LOOPBACK_LISTENER
fi

echo '--- ssh.service ---'
systemctl show ssh.service --no-pager \
  --property=LoadState,ActiveState,SubState,MainPID,NRestarts,FragmentPath 2>&1 || true
ssh_active="$(systemctl show ssh.service --property=ActiveState --value 2>/dev/null || true)"
[[ "${ssh_active}" == "active" ]] && record_pass SSH_SERVICE || record_failure SSH_SERVICE

echo '--- public HTTP probes ---'
camouflage_code="$(curl --silent --show-error --max-time 8 --resolve "${public_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${public_host}/" 2>/dev/null || true)"
panel_code="$(curl --silent --show-error --max-time 8 --resolve "${public_host}:443:127.0.0.1" --output /dev/null --write-out '%{http_code}' "https://${public_host}/runtime/naive" 2>/dev/null || true)"
echo "CAMOUFLAGE_ROOT_HTTP=${camouflage_code:-unavailable}"
echo "PANEL_RUNTIME_NAIVE_HTTP=${panel_code:-unavailable}"

echo '--- database schema ---'
if [[ -r "${db_env}" && -x "$(command -v psql 2>/dev/null || true)" ]]; then
  expected_schema="$(awk -F= '$1=="PVNAIVE_EXPECTED_SCHEMA_VERSION" {print $2}' "${db_env}" | tail -n1)"
  echo "DB_EXPECTED_SCHEMA_VERSION=${expected_schema:-unknown}"
  set -a
  # shellcheck disable=SC1091
  source "${db_env}"
  set +a
  db_schema="$(psql --no-psqlrc --tuples-only --no-align --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST:-127.0.0.1}" --port "${PVNAIVE_DB_PORT:-5432}" \
    --username "${PVNAIVE_DB_USER:-pvnaive_app}" --dbname "${PVNAIVE_DB_NAME:-pvnaive}" \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations' 2>/dev/null || true)"
  echo "DB_SCHEMA_VERSION=${db_schema:-unavailable}"
else
  echo 'DB_SCHEMA_VERSION=unavailable'
fi

echo '--- firewall ---'
if command -v ufw >/dev/null 2>&1; then
  ufw status verbose 2>&1 || true
else
  echo 'UFW=not-installed'
fi
if command -v nft >/dev/null 2>&1; then
  nft list ruleset 2>&1 || true
else
  echo 'NFT=not-installed'
fi

if ((${#failures[@]} > 0)); then
  printf 'PREFLIGHT_FAILURE=%s\n' "${failures[@]}"
  echo 'PREFLIGHT_RESULT=FAIL'
  exit 1
fi

echo 'PREFLIGHT_RESULT=PASS'