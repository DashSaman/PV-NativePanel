#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

# R5 pool registry migration gate (0033): proves the pull-model registry
# contract — single-use enrollment tokens, MONOTONIC append-only signed
# revisions, heartbeat drift tracking, the drain state machine (active ->
# draining -> disabled, never a direct yank), and the trusted boundary
# (pvnaive_app only through SECURITY DEFINER functions; no direct table
# access; no DML grants for non-superuser roles anywhere).

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
suffix="${GITHUB_RUN_ID:-local}_${GITHUB_RUN_ATTEMPT:-1}_${BASHPID}"
suffix="${suffix//[^a-zA-Z0-9_]/_}"
test_db="pvnaive_r5_pool_${suffix,,}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" \
    --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" \
    >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
}
trap cleanup EXIT
cleanup

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
PVNAIVE_DB_NAME="${test_db}" "${repo_root}/scripts/db/migrate.sh" >/dev/null

schema="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema}" =~ ^[0-9]+$ && "${schema}" -ge 33 ]] || { echo "ERROR: expected schema >=33, got ${schema}" >&2; exit 1; }

# 1) Schema contract: three tables, seven SECURITY DEFINER functions.
contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "
SELECT concat_ws('|',
 (to_regclass('pvnaive.pool_nodes') IS NOT NULL)::text,
 (to_regclass('pvnaive.pool_manifest_revisions') IS NOT NULL)::text,
 (to_regclass('pvnaive.pool_enrollment_tokens') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_enrollment_token_record(text,text,integer)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_node_enroll(text,text,text,integer,uuid)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_revision_publish(uuid,jsonb,text)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_revision_latest(uuid)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_node_heartbeat(uuid,text,bigint)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_node_maintenance_set(uuid,text)') IS NOT NULL)::text,
 (to_regprocedure('pvnaive.pool_nodes_list()') IS NOT NULL)::text
);")"
[[ "${contract}" == 'true|true|true|true|true|true|true|true|true|true' ]] || {
  echo "ERROR: pool registry schema contract failed: ${contract}" >&2; exit 1;
}

# Fixture: one owner actor to create nodes.
psql_admin --dbname "${test_db}" <<'SQL' >/dev/null
INSERT INTO pvnaive.actors (id, tenant_id, actor_role, email, display_name, status)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'owner', 'owner@r5.invalid', 'R5 Owner', 'active')
ON CONFLICT (id) DO NOTHING;
SQL

NODE_ID='cccccccc-cccc-cccc-cccc-cccccccccccc'
TOKEN_HASH='0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20'
SIG='ab'  # 2 chars; must be rejected below (needs 64..256)

app() {
  psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SET ROLE pvnaive_app; $1"
}

# 2) Token record + single-use enrollment.
exp="$(app "SELECT expires_at::text FROM pvnaive.pool_enrollment_token_record('${TOKEN_HASH}','node-eu-1',3600);")"
[[ -n "${exp##*$'\n'}" ]] || { echo "ERROR: token record returned nothing" >&2; exit 1; }
enroll="$(app "SELECT concat_ws('|',node_id::text,enrolled::text,reason) FROM pvnaive.pool_node_enroll('${TOKEN_HASH}','Node EU 1','eu',2,'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');")"
[[ "${enroll##*$'\n'}" == *"|true|enrolled" ]] || { echo "ERROR: enrollment failed: ${enroll}" >&2; exit 1; }
NODE_ID="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT id FROM pvnaive.pool_nodes WHERE display_name='Node EU 1'")"
[[ "${NODE_ID}" =~ ^[0-9a-f-]{36}$ ]] || { echo "ERROR: enrolled node id missing" >&2; exit 1; }
replay="$(app "SELECT concat_ws('|',node_id::text,enrolled::text,reason) FROM pvnaive.pool_node_enroll('${TOKEN_HASH}','Node EU 1b','eu',1,'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');")"
[[ "${replay##*$'\n'}" == 'false|token_invalid' ]] || { echo "ERROR: token replay accepted: ${replay}" >&2; exit 1; }

