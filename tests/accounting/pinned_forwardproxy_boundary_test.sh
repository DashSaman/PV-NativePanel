#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

UPSTREAM_REPO='https://github.com/klzgrad/forwardproxy.git'
UPSTREAM_COMMIT='d62c80d3dd2c706b6b87579844d2397bddd18317'

tmp_root="$(mktemp -d)"
cleanup() {
  rm -rf -- "${tmp_root}"
}
trap cleanup EXIT HUP INT TERM

repo_dir="${tmp_root}/forwardproxy"
source_file="${tmp_root}/forwardproxy.go"

git init -q "${repo_dir}"
git -C "${repo_dir}" remote add origin "${UPSTREAM_REPO}"
git -C "${repo_dir}" -c protocol.version=2 fetch --quiet --no-tags --depth=1 origin "${UPSTREAM_COMMIT}"

resolved_commit="$(git -C "${repo_dir}" rev-parse FETCH_HEAD)"
[[ "${resolved_commit}" == "${UPSTREAM_COMMIT}" ]] || {
  echo "ERROR: pinned forwardproxy resolved to ${resolved_commit}, expected ${UPSTREAM_COMMIT}" >&2
  exit 1
}

git -C "${repo_dir}" show FETCH_HEAD:forwardproxy.go >"${source_file}"

grep -Fq 'authErr = h.checkCredentials(r)' "${source_file}" || {
  echo 'ERROR: pinned forwardproxy authentication boundary changed' >&2
  exit 1
}
grep -Fq 'targetConn, err := h.dialContextCheckACL(ctx, "tcp", hostPort)' "${source_file}" || {
  echo 'ERROR: pinned forwardproxy no longer owns the CONNECT target dial at the expected boundary' >&2
  exit 1
}
grep -Fq 'return serveHijack(w, targetConn)' "${source_file}" || {
  echo 'ERROR: pinned HTTP/1 CONNECT streaming boundary changed' >&2
  exit 1
}
grep -Fq 'return dualStream(targetConn, r.Body, w,' "${source_file}" || {
  echo 'ERROR: pinned HTTP/2/3 CONNECT streaming boundary changed' >&2
  exit 1
}

auth_line="$(grep -n -F 'authErr = h.checkCredentials(r)' "${source_file}" | head -n1 | cut -d: -f1)"
dial_line="$(grep -n -F 'targetConn, err := h.dialContextCheckACL(ctx, "tcp", hostPort)' "${source_file}" | head -n1 | cut -d: -f1)"
hijack_line="$(grep -n -F 'return serveHijack(w, targetConn)' "${source_file}" | head -n1 | cut -d: -f1)"
dual_line="$(grep -n -F 'return dualStream(targetConn, r.Body, w,' "${source_file}" | head -n1 | cut -d: -f1)"

for value in "${auth_line}" "${dial_line}" "${hijack_line}" "${dual_line}"; do
  [[ "${value}" =~ ^[1-9][0-9]*$ ]] || {
    echo 'ERROR: failed to resolve a pinned forwardproxy source line' >&2
    exit 1
  }
done

((auth_line < dial_line)) || {
  echo 'ERROR: authentication no longer precedes CONNECT target ownership' >&2
  exit 1
}
((dial_line < hijack_line && dial_line < dual_line)) || {
  echo 'ERROR: CONNECT target ownership no longer precedes both streaming paths' >&2
  exit 1
}

echo "FORWARDPROXY_UPSTREAM_COMMIT=${resolved_commit}"
echo "FORWARDPROXY_AUTH_LINE=${auth_line}"
echo "FORWARDPROXY_TARGET_DIAL_LINE=${dial_line}"
echo "FORWARDPROXY_HTTP1_STREAM_LINE=${hijack_line}"
echo "FORWARDPROXY_HTTP23_STREAM_LINE=${dual_line}"
echo 'PINNED_FORWARDPROXY_BOUNDARY_PROOF=PASSED'
