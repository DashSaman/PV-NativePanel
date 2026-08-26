#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

if [[ ${EUID} -ne 0 ]]; then
  echo "ERROR: run as root" >&2
  exit 1
fi

stage_id="S02-FOUNDATION"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
base="/opt/pvnaive"
backup_root="/var/backups/pvnaive"
backup_dir="${backup_root}/${stamp}"

echo "=== ${stage_id} ==="
echo "Host: $(hostname)"
echo "UTC:  ${stamp}"

[[ "$(hostname)" == "testAmir5-3" ]] || { echo "ERROR: unexpected host"; exit 1; }
getent ahostsv4 namir.softarg.ir | awk '{print $1}' | grep -qx "91.107.182.147" || {
  echo "ERROR: DNS does not match expected IPv4"
  exit 1
}
[[ -x /usr/local/bin/caddy ]] || { echo "ERROR: Caddy binary missing"; exit 1; }
[[ -f /etc/caddy/Caddyfile ]] || { echo "ERROR: Caddyfile missing"; exit 1; }
systemctl is-active --quiet caddy-naive.service || { echo "ERROR: caddy-naive is not active"; exit 1; }
/usr/local/bin/caddy validate --config /etc/caddy/Caddyfile >/dev/null

install -d -m 0700 "${backup_root}" "${backup_dir}"
cp -a /etc/caddy/Caddyfile "${backup_dir}/Caddyfile"
cp -a /etc/systemd/system/caddy-naive.service "${backup_dir}/caddy-naive.service"
if [[ -d /var/www/naive ]]; then
  tar --xattrs --acls -C /var/www -czf "${backup_dir}/public-site.tar.gz" naive
fi
sha256sum "${backup_dir}"/* > "${backup_dir}/SHA256SUMS"
chmod -R go-rwx "${backup_dir}"

if ! getent group pvnaive >/dev/null; then
  groupadd --system pvnaive
fi
if ! id pvnaive >/dev/null 2>&1; then
  useradd --system --gid pvnaive --home-dir "${base}" --shell /usr/sbin/nologin pvnaive
fi

install -d -o root -g pvnaive -m 0750 "${base}"
install -d -o pvnaive -g pvnaive -m 0750 "${base}/bin" "${base}/web"
install -d -o pvnaive -g pvnaive -m 0700 "${base}/data" "${base}/secrets"
install -d -o pvnaive -g pvnaive -m 0750 /var/lib/pvnaive /var/log/pvnaive
install -d -o root -g pvnaive -m 0750 /etc/pvnaive

cat > "${base}/FOUNDATION.json" <<EOF
{
  "stage": "${stage_id}",
  "created_at_utc": "${stamp}",
  "host": "$(hostname)",
  "domain": "namir.softarg.ir",
  "caddyfile_sha256": "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')",
  "backup_dir": "${backup_dir}",
  "caddy_restarted": false,
  "firewall_changed": false,
  "ssh_changed": false,
  "packages_installed": false
}
EOF
chown root:pvnaive "${base}/FOUNDATION.json"
chmod 0640 "${base}/FOUNDATION.json"

echo
echo "=== Verification ==="
id pvnaive
stat -c '%a %U:%G %n' "${base}" "${base}/data" "${base}/secrets" /etc/pvnaive
sha256sum -c "${backup_dir}/SHA256SUMS"
systemctl is-active caddy-naive.service
ss -lntp | awk 'NR==1 || $4 ~ /:(22|80|443)$/'
echo
echo "S02_RESULT=PASSED"
echo "BACKUP_DIR=${backup_dir}"
echo "Paste the complete output back into the chat."