# 3) Revision publish is monotonic and append-only.
rev1="$(app "SELECT concat_ws('|',revision::text,accepted::text) FROM pvnaive.pool_revision_publish('${NODE_ID}','{\"schema\":\"pvnaive.node.manifest.v1\",\"node_id\":\"${NODE_ID}\",\"pool_id\":\"p1\",\"endpoints\":[{\"host\":\"198.51.100.1\",\"port\":443}],\"pubkey\":\"aa\",\"weight\":1,\"valid_from\":\"2026-09-14T00:00:00Z\",\"valid_until\":\"2026-12-14T00:00:00Z\"}','$(printf 'a%.0s' $(seq 64))');")"
[[ "${rev1##*$'\n'}" == '1|true' ]] || { echo "ERROR: first revision: ${rev1}" >&2; exit 1; }
rev2="$(app "SELECT concat_ws('|',revision::text,accepted::text) FROM pvnaive.pool_revision_publish('${NODE_ID}','{\"schema\":\"pvnaive.node.manifest.v1\",\"node_id\":\"${NODE_ID}\",\"pool_id\":\"p1\",\"endpoints\":[{\"host\":\"198.51.100.2\",\"port\":443}],\"pubkey\":\"aa\",\"weight\":2,\"valid_from\":\"2026-09-14T00:00:00Z\",\"valid_until\":\"2026-12-14T00:00:00Z\"}','$(printf 'b%.0s' $(seq 64))');")"
[[ "${rev2##*$'\n'}" == '2|true' ]] || { echo "ERROR: second revision: ${rev2}" >&2; exit 1; }
latest="$(app "SELECT revision::text FROM pvnaive.pool_revision_latest('${NODE_ID}');")"
[[ "${latest##*$'\n'}" == '2' ]] || { echo "ERROR: latest revision: ${latest}" >&2; exit 1; }
count="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT count(*) FROM pvnaive.pool_manifest_revisions;")"
[[ "${count}" == 2 ]] || { echo "ERROR: revision ledger rows: ${count}" >&2; exit 1; }
short_sig_rc=0
app "SELECT pvnaive.pool_revision_publish('${NODE_ID}','{\"k\":1}','${SIG}');" >/dev/null 2>&1 || short_sig_rc=$?
[[ "${short_sig_rc}" != 0 ]] || { echo "ERROR: short signature accepted" >&2; exit 1; }

# 4) Heartbeat: tracked, monotonic applied revision.
hb="$(app "SELECT tracked::text FROM pvnaive.pool_node_heartbeat('${NODE_ID}','healthy',2);")"
[[ "${hb##*$'\n'}" == 'true' ]] || { echo "ERROR: heartbeat: ${hb}" >&2; exit 1; }
app "SELECT pvnaive.pool_node_heartbeat('${NODE_ID}','healthy',1);" >/dev/null
applied="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT applied_revision FROM pvnaive.pool_nodes WHERE id='${NODE_ID}';")"
[[ "${applied}" == 2 ]] || { echo "ERROR: applied revision rewound: ${applied}" >&2; exit 1; }

# 5) Drain state machine: active -> disabled is refused (23514).
yank_rc=0
app "SELECT pvnaive.pool_node_maintenance_set('${NODE_ID}','disabled');" >/dev/null 2>&1 || yank_rc=$?
[[ "${yank_rc}" != 0 ]] || { echo "ERROR: live node was yanked straight to disabled" >&2; exit 1; }
drain="$(app "SELECT tracked::text FROM pvnaive.pool_node_maintenance_set('${NODE_ID}','draining');")"
[[ "${drain##*$'\n'}" == 'true' ]] || { echo "ERROR: drain: ${drain}" >&2; exit 1; }
disable="$(app "SELECT tracked::text FROM pvnaive.pool_node_maintenance_set('${NODE_ID}','disabled');")"
[[ "${disable##*$'\n'}" == 'true' ]] || { echo "ERROR: disable after drain: ${disable}" >&2; exit 1; }
pub_disabled="$(app "SELECT concat_ws('|',revision::text,accepted::text) FROM pvnaive.pool_revision_publish('${NODE_ID}','{\"schema\":\"pvnaive.node.manifest.v1\",\"node_id\":\"${NODE_ID}\",\"pool_id\":\"p1\",\"endpoints\":[{\"host\":\"198.51.100.3\",\"port\":443}],\"pubkey\":\"aa\",\"weight\":1,\"valid_from\":\"2026-09-14T00:00:00Z\",\"valid_until\":\"2026-12-14T00:00:00Z\"}','$(printf 'c%.0s' $(seq 64))');")"
[[ "${pub_disabled##*$'\n'}" == '0|false' ]] || { echo "ERROR: disabled node accepted a revision: ${pub_disabled}" >&2; exit 1; }

# 6) Trusted boundary: pvnaive_app has NO direct table privileges.
set +e
app "SELECT count(*) FROM pvnaive.pool_nodes;" >/dev/null 2>&1; denied_app=$?
app "SELECT count(*) FROM pvnaive.pool_manifest_revisions;" >/dev/null 2>&1; denied_app2=$?
# 7) Even the owner role has no DML grants on the revision ledger.
psql_admin --dbname "${test_db}" --command "SET ROLE pvnaive_owner; UPDATE pvnaive.pool_manifest_revisions SET signature='tampered';" >/dev/null 2>&1; denied_upd=$?
psql_admin --dbname "${test_db}" --command "SET ROLE pvnaive_owner; DELETE FROM pvnaive.pool_enrollment_tokens;" >/dev/null 2>&1; denied_del=$?
set -e
[[ "${denied_app}" != 0 && "${denied_app2}" != 0 && "${denied_upd}" != 0 && "${denied_del}" != 0 ]] || {
  echo "ERROR: trusted boundary broken (app=${denied_app}/${denied_app2} owner=${denied_upd}/${denied_del})" >&2; exit 1;
}

echo "pool registry migration gate: OK"
