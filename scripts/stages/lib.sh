#!/usr/bin/env bash

die() {
  echo "ERROR: $*" >&2
  return 1
}

pvnaive_is_root_bash_process() {
  local expected_bashpid="$1"
  [[ "${BASHPID}" == "${expected_bashpid}" ]]
}

pvnaive_tcp_port_is_listening() {
  local required_port="$1"
  [[ "${required_port}" =~ ^[1-9][0-9]{0,4}$ ]] || return 2
  ((10#${required_port} <= 65535)) || return 2

  awk -v port="${required_port}" '$4 ~ (":" port "$") {found=1} END {exit !found}'
}

pvnaive_tcp_listener_snapshot() {
  ss -H -lnt
}

pvnaive_tcp_has_non_loopback_postgres_listener() {
  awk '$4 ~ /:5432$/ && $4 !~ /^(127\.0\.0\.1|\[::1\]):5432$/ {bad=1} END {exit !bad}'
}
