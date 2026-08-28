#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
api_binary="${PVNAIVE_REHEARSAL_API_BINARY:-${repo_root}/dist/rehearsal/pvnaive}"
password_binary="${PVNAIVE_REHEARSAL_PASSWORD_BINARY:-${repo_root}/dist/rehearsal/pvnaive-password}"
agent_binary="${PVNAIVE_REHEARSAL_RUNTIME_AGENT_BINARY:-${repo_root}/dist/rehearsal/pvnaive-runtime-agent-rehearsal}"
: "${PVNAIVE_DB_HOST:=127.0.0.1}"
: "${PVNAIVE_DB_PORT:=5432}"
: "${PVNAIVE_DB_USER:=postgres}"
export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_USER

for required in psql createdb dropdb curl jq sha256sum; do
  command -v "${required}" >/dev/null 2>&1 || { echo "ERROR: missing ${required}" >&2; exit 1; }
done
for binary in "${api_binary}" "${password_binary}" "${agent_binary}"; do
  [[ -x "${binary}" ]] || { echo "ERROR: rehearsal binary missing: ${binary}" >&2; exit 1; }
done

test_db="pvnaive"
api_port="18081"
api_pid=""
agent_pid=""
tmpdir=""
password='S04R-Rehearsal-Owner-Password'
owner_email='owner-s04r@example.invalid'

psql_admin() {
  psql --no-psqlrc --set ON_ERROR_STOP=1 \
    --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" \
    --username "${PVNAIVE_DB_USER}" "$@"
}

cleanup() {
  if [[ -n "${api_pid}" ]]; then
    kill "${api_pid}" >/dev/null 2>&1 || true
    wait "${api_pid}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${agent_pid}" ]]; then
    kill "${agent_pid}" >/dev/null 2>&1 || true
    wait "${agent_pid}" >/dev/null 2>&1 || true
  fi
  psql_admin --dbname postgres --command \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='${test_db}' AND pid <> pg_backend_pid()" >/dev/null 2>&1 || true
  dropdb --if-exists --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" "${test_db}" >/dev/null 2>&1 || true
  psql_admin --dbname postgres --command 'DROP ROLE IF EXISTS pvnaive_app; DROP ROLE IF EXISTS pvnaive_owner;' >/dev/null 2>&1 || true
  if [[ -n "${tmpdir}" ]]; then
    rm -rf -- "${tmpdir}" || true
  fi
}
cleanup
tmpdir="$(mktemp -d)"
trap cleanup EXIT HUP INT TERM

psql_admin --dbname postgres <<'SQL' >/dev/null
CREATE ROLE pvnaive_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE pvnaive_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD 'pvnaive-s04r-ci-only';
ALTER ROLE pvnaive_app SET row_security = on;
SQL
createdb --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}" \
  --owner pvnaive_owner --encoding UTF8 --template template0 "${test_db}"

PVNAIVE_DB_NAME="${test_db}" PVNAIVE_MIGRATIONS_DIR="${repo_root}/db/migrations" \
  "${repo_root}/scripts/db/migrate.sh" >/dev/null
schema_version="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command 'SELECT COALESCE(MAX(version),0) FROM pvnaive.schema_migrations')"
[[ "${schema_version}" == "4" ]] || { echo "ERROR: S04R rehearsal schema version=${schema_version}" >&2; exit 1; }

phc="$(printf '%s\n' "${password}" | "${password_binary}")"
[[ "${phc}" == '$argon2id$'* ]] || { echo 'ERROR: password helper returned invalid PHC' >&2; exit 1; }
psql_admin --dbname "${test_db}" --set=owner_email="${owner_email}" --set=owner_hash="${phc}" <<'SQL' >/dev/null
SET ROLE pvnaive_owner;
INSERT INTO pvnaive.actors (tenant_id, actor_role, email, display_name, password_hash, mfa_required, status, password_changed_at)
VALUES (NULL, 'owner', :'owner_email', 'S04R Rehearsal Owner', :'owner_hash', false, 'active', clock_timestamp());
SQL
unset phc

dd if=/dev/urandom of="${tmpdir}/auth.key" bs=32 count=1 status=none
dd if=/dev/urandom of="${tmpdir}/runtime.key" bs=32 count=1 status=none
chmod 0600 "${tmpdir}/auth.key" "${tmpdir}/runtime.key"
socket="${tmpdir}/runtime-agent.sock"

PVNAIVE_RUNTIME_AGENT_SOCKET="${socket}" "${agent_binary}" >"${tmpdir}/agent.log" 2>&1 &
agent_pid=$!
for _ in $(seq 1 30); do
  if [[ -S "${socket}" ]] && curl --fail --silent --unix-socket "${socket}" http://unix/v1/health | grep -q '"status":"ok"'; then
    break
  fi
  sleep 1
