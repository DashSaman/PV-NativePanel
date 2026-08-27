#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S04-AUTH"
stage_root_bashpid="${BASHPID}"
expected_host="testAmir5-3"
expected_domain="namir.softarg.ir"
expected_ipv4="91.107.182.147"
expected_caddy_sha256="101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1"
expected_schema_before="1"
expected_schema_after="2"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
# shellcheck source=scripts/stages/lib.sh
source "${script_dir}/lib.sh"

marker="/opt/pvnaive/S04_AUTH.json"
s03_marker="/opt/pvnaive/S03_DATABASE.json"
auth_key="/etc/pvnaive/auth.key"
api_unit="/etc/systemd/system/pvnaive-api.service"
api_binary="/opt/pvnaive/bin/pvnaive"
password_binary="/opt/pvnaive/bin/pvnaive-password"
release_root="/opt/pvnaive/auth/releases"
release_dir="${release_root}/${stamp}"
release_link="/opt/pvnaive/auth/current"
web_release_root="/opt/pvnaive/web/releases"
web_release_dir="${web_release_root}/${stamp}"
web_release_link="/opt/pvnaive/web/current"
pre_backup=""
rollback_backup=""
migration_owned=0
release_created=0
web_release_created=0
binaries_installed=0
auth_key_created=0
unit_installed=0
service_enabled=0
marker_created=0
rollback_failures=()

postgres_psql() {
  runuser -u postgres -- psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql --port "${PVNAIVE_DB_PORT}" --username postgres "$@"
}

