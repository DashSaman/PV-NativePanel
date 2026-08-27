#!/usr/bin/env bash
# Shared PVNaive PostgreSQL command helpers.

pvnaive_die() {
  echo "ERROR: $*" >&2
  exit 1
}

pvnaive_require_command() {
  command -v "$1" >/dev/null 2>&1 || pvnaive_die "required command not found: $1"
}

pvnaive_require_identifier() {
  local label="$1"
  local value="$2"
  [[ "${value}" =~ ^[a-z_][a-z0-9_]{0,62}$ ]] || pvnaive_die "unsafe ${label}: ${value}"
}

pvnaive_validate_storage_root() {
  local storage_root="$1"
  local canonical_storage_root
  pvnaive_require_command realpath
  canonical_storage_root="$(realpath --canonicalize-missing -- "${storage_root}")"
  [[ "${storage_root%/}" == "${canonical_storage_root}" ]] || \
    pvnaive_die "storage root must be an absolute canonical path without symlink traversal"
  [[ "${canonical_storage_root}" =~ ^/[^/]+/[^/]+/[^/]+(/.*)?$ ]] || \
    pvnaive_die "storage root is too broad"
}

pvnaive_db_defaults() {
  : "${PVNAIVE_DB_HOST:=127.0.0.1}"
  : "${PVNAIVE_DB_PORT:=5432}"
  : "${PVNAIVE_DB_NAME:=pvnaive}"
  : "${PVNAIVE_DB_USER:=pvnaive_app}"
  : "${PVNAIVE_DB_CONNECT_TIMEOUT:=5}"
  export PVNAIVE_DB_HOST PVNAIVE_DB_PORT PVNAIVE_DB_NAME PVNAIVE_DB_USER PVNAIVE_DB_CONNECT_TIMEOUT
  export PGCONNECT_TIMEOUT="${PVNAIVE_DB_CONNECT_TIMEOUT}"
  pvnaive_require_identifier "database name" "${PVNAIVE_DB_NAME}"
  pvnaive_require_identifier "database user" "${PVNAIVE_DB_USER}"
  [[ "${PVNAIVE_DB_PORT}" =~ ^[1-9][0-9]{0,4}$ ]] || pvnaive_die "invalid database port"
  ((10#${PVNAIVE_DB_PORT} <= 65535)) || pvnaive_die "invalid database port"
}

pvnaive_db_tool() {
  local tool="$1"
  shift
  local command=("${tool}" --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}")
  if [[ -n "${PVNAIVE_RUN_AS_OS_USER:-}" ]]; then
    runuser -u "${PVNAIVE_RUN_AS_OS_USER}" -- "${command[@]}" "$@"
  else
    "${command[@]}" "$@"
  fi
}

pvnaive_psql() {
  pvnaive_db_tool psql --no-psqlrc --set ON_ERROR_STOP=1 --dbname "${PVNAIVE_DB_NAME}" "$@"
}

pvnaive_psql_at() {
  pvnaive_psql --tuples-only --no-align "$@"
}

pvnaive_admin_tool() {
  local tool="$1"
  shift
  local command=("${tool}" --host "${PVNAIVE_DB_HOST}" --port "${PVNAIVE_DB_PORT}" --username "${PVNAIVE_DB_USER}")
  if [[ -n "${PVNAIVE_RUN_AS_OS_USER:-}" ]]; then
    runuser -u "${PVNAIVE_RUN_AS_OS_USER}" -- "${command[@]}" "$@"
  else
    "${command[@]}" "$@"
  fi
}
