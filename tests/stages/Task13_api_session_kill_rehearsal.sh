#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
api_binary="${PVNAIVE_REHEARSAL_API_BINARY:-${repo_root}/dist/rehearsal/pvnaive}"
agent_binary="${PVNAIVE_REHEARSAL_RUNTIME_AGENT_BINARY:-${repo_root}/dist/rehearsal/pvnaive-runtime-agent-rehearsal}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

for required in psql createdb dropdb curl jq sha256sum python3; do
  command -v "${required}" >/dev/null 2>&1 || { echo "ERROR: missing ${required}" >&2; exit 1; }
done
for binary in "${api_binary}" "${agent_binary}"; do
  [[ -x "${binary}" ]] || { echo "ERROR: rehearsal binary missing: ${binary}" >&2; exit 1; }
done

test_db="pvnaive"
app_password='pvnaive-task13-ci-only'
api_port=18083
api_pid=''
agent_pid=''
control_pid=''
tmpdir=''

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  for pid in "${api_pid}" "${agent_pid}" "${control_pid}"; do
    if [[ -n "${pid}" ]]; then kill "${pid}" >/dev/null 2>&1 || true; wait "${pid}" >/dev/null 2>&1 || true; fi
  done
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
  [[ -z "${tmpdir}" ]] || rm -rf -- "${tmpdir}" || true
}
cleanup
tmpdir="$(mktemp -d)"
trap cleanup EXIT HUP INT TERM

psql_admin --dbname postgres <<SQL >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN PASSWORD '${app_password}' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"
PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == 20 ]] || { echo "ERROR: Task13 rehearsal expected schema20, got ${schema_version}" >&2; exit 1; }

tenant_a='17170000-0000-0000-0000-000000001002'
tenant_b='17170000-0000-0000-0000-000000001003'
actor_a='17170000-0000-0000-0000-000000001011'
actor_b='17170000-0000-0000-0000-000000001012'
user_a='17170000-0000-0000-0000-000000001041'
user_b='17170000-0000-0000-0000-000000001042'
cred_a='17170000-0000-0000-0000-000000001051'
cred_b='17170000-0000-0000-0000-000000001052'
term_a='17170000-0000-0000-0000-000000001061'
term_b='17170000-0000-0000-0000-000000001062'
boot_a='17170000-0000-0000-0000-000000001081'
boot_b='17170000-0000-0000-0000-000000001082'
session_a='17170000-0000-0000-0000-000000001091'
session_b='17170000-0000-0000-0000-000000001092'
raw_session_a='task13-session-token-a'
raw_session_b='task13-session-token-b'
raw_csrf_a='task13-csrf-token-a'
raw_csrf_b='task13-csrf-token-b'
token_hash_a="$(printf %s "${raw_session_a}" | sha256sum | awk '{print $1}')"
token_hash_b="$(printf %s "${raw_session_b}" | sha256sum | awk '{print $1}')"
csrf_hash_a="$(printf %s "${raw_csrf_a}" | sha256sum | awk '{print $1}')"
csrf_hash_b="$(printf %s "${raw_csrf_b}" | sha256sum | awk '{print $1}')"

psql_admin --dbname "${test_db}" \
  --set=tenant_a="${tenant_a}" --set=tenant_b="${tenant_b}" \
  --set=actor_a="${actor_a}" --set=actor_b="${actor_b}" \
  --set=user_a="${user_a}" --set=user_b="${user_b}" \
  --set=cred_a="${cred_a}" --set=cred_b="${cred_b}" \
  --set=term_a="${term_a}" --set=term_b="${term_b}" \
  --set=boot_a="${boot_a}" --set=boot_b="${boot_b}" \
  --set=session_a="${session_a}" --set=session_b="${session_b}" \
  --set=token_hash_a="${token_hash_a}" --set=token_hash_b="${token_hash_b}" \
  --set=csrf_hash_a="${csrf_hash_a}" --set=csrf_hash_b="${csrf_hash_b}" <<'SQL' >/dev/null
