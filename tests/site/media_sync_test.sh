#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
sync_script="${repo_root}/scripts/site/sync_public_media.py"
[[ -f "${sync_script}" ]] || { echo 'ERROR: missing public media sync tool' >&2; exit 1; }

tmp="$(mktemp -d)"
server_pid=''
cleanup() {
  if [[ -n "${server_pid}" ]]; then kill "${server_pid}" >/dev/null 2>&1 || true; fi
  rm -rf "${tmp}"
}
trap cleanup EXIT

serve="${tmp}/serve"
root="${tmp}/root"
mkdir -p "${serve}" "${root}"
printf 'verified-webm-fixture\n' > "${serve}/fixture.webm"
printf 'wrong mime fixture\n' > "${serve}/fixture.txt"
bytes="$(wc -c < "${serve}/fixture.webm" | tr -d ' ')"
sha="$(sha256sum "${serve}/fixture.webm" | awk '{print $1}')"
port="$(python3 - <<'PY'
import socket
s=socket.socket(); s.bind(('127.0.0.1',0)); print(s.getsockname()[1]); s.close()
PY
)"
python3 -m http.server "${port}" --bind 127.0.0.1 --directory "${serve}" >"${tmp}/http.log" 2>&1 &
server_pid=$!
for _ in $(seq 1 30); do
  python3 - <<PY >/dev/null 2>&1 && break || true
import urllib.request
urllib.request.urlopen('http://127.0.0.1:${port}/fixture.webm', timeout=.3).read(1)
PY
  sleep .1
done

write_manifest() {
  local path="$1" url="$2" local_path="$3" expected_bytes="$4" expected_sha="$5" mime="$6"
  cat >"${path}" <<JSON
{"version":1,"items":[{"id":"fixture","slug":"fixture","kind":"video","title":"Fixture","summary":"Fixture","source_name":"Fixture Source","source_domain":"127.0.0.1","source_url":"https://example.com/source","attribution":"Fixture attribution","rights_note":"CC BY 4.0 fixture","mirror_allowed":true,"qualities":[{"label":"Original","mime":"${mime}","bytes":${expected_bytes},"url":"${url}","local_path":"${local_path}","sha256":"${expected_sha}"}]}]}
JSON
}

manifest="${tmp}/media.json"
write_manifest "${manifest}" "http://127.0.0.1:${port}/fixture.webm" 'video/fixture.webm' "${bytes}" "${sha}" 'video/webm'

# Production policy must reject HTTP even if the fixture itself is otherwise valid.
if python3 "${sync_script}" --manifest "${manifest}" --root "${root}" >/dev/null 2>"${tmp}/http.err"; then
  echo 'ERROR: sync accepted HTTP without fixture override' >&2; exit 1
fi
grep -qi 'HTTPS' "${tmp}/http.err" || { echo 'ERROR: HTTP rejection did not explain HTTPS policy' >&2; exit 1; }

# Dry-run must not create the destination file.
python3 "${sync_script}" --manifest "${manifest}" --root "${root}" --allow-http-fixtures >"${tmp}/dry.out"
grep -q 'PLAN mirror fixture' "${tmp}/dry.out" || { echo 'ERROR: dry-run did not report planned mirror' >&2; exit 1; }
[[ ! -e "${root}/video/fixture.webm" ]] || { echo 'ERROR: dry-run wrote media' >&2; exit 1; }

# Apply must verify and atomically promote the exact file.
python3 "${sync_script}" --manifest "${manifest}" --root "${root}" --allow-http-fixtures --apply >"${tmp}/apply.out"
grep -q 'SYNCED fixture' "${tmp}/apply.out" || { echo 'ERROR: apply did not report sync' >&2; exit 1; }
cmp -s "${serve}/fixture.webm" "${root}/video/fixture.webm" || { echo 'ERROR: mirrored file differs from source fixture' >&2; exit 1; }

# Second apply is idempotent and must keep the verified file.
python3 "${sync_script}" --manifest "${manifest}" --root "${root}" --allow-http-fixtures --apply >"${tmp}/again.out"
grep -q 'SKIP verified fixture' "${tmp}/again.out" || { echo 'ERROR: second sync was not idempotent' >&2; exit 1; }

# Traversal must be rejected before any write.
write_manifest "${tmp}/traversal.json" "http://127.0.0.1:${port}/fixture.webm" '../escape.webm' "${bytes}" "${sha}" 'video/webm'
if python3 "${sync_script}" --manifest "${tmp}/traversal.json" --root "${root}" --allow-http-fixtures --apply >/dev/null 2>&1; then
  echo 'ERROR: sync accepted path traversal' >&2; exit 1
fi
[[ ! -e "${tmp}/escape.webm" ]] || { echo 'ERROR: traversal escaped media root' >&2; exit 1; }

# Wrong response MIME must be rejected.
write_manifest "${tmp}/mime.json" "http://127.0.0.1:${port}/fixture.txt" 'video/wrong.webm' "$(wc -c < "${serve}/fixture.txt" | tr -d ' ')" "$(sha256sum "${serve}/fixture.txt" | awk '{print $1}')" 'video/webm'
if python3 "${sync_script}" --manifest "${tmp}/mime.json" --root "${root}" --allow-http-fixtures --apply >/dev/null 2>"${tmp}/mime.err"; then
  echo 'ERROR: sync accepted wrong MIME' >&2; exit 1
fi
grep -qi 'MIME' "${tmp}/mime.err" || { echo 'ERROR: MIME rejection missing diagnostic' >&2; exit 1; }

# Size mismatch must be rejected without replacing a good target.
write_manifest "${tmp}/size.json" "http://127.0.0.1:${port}/fixture.webm" 'video/size.webm' "$((bytes+1))" "${sha}" 'video/webm'
if python3 "${sync_script}" --manifest "${tmp}/size.json" --root "${root}" --allow-http-fixtures --apply >/dev/null 2>"${tmp}/size.err"; then
  echo 'ERROR: sync accepted size mismatch' >&2; exit 1
fi
grep -qi 'size' "${tmp}/size.err" || { echo 'ERROR: size rejection missing diagnostic' >&2; exit 1; }

# Checksum mismatch must be rejected.
bad_sha="$(printf '0%.0s' $(seq 1 64))"
write_manifest "${tmp}/sha.json" "http://127.0.0.1:${port}/fixture.webm" 'video/sha.webm' "${bytes}" "${bad_sha}" 'video/webm'
if python3 "${sync_script}" --manifest "${tmp}/sha.json" --root "${root}" --allow-http-fixtures --apply >/dev/null 2>"${tmp}/sha.err"; then
  echo 'ERROR: sync accepted checksum mismatch' >&2; exit 1
fi
grep -qi 'checksum' "${tmp}/sha.err" || { echo 'ERROR: checksum rejection missing diagnostic' >&2; exit 1; }

# Explicit item cap must stop oversized media before promotion.
if python3 "${sync_script}" --manifest "${manifest}" --root "${tmp}/small-root" --allow-http-fixtures --apply --max-item-bytes 1 >/dev/null 2>"${tmp}/cap.err"; then
  echo 'ERROR: sync ignored max item size' >&2; exit 1
fi
grep -qi 'limit' "${tmp}/cap.err" || { echo 'ERROR: size-limit rejection missing diagnostic' >&2; exit 1; }

echo 'PUBLIC_MEDIA_SYNC_CONTRACT=PASSED'