verify_caddy_invariants() {
  local current_sha listener_snapshot
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

current_schema_version() {
  postgres_psql --dbname pvnaive --tuples-only --no-align \
    --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations'
}

rollback_on_error() {
  local line="$1" code="$2"
  if ! pvnaive_is_root_bash_process "${stage_root_bashpid}"; then
    trap - ERR
    exit "${code}"
  fi
  trap - ERR HUP INT TERM
  set +e
  echo
  echo "S04_RESULT=FAILED"
  echo "FAILED_LINE=${line}"
  echo "FAILED_EXIT=${code}"
  echo "ROLLBACK=STARTED"

  if [[ "${service_enabled}" == "1" ]]; then
    systemctl disable --now pvnaive-api.service >/dev/null 2>&1 || rollback_failures+=("disable-api-service")
  fi
  if [[ "${unit_installed}" == "1" ]]; then
    rm -f -- "${api_unit}" || rollback_failures+=("remove-api-unit")
    systemctl daemon-reload || rollback_failures+=("systemd-daemon-reload")
  fi
  if [[ "${auth_key_created}" == "1" ]]; then
    rm -f -- "${auth_key}" || rollback_failures+=("remove-auth-key")
  fi
  if [[ "${binaries_installed}" == "1" ]]; then
    rm -f -- "${api_binary}" "${password_binary}" || rollback_failures+=("remove-binaries")
  fi
  if [[ "${release_created}" == "1" ]]; then
    [[ ! -L "${release_link}" || "$(readlink -f "${release_link}")" != "${release_dir}" ]] || rm -f -- "${release_link}"
    rm -rf -- "${release_dir}" || rollback_failures+=("remove-auth-release")
  fi
  if [[ "${web_release_created}" == "1" ]]; then
    [[ ! -L "${web_release_link}" || "$(readlink -f "${web_release_link}")" != "${web_release_dir}" ]] || rm -f -- "${web_release_link}"
    rm -rf -- "${web_release_dir}" || rollback_failures+=("remove-web-release")
  fi
  if [[ "${marker_created}" == "1" ]]; then
    rm -f -- "${marker}" || rollback_failures+=("remove-S04-marker")
  fi

  if [[ "${migration_owned}" == "1" ]]; then
    if [[ -n "${rollback_backup}" && -f "${rollback_backup}" ]]; then
      PVNAIVE_DB_HOST=/var/run/postgresql \
      PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
      PVNAIVE_DB_NAME=pvnaive \
      PVNAIVE_DB_USER=postgres \
      PVNAIVE_RUN_AS_OS_USER=postgres \
      PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
      PVNAIVE_ALLOW_DESTRUCTIVE_ROLLBACK=ROLLBACK_ONE_MIGRATION \
      PVNAIVE_CONFIRMED_BACKUP="${rollback_backup}" \
        "${repo_root}/scripts/db/rollback.sh" >/dev/null 2>&1 || rollback_failures+=("rollback-migration-0002")
    else
      rollback_failures+=("rollback-migration-0002-no-schema2-backup")
    fi
  fi

  verify_caddy_invariants >/dev/null 2>&1 || rollback_failures+=("caddy-invariant")
  if ((${#rollback_failures[@]} == 0)); then
    echo "ROLLBACK=COMPLETED"
  else
    echo "ROLLBACK=INCOMPLETE"
    printf 'ROLLBACK_FAILED_STEP=%s\n' "${rollback_failures[@]}"
  fi
  echo "CADDY_ACTION=none"
  echo "SSH_ACTION=none"
  echo "FIREWALL_ACTION=none"
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
echo "CADDY_ACTION=none"
echo "SSH_ACTION=none"
echo "FIREWALL_ACTION=none"
echo "API_TARGET=127.0.0.1:8080"
echo "ROLLBACK_PLAN=Disable/remove the new API service, key and binaries; roll migration 0002 back only with a verified schema-v2 encrypted backup; never mutate Caddy, SSH or firewall."

[[ "$(hostname)" == "${expected_host}" ]] || die "unexpected host"
getent ahostsv4 "${expected_domain}" | awk -v expected="${expected_ipv4}" '$1 == expected {found=1} END {exit !found}' || die "DNS mismatch"
[[ -f "${s03_marker}" ]] || die "S03_DATABASE.json is missing"
grep -Fqx '  "stage": "S03-DATABASE",' "${s03_marker}" || die "invalid S03 marker"
grep -Fqx '  "host": "testAmir5-3",' "${s03_marker}" || die "S03 marker host mismatch"
[[ -r /etc/pvnaive/db.env ]] || die "database environment is missing"
[[ -r /etc/pvnaive/backup.agekey && -r /etc/pvnaive/backup.recipient ]] || die "database backup encryption material is missing"
[[ -x /usr/local/bin/caddy && -f /etc/caddy/Caddyfile ]] || die "Caddy baseline is missing"
verify_caddy_invariants

for required_source in \
  db/migrations/0001_initial.up.sql db/migrations/0001_initial.down.sql \
  db/migrations/0002_auth_foundation.up.sql db/migrations/0002_auth_foundation.down.sql \
  db/migrations/SHA256SUMS \
  scripts/db/lib.sh scripts/db/migrate.sh scripts/db/rollback.sh scripts/db/backup.sh \
  scripts/auth/bootstrap-owner.sh ops/systemd/pvnaive-api.service \
  dist/s04/linux-amd64/pvnaive dist/s04/linux-amd64/pvnaive-password; do
  [[ -f "${repo_root}/${required_source}" ]] || die "bundle file missing: ${required_source}"
done
[[ -d "${repo_root}/dist/s04/web" ]] || die "bundled web build is missing"
[[ -f "${repo_root}/S04_SHA256SUMS" ]] || die "S04 bundle checksum manifest is missing"
(
  cd "${repo_root}"
  sha256sum --check --strict S04_SHA256SUMS
) >/dev/null || die "S04 bundle checksum verification failed"
(
  cd "${repo_root}/db/migrations"
  sha256sum --check --strict SHA256SUMS
) >/dev/null || die "migration checksum verification failed"
bash -n "${repo_root}"/scripts/db/*.sh "${repo_root}"/scripts/auth/*.sh "${repo_root}"/scripts/stages/*.sh
[[ -x "${repo_root}/dist/s04/linux-amd64/pvnaive" && -x "${repo_root}/dist/s04/linux-amd64/pvnaive-password" ]] || die "bundled binaries are not executable"
file "${repo_root}/dist/s04/linux-amd64/pvnaive" | grep -Eq 'ELF 64-bit LSB.*x86-64' || die "pvnaive binary architecture mismatch"
file "${repo_root}/dist/s04/linux-amd64/pvnaive-password" | grep -Eq 'ELF 64-bit LSB.*x86-64' || die "password helper architecture mismatch"

set -a
# shellcheck disable=SC1091
source /etc/pvnaive/db.env
set +a
[[ "${PVNAIVE_DB_HOST:-}" == "127.0.0.1" ]] || die "database host is not IPv4 loopback"
[[ "${PVNAIVE_DB_NAME:-}" == "pvnaive" ]] || die "unexpected database name"
[[ "${PVNAIVE_DB_PORT:-}" =~ ^[0-9]+$ ]] || die "invalid database port"

if [[ -f "${marker}" ]]; then
  grep -Fqx '  "stage": "S04-AUTH",' "${marker}" || die "invalid S04 marker"
  grep -Fqx '  "host": "testAmir5-3",' "${marker}" || die "S04 marker host mismatch"
  [[ "$(current_schema_version)" == "${expected_schema_after}" ]] || die "S04 marker exists but schema is not version 2"
  systemctl is-active --quiet pvnaive-api.service || die "S04 marker exists but API service is not active"
  ss -H -lnt | awk '$4 == "127.0.0.1:8080" {found=1} END {exit !found}' || die "API is not loopback-only on 127.0.0.1:8080"
  curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/live >/dev/null
  curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || die "API readiness check failed"
  verify_caddy_invariants
  trap - ERR HUP INT TERM
  echo "S04_RESULT=PASSED"
  echo "S04_MODE=VERIFIED_EXISTING"
  exit 0
fi

schema_before="$(current_schema_version)"
if [[ "${schema_before}" == "${expected_schema_before}" ]]; then
  :
elif [[ "${schema_before}" == "${expected_schema_after}" ]]; then
  stored_0002="$(postgres_psql --dbname pvnaive --tuples-only --no-align --command "SELECT filename || '|' || checksum_sha256 FROM pvnaive.schema_migrations WHERE version=2")"
  manifest_0002="$(awk '$2 == "0002_auth_foundation.up.sql" {print $1}' "${repo_root}/db/migrations/SHA256SUMS")"
  [[ "${stored_0002}" == "0002_auth_foundation.up.sql|${manifest_0002}" ]] || die "unowned schema-v2 state does not match this S04 migration"
  echo "RECOVERY_MODE=SCHEMA2_WITHOUT_MARKER"
else
  die "unexpected schema version ${schema_before}; expected 1 or recoverable 2"
fi

for new_path in "${auth_key}" "${api_unit}" "${api_binary}" "${password_binary}" "${release_link}" "${web_release_link}"; do
  [[ ! -e "${new_path}" && ! -L "${new_path}" ]] || die "pre-existing S04 artifact requires manual inspection: ${new_path}"
done

install -d -o root -g pvnaive -m 0750 /opt/pvnaive/bin /opt/pvnaive/auth "${release_root}" /opt/pvnaive/web "${web_release_root}"
install -d -o root -g pvnaive -m 0750 "${release_dir}" "${web_release_dir}"
release_created=1
web_release_created=1
cp -a "${repo_root}/db" "${release_dir}/db"
mkdir -p "${release_dir}/scripts"
cp -a "${repo_root}/scripts/db" "${release_dir}/scripts/db"
cp -a "${repo_root}/scripts/auth" "${release_dir}/scripts/auth"
cp -a "${repo_root}/dist/s04/web/." "${web_release_dir}/"
chown -R root:pvnaive "${release_dir}" "${web_release_dir}"
chmod -R go-rwx "${release_dir}"
find "${web_release_dir}" -type d -exec chmod 0750 {} +
find "${web_release_dir}" -type f -exec chmod 0640 {} +
ln -s "${release_dir}" "${release_link}.new"
mv -Tf "${release_link}.new" "${release_link}"
ln -s "${web_release_dir}" "${web_release_link}.new"
mv -Tf "${web_release_link}.new" "${web_release_link}"

install -o root -g pvnaive -m 0750 "${repo_root}/dist/s04/linux-amd64/pvnaive" "${api_binary}"
install -o root -g pvnaive -m 0750 "${repo_root}/dist/s04/linux-amd64/pvnaive-password" "${password_binary}"
binaries_installed=1

if [[ "${schema_before}" == "${expected_schema_before}" ]]; then
  backup_output="$(
    PVNAIVE_DB_HOST=/var/run/postgresql \
    PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
    PVNAIVE_DB_NAME=pvnaive \
    PVNAIVE_DB_USER=postgres \
    PVNAIVE_RUN_AS_OS_USER=postgres \
      "${release_link}/scripts/db/backup.sh"
  )"
  echo "${backup_output}"
  pre_backup="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {print $2}' <<<"${backup_output}")"
  [[ -f "${pre_backup}" ]] || die "pre-S04 encrypted backup was not produced"

  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
  PVNAIVE_MIGRATIONS_DIR="${release_link}/db/migrations" \
    "${release_link}/scripts/db/migrate.sh"
  [[ "$(current_schema_version)" == "${expected_schema_after}" ]] || die "schema did not reach version 2"
  migration_owned=1
else
  pre_backup="RECOVERY_SCHEMA_ALREADY_2"
fi

backup_output="$(
  PVNAIVE_DB_HOST=/var/run/postgresql \
  PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
  PVNAIVE_DB_NAME=pvnaive \
  PVNAIVE_DB_USER=postgres \
  PVNAIVE_RUN_AS_OS_USER=postgres \
    "${release_link}/scripts/db/backup.sh"
)"
echo "${backup_output}"
rollback_backup="$(awk -F= '/^PVNAIVE_BACKUP_PATH=/ {print $2}' <<<"${backup_output}")"
[[ -f "${rollback_backup}" ]] || die "schema-v2 encrypted backup was not produced"

install -o root -g pvnaive -m 0640 /dev/null "${auth_key}"
dd if=/dev/urandom of="${auth_key}" bs=32 count=1 status=none
[[ "$(stat -c '%s' "${auth_key}")" == "32" ]] || die "authentication key generation failed"
auth_key_created=1

install -o root -g root -m 0644 "${repo_root}/ops/systemd/pvnaive-api.service" "${api_unit}"
unit_installed=1
systemctl daemon-reload
systemctl enable --now pvnaive-api.service
service_enabled=1

for _ in $(seq 1 20); do
  if curl --fail --silent http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true'; then
    break
  fi
  sleep 1
done
systemctl is-active --quiet pvnaive-api.service || die "pvnaive-api.service is not active"
ss -H -lnt | awk '$4 == "127.0.0.1:8080" {found=1} END {exit !found}' || die "API listener is missing or not IPv4 loopback"
if ss -H -lnt | awk '$4 ~ /:8080$/ && $4 != "127.0.0.1:8080" {bad=1} END {exit !bad}'; then
  die "API has a non-loopback listener"
fi
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/live | grep -q '"status":"ok"' || die "API liveness check failed"
curl --fail --silent --show-error http://127.0.0.1:8080/api/v1/health/ready | grep -q '"ready":true' || die "API readiness check failed"
[[ "$(current_schema_version)" == "${expected_schema_after}" ]] || die "post-service schema verification failed"
verify_caddy_invariants

marker_tmp="${marker}.tmp.${BASHPID}"
printf '{\n  "stage": "S04-AUTH",\n  "host": "testAmir5-3",\n  "completed_at_utc": "%s",\n  "schema_version": 2,\n  "api_listener": "127.0.0.1:8080",\n  "api_service": "pvnaive-api.service",\n  "auth_release": "%s",\n  "web_release": "%s",\n  "pre_migration_backup": "%s",\n  "rollback_backup": "%s",\n  "caddyfile_sha256": "%s",\n  "caddy_changed": false,\n  "ssh_changed": false,\n  "firewall_changed": false\n}\n' \
  "${stamp}" "${release_dir}" "${web_release_dir}" "${pre_backup}" "${rollback_backup}" "${expected_caddy_sha256}" > "${marker_tmp}"
chown root:pvnaive "${marker_tmp}"
chmod 0640 "${marker_tmp}"
mv -f "${marker_tmp}" "${marker}"
marker_created=1

verify_caddy_invariants
trap - ERR HUP INT TERM
echo "S04_RESULT=PASSED"
echo "S04_MODE=LOCALHOST_READY"
echo "S04_SCHEMA_VERSION=2"
echo "S04_API_LISTENER=127.0.0.1:8080"
echo "S04_OWNER_BOOTSTRAP=${release_link}/scripts/auth/bootstrap-owner.sh"
echo "CADDY_ACTION=none"
echo "SSH_ACTION=none"
echo "FIREWALL_ACTION=none"
