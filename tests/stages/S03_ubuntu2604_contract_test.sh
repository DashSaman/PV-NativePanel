#!/usr/bin/env bash
set -Eeuo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd -- "${script_dir}/../.." && pwd -P)"
stage="${repo_root}/scripts/stages/S03-database.sh"
backup="${repo_root}/scripts/db/backup.sh"
restore="${repo_root}/scripts/db/restore.sh"

for file in "${stage}" "${backup}" "${restore}"; do
  [[ -f "${file}" ]] || { echo "ERROR: missing ${file}" >&2; exit 1; }
  bash -n "${file}"
done

grep -Fqx 'expected_postgres_major="18"' "${stage}"
grep -Fqx 'packages=(postgresql postgresql-client age)' "${stage}"
if grep -Fq 'postgresql-contrib' "${stage}"; then
  echo 'ERROR: S03 must not request the obsolete postgresql-contrib meta package on Ubuntu 26.04' >&2
  exit 1
fi

grep -Fq 'apt-get --simulate install --no-install-recommends' "${stage}"
grep -Fq 'pgcrypto.control' "${stage}"
grep -Fq 'package install requires a reboot before S03 can continue' "${stage}"

if grep -Fq "awk '/Candidate:/ {print \$2; exit}'" "${stage}"; then
  echo 'ERROR: early-exit APT candidate parser can trigger SIGPIPE under pipefail' >&2
  exit 1
fi

grep -Fq "awk '/Candidate:/ {print \$2}'" "${stage}"
grep -Fq 'awk -v expected="${expected_ipv4}"' "${stage}"

if grep -Eq -- '--no-(owner|acl)' "${backup}"; then
  echo 'ERROR: database backup must preserve ownership and ACLs' >&2
  exit 1
fi

grep -Fq '"ownership_and_acls": true' "${backup}"
grep -Fq 'restore connection must be a PostgreSQL superuser' "${restore}"
if grep -Eq -- '--no-(owner|acl)' "${restore}"; then
  echo 'ERROR: restore drill must replay ownership and ACLs' >&2
  exit 1
fi
if grep -Fq -- '--role pvnaive_owner' "${restore}"; then
  echo 'ERROR: restore drill must load data as superuser so FORCE RLS cannot block backup data replay' >&2
  exit 1
fi

grep -Fq 'PVNAIVE_RESTORE_OWNERSHIP=PASSED' "${restore}"
grep -Fq 'PVNAIVE_RESTORE_ACLS=PASSED' "${restore}"
grep -Fq 'PVNAIVE_RESTORE_SIGNING_KEY=PASSED' "${restore}"

echo 'S03_UBUNTU2604_CONTRACT_TEST=PASSED'
