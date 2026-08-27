#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S03-RETRY-BOOTSTRAP"
repo_url="https://github.com/DashSaman/PV-NativePanel.git"
source_commit="6528cabee187f3fdd1c91a392d31f118713646b1"
expected_host="testAmir5-3"
expected_caddy_sha256="101884de2dd11cb9d276df8e72cd068bed50e4ec6eb4ebb477184dda7a86e8b1"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
work="$(mktemp -d /root/pvnaive-s03-retry-src.${stamp}.XXXXXX)"
log="/root/pvnaive-s03-retry-${stamp}.log"
acl_active=0
acl_schema=""
acl_owner=""
acl_app=""
acl_pgpass=""
cluster_port=""

fail() {
  echo "S03_RETRY_BOOTSTRAP=FAILED"
  echo "ERROR=$*"
  return 1
}

postgres_psql() {
  runuser -u postgres -- psql \
    --no-psqlrc \
    --set ON_ERROR_STOP=1 \
    --host /var/run/postgresql \
    --port "${cluster_port}" \
    --username postgres \
    --dbname postgres \
    "$@"
}

cleanup_acl() {
  set +e
  if [[ "${acl_active}" == "1" && -n "${cluster_port}" ]]; then
    if [[ -n "${acl_schema}" ]]; then
      postgres_psql --command "DROP SCHEMA IF EXISTS \"${acl_schema}\" CASCADE" >/dev/null 2>&1
    fi
    if [[ -n "${acl_app}" ]]; then
      postgres_psql --command "DROP OWNED BY \"${acl_app}\"" >/dev/null 2>&1
      postgres_psql --command "DROP ROLE IF EXISTS \"${acl_app}\"" >/dev/null 2>&1
    fi
    if [[ -n "${acl_owner}" ]]; then
      postgres_psql --command "DROP OWNED BY \"${acl_owner}\"" >/dev/null 2>&1
      postgres_psql --command "DROP ROLE IF EXISTS \"${acl_owner}\"" >/dev/null 2>&1
    fi
  fi
  [[ -z "${acl_pgpass}" ]] || rm -f -- "${acl_pgpass}"
  acl_active=0
  set -e
}

cleanup() {
  local rc="$1"
  trap - EXIT HUP INT TERM
  cleanup_acl || true
  if [[ "${rc}" == "0" ]]; then
    rm -rf -- "${work}"
  else
    echo "SOURCE_WORKDIR_PRESERVED=${work}"
  fi
  exit "${rc}"
}
trap 'cleanup "$?"' EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

[[ ${EUID} -eq 0 ]] || fail "run as root"
[[ "$(hostname)" == "${expected_host}" ]] || fail "unexpected host"
for cmd in git psql pg_lsclusters openssl runuser sha256sum systemctl; do
  command -v "${cmd}" >/dev/null 2>&1 || fail "required command missing: ${cmd}"
done
[[ -f /opt/pvnaive/FOUNDATION.json ]] || fail "S02 foundation marker is missing"
[[ ! -f /opt/pvnaive/S03_DATABASE.json ]] || fail "S03 success marker already exists"
[[ ! -f /var/run/reboot-required ]] || fail "reboot is required before retry"
[[ -f /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER ]] || fail "S03 PostgreSQL provenance marker is missing"
grep -Fqx 'stage=S03-DATABASE' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid PostgreSQL provenance marker"
grep -Fqx 'host=testAmir5-3' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid PostgreSQL provenance host"
grep -Fqx 'purpose=pvnaive-dedicated-postgresql' /var/lib/pvnaive/S03_POSTGRES_CLUSTER_OWNER || fail "invalid PostgreSQL provenance purpose"

current_caddy_sha="$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')"
[[ "${current_caddy_sha}" == "${expected_caddy_sha256}" ]] || fail "Caddyfile checksum mismatch"
systemctl is-active --quiet caddy-naive.service || fail "caddy-naive.service is not active"