INSERT INTO pvnaive.tenants(id,tenant_type,slug,display_name,status) VALUES
(:'tenant_a','reseller','task13-a','Task13 A','active'),
(:'tenant_b','reseller','task13-b','Task13 B','active');
INSERT INTO pvnaive.actors(id,tenant_id,actor_role,email,display_name,status) VALUES
(:'actor_a',:'tenant_a','reseller','task13-a@example.invalid','Task13 A','active'),
(:'actor_b',:'tenant_b','reseller','task13-b@example.invalid','Task13 B','active');
INSERT INTO pvnaive.auth_sessions(id,tenant_id,actor_id,token_hash,refresh_family_id,user_agent_hash,expires_at,absolute_expires_at,csrf_token_hash) VALUES
('17170000-0000-0000-0000-000000001021',:'tenant_a',:'actor_a',decode(:'token_hash_a','hex'),'17170000-0000-0000-0000-000000001031',decode(repeat('81',32),'hex'),clock_timestamp()+interval '1 hour',clock_timestamp()+interval '12 hours',decode(:'csrf_hash_a','hex')),
('17170000-0000-0000-0000-000000001022',:'tenant_b',:'actor_b',decode(:'token_hash_b','hex'),'17170000-0000-0000-0000-000000001032',decode(repeat('82',32),'hex'),clock_timestamp()+interval '1 hour',clock_timestamp()+interval '12 hours',decode(:'csrf_hash_b','hex'));
INSERT INTO pvnaive.users(id,tenant_id,username,display_name,status,created_by_actor_id) VALUES
(:'user_a',:'tenant_a','task13-user-a','Task13 User A','active',:'actor_a'),
(:'user_b',:'tenant_b','task13-user-b','Task13 User B','active',:'actor_b');
INSERT INTO pvnaive.naive_runtime_credentials(id,username,secret_hash,secret_ciphertext,secret_nonce,encryption_key_id,status,origin,created_by_actor_id,updated_by_actor_id) VALUES
(:'cred_a','task13-runtime-a',decode(repeat('41',32),'hex'),decode(repeat('51',16),'hex'),decode(repeat('61',12),'hex'),'runtime-v1','active','panel',:'actor_a',:'actor_a'),
(:'cred_b','task13-runtime-b',decode(repeat('42',32),'hex'),decode(repeat('52',16),'hex'),decode(repeat('62',12),'hex'),'runtime-v1','active','panel',:'actor_b',:'actor_b');
BEGIN;
SELECT pvnaive.direct_naive_accounting_enter_context();
INSERT INTO pvnaive.service_terms(id,tenant_id,user_id,quota_bytes,duration_seconds,start_policy,purchased_at,starts_at,expires_at,state,accounting_baseline_state,accounting_baseline_source,accounting_baseline_cutoff_at,accounting_baseline_upload_bytes,accounting_baseline_download_bytes,concurrency_limit,unique_ip_limit) VALUES
(:'term_a',:'tenant_a',:'user_a',1000000,3600,'on_creation',clock_timestamp()-interval '5 minutes',clock_timestamp()-interval '5 minutes',clock_timestamp()+interval '55 minutes','active','known','fresh_managed_term',clock_timestamp()-interval '5 minutes',0,0,2,2),
(:'term_b',:'tenant_b',:'user_b',1000000,3600,'on_creation',clock_timestamp()-interval '5 minutes',clock_timestamp()-interval '5 minutes',clock_timestamp()+interval '55 minutes','active','known','fresh_managed_term',clock_timestamp()-interval '5 minutes',0,0,2,2);
SELECT pvnaive.direct_naive_accounting_leave_context();
COMMIT;
INSERT INTO pvnaive.user_runtime_credentials(id,tenant_id,user_id,service_term_id,runtime_credential_id,role,bound_at) VALUES
('17170000-0000-0000-0000-000000001071',:'tenant_a',:'user_a',:'term_a',:'cred_a','primary',clock_timestamp()-interval '5 minutes'),
('17170000-0000-0000-0000-000000001072',:'tenant_b',:'user_b',:'term_b',:'cred_b','primary',clock_timestamp()-interval '5 minutes');
INSERT INTO pvnaive.direct_naive_accounting_sessions(runtime_credential_id,node_id,boot_id,session_id,service_term_id,first_observed_at,last_observed_at,last_sequence,upload_cumulative,download_cumulative,final,accounting_complete) VALUES
(:'cred_a','node-task13-a',:'boot_a',:'session_a',:'term_a',clock_timestamp()-interval '20 seconds',clock_timestamp()-interval '1 second',3,100,200,false,true),
(:'cred_b','node-task13-b',:'boot_b',:'session_b',:'term_b',clock_timestamp()-interval '20 seconds',clock_timestamp()-interval '1 second',3,300,400,false,true);
SQL

