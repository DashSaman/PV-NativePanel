#!/usr/bin/env bash
# PVNaive all-in-one Docker entrypoint.
# Orchestrates: PostgreSQL 18 -> schema migrations -> secrets -> owner bootstrap
# -> runtime/telemetry agents -> pvnaive API -> pinned accounting Caddy (foreground).
set -Eeuo pipefail

PVNAIVE_ROOT=/opt/pvnaive
SECRETS_DIR=/var/lib/pvnaive/secrets
PGDATA="${PGDATA:-/var/lib/postgresql/data}"
PG_USER=postgres
PG_SOCKET=/var/run/postgresql
DB_NAME=pvnaive
DB_APP_ROLE=pvnaive_app
DB_OWNER_ROLE=pvnaive_owner
export PGDATA

log() { printf '[pvnaive] %s\n' "$*" >&2; }
fail() { printf '[pvnaive] FATAL: %s\n' "$*" >&2; exit 1; }

psql_admin() { gosu postgres psql --no-psqlrc --set ON_ERROR_STOP=1 -h "${PG_SOCKET}" -U "${PG_USER}" "$@"; }

# ---------------------------------------------------------------- PostgreSQL
mkdir -p "${PGDATA}" "${SECRETS_DIR}" /run/pvnaive /etc/caddy
chmod 0700 "${PGDATA}" "${SECRETS_DIR}" 2>/dev/null || true

if [[ ! -s "${PGDATA}/PG_VERSION" ]]; then
  log "initialising PostgreSQL data directory"
  chown postgres:postgres "${PGDATA}" 2>/dev/null || true
  gosu postgres initdb --pgdata="${PGDATA}" --auth-local=trust --auth-host=scram-sha-256 >/dev/null
  cat >> "${PGDATA}/pg_hba.conf" <<'HBA'
# PVNaive hardening: app role uses scram only over loopback TCP.
hostssl all all 127.0.0.1/32 reject
HBA
fi

log "starting PostgreSQL"
gosu postgres pg_ctl -D "${PGDATA}" -o "-c listen_addresses='127.0.0.1' -p ${PVNAIVE_DB_PORT:-5432} -c unix_socket_directories='${PG_SOCKET}'" -w -t 60 start >/dev/null

# ------------------------------------------------------------- roles + database
if ! psql_admin -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='${DB_APP_ROLE}'" | grep -q 1; then
  log "creating database roles"
  app_password="$(openssl rand -hex 32)"
  psql_admin -d postgres <<SQL
CREATE ROLE ${DB_OWNER_ROLE} NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
CREATE ROLE ${DB_APP_ROLE} LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 40;
ALTER ROLE ${DB_APP_ROLE} PASSWORD '${app_password}';
ALTER ROLE ${DB_APP_ROLE} SET statement_timeout = '30s';
ALTER ROLE ${DB_APP_ROLE} SET lock_timeout = '5s';
ALTER ROLE ${DB_APP_ROLE} SET idle_in_transaction_session_timeout = '30s';
ALTER ROLE ${DB_APP_ROLE} SET row_security = on;
SQL
  install -d -m 0700 "${SECRETS_DIR}"
  printf '127.0.0.1:%s:%s:%s:%s\n' "${PVNAIVE_DB_PORT:-5432}" "${DB_NAME}" "${DB_APP_ROLE}" "${app_password}" > "${SECRETS_DIR}/db.pgpass"
else
  log "database roles already exist"
fi
# The pvnaive group needs to traverse the secrets dir for the agent pgpass.
chown root:pvnaive "${SECRETS_DIR}" 2>/dev/null || true
chmod 0750 "${SECRETS_DIR}" 2>/dev/null || true
# libpq only accepts a PGPASSFILE owned by the running user, so keep two
# identical copies: root-owned for the API, pvnaive-owned for the telemetry
# agent. The API/auth/runtime keys stay root-only.
if [[ -s "${SECRETS_DIR}/db.pgpass" ]]; then
  chown pvnaive:pvnaive "${SECRETS_DIR}/db.pgpass.agent" 2>/dev/null || true
  cp "${SECRETS_DIR}/db.pgpass" "${SECRETS_DIR}/db.pgpass.agent"
  chown pvnaive:pvnaive "${SECRETS_DIR}/db.pgpass.agent"
  chmod 0600 "${SECRETS_DIR}/db.pgpass.agent"
fi
chmod 0600 "${SECRETS_DIR}/db.pgpass" 2>/dev/null || true
chown root:root "${SECRETS_DIR}/db.pgpass" 2>/dev/null || true
export PGPASSFILE="${SECRETS_DIR}/db.pgpass"

