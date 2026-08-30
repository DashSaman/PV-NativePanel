#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: backup-run must run as root' >&2; exit 1; }
for cmd in age tar sha256sum awk; do
  command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }
done

repo_db="${PVNAIVE_DB_SCRIPT_ROOT:-/opt/pvnaive/db/current/scripts/db}"
backup_script="${repo_db}/backup.sh"
[[ -f "$backup_script" ]] || { echo "ERROR: database backup script missing: $backup_script" >&2; exit 1; }
recipient_file="${PVNAIVE_BACKUP_RECIPIENT_FILE:-/etc/pvnaive/backup.recipient}"
[[ -r "$recipient_file" ]] || { echo 'ERROR: age recipient unavailable' >&2; exit 1; }
recipient="$(tr -d '[:space:]' < "$recipient_file")"
[[ "$recipient" == age1* ]] || { echo 'ERROR: invalid age recipient' >&2; exit 1; }

root="${PVNAIVE_SCHEDULED_BACKUP_ROOT:-/var/backups/pvnaive/scheduled}"
db_root="${PVNAIVE_BACKUP_ROOT:-/var/backups/pvnaive/database}"
[[ "$root" == /var/backups/pvnaive/* ]] || { echo 'ERROR: unsafe scheduled backup root' >&2; exit 1; }
[[ "$db_root" == /var/backups/pvnaive/* ]] || { echo 'ERROR: unsafe database backup root' >&2; exit 1; }

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
install -d -m 0700 "$root" "$db_root"
dir="${root}/${stamp}-${BASHPID}"
[[ ! -e "$dir" ]] || { echo 'ERROR: scheduled backup destination already exists' >&2; exit 1; }
install -d -m 0700 "$dir"

# Stream known configuration directly into age. The legacy
# pvnaive-runtime.conf path is included if present for compatibility, while the
# current shared runtime namespace is normally /etc/tmpfiles.d/pvnaive.conf.
tar_inputs=()
for path in \
  /etc/pvnaive \
  /etc/caddy/Caddyfile \
  /etc/systemd/system/pvnaive-api.service \
  /etc/systemd/system/pvnaive-runtime-agent.service \
  /etc/systemd/system/pvnaive-telemetry-agent.service \
  /etc/systemd/system/pvnaive-backup.service \
  /etc/systemd/system/pvnaive-backup.timer \
  /etc/systemd/system/pvnaive-restore-drill.service \
  /etc/systemd/system/pvnaive-restore-drill.timer \
  /etc/tmpfiles.d/pvnaive.conf \
  /etc/tmpfiles.d/pvnaive-runtime.conf; do
  [[ -e "$path" ]] && tar_inputs+=("${path#/}")
done
(( ${#tar_inputs[@]} > 0 )) || { echo 'ERROR: no PVNaive configuration inputs found' >&2; exit 1; }
( cd / && tar --numeric-owner --acls --xattrs -cf - "${tar_inputs[@]}" ) |
  age --recipient "$recipient" --output "${dir}/config.tar.age"

# Database backup uses PostgreSQL's local Unix socket and the postgres OS role;
# no DB password is read or copied into this scheduled-backup wrapper.
db_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT:-5432}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_DB_CONNECT_TIMEOUT=5 \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_BACKUP_ROOT="$db_root" \
  bash "$backup_script"
)"
db_path="$(awk -F= '$1=="PVNAIVE_BACKUP_PATH"{print substr($0,index($0,"=")+1)}' <<<"$db_output")"
[[ -n "$db_path" && -f "$db_path" ]] || { echo 'ERROR: encrypted DB backup did not produce an archive' >&2; exit 1; }

printf '{\n  "product":"PVNaive",\n  "created_at_utc":"%s",\n  "database_backup":"%s",\n  "config_encrypted":true,\n  "database_encrypted":true\n}\n' "$stamp" "$db_path" >"${dir}/metadata.json"
(
  cd "$dir"
  sha256sum config.tar.age metadata.json > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)
chmod 0600 "$dir"/*

echo 'PVNAIVE_SCHEDULED_BACKUP_RESULT=PASSED'
echo "PVNAIVE_CONFIG_BACKUP=${dir}/config.tar.age"
echo "PVNAIVE_DB_BACKUP=${db_path}"
