#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S03-DATABASE"
stage_root_bashpid="${BASHPID}"
expected_host="testAmir5-3"
expected_domain="namir.softarg.ir"
expected_ipv4="91.107.182.147"
expected_caddy_sha256="101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
# shellcheck source=scripts/stages/lib.sh
source "${script_dir}/lib.sh"
backup_root="/var/backups/pvnaive"
rollback_backup_dir=""
cluster_version=""
cluster_name=""
cluster_port=""
hba_file=""
auto_conf_file=""
release_dir=""
release_link="/opt/pvnaive/db/current"
database_created=0
roles_created=0
config_changed=0
release_created=0
secrets_created=0
units_installed=0
timer_enabled=0
restore_test_db=""
marker_temp=""
marker_owned_by_run=0
cluster_owner_marker="/var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER"
cluster_owner_temp=""

postgres_psql() {
  runuser -u postgres -- psql --no-psqlrc --set ON_ERROR_STOP=1 --host /var/run/postgresql --port "${cluster_port}" --username postgres "$@"
}

verify_caddy_invariants() {
  local current_sha
  local listener_snapshot
  current_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
  [[ "${current_sha}" == "${expected_caddy_sha256}" ]] || die "Caddyfile checksum changed: ${current_sha}"
  systemctl is-active --quiet caddy-naive.service || die "caddy-naive.service is not active"
  /usr/local/bin/caddy validate --config /etc/caddy/Caddyfile >/dev/null
  listener_snapshot="$(pvnaive_tcp_listener_snapshot)"
  for required_port in 22 80 443; do
    printf '%s\n' "${listener_snapshot}" | pvnaive_tcp_port_is_listening "${required_port}" || \
      die "required TCP port ${required_port} is not listening"
  done
}

