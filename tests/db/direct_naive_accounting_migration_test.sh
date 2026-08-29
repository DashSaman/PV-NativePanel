#!/usr/bin/env bash
set -Eeuo pipefail
suite="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)/direct_naive_accounting_pg18_test.sh"
exec bash "${suite}" "$@"
