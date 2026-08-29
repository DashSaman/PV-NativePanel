#!/usr/bin/env bash
set -Eeuo pipefail

target="scripts/stages/S04R-preflight.sh"
[[ -f "${target}" ]] || { echo "ERROR: missing ${target}" >&2; exit 1; }
bash -n "${target}"

for required in \
  'sha256sum /etc/caddy/Caddyfile' \
  '/usr/local/bin/caddy validate' \
  'systemctl show caddy-naive.service' \
  'MainPID' \
  'NRestarts' \
  'ss -H -lntp' \
  'pvnaive-api.service' \
  'ssh.service' \
  'runtime.key' \
  'ufw status verbose' \
  'nft list ruleset' \
  'PREFLIGHT_RESULT=PASS'; do
  grep -Fq -- "${required}" "${target}" || { echo "ERROR: preflight missing contract token: ${required}" >&2; exit 1; }
done

for forbidden in \
  'systemctl restart' \
  'systemctl reload' \
  'systemctl start' \
  'systemctl stop' \
  'systemctl enable' \
  'systemctl disable' \
  'systemctl daemon-reload' \
  'apt ' \
  'apt-get ' \
  'install -m' \
  'mkdir ' \
  'chmod ' \
  'chown ' \
  'cp ' \
  'mv ' \
  'rm ' \
  'tee ' \
  'sed -i' \
  'truncate ' \
  '> /etc/' \
  '>> /etc/'; do
  if grep -Fq -- "${forbidden}" "${target}"; then
    echo "ERROR: read-only preflight contains forbidden mutation token: ${forbidden}" >&2
    exit 1
  fi
done

if grep -Eq 'cat[[:space:]]+/etc/pvnaive/(runtime|auth)\.key' "${target}"; then
  echo 'ERROR: preflight may not print key material' >&2
  exit 1
fi

echo 'S04R_PREFLIGHT_CONTRACT=PASSED'