done
kill -0 "${agent_pid}" >/dev/null 2>&1 || { cat "${tmpdir}/agent.log" >&2; echo 'ERROR: rehearsal runtime agent exited' >&2; exit 1; }
[[ -S "${socket}" ]] || { echo 'ERROR: rehearsal runtime socket missing' >&2; exit 1; }

initial_sha="$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq -r '.caddy_sha256')"
[[ "${#initial_sha}" == 64 ]] || { echo 'ERROR: invalid initial runtime SHA' >&2; exit 1; }

PVNAIVE_DB_HOST="${PVNAIVE_DB_HOST}" \
PVNAIVE_DB_PORT="${PVNAIVE_DB_PORT}" \
PVNAIVE_DB_NAME="${test_db}" \
PVNAIVE_DB_USER=pvnaive_app \
PVNAIVE_DB_CONNECT_TIMEOUT=5 \
PGPASSWORD=pvnaive-s04r-ci-only \
PGSSLMODE=disable \
PVNAIVE_AUTH_KEY_FILE="${tmpdir}/auth.key" \
PVNAIVE_RUNTIME_KEY_FILE="${tmpdir}/runtime.key" \
PVNAIVE_RUNTIME_KEY_ID=runtime-rehearsal-v1 \
PVNAIVE_RUNTIME_AGENT_SOCKET="${socket}" \
PVNAIVE_LISTEN="127.0.0.1:${api_port}" \
  "${api_binary}" >"${tmpdir}/api.log" 2>&1 &
api_pid=$!

for _ in $(seq 1 30); do
  if curl --fail --silent "http://127.0.0.1:${api_port}/api/v1/health/ready" | grep -q '"ready":true'; then
    break
  fi
  sleep 1
done
kill -0 "${api_pid}" >/dev/null 2>&1 || { cat "${tmpdir}/api.log" >&2; echo 'ERROR: rehearsal API exited' >&2; exit 1; }

login_body="$(printf '{"email":"%s","password":"%s","totp_code":""}' "${owner_email}" "${password}")"
curl --fail --silent --show-error \
  --cookie-jar "${tmpdir}/cookies.txt" \
  --header 'Content-Type: application/json' \
  --data "${login_body}" \
  "http://127.0.0.1:${api_port}/api/v1/auth/login" >"${tmpdir}/login.json"
jq -e '.status == "authenticated" and .role == "owner"' "${tmpdir}/login.json" >/dev/null
csrf="$(awk '$6 == "__Host-pvnaive_csrf" {print $7}' "${tmpdir}/cookies.txt" | tail -n1)"
[[ -n "${csrf}" ]] || { echo 'ERROR: CSRF cookie missing' >&2; exit 1; }

api_get() {
  curl --fail --silent --show-error --cookie "${tmpdir}/cookies.txt" "http://127.0.0.1:${api_port}$1"
}

api_mutate() {
  local method="$1" path="$2" idem="$3" body="$4" revision="${5:-}"
  local args=(--fail --silent --show-error --cookie "${tmpdir}/cookies.txt" --request "${method}"
    --header 'Content-Type: application/json' --header "X-CSRF-Token: ${csrf}" --header "Idempotency-Key: ${idem}")
  if [[ -n "${revision}" ]]; then
    args+=(--header "If-Match: ${revision}")
  fi
  curl "${args[@]}" --data "${body}" "http://127.0.0.1:${api_port}${path}"
}

api_get /api/v1/runtime/naive/credentials >"${tmpdir}/before.json"
jq -e '.credentials | length == 0' "${tmpdir}/before.json" >/dev/null

api_mutate POST /api/v1/runtime/naive/import runtime-import-0001 '{}' >"${tmpdir}/import.json"
jq -e '.status == "imported" and (.credentials | length == 1) and .credentials[0].username == "legacy.user" and .credentials[0].origin == "imported"' "${tmpdir}/import.json" >/dev/null
if grep -Eq 'legacy-pass|secret_(hash|ciphertext|nonce)|generated_password' "${tmpdir}/import.json"; then
  echo 'ERROR: import response exposed secret material' >&2
  exit 1
fi
after_import_sha="$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq -r '.caddy_sha256')"
[[ "${after_import_sha}" == "${initial_sha}" ]] || { echo 'ERROR: guarded import changed runtime state' >&2; exit 1; }

api_mutate POST /api/v1/runtime/naive/credentials runtime-create-0001 \
  '{"username":"customer.one","password":"","generate_password":true}' >"${tmpdir}/create.json"
