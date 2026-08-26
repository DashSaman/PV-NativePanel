#!/usr/bin/env bash
set -Eeuo pipefail

# PVNaive read-only preflight for testAmir5-3. It changes nothing.
if [[ ${EUID} -ne 0 ]]; then
  echo "Run as root: sudo bash scripts/preflight-testamir5-3.sh" >&2
  exit 1
fi

echo "PVNaive preflight (read-only)"
echo "UTC: $(date -u +%FT%TZ)"
echo "Host: $(hostname)"
echo "OS: $(. /etc/os-release && printf '%s %s' "$NAME" "$VERSION_ID")"
echo "Kernel: $(uname -r)"
echo
echo "Expected endpoint DNS:"
getent ahostsv4 namir.softarg.ir || true
echo
echo "Listening TCP sockets (22/80/443/8080/8443):"
ss -lntp | awk 'NR==1 || $4 ~ /:(22|80|443|8080|8443)$/'
echo
echo "Caddy binary/modules:"
command -v caddy || true
caddy version 2>/dev/null || true
caddy list-modules 2>/dev/null | grep -E 'forward_proxy|http.handlers.file_server' || true
echo
echo "Services:"
systemctl --no-pager --full status caddy-naive.service 2>/dev/null | sed -n '1,22p' || true
systemctl --no-pager --full status pvnaive.service 2>/dev/null | sed -n '1,12p' || true
echo
echo "Caddy config validation:"
if [[ -f /etc/caddy/Caddyfile ]]; then
  caddy validate --config /etc/caddy/Caddyfile 2>&1 || true
  sha256sum /etc/caddy/Caddyfile
  stat -c '%a %U:%G %n' /etc/caddy/Caddyfile
else
  echo "/etc/caddy/Caddyfile missing"
fi
echo
echo "Capacity:"
df -h /
free -h
echo
echo "Firewall summary:"
command -v ufw >/dev/null && ufw status verbose || true
command -v nft >/dev/null && nft list ruleset 2>/dev/null | sed -n '1,100p' || true
echo
echo "Recent Caddy errors (credentials are not printed from config):"
journalctl -u caddy-naive.service -p warning --since '-24 hours' --no-pager | tail -80 || true