PGPASSWORD="${app_password}" psql --no-psqlrc --set ON_ERROR_STOP=1 --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username pvnaive_app --dbname "${test_db}" --command \
  "SELECT * FROM pvnaive.direct_naive_accounting_record_session_peer('${cred_a}','node-task13-a','${boot_a}','${session_a}','203.0.113.21',clock_timestamp()); SELECT * FROM pvnaive.direct_naive_accounting_record_session_peer('${cred_b}','node-task13-b','${boot_b}','${session_b}','203.0.113.22',clock_timestamp());" >/dev/null

dd if=/dev/urandom of="${tmpdir}/auth.key" bs=32 count=1 status=none
dd if=/dev/urandom of="${tmpdir}/runtime.key" bs=32 count=1 status=none
chmod 0600 "${tmpdir}/auth.key" "${tmpdir}/runtime.key"
runtime_socket="${tmpdir}/runtime-agent.sock"
control_socket="${tmpdir}/session-control.sock"
control_log="${tmpdir}/control.jsonl"

PVNAIVE_RUNTIME_AGENT_SOCKET="${runtime_socket}" "${agent_binary}" >"${tmpdir}/agent.log" 2>&1 &
agent_pid=$!
for _ in $(seq 1 30); do
  if [[ -S "${runtime_socket}" ]] && curl --fail --silent --unix-socket "${runtime_socket}" http://unix/v1/health | grep -q '"status":"ok"'; then break; fi
  sleep 1
done
[[ -S "${runtime_socket}" ]] || { cat "${tmpdir}/agent.log" >&2; echo 'ERROR: runtime agent socket missing' >&2; exit 1; }

python3 - "${control_socket}" "${control_log}" <<'PY' >"${tmpdir}/control.log" 2>&1 &
import json, os, socket, sys
path, log_path = sys.argv[1:3]
try: os.unlink(path)
except FileNotFoundError: pass
server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
server.bind(path)
server.listen(2)
conn, _ = server.accept()
data = b""
while b"\r\n\r\n" not in data:
    chunk = conn.recv(4096)
    if not chunk: break
    data += chunk
head, body = data.split(b"\r\n\r\n", 1)
content_length = 0
for line in head.decode("iso-8859-1").split("\r\n")[1:]:
    if line.lower().startswith("content-length:"):
        content_length = int(line.split(":", 1)[1].strip())
while len(body) < content_length:
    body += conn.recv(4096)
payload = json.loads(body[:content_length].decode("utf-8"))
with open(log_path, "a", encoding="utf-8") as fh:
    fh.write(json.dumps(payload, sort_keys=True) + "\n")
response = b'{"found":true,"killed":true}\n'
conn.sendall(b"HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: " + str(len(response)).encode() + b"\r\nConnection: close\r\n\r\n" + response)
conn.close(); server.close()
PY
control_pid=$!
for _ in $(seq 1 30); do [[ -S "${control_socket}" ]] && break; sleep 0.1; done
[[ -S "${control_socket}" ]] || { cat "${tmpdir}/control.log" >&2; echo 'ERROR: fake session-control socket missing' >&2; exit 1; }

PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" PVNAIVE_DB_NAME="${test_db}" \
PVNAIVE_DB_USER=pvnaive_app PVNAIVE_DB_CONNECT_TIMEOUT=5 PGPASSWORD="${app_password}" PGSSLMODE=disable \
PVNAIVE_AUTH_KEY_FILE="${tmpdir}/auth.key" PVNAIVE_RUNTIME_KEY_FILE="${tmpdir}/runtime.key" PVNAIVE_RUNTIME_KEY_ID=runtime-task13-v1 \
PVNAIVE_RUNTIME_AGENT_SOCKET="${runtime_socket}" PVNAIVE_SESSION_CONTROL_SOCKET="${control_socket}" \
PVNAIVE_NAIVE_PUBLIC_HOST="naive-task13.example.invalid:443" PVNAIVE_EXPECTED_SCHEMA_VERSION=20 PVNAIVE_LISTEN="127.0.0.1:${api_port}" \
  "${api_binary}" >"${tmpdir}/api.log" 2>&1 &
api_pid=$!
for _ in $(seq 1 30); do
  if curl --fail --silent "http://127.0.0.1:${api_port}/api/v1/health/ready" | grep -q '"ready":true'; then break; fi
  sleep 1
done
kill -0 "${api_pid}" >/dev/null 2>&1 || { cat "${tmpdir}/api.log" >&2; echo 'ERROR: Task13 rehearsal API exited' >&2; exit 1; }

cookie_a="__Host-pvnaive_session=${raw_session_a}; __Host-pvnaive_csrf=${raw_csrf_a}"
base="http://127.0.0.1:${api_port}"

csrf_status="$(curl --silent --output "${tmpdir}/csrf.json" --write-out '%{http_code}' --cookie "${cookie_a}" --request DELETE "${base}/api/v1/users/${user_a}/sessions/${session_a}")"
[[ "${csrf_status}" == 403 ]] || { cat "${tmpdir}/csrf.json" >&2; echo "ERROR: missing-CSRF session kill returned ${csrf_status}" >&2; exit 1; }
[[ ! -s "${control_log}" ]] || { echo 'ERROR: CSRF failure reached session-control side effect' >&2; exit 1; }

idor_status="$(curl --silent --output "${tmpdir}/idor.json" --write-out '%{http_code}' --cookie "${cookie_a}" --header "X-CSRF-Token: ${raw_csrf_a}" --request DELETE "${base}/api/v1/users/${user_b}/sessions/${session_b}")"
[[ "${idor_status}" == 404 ]] || { cat "${tmpdir}/idor.json" >&2; echo "ERROR: cross-tenant session kill returned ${idor_status}" >&2; exit 1; }
[[ ! -s "${control_log}" ]] || { echo 'ERROR: IDOR attempt reached session-control side effect' >&2; exit 1; }

own_status="$(curl --silent --output "${tmpdir}/own.json" --write-out '%{http_code}' --cookie "${cookie_a}" --header "X-CSRF-Token: ${raw_csrf_a}" --request DELETE "${base}/api/v1/users/${user_a}/sessions/${session_a}")"
[[ "${own_status}" == 200 ]] || { cat "${tmpdir}/own.json" >&2; cat "${tmpdir}/api.log" >&2; echo "ERROR: owned session kill returned ${own_status}" >&2; exit 1; }
jq -e --arg sid "${session_a}" '.status=="completed" and .found==true and .killed==true and .session_id==$sid and .credential_mutated==false' "${tmpdir}/own.json" >/dev/null
wait "${control_pid}"; control_pid=''
[[ "$(wc -l < "${control_log}")" == 1 ]] || { echo 'ERROR: expected exactly one session-control request' >&2; exit 1; }
jq -e --arg cred "${cred_a}" --arg boot "${boot_a}" --arg sid "${session_a}" \
  '.runtime_credential_id==$cred and .node_id=="node-task13-a" and .boot_id==$boot and .session_id==$sid' "${control_log}" >/dev/null

credential_state="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT status FROM pvnaive.naive_runtime_credentials WHERE id='${cred_a}'")"
[[ "${credential_state}" == active ]] || { echo "ERROR: session kill mutated credential state=${credential_state}" >&2; exit 1; }

echo 'TASK13_API_SESSION_KILL_REHEARSAL=PASSED'
