#!/usr/bin/env bash
set -Eeuo pipefail
suite="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)/direct_naive_accounting_pg18_test.sh"

# Schema20 extends the public ingest boundary with trusted client_ip while
# retaining a default so legacy 11-argument calls used by this baseline suite
# keep exercising the same accounting semantics. Adapt only the exact
# to_regprocedure() signature assertion; do not rewrite any behavior calls.
if grep -q "direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text)" \
  "${suite%/*}/../../db/migrations/0020_unique_ip_limit.up.sql" 2>/dev/null; then
  tmp="${suite%/*}/.direct_naive_accounting_pg18_schema20_${BASHPID}.sh"
  trap 'rm -f -- "${tmp:-}"' EXIT
  sed "s/direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean)')/direct_naive_accounting_ingest(uuid,text,text,uuid,uuid,bigint,timestamptz,boolean,bigint,bigint,boolean,text)')/" \
    "${suite}" >"${tmp}"
  bash "${tmp}" "$@"
  exit $?
fi

exec bash "${suite}" "$@"
