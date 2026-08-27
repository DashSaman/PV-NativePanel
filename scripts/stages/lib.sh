#!/usr/bin/env bash

die() {
  echo "ERROR: $*" >&2
  return 1
}

pvnaive_tcp_port_is_listening() {
  local required_port="$1"
  [[ "${required_port}" =~ ^[0-9]+$ ]] || return 2

  awk -v port="${required_port}" '$4 ~ (":" port "$") {found=1} END {exit !found}'
}
