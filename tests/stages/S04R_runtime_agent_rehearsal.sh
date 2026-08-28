#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "${repo_root}"

command -v go >/dev/null 2>&1 || { echo "ERROR: go is required" >&2; exit 1; }

# The package tests inject a fixed command runner and disposable Caddyfile /
# backup roots. This exercises validate -> backup -> reload-only -> verify and
# rollback/compensation without touching production paths or secrets.
go test ./internal/runtimeagent -run '^TestOperator' -count=1

# Static capability gates ensure the production implementation cannot silently
# drift from the narrow contract even if a test fake changes later.
grep -Fq 'runSystemctl(ctx, "reload", o.config.serviceName)' internal/runtimeagent/operator.go
if grep -Eq 'runSystemctl\([^\n]*"restart"|runner\.Run\([^\n]*"restart"' internal/runtimeagent/operator.go; then
  echo "ERROR: runtime operator contains forbidden restart execution" >&2
  exit 1
fi

grep -Fq 'productionCaddyfilePath = "/etc/caddy/Caddyfile"' internal/runtimeagent/operator.go
grep -Fq 'productionCaddyBinary   = "/usr/local/bin/caddy"' internal/runtimeagent/operator.go
grep -Fq 'productionServiceName   = "caddy-naive.service"' internal/runtimeagent/operator.go
grep -Fq 'productionBackupRoot    = "/var/backups/pvnaive/caddy"' internal/runtimeagent/operator.go

grep -Fq 'RestrictAddressFamilies=AF_UNIX' ops/systemd/pvnaive-runtime-agent.service
grep -Fq 'IPAddressDeny=any' ops/systemd/pvnaive-runtime-agent.service
if grep -Eq 'AF_INET|AF_INET6' ops/systemd/pvnaive-runtime-agent.service; then
  echo "ERROR: runtime agent unit unexpectedly permits IP sockets" >&2
  exit 1
fi

echo "S04R_RUNTIME_AGENT_REHEARSAL=PASSED"