rollback_on_error() {
  local line="$1"
  local code="$2"
  if ! pvnaive_is_root_bash_process "${stage_root_bashpid}"; then
    trap - ERR
    exit "${code}"
  fi

  trap - ERR HUP INT TERM
  set +e
  local rollback_failures=()
  echo
  echo "S03_RESULT=FAILED"
  echo "FAILED_LINE=${line}"
  echo "FAILED_EXIT=${code}"
  echo "ROLLBACK=STARTED"

  if [[ "${timer_enabled}" == "1" ]]; then
    if ! systemctl disable --now pvnaive-db-health.timer; then
      rollback_failures+=("disable-health-timer")
    fi
  fi
  if [[ "${units_installed}" == "1" ]]; then
    if ! rm -f /etc/systemd/system/pvnaive-db-health.service /etc/systemd/system/pvnaive-db-health.timer; then
      rollback_failures+=("remove-health-units")
    fi
    if ! systemctl daemon-reload; then
      rollback_failures+=("systemd-daemon-reload")
    fi
  fi
  if [[ "${config_changed}" == "1" && -n "${rollback_backup_dir}" ]]; then
    if ! cp -a "${rollback_backup_dir}/pg_hba.conf" "${hba_file}"; then
      rollback_failures+=("restore-pg-hba")
    fi
    if [[ -f "${rollback_backup_dir}/postgresql.auto.conf" ]]; then
      if ! cp -a "${rollback_backup_dir}/postgresql.auto.conf" "${auto_conf_file}"; then
        rollback_failures+=("restore-postgresql-auto-conf")
      fi
    else
      if ! rm -f -- "${auto_conf_file}"; then
        rollback_failures+=("remove-postgresql-auto-conf")
      fi
    fi
    if ! pg_ctlcluster "${cluster_version}" "${cluster_name}" restart; then
      rollback_failures+=("restart-postgresql-after-config-restore")
    fi
  fi
  if [[ "${database_created}" == "1" && -n "${cluster_port}" ]]; then
    if [[ -n "${restore_test_db}" ]]; then
      if ! postgres_psql --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${restore_test_db}' AND pid <> pg_backend_pid()" >/dev/null; then
        rollback_failures+=("terminate-restore-test-connections")
      fi
      if ! runuser -u postgres -- dropdb --if-exists --host /var/run/postgresql --port "${cluster_port}" "${restore_test_db}"; then
        rollback_failures+=("drop-restore-test-db")
      fi
    fi
    if ! postgres_psql --dbname postgres --command "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'pvnaive' AND pid <> pg_backend_pid()" >/dev/null; then
      rollback_failures+=("terminate-pvnaive-connections")
    fi
    if ! runuser -u postgres -- dropdb --if-exists --host /var/run/postgresql --port "${cluster_port}" pvnaive; then
      rollback_failures+=("drop-pvnaive-db")
    fi
  fi
  if [[ "${roles_created}" == "1" && -n "${cluster_port}" ]]; then
    if ! postgres_psql --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null; then
      rollback_failures+=("drop-pvnaive-roles")
    fi
  fi
  if [[ "${secrets_created}" == "1" ]]; then
    if ! rm -f /etc/pvnaive/db.env /etc/pvnaive/db.pgpass /etc/pvnaive/backup.agekey /etc/pvnaive/backup.recipient; then
      rollback_failures+=("remove-database-secrets")
    fi
  fi
  if [[ "${release_created}" == "1" ]]; then
    if ! rm -f -- "${release_link}.new"; then
      rollback_failures+=("remove-release-staging-link")
    fi
    if [[ -L "${release_link}" && "$(readlink -f "${release_link}")" == "${release_dir}" ]]; then
      if ! rm -f -- "${release_link}"; then
        rollback_failures+=("remove-current-release-link")
      fi
    fi
    if ! rm -rf -- "${release_dir}"; then
      rollback_failures+=("remove-database-release")
    fi
  fi
  if [[ "${marker_owned_by_run}" == "1" ]]; then
    if [[ -n "${marker_temp}" ]] && ! rm -f -- "${marker_temp}"; then
      rollback_failures+=("remove-S03-marker-temp")
    fi
    if ! rm -f -- /opt/pvnaive/S03_DATABASE.json; then
      rollback_failures+=("remove-S03-marker")
    fi
  fi
  if [[ -n "${cluster_owner_temp}" ]] && ! rm -f -- "${cluster_owner_temp}"; then
    rollback_failures+=("remove-cluster-owner-marker-temp")
  fi

  systemctl is-active caddy-naive.service 2>/dev/null || true
  if ((${#rollback_failures[@]} == 0)); then
    echo "ROLLBACK=COMPLETED"
  else
    echo "ROLLBACK=FAILED"
    printf 'ROLLBACK_FAILED_STEP=%s\n' "${rollback_failures[@]}"
  fi
  echo "PACKAGES_REMOVED=false"
  echo "CADDY_RESTARTED=false"
  echo "SSH_CHANGED=false"
  echo "FIREWALL_CHANGED=false"
  echo "Inspect ${rollback_backup_dir:-no-backup-created} before retrying."
  exit "${code}"
}
trap 'rollback_on_error "${LINENO}" "$?"' ERR
trap 'rollback_on_error "${LINENO}" 129' HUP
trap 'rollback_on_error "${LINENO}" 130' INT
trap 'rollback_on_error "${LINENO}" 143' TERM

[[ ${EUID} -eq 0 ]] || die "run as root"
echo "=== ${stage_id} ==="
echo "Host: $(hostname)"
echo "UTC:  ${stamp}"
echo "ROLLBACK_PLAN=On any error, remove the new pvnaive DB/roles/secrets/units, restore PostgreSQL config, restart only PostgreSQL, and leave pinned packages installed for inspection."
echo "ROLLBACK_PLAN_DATA=An encrypted DB backup and a restore drill are mandatory before PASSED."
echo "ROLLBACK_PLAN_CLUSTER=The dedicated PostgreSQL cluster and pinned packages remain for retry; their ownership marker is retained."
echo "CADDY_ACTION=none"
echo "SSH_ACTION=none"
echo "FIREWALL_ACTION=none"

[[ "$(hostname)" == "${expected_host}" ]] || die "unexpected host"
getent ahostsv4 "${expected_domain}" | awk '{print $1}' | grep -qx "${expected_ipv4}" || die "DNS mismatch"
[[ -f /opt/pvnaive/FOUNDATION.json ]] || die "S02 foundation marker is missing"
[[ -x /usr/local/bin/caddy ]] || die "Caddy binary is missing"
[[ -f /etc/caddy/Caddyfile ]] || die "Caddyfile is missing"
verify_caddy_invariants

for required_source in \
  db/migrations/0001_initial.up.sql \
  db/migrations/0001_initial.down.sql \
  db/migrations/SHA256SUMS \
  scripts/db/lib.sh scripts/db/migrate.sh scripts/db/rollback.sh \
  scripts/db/backup.sh scripts/db/restore.sh scripts/db/health.sh \
  scripts/stages/lib.sh \
  ops/systemd/pvnaive-db-health.service ops/systemd/pvnaive-db-health.timer; do
  [[ -f "${repo_root}/${required_source}" ]] || die "bundle file missing: ${required_source}"
done
(
  cd "${repo_root}/db/migrations"
  sha256sum --check --strict SHA256SUMS
) >/dev/null || die "bundled migration checksums failed"
bash -n "${repo_root}"/scripts/db/*.sh "${repo_root}"/scripts/stages/*.sh

if [[ -f /opt/pvnaive/S03_DATABASE.json ]]; then
  marker_file="/opt/pvnaive/S03_DATABASE.json"
  grep -Fqx '  "stage": "S03-DATABASE",' "${marker_file}" || die "S03 marker stage is invalid"
  grep -Fqx '  "host": "testAmir5-3",' "${marker_file}" || die "S03 marker host is invalid"
  grep -Fqx '  "schema_version": 1,' "${marker_file}" || die "S03 marker schema version is invalid"
  grep -Fqx '  "restore_drill": "passed",' "${marker_file}" || die "S03 marker restore result is invalid"
  grep -Fqx '  "health_check_result": "success",' "${marker_file}" || die "S03 marker health result is invalid"
  grep -Fqx "  \"caddyfile_sha256\": \"${expected_caddy_sha256}\"," "${marker_file}" || die "S03 marker Caddy checksum is invalid"
  grep -Fqx '  "caddy_restarted": false,' "${marker_file}" || die "S03 marker Caddy action is invalid"
  grep -Fqx '  "ssh_changed": false,' "${marker_file}" || die "S03 marker SSH action is invalid"
  grep -Fqx '  "firewall_changed": false' "${marker_file}" || die "S03 marker firewall action is invalid"
  existing_backup="$(sed -n 's/^  "encrypted_backup": "\(.*\)",$/\1/p' "${marker_file}")"
  existing_prechange_backup="$(sed -n 's/^  "prechange_backup": "\(.*\)",$/\1/p' "${marker_file}")"
  [[ "${existing_backup}" =~ ^/var/backups/pvnaive/database/[0-9]{8}T[0-9]{6}Z/pvnaive\.dump\.age$ ]] || die "S03 marker database backup path is invalid"
  [[ "${existing_prechange_backup}" =~ ^/var/backups/pvnaive/[0-9]{8}T[0-9]{6}Z-S03-pre$ ]] || die "S03 marker prechange backup path is invalid"
  [[ -f "${existing_backup}" ]] || die "S03 encrypted database backup is missing"
  [[ -f "$(dirname -- "${existing_backup}")/SHA256SUMS" ]] || die "S03 database backup manifest is missing"
  (cd "$(dirname -- "${existing_backup}")" && sha256sum --check --strict SHA256SUMS) >/dev/null || die "S03 database backup checksum failed"
  [[ -f "${existing_prechange_backup}/SHA256SUMS" ]] || die "S03 prechange backup manifest is missing"
  (cd "${existing_prechange_backup}" && sha256sum --check --strict SHA256SUMS) >/dev/null || die "S03 prechange backup checksum failed"
  [[ -r /etc/pvnaive/db.env ]] || die "S03 marker exists but db.env is missing"
  [[ -x /opt/pvnaive/db/current/scripts/db/health.sh ]] || die "S03 current database release is missing"
  set -a
  # shellcheck disable=SC1091
  source /etc/pvnaive/db.env
  set +a
  /opt/pvnaive/db/current/scripts/db/health.sh
  systemctl is-active --quiet pvnaive-db-health.timer || die "S03 health timer is not active"
  systemctl start pvnaive-db-health.service
  existing_health_result="$(systemctl show --property=Result --value pvnaive-db-health.service)"
  [[ "${existing_health_result}" == "success" ]] || die "S03 health service result is ${existing_health_result}"
  verify_caddy_invariants
  trap - ERR HUP INT TERM
  echo "S03_RESULT=PASSED"
  echo "S03_MODE=VERIFIED_EXISTING"
  exit 0
fi

available_kib="$(awk '/MemAvailable:/ {print $2}' /proc/meminfo)"
((available_kib >= 1048576)) || die "less than 1 GiB memory is available"
available_blocks="$(df -Pk /var | awk 'NR==2 {print $4}')"
((available_blocks >= 5242880)) || die "less than 5 GiB is available on /var"
swap_kib="$(awk '/SwapTotal:/ {print $2}' /proc/meminfo)"
echo "MEM_AVAILABLE_KIB=${available_kib}"
echo "DISK_AVAILABLE_KIB=${available_blocks}"
echo "SWAP_TOTAL_KIB=${swap_kib}"
[[ "${swap_kib}" != "0" ]] || echo "WARNING: swap remains disabled; this is recorded but does not block S03."

[[ -r /etc/os-release ]] || die "/etc/os-release is missing"
# shellcheck disable=SC1091
source /etc/os-release
[[ "${ID:-}" == "ubuntu" && "${VERSION_ID:-}" == "26.04" ]] || die "unsupported OS ${ID:-unknown} ${VERSION_ID:-unknown}; expected Ubuntu 26.04"

# S03 owns the first PostgreSQL cluster on this dedicated host. A durable
# provenance marker makes a post-APT retry distinguishable from an unrelated
# pre-existing cluster without silently repurposing operator data.
if [[ -f "${cluster_owner_marker}" ]]; then
  grep -Fqx 'stage=S03-DATABASE' "${cluster_owner_marker}" || die "PostgreSQL cluster owner marker stage is invalid"
  grep -Fqx 'host=testAmir5-3' "${cluster_owner_marker}" || die "PostgreSQL cluster owner marker host is invalid"
  grep -Fqx 'purpose=pvnaive-dedicated-postgresql' "${cluster_owner_marker}" || die "PostgreSQL cluster owner marker purpose is invalid"
else
  if command -v pg_lsclusters >/dev/null 2>&1; then
    preexisting_clusters="$(pg_lsclusters --no-header)"
    [[ -z "${preexisting_clusters//[[:space:]]/}" ]] || die "unowned pre-existing PostgreSQL cluster requires manual inspection"
  fi
  [[ -d /var/lib/pvnaive ]] || die "/var/lib/pvnaive foundation directory is missing"
  cluster_owner_temp="$(mktemp /var/lib/pvnaive/.S03_POSTGRES_CLUSTER_OWNER.XXXXXX)"
  printf '%s\n' \
    'stage=S03-DATABASE' \
    'host=testAmir5-3' \
    'purpose=pvnaive-dedicated-postgresql' > "${cluster_owner_temp}"
  chown root:pvnaive "${cluster_owner_temp}"
  chmod 0640 "${cluster_owner_temp}"
  mv -T -- "${cluster_owner_temp}" "${cluster_owner_marker}"
  cluster_owner_temp=""
fi

export DEBIAN_FRONTEND=noninteractive
export NEEDRESTART_MODE=l
apt-get update
packages=(postgresql postgresql-client postgresql-contrib age)
install_specs=()
for package_name in "${packages[@]}"; do
  candidate="$(apt-cache policy "${package_name}" | awk '/Candidate:/ {print $2}')"
  [[ -n "${candidate}" && "${candidate}" != "(none)" ]] || die "no signed APT candidate for ${package_name}"
  install_specs+=("${package_name}=${candidate}")
done
apt-get install --yes --no-install-recommends "${install_specs[@]}"

for required_command in psql pg_isready pg_dump pg_restore pg_lsclusters pg_ctlcluster pg_createcluster age age-keygen openssl systemd-analyze; do
  command -v "${required_command}" >/dev/null 2>&1 || die "installed command missing: ${required_command}"
done

clusters_output="$(pg_lsclusters --no-header)"
if [[ -z "${clusters_output//[[:space:]]/}" ]]; then
  mapfile -t postgres_versions < <(find /usr/lib/postgresql -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort -V)
  ((${#postgres_versions[@]} == 1)) || die "cannot identify exactly one installed PostgreSQL major version"
  pg_createcluster --start "${postgres_versions[0]}" main
  clusters_output="$(pg_lsclusters --no-header)"
fi
mapfile -t clusters < <(awk '{print $1 "|" $2 "|" $3 "|" $4}' <<< "${clusters_output}")
((${#clusters[@]} == 1)) || die "expected exactly one PostgreSQL cluster, found ${#clusters[@]}"
IFS='|' read -r cluster_version cluster_name cluster_port cluster_status <<< "${clusters[0]}"
[[ "${cluster_port}" == "5432" ]] || die "unexpected PostgreSQL port ${cluster_port}"
if [[ "${cluster_status}" != "online" ]]; then
  pg_ctlcluster "${cluster_version}" "${cluster_name}" start
fi

hba_file="$(postgres_psql --dbname postgres --tuples-only --no-align --command 'SHOW hba_file')"
auto_conf_file="$(postgres_psql --dbname postgres --tuples-only --no-align --command "SELECT current_setting('data_directory') || '/postgresql.auto.conf'")"
[[ -f "${hba_file}" ]] || die "pg_hba.conf not found"
if grep -q '^# BEGIN PVNAIVE S03$' "${hba_file}"; then
  die "PVNaive HBA block already exists without a completed S03 marker"
fi

rollback_backup_dir="${backup_root}/${stamp}-S03-pre"
install -d -m 0700 "${backup_root}" "${rollback_backup_dir}"
cp -a "${hba_file}" "${rollback_backup_dir}/pg_hba.conf"
[[ ! -f "${auto_conf_file}" ]] || cp -a "${auto_conf_file}" "${rollback_backup_dir}/postgresql.auto.conf"
dpkg-query -W -f='${Package}=${Version}\n' "${packages[@]}" > "${rollback_backup_dir}/PACKAGES"
pg_lsclusters > "${rollback_backup_dir}/PG_CLUSTERS"
(
  cd "${rollback_backup_dir}"
  find . -maxdepth 1 -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
)
chmod -R go-rwx "${rollback_backup_dir}"

db_exists="$(postgres_psql --dbname postgres --tuples-only --no-align --command "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'pvnaive')")"
owner_exists="$(postgres_psql --dbname postgres --tuples-only --no-align --command "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'pvnaive_owner')")"
app_exists="$(postgres_psql --dbname postgres --tuples-only --no-align --command "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'pvnaive_app')")"
[[ "${db_exists}|${owner_exists}|${app_exists}" == "f|f|f" ]] || die "pre-existing pvnaive DB or role requires manual inspection"
for new_path in \
  /etc/pvnaive/db.env /etc/pvnaive/db.pgpass \
  /etc/pvnaive/backup.agekey /etc/pvnaive/backup.recipient \
  /etc/systemd/system/pvnaive-db-health.service \
  /etc/systemd/system/pvnaive-db-health.timer \
  "${release_link}" "${release_link}.new"; do
  [[ ! -e "${new_path}" && ! -L "${new_path}" ]] || die "pre-existing S03 artifact requires manual inspection: ${new_path}"
done

hba_temp="$(mktemp)"
{
  echo '# BEGIN PVNAIVE S03'
  echo 'local   pvnaive   pvnaive_app                              scram-sha-256'
  echo 'host    pvnaive   pvnaive_app   127.0.0.1/32               scram-sha-256'
  echo 'host    pvnaive   pvnaive_app   ::1/128                    scram-sha-256'
  echo '# END PVNAIVE S03'
  cat "${hba_file}"
} > "${hba_temp}"
chown --reference="${hba_file}" "${hba_temp}"
chmod --reference="${hba_file}" "${hba_temp}"
config_changed=1
mv -- "${hba_temp}" "${hba_file}"
postgres_psql --dbname postgres <<'SQL' >/dev/null
ALTER SYSTEM SET listen_addresses = '127.0.0.1,::1';
ALTER SYSTEM SET password_encryption = 'scram-sha-256';
SQL
pg_ctlcluster "${cluster_version}" "${cluster_name}" restart
postgres_psql --dbname postgres --command "SELECT pg_reload_conf()" >/dev/null
hba_errors="$(postgres_psql --dbname postgres --tuples-only --no-align --command 'SELECT COUNT(*) FROM pg_hba_file_rules WHERE error IS NOT NULL')"
[[ "${hba_errors}" == "0" ]] || die "pg_hba.conf validation reported ${hba_errors} errors"
postgres_psql --dbname postgres --tuples-only --no-align --command 'SHOW listen_addresses' | grep -qx '127.0.0.1,::1' || die "listen_addresses is not loopback-only"
postgres_listener_snapshot="$(pvnaive_tcp_listener_snapshot)"
if printf '%s\n' "${postgres_listener_snapshot}" | pvnaive_tcp_has_non_loopback_postgres_listener; then
  die "PostgreSQL has a non-loopback listener"
fi

roles_created=1
postgres_psql --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 40;
SQL
app_password="$(openssl rand -hex 32)"
postgres_psql --dbname postgres <<SQL >/dev/null
ALTER ROLE pvnaive_app PASSWORD '${app_password}';
ALTER ROLE pvnaive_app SET statement_timeout = '30s';
ALTER ROLE pvnaive_app SET lock_timeout = '5s';
ALTER ROLE pvnaive_app SET idle_in_transaction_session_timeout = '30s';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
database_created=1
runuser -u postgres -- createdb --host /var/run/postgresql --port "${cluster_port}" --owner pvnaive_owner --encoding UTF8 --template template0 pvnaive
postgres_psql --dbname pvnaive <<'SQL' >/dev/null
REVOKE ALL ON DATABASE pvnaive FROM PUBLIC;
GRANT CONNECT ON DATABASE pvnaive TO pvnaive_app;
REVOKE TEMPORARY ON DATABASE pvnaive FROM pvnaive_app;
ALTER ROLE pvnaive_app IN DATABASE pvnaive SET search_path = 'pg_catalog,pvnaive';
SQL

secrets_created=1
install -o root -g pvnaive -m 0640 /dev/null /etc/pvnaive/db.env
printf '%s\n' \
  'PVNAIVE_DB_HOST=127.0.0.1' \
  "PVNAIVE_DB_PORT=${cluster_port}" \
  'PVNAIVE_DB_NAME=pvnaive' \
  'PVNAIVE_DB_USER=pvnaive_app' \
  'PVNAIVE_DB_CONNECT_TIMEOUT=5' \
  'PVNAIVE_EXPECTED_SCHEMA_VERSION=1' \
  'PGPASSFILE=/etc/pvnaive/db.pgpass' > /etc/pvnaive/db.env
install -o pvnaive -g pvnaive -m 0600 /dev/null /etc/pvnaive/db.pgpass
printf '127.0.0.1:%s:pvnaive:pvnaive_app:%s\n' "${cluster_port}" "${app_password}" > /etc/pvnaive/db.pgpass
unset app_password
age-keygen -o /etc/pvnaive/backup.agekey >/dev/null 2>&1
chmod 0600 /etc/pvnaive/backup.agekey
chown root:root /etc/pvnaive/backup.agekey
age-keygen -y /etc/pvnaive/backup.agekey > /etc/pvnaive/backup.recipient
chown root:pvnaive /etc/pvnaive/backup.recipient
chmod 0640 /etc/pvnaive/backup.recipient

migration_checksum="$(sha256sum "${repo_root}/db/migrations/0001_initial.up.sql" | awk '{print $1}')"
release_id="0001-${migration_checksum:0:12}"
release_dir="/opt/pvnaive/db/releases/${release_id}"
[[ ! -e "${release_dir}" ]] || die "database release already exists without S03 marker"
release_created=1
install -d -o root -g pvnaive -m 0750 /opt/pvnaive/db /opt/pvnaive/db/releases "${release_dir}" "${release_dir}/db" "${release_dir}/db/migrations" "${release_dir}/scripts" "${release_dir}/scripts/db"
install -o root -g pvnaive -m 0640 "${repo_root}"/db/migrations/* "${release_dir}/db/migrations/"
install -o root -g pvnaive -m 0750 "${repo_root}"/scripts/db/*.sh "${release_dir}/scripts/db/"
ln -s "${release_dir}" "${release_link}.new"
mv -T "${release_link}.new" "${release_link}"

PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${cluster_port}" \
PVNAIVE_DB_NAME=pvnaive \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_MIGRATIONS_DIR="${release_link}/db/migrations" \
  "${release_link}/scripts/db/migrate.sh"

set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
runuser -u pvnaive -- env \
  PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME="${PVNAIVE_DB_NAME}" PVNAIVE_DB_USER="${PVNAIVE_DB_USER}" \
  PVNAIVE_DB_CONNECT_TIMEOUT="${PVNAIVE_DB_CONNECT_TIMEOUT}" \
  PVNAIVE_EXPECTED_SCHEMA_VERSION="${PVNAIVE_EXPECTED_SCHEMA_VERSION}" \
  PGPASSFILE="${PGPASSFILE}" \
  "${release_link}/scripts/db/health.sh"

backup_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${cluster_port}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
    "${release_link}/scripts/db/backup.sh"
)"
echo "${backup_output}"
database_backup="$(echo "${backup_output}" | awk -F= '/^PVNAIVE_BACKUP_PATH=/ {print $2}')"
[[ -f "${database_backup}" ]] || die "encrypted database backup path was not produced"
restore_test_db="pvnaive_restore_test_${stamp,,}"
restore_test_db="${restore_test_db//[^a-z0-9_]/_}"
PVNAIVE_DB_HOST=/var/run/postgresql \
PVNAIVE_DB_PORT="${cluster_port}" \
PVNAIVE_DB_NAME=pvnaive \
PVNAIVE_DB_USER=postgres \
PVNAIVE_RUN_AS_OS_USER=postgres \
PVNAIVE_RESTORE_BACKUP="${database_backup}" \
PVNAIVE_RESTORE_TARGET_DB="${restore_test_db}" \
  "${release_link}/scripts/db/restore.sh"
runuser -u postgres -- dropdb --host /var/run/postgresql --port "${cluster_port}" "${restore_test_db}"
restore_test_db=""

units_installed=1
install -o root -g root -m 0644 "${repo_root}/ops/systemd/pvnaive-db-health.service" /etc/systemd/system/pvnaive-db-health.service
install -o root -g root -m 0644 "${repo_root}/ops/systemd/pvnaive-db-health.timer" /etc/systemd/system/pvnaive-db-health.timer
systemd-analyze verify /etc/systemd/system/pvnaive-db-health.service /etc/systemd/system/pvnaive-db-health.timer
systemctl daemon-reload
timer_enabled=1
systemctl enable --now pvnaive-db-health.timer
systemctl start pvnaive-db-health.service
systemctl is-active --quiet pvnaive-db-health.timer
health_service_result="$(systemctl show --property=Result --value pvnaive-db-health.service)"
[[ "${health_service_result}" == "success" ]] || die "database health service result is ${health_service_result}"

installed_packages="$(dpkg-query -W -f='${Package}=${Version};' "${packages[@]}")"
verify_caddy_invariants
systemctl is-active --quiet postgresql.service || die "postgresql.service is not active"
postgres_listener_snapshot="$(pvnaive_tcp_listener_snapshot)"
printf '%s\n' "${postgres_listener_snapshot}" | pvnaive_tcp_port_is_listening 5432 || die "PostgreSQL TCP port 5432 is not listening"
if printf '%s\n' "${postgres_listener_snapshot}" | pvnaive_tcp_has_non_loopback_postgres_listener; then
  die "PostgreSQL has a non-loopback listener at the final gate"
fi
systemctl is-active --quiet pvnaive-db-health.timer || die "database health timer is not active at the final gate"
systemctl start pvnaive-db-health.service
health_service_result="$(systemctl show --property=Result --value pvnaive-db-health.service)"
[[ "${health_service_result}" == "success" ]] || die "database health service final result is ${health_service_result}"
systemctl --no-pager --full status pvnaive-db-health.service | sed -n '1,18p' || true
(
  cd "$(dirname -- "${database_backup}")"
  sha256sum --check --strict SHA256SUMS
)
(
  cd "${rollback_backup_dir}"
  sha256sum --check --strict SHA256SUMS
)

marker_owned_by_run=1
marker_temp="$(mktemp /opt/pvnaive/.S03_DATABASE.json.XXXXXX)"
cat > "${marker_temp}" <<EOF
{
  "stage": "${stage_id}",
  "completed_at_utc": "${stamp}",
  "host": "$(hostname)",
  "database": "pvnaive",
  "schema_version": 1,
  "postgres_cluster": "${cluster_version}/${cluster_name}",
  "postgres_port": ${cluster_port},
  "listen_addresses": "127.0.0.1,::1",
  "migration_sha256": "${migration_checksum}",
  "encrypted_backup": "${database_backup}",
  "restore_drill": "passed",
  "health_check_result": "${health_service_result}",
  "prechange_backup": "${rollback_backup_dir}",
  "packages": "${installed_packages}",
  "caddyfile_sha256": "${expected_caddy_sha256}",
  "caddy_restarted": false,
  "ssh_changed": false,
  "firewall_changed": false
}
EOF
chown root:pvnaive "${marker_temp}"
chmod 0640 "${marker_temp}"
mv -T -- "${marker_temp}" /opt/pvnaive/S03_DATABASE.json
marker_temp=""

trap - ERR HUP INT TERM
echo
echo "S03_RESULT=PASSED"
echo "SCHEMA_VERSION=1"
echo "PRECHANGE_BACKUP=${rollback_backup_dir}"
echo "DATABASE_BACKUP=${database_backup}"
echo "CADDY_RESTARTED=false"
echo "SSH_CHANGED=false"
echo "FIREWALL_CHANGED=false"
echo "Paste the complete output back into the chat."
