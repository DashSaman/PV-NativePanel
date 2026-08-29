#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

repo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "${repo_root}"

command -v go >/dev/null 2>&1 || { echo "ERROR: go is required" >&2; exit 1; }

go test ./internal/runtimeagent -run '^TestOperator' -count=1

grep -Fq 'runSystemctl(ctx, "reload", o.config.serviceName)' internal/runtimeagent/operator.go
if grep -Eq 'runSystemctl\([^\n]*"restart"|runner\.Run\([^\n]*"restart"' internal/runtimeagent/operator.go; then
  echo "ERROR: runtime operator contains forbidden restart execution" >&2
  exit 1
fi

grep -Fq 'productionCaddyfilePath = "/etc/caddy/Caddyfile"' internal/runtimeagent/operator.go
grep -Fq 'productionCaddyBinary   = "/usr/local/bin/caddy"' internal/runtimeagent/operator.go
grep -Fq 'productionServiceName   = "caddy-naive.service"' internal/runtimeagent/operator.go
grep -Fq 'productionBackupRoot    = "/var/backups/pvnaive/caddy"' internal/runtimeagent/operator.go

# /run/pvnaive is shared by two independently-owned sockets. A service-level
# RuntimeDirectory declaration is forbidden because systemd may re-own existing
# entries when that service starts. The shared namespace is created at boot by
# tmpfiles instead, and each process owns only its own socket.
if grep -Eq '^RuntimeDirectory(Mode|Preserve)?=' ops/systemd/pvnaive-runtime-agent.service; then
  echo "ERROR: runtime agent must not own the shared /run/pvnaive lifecycle" >&2
  exit 1
fi
grep -Fq 'After=caddy-naive.service systemd-tmpfiles-setup.service' ops/systemd/pvnaive-runtime-agent.service || {
  echo "ERROR: runtime agent must start after tmpfiles creates the shared runtime directory" >&2
  exit 1
}
grep -Fxq 'd /run/pvnaive 0771 root pvnaive -' ops/tmpfiles/pvnaive.conf || {
  echo "ERROR: tmpfiles contract for shared /run/pvnaive is missing" >&2
  exit 1
}
grep -Fq 'os.MkdirAll(runtimeDir, 0771)' cmd/pvnaive-runtime-agent/main.go || {
  echo "ERROR: runtime agent should retain a defensive runtime directory check" >&2
  exit 1
}
grep -Fq 'os.Chmod(runtimeDir, 0771)' cmd/pvnaive-runtime-agent/main.go || {
  echo "ERROR: runtime agent must enforce directory mode without touching peer sockets" >&2
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
