#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
helper="${repo_root}/cmd/pvnaive-password/main.go"
script="${repo_root}/scripts/auth/bootstrap-owner.sh"

[[ -f "${helper}" ]] || { echo "ERROR: missing password helper" >&2; exit 1; }
[[ -f "${script}" ]] || { echo "ERROR: missing owner bootstrap" >&2; exit 1; }

# Static production-safety contract. The script itself remains interactive and
# is exercised on the target only after the S04 localhost gate passes.
grep -Fq '[[ ${EUID} -eq 0 ]]' "${script}"
grep -Fq 'testAmir5-3' "${script}"
grep -Fq '/dev/tty' "${script}"
grep -Fq 'S03_DATABASE.json' "${script}"
grep -Fq 'schema_migrations' "${script}"
grep -Fq "actor_role = 'owner'" "${script}"
grep -Fq 'owner already exists' "${script}"
grep -Fq 'pvnaive-password' "${script}"

if grep -Eiq '(default[_ -]?password|password[[:space:]]*=[[:space:]]*["'"'][^"'"']+["'"'])' "${script}"; then
  echo "ERROR: bootstrap appears to contain a default password" >&2
  exit 1
fi
if grep -Eq 'set[[:space:]]+-x|echo[[:space:]].*password|printf[[:space:]].*password_hash' "${script}"; then
  echo "ERROR: bootstrap may log password material" >&2
  exit 1
fi

echo "PVNAIVE_BOOTSTRAP_OWNER_TEST=PASSED"
