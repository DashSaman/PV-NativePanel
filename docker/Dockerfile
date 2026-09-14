# PVNaive — all-in-one Docker image (PostgreSQL 18 + API + agents + accounting Caddy)
# Build from the repository root:
#   docker build -t pvnaive:latest -f docker/Dockerfile .
#
# Contents:
#   - PostgreSQL 18 (PGDG bookworm base) — data in /var/lib/postgresql
#   - pvnaive (API), pvnaive-runtime-agent, pvnaive-telemetry-agent, pvnaive-password
#   - Caddy pinned to klzgrad/forwardproxy with the PVNaive exact-accounting patch
#   - web panel static files (base path /panel/)
FROM postgres:18-bookworm

LABEL org.opencontainers.image.title="PVNaive" \
      org.opencontainers.image.description="NaiveProxy management panel with exact accounting" \
      org.opencontainers.image.source="https://github.com/DashSaman/PV-NativePanel"

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates curl bash openssl \
 && rm -rf /var/lib/apt/lists/* \
 && groupadd --system pvnaive \
 && useradd --system --gid pvnaive --no-create-home --shell /usr/sbin/nologin pvnaive

# --- pvnaive binaries (prebuilt; see scripts/release/build-docker-context.sh) ---
COPY dist/docker/pvnaive-linux-amd64               /opt/pvnaive/bin/pvnaive
COPY dist/docker/pvnaive-password-linux-amd64      /opt/pvnaive/bin/pvnaive-password
COPY dist/docker/pvnaive-runtime-agent-linux-amd64 /opt/pvnaive/bin/pvnaive-runtime-agent
COPY dist/docker/pvnaive-telemetry-agent-linux-amd64 /opt/pvnaive/bin/pvnaive-telemetry-agent
COPY dist/docker/caddy-pvnaive-accounting          /opt/pvnaive/bin/caddy
COPY dist/docker/web/                              /opt/pvnaive/web/

RUN chmod 0755 /opt/pvnaive/bin/* \
 && /opt/pvnaive/bin/caddy version

# --- migrations (checksum manifest verified at runtime) ---
COPY db/migrations/ /opt/pvnaive/db/migrations/
COPY scripts/db/migrate.sh scripts/db/lib.sh /opt/pvnaive/scripts/db/

# --- runtime orchestration ---
COPY docker/entrypoint.sh docker/Caddyfile.template /opt/pvnaive/docker/
RUN chmod 0755 /opt/pvnaive/docker/entrypoint.sh \
 && chmod 0644 /opt/pvnaive/docker/Caddyfile.template \
 && ln -sf /opt/pvnaive/docker/entrypoint.sh /usr/local/bin/pvnaive-docker-entrypoint \
 && mkdir -p /run/pvnaive /etc/caddy /etc/pvnaive /var/backups/pvnaive/caddy

ENV PVNAIVE_LISTEN=127.0.0.1:8080 \
    PVNAIVE_DB_HOST=127.0.0.1 \
    PVNAIVE_DB_PORT=5432 \
    PVNAIVE_DB_NAME=pvnaive \
    PVNAIVE_DB_USER=pvnaive_app \
    PVNAIVE_DB_CONNECT_TIMEOUT=5 \
    PVNAIVE_AUTH_KEY_FILE=/var/lib/pvnaive/secrets/auth.key \
    PVNAIVE_RUNTIME_KEY_FILE=/var/lib/pvnaive/secrets/runtime.key \
    PVNAIVE_RUNTIME_KEY_ID=runtime-v1 \
    PVNAIVE_RUNTIME_AGENT_SOCKET=/run/pvnaive/runtime-agent.sock \
    PVNAIVE_SESSION_CONTROL_SOCKET=/run/pvnaive/session-control.sock \
    PVNAIVE_RUNTIME_RELOAD_MODE=caddy-admin \
    PVNAIVE_RUNTIME_CADDY_BINARY=/opt/pvnaive/bin/caddy \
    PVNAIVE_RUNTIME_CADDYFILE_PATH=/etc/caddy/Caddyfile \
    PGDATA=/var/lib/postgresql/data

EXPOSE 80 443

ENTRYPOINT ["/usr/local/bin/pvnaive-docker-entrypoint"]