created_password="$(jq -r '.generated_password // empty' "${tmpdir}/create.json")"
[[ "${#created_password}" == 32 ]] || { echo 'ERROR: generated password was not returned once after commit' >&2; exit 1; }
jq -e '.credential.username == "customer.one" and .credential.status == "active" and .credential.origin == "panel"' "${tmpdir}/create.json" >/dev/null

replay_status="$(curl --silent --output "${tmpdir}/replay.json" --write-out '%{http_code}' \
  --cookie "${tmpdir}/cookies.txt" --request POST \
  --header 'Content-Type: application/json' --header "X-CSRF-Token: ${csrf}" --header 'Idempotency-Key: runtime-create-0001' \
  --data '{"username":"customer.replay","password":"","generate_password":true}' \
  "http://127.0.0.1:${api_port}/api/v1/runtime/naive/credentials")"
[[ "${replay_status}" == "409" ]] || { echo "ERROR: idempotency replay returned ${replay_status}" >&2; exit 1; }
if grep -q 'generated_password' "${tmpdir}/replay.json"; then
  echo 'ERROR: idempotency replay exposed generated secret' >&2
  exit 1
fi

agent_count="$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq '.credentials | length')"
[[ "${agent_count}" == "2" ]] || { echo "ERROR: runtime agent credential count after create=${agent_count}" >&2; exit 1; }

api_get /api/v1/runtime/naive/credentials >"${tmpdir}/list.json"
if grep -Eq 'legacy-pass|customer.one.*password|secret_(hash|ciphertext|nonce)|generated_password' "${tmpdir}/list.json"; then
  echo 'ERROR: list response exposed secret material' >&2
  exit 1
fi
customer_id="$(jq -r '.credentials[] | select(.username == "customer.one") | .id' "${tmpdir}/list.json")"
customer_revision="$(jq -r '.credentials[] | select(.username == "customer.one") | .revision' "${tmpdir}/list.json")"
legacy_id="$(jq -r '.credentials[] | select(.username == "legacy.user") | .id' "${tmpdir}/list.json")"
legacy_revision="$(jq -r '.credentials[] | select(.username == "legacy.user") | .revision' "${tmpdir}/list.json")"
[[ -n "${customer_id}" && -n "${legacy_id}" ]] || { echo 'ERROR: imported/created IDs missing' >&2; exit 1; }

api_mutate PATCH "/api/v1/runtime/naive/credentials/${customer_id}" runtime-rename-0001 \
  '{"username":"customer.renamed","status":"active"}' "${customer_revision}" >"${tmpdir}/rename.json"
renamed_revision="$(jq -r '.credential.revision' "${tmpdir}/rename.json")"
jq -e '.credential.username == "customer.renamed" and .credential.status == "active"' "${tmpdir}/rename.json" >/dev/null
jq -e '[.credentials[] | select(.username == "customer.renamed")] | length == 1' \
  < <(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect) >/dev/null

stale_status="$(curl --silent --output "${tmpdir}/stale.json" --write-out '%{http_code}' \
  --cookie "${tmpdir}/cookies.txt" --request PATCH \
  --header 'Content-Type: application/json' --header "X-CSRF-Token: ${csrf}" --header 'Idempotency-Key: runtime-stale-revision-0001' \
  --header "If-Match: ${customer_revision}" --data '{"username":"customer.stale","status":"active"}' \
  "http://127.0.0.1:${api_port}/api/v1/runtime/naive/credentials/${customer_id}")"
[[ "${stale_status}" == "409" ]] || { echo "ERROR: stale revision returned ${stale_status}" >&2; exit 1; }
jq -e '.code == "revision_conflict"' "${tmpdir}/stale.json" >/dev/null

api_mutate PATCH "/api/v1/runtime/naive/credentials/${customer_id}" runtime-disable-0001 \
  '{"username":"customer.renamed","status":"disabled"}' "${renamed_revision}" >"${tmpdir}/disable.json"
disabled_revision="$(jq -r '.credential.revision' "${tmpdir}/disable.json")"
[[ "$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq '.credentials | length')" == "1" ]] || { echo 'ERROR: disabled credential remained active in runtime' >&2; exit 1; }

last_active_status="$(curl --silent --output "${tmpdir}/last-active.json" --write-out '%{http_code}' \
  --cookie "${tmpdir}/cookies.txt" --request DELETE \
  --header 'Content-Type: application/json' --header "X-CSRF-Token: ${csrf}" --header 'Idempotency-Key: runtime-last-active-0001' \
  --header "If-Match: ${legacy_revision}" --data '{}' \
  "http://127.0.0.1:${api_port}/api/v1/runtime/naive/credentials/${legacy_id}")"
[[ "${last_active_status}" == "409" ]] || { echo "ERROR: last active revoke returned ${last_active_status}" >&2; exit 1; }
jq -e '.code == "last_active_credential"' "${tmpdir}/last-active.json" >/dev/null

