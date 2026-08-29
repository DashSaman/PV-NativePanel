#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage_id="S05-PREFLIGHT"
panel_host="${PVNAIVE_PUBLIC_HOST:-namir.softarg.ir}"
naive_public_host="${PVNAIVE_NAIVE_PUBLIC_HOST:-}"
caddy_file="/etc/caddy/Caddyfile"
caddy_bin="/usr/local/bin/caddy"
db_env="/etc/pvnaive/db.env"
api_env="/etc/pvnaive/api.env"
failures=()

record_failure() { failures+=("$1"); printf 'CHECK_%s=FAIL\n' "$1"; }
record_pass() { printf 'CHECK_%s=PASS\n' "$1"; }

validate_public_host() {
  local value="$1" host port=""
  [[ -n "${value}" ]] || return 1
  [[ "${value}" != *://* && "${value}" != */* && "${value}" != *' '* ]] || return 1
  host="${value}"
  if [[ "${value}" == *:* ]]; then
    host="${value%:*}"
    port="${value##*:}"
    [[ "${port}" =~ ^[1-9][0-9]{0,4}$ ]] || return 1
    ((10#${port} <= 65535)) || return 1
  fi
  [[ "${host}" =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]]
}

# Bash regex above intentionally has no shell expansion; normalize the final
# host check separately to avoid accepting scheme/path/whitespace.
validate_public_host() {
  local value="$1" host port=""
  [[ -n "${value}" ]] || return 1
  [[ "${value}" != *://* && "${value}" != */* && "${value}" != *' '* ]] || return 1
  host="${value}"
  if [[ "${value}" == *:* ]]; then
    host="${value%:*}"
    port="${value##*:}"
    [[ "${port}" =~ ^[1-9][0-9]{0,4}$ ]] || return 1
    ((10#${port} <= 65535)) || return 1
  fi
  [[ "${host}" =~ ^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$ ]]
}
