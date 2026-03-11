# ── Stage 1: Build ────────────────────────────────────────────────────────────
# Compile TypeScript and prune dev dependencies.
FROM node:22-alpine AS builder

WORKDIR /build

COPY package*.json ./
RUN npm ci

COPY tsconfig.json ./
COPY src/ ./src/

RUN npm run build \
 && npm prune --omit=dev

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
# node:22-alpine — same base as builder to keep native module ABIs compatible.
# docker-cli     — provides the `docker` binary.
# docker-compose — installs Compose V2 as a Docker CLI plugin (`docker compose`).
FROM node:22-alpine AS runtime

RUN apk add --no-cache docker-cli docker-compose

WORKDIR /app

COPY --from=builder /build/node_modules ./node_modules
COPY --from=builder /build/dist        ./dist
COPY public/ ./public/
COPY package.json ./

ENV NODE_ENV=production

# Metrics / health endpoint (configurable via MONITOR_PORT)
EXPOSE 7070

# ── Important runtime constraints ─────────────────────────────────────────────
# 1. /var/run/docker.sock must be mounted from the host (docker compose control).
# 2. The project root must be mounted at the SAME absolute path as on the host
#    so that `docker compose -f <path>` build contexts resolve correctly.
# 3. Compose file paths (ROOT_COMPOSE_FILE, ENGINE_COMPOSE_FILE, …) must be set
#    to absolute paths inside the container.
# 4. network_mode: host is recommended so that localhost:{5432,6379,1323} resolve
#    to the host services as they do in native (non-container) mode.
#
# See services/monitoring/docker-compose.yml for a ready-made deployment example.
# ──────────────────────────────────────────────────────────────────────────────

CMD ["node", "dist/index.js"]
