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

# systemd evaluates ReadWritePaths before ExecStart. The runtime directory must
# therefore be created by systemd itself; the Go process cannot be responsible
# for creating a path required by its own mount namespace setup.
#
# Exact accounting adds a non-root telemetry writer in group pvnaive and a
# Caddy reader in group pvnaive-telemetry. 0771 is deliberate: group pvnaive
# can create accounting.sock, while unrelated users (including Caddy before
# its telemetry supplementary group is applied) get traverse-only access and
# cannot list the directory or access runtime-agent.sock (0660 root:pvnaive).
grep -Fq 'RuntimeDirectory=pvnaive' ops/systemd/pvnaive-runtime-agent.service || {
  echo "ERROR: runtime agent unit must let systemd create /run/pvnaive before namespace setup" >&2
  exit 1
}
grep -Fq 'RuntimeDirectoryMode=0771' ops/systemd/pvnaive-runtime-agent.service || {
  echo "ERROR: runtime agent RuntimeDirectoryMode must be 0771 for telemetry writer + Caddy traversal" >&2
  exit 1
}
grep -Fq 'os.MkdirAll(runtimeDir, 0771)' cmd/pvnaive-runtime-agent/main.go || {
  echo "ERROR: runtime agent must create /run/pvnaive with mode 0771" >&2
  exit 1
}
grep -Fq 'os.Chmod(runtimeDir, 0771)' cmd/pvnaive-runtime-agent/main.go || {
  echo "ERROR: runtime agent must enforce /run/pvnaive mode 0771 after chown" >&2
  exit 1
}
grep -Fq 'os.Chmod(runtimeagent.DefaultSocketPath, 0660)' cmd/pvnaive-runtime-agent/main.go || {
  echo "ERROR: runtime-agent socket must remain 0660" >&2
  exit 1
}
grep -Fq 'ReadWritePaths=/run/pvnaive' ops/systemd/pvnaive-runtime-agent.service

grep -Fq 'RestrictAddressFamilies=AF_UNIX' ops/systemd/pvnaive-runtime-agent.service
grep -Fq 'IPAddressDeny=any' ops/systemd/pvnaive-runtime-agent.service
if grep -Eq 'AF_INET|AF_INET6' ops/systemd/pvnaive-runtime-agent.service; then
  echo "ERROR: runtime agent unit unexpectedly permits IP sockets" >&2
  exit 1
fi

echo "S04R_RUNTIME_AGENT_REHEARSAL=PASSED"
