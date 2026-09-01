#!/usr/bin/env bash
set -Eeuo pipefail
suite="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)/direct_naive_accounting_pg18_test.sh"

# Schema20 extends the public ingest boundary with trusted client_ip while
# retaining a default so legacy 11-argument calls used by this baseline suite
# keep exercising the same accounting semantics. Adapt only the exact
# to_regprocedure() signature assertion; do not rewrite any behavior calls.
tmp="${suite%/*}/.direct_naive_accounting_pg18_current_${BASHPID}.sh"
trap 'rm -f -- "${tmp:-}"' EXIT
sed "s/direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean)')/direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text)')/" \
  "${suite}" >"${tmp}"

if cmp -s "${suite}" "${tmp}"; then
  echo 'ERROR: current accounting ingest signature adapter made no change' >&2
  exit 1
fi

bash "${tmp}" "$@"