if ! psql_admin -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1; then
  log "creating database ${DB_NAME}"
  psql_admin -d postgres -c "CREATE DATABASE ${DB_NAME} OWNER ${DB_OWNER_ROLE} ENCODING 'UTF8' TEMPLATE template0" >/dev/null
  psql_admin -d "${DB_NAME}" <<'SQL'
REVOKE ALL ON DATABASE pvnaive FROM PUBLIC;
GRANT CONNECT ON DATABASE pvnaive TO pvnaive_app;
REVOKE TEMPORARY ON DATABASE pvnaive FROM pvnaive_app;
ALTER ROLE pvnaive_app IN DATABASE pvnaive SET search_path = 'pg_catalog,pvnaive';
SQL
fi

# ----------------------------------------------------------------- migrations
# Migrations run over the local Unix socket as the postgres superuser (the
# native S04 flow does the same); the API itself keeps using the least-
# privilege pvnaive_app role over TCP with a password file.
log "applying checksum-verified schema migrations"
PVNAIVE_DB_HOST="${PG_SOCKET}" PVNAIVE_DB_PORT=5432 PVNAIVE_DB_NAME="${DB_NAME}" PVNAIVE_DB_USER="${PG_USER}" \
  bash "${PVNAIVE_ROOT}/scripts/db/migrate.sh"

# -------------------------------------------------------------------- secrets
if [[ ! -s "${PVNAIVE_AUTH_KEY_FILE}" ]]; then
  log "generating authentication key (32 bytes)"
  openssl rand -out "${PVNAIVE_AUTH_KEY_FILE}" 32
fi
if [[ ! -s "${PVNAIVE_RUNTIME_KEY_FILE}" ]]; then
  log "generating runtime encryption key (32 bytes)"
  openssl rand -out "${PVNAIVE_RUNTIME_KEY_FILE}" 32
fi
chmod 0600 "${PVNAIVE_AUTH_KEY_FILE}" "${PVNAIVE_RUNTIME_KEY_FILE}"

# ------------------------------------------------------------- owner bootstrap
owner_count="$(psql_admin -d "${DB_NAME}" -tAc "SELECT COUNT(*) FROM pvnaive.actors WHERE actor_role='owner' AND email <> 'scheduler@pvnaive.invalid'")" || owner_count="?"
if [[ "${owner_count}" == "0" ]]; then
  owner_email="${PVNAIVE_OWNER_EMAIL:-}"
  owner_password="${PVNAIVE_OWNER_PASSWORD:-}"
  if [[ -z "${owner_email}" ]]; then
    log "no owner account exists; set PVNAIVE_OWNER_EMAIL (and optionally PVNAIVE_OWNER_PASSWORD) and restart to create one"
  else
    [[ "${owner_email}" =~ ^[^[:space:]@]+@[^[:space:]@]+\.[^[:space:]@]+$ ]] || fail "invalid PVNAIVE_OWNER_EMAIL"
    if [[ -z "${owner_password}" ]]; then
      owner_password="$(openssl rand -base64 18)"
      log "generated owner password: ${owner_password}"
    fi
    hash="$(printf '%s\n' "${owner_password}" | "${PVNAIVE_ROOT}/bin/pvnaive-password")"
    [[ "${hash}" == '$argon2id$'* ]] || fail "password hashing failed"
    log "creating owner account ${owner_email}"
    psql_admin -d "${DB_NAME}" <<SQL
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('pvnaive-owner-bootstrap'));
SET LOCAL ROLE ${DB_OWNER_ROLE};
DO \$block\$
BEGIN
  IF EXISTS (SELECT 1 FROM pvnaive.actors WHERE actor_role = 'owner' AND email <> 'scheduler@pvnaive.invalid') THEN
    RAISE EXCEPTION 'owner already exists' USING ERRCODE = '23505';
  END IF;
END
\$block\$;
INSERT INTO pvnaive.actors (
  tenant_id, actor_role, email, display_name, password_hash, mfa_required, status, password_changed_at
) VALUES (
  NULL, 'owner', '${owner_email//\'/\'\'}', 'Owner', '${hash//\'/\'\'}', false, 'active', clock_timestamp()
);
COMMIT;
SQL
  fi
else
  log "owner account present (count=${owner_count})"
fi