mapfile -t clusters < <(pg_lsclusters --no-header | awk '{print $1 "|" $2 "|" $3 "|" $4}')
((${#clusters[@]} == 1)) || fail "expected exactly one PostgreSQL cluster, found ${#clusters[@]}"
IFS='|' read -r cluster_version cluster_name cluster_port cluster_status <<< "${clusters[0]}"
[[ "${cluster_version}" == "18" ]] || fail "unexpected PostgreSQL major ${cluster_version}"
[[ "${cluster_name}" == "main" ]] || fail "unexpected PostgreSQL cluster name ${cluster_name}"
[[ "${cluster_port}" == "5432" ]] || fail "unexpected PostgreSQL port ${cluster_port}"
[[ "${cluster_status}" == "online" ]] || fail "PostgreSQL cluster is not online"

preexisting_state="$(postgres_psql --tuples-only --no-align --command "SELECT (SELECT COUNT(*) FROM pg_database WHERE datname='pvnaive') || '|' || (SELECT COUNT(*) FROM pg_roles WHERE rolname IN ('pvnaive_owner','pvnaive_app'))")"
[[ "${preexisting_state}" == "0|0" ]] || fail "pvnaive database or roles remain after prior rollback: ${preexisting_state}"
for path in /etc/pvnaive/db.env /etc/pvnaive/db.pgpass /etc/pvnaive/backup.agekey /etc/pvnaive/backup.recipient /opt/pvnaive/db/current; do
  [[ ! -e "${path}" && ! -L "${path}" ]] || fail "unexpected leftover S03 artifact: ${path}"
done

echo "=== ${stage_id} ==="
echo "UTC=${stamp}"
echo "HOST=$(hostname)"
echo "POSTGRES_CLUSTER=${cluster_version}/${cluster_name}"
echo "POSTGRES_PORT=${cluster_port}"
echo "PINNED_SOURCE_COMMIT=${source_commit}"
echo "LOG=${log}"

# Fetch the exact audited production source commit.
git -C "${work}" init --quiet
git -C "${work}" remote add origin "${repo_url}"
GIT_TERMINAL_PROMPT=0 git -C "${work}" fetch --quiet --depth=1 origin "${source_commit}"
[[ "$(git -C "${work}" rev-parse FETCH_HEAD)" == "${source_commit}" ]] || fail "fetched commit mismatch"
git -C "${work}" checkout --quiet --detach "${source_commit}"
[[ "$(git -C "${work}" rev-parse HEAD)" == "${source_commit}" ]] || fail "checkout commit mismatch"
git -C "${work}" fsck --full --no-dangling >/dev/null

echo "GIT_SOURCE_VERIFY=PASSED"

bash -n "${work}"/scripts/db/*.sh "${work}"/scripts/stages/*.sh "${work}"/tests/stages/*.sh
(
  cd "${work}/db/migrations"
  sha256sum --check --strict SHA256SUMS
)
bash "${work}/tests/stages/S03_preflight_test.sh"
bash "${work}/tests/stages/S03_apt_candidate_pipefail_test.sh"
bash "${work}/tests/stages/S03_ubuntu2604_contract_test.sh"
echo "STATIC_AND_REGRESSION_GATES=PASSED"

# Disposable ACL/identity proof on the existing PostgreSQL 18 cluster.
suffix="${BASHPID}"
acl_schema="pvnaive_acl_probe_${suffix}"
acl_owner="pvnaive_acl_owner_${suffix}"
acl_app="pvnaive_acl_app_${suffix}"
acl_password="$(openssl rand -hex 32)"
acl_pgpass="/var/lib/pvnaive/.acl-probe-${suffix}.pgpass"
acl_active=1

for identifier in "${acl_schema}" "${acl_owner}" "${acl_app}"; do
  [[ "${identifier}" =~ ^[a-z_][a-z0-9_]{0,62}$ ]] || fail "unsafe ACL probe identifier: ${identifier}"
done

probe_existing="$(postgres_psql --tuples-only --no-align --command "SELECT (SELECT COUNT(*) FROM pg_namespace WHERE nspname='${acl_schema}') || '|' || (SELECT COUNT(*) FROM pg_roles WHERE rolname IN ('${acl_owner}','${acl_app}'))")"
[[ "${probe_existing}" == "0|0" ]] || fail "ACL probe names already exist"

postgres_psql <<SQL >/dev/null
CREATE ROLE "${acl_owner}" NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE "${acl_app}" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD '${acl_password}';
CREATE SCHEMA "${acl_schema}" AUTHORIZATION "${acl_owner}";
SET ROLE "${acl_owner}";
CREATE TABLE "${acl_schema}".visible (id bigint PRIMARY KEY);
CREATE TABLE "${acl_schema}".secret (signing_key bytea NOT NULL);
REVOKE ALL ON ALL TABLES IN SCHEMA "${acl_schema}" FROM PUBLIC;
GRANT USAGE ON SCHEMA "${acl_schema}" TO "${acl_app}";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA "${acl_schema}" TO "${acl_app}";
REVOKE ALL ON "${acl_schema}".secret FROM "${acl_app}";
RESET ROLE;
SQL

catalog_secret_select="$(postgres_psql --tuples-only --no-align --command "SELECT has_table_privilege('${acl_app}', '${acl_schema}.secret', 'SELECT')")"
catalog_visible_select="$(postgres_psql --tuples-only --no-align --command "SELECT has_table_privilege('${acl_app}', '${acl_schema}.visible', 'SELECT')")"
role_is_super="$(postgres_psql --tuples-only --no-align --command "SELECT rolsuper FROM pg_roles WHERE rolname='${acl_app}'")"
role_is_owner_member="$(postgres_psql --tuples-only --no-align --command "SELECT pg_has_role('${acl_app}', '${acl_owner}', 'MEMBER')")"
[[ "${catalog_secret_select}" == "f" ]] || fail "ACL probe catalog says secret SELECT is allowed"
[[ "${catalog_visible_select}" == "t" ]] || fail "ACL probe catalog says visible SELECT is denied"
[[ "${role_is_super}" == "f" ]] || fail "ACL probe app role became superuser"
[[ "${role_is_owner_member}" == "f" ]] || fail "ACL probe app role unexpectedly inherits owner"

install -o pvnaive -g pvnaive -m 0600 /dev/null "${acl_pgpass}"
printf '127.0.0.1:%s:postgres:%s:%s\n' "${cluster_port}" "${acl_app}" "${acl_password}" > "${acl_pgpass}"
chown pvnaive:pvnaive "${acl_pgpass}"
chmod 0600 "${acl_pgpass}"
unset acl_password

probe_identity="$(runuser -u pvnaive -- env -i PATH=/usr/bin:/bin PGPASSFILE="${acl_pgpass}" psql --no-psqlrc --set ON_ERROR_STOP=1 --host 127.0.0.1 --port "${cluster_port}" --username "${acl_app}" --dbname postgres --tuples-only --no-align --command "SELECT current_user || '|' || session_user")"
[[ "${probe_identity}" == "${acl_app}|${acl_app}" ]] || fail "TCP identity probe mismatch: ${probe_identity}"

runuser -u pvnaive -- env -i PATH=/usr/bin:/bin PGPASSFILE="${acl_pgpass}" psql --no-psqlrc --set ON_ERROR_STOP=1 --host 127.0.0.1 --port "${cluster_port}" --username "${acl_app}" --dbname postgres --command "SELECT id FROM \"${acl_schema}\".visible LIMIT 0" >/dev/null

set +e
runuser -u pvnaive -- env -i PATH=/usr/bin:/bin PGPASSFILE="${acl_pgpass}" psql --no-psqlrc --set ON_ERROR_STOP=1 --host 127.0.0.1 --port "${cluster_port}" --username "${acl_app}" --dbname postgres --command "SELECT signing_key FROM \"${acl_schema}\".secret LIMIT 0" >/dev/null 2>&1
secret_probe_rc=$?
set -e
[[ "${secret_probe_rc}" != "0" ]] || fail "actual TCP secret SELECT unexpectedly succeeded"

echo "ACL_CATALOG_SECRET_SELECT=DENIED"
echo "ACL_TCP_IDENTITY=${probe_identity}"
echo "ACL_TCP_SECRET_SELECT=DENIED"
echo "ACL_AND_IDENTITY_PROBE=PASSED"

cleanup_acl

probe_cleanup="$(postgres_psql --tuples-only --no-align --command "SELECT (SELECT COUNT(*) FROM pg_namespace WHERE nspname='${acl_schema}') || '|' || (SELECT COUNT(*) FROM pg_roles WHERE rolname IN ('${acl_owner}','${acl_app}'))")"
[[ "${probe_cleanup}" == "0|0" ]] || fail "ACL probe cleanup incomplete: ${probe_cleanup}"
echo "ACL_PROBE_CLEANUP=PASSED"

# Ensure the real S03 cannot inherit libpq/admin helper identity from the caller.
cat <<'EOF'
==================================================
ALL RETRY GATES PASSED — STARTING REAL S03
==================================================
EOF

set +e
env \
  -u PVNAIVE_RUN_AS_OS_USER \
  -u PGUSER \
  -u PGHOST \
  -u PGPORT \
  -u PGDATABASE \
  -u PGSERVICE \
  -u PGPASSFILE \
  bash "${work}/scripts/stages/S03-database.sh" 2>&1 | tee "${log}"
stage_rc=${PIPESTATUS[0]}
set -e

echo "S03_STAGE_EXIT=${stage_rc}"
echo "S03_LOG=${log}"

[[ "$(sha256sum /etc/caddy/Caddyfile | awk '{print $1}')" == "${expected_caddy_sha256}" ]] || fail "Caddyfile changed during retry"
systemctl is-active --quiet caddy-naive.service || fail "Caddy inactive after retry"

if [[ "${stage_rc}" != "0" ]]; then
  echo "S03_RETRY_BOOTSTRAP=FAILED"
  exit "${stage_rc}"
fi

[[ -f /opt/pvnaive/S03_DATABASE.json ]] || fail "S03 exited zero without success marker"
grep -Fqx '  "stage": "S03-DATABASE",' /opt/pvnaive/S03_DATABASE.json || fail "invalid S03 success marker"
systemctl is-active --quiet pvnaive-db-health.timer || fail "database health timer inactive"
systemctl start pvnaive-db-health.service
[[ "$(systemctl show --property=Result --value pvnaive-db-health.service)" == "success" ]] || fail "database health service final result is not success"

echo "S03_RETRY_BOOTSTRAP=PASSED"
echo "PINNED_SOURCE_COMMIT=${source_commit}"
echo "S03_LOG=${log}"
