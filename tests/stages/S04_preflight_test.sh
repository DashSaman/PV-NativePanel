#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
stage="${repo_root}/scripts/stages/S04-auth.sh"
unit="${repo_root}/ops/systemd/pvnaive-api.service"
main="${repo_root}/cmd/pvnaive/main.go"

[[ -f "${stage}" ]] || { echo "ERROR: missing S04 stage" >&2; exit 1; }
[[ -f "${unit}" ]] || { echo "ERROR: missing API systemd unit" >&2; exit 1; }
[[ -f "${main}" ]] || { echo "ERROR: missing API main" >&2; exit 1; }

for needle in \
  'stage_id="S04-AUTH"' \
  'expected_host="testAmir5-3"' \
  'S03_DATABASE.json' \
  '101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1' \
  '/etc/pvnaive/auth.key' \
  '/opt/pvnaive/S04_AUTH.json' \
  '127.0.0.1:8080' \
  'CADDY_ACTION=none' \
  'SSH_ACTION=none' \
  'FIREWALL_ACTION=none'; do
  grep -Fq "${needle}" "${stage}" || { echo "ERROR: S04 stage missing contract: ${needle}" >&2; exit 1; }
done

for needle in \
  'User=pvnaive' \
  'Group=pvnaive' \
  'EnvironmentFile=/etc/pvnaive/db.env' \
  'PVNAIVE_AUTH_KEY_FILE=/etc/pvnaive/auth.key' \
  'PVNAIVE_LISTEN=127.0.0.1:8080' \
  'ExecStart=/opt/pvnaive/bin/pvnaive' \
  'NoNewPrivileges=true' \
  'PrivateTmp=true' \
  'ProtectSystem=strict' \
  'ProtectHome=true' \
  'UMask=0077'; do
  grep -Fq "${needle}" "${unit}" || { echo "ERROR: systemd unit missing hardening: ${needle}" >&2; exit 1; }
done

if grep -Eiq '(ufw|iptables|nft[[:space:]]|systemctl[[:space:]]+(restart|reload)[[:space:]]+caddy|caddy[[:space:]]+reload)' "${stage}"; then
  echo "ERROR: S04 localhost stage mutates firewall or Caddy" >&2
  exit 1
fi

grep -Fq '127.0.0.1' "${main}" || { echo "ERROR: API main does not enforce loopback" >&2; exit 1; }
grep -Fq 'PVNAIVE_AUTH_KEY_FILE' "${main}" || { echo "ERROR: API main does not require auth key" >&2; exit 1; }
grep -Fq 'sql.Open("pgx"' "${main}" || { echo "ERROR: API main does not open PostgreSQL via pgx stdlib" >&2; exit 1; }

echo "S04_PREFLIGHT_TEST=PASSED"