# --------------------------------------------------------------------- Caddyfile
# The persisted Caddyfile survives container recreation in the bind mount, so
# the initial template render only happens when it is genuinely absent. The
# bootstrap credential below is a SEED for the first runtime apply: as soon as
# the first customer credential exists, `reconcile-runtime-config` and every
# later runtime apply replace the whole basic_auth block with the database
# truth (DEPLOY-001: never render bootstrap-only over existing customers).
if [[ ! -f /etc/caddy/Caddyfile ]]; then
  log "rendering /etc/caddy/Caddyfile"
  : "${PVNAIVE_PROXY_USER:?PVNAIVE_PROXY_USER must be set (bootstrap proxy account)}"
  : "${PVNAIVE_PROXY_PASSWORD:?PVNAIVE_PROXY_PASSWORD must be set (bootstrap proxy account)}"
  domain="${PVNAIVE_DOMAIN:-localhost}"
  # Literal values are written (never Caddy {$VAR} placeholders): the runtime
  # agent re-parses this file with its own strict Caddyfile lexer.
  tls_line=""
  if [[ "${PVNAIVE_TLS_MODE:-}" == "internal" ]] || [[ "${domain}" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    tls_line=$'\ttls internal'
    log "TLS mode: internal (self-signed) for ${domain}"
  fi
  cat > /etc/caddy/Caddyfile <<CADDYEOF
{
        order forward_proxy before file_server
        default_sni ${domain}
}

:443, ${domain} {
        encode gzip
${tls_line}

        handle_path /panel/* {
                root * /opt/pvnaive/web
                try_files {path} /index.html
                file_server
        }

        handle /api/v1/* {
                reverse_proxy 127.0.0.1:8080
        }

        handle /sub/* {
                reverse_proxy 127.0.0.1:8080
        }

        handle /s/* {
                reverse_proxy 127.0.0.1:8080
        }

        forward_proxy {
                basic_auth ${PVNAIVE_PROXY_USER} "${PVNAIVE_PROXY_PASSWORD}"
                hide_ip
                hide_via
                probe_resistance
                pvnaive_accounting_socket /run/pvnaive/accounting.sock
                pvnaive_node_id pvnaive-node-1
                pvnaive_runtime_credential ${PVNAIVE_PROXY_USER} 00000000-0000-0000-0000-100000000001
        }

        root * /opt/pvnaive/camouflage
        file_server
}
CADDYEOF
  chmod 0644 /etc/caddy/Caddyfile
fi

# ------------------------------------------------- expected schema (lineage truth)
# The readiness gate compares the database schema version against
# PVNAIVE_EXPECTED_SCHEMA_VERSION. Baking the value into the image made image
# and database disagree after every lineage extension (readiness 503 with
# schema=error). Derive it from the bundled migration count instead; an
# explicitly configured value still wins for controlled downgrades.
if [[ -z "${PVNAIVE_EXPECTED_SCHEMA_VERSION:-}" ]]; then
  derived_schema="$(find /opt/pvnaive/db/migrations -maxdepth 1 -type f -name '*.up.sql' | wc -l)"
  export PVNAIVE_EXPECTED_SCHEMA_VERSION="${derived_schema}"
  log "derived PVNAIVE_EXPECTED_SCHEMA_VERSION=${derived_schema} from bundled migrations"
fi

# ------------------------------------------------- boot credential reconcile (DEPLOY-001)
# Merge the persisted config's forward_proxy credentials with the active
# database credentials on EVERY boot. The database is the single source of
# truth; the reconcile renders through the exact production renderer,
# validates the candidate with the pinned proxy binary and atomically swaps
# it in before the proxy starts. Failure is non-fatal: boot continues on the
# existing file (degraded to the previous behaviour) with a loud warning.
log "reconciling proxy credentials with database truth"
if ! "${PVNAIVE_ROOT}/bin/pvnaive" reconcile-runtime-config \
     --config /etc/caddy/Caddyfile \
     --caddy-binary "${PVNAIVE_RUNTIME_CADDY_BINARY:-/opt/pvnaive/bin/caddy}"; then
  log "WARNING: reconcile-runtime-config failed — keeping the existing Caddyfile; customer credentials may be stale"
fi

# ----------------------------------------------------------- background services
gosu pvnaive env PGPASSFILE="${SECRETS_DIR}/db.pgpass.agent" "${PVNAIVE_ROOT}/bin/pvnaive-telemetry-agent" &
telemetry_pid=$!
"${PVNAIVE_ROOT}/bin/pvnaive-runtime-agent" &
runtime_pid=$!
"${PVNAIVE_ROOT}/bin/pvnaive" &
api_pid=$!

shutdown() {
  log "shutting down"
  kill "${api_pid}" "${runtime_pid}" "${telemetry_pid}" 2>/dev/null || true
  gosu postgres pg_ctl -D "${PGDATA}" -m fast -w stop >/dev/null 2>&1 || true
  exit 0
}
trap shutdown INT TERM

# ------------------------------------------------------------ readiness + caddy
ready=0
for _ in $(seq 1 30); do
  if curl -fsS "http://127.0.0.1:${PVNAIVE_LISTEN##*:}/api/v1/health/live" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
[[ "${ready}" == "1" ]] || fail "pvnaive API did not become ready"
log "pvnaive API ready — starting accounting Caddy on :80/:443"
exec "${PVNAIVE_ROOT}/bin/caddy" run --config /etc/caddy/Caddyfile --adapter caddyfile
