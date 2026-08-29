#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

[[ ${EUID} -eq 0 ]] || { echo 'ERROR: restore drill must run as root' >&2; exit 1; }
restore_script="${PVNAIVE_DB_SCRIPT_ROOT:-/opt/pvnaive/db/current/scripts/db}/restore.sh"
[[ -f "$restore_script" ]] || { echo "ERROR: restore script missing: $restore_script" >&2; exit 1; }
root="${PVNAIVE_BACKUP_ROOT:-/var/backups/pvnaive/database}"
latest="$(find "$root" -mindepth 2 -maxdepth 2 -type f -name pvnaive.dump.age -printf '%T@ %p\n' | sort -nr | head -n1 | cut -d' ' -f2-)"
[[ -n "$latest" && -f "$latest" ]] || { echo 'ERROR: no encrypted database backup available for restore drill' >&2; exit 1; }

set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
target="pvnaive_restore_test_$(date -u +%Y%m%d_%H%M%S)_$RANDOM"
output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_RESTORE_BACKUP="$latest" \
  PVNAIVE_RESTORE_TARGET_DB="$target" \
  bash "$restore_script"
)"
printf '%s\n' "$output"
grep -q '^PVNAIVE_RESTORE_RESULT=PASSED$' <<<"$output" || { echo 'ERROR: disposable restore drill failed' >&2; exit 1; }
# restore.sh intentionally leaves the verified test DB for caller-side inspection.
runuser -u postgres -- dropdb --if-exists --force "$target"
echo 'PVNAIVE_RESTORE_DRILL_RESULT=PASSED'