api_mutate PATCH "/api/v1/runtime/naive/credentials/${customer_id}" runtime-enable-0001 \
  '{"username":"customer.renamed","status":"active"}' "${disabled_revision}" >"${tmpdir}/enable.json"
enabled_revision="$(jq -r '.credential.revision' "${tmpdir}/enable.json")"

api_mutate POST "/api/v1/runtime/naive/credentials/${customer_id}/rotate-password" runtime-rotate-0001 \
  '{"password":"","generate_password":true}' "${enabled_revision}" >"${tmpdir}/rotate.json"
rotated_password="$(jq -r '.generated_password // empty' "${tmpdir}/rotate.json")"
rotated_revision="$(jq -r '.credential.revision' "${tmpdir}/rotate.json")"
[[ "${#rotated_password}" == 32 && "${rotated_password}" != "${created_password}" ]] || { echo 'ERROR: rotation did not return a fresh one-time password' >&2; exit 1; }

api_mutate DELETE "/api/v1/runtime/naive/credentials/${customer_id}" runtime-revoke-0001 '{}' "${rotated_revision}" >"${tmpdir}/revoke.json"
jq -e '.credential.status == "revoked"' "${tmpdir}/revoke.json" >/dev/null
[[ "$(curl --fail --silent --unix-socket "${socket}" http://unix/v1/inspect | jq '.credentials | length')" == "1" ]] || { echo 'ERROR: revoked credential remained active in runtime' >&2; exit 1; }

api_get /api/v1/runtime/naive/credentials >"${tmpdir}/final-list.json"
jq -e '(.credentials | length == 2) and ([.credentials[] | select(.username == "legacy.user" and .status == "active")] | length == 1) and ([.credentials[] | select(.username == "customer.renamed" and .status == "revoked")] | length == 1)' "${tmpdir}/final-list.json" >/dev/null
if grep -Eq "${created_password}|${rotated_password}|legacy-pass|secret_(hash|ciphertext|nonce)|generated_password" "${tmpdir}/final-list.json"; then
  echo 'ERROR: final list exposed secret material' >&2
  exit 1
fi

state_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT COUNT(*) || '|' || COUNT(*) FILTER (WHERE origin='imported') || '|' || COUNT(*) FILTER (WHERE origin='panel') || '|' || COUNT(*) FILTER (WHERE status='revoked') FROM pvnaive.naive_runtime_credentials")"
[[ "${state_contract}" == "2|1|1|1" ]] || { echo "ERROR: runtime credential DB contract=${state_contract}" >&2; exit 1; }

envelope_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT COUNT(*) FILTER (WHERE octet_length(secret_hash)=32 AND octet_length(secret_nonce)=12 AND octet_length(secret_ciphertext)>16 AND encryption_key_id='runtime-rehearsal-v1') FROM pvnaive.naive_runtime_credentials")"
[[ "${envelope_contract}" == "2" ]] || { echo "ERROR: encrypted credential envelope contract=${envelope_contract}" >&2; exit 1; }

plaintext_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align \
  --set=legacy_password='legacy-pass' --set=created_password="${created_password}" --set=rotated_password="${rotated_password}" <<'SQL'
SELECT CASE WHEN
  bool_and(position(convert_to(:'legacy_password','UTF8') in secret_ciphertext)=0
    AND position(convert_to(:'created_password','UTF8') in secret_ciphertext)=0
    AND position(convert_to(:'rotated_password','UTF8') in secret_ciphertext)=0)
  AND (SELECT bool_and(position(convert_to(:'legacy_password','UTF8') in config_ciphertext)=0
    AND position(convert_to(:'created_password','UTF8') in config_ciphertext)=0
    AND position(convert_to(:'rotated_password','UTF8') in config_ciphertext)=0)
    FROM pvnaive.runtime_revisions WHERE protocol_id='naive' AND tenant_id IS NULL)
THEN 'clean' ELSE 'plaintext-found' END
FROM pvnaive.naive_runtime_credentials;
SQL
)"
[[ "${plaintext_contract}" == "clean" ]] || { echo 'ERROR: plaintext credential material found in DB encrypted columns' >&2; exit 1; }

revision_contract="$(psql_admin --dbname "${test_db}" --tuples-only --no-align --command "SELECT COUNT(*) || '|' || COUNT(*) FILTER (WHERE state='applied') FROM pvnaive.runtime_revisions WHERE protocol_id='naive' AND tenant_id IS NULL")"
[[ "${revision_contract}" == "6|6" ]] || { echo "ERROR: runtime revision contract=${revision_contract}" >&2; exit 1; }

echo 'S04R_FULL_REHEARSAL=PASSED'
