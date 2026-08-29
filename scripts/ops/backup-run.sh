#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: backup-run must run as root' >&2; exit 1; }
for cmd in age tar sha256sum find sort awk stat runuser; do command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }; done

repo_db="${PVNAIVE_DB_SCRIPT_ROOT:-/opt/pvnaive/db/current/scripts/db}"
backup_script="${repo_db}/backup.sh"
[[ -x "$backup_script" || -f "$backup_script" ]] || { echo "ERROR: database backup script missing: $backup_script" >&2; exit 1; }
recipient_file="${PVNAIVE_BACKUP_RECIPIENT_FILE:-/etc/pvnaive/backup.recipient}"
[[ -r "$recipient_file" ]] || { echo 'ERROR: age recipient unavailable' >&2; exit 1; }
recipient="$(tr -d '[:space:]' < "$recipient_file")"
[[ "$recipient" == age1* ]] || { echo 'ERROR: invalid age recipient' >&2; exit 1; }

root="${PVNAIVE_SCHEDULED_BACKUP_ROOT:-/var/backups/pvnaive/scheduled}"
retention_days="${PVNAIVE_BACKUP_RETENTION_DAYS:-14}"
[[ "$root" == /var/backups/pvnaive/* ]] || { echo 'ERROR: unsafe backup root' >&2; exit 1; }
[[ "$retention_days" =~ ^[1-9][0-9]*$ && "$retention_days" -le 365 ]] || { echo 'ERROR: retention must be 1..365 days' >&2; exit 1; }

stamp="$(date -u +%Y%m%dT%H%M%SZ)"
dir="${root}/${stamp}"
install -d -m 0700 "$dir"
cleanup(){ code=$?; if ((code != 0)); then rm -rf -- "$dir"; fi; exit "$code"; }
trap cleanup EXIT HUP INT TERM

# Sensitive configuration is never persisted as plaintext by this workflow.
# tar writes directly into age. Files absent on older installations are ignored.
tar_args=()
for path in \
  /etc/pvnaive \
  /etc/caddy/Caddyfile \
  /etc/systemd/system/pvnaive-api.service \
  /etc/systemd/system/pvnaive-runtime-agent.service \
  /etc/systemd/system/pvnaive-backup.service \
  /etc/systemd/system/pvnaive-backup.timer; do
  [[ -e "$path" ]] && tar_args+=("${path#/}")
done
( cd / && tar --numeric-owner --acls --xattrs -cf - "${tar_args[@]}" ) |
  age --recipient "$recipient" --output "${dir}/config.tar.age"

set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
db_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_BACKUP_ROOT=/var/backups/pvnaive/database \
  bash "$backup_script"
)"
db_path="$(awk -F= '$1=="PVNAIVE_BACKUP_PATH"{print $2}' <<<"$db_output")"
[[ -f "$db_path" ]] || { echo 'ERROR: encrypted DB backup did not produce an archive' >&2; exit 1; }

db_metadata="$(dirname -- "$db_path")/metadata.json"
[[ -r "$db_metadata" ]] || { echo 'ERROR: DB backup metadata missing' >&2; exit 1; }
printf '{\n  "product":"PVNaive",\n  "created_at_utc":"%s",\n  "database_backup":"%s",\n  "config_encrypted":true,\n  "retention_days":%s\n}\n' "$stamp" "$db_path" "$retention_days" >"${dir}/metadata.json"
(
  cd "$dir"
  sha256sum config.tar.age metadata.json > SHA256SUMS
  sha256sum --check --strict SHA256SUMS >/dev/null
)
chmod 0600 "$dir"/*

# Retention is constrained to known PVNaive backup roots and completed timestamp directories.
find "$root" -mindepth 1 -maxdepth 1 -type d -name '20????????T??????Z' -mtime "+$retention_days" -print0 | xargs -0r rm -rf --
find /var/backups/pvnaive/database -mindepth 1 -maxdepth 1 -type d -name '20????????T??????Z-*' -mtime "+$retention_days" -print0 | xargs -0r rm -rf --

trap - EXIT HUP INT TERM
echo 'PVNAIVE_SCHEDULED_BACKUP_RESULT=PASSED'
echo "PVNAIVE_CONFIG_BACKUP=${dir}/config.tar.age"
echo "PVNAIVE_DB_BACKUP=${db_path}"
