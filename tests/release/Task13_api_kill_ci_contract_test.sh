#!/usr/bin/env bash
set -Eeuo pipefail
root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
workflow="${root}/.github/workflows/ci.yml"
grep -Fq 'bash tests/stages/Task13_api_session_kill_rehearsal.sh' "${workflow}" || {
  echo 'ERROR: CI does not execute the Task13 DB/auth session-kill rehearsal' >&2
  exit 1
}
grep -Fq 'apt-get install --yes --no-install-recommends curl jq python3 >/dev/null' "${workflow}" || {
  echo 'ERROR: Task13 rehearsal dependency python3 is not provisioned in CI' >&2
  exit 1
}
echo 'TASK13_API_KILL_CI_CONTRACT=PASSED'
