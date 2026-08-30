#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: restore drill must run as root' >&2; exit 1; }
for cmd in find sort runuser dropdb; do
  command -v "$cmd" >/dev/null || { echo "ERROR: missing $cmd" >&2; exit 1; }
done

restore_script="${PVNAIVE_DB_SCRIPT_ROOT:-/opt/pvnaive/db/current/scripts/db}/restore.sh"
[[ -f "$restore_script" ]] || { echo "ERROR: restore script missing: $restore_script" >&2; exit 1; }
root="${PVNAIVE_BACKUP_ROOT:-/var/backups/pvnaive/database}"
[[ "$root" == /var/backups/pvnaive/* ]] || { echo 'ERROR: unsafe backup root' >&2; exit 1; }
latest="$(find "$root" -mindepth 2 -maxdepth 2 -type f -name pvnaive.dump.age -printf '%T@ %p\n' | sort -nr | head -n1 | cut -d' ' -f2-)"
[[ -n "$latest" && -f "$latest" ]] || { echo 'ERROR: no encrypted database backup available for restore drill' >&2; exit 1; }

target="pvnaive_restore_test_$(date -u +%Y%m%d_%H%M%S)_${RANDOM}"
[[ "$target" =~ ^pvnaive_restore_test_[0-9]{8}_[0-9]{6}_[0-9]+$ ]] || { echo 'ERROR: unsafe restore target' >&2; exit 1; }

output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT:-5432}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_DB_CONNECT_TIMEOUT=5 \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_RESTORE_BACKUP="$latest" \
  PVNAIVE_RESTORE_TARGET_DB="$target" \
  bash "$restore_script"
)"
printf '%s\n' "$output"
grep -q '^PVNAIVE_RESTORE_RESULT=PASSED$' <<<"$output" || { echo 'ERROR: disposable restore drill failed' >&2; exit 1; }

# Only the generated pvnaive_restore_test_* database is removed. Production
# database name `pvnaive` is never accepted as the drill target.
runuser -u postgres -- dropdb --if-exists --force "$target"
echo 'PVNAIVE_RESTORE_DRILL_RESULT=PASSED'
