#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
site_root="${repo_root}/site"
css="${site_root}/assets/site.css"

video_page="${site_root}/videos/khamenei-1999-speech-clip.html"
audio_page="${site_root}/audio/khamenei-1999-speech-audio.html"
downloads="${site_root}/downloads/index.html"

for path in "${video_page}" "${audio_page}" "${downloads}"; do
  [[ -f "${path}" ]] || { echo "ERROR: missing media page ${path#${repo_root}/}" >&2; exit 1; }
done

grep -Fq '<video controls preload="metadata"' "${video_page}" || {
  echo 'ERROR: video detail must use native HTML5 playback' >&2
  exit 1
}
grep -Fq '<audio controls preload="metadata"' "${audio_page}" || {
  echo 'ERROR: audio detail must use native HTML5 playback' >&2
  exit 1
}
grep -Fq 'download>دانلود از این سایت</a>' "${video_page}" || {
  echo 'ERROR: mirror-approved video must expose local download action' >&2
  exit 1
}
grep -Fq '2.87 MB' "${video_page}" || {
  echo 'ERROR: verified 3,009,132-byte media size must render truthfully' >&2
  exit 1
}
grep -Fq '2.87 MB' "${downloads}" || {
  echo 'ERROR: downloads library must show verified size metadata' >&2
  exit 1
}
grep -Fq '<span class="source-domain">farsi.khamenei.ir</span>' "${video_page}" || {
  echo 'ERROR: media detail must visibly identify its canonical source domain' >&2
  exit 1
}

for css_marker in '.portal-masthead' '.portal-main' '.portal-grid' '.media-player' '.download-table' '.portal-gallery'; do
  grep -Fq "${css_marker}" "${css}" || {
    echo "ERROR: public portal styling missing ${css_marker}" >&2
    exit 1
  }
done

# At least one remote-only item stays explicit about source-hosted delivery.
remote_page="${site_root}/videos/khamenei-1997-inauguration.html"
grep -Fq 'پخش/دانلود از منبع' "${remote_page}" || {
  echo 'ERROR: remote-only media must identify source-hosted delivery' >&2
  exit 1
}
if grep -Fq 'دانلود از این سایت' "${remote_page}"; then
  echo 'ERROR: remote-only media must not claim a local mirror' >&2
  exit 1
fi

echo 'PUBLIC_MEDIA_PAGES_CONTRACT=PASSED'