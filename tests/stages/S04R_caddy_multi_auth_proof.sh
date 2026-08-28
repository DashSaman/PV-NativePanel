#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

caddy_bin="${CADDY_BIN:-}"
[[ -n "${caddy_bin}" && -x "${caddy_bin}" ]] || { echo 'ERROR: CADDY_BIN must point to the pinned executable' >&2; exit 1; }

version="$(${caddy_bin} version)"
[[ "${version}" == v2.11.2* ]] || { echo "ERROR: unexpected Caddy version: ${version}" >&2; exit 1; }
"${caddy_bin}" list-modules | grep -Fx 'http.handlers.forward_proxy' >/dev/null || {
  echo 'ERROR: pinned Caddy lacks http.handlers.forward_proxy' >&2
  exit 1
}

tmpdir="$(mktemp -d)"
trap 'rm -rf -- "${tmpdir}"' EXIT HUP INT TERM
cat >"${tmpdir}/Caddyfile" <<'CADDY'
{
  admin off
  auto_https off
  order forward_proxy before file_server
}

:18443 {
  forward_proxy {
    basic_auth proof.one proof-password-one-123
    basic_auth proof.two proof-password-two-456
    hide_ip
    hide_via
    probe_resistance proof-path-123
  }

  respond "camouflage" 200
}
CADDY

"${caddy_bin}" validate --config "${tmpdir}/Caddyfile" --adapter caddyfile >/dev/null
"${caddy_bin}" adapt --config "${tmpdir}/Caddyfile" --adapter caddyfile >/dev/null

binary_sha="$(sha256sum "${caddy_bin}" | awk '{print $1}')"
[[ "${#binary_sha}" == 64 ]] || { echo 'ERROR: could not fingerprint Caddy binary' >&2; exit 1; }
printf 'CADDY_MULTI_BASIC_AUTH_PROOF=PASSED\nCADDY_VERSION=%s\nCADDY_BINARY_SHA256=%s\n' "${version}" "${binary_sha}"
